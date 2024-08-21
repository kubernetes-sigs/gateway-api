/*
Copyright 2024 The Kubernetes Authors.

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

package features

import "k8s.io/apimachinery/pkg/util/sets"

// -----------------------------------------------------------------------------
// Features - HTTPRoute Conformance (Core)
// -----------------------------------------------------------------------------

const (
	// This option indicates support for HTTPRoute
	SupportHTTPRoute FeatureName = "HTTPRoute"
)

var (
	HTTPRouteFeature = Feature{
		Name:   SupportHTTPRoute,
		Status: FeatureStatusStable,
	}
)

// HTTPRouteCoreFeatures includes all SupportedFeatures needed to be conformant with
// the HTTPRoute resource.
var HTTPRouteCoreFeatures = sets.New(
	HTTPRouteFeature,
)

// -----------------------------------------------------------------------------
// Features - HTTPRoute Conformance (Extended)
// -----------------------------------------------------------------------------

const (
	// This option indicates support for HTTPRoute backend request header modification
	SupportHTTPRouteBackendRequestHeaderModification FeatureName = "HTTPRouteBackendRequestHeaderModification"

	// This option indicates support for HTTPRoute query param matching (extended conformance).
	SupportHTTPRouteQueryParamMatching FeatureName = "HTTPRouteQueryParamMatching"

	// This option indicates support for HTTPRoute method matching (extended conformance).
	SupportHTTPRouteMethodMatching FeatureName = "HTTPRouteMethodMatching"

	// This option indicates support for HTTPRoute response header modification (extended conformance).
	SupportHTTPRouteResponseHeaderModification FeatureName = "HTTPRouteResponseHeaderModification"

	// This option indicates support for HTTPRoute port redirect (extended conformance).
	SupportHTTPRoutePortRedirect FeatureName = "HTTPRoutePortRedirect"

	// This option indicates support for HTTPRoute scheme redirect (extended conformance).
	SupportHTTPRouteSchemeRedirect FeatureName = "HTTPRouteSchemeRedirect"

	// This option indicates support for HTTPRoute path redirect (extended conformance).
	SupportHTTPRoutePathRedirect FeatureName = "HTTPRoutePathRedirect"

	// This option indicates support for HTTPRoute host rewrite (extended conformance)
	SupportHTTPRouteHostRewrite FeatureName = "HTTPRouteHostRewrite"

	// This option indicates support for HTTPRoute path rewrite (extended conformance)
	SupportHTTPRoutePathRewrite FeatureName = "HTTPRoutePathRewrite"

	// This option indicates support for HTTPRoute request mirror (extended conformance).
	SupportHTTPRouteRequestMirror FeatureName = "HTTPRouteRequestMirror"

	// This option indicates support for multiple RequestMirror filters within the same HTTPRoute rule (extended conformance).
	SupportHTTPRouteRequestMultipleMirrors FeatureName = "HTTPRouteRequestMultipleMirrors"

	// This option indicates support for HTTPRoute request timeouts (extended conformance).
	SupportHTTPRouteRequestTimeout FeatureName = "HTTPRouteRequestTimeout"

	// This option indicates support for HTTPRoute backendRequest timeouts (extended conformance).
	SupportHTTPRouteBackendTimeout FeatureName = "HTTPRouteBackendTimeout"

	// This option indicates support for HTTPRoute parentRef port (extended conformance).
	SupportHTTPRouteParentRefPort FeatureName = "HTTPRouteParentRefPort"

	// This option indicates support for HTTPRoute with a backendref with an appProtocol 'kubernetes.io/h2c' (extended conformance)
	SupportHTTPRouteBackendProtocolH2C FeatureName = "HTTPRouteBackendProtocolH2C"

	// This option indicates support for HTTPRoute with a backendref with an appProtoocol 'kubernetes.io/ws' (extended conformance)
	SupportHTTPRouteBackendProtocolWebSocket FeatureName = "HTTPRouteBackendProtocolWebSocket"
)

var (
	HTTPRouteBackendRequestHeaderModificationFeature = Feature{
		Name:   SupportHTTPRouteBackendRequestHeaderModification,
		Status: FeatureStatusStable,
	}

	HTTPRouteQueryParamMatchingFeature = Feature{
		Name:   SupportHTTPRouteQueryParamMatching,
		Status: FeatureStatusStable,
	}

	HTTPRouteMethodMatchingFeature = Feature{
		Name:   SupportHTTPRouteMethodMatching,
		Status: FeatureStatusStable,
	}

	HTTPRouteResponseHeaderModificationFeature = Feature{
		Name:   SupportHTTPRouteResponseHeaderModification,
		Status: FeatureStatusStable,
	}

	HTTPRoutePortRedirectFeature = Feature{
		Name:   SupportHTTPRoutePortRedirect,
		Status: FeatureStatusStable,
	}

	HTTPRouteSchemeRedirectFeature = Feature{
		Name:   SupportHTTPRouteSchemeRedirect,
		Status: FeatureStatusStable,
	}

	HTTPRoutePathRedirectFeature = Feature{
		Name:   SupportHTTPRoutePathRedirect,
		Status: FeatureStatusStable,
	}

	HTTPRouteHostRewriteFeature = Feature{
		Name:   SupportHTTPRouteHostRewrite,
		Status: FeatureStatusStable,
	}

	HTTPRoutePathRewriteFeature = Feature{
		Name:   SupportHTTPRoutePathRewrite,
		Status: FeatureStatusStable,
	}

	HTTPRouteRequestMirrorFeature = Feature{
		Name:   SupportHTTPRouteRequestMirror,
		Status: FeatureStatusStable,
	}

	HTTPRouteRequestMultipleMirrorsFeature = Feature{
		Name:   SupportHTTPRouteRequestMultipleMirrors,
		Status: FeatureStatusStable,
	}

	HTTPRouteRequestTimeoutFeature = Feature{
		Name:   SupportHTTPRouteRequestTimeout,
		Status: FeatureStatusStable,
	}

	HTTPRouteBackendTimeoutFeature = Feature{
		Name:   SupportHTTPRouteBackendTimeout,
		Status: FeatureStatusStable,
	}

	HTTPRouteParentRefPortFeature = Feature{
		Name:   SupportHTTPRouteParentRefPort,
		Status: FeatureStatusStable,
	}

	HTTPRouteBackendProtocolH2CFeature = Feature{
		Name:   SupportHTTPRouteBackendProtocolH2C,
		Status: FeatureStatusStable,
	}

	HTTPRouteBackendProtocolWebSocketFeature = Feature{
		Name:   SupportHTTPRouteBackendProtocolWebSocket,
		Status: FeatureStatusStable,
	}
)

// HTTPRouteExtendedFeatures includes all extended features for HTTPRoute
// conformance and can be used to opt-in to run all HTTPRoute extended features tests.
// This does not include any Core Features.
var HTTPRouteExtendedFeatures = sets.New(
	HTTPRouteBackendRequestHeaderModificationFeature,
	HTTPRouteQueryParamMatchingFeature,
	HTTPRouteMethodMatchingFeature,
	HTTPRouteResponseHeaderModificationFeature,
	HTTPRoutePortRedirectFeature,
	HTTPRouteSchemeRedirectFeature,
	HTTPRoutePathRedirectFeature,
	HTTPRouteHostRewriteFeature,
	HTTPRoutePathRewriteFeature,
	HTTPRouteRequestMirrorFeature,
	HTTPRouteRequestMultipleMirrorsFeature,
	HTTPRouteRequestTimeoutFeature,
	HTTPRouteBackendTimeoutFeature,
	HTTPRouteParentRefPortFeature,
	HTTPRouteBackendProtocolH2CFeature,
	HTTPRouteBackendProtocolWebSocketFeature,
)

// -----------------------------------------------------------------------------
// Features - HTTPRoute Conformance (Experimental)
// -----------------------------------------------------------------------------

const (
	// This option indicates support for Destination Port matching.
	SupportHTTPRouteDestinationPortMatching FeatureName = "HTTPRouteDestinationPortMatching"
)

var (
	HTTPRouteDestinationPortMatchingFeature = Feature{
		Name:   SupportHTTPRouteDestinationPortMatching,
		Status: FeatureStatusTrial,
	}
)

// HTTPRouteExperimentalFeatures includes all the supported experimental features, currently only
// available in our experimental release channel.
// Implementations have the flexibility to opt-in for either specific features or the entire set.
var HTTPRouteExperimentalFeatures = sets.New(
	HTTPRouteDestinationPortMatchingFeature,
)
