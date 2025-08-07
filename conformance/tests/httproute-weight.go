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
	ConformanceTests = append(ConformanceTests, HTTPRouteWeight)
}

var HTTPRouteWeight = suite.ConformanceTest{
	ShortName:   "HTTPRouteWeight",
	Description: "An HTTPRoute with weighted backends",
	Manifests:   []string{"tests/httproute-weight.yaml"},
	Features: []features.FeatureName{
		features.SupportGateway,
		features.SupportHTTPRoute,
	},
	Test: func(t *testing.T, suite *suite.ConformanceTestSuite) {
		var (
			ns      = "gateway-conformance-infra"
			routeNN = types.NamespacedName{Name: "weighted-backends", Namespace: ns}
			gwNN    = types.NamespacedName{Name: "same-namespace", Namespace: ns}
			gwAddr  = kubernetes.GatewayAndHTTPRoutesMustBeAccepted(t, suite.Client, suite.TimeoutConfig, suite.ControllerName, kubernetes.NewGatewayRef(gwNN), routeNN)
		)

		kubernetes.HTTPRouteMustHaveResolvedRefsConditionsTrue(t, suite.Client, suite.TimeoutConfig, routeNN, gwNN)

		t.Run("Requests should have a distribution that matches the weight", func(t *testing.T) {
			expected := http.ExpectedResponse{
				Request:   http.Request{Path: "/"},
				Response:  http.Response{StatusCode: 200},
				Namespace: "gateway-conformance-infra",
			}

			// Assert request succeeds before doing our distribution check
			http.MakeRequestAndExpectEventuallyConsistentResponse(t, suite.RoundTripper, suite.TimeoutConfig, gwAddr, expected)

			expectedWeights := map[string]float64{
				"infra-backend-v1": 0.7,
				"infra-backend-v2": 0.3,
				"infra-backend-v3": 0.0,
			}

			sender := newHTTPRequestSender(t, suite, gwAddr, expected)

			for i := 0; i < 10; i++ {
				if err := testWeightedDistribution(sender, expectedWeights); err != nil {
					t.Logf("Traffic distribution test failed (%d/10): %s", i+1, err)
				} else {
					return
				}
			}
			t.Fatal("Weighted distribution tests failed")
		})
	},
}

