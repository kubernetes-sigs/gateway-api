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

	"google.golang.org/grpc/codes"
	"k8s.io/apimachinery/pkg/types"

	v1 "sigs.k8s.io/gateway-api/apis/v1"
	pb "sigs.k8s.io/gateway-api/conformance/echo-basic/grpcechoserver"
	"sigs.k8s.io/gateway-api/conformance/utils/grpc"
	"sigs.k8s.io/gateway-api/conformance/utils/kubernetes"
	confsuite "sigs.k8s.io/gateway-api/conformance/utils/suite"
	"sigs.k8s.io/gateway-api/pkg/features"
)

func init() {
	ConformanceTests = append(ConformanceTests, GRPCRouteAllBackendsWeightZero)
}

var GRPCRouteAllBackendsWeightZero = confsuite.ConformanceTest{
	ShortName:   "GRPCRouteAllBackendsWeightZero",
	Description: "A GRPCRoute with all backend weights set to 0 returns UNAVAILABLE",
	Manifests:   []string{"tests/grpcroute-all-backends-weight-zero.yaml"},
	Features: []features.FeatureName{
		features.SupportGateway,
		features.SupportGRPCRoute,
	},
	Test: func(t *testing.T, suite *confsuite.ConformanceTestSuite) {
		var (
			ns      = confsuite.InfrastructureNamespace
			routeNN = types.NamespacedName{Name: "all-backends-weight-zero", Namespace: ns}
			gwNN    = types.NamespacedName{Name: "same-namespace", Namespace: ns}
			gwAddr  = kubernetes.GatewayAndRoutesMustBeAccepted(t, suite.Client, suite.TimeoutConfig, suite.ControllerName, kubernetes.NewGatewayRef(gwNN), &v1.GRPCRoute{}, true, routeNN)
		)

		t.Run("Requests should return UNAVAILABLE when all backend weights are 0", func(t *testing.T) {
			grpc.MakeRequestAndExpectEventuallyConsistentResponse(t, suite.GRPCClient, suite.TimeoutConfig, gwAddr, grpc.ExpectedResponse{
				EchoRequest: &pb.EchoRequest{},
				Response: grpc.Response{
					Code: codes.Unavailable,
				},
			})
		})
	},
}
