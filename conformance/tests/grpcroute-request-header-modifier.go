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

	"google.golang.org/grpc/metadata"
	"k8s.io/apimachinery/pkg/types"

	v1 "sigs.k8s.io/gateway-api/apis/v1"
	pb "sigs.k8s.io/gateway-api/conformance/echo-basic/grpcechoserver"
	"sigs.k8s.io/gateway-api/conformance/utils/grpc"
	"sigs.k8s.io/gateway-api/conformance/utils/kubernetes"
	"sigs.k8s.io/gateway-api/conformance/utils/suite"
	"sigs.k8s.io/gateway-api/pkg/features"
)

func init() {
	ConformanceTests = append(ConformanceTests, GRPCRouteRequestHeaderModifier)
}

var GRPCRouteRequestHeaderModifier = suite.ConformanceTest{
	ShortName:   "GRPCRouteRequestHeaderModifier",
	Description: "A GRPCRoute with RequestHeaderModifier filter should modify request headers",
	Manifests:   []string{"tests/grpcroute-request-header-modifier.yaml"},
	Features: []features.FeatureName{
		features.SupportGateway,
		features.SupportGRPCRoute,
		features.SupportGRPCRouteRequestHeaderModifier,
	},
	Test: func(t *testing.T, suite *suite.ConformanceTestSuite) {
		ns := "gateway-conformance-infra"
		routeNN := types.NamespacedName{Name: "grpc-request-header-modifier", Namespace: ns}
		gwNN := types.NamespacedName{Name: "same-namespace", Namespace: ns}
		gwAddr := kubernetes.GatewayAndRoutesMustBeAccepted(t, suite.Client, suite.TimeoutConfig, suite.ControllerName, kubernetes.NewGatewayRef(gwNN), &v1.GRPCRoute{}, true, routeNN)

		testCases := []grpc.ExpectedResponse{
			{
				TestCaseName: "Set headers -- X-Header-Set should have the original value",
				EchoRequest:  &pb.EchoRequest{},
				RequestMetadata: &grpc.RequestMetadata{
					Authority: "gateway_api_conformance.echo_basic.grpcecho.GrpcEcho",
					Metadata: map[string]string{
						"Some-Other-Header": "this-header-should-be-set",
					},
				},
				Response: grpc.Response{
					Headers: &metadata.MD{
						"Some-Other-Header": {"this-header-should-be-set"},
						"X-Header-Set":      {"set-overwrites-values"},
					},
				},
				Backend:   "grpc-infra-backend-v1",
				Namespace: ns,
			}, {
				TestCaseName: "Set headers -- X-Header-Set should get overwritten with the original value",
				EchoRequest:  &pb.EchoRequest{},
				RequestMetadata: &grpc.RequestMetadata{
					Authority: "gateway_api_conformance.echo_basic.grpcecho.GrpcEcho",
					Metadata: map[string]string{
						"Some-Other-Header": "this-header-should-be-set",
						"X-Header-Set":      "this-value-should-be-overwritten",
					},
				},
				Response: grpc.Response{
					Headers: &metadata.MD{
						"Some-Other-Header": {"this-header-should-be-set"},
						"X-Header-Set":      {"set-overwrites-values"},
					},
				},
				Backend:   "grpc-infra-backend-v1",
				Namespace: ns,
			}, {
				TestCaseName: "Add headers -- X-Header-Add should have the original value",
				EchoRequest:  &pb.EchoRequest{},
				RequestMetadata: &grpc.RequestMetadata{
					Authority: "gateway_api_conformance.echo_basic.grpcecho.GrpcEcho",
					Metadata: map[string]string{
						"Some-Other-Header": "this-header-should-be-set",
					},
				},
				Response: grpc.Response{
					Headers: &metadata.MD{
						"Some-Other-Header": {"this-header-should-be-set"},
						"X-Header-Add":      {"add-appends-values"},
					},
				},
				Backend:   "grpc-infra-backend-v1",
				Namespace: ns,
			},
			{
				TestCaseName: "Add headers -- X-Header-Add should append the new value to the original value",
				EchoRequest:  &pb.EchoRequest{},
				RequestMetadata: &grpc.RequestMetadata{
					Authority: "gateway_api_conformance.echo_basic.grpcecho.GrpcEcho",
					Metadata: map[string]string{
						"Some-Other-Header": "this-header-should-be-set",
						"X-Header-Add":      "this-value-should-be-appended",
					},
				},
				Response: grpc.Response{
					Headers: &metadata.MD{
						"Some-Other-Header": {"this-header-should-be-set"},
						"X-Header-Add":      {"this-value-should-be-appended", "add-appends-values"},
					},
				},
				Backend:   "grpc-infra-backend-v1",
				Namespace: ns,
			}, {
				TestCaseName: "Remove headers -- X-Header-Remove should be removed",
				EchoRequest:  &pb.EchoRequest{},
				RequestMetadata: &grpc.RequestMetadata{
					Authority: "gateway_api_conformance.echo_basic.grpcecho.GrpcEcho",
					Metadata: map[string]string{
						"X-Header-Remove": "this-should-be-removed",
					},
				},
				Response: grpc.Response{
					Headers: &metadata.MD{},
				},
				Backend:   "grpc-infra-backend-v1",
				Namespace: ns,
			},
			{
				TestCaseName:   "Multiple operations - all header operations should be applied",
				EchoTwoRequest: &pb.EchoRequest{},
				RequestMetadata: &grpc.RequestMetadata{
					Authority: "gateway_api_conformance.echo_basic.grpcecho.GrpcEcho",
					Metadata: map[string]string{
						"X-Header-Set-2":    "set-header-2",
						"X-Header-Add-2":    "add-header-2",
						"X-Header-Remove-2": "should-be-removed-2",
						"Some-Other-Header": "another-header-val",
					},
				},
				Response: grpc.Response{
					Headers: &metadata.MD{
						"X-Header-Set-1":    {"header-set-1"},
						"X-Header-Set-2":    {"header-set-2"},
						"X-Header-Add-1":    {"header-add-1"},
						"X-Header-Add-2":    {"header-add-1,add-header-2"},
						"Some-Other-Header": {"another-header-val"},
					},
				},
				Backend:   "grpc-infra-backend-v2",
				Namespace: ns,
			},
			{
				TestCaseName:   "Case sensitivity check for header names",
				EchoTwoRequest: &pb.EchoRequest{},
				RequestMetadata: &grpc.RequestMetadata{
					Authority: "gateway_api_conformance.echo_basic.grpcecho.GrpcEcho",
					// The filter uses canonicalized header names,
					// the request uses lowercase names.
					Metadata: map[string]string{
						"x-header-set-1":    "original-set-1",
						"x-header-add-1":    "existing-add-1",
						"x-header-remove-1": "should-be-removed-1",
					},
				},
				Response: grpc.Response{
					Headers: &metadata.MD{
						"X-Header-Set-1": {"set-overwrites-values"},
						"X-Header-Set-2": {"header-set-2"},
						"X-Header-Add-1": {"existing-add-1", "add-appends-values"},
						"X-Header-Add-2": {"header-add-2"},
					},
				},
				Backend:   "grpc-infra-backend-v2",
				Namespace: ns,
			},
		}

		for i := range testCases {
			tc := testCases[i]
			t.Run(tc.TestCaseName, func(t *testing.T) {
				t.Parallel()
				grpc.MakeRequestAndExpectEventuallyConsistentResponse(t, suite.GRPCClient, suite.TimeoutConfig, gwAddr, tc)
			})
		}
	},
}
