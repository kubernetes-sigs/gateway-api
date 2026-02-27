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

package tests

import (
	"testing"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"k8s.io/apimachinery/pkg/types"

	v1 "sigs.k8s.io/gateway-api/apis/v1"
	pb "sigs.k8s.io/gateway-api/conformance/echo-basic/grpcechoserver"
	"sigs.k8s.io/gateway-api/conformance/utils/grpc"
	"sigs.k8s.io/gateway-api/conformance/utils/http"
	"sigs.k8s.io/gateway-api/conformance/utils/kubernetes"
	"sigs.k8s.io/gateway-api/conformance/utils/suite"
	"sigs.k8s.io/gateway-api/pkg/features"
)

func init() {
	ConformanceTests = append(ConformanceTests, GRPCRouteRequestMirror)
}

var GRPCRouteRequestMirror = suite.ConformanceTest{
	ShortName:   "GRPCRouteRequestMirror",
	Description: "A GRPCRoute with request mirror filter",
	Manifests:   []string{"tests/grpcroute-request-mirror.yaml"},
	Features: []features.FeatureName{
		features.SupportGRPCRoute,
		features.SupportGateway,
		features.SupportGRPCRouteRequestMirror,
	},
	Test: func(t *testing.T, suite *suite.ConformanceTestSuite) {
		ns := "gateway-conformance-infra"
		routeNN := types.NamespacedName{Name: "grpc-request-mirror", Namespace: ns}
		gwNN := types.NamespacedName{Name: "same-namespace", Namespace: ns}
		gwAddr := kubernetes.GatewayAndRoutesMustBeAccepted(t, suite.Client, suite.TimeoutConfig, suite.ControllerName, kubernetes.NewGatewayRef(gwNN), &v1.GRPCRoute{}, true, routeNN)

		testCases := []grpc.ExpectedResponse{
			{
				EchoRequest: &pb.EchoRequest{},
				Backend:     "grpc-infra-backend-v1",
				Namespace:   ns,
				MirroredTo: []http.MirroredBackend{
					{
						BackendRef: http.BackendRef{
							Name:      "grpc-infra-backend-v2",
							Namespace: ns,
						},
					},
				},
				Response: grpc.Response{
					Code: codes.OK,
				},
			},
			{
				EchoTwoRequest: &pb.EchoRequest{},
				Backend:        "grpc-infra-backend-v1",
				RequestMetadata: &grpc.RequestMetadata{
					Metadata: map[string]string{
						"X-Header-Remove":     "remove-val",
						"X-Header-Add-Append": "append-val-1",
					},
				},
				Namespace: ns,
				MirroredTo: []http.MirroredBackend{
					{
						BackendRef: http.BackendRef{
							Name:      "grpc-infra-backend-v2",
							Namespace: ns,
						},
					},
				},
				Response: grpc.Response{
					Code: codes.OK,
					Headers: func() *metadata.MD {
						md := metadata.Pairs(
							"x-header-set", "set-overwrites-values",
							"x-header-add", "header-val-1",
							"x-header-add-append", "append-val-1",
							"x-header-add-append", "header-val-2",
						)
						return &md
					}(),
					AbsentHeaders: []string{"X-Header-Remove"},
				},
			},
		}
		for i := range testCases {
			tc := testCases[i]
			t.Run(tc.GetTestCaseName(i), func(t *testing.T) {
				t.Parallel()
				grpc.MakeRequestAndExpectEventuallyConsistentResponse(t, suite.GRPCClient, suite.TimeoutConfig, gwAddr, tc)
				grpc.ExpectMirroredRequest(t, suite.Client, suite.Clientset, tc.MirroredTo, suite.TimeoutConfig)
			})
		}
	},
}
