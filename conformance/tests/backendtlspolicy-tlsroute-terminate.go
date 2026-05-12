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

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	gatewayv1 "sigs.k8s.io/gateway-api/apis/v1"
	"sigs.k8s.io/gateway-api/conformance/utils/kubernetes"
	"sigs.k8s.io/gateway-api/conformance/utils/suite"
	"sigs.k8s.io/gateway-api/conformance/utils/tcp"
	"sigs.k8s.io/gateway-api/pkg/features"
)

func init() {
	ConformanceTests = append(ConformanceTests, BackendTLSPolicyTLSRouteTerminate)
}

var BackendTLSPolicyTLSRouteTerminate = suite.ConformanceTest{
	ShortName:   "BackendTLSPolicyTLSRouteTerminate",
	Description: "A BackendTLSPolicy attached to a Service consumed by a TLSRoute using Terminate mode should re-encrypt traffic to the backend",
	Features: []features.FeatureName{
		features.SupportGateway,
		features.SupportTLSRoute,
		features.SupportTLSRouteModeTerminate,
		features.SupportBackendTLSPolicy,
	},
	Provisional: true,
	Manifests:   []string{"tests/backendtlspolicy-tlsroute-terminate.yaml"},
	Test: func(t *testing.T, suite *suite.ConformanceTestSuite) {
		ns := "gateway-conformance-infra"
		routeNN := types.NamespacedName{Name: "backendtlspolicy-tlsroute-terminate", Namespace: ns}
		gwNN := types.NamespacedName{Name: "gateway-btls-tlsroute-terminate", Namespace: ns}
		policyNN := types.NamespacedName{Name: "btls-tlsroute-terminate-test", Namespace: ns}
		caCertNN := types.NamespacedName{Name: "tls-checks-ca-certificate", Namespace: ns}

		kubernetes.NamespacesMustBeReady(t, suite.Client, suite.TimeoutConfig, []string{ns})

		gwAddr, hostnames := kubernetes.GatewayAndTLSRoutesMustBeAccepted(t, suite.Client, suite.TimeoutConfig, suite.ControllerName, kubernetes.NewGatewayRef(gwNN), routeNN)
		if len(hostnames) != 1 {
			t.Fatalf("unexpected error in test configuration, found %d hostnames", len(hostnames))
		}
		serverStr := string(hostnames[0])

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
		kubernetes.BackendTLSPolicyMustHaveCondition(t, suite.Client, suite.TimeoutConfig, policyNN, gwNN, acceptedCond)
		kubernetes.BackendTLSPolicyMustHaveCondition(t, suite.Client, suite.TimeoutConfig, policyNN, gwNN, resolvedRefsCond)

		caConfigMap, err := kubernetes.GetConfigMapData(suite.Client, suite.TimeoutConfig, caCertNN)
		if err != nil {
			t.Fatalf("unexpected error finding CA certificate ConfigMap: %v", err)
		}
		caString, ok := caConfigMap["ca.crt"]
		if !ok {
			t.Fatalf("ca.crt not found in configmap: %s/%s", caCertNN.Namespace, caCertNN.Name)
		}

		t.Run("TLS request matching terminated TLSRoute with BackendTLSPolicy should reach backend with re-encrypted TLS", func(t *testing.T) {
			tcp.MakeTCPRequestAndExpectEventuallyValidResponse(t, suite.TimeoutConfig, gwAddr, []byte(caString), serverStr, true,
				tcp.ExpectedResponse{
					BackendIsTLS: true,
					Backend:      "btls-terminate-tcp-backend",
					Namespace:    "gateway-conformance-infra",
					Hostname:     "abc.example.com",
				})
		})
	},
}
