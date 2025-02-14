package fetcher

import (
	"context"
	"encoding/json"
	"os"
	"testing"

	"github.com/facebookincubator/nvdtools/cvefeed/nvd/schema"
	"github.com/golang/mock/gomock"
	mockClusterDataStore "github.com/stackrox/rox/central/cluster/datastore/mocks"
	mockClusterEdgeDataStore "github.com/stackrox/rox/central/clustercveedge/datastore/mocks"
	"github.com/stackrox/rox/central/cve/converter"
	"github.com/stackrox/rox/central/cve/converter/utils"
	mockCVEDataStore "github.com/stackrox/rox/central/cve/datastore/mocks"
	"github.com/stackrox/rox/central/cve/matcher"
	mockImageDataStore "github.com/stackrox/rox/central/image/datastore/mocks"
	mockNSDataStore "github.com/stackrox/rox/central/namespace/datastore/mocks"
	"github.com/stackrox/rox/generated/storage"
	"github.com/stackrox/rox/pkg/dackbox/edges"
	"github.com/stackrox/rox/pkg/env"
	"github.com/stackrox/rox/pkg/kubernetes"
	"github.com/stackrox/rox/pkg/scanners/types"
	"github.com/stackrox/rox/pkg/search"
	"github.com/stackrox/rox/pkg/set"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	cveListPath  = "testdata/cve-list.json"
	checksumPath = "testdata/checksum"
)

func TestJSONUnmarshalCVEs(t *testing.T) {
	data, err := os.ReadFile(cveListPath)
	require.NoError(t, err)
	var cveEntries []schema.NVDCVEFeedJSON10DefCVEItem
	assert.NoError(t, json.Unmarshal(data, &cveEntries))
	assert.Len(t, cveEntries, 22)
}

func TestReadChecksum(t *testing.T) {
	data, err := os.ReadFile(checksumPath)
	require.Nil(t, err)
	assert.Equal(t, "9c210f795cdc93a458f06053033793e2907a75b28ec0aaeadd22b6d2e70f0e5e", string(data))
}

