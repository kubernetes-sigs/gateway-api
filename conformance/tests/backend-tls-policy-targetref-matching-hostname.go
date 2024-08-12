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

package tests

import (
	"testing"

	"sigs.k8s.io/gateway-api/conformance/utils/suite"
	"sigs.k8s.io/gateway-api/pkg/features"
)

func init() {
	ConformanceTests = append(ConformanceTests, BackendTLSPolicyTargetRefMatchingHostname)
}

var BackendTLSPolicyTargetRefMatchingHostname = suite.ConformanceTest{
	ShortName:   "BackendTLSPolicyTargetRefMatchingHostname",
	Description: "A valid BackendTLSPolicy with a targetRef/servce using CACertificateRef matching for hostname",
	Features: []features.SupportedFeature{
		features.SupportGateway,
		features.SupportHTTPRoute,
		features.SupportBackendTLSPolicy,
	},
	Manifests: []string{"tests/backend-tls-policy-targetref-matching-hostname.yaml"},
	Test: func(t *testing.T, suite *suite.ConformanceTestSuite) {

		// Setup cluster with:
		// 	 gateway: gateway-conformance-infra/tls-backend-gateway with certificateRef and example.org hostname
		//   httpRoute:  gateway-conformance-infra/tls-backend-route
		// 	 deployment: gateway-conformance-infra/tls-backend that serves a cert with common name "www.example.org"
		//   backend service with a cert: gateway-conformance-infra/tls-backend
		// 	 backendTLSPolicy: gateway-conformance-infra/tls-backend-tls-policy with CACertificateRef gateway-conformance-infra/tls-validity-checks-certificate

	},
}
