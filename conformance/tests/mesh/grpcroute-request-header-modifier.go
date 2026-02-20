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
	"maps"
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
			name             string
			client           echo.MeshPod
			matchHeaderValue string
			targetHost       string
			headers          map[string]string
			absentHeaders    []string
			expectedHeaders  map[string]string
		}{
			{
				name:             "Set headers -- overwrites existing header",
				client:           clientV1,
				targetHost:       "echo-v1",
				matchHeaderValue: "set",
				headers: map[string]string{
					"some-other-header": "this-header-should-be-set",
					"x-header-set":      "this-value-should-be-overwritten",
				},
				expectedHeaders: map[string]string{
					"x-test-case":       "set",
					"some-other-header": "this-header-should-be-set",
					"x-header-set":      "set-overwrites-values",
				},
			},
			{
				name:             "Add headers -- appends to existing header",
				client:           clientV1,
				targetHost:       "echo-v1",
				matchHeaderValue: "add",
				headers: map[string]string{
					"x-header-add": "this-value-should-be-appended",
				},
				expectedHeaders: map[string]string{
					"x-test-case":  "add",
					"x-header-add": "this-value-should-be-appended,add-appends-values",
				},
			},
			{
				name:             "Remove headers -- removes headers",
				client:           clientV1,
				targetHost:       "echo-v1",
				matchHeaderValue: "remove",
				headers: map[string]string{
					"x-header-remove": "this-should-be-removed",
				},
				absentHeaders: []string{"x-header-remove"},
				expectedHeaders: map[string]string{
					"x-test-case": "remove",
				},
			},
			{
				name:             "Multiple operations -- set, add, and remove headers",
				client:           clientV2,
				targetHost:       "echo-v2",
				matchHeaderValue: "multi",
				headers: map[string]string{
					"x-header-set-2":    "set-header-2",
					"x-header-add-2":    "add-header-2",
					"x-header-remove-2": "should-be-removed-2",
				},
				absentHeaders: []string{"x-header-remove-1", "x-header-remove-2"},
				expectedHeaders: map[string]string{
					"x-test-case":    "multi",
					"x-header-set-1": "header-set-1",
					"x-header-set-2": "header-set-2",
					"x-header-add-1": "header-add-1",
					"x-header-add-2": "add-header-2,header-add-2",
				},
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				t.Parallel()

				hdrs := map[string]string{"x-test-case": tc.matchHeaderValue}
				maps.Copy(hdrs, tc.headers)

				req := http.Request{
					Protocol: "grpc",
					Host:     "echo:7070",
					Headers:  hdrs,
				}

				exp := http.ExpectedResponse{
					Request: req,
					ExpectedRequest: &http.ExpectedRequest{
						Request: http.Request{
							Protocol: "grpc",
							Host:     "echo:7070",
							Headers:  tc.expectedHeaders,
						},
						AbsentHeaders: tc.absentHeaders,
					},
					Response: http.Response{
						StatusCodes: []int{200},
					},
					Backend:   tc.targetHost,
					Namespace: "gateway-conformance-mesh",
				}

				tc.client.MakeRequestAndExpectEventuallyConsistentResponse(t, exp, s.TimeoutConfig)
			})
		}
	},
}
