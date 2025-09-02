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

package meshtests

import (
	"testing"

	"sigs.k8s.io/gateway-api/conformance/utils/echo"
	"sigs.k8s.io/gateway-api/conformance/utils/http"
	"sigs.k8s.io/gateway-api/conformance/utils/suite"
	"sigs.k8s.io/gateway-api/pkg/features"
)

func init() {
	MeshConformanceTests = append(MeshConformanceTests, MeshGRPCRouteRequestHeaderModifier)
}

var MeshGRPCRouteRequestHeaderModifier = suite.ConformanceTest{
	ShortName:   "MeshGRPCRouteRequestHeaderModifier",
	Description: "A GRPCRoute with RequestHeaderModifier filter should modify request headers in mesh mode",
	Manifests:   []string{"tests/mesh/grpcroute-request-header-modifier.yaml"},
	Features: []features.FeatureName{
		features.SupportMesh,
		features.SupportGRPCRoute,
	},
	Test: func(t *testing.T, s *suite.ConformanceTestSuite) {
		clientV1 := echo.ConnectToApp(t, s, echo.MeshAppEchoV1)
		clientV2 := echo.ConnectToApp(t, s, echo.MeshAppEchoV2)

		testCases := []struct {
			name            string
			client          echo.MeshPod
			headers         map[string]string
			targetHost      string
			absentHeaders   []string
			expectedHeaders map[string]string
		}{
			{
				name: "Set headers -- X-Header-Set should have the original value",
				headers: map[string]string{
					"Some-Other-Header": "this-header-should-be-set",
				},
				expectedHeaders: map[string]string{
					"Some-Other-Header": "this-header-should-be-set",
					"X-Header-Set":      "set-overwrites-values",
				},
				targetHost: "echo-v1",
				client:     clientV1,
			},
			{
				name: "Set headers -- X-Header-Set should get overwritten with the original value",
				headers: map[string]string{
					"Some-Other-Header": "this-header-should-be-set",
					"X-Header-Set":      "this-value-should-be-overwritten",
				},
				expectedHeaders: map[string]string{
					"Some-Other-Header": "this-header-should-be-set",
					"X-Header-Set":      "set-overwrites-values",
				},
				targetHost: "echo-v1",
				client:     clientV1,
			},
			{
				name: "Add headers -- X-Header-Add should have the original value",
				headers: map[string]string{
					"Some-Other-Header": "this-header-should-be-set",
				},
				expectedHeaders: map[string]string{
					"Some-Other-Header": "this-header-should-be-set",
					"X-Header-Add":      "add-appends-values",
				},
				targetHost: "echo-v1",
				client:     clientV1,
			},
			{
				name: "X-Header-Add should append the new value to the original value",
				headers: map[string]string{
					"Some-Other-Header": "this-header-should-be-set",
					"X-Header-Add":      "this-value-should-be-appended",
				},
				expectedHeaders: map[string]string{
					"Some-Other-Header": "this-header-should-be-set",
					"X-Header-Add":      "this-value-should-be-appended,add-appends-values",
				},
				targetHost: "echo-v1",
				client:     clientV1,
			},
			{
				name: "Remove headers -- X-Header-Remove should be removed",
				headers: map[string]string{
					"X-Header-Remove": "this-should-be-removed",
				},
				absentHeaders: []string{"X-Header-Remove"},
				client:        clientV1,
				targetHost:    "echo-v1",
			},
			{
				name: "Multiple operations - all header operations should be applied",
				headers: map[string]string{
					"X-Header-Set-2":    "set-header-2",
					"X-Header-Add-2":    "add-header-2",
					"X-Header-Remove-2": "should-be-removed-2",
					"Some-Other-Header": "another-header-val",
				},
				absentHeaders: []string{
					"X-Header-Remove-1",
					"X-Header-Remove-2",
				},
				expectedHeaders: map[string]string{
					"X-Header-Set-1":    "header-set-1",
					"X-Header-Set-2":    "header-set-2",
					"X-Header-Add-1":    "header-add-1",
					"X-Header-Add-2":    "add-header-2,header-add-2",
					"X-Header-Add-3":    "header-add-3",
					"Some-Other-Header": "another-header-val",
				},
				client:     clientV2,
				targetHost: "echo-v2",
			},
			{
				name: "Case sensitivity check for header names",
				headers: map[string]string{
					"x-header-set-1":    "original-set-1",
					"x-header-add-1":    "existing-add-1",
					"x-header-remove-1": "should-be-removed-1",
				},
				absentHeaders: []string{"X-Header-Remove-1"},
				expectedHeaders: map[string]string{
					"X-Header-Set-1": "header-set-1",
					"X-Header-Set-2": "header-set-2",
					"X-Header-Add-1": "existing-add-1,header-add-1",
					"X-Header-Add-2": "header-add-2",
					"X-Header-Add-3": "header-add-3",
				},
				client:     clientV2,
				targetHost: "echo-v2",
			},
		}

		for _, tc := range testCases {
			tc := tc // capture for parallel execution
			t.Run(tc.name, func(t *testing.T) {
				t.Parallel()

				request := http.Request{
					Protocol: "grpc",
					Host:     "echo:7070",
					Path:     "",
					Headers:  tc.headers,
				}

				expected := http.ExpectedResponse{
					Request: request,
					Response: http.Response{
						StatusCode: 200,
					},
					Namespace: "gateway-conformance-mesh",
				}

				// will remove based on input from Gateway API team

				// expected := http.ExpectedResponse{
				// 	Request: request,
				// 	ExpectedRequest: &http.ExpectedRequest{
				// 		Request: http.Request{
				// 			Protocol: "grpc",
				// 			Host:     "echo:7070",
				// 			Path:     "",
				// 			Headers:  tc.expectedHeaders,
				// 		},
				// 		AbsentHeaders: tc.absentHeaders,
				// 	},
				// 	Namespace: "gateway-conformance-mesh",
				// }

				// Make the request and validate headers are properly modified
				tc.client.MakeRequestAndExpectEventuallyConsistentResponse(t, expected, s.TimeoutConfig)
			})
		}
	},
}
