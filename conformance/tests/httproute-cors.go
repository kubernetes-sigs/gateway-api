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

	"k8s.io/apimachinery/pkg/types"

	"sigs.k8s.io/gateway-api/conformance/utils/http"
	"sigs.k8s.io/gateway-api/conformance/utils/kubernetes"
	"sigs.k8s.io/gateway-api/conformance/utils/suite"
	"sigs.k8s.io/gateway-api/pkg/features"
)

func init() {
	ConformanceTests = append(ConformanceTests, HTTPRouteRewritePath)
}

var HTTPRouteCORS = suite.ConformanceTest{
	ShortName:   "HTTPRouteCORS",
	Description: "An HTTPRoute with CORS filter",
	Manifests:   []string{"tests/httproute-cors.yaml"},
	Features: []features.FeatureName{
		features.SupportGateway,
		features.SupportHTTPRoute,
		features.SupportHTTPRouteCORS,
	},
	Test: func(t *testing.T, suite *suite.ConformanceTestSuite) {
		ns := "gateway-conformance-infra"
		routeNN := types.NamespacedName{Name: "cors", Namespace: ns}
		gwNN := types.NamespacedName{Name: "same-namespace", Namespace: ns}
		gwAddr := kubernetes.GatewayAndHTTPRoutesMustBeAccepted(t, suite.Client, suite.TimeoutConfig, suite.ControllerName, kubernetes.NewGatewayRef(gwNN), routeNN)
		kubernetes.HTTPRouteMustHaveResolvedRefsConditionsTrue(t, suite.Client, suite.TimeoutConfig, routeNN, gwNN)

		testCases := []http.ExpectedResponse{
			{
				Request: http.Request{
					Path:   "/",
					Method: "OPTIONS",
					Headers: map[string]string{
						"Origin":                         "https://www.foo.com",
						"access-control-request-method":  "GET",
						"access-control-request-headers": "x-header-1, x-header-2",
					},
				},
				// Set the expected request properties and namespace to empty strings.
				// This is a workaround to avoid the test failure.
				// The response body is empty because the request is a preflight request,
				// so we can't get the request properties from the echoserver.
				ExpectedRequest: &http.ExpectedRequest{
					Request: http.Request{
						Host:    "",
						Method:  "OPTIONS",
						Path:    "",
						Headers: nil,
					},
				},
				Namespace: "",
				Response: http.Response{
					StatusCode: 200,
					Headers: map[string]string{
						"access-control-allow-origin":   "https://www.foo.com",
						"access-control-allow-methods":  "GET, POST, PUT, PATCH, DELETE, OPTIONS",
						"access-control-allow-headers":  "x-header-1, x-header-2",
						"access-control-expose-headers": "x-header-3, x-header-4",
					},
				},
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
