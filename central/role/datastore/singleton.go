package datastore

import (
	"context"

	"github.com/stackrox/rox/central/globaldb"
	groupFilter "github.com/stackrox/rox/central/group/datastore/filter"
	rolePkg "github.com/stackrox/rox/central/role"
	"github.com/stackrox/rox/central/role/resources"
	"github.com/stackrox/rox/central/role/store"
	PermissionSetPGStore "github.com/stackrox/rox/central/role/store/permissionset/postgres"
	permissionSetPGStore "github.com/stackrox/rox/central/role/store/permissionset/rocksdb"
	postgresRolePGStore "github.com/stackrox/rox/central/role/store/role/postgres"
	roleStore "github.com/stackrox/rox/central/role/store/role/rocksdb"
	postgresSimpleAccessScopeStore "github.com/stackrox/rox/central/role/store/simpleaccessscope/postgres"
	simpleAccessScopeStore "github.com/stackrox/rox/central/role/store/simpleaccessscope/rocksdb"
	"github.com/stackrox/rox/generated/storage"
	"github.com/stackrox/rox/pkg/auth/permissions"
	permissionsUtils "github.com/stackrox/rox/pkg/auth/permissions/utils"
	"github.com/stackrox/rox/pkg/defaults/accesscontrol"
	"github.com/stackrox/rox/pkg/env"
	"github.com/stackrox/rox/pkg/sac"
	"github.com/stackrox/rox/pkg/sync"
	"github.com/stackrox/rox/pkg/utils"
)

var (
	ds   DataStore
	once sync.Once
)

// Singleton returns the singleton providing access to the roles store.
func Singleton() DataStore {
	once.Do(func() {
		var roleStorage store.RoleStore
		var permissionSetStorage store.PermissionSetStore
		var accessScopeStorage store.SimpleAccessScopeStore
		if env.PostgresDatastoreEnabled.BooleanSetting() {
			roleStorage = postgresRolePGStore.New(globaldb.GetPostgres())
			permissionSetStorage = PermissionSetPGStore.New(globaldb.GetPostgres())
			accessScopeStorage = postgresSimpleAccessScopeStore.New(globaldb.GetPostgres())
		} else {
			var err error
			roleStorage, err = roleStore.New(globaldb.GetRocksDB())
			utils.CrashOnError(err)
			permissionSetStorage, err = permissionSetPGStore.New(globaldb.GetRocksDB())
			utils.CrashOnError(err)
			accessScopeStorage, err = simpleAccessScopeStore.New(globaldb.GetRocksDB())
			utils.CrashOnError(err)
		}
		// Which role format is used is determined solely by the feature flag.
		ds = New(roleStorage, permissionSetStorage, accessScopeStorage, groupFilter.GetFiltered)

		ctx := sac.WithGlobalAccessScopeChecker(context.Background(),
			sac.AllowFixedScopes(
				sac.AccessModeScopeKeys(storage.Access_READ_ACCESS, storage.Access_READ_WRITE_ACCESS),
				sac.ResourceScopeKeys(resources.Access)))
		roles, permissionSets, accessScopes := getDefaultObjects()
		utils.Must(roleStorage.UpsertMany(ctx, roles))
		utils.Must(permissionSetStorage.UpsertMany(ctx, permissionSets))
		utils.Must(accessScopeStorage.UpsertMany(ctx, accessScopes))
	})
	return ds
}

type roleAttributes struct {
	idSuffix           string
	postgresID         string // postgresID should be populated with valid UUID values.
	description        string
	resourceWithAccess []permissions.ResourceWithAccess
}

func (attributes *roleAttributes) getID() string {
	if env.PostgresDatastoreEnabled.BooleanSetting() {
		return attributes.postgresID
	}
	return rolePkg.EnsureValidPermissionSetID(attributes.idSuffix)
}

