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

package meshtests

import (
	"testing"

	"sigs.k8s.io/gateway-api/conformance/tests"
	"sigs.k8s.io/gateway-api/conformance/utils/echo"
	"sigs.k8s.io/gateway-api/conformance/utils/http"
	"sigs.k8s.io/gateway-api/conformance/utils/suite"
	"sigs.k8s.io/gateway-api/pkg/features"
)

func init() {
	MeshConformanceTests = append(MeshConformanceTests, MeshHTTPRouteWeight)
}

var MeshHTTPRouteWeight = suite.ConformanceTest{
	ShortName:   "MeshHTTPRouteWeight",
	Description: "An HTTPRoute with weighted backends",
	Manifests:   []string{"tests/mesh/httproute-weight.yaml"},
	Features: []features.FeatureName{
		features.SupportMesh,
		features.SupportHTTPRoute,
	},
	Test: func(t *testing.T, s *suite.ConformanceTestSuite) {
		client := echo.ConnectToApp(t, s, echo.MeshAppEchoV1)

		t.Run("Requests should have a distribution that matches the weight", func(t *testing.T) {
			host := "echo"
			expected := http.ExpectedResponse{
				Request:   http.Request{Path: "/", Host: host},
				Response:  http.Response{StatusCode: 200},
				Namespace: "gateway-conformance-mesh",
			}

			// Assert request succeeds before doing our distribution check
			client.MakeRequestAndExpectEventuallyConsistentResponse(t, expected, s.TimeoutConfig)

			expectedWeights := map[string]float64{
				"echo-v1": 0.7,
				"echo-v2": 0.3,
			}

			sender := tests.NewMeshRequestSender(t, client, expected)

			for i := 0; i < 10; i++ {
				if err := tests.TestWeightedDistribution(sender, expectedWeights); err != nil {
					t.Logf("Traffic distribution test failed (%d/10): %s", i+1, err)
				} else {
					return
				}
			}
			t.Fatal("Weighted distribution tests failed")
		})
	},
}

