/*
Copyright 2022 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package suite

import "k8s.io/apimachinery/pkg/util/sets"

// -----------------------------------------------------------------------------
// Features - Types
// -----------------------------------------------------------------------------

// SupportedFeature allows opting in to additional conformance tests at an
// individual feature granularity.
type SupportedFeature string

// -----------------------------------------------------------------------------
// Features - All Profiles (Core)
// -----------------------------------------------------------------------------

const (
	// This option indicates support for ReferenceGrant (core conformance).
	// Opting out of this requires an implementation to have clearly implemented
	// and documented equivalent safeguards.
	SupportReferenceGrant SupportedFeature = "ReferenceGrant"
)

// StandardCoreFeatures are the features that are required to be conformant with
// the Core API features. All conformance profiles include this standard set.
//
// TODO: we need clarity for standard vs experimental features.
// See: https://github.com/kubernetes-sigs/gateway-api/issues/1891
var StandardCoreFeatures = sets.New(
	SupportReferenceGrant,
)

// -----------------------------------------------------------------------------
// Features - All Profiles (Extended)
// -----------------------------------------------------------------------------

const (
	// This option indicates GatewayClass will update the observedGeneration in
	// it's conditions when reconciling (extended conformance).
	//
	// NOTE: we intend to make this core and require implementations to do it
	//       as we expect this is something every implementation should be able
	//       to do and it's ideal behavior.
	//
	//       See: https://github.com/kubernetes-sigs/gateway-api/issues/1780
	SupportGatewayClassObservedGenerationBump SupportedFeature = "GatewayClassObservedGenerationBump"

	// This option indicates support for Destination Port matching.
	SupportRouteDestinationPortMatching SupportedFeature = "RouteDestinationPortMatching"
)

// StandardExtendedFeatures are extra generic features that implementations may
// choose to support as an opt-in.
//
// TODO: we need clarity for standard vs experimental features.
// See: https://github.com/kubernetes-sigs/gateway-api/issues/1891
var StandardExtendedFeatures = sets.New(
	SupportGatewayClassObservedGenerationBump,
	SupportRouteDestinationPortMatching,
).Insert(StandardCoreFeatures.UnsortedList()...)

// -----------------------------------------------------------------------------
// Features - HTTP Conformance Profile (Core)
// -----------------------------------------------------------------------------

const (
	// This option indicates support for HTTPRoute
	SupportHTTPRoute SupportedFeature = "HTTPRoute"
)

// HTTPCoreFeatures includes all SupportedFeatures needed to be conformant with
// the HTTP conformance profile at a Core level of support.
var HTTPCoreFeatures = sets.New(
	SupportHTTPRoute,
)

// -----------------------------------------------------------------------------
// Features - HTTP Conformance Profile (Extended)
// -----------------------------------------------------------------------------

const (
	// This option indicates support for HTTPRoute query param matching (extended conformance).
	SupportHTTPRouteQueryParamMatching SupportedFeature = "HTTPRouteQueryParamMatching"

	// This option indicates support for HTTPRoute method matching (extended conformance).
	SupportHTTPRouteMethodMatching SupportedFeature = "HTTPRouteMethodMatching"

	// This option indicates support for HTTPRoute response header modification (extended conformance).
	SupportHTTPResponseHeaderModification SupportedFeature = "HTTPResponseHeaderModification"

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

// HTTPExtendedFeatures includes all the supported features for the HTTP conformance
// profile and can be used to opt-in to run all HTTP tests (including extended
// and experimental features).
var HTTPExtendedFeatures = sets.New(
	SupportHTTPRouteQueryParamMatching,
	SupportHTTPRouteMethodMatching,
	SupportHTTPResponseHeaderModification,
	SupportHTTPRoutePortRedirect,
	SupportHTTPRouteSchemeRedirect,
	SupportHTTPRoutePathRedirect,
	SupportHTTPRouteHostRewrite,
	SupportHTTPRoutePathRewrite,
).Insert(HTTPCoreFeatures.UnsortedList()...)

// -----------------------------------------------------------------------------
// Features - TLS Conformance Profile
// -----------------------------------------------------------------------------

const (
	// This option indicates support for TLSRoute
	SupportTLSRoute SupportedFeature = "TLSRoute"
)

// TLSCoreFeatures includes all the supported features for the TLS conformance
// profile at a Core level of support.
var TLSCoreFeatures = sets.New(
	SupportTLSRoute,
)

// -----------------------------------------------------------------------------
// Features - Compilations
// -----------------------------------------------------------------------------

// AllFeatures contains all the supported features and can be used to run all
// conformance tests with `all-features` flag.
//
// NOTE: as new profiles and levels are added, they should be inserted into
// this set.
var AllFeatures = sets.New[SupportedFeature]().
	Insert(StandardExtendedFeatures.UnsortedList()...).
	Insert(HTTPExtendedFeatures.UnsortedList()...).
	Insert(TLSCoreFeatures.UnsortedList()...)
