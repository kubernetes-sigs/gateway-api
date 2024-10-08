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

package tests

import (
	"testing"

	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/gateway-api/conformance/utils/http"
	"sigs.k8s.io/gateway-api/conformance/utils/kubernetes"
	"sigs.k8s.io/gateway-api/conformance/utils/suite"
	"sigs.k8s.io/gateway-api/pkg/features"
)

func init() {
	ConformanceTests = append(ConformanceTests, BackendTLSPolicy)
}

var BackendTLSPolicy = suite.ConformanceTest{
	ShortName:   "BackendTLSPolicy",
	Description: "A single service that is targeted by a BackendTLSPolicy must successfully complete TLS termination",
	Features: []features.FeatureName{
		features.SupportGateway,
		features.SupportBackendTLSPolicy,
	},
	Manifests: []string{"tests/backendtlspolicy.yaml"},
	Test: func(t *testing.T, suite *suite.ConformanceTestSuite) {
		ns := "gateway-conformance-infra"
		routeNN := types.NamespacedName{Name: "gateway-conformance-infra-test", Namespace: ns}
		gwNN := types.NamespacedName{Name: "gateway-backendtlspolicy", Namespace: ns}

		kubernetes.NamespacesMustBeReady(t, suite.Client, suite.TimeoutConfig, []string{ns})

		gwAddr := kubernetes.GatewayAndHTTPRoutesMustBeAccepted(t, suite.Client, suite.TimeoutConfig, suite.ControllerName, kubernetes.NewGatewayRef(gwNN), routeNN)

		serverStr := "abc.example.com"
		headers := make(map[string]string)
		headers["Host"] = serverStr

		// Verify that the response to a call to /backendTLS will return the matching SNI.
		t.Run("Simple request targeting BackendTLSPolicy should reach infra-backend", func(t *testing.T) {
			http.MakeHTTPSRequestAndExpectEventuallyConsistentResponse(t, suite.RoundTripper, suite.TimeoutConfig, gwAddr,
				http.ExpectedResponse{
					Request: http.Request{
						Headers: headers,
						Host:    serverStr,
						Path:    "/backendTLS",
					},
					Response:   http.Response{StatusCode: 200},
					Namespace:  "gateway-conformance-infra",
					ServerName: serverStr,
				})
		})
	},
}
