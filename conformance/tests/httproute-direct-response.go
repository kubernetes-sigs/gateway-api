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

package tests

import (
	"testing"

	"k8s.io/apimachinery/pkg/types"

	"sigs.k8s.io/gateway-api/conformance/utils/http"
	"sigs.k8s.io/gateway-api/conformance/utils/kubernetes"
	confsuite "sigs.k8s.io/gateway-api/conformance/utils/suite"
	"sigs.k8s.io/gateway-api/pkg/features"
)

func init() {
	ConformanceTests = append(ConformanceTests, HTTPRouteDirectResponse)
}

var HTTPRouteDirectResponse = confsuite.ConformanceTest{
	ShortName:   "HTTPRouteDirectResponse",
	Description: "An HTTPRoute with a DirectResponse filter should reply directly from the gateway without forwarding to a backend",
	Manifests:   []string{"tests/httproute-direct-response.yaml"},
	Features: []features.FeatureName{
		features.SupportGateway,
		features.SupportHTTPRoute,
		features.SupportHTTPRouteDirectResponse,
	},
	Test: func(t *testing.T, suite *confsuite.ConformanceTestSuite) {
		ns := confsuite.InfrastructureNamespace
		routeNN1 := types.NamespacedName{Name: "direct-response-with-body", Namespace: ns}
		routeNN2 := types.NamespacedName{Name: "direct-response-no-body", Namespace: ns}
		routeNN3 := types.NamespacedName{Name: "direct-response-forbidden", Namespace: ns}
		gwNN := types.NamespacedName{Name: "same-namespace", Namespace: ns}

		gwAddr := kubernetes.GatewayAndHTTPRoutesMustBeAccepted(t, suite.Client, suite.TimeoutConfig, suite.ControllerName, kubernetes.NewGatewayRef(gwNN), routeNN1, routeNN2, routeNN3)

		testCases := []http.ExpectedResponse{
			{
				TestCaseName: "DirectResponse filter should return configured status code and body",
				Namespace:    ns,
				Request: http.Request{
					Path: "/direct-response/body",
				},
				ExpectedRequest: &http.ExpectedRequest{
					Request: http.Request{
						Host:   "",
						Path:   "",
						Method: "",
					},
				},
				Response: http.Response{
					StatusCode: 200,
				},
			},
			{
				TestCaseName: "DirectResponse filter without body should return configured status code with empty body",
				Namespace:    ns,
				Request: http.Request{
					Path: "/direct-response/no-body",
				},
				ExpectedRequest: &http.ExpectedRequest{
					Request: http.Request{
						Host:   "",
						Path:   "",
						Method: "",
					},
				},
				Response: http.Response{
					StatusCode: 204,
				},
			},
			{
				TestCaseName: "DirectResponse filter should return 403 with body for blocked path",
				Namespace:    ns,
				Request: http.Request{
					Path: "/direct-response/forbidden",
				},
				ExpectedRequest: &http.ExpectedRequest{
					Request: http.Request{
						Host:   "",
						Path:   "",
						Method: "",
					},
				},
				Response: http.Response{
					StatusCode: 403,
				},
			},
		}

		for i := range testCases {
			tc := testCases[i]
			t.Run(tc.GetTestCaseName(i), func(t *testing.T) {
				t.Parallel()
				http.MakeRequestAndExpectEventuallyConsistentResponse(t, suite.RoundTripper, suite.TimeoutConfig, gwAddr, tc)
			})
		}
	},
}
