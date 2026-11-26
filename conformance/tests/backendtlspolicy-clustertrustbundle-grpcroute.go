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

	pb "sigs.k8s.io/gateway-api-conformance-images/echo-basic/grpcechoserver"
	gatewayv1 "sigs.k8s.io/gateway-api/apis/v1"
	"sigs.k8s.io/gateway-api/conformance/utils/grpc"
	"sigs.k8s.io/gateway-api/conformance/utils/kubernetes"
	confsuite "sigs.k8s.io/gateway-api/conformance/utils/suite"
	"sigs.k8s.io/gateway-api/pkg/features"
)

func init() {
	ConformanceTests = append(ConformanceTests, BackendTLSPolicyClusterTrustBundleGRPCRoute)
}

var BackendTLSPolicyClusterTrustBundleGRPCRoute = confsuite.ConformanceTest{
	ShortName:   "BackendTLSPolicyClusterTrustBundleGRPCRoute",
	Description: "A BackendTLSPolicy that references a ClusterTrustBundle for validation should verify the gRPC backend certificate using the trust anchor contained in the bundle",
	Features: []features.FeatureName{
		features.SupportGateway,
		features.SupportGRPCRoute,
		features.SupportBackendTLSPolicy,
		features.SupportClusterTrustBundle,
	},
	Manifests: []string{"tests/backendtlspolicy-clustertrustbundle-grpcroute.yaml"},
	Parallel:  true,
	Test: func(t *testing.T, suite *confsuite.ConformanceTestSuite) {
		ns := confsuite.InfrastructureNamespace
		routeNN := types.NamespacedName{Name: "backendtlspolicy-clustertrustbundle-grpcroute", Namespace: ns}
		gwNN := types.NamespacedName{Name: "same-namespace", Namespace: ns}

		// Build a ClusterTrustBundle from the CA certificate that signed the
		// gRPC backend serving certificate and apply it. This exercises the
		// ClusterTrustBundle reference feature rather than a ConfigMap.
		mustApplyClusterTrustBundleFromCAConfigMap(t, suite, "backendtlspolicy-cluster-trust-bundle-grpcroute", "tls-checks-ca-certificate")

		kubernetes.NamespacesMustBeReady(t, suite.Client, suite.TimeoutConfig, []string{ns})

		gwAddr := kubernetes.GatewayAndRoutesMustBeAccepted(t, suite.Client, suite.TimeoutConfig, suite.ControllerName, kubernetes.NewGatewayRef(gwNN), &gatewayv1.GRPCRoute{}, true, routeNN)

		policyCond := metav1.Condition{
			Type:   string(gatewayv1.PolicyConditionAccepted),
			Status: metav1.ConditionTrue,
			Reason: string(gatewayv1.PolicyReasonAccepted),
		}
		resolvedRefsCond := metav1.Condition{
			Type:   string(gatewayv1.BackendTLSPolicyConditionResolvedRefs),
			Status: metav1.ConditionTrue,
			Reason: string(gatewayv1.BackendTLSPolicyReasonResolvedRefs),
		}

		t.Run("gRPC request sent to Service targeted by BackendTLSPolicy referencing a ClusterTrustBundle with matching hostname should succeed", func(t *testing.T) {
			policyNN := types.NamespacedName{Name: "cluster-trust-bundle-grpcroute", Namespace: ns}
			kubernetes.BackendTLSPolicyMustHaveCondition(t, suite.Client, suite.TimeoutConfig, policyNN, gwNN, policyCond)
			kubernetes.BackendTLSPolicyMustHaveCondition(t, suite.Client, suite.TimeoutConfig, policyNN, gwNN, resolvedRefsCond)

			grpc.MakeRequestAndExpectEventuallyConsistentResponse(t, nil, suite.TimeoutConfig, gwAddr,
				grpc.ExpectedResponse{
					EchoRequest: &pb.EchoRequest{},
					RequestMetadata: &grpc.RequestMetadata{
						Authority: "other.example.com",
					},
					Backend:   "grpc-infra-backend-v1",
					Namespace: ns,
				})
		})
	},
}