func TestReconcileCVEsInDB(t *testing.T) {
	if env.PostgresDatastoreEnabled.BooleanSetting() {
		t.Skip("Skip non-postgres test")
	}

	cluster := &storage.Cluster{
		Id:   "test_cluster_id1",
		Name: "cluster1",
		Status: &storage.ClusterStatus{
			OrchestratorMetadata: &storage.OrchestratorMetadata{
				Version: "v1.10.6",
			},
		},
	}

	nvdCVEs := []*schema.NVDCVEFeedJSON10DefCVEItem{
		{
			CVE: &schema.CVEJSON40{
				CVEDataMeta: &schema.CVEJSON40CVEDataMeta{
					ID: "CVE-1",
				},
			},
			Configurations: &schema.NVDCVEFeedJSON10DefConfigurations{
				Nodes: []*schema.NVDCVEFeedJSON10DefNode{
					{
						Operator: "OR",
						CPEMatch: []*schema.NVDCVEFeedJSON10DefCPEMatch{
							{
								Vulnerable:            true,
								Cpe23Uri:              "cpe:2.3:a:kubernetes:kubernetes:*:*:*:*:*:*:*:*",
								VersionStartIncluding: "1.10.1",
								VersionEndExcluding:   "1.10.9",
							},
						},
					},
				},
			},
			Impact: &schema.NVDCVEFeedJSON10DefImpact{
				BaseMetricV3: &schema.NVDCVEFeedJSON10DefImpactBaseMetricV3{
					CVSSV3: &schema.CVSSV30{
						BaseScore:    6.1,
						VectorString: "AV:L/AC:L/PR:L/UI:N/S:U/C:N/I:L/A:H",
						Version:      "3.0",
					},
					ExploitabilityScore: 1.8,
					ImpactScore:         4.2,
				},
			},
		},
		{
			CVE: &schema.CVEJSON40{
				CVEDataMeta: &schema.CVEJSON40CVEDataMeta{
					ID: "CVE-2",
				},
			},
			Configurations: &schema.NVDCVEFeedJSON10DefConfigurations{
				Nodes: []*schema.NVDCVEFeedJSON10DefNode{
					{
						Operator: "OR",
						CPEMatch: []*schema.NVDCVEFeedJSON10DefCPEMatch{
							{
								Vulnerable: true,
								Cpe23Uri:   "cpe:2.3:a:kubernetes:kubernetes:1.10.6:*:*:*:*:*:*:*",
							},
						},
					},
				},
			},
			Impact: &schema.NVDCVEFeedJSON10DefImpact{
				BaseMetricV3: &schema.NVDCVEFeedJSON10DefImpactBaseMetricV3{
					CVSSV3: &schema.CVSSV30{
						BaseScore:    6.1,
						VectorString: "AV:L/AC:L/PR:L/UI:N/S:U/C:N/I:L/A:H",
						Version:      "3.0",
					},
					ExploitabilityScore: 1.8,
					ImpactScore:         4.2,
				},
			},
		},
		{
			CVE: &schema.CVEJSON40{
				CVEDataMeta: &schema.CVEJSON40CVEDataMeta{
					ID: "CVE-3",
				},
			},
			Configurations: &schema.NVDCVEFeedJSON10DefConfigurations{
				Nodes: []*schema.NVDCVEFeedJSON10DefNode{
					{
						Operator: "OR",
						CPEMatch: []*schema.NVDCVEFeedJSON10DefCPEMatch{
							{
								Vulnerable:            true,
								Cpe23Uri:              "cpe:2.3:a:kubernetes:kubernetes:*:*:*:*:*:*:*:*",
								VersionStartIncluding: "1.10.3",
								VersionEndIncluding:   "1.10.7",
							},
						},
					},
				},
			},
			Impact: &schema.NVDCVEFeedJSON10DefImpact{
				BaseMetricV3: &schema.NVDCVEFeedJSON10DefImpactBaseMetricV3{
					CVSSV3: &schema.CVSSV30{
						BaseScore:    6.1,
						VectorString: "AV:L/AC:L/PR:L/UI:N/S:U/C:N/I:L/A:H",
						Version:      "3.0",
					},
					ExploitabilityScore: 1.8,
					ImpactScore:         4.2,
				},
			},
		},
	}

	embeddedCVEs, err := utils.NVDCVEsToEmbeddedCVEs(nvdCVEs, utils.K8s)
	require.NoError(t, err)

	embeddedCVEToClusters := map[string][]*storage.Cluster{
		"CVE-1": {
			cluster,
		},
		"CVE-2": {
			cluster,
		},
		"CVE-3": {
			cluster,
		},
	}

	cvesToUpsert := []converter.ClusterCVEParts{
		{
			CVE: &storage.CVE{
				Id:           "CVE-1",
				Link:         "https://nvd.nist.gov/vuln/detail/CVE-1",
				Type:         storage.CVE_K8S_CVE,
				Cvss:         6.1,
				ImpactScore:  4.2,
				ScoreVersion: storage.CVE_V3,
				CvssV3: &storage.CVSSV3{
					Vector:              "AV:L/AC:L/PR:L/UI:N/S:U/C:N/I:L/A:H",
					ExploitabilityScore: 1.8,
					ImpactScore:         4.2,
					AttackVector:        storage.CVSSV3_ATTACK_LOCAL,
					AttackComplexity:    storage.CVSSV3_COMPLEXITY_LOW,
					PrivilegesRequired:  storage.CVSSV3_PRIVILEGE_LOW,
					UserInteraction:     storage.CVSSV3_UI_NONE,
					Scope:               storage.CVSSV3_UNCHANGED,
					Confidentiality:     storage.CVSSV3_IMPACT_NONE,
					Integrity:           storage.CVSSV3_IMPACT_LOW,
					Availability:        storage.CVSSV3_IMPACT_HIGH,
					Score:               6.1,
				},
			},
			Children: []converter.EdgeParts{
				{
					Edge: &storage.ClusterCVEEdge{
						Id:        edges.EdgeID{ParentID: "test_cluster_id1", ChildID: "CVE-1"}.ToString(),
						IsFixable: true,
						HasFixedBy: &storage.ClusterCVEEdge_FixedBy{
							FixedBy: "1.10.9",
						},
					},
					ClusterID: "test_cluster_id1",
				},
			},
		},
		{
			CVE: &storage.CVE{
				Id:           "CVE-2",
				Link:         "https://nvd.nist.gov/vuln/detail/CVE-2",
				Type:         storage.CVE_K8S_CVE,
				Cvss:         6.1,
				ImpactScore:  4.2,
				ScoreVersion: storage.CVE_V3,
				CvssV3: &storage.CVSSV3{
					Vector:              "AV:L/AC:L/PR:L/UI:N/S:U/C:N/I:L/A:H",
					ExploitabilityScore: 1.8,
					ImpactScore:         4.2,
					AttackVector:        storage.CVSSV3_ATTACK_LOCAL,
					AttackComplexity:    storage.CVSSV3_COMPLEXITY_LOW,
					PrivilegesRequired:  storage.CVSSV3_PRIVILEGE_LOW,
					UserInteraction:     storage.CVSSV3_UI_NONE,
					Scope:               storage.CVSSV3_UNCHANGED,
					Confidentiality:     storage.CVSSV3_IMPACT_NONE,
					Integrity:           storage.CVSSV3_IMPACT_LOW,
					Availability:        storage.CVSSV3_IMPACT_HIGH,
					Score:               6.1,
				},
			},
			Children: []converter.EdgeParts{
				{
					Edge: &storage.ClusterCVEEdge{
						Id:        edges.EdgeID{ParentID: "test_cluster_id1", ChildID: "CVE-2"}.ToString(),
						IsFixable: false,
					},
					ClusterID: "test_cluster_id1",
				},
			},
		},
		{
			CVE: &storage.CVE{
				Id:           "CVE-3",
				Link:         "https://nvd.nist.gov/vuln/detail/CVE-3",
				Type:         storage.CVE_K8S_CVE,
				Cvss:         6.1,
				ImpactScore:  4.2,
				ScoreVersion: storage.CVE_V3,
				CvssV3: &storage.CVSSV3{
					Vector:              "AV:L/AC:L/PR:L/UI:N/S:U/C:N/I:L/A:H",
					ExploitabilityScore: 1.8,
					ImpactScore:         4.2,
					AttackVector:        storage.CVSSV3_ATTACK_LOCAL,
					AttackComplexity:    storage.CVSSV3_COMPLEXITY_LOW,
					PrivilegesRequired:  storage.CVSSV3_PRIVILEGE_LOW,
					UserInteraction:     storage.CVSSV3_UI_NONE,
					Scope:               storage.CVSSV3_UNCHANGED,
					Confidentiality:     storage.CVSSV3_IMPACT_NONE,
					Integrity:           storage.CVSSV3_IMPACT_LOW,
					Availability:        storage.CVSSV3_IMPACT_HIGH,
					Score:               6.1,
				},
			},
			Children: []converter.EdgeParts{
				{
					Edge: &storage.ClusterCVEEdge{
						Id:        edges.EdgeID{ParentID: "test_cluster_id1", ChildID: "CVE-3"}.ToString(),
						IsFixable: false,
					},
					ClusterID: "test_cluster_id1",
				},
			},
		},
	}

	ctrl := gomock.NewController(t)
	mockClusters := mockClusterDataStore.NewMockDataStore(ctrl)
	mockClusterCveEdge := mockClusterEdgeDataStore.NewMockDataStore(ctrl)
	mockNamespaces := mockNSDataStore.NewMockDataStore(ctrl)
	mockImages := mockImageDataStore.NewMockDataStore(ctrl)
	mockCVEs := mockCVEDataStore.NewMockDataStore(ctrl)

	cveMatcher, err := matcher.NewCVEMatcher(mockClusters, mockNamespaces, mockImages)
	require.NoError(t, err)

	cveManager := &orchestratorIstioCVEManagerImpl{
		orchestratorCVEMgr: &orchestratorCVEManager{
			clusterCVEEdgeDataStore: mockClusterCveEdge,
			clusterDataStore:        mockClusters,
			legacyCVEDataStore:      mockCVEs,
			cveMatcher:              cveMatcher,
		},
	}

	mockCVEs.EXPECT().Search(gomock.Any(), gomock.Any()).Return(nil, nil)
	mockClusters.EXPECT().GetClusters(gomock.Any()).Return([]*storage.Cluster{cluster}, nil).AnyTimes()
	mockNamespaces.EXPECT().Search(gomock.Any(), gomock.Any()).Return(nil, nil).AnyTimes()

	mockClusterCveEdge.EXPECT().Upsert(gomock.Any(), cvesToUpsert).Return(nil)

	mockClusterCveEdge.EXPECT().Search(gomock.Any(), gomock.Any()).Return([]search.Result{}, nil).AnyTimes()
	err = cveManager.orchestratorCVEMgr.updateCVEs(embeddedCVEs, embeddedCVEToClusters, utils.K8s)
	assert.NoError(t, err)
}

