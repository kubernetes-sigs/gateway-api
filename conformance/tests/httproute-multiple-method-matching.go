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
	ConformanceTests = append(ConformanceTests, HTTPRouteMultipleMethodMatching)
}

var HTTPRouteMultipleMethodMatching = confsuite.ConformanceTest{
	ShortName:   "HTTPRouteMultipleMethodMatching",
	Description: "A single HTTPRouteMatch matches multiple HTTP methods listed in methods",
	Manifests:   []string{"tests/httproute-multiple-method-matching.yaml"},
	Features: []features.FeatureName{
		features.SupportGateway,
		features.SupportHTTPRoute,
		features.SupportHTTPRouteMultipleMethodMatching,
	},
	Test: func(t *testing.T, suite *confsuite.ConformanceTestSuite) {
		ns := confsuite.InfrastructureNamespace
		routeNN := types.NamespacedName{Name: "multiple-method-matching", Namespace: ns}
		gwNN := types.NamespacedName{Name: "same-namespace", Namespace: ns}
		gwAddr := kubernetes.GatewayAndHTTPRoutesMustBeAccepted(t, suite.Client, suite.TimeoutConfig, suite.ControllerName, kubernetes.NewGatewayRef(gwNN), routeNN)
		kubernetes.HTTPRouteMustHaveResolvedRefsConditionsTrue(t, suite.Client, suite.TimeoutConfig, routeNN, gwNN)

		testCases := []http.ExpectedResponse{
			{
				Request:   http.Request{Method: "GET", Path: "/multi-method"},
				Backend:   confsuite.InfraBackendServiceNameV1,
				Namespace: ns,
			},
			{
				Request:   http.Request{Method: "POST", Path: "/multi-method"},
				Backend:   confsuite.InfraBackendServiceNameV1,
				Namespace: ns,
			},
			{
				Request:   http.Request{Method: "PUT", Path: "/multi-method"},
				Backend:   confsuite.InfraBackendServiceNameV1,
				Namespace: ns,
			},
			{
				Request:  http.Request{Method: "DELETE", Path: "/multi-method"},
				Response: http.Response{StatusCode: 404},
			},
			{
				Request: http.Request{
					Method:  "DELETE",
					Path:    "/multi-method-with-header",
					Headers: map[string]string{"version": "one"},
				},
				Backend:   confsuite.InfraBackendServiceNameV2,
				Namespace: ns,
			},
			{
				Request: http.Request{
					Method:  "PATCH",
					Path:    "/multi-method-with-header",
					Headers: map[string]string{"version": "one"},
				},
				Backend:   confsuite.InfraBackendServiceNameV2,
				Namespace: ns,
			},
			{
				Request: http.Request{
					Method:  "DELETE",
					Path:    "/multi-method-with-header",
					Headers: map[string]string{"version": "two"},
				},
				Response: http.Response{StatusCode: 404},
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