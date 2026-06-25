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
	ConformanceTests = append(ConformanceTests, HTTPRouteHostnamePrecedence)
}

var HTTPRouteHostnamePrecedence = suite.ConformanceTest{
	ShortName:   "HTTPRouteHostnamePrecedence",
	Description: "When a Gateway has both an HTTPRoute with a specified hostname and another without a hostname, the route with the specified hostname should take precedence",
	Manifests:   []string{"tests/httproute-hostname-precedence.yaml"},
	Features: []features.FeatureName{
		features.SupportGateway,
		features.SupportHTTPRoute,
	},
	Test: func(t *testing.T, suite *suite.ConformanceTestSuite) {
		ns := "gateway-conformance-infra"
		specificRouteNN := types.NamespacedName{Name: "hostname-precedence-with-specific", Namespace: ns}
		nonSpecificRouteNN := types.NamespacedName{Name: "hostname-precedence-without-specific", Namespace: ns}
		gwNN := types.NamespacedName{Name: "same-namespace", Namespace: ns}

		gwAddr := kubernetes.GatewayAndHTTPRoutesMustBeAccepted(t, suite.Client, suite.TimeoutConfig, suite.ControllerName, kubernetes.NewGatewayRef(gwNN), specificRouteNN)
		kubernetes.GatewayAndHTTPRoutesMustBeAccepted(t, suite.Client, suite.TimeoutConfig, suite.ControllerName, kubernetes.NewGatewayRef(gwNN), nonSpecificRouteNN)

		kubernetes.HTTPRouteMustHaveResolvedRefsConditionsTrue(t, suite.Client, suite.TimeoutConfig, nonSpecificRouteNN, gwNN)
		kubernetes.HTTPRouteMustHaveResolvedRefsConditionsTrue(t, suite.Client, suite.TimeoutConfig, specificRouteNN, gwNN)

		expectedHeaders := map[string]string{
			"test": "true",
		}

		testCases := []http.ExpectedResponse{
			{
				Request:   http.Request{Path: "/hostname-precedence", Host: "test.com"},
				Response:  http.Response{Headers: expectedHeaders},
				Namespace: ns,
			},
			{
				Request:   http.Request{Path: "/hostname-precedence"},
				Response:  http.Response{AbsentHeaders: []string{"test"}},
				Namespace: ns,
			},
		}

		for i := range testCases {
			// Declare tc here to avoid loop variable
			// reuse issues across parallel tests.
			tc := testCases[i]
			t.Run(tc.GetTestCaseName(i), func(t *testing.T) {
				t.Parallel()
				http.MakeRequestAndExpectEventuallyConsistentResponse(t, suite.RoundTripper, suite.TimeoutConfig, gwAddr, tc)
			})
		}
	},
}