func TestOrchestratorManager_ReconcileCVEs(t *testing.T) {
	if env.PostgresDatastoreEnabled.BooleanSetting() {
		t.Skip("Skip non-postgres test")
	}

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClusters := mockClusterDataStore.NewMockDataStore(ctrl)
	mockClusterCveEdge := mockClusterEdgeDataStore.NewMockDataStore(ctrl)
	mockNamespaces := mockNSDataStore.NewMockDataStore(ctrl)
	mockImages := mockImageDataStore.NewMockDataStore(ctrl)
	mockCVEs := mockCVEDataStore.NewMockDataStore(ctrl)
	clusters := []*storage.Cluster{
		{
			Id:   "test_cluster_id1",
			Name: "cluster1",
			Status: &storage.ClusterStatus{
				OrchestratorMetadata: &storage.OrchestratorMetadata{
					Version: "v1.10.6",
				},
			},
		},
		{
			Id:   "test_cluster_id2",
			Name: "cluster2",
			Status: &storage.ClusterStatus{
				OrchestratorMetadata: &storage.OrchestratorMetadata{
					Version: "v1.10.9",
				},
			},
		},
		{
			Id:   "test_cluster_id3",
			Name: "cluster3",
			Status: &storage.ClusterStatus{
				OrchestratorMetadata: &storage.OrchestratorMetadata{
					Version: "v1.10.10",
					IsOpenshift: &storage.OrchestratorMetadata_OpenshiftVersion{
						OpenshiftVersion: "4.7.7",
					},
				},
			},
		},
	}

	mockCVEs.EXPECT().Search(gomock.Any(), gomock.Any()).Return(nil, nil).Times(4)
	mockClusterCveEdge.EXPECT().Search(gomock.Any(), gomock.Any()).Return([]search.Result{}, nil).Times(4)

	mockClusterCveEdge.EXPECT().Upsert(gomock.Any(), gomock.Any()).Return(nil).Times(1).Do(func(arg0 context.Context, cves ...converter.ClusterCVEParts) {
		assert.Equal(t, len(cves), 3)
		for _, cve := range cves {
			switch cve.CVE.GetId() {
			case "CVE-1":
				assert.Equal(t, len(cve.Children), 2) // Cluster 1, 2
				assert.Contains(t, []string{cve.Children[0].ClusterID, cve.Children[1].ClusterID}, clusters[0].GetId())
				assert.Contains(t, []string{cve.Children[0].ClusterID, cve.Children[1].ClusterID}, clusters[1].GetId())
			case "CVE-2":
				assert.Equal(t, len(cve.Children), 1) // Cluster 2
				assert.Equal(t, cve.Children[0].ClusterID, clusters[0].GetId())
			case "CVE-3":
				assert.Equal(t, len(cve.Children), 2) // Cluster 1, 2
				ss := set.StringSet{}
				ss.AddAll(cve.Children[0].ClusterID, cve.Children[1].ClusterID)
				assert.Equal(t, len(ss), 2)
			}
		}
	})

	cvesWithComponents := []*nvdCVEWithComponents{
		{
			nvdCVE: &schema.NVDCVEFeedJSON10DefCVEItem{
				CVE: &schema.CVEJSON40{
					CVEDataMeta: &schema.CVEJSON40CVEDataMeta{
						ID: "CVE-1",
					},
				},
				Configurations: &schema.NVDCVEFeedJSON10DefConfigurations{
					Nodes: []*schema.NVDCVEFeedJSON10DefNode{
						{
							Operator: "OR",
							CPEMatch: []*schema.NVDCVEFeedJSON10DefCPEMatch{
								{
									Vulnerable:            true,
									Cpe23Uri:              "cpe:2.3:a:kubernetes:kubernetes:*:*:*:*:*:*:*:*",
									VersionStartIncluding: "1.10.1",
									VersionEndExcluding:   "1.10.10",
								},
							},
						},
					},
				},
				Impact: &schema.NVDCVEFeedJSON10DefImpact{
					BaseMetricV3: &schema.NVDCVEFeedJSON10DefImpactBaseMetricV3{
						CVSSV3: &schema.CVSSV30{
							BaseScore:    6.1,
							VectorString: "AV:L/AC:L/PR:L/UI:N/S:U/C:N/I:L/A:H",
							Version:      "3.0",
						},
						ExploitabilityScore: 1.8,
						ImpactScore:         4.2,
					},
				},
			},
			components: []string{
				kubernetes.KubeControllerManager,
			},
		},
		{
			nvdCVE: &schema.NVDCVEFeedJSON10DefCVEItem{
				CVE: &schema.CVEJSON40{
					CVEDataMeta: &schema.CVEJSON40CVEDataMeta{
						ID: "CVE-2",
					},
				},
				Configurations: &schema.NVDCVEFeedJSON10DefConfigurations{
					Nodes: []*schema.NVDCVEFeedJSON10DefNode{
						{
							Operator: "OR",
							CPEMatch: []*schema.NVDCVEFeedJSON10DefCPEMatch{
								{
									Vulnerable: true,
									Cpe23Uri:   "cpe:2.3:a:kubernetes:kubernetes:1.10.6:*:*:*:*:*:*:*",
								},
							},
						},
					},
				},
				Impact: &schema.NVDCVEFeedJSON10DefImpact{
					BaseMetricV3: &schema.NVDCVEFeedJSON10DefImpactBaseMetricV3{
						CVSSV3: &schema.CVSSV30{
							BaseScore:    6.1,
							VectorString: "AV:L/AC:L/PR:L/UI:N/S:U/C:N/I:L/A:H",
							Version:      "3.0",
						},
						ExploitabilityScore: 1.8,
						ImpactScore:         4.2,
					},
				},
			},
			components: []string{
				kubernetes.KubeAPIServer,
				kubernetes.KubeControllerManager,
				kubernetes.KubeAggregator,
			},
		},
		{
			nvdCVE: &schema.NVDCVEFeedJSON10DefCVEItem{
				CVE: &schema.CVEJSON40{
					CVEDataMeta: &schema.CVEJSON40CVEDataMeta{
						ID: "CVE-3",
					},
				},
				Configurations: &schema.NVDCVEFeedJSON10DefConfigurations{
					Nodes: []*schema.NVDCVEFeedJSON10DefNode{
						{
							Operator: "OR",
							CPEMatch: []*schema.NVDCVEFeedJSON10DefCPEMatch{
								{
									Vulnerable:            true,
									Cpe23Uri:              "cpe:2.3:a:kubernetes:kubernetes:*:*:*:*:*:*:*:*",
									VersionStartIncluding: "1.10.3",
									VersionEndIncluding:   "1.10.11",
								},
							},
						},
					},
				},
				Impact: &schema.NVDCVEFeedJSON10DefImpact{
					BaseMetricV3: &schema.NVDCVEFeedJSON10DefImpactBaseMetricV3{
						CVSSV3: &schema.CVSSV30{
							BaseScore:    6.1,
							VectorString: "AV:L/AC:L/PR:L/UI:N/S:U/C:N/I:L/A:H",
							Version:      "3.0",
						},
						ExploitabilityScore: 1.8,
						ImpactScore:         4.2,
					},
				},
			},
			components: []string{
				kubernetes.KubeAggregator,
			},
		},
		{
			nvdCVE: &schema.NVDCVEFeedJSON10DefCVEItem{
				CVE: &schema.CVEJSON40{
					CVEDataMeta: &schema.CVEJSON40CVEDataMeta{
						ID: "CVE-4",
					},
				},
				Configurations: &schema.NVDCVEFeedJSON10DefConfigurations{
					Nodes: []*schema.NVDCVEFeedJSON10DefNode{
						{
							Operator: "OR",
							CPEMatch: []*schema.NVDCVEFeedJSON10DefCPEMatch{
								{
									Vulnerable:            true,
									Cpe23Uri:              "cpe:2.3:a:openshift:openshift:*:*:*:*:*:*:*:*",
									VersionStartIncluding: "4.7.3",
									VersionEndIncluding:   "4.7.10",
								},
							},
						},
					},
				},
				Impact: &schema.NVDCVEFeedJSON10DefImpact{
					BaseMetricV3: &schema.NVDCVEFeedJSON10DefImpactBaseMetricV3{
						CVSSV3: &schema.CVSSV30{
							BaseScore:    6.1,
							VectorString: "AV:L/AC:L/PR:L/UI:N/S:U/C:N/I:L/A:H",
							Version:      "3.0",
						},
						ExploitabilityScore: 1.8,
						ImpactScore:         4.2,
					},
				},
			},
			components: []string{
				"openshift",
			},
		},
		{
			nvdCVE: &schema.NVDCVEFeedJSON10DefCVEItem{
				CVE: &schema.CVEJSON40{
					CVEDataMeta: &schema.CVEJSON40CVEDataMeta{
						ID: "CVE-5",
					},
				},
				Configurations: &schema.NVDCVEFeedJSON10DefConfigurations{
					Nodes: []*schema.NVDCVEFeedJSON10DefNode{
						{
							Operator: "OR",
							CPEMatch: []*schema.NVDCVEFeedJSON10DefCPEMatch{
								{
									Vulnerable:            true,
									Cpe23Uri:              "cpe:2.3:a:openshift:openshift:*:*:*:*:*:*:*:*",
									VersionStartIncluding: "4.7.8",
									VersionEndIncluding:   "4.7.12",
								},
							},
						},
					},
				},
				Impact: &schema.NVDCVEFeedJSON10DefImpact{
					BaseMetricV3: &schema.NVDCVEFeedJSON10DefImpactBaseMetricV3{
						CVSSV3: &schema.CVSSV30{
							BaseScore:    6.1,
							VectorString: "AV:L/AC:L/PR:L/UI:N/S:U/C:N/I:L/A:H",
							Version:      "3.0",
						},
						ExploitabilityScore: 1.8,
						ImpactScore:         4.2,
					},
				},
			},
			components: []string{
				"openshift",
			},
		},
		{
			nvdCVE: &schema.NVDCVEFeedJSON10DefCVEItem{
				CVE: &schema.CVEJSON40{
					CVEDataMeta: &schema.CVEJSON40CVEDataMeta{
						ID: "CVE-6",
					},
				},
				Configurations: &schema.NVDCVEFeedJSON10DefConfigurations{
					Nodes: []*schema.NVDCVEFeedJSON10DefNode{
						{
							Operator: "OR",
							CPEMatch: []*schema.NVDCVEFeedJSON10DefCPEMatch{
								{
									Vulnerable:            true,
									Cpe23Uri:              "cpe:2.3:a:istio:istio:*:*:*:*:*:*:*:*",
									VersionStartIncluding: "1.13.12",
									VersionEndExcluding:   "1.13.19",
								},
							},
						},
					},
				},
				Impact: &schema.NVDCVEFeedJSON10DefImpact{
					BaseMetricV3: &schema.NVDCVEFeedJSON10DefImpactBaseMetricV3{
						CVSSV3: &schema.CVSSV30{
							BaseScore:    6.1,
							VectorString: "AV:L/AC:L/PR:L/UI:N/S:U/C:N/I:L/A:H",
							Version:      "3.0",
						},
						ExploitabilityScore: 1.8,
						ImpactScore:         4.2,
					},
				},
			},
			components: []string{
				"istio",
			}},
		{
			nvdCVE: &schema.NVDCVEFeedJSON10DefCVEItem{
				CVE: &schema.CVEJSON40{
					CVEDataMeta: &schema.CVEJSON40CVEDataMeta{
						ID: "CVE-7",
					},
				},
				Configurations: &schema.NVDCVEFeedJSON10DefConfigurations{
					Nodes: []*schema.NVDCVEFeedJSON10DefNode{
						{
							Operator: "OR",
							CPEMatch: []*schema.NVDCVEFeedJSON10DefCPEMatch{
								{
									Vulnerable:            true,
									Cpe23Uri:              "cpe:2.3:a:istio:istio:*:*:*:*:*:*:*:*",
									VersionStartIncluding: "1.13.13",
									VersionEndExcluding:   "1.13.18",
								},
							},
						},
					},
				},
				Impact: &schema.NVDCVEFeedJSON10DefImpact{
					BaseMetricV3: &schema.NVDCVEFeedJSON10DefImpactBaseMetricV3{
						CVSSV3: &schema.CVSSV30{
							BaseScore:    6.1,
							VectorString: "AV:L/AC:L/PR:L/UI:N/S:U/C:N/I:L/A:H",
							Version:      "3.0",
						},
						ExploitabilityScore: 1.8,
						ImpactScore:         4.2,
					},
				},
			},
			components: []string{
				"istio",
			}},
	}

	cveMatcher, err := matcher.NewCVEMatcher(mockClusters, mockNamespaces, mockImages)
	require.NoError(t, err)

	scanner := mockScanner{
		cveMatcher: cveMatcher,
		nvdCVEs:    cvesWithComponents,
	}

	orchestratorCVEMgr := &orchestratorCVEManager{
		clusterCVEEdgeDataStore: mockClusterCveEdge,
		clusterDataStore:        mockClusters,
		legacyCVEDataStore:      mockCVEs,
		cveMatcher:              cveMatcher,
		scanners:                make(map[string]types.OrchestratorScanner),
	}
	orchestratorCVEMgr.scanners["someName"] = &scanner

	err = orchestratorCVEMgr.reconcileCVEs(clusters, utils.K8s)
	assert.NoError(t, err)

	mockClusterCveEdge.EXPECT().Upsert(gomock.Any(), gomock.Any()).Return(nil).Times(1).Do(func(arg0 context.Context, cves ...converter.ClusterCVEParts) {
		assert.Equal(t, 1, len(cves))
		assert.Equal(t, "CVE-4", cves[0].CVE.GetId())
		assert.Equal(t, 1, len(cves[0].Children))
		assert.Contains(t, clusters[2].GetId(), cves[0].Children[0].ClusterID)
	})
	err = orchestratorCVEMgr.reconcileCVEs(clusters, utils.OpenShift)
	assert.NoError(t, err)

	mockClusterCveEdge.EXPECT().Upsert(gomock.Any(), gomock.Any()).Return(nil).Times(1).Do(func(arg0 context.Context, cves ...converter.ClusterCVEParts) {
		assert.Equal(t, 2, len(cves)) // CVE 1, 3
	})

	clusters = clusters[1:2]
	err = orchestratorCVEMgr.reconcileCVEs(clusters, utils.K8s)
	assert.NoError(t, err)

	mockClusterCveEdge.EXPECT().Upsert(gomock.Any(), gomock.Any()).Return(nil).Times(1).Do(func(arg0 context.Context, cves ...converter.ClusterCVEParts) {
		assert.Empty(t, cves)
	})

	err = orchestratorCVEMgr.reconcileCVEs(clusters, utils.OpenShift)
	assert.NoError(t, err)

	cves := []string{"to_be_removed_0", "to_be_removed_1", "CVE-1", "CVE-3"}
	var existingCVEs []search.Result
	for _, cve := range cves {
		existingCVEs = append(existingCVEs, search.Result{ID: cve})
	}
	mockCVEs.EXPECT().Search(gomock.Any(), gomock.Any()).Return(existingCVEs, nil)

	edges := []edges.EdgeID{
		{ParentID: "cluster1", ChildID: cves[0]},
		{ParentID: "cluster2", ChildID: cves[1]},
		{ParentID: "cluster1", ChildID: cves[1]},
		{ParentID: clusters[0].Id, ChildID: cves[2]},
		{ParentID: clusters[0].Id, ChildID: cves[3]},
	}

	var existingEdges []search.Result
	for _, edge := range edges {
		existingEdges = append(existingEdges, search.Result{ID: edge.ToString()})
	}
	mockClusterCveEdge.EXPECT().Search(gomock.Any(), gomock.Any()).Return(existingEdges, nil)
	mockClusterCveEdge.EXPECT().Upsert(gomock.Any(), gomock.Any()).Return(nil).Times(1).Do(func(arg0 context.Context, cves ...converter.ClusterCVEParts) {
		assert.Equal(t, 2, len(cves))
	})
	mockClusterCveEdge.EXPECT().Delete(gomock.Any(), gomock.Any()).Return(nil).Times(1).Do(func(arg0 context.Context, ids ...string) {
		assert.Equal(t, 3, len(ids))
	})
	mockCVEs.EXPECT().Delete(gomock.Any(), gomock.Any()).Return(nil).Times(1).Do(func(arg0 context.Context, ids ...string) {
		assert.Equal(t, 2, len(ids))
		assert.Contains(t, ids, cves[0])
		assert.Contains(t, ids, cves[1])
	})
	err = orchestratorCVEMgr.reconcileCVEs(clusters, utils.K8s)
	assert.NoError(t, err)

	imgName := &storage.ImageName{
		FullName: "istioTestImg",
		Tag:      "1.13.16",
	}
	img := &storage.Image{
		Name: imgName,
		Id:   "12345",
	}

	mockClusterCveEdge.EXPECT().Upsert(gomock.Any(), gomock.Any()).Return(nil).Times(1).Do(func(arg0 context.Context, cves ...converter.ClusterCVEParts) {
		assert.Equal(t, 2, len(cves))
	})
	mockImages.EXPECT().SearchRawImages(gomock.Any(), gomock.Any()).Return([]*storage.Image{img}, nil).Times(1)
	searchResults := make([]search.Result, 0, 1)
	searchResults = append(searchResults, search.Result{
		ID: "testID",
	})
	mockNamespaces.EXPECT().Search(gomock.Any(), gomock.Any()).Return(searchResults, nil)
	mockCVEs.EXPECT().Search(gomock.Any(), gomock.Any()).Return(nil, nil).Times(1)
	mockClusterCveEdge.EXPECT().Search(gomock.Any(), gomock.Any()).Return(existingEdges, nil)
	mockClusterCveEdge.EXPECT().Delete(gomock.Any(), gomock.Any()).Return(nil).Times(1)
	mockCVEs.EXPECT().Delete(gomock.Any(), gomock.Any()).Return(nil).Times(1)

	err = orchestratorCVEMgr.reconcileCVEs(clusters, utils.Istio)
	assert.NoError(t, err)
}
