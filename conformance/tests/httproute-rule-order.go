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
	ConformanceTests = append(ConformanceTests, HTTPRouteRuleOrder)
}

var HTTPRouteRuleOrder = confsuite.ConformanceTest{
	ShortName:   "HTTPRouteRuleOrder",
	Description: "HTTPRoutes with identical matches grant precedence to the first matching rule",
	Features: []features.FeatureName{
		features.SupportGateway,
		features.SupportHTTPRoute,
	},
	Manifests: []string{"tests/httproute-rule-order.yaml"},
	Test: func(t *testing.T, suite *confsuite.ConformanceTestSuite) {
		ns := confsuite.InfrastructureNamespace
		routeNN := types.NamespacedName{Namespace: ns, Name: "rule-order"}
		invalidBackendRouteNN := types.NamespacedName{Namespace: ns, Name: "rule-order-invalid-backend"}
		gwNN := types.NamespacedName{Namespace: ns, Name: "same-namespace"}

		kubernetes.NamespacesMustBeReady(t, suite.Client, suite.TimeoutConfig, []string{ns})

		gwAddr := kubernetes.GatewayAndHTTPRoutesMustBeAccepted(
			t,
			suite.Client,
			suite.TimeoutConfig,
			suite.ControllerName,
			kubernetes.NewGatewayRef(gwNN),
			routeNN,
			invalidBackendRouteNN,
		)
		kubernetes.HTTPRouteMustHaveResolvedRefsConditionsTrue(t, suite.Client, suite.TimeoutConfig, routeNN, gwNN)
		kubernetes.HTTPRouteMustHaveResolvedRefsMustHaveBackendsNotFound(t, suite.Client, suite.TimeoutConfig, invalidBackendRouteNN, gwNN)

		testCases := []http.ExpectedResponse{
			{
				TestCaseName: "first matching rule routes to its backend",
				Request: http.Request{
					Path: "/rule-order",
					Headers: map[string]string{
						"x-color":   "blue",
						"x-version": "one",
					},
				},
				Backend:   confsuite.InfraBackendServiceNameV1,
				Namespace: ns,
			},
			{
				TestCaseName: "invalid first matching rule returns 500",
				Request:      http.Request{Path: "/rule-order-invalid-backend"},
				Response:     http.Response{StatusCode: 500},
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