var defaultRoles = map[string]roleAttributes{
	accesscontrol.Admin: {
		idSuffix:           "admin",
		postgresID:         accesscontrol.DefaultPermissionSetIDs[accesscontrol.Admin],
		description:        "For users: use it to provide read and write access to all the resources",
		resourceWithAccess: resources.AllResourcesModifyPermissions(),
	},
	accesscontrol.Analyst: {
		idSuffix:           "analyst",
		postgresID:         accesscontrol.DefaultPermissionSetIDs[accesscontrol.Analyst],
		resourceWithAccess: rolePkg.GetAnalystPermissions(),
		description:        "For users: use it to give read-only access to all the resources",
	},
	accesscontrol.ContinuousIntegration: {
		idSuffix:    "continuousintegration",
		postgresID:  accesscontrol.DefaultPermissionSetIDs[accesscontrol.ContinuousIntegration],
		description: "For automation: it includes the permissions required to enforce deployment policies",
		resourceWithAccess: []permissions.ResourceWithAccess{
			permissions.View(resources.Detection),
			permissions.Modify(resources.Image),
		},
	},
	accesscontrol.NetworkGraphViewer: {
		idSuffix:    "networkgraphviewer",
		postgresID:  accesscontrol.DefaultPermissionSetIDs[accesscontrol.NetworkGraphViewer],
		description: "For users: use it to give read-only access to the NetworkGraph pages",
		resourceWithAccess: []permissions.ResourceWithAccess{
			permissions.View(resources.Deployment),
			permissions.View(resources.NetworkGraph),
			permissions.View(resources.NetworkPolicy),
		},
	},
	accesscontrol.None: {
		idSuffix:    "none",
		postgresID:  accesscontrol.DefaultPermissionSetIDs[accesscontrol.None],
		description: "For users: use it to provide no read and write access to any resource",
	},
	accesscontrol.SensorCreator: {
		idSuffix:    "sensorcreator",
		postgresID:  accesscontrol.DefaultPermissionSetIDs[accesscontrol.SensorCreator],
		description: "For automation: it consists of the permissions to create Sensors in secured clusters",
		resourceWithAccess: []permissions.ResourceWithAccess{
			permissions.View(resources.Cluster),
			permissions.Modify(resources.Cluster),
			permissions.Modify(resources.Administration),
		},
	},
	accesscontrol.VulnMgmtApprover: {
		idSuffix:    "vulnmgmtapprover",
		postgresID:  accesscontrol.DefaultPermissionSetIDs[accesscontrol.VulnMgmtApprover],
		description: "For users: use it to provide access to approve vulnerability deferrals or false positive requests",
		resourceWithAccess: []permissions.ResourceWithAccess{
			permissions.View(resources.VulnerabilityManagementApprovals),
			permissions.Modify(resources.VulnerabilityManagementApprovals),
		},
	},
	accesscontrol.VulnMgmtRequester: {
		idSuffix:    "vulnmgmtrequester",
		postgresID:  accesscontrol.DefaultPermissionSetIDs[accesscontrol.VulnMgmtRequester],
		description: "For users: use it to provide access to request vulnerability deferrals or false positives",
		resourceWithAccess: []permissions.ResourceWithAccess{
			permissions.View(resources.VulnerabilityManagementRequests),
			permissions.Modify(resources.VulnerabilityManagementRequests),
		},
	},
	// TODO ROX-13888 when we migrate to WorkflowAdministration we can remove VulnerabilityReports and Role resources
	accesscontrol.VulnReporter: {
		idSuffix:    "vulnreporter",
		postgresID:  accesscontrol.DefaultPermissionSetIDs[accesscontrol.VulnReporter],
		description: "For users: use it to create and manage vulnerability reporting configurations for scheduled vulnerability reports",
		resourceWithAccess: func() []permissions.ResourceWithAccess {
			if !env.PostgresDatastoreEnabled.BooleanSetting() {
				return []permissions.ResourceWithAccess{
					permissions.View(resources.Access),                 // required for scopes
					permissions.View(resources.Integration),            // required for vuln report configurations
					permissions.View(resources.VulnerabilityReports),   // required for vuln report configurations prior to collections
					permissions.Modify(resources.VulnerabilityReports), // required for vuln report configurations prior to collections
				}
			}
			return []permissions.ResourceWithAccess{
				permissions.View(resources.WorkflowAdministration),   // required for vuln report configurations
				permissions.Modify(resources.WorkflowAdministration), // required for vuln report configurations
				permissions.View(resources.Integration),              // required for vuln report configurations
			}
		}(),
	},
	accesscontrol.VulnerabilityManager: {
		idSuffix:    "vulnmgmt",
		postgresID:  accesscontrol.DefaultPermissionSetIDs[accesscontrol.VulnerabilityManager],
		description: "For users: use it to provide access to analyze and manage system vulnerabilities",
		resourceWithAccess: []permissions.ResourceWithAccess{
			permissions.View(resources.Cluster),
			permissions.View(resources.Node),
			permissions.View(resources.Namespace),
			permissions.View(resources.Deployment),
			permissions.View(resources.Image),
			permissions.View(resources.Integration),
			permissions.Modify(resources.WatchedImage),
			permissions.Modify(resources.VulnerabilityManagementRequests),
			permissions.Modify(resources.VulnerabilityReports),
			permissions.Modify(resources.WorkflowAdministration),
		},
	},
}

func getDefaultObjects() ([]*storage.Role, []*storage.PermissionSet, []*storage.SimpleAccessScope) {
	roles := make([]*storage.Role, 0, len(defaultRoles))
	permissionSets := make([]*storage.PermissionSet, 0, len(defaultRoles))

	for roleName, attributes := range defaultRoles {
		resourceToAccess := permissionsUtils.FromResourcesWithAccess(attributes.resourceWithAccess...)

		role := &storage.Role{
			Name:          roleName,
			Description:   attributes.description,
			AccessScopeId: rolePkg.AccessScopeIncludeAll.GetId(),
			Traits: &storage.Traits{
				Origin: storage.Traits_DEFAULT,
			},
		}

		permissionSet := &storage.PermissionSet{
			Id:               attributes.getID(),
			Name:             role.Name,
			Description:      role.Description,
			ResourceToAccess: resourceToAccess,
			Traits: &storage.Traits{
				Origin: storage.Traits_DEFAULT,
			},
		}
		role.PermissionSetId = permissionSet.Id
		permissionSets = append(permissionSets, permissionSet)

		roles = append(roles, role)

	}
	simpleAccessScopes := []*storage.SimpleAccessScope{
		rolePkg.AccessScopeIncludeAll,
		rolePkg.AccessScopeExcludeAll}

	return roles, permissionSets, simpleAccessScopes
}
