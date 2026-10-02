/*
Copyright The Kubernetes Authors.

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
// Features - Backend Conformance (Core)
// -----------------------------------------------------------------------------

const (
	// This option indicates support for Backend.
	SupportBackend FeatureName = "Backend"
)

// BackendFeature contains metadata for the Backend feature.
var BackendFeature = Feature{
	Name:    SupportBackend,
	Channel: FeatureChannelExperimental,
}

// BackendCoreFeatures includes all the supported features for the Backend API
// at a Core level of support.
var BackendCoreFeatures = sets.New(
	BackendFeature,
)

// -----------------------------------------------------------------------------
// Features - Backend Conformance (Extended)
// -----------------------------------------------------------------------------

const (
	// This option indicates support for ExternalHostname Backend destinations.
	SupportBackendExternalHostname FeatureName = "BackendExternalHostname"

	// This option indicates support for Backend TLS configuration.
	SupportBackendTLS FeatureName = "BackendTLS"

	// This option indicates support for the HTTP2 Backend protocol.
	SupportBackendProtocolHTTP2 FeatureName = "BackendProtocolHTTP2"

	// This option indicates support for the TCP Backend protocol.
	SupportBackendProtocolTCP FeatureName = "BackendProtocolTCP"

	// This option indicates support for the GRPC Backend protocol.
	SupportBackendProtocolGRPC FeatureName = "BackendProtocolGRPC"

	// This option indicates support for the MCP Backend protocol.
	SupportBackendProtocolMCP FeatureName = "BackendProtocolMCP"

	// This option indicates support for the WSS Backend protocol.
	SupportBackendProtocolWSS FeatureName = "BackendProtocolWSS"

	// This option indicates support for cookie-based Backend session persistence.
	SupportBackendSessionPersistenceCookie FeatureName = "BackendSessionPersistenceCookie"

	// This option indicates support for header-based Backend session persistence.
	SupportBackendSessionPersistenceHeader FeatureName = "BackendSessionPersistenceHeader"

	// This option indicates support for absolute timeout in Backend session persistence.
	SupportBackendSessionPersistenceSessionCookieAbsoluteTimeout FeatureName = "BackendSessionPersistenceSessionCookieAbsoluteTimeout"

	// This option indicates support for configuring the cookie path in Backend session persistence.
	SupportBackendSessionPersistenceCookiePath FeatureName = "BackendSessionPersistenceCookiePath"
)

var (
	// BackendExternalHostnameFeature contains metadata for the
	// BackendExternalHostname feature.
	BackendExternalHostnameFeature = Feature{
		Name:    SupportBackendExternalHostname,
		Channel: FeatureChannelExperimental,
	}

	// BackendTLSFeature contains metadata for the BackendTLS feature.
	BackendTLSFeature = Feature{
		Name:    SupportBackendTLS,
		Channel: FeatureChannelExperimental,
	}

	// BackendProtocolHTTP2Feature contains metadata for the BackendProtocolHTTP2
	// feature.
	BackendProtocolHTTP2Feature = Feature{
		Name:    SupportBackendProtocolHTTP2,
		Channel: FeatureChannelExperimental,
	}

	// BackendProtocolTCPFeature contains metadata for the BackendProtocolTCP
	// feature.
	BackendProtocolTCPFeature = Feature{
		Name:    SupportBackendProtocolTCP,
		Channel: FeatureChannelExperimental,
	}

	// BackendProtocolGRPCFeature contains metadata for the BackendProtocolGRPC
	// feature.
	BackendProtocolGRPCFeature = Feature{
		Name:    SupportBackendProtocolGRPC,
		Channel: FeatureChannelExperimental,
	}

	// BackendProtocolMCPFeature contains metadata for the BackendProtocolMCP
	// feature.
	BackendProtocolMCPFeature = Feature{
		Name:    SupportBackendProtocolMCP,
		Channel: FeatureChannelExperimental,
	}

	// BackendProtocolWSSFeature contains metadata for the BackendProtocolWSS
	// feature.
	BackendProtocolWSSFeature = Feature{
		Name:    SupportBackendProtocolWSS,
		Channel: FeatureChannelExperimental,
	}

	// BackendSessionPersistenceCookieFeature contains metadata for the
	// BackendSessionPersistenceCookie feature.
	BackendSessionPersistenceCookieFeature = Feature{
		Name:    SupportBackendSessionPersistenceCookie,
		Channel: FeatureChannelExperimental,
	}

	// BackendSessionPersistenceHeaderFeature contains metadata for the
	// BackendSessionPersistenceHeader feature.
	BackendSessionPersistenceHeaderFeature = Feature{
		Name:    SupportBackendSessionPersistenceHeader,
		Channel: FeatureChannelExperimental,
	}

	// BackendSessionPersistenceSessionCookieAbsoluteTimeoutFeature contains metadata for the
	// BackendSessionPersistenceSessionCookieAbsoluteTimeout feature.
	BackendSessionPersistenceSessionCookieAbsoluteTimeoutFeature = Feature{
		Name:    SupportBackendSessionPersistenceSessionCookieAbsoluteTimeout,
		Channel: FeatureChannelExperimental,
	}

	// BackendSessionPersistenceCookiePathFeature contains metadata for the
	// BackendSessionPersistenceCookiePath feature.
	BackendSessionPersistenceCookiePathFeature = Feature{
		Name:    SupportBackendSessionPersistenceCookiePath,
		Channel: FeatureChannelExperimental,
	}
)

// BackendExtendedFeatures includes all the supported features for the
// Backend API at an Extended level of support.
var BackendExtendedFeatures = sets.New(
	BackendExternalHostnameFeature,
	BackendTLSFeature,
	BackendProtocolHTTP2Feature,
	BackendProtocolTCPFeature,
	BackendProtocolGRPCFeature,
	BackendProtocolMCPFeature,
	BackendProtocolWSSFeature,
	BackendSessionPersistenceCookieFeature,
	BackendSessionPersistenceHeaderFeature,
	BackendSessionPersistenceSessionCookieAbsoluteTimeoutFeature,
	BackendSessionPersistenceCookiePathFeature,
)
