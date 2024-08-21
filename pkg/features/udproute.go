package features

// -----------------------------------------------------------------------------
// Features - UDPRoute Conformance (Core)
// -----------------------------------------------------------------------------

const (
	// This option indicates support for UDPRoute
	SupportUDPRoute FeatureName = "UDPRoute"
)

var (
	UDPRouteFeature = Feature{
		Name:   SupportUDPRoute,
		Status: FeatureStatusTrial,
	}
)

// UDPRouteCoreFeatures includes all SupportedFeatures needed to be conformant with
// the UDPRoute resource.
var UDPRouteFeatures = map[FeatureName]Feature{
	SupportUDPRoute: UDPRouteFeature,
}
