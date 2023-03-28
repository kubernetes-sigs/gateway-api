package suite

import "k8s.io/apimachinery/pkg/util/sets"

// SupportedFeature allows opting in to additional conformance tests at an
// individual feature granularity.
type SupportedFeature string

const (
	// This option indicates support for ReferenceGrant (core conformance).
	// Opting out of this requires an implementation to have clearly implemented
	// and documented equivalent safeguards.
	SupportReferenceGrant SupportedFeature = "ReferenceGrant"

	// This option indicates support for TLSRoute (extended conformance).
	SupportTLSRoute SupportedFeature = "TLSRoute"

	// This option indicates support for HTTPRoute query param matching (extended conformance).
	SupportHTTPRouteQueryParamMatching SupportedFeature = "HTTPRouteQueryParamMatching"

	// This option indicates support for HTTPRoute method matching (extended conformance).
	SupportHTTPRouteMethodMatching SupportedFeature = "HTTPRouteMethodMatching"

	// This option indicates support for HTTPRoute response header modification (extended conformance).
	SupportHTTPResponseHeaderModification SupportedFeature = "HTTPResponseHeaderModification"

	// This option indicates support for Destination Port matching (extended conformance).
	SupportRouteDestinationPortMatching SupportedFeature = "RouteDestinationPortMatching"

	// This option indicates GatewayClass will update the observedGeneration in it's conditions when reconciling
	SupportGatewayClassObservedGenerationBump SupportedFeature = "GatewayClassObservedGenerationBump"

	// This option indicates support for HTTPRoute port redirect (extended conformance).
	SupportHTTPRoutePortRedirect SupportedFeature = "HTTPRoutePortRedirect"

	// This option indicates support for HTTPRoute scheme redirect (extended conformance).
	SupportHTTPRouteSchemeRedirect SupportedFeature = "HTTPRouteSchemeRedirect"

	// This option indicates support for HTTPRoute path redirect (experimental conformance).
	SupportHTTPRoutePathRedirect SupportedFeature = "HTTPRoutePathRedirect"

	// This option indicates support for HTTPRoute host rewrite (experimental conformance)
	SupportHTTPRouteHostRewrite SupportedFeature = "HTTPRouteHostRewrite"

	// This option indicates support for HTTPRoute path rewrite (experimental conformance)
	SupportHTTPRoutePathRewrite SupportedFeature = "HTTPRoutePathRewrite"
)

// StandardCoreFeatures are the features that are required to be conformant with
// the Core API features that are part of the Standard release channel.
var StandardCoreFeatures = sets.New(
	SupportReferenceGrant,
)

// AllFeatures contains all the supported features and can be used to run all
// conformance tests with `all-features` flag.
//
// Note that the AllFeatures must in sync with defined features when the
// feature constants change.
var AllFeatures = sets.New(
	SupportReferenceGrant,
	SupportTLSRoute,
	SupportHTTPRouteQueryParamMatching,
	SupportHTTPRouteMethodMatching,
	SupportHTTPResponseHeaderModification,
	SupportRouteDestinationPortMatching,
	SupportGatewayClassObservedGenerationBump,
	SupportHTTPRoutePortRedirect,
	SupportHTTPRouteSchemeRedirect,
	SupportHTTPRoutePathRedirect,
	SupportHTTPRouteHostRewrite,
	SupportHTTPRoutePathRewrite,
)
