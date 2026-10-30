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

package tests

import (
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	client "sigs.k8s.io/controller-runtime/pkg/client"

	gatewayv1 "sigs.k8s.io/gateway-api/apis/v1"
	h "sigs.k8s.io/gateway-api/conformance/utils/http"
	"sigs.k8s.io/gateway-api/conformance/utils/kubernetes"
	confsuite "sigs.k8s.io/gateway-api/conformance/utils/suite"
	"sigs.k8s.io/gateway-api/pkg/features"
)

func init() {
	ConformanceTests = append(ConformanceTests, BackendTLSPolicyClusterTrustBundle)
}

var BackendTLSPolicyClusterTrustBundle = confsuite.ConformanceTest{
	ShortName:   "BackendTLSPolicyClusterTrustBundle",
	Description: "A BackendTLSPolicy that references a ClusterTrustBundle for validation should verify the backend certificate using the trust anchor contained in the bundle",
	Features: []features.FeatureName{
		features.SupportGateway,
		features.SupportHTTPRoute,
		features.SupportBackendTLSPolicy,
		features.SupportClusterTrustBundle,
	},
	Manifests: []string{"tests/backendtlspolicy-clustertrustbundle.yaml"},
	Parallel:  true,
	Test: func(t *testing.T, suite *confsuite.ConformanceTestSuite) {
		ns := confsuite.InfrastructureNamespace
		routeNN := types.NamespacedName{Name: "backendtlspolicy-clustertrustbundle-test", Namespace: ns}
		gwNN := types.NamespacedName{Name: "same-namespace", Namespace: ns}

		kubernetes.NamespacesMustBeReady(t, suite.Client, suite.TimeoutConfig, []string{ns})

		// Build a ClusterTrustBundle from the CA certificate that signed the
		// tls-backend serving certificate and apply it. This exercises the
		// ClusterTrustBundle reference feature rather than a ConfigMap.
		bundle := kubernetes.MustCreateClusterTrustBundleFromCAConfigMap(t, suite.Client, suite.TimeoutConfig, "backendtlspolicy-cluster-trust-bundle", types.NamespacedName{Name: "tls-checks-ca-certificate", Namespace: ns})
		suite.Applier.MustApplyObjectsWithCleanup(t, suite.Client, suite.TimeoutConfig, []client.Object{bundle}, suite.CleanupTestResources)

		gwAddr := kubernetes.GatewayAndRoutesMustBeAccepted(t, suite.Client, suite.TimeoutConfig, suite.ControllerName, kubernetes.NewGatewayRef(gwNN), &gatewayv1.HTTPRoute{}, false, routeNN)
		kubernetes.HTTPRouteMustHaveResolvedRefsConditionsTrue(t, suite.Client, suite.TimeoutConfig, routeNN, gwNN)

		policyCond := metav1.Condition{
			Type:   string(gatewayv1.PolicyConditionAccepted),
			Status: metav1.ConditionTrue,
			Reason: string(gatewayv1.PolicyReasonAccepted),
		}
		serverStr := "abc.example.com"

		// Verify that the request sent to a Service targeted by a BackendTLSPolicy
		// whose validation references a ClusterTrustBundle and matches the backend
		// certificate SAN should succeed.
		t.Run("HTTP request sent to Service targeted by BackendTLSPolicy referencing a ClusterTrustBundle with matching hostname should succeed", func(t *testing.T) {
			policyNN := types.NamespacedName{Name: "cluster-trust-bundle-valid", Namespace: ns}
			kubernetes.BackendTLSPolicyMustHaveCondition(t, suite.Client, suite.TimeoutConfig, policyNN, gwNN, policyCond)

			h.MakeRequestAndExpectEventuallyConsistentResponse(t, suite.RoundTripper, suite.TimeoutConfig, gwAddr,
				h.ExpectedResponse{
					Namespace: ns,
					Request: h.Request{
						Host: serverStr,
						Path: "/backendtlspolicy-clustertrustbundle-valid",
						SNI:  serverStr,
					},
					Response: h.Response{StatusCodes: []int{200}},
				})
		})

		// Verify that the request sent to a Service targeted by a BackendTLSPolicy
		// whose validation references a ClusterTrustBundle but expects a hostname
		// that does not match the backend certificate should fail.
		t.Run("HTTP request sent to Service targeted by BackendTLSPolicy referencing a ClusterTrustBundle with mismatched hostname should return an HTTP error", func(t *testing.T) {
			policyNN := types.NamespacedName{Name: "cluster-trust-bundle-mismatch", Namespace: ns}
			kubernetes.BackendTLSPolicyMustHaveCondition(t, suite.Client, suite.TimeoutConfig, policyNN, gwNN, policyCond)

			h.MakeRequestAndExpectFailure(t, suite.RoundTripper, suite.TimeoutConfig, gwAddr,
				h.ExpectedResponse{
					Namespace: ns,
					Request: h.Request{
						Host: serverStr,
						Path: "/backendtlspolicy-clustertrustbundle-mismatch",
						SNI:  serverStr,
					},
				})
		})
	},
}
