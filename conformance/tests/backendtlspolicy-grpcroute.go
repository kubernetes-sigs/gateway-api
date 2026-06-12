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

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	gatewayv1 "sigs.k8s.io/gateway-api/apis/v1"
	pb "sigs.k8s.io/gateway-api/conformance/echo-basic/grpcechoserver"
	"sigs.k8s.io/gateway-api/conformance/utils/grpc"
	"sigs.k8s.io/gateway-api/conformance/utils/kubernetes"
	confsuite "sigs.k8s.io/gateway-api/conformance/utils/suite"
	"sigs.k8s.io/gateway-api/pkg/features"
)

func init() {
	ConformanceTests = append(ConformanceTests, BackendTLSPolicyGRPCRoute)
}

var BackendTLSPolicyGRPCRoute = confsuite.ConformanceTest{
	ShortName:   "BackendTLSPolicyGRPCRoute",
	Description: "A BackendTLSPolicy attached to a Service consumed by a GRPCRoute should configure TLS between gateway and backend",
	Features: []features.FeatureName{
		features.SupportGateway,
		features.SupportGRPCRoute,
		features.SupportBackendTLSPolicy,
	},
	Manifests: []string{"tests/backendtlspolicy-grpcroute.yaml"},
	Test: func(t *testing.T, suite *confsuite.ConformanceTestSuite) {
		ns := confsuite.InfrastructureNamespace

		acceptedCond := metav1.Condition{
			Type:   string(gatewayv1.PolicyConditionAccepted),
			Status: metav1.ConditionTrue,
			Reason: string(gatewayv1.PolicyReasonAccepted),
		}
		resolvedRefsCond := metav1.Condition{
			Type:   string(gatewayv1.BackendTLSPolicyConditionResolvedRefs),
			Status: metav1.ConditionTrue,
			Reason: string(gatewayv1.BackendTLSPolicyReasonResolvedRefs),
		}

		routeNN := types.NamespacedName{Name: "backendtlspolicy-grpcroute", Namespace: ns}
		gwNN := types.NamespacedName{Name: "same-namespace", Namespace: ns}

		kubernetes.NamespacesMustBeReady(t, suite.Client, suite.TimeoutConfig, []string{ns})
		gwAddr := kubernetes.GatewayAndRoutesMustBeAccepted(t, suite.Client, suite.TimeoutConfig, suite.ControllerName, kubernetes.NewGatewayRef(gwNN), &gatewayv1.GRPCRoute{}, true, routeNN)

		t.Run("gRPC request sent to Service with valid BackendTLSPolicy should succeed", func(t *testing.T) {
			validPolicyNN := types.NamespacedName{Name: "grpc-normative-test", Namespace: ns}
			kubernetes.BackendTLSPolicyMustHaveCondition(t, suite.Client, suite.TimeoutConfig, validPolicyNN, gwNN, acceptedCond)
			kubernetes.BackendTLSPolicyMustHaveCondition(t, suite.Client, suite.TimeoutConfig, validPolicyNN, gwNN, resolvedRefsCond)

			grpc.MakeRequestAndExpectEventuallyConsistentResponse(t, nil, suite.TimeoutConfig, gwAddr,
				grpc.ExpectedResponse{
					EchoRequest: &pb.EchoRequest{},
					RequestMetadata: &grpc.RequestMetadata{
						Authority: "abc.example.com",
					},
					Backend:   "grpc-infra-backend-v1",
					Namespace: ns,
				})
		})

		t.Run("gRPC request sent to Service targeted by BackendTLSPolicy with mismatched hostname should fail", func(t *testing.T) {
			invalidPolicyNN := types.NamespacedName{Name: "grpc-host-mismatch", Namespace: ns}
			kubernetes.BackendTLSPolicyMustHaveCondition(t, suite.Client, suite.TimeoutConfig, invalidPolicyNN, gwNN, acceptedCond)
			kubernetes.BackendTLSPolicyMustHaveCondition(t, suite.Client, suite.TimeoutConfig, invalidPolicyNN, gwNN, resolvedRefsCond)

			grpc.MakeRequestAndExpectEventuallyConsistentFailure(t, nil, suite.TimeoutConfig, gwAddr,
				grpc.ExpectedResponse{
					EchoTwoRequest: &pb.EchoRequest{},
					RequestMetadata: &grpc.RequestMetadata{
						Authority: "abc.example.com",
					},
				})
		})

		t.Run("gRPC request sent to Service targeted by BackendTLSPolicy with mismatched cert should fail", func(t *testing.T) {
			certMismatchRouteNN := types.NamespacedName{Name: "backendtlspolicy-grpcroute-cert-mismatch", Namespace: ns}
			kubernetes.GatewayAndRoutesMustBeAccepted(t, suite.Client, suite.TimeoutConfig, suite.ControllerName, kubernetes.NewGatewayRef(gwNN), &gatewayv1.GRPCRoute{}, true, certMismatchRouteNN)

			invalidCertPolicyNN := types.NamespacedName{Name: "grpc-cert-mismatch", Namespace: ns}
			kubernetes.BackendTLSPolicyMustHaveCondition(t, suite.Client, suite.TimeoutConfig, invalidCertPolicyNN, gwNN, acceptedCond)
			kubernetes.BackendTLSPolicyMustHaveCondition(t, suite.Client, suite.TimeoutConfig, invalidCertPolicyNN, gwNN, resolvedRefsCond)

			grpc.MakeRequestAndExpectEventuallyConsistentFailure(t, nil, suite.TimeoutConfig, gwAddr,
				grpc.ExpectedResponse{
					EchoRequest: &pb.EchoRequest{},
					RequestMetadata: &grpc.RequestMetadata{
						Authority: "cert-mismatch.example.com",
					},
				})
		})
	},
}
