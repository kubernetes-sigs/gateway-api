/*
Copyright 2026 The Kubernetes Authors.

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

	"sigs.k8s.io/gateway-api/conformance/utils/echo"
	"sigs.k8s.io/gateway-api/conformance/utils/grpc"
	"sigs.k8s.io/gateway-api/conformance/utils/http"
	"sigs.k8s.io/gateway-api/conformance/utils/suite"
	"sigs.k8s.io/gateway-api/pkg/features"
)

func init() {
	MeshConformanceTests = append(MeshConformanceTests, MeshGRPCRouteRequestMirror)
}

var MeshGRPCRouteRequestMirror = suite.ConformanceTest{
	ShortName:   "MeshGRPCRouteRequestMirror",
	Description: "A GRPCRoute with request mirror filter in mesh mode",
	Manifests:   []string{"tests/mesh/grpcroute-request-mirror.yaml"},
	Features: []features.FeatureName{
		features.SupportMesh,
		features.SupportGRPCRoute,
		features.SupportGRPCRouteRequestMirror,
	},
	Test: func(t *testing.T, s *suite.ConformanceTestSuite) {
		ns := "gateway-conformance-mesh"
		client := echo.ConnectToApp(t, s, echo.MeshAppEchoV1)

		mirrorPods := []http.MirroredBackend{
			{
				BackendRef: http.BackendRef{
					Name:      "echo-v2",
					Namespace: ns,
				},
				Labels:    map[string]string{"app": "echo", "version": "v2"},
				Container: "echo",
			},
		}

		testCases := []http.ExpectedResponse{
			{
				Request: http.Request{
					Protocol: "grpc",
					Host:     "echo:7070",
					Headers:  map[string]string{"x-mirror-only": "true"},
				},
				Response:  http.Response{StatusCode: 200},
				Backend:   "echo-v1",
				Namespace: ns,
			},
			{
				Request: http.Request{
					Protocol: "grpc",
					Host:     "echo:7070",
					Headers: map[string]string{
						"x-mirror-and-modify": "true",
						"X-Header-Remove":     "remove-val",
						"X-Header-Add-Append": "append-val-1",
					},
				},
				ExpectedRequest: &http.ExpectedRequest{
					Request: http.Request{
						Headers: map[string]string{
							"X-Header-Set":        "set-overwrites-values",
							"X-Header-Add":        "header-val-1",
							"X-Header-Add-Append": "append-val-1,header-val-2",
						},
					},
					AbsentHeaders: []string{"X-Header-Remove"},
				},
				Response:  http.Response{StatusCode: 200},
				Backend:   "echo-v1",
				Namespace: ns,
			},
		}
		for i := range testCases {
			tc := testCases[i]
			t.Run(tc.GetTestCaseName(i), func(t *testing.T) {
				client.MakeRequestAndExpectEventuallyConsistentResponse(t, tc, s.TimeoutConfig)
				grpc.ExpectMeshMirroredRequest(t, s.Client, s.Clientset, mirrorPods, s.TimeoutConfig)
			})
		}
	},
}
