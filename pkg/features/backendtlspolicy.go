/*
Copyright 2025 The Kubernetes Authors.

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
// Features - BackendTLSPolicy Conformance (Core)
// -----------------------------------------------------------------------------

const (
	// This option indicates support for BackendTLSPolicy.
	SupportBackendTLSPolicy FeatureName = "BackendTLSPolicy"

	// This option indicates support for BackendTLSPolicy SubjectAltName Validation.
	SupportBackendTLSPolicySANValidation FeatureName = "BackendTLSPolicySANValidation"

	// This option indicates support for ClusterTrustBundle references in BackendTLSPolicyValidation.
	SupportBackendTLSPolicyClusterTrustBundle FeatureName = "BackendTLSPolicyClusterTrustBundle"

	// This option indicates support for ClusterTrustBundle references in Gateway frontend TLS validation.
	SupportGatewayFrontendClusterTrustBundle FeatureName = "GatewayFrontendClusterTrustBundle"
)

// BackendTLSPolicyFeature contains metadata for the BackendTLSPolicy feature.
var BackendTLSPolicyFeature = Feature{
	Name:    SupportBackendTLSPolicy,
	Channel: FeatureChannelStandard,
}

// BackendTLSPolicySanValidationFeature contains metadata for the BackendTLSPolicy
// SubjectAltName Validation feature.
var BackendTLSPolicySanValidationFeature = Feature{
	Name:    SupportBackendTLSPolicySANValidation,
	Channel: FeatureChannelStandard,
}

// BackendTLSPolicyClusterTrustBundleFeature contains metadata for the BackendTLSPolicy
// ClusterTrustBundle reference feature.
var BackendTLSPolicyClusterTrustBundleFeature = Feature{
	Name:    SupportBackendTLSPolicyClusterTrustBundle,
	Channel: FeatureChannelExperimental,
}

// GatewayFrontendClusterTrustBundleFeature contains metadata for the Gateway
// frontend TLS ClusterTrustBundle reference feature.
var GatewayFrontendClusterTrustBundleFeature = Feature{
	Name:    SupportGatewayFrontendClusterTrustBundle,
	Channel: FeatureChannelExperimental,
}

// BackendTLSPolicyCoreFeatures includes all the supported features for the
// BackendTLSPolicy API at a Core level of support.
var BackendTLSPolicyCoreFeatures = sets.New(
	BackendTLSPolicyFeature,
)

// BackendTLSPolicyExtendedFeatures includes all the supported features for the
// BackendTLSPolicy API at a Extended level of support.
var BackendTLSPolicyExtendedFeatures = sets.New(
	BackendTLSPolicySanValidationFeature,
	BackendTLSPolicyClusterTrustBundleFeature,
	GatewayFrontendClusterTrustBundleFeature,
)
