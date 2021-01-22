package dackbox

import (
	clusterDackBox "github.com/stackrox/rox/central/cluster/dackbox"
	cveDackBox "github.com/stackrox/rox/central/cve/dackbox"
	componentDackBox "github.com/stackrox/rox/central/imagecomponent/dackbox"
	nodeDackBox "github.com/stackrox/rox/central/node/dackbox"
	v1 "github.com/stackrox/rox/generated/api/v1"
	"github.com/stackrox/rox/pkg/dackbox/keys/transformation"
)

var (
	// NodeTransformations holds the transformations to go from a node id to the ids of the given category.
	NodeTransformations = map[v1.SearchCategory]transformation.OneToMany{
		// Node (backwards) Clusters
		v1.SearchCategory_CLUSTERS: transformation.AddPrefix(nodeDackBox.Bucket).
			ThenMapToMany(transformation.BackwardFromContext()).
			Then(transformation.Dedupe()).
			Then(transformation.HasPrefix(clusterDackBox.Bucket)).
			ThenMapEachToOne(transformation.StripPrefix(clusterDackBox.Bucket)),

		// Node
		v1.SearchCategory_NODES: DoNothing,

		// Combine ( { k1, k2 }
		//          Node,
		//          Node (forwards) Components (forwards) CVEs,
		//          )
		v1.SearchCategory_NODE_VULN_EDGE: transformation.ForwardEdgeKeys(
			DoNothing,
			transformation.AddPrefix(nodeDackBox.Bucket).
				ThenMapToMany(transformation.ForwardFromContext()).
				Then(transformation.HasPrefix(componentDackBox.Bucket)).
				ThenMapEachToMany(transformation.ForwardFromContext()).
				Then(transformation.Dedupe()).
				Then(transformation.HasPrefix(cveDackBox.Bucket)).
				ThenMapEachToOne(transformation.StripPrefix(cveDackBox.Bucket)),
		),

		// Combine ( { k1, k2 }
		//          Node,
		//          Node (forwards) Components,
		//          )
		v1.SearchCategory_NODE_COMPONENT_EDGE: transformation.ForwardEdgeKeys(
			DoNothing,
			transformation.AddPrefix(nodeDackBox.Bucket).
				ThenMapToMany(transformation.ForwardFromContext()).
				Then(transformation.HasPrefix(componentDackBox.Bucket)).
				ThenMapEachToOne(transformation.StripPrefix(componentDackBox.Bucket)),
		),

		// Node (forwards) Components
		v1.SearchCategory_IMAGE_COMPONENTS: transformation.AddPrefix(nodeDackBox.Bucket).
			ThenMapToMany(transformation.ForwardFromContext()).
			Then(transformation.HasPrefix(componentDackBox.Bucket)).
			ThenMapEachToOne(transformation.StripPrefix(componentDackBox.Bucket)),

		// Combine ( { k1, k2 }
		//          Node (forwards) Components,
		//          Components (forwards) CVEs,
		//          )
		v1.SearchCategory_COMPONENT_VULN_EDGE: transformation.ForwardEdgeKeys(
			transformation.AddPrefix(nodeDackBox.Bucket).
				ThenMapToMany(transformation.ForwardFromContext()).
				Then(transformation.Dedupe()).
				Then(transformation.HasPrefix(componentDackBox.Bucket)).
				ThenMapEachToOne(transformation.StripPrefix(componentDackBox.Bucket)),
			transformation.AddPrefix(componentDackBox.Bucket).
				ThenMapToMany(transformation.ForwardFromContext()).
				Then(transformation.HasPrefix(cveDackBox.Bucket)).
				ThenMapEachToOne(transformation.StripPrefix(cveDackBox.Bucket)),
		),

		// We don't want to surface cluster level CVEs from a node scope, so we just descend to the CVEs.
		// Node (forwards) Components (forwards) CVEs
		v1.SearchCategory_VULNERABILITIES: transformation.AddPrefix(nodeDackBox.Bucket).
			ThenMapToMany(transformation.ForwardFromContext()).
			Then(transformation.HasPrefix(componentDackBox.Bucket)).
			ThenMapEachToMany(transformation.ForwardFromContext()).
			Then(transformation.Dedupe()).
			Then(transformation.HasPrefix(cveDackBox.Bucket)).
			ThenMapEachToOne(transformation.StripPrefix(cveDackBox.Bucket)),

		// We don't want to surface cluster level CVEs from a node scope.
		v1.SearchCategory_CLUSTER_VULN_EDGE: ReturnNothing,

		v1.SearchCategory_NAMESPACES:           ReturnNothing,
		v1.SearchCategory_DEPLOYMENTS:          ReturnNothing,
		v1.SearchCategory_IMAGES:               ReturnNothing,
		v1.SearchCategory_IMAGE_COMPONENT_EDGE: ReturnNothing,
		v1.SearchCategory_IMAGE_VULN_EDGE:      ReturnNothing,
	}
)
