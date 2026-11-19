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
	"net"
	"testing"

	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	gatewayv1 "sigs.k8s.io/gateway-api/apis/v1"
	"sigs.k8s.io/gateway-api/conformance/utils/kubernetes"
	confsuite "sigs.k8s.io/gateway-api/conformance/utils/suite"
	"sigs.k8s.io/gateway-api/conformance/utils/tcp"
	"sigs.k8s.io/gateway-api/conformance/utils/tls"
	"sigs.k8s.io/gateway-api/pkg/features"
)

func init() {
	ConformanceTests = append(ConformanceTests, ListenerSetProtocolAndRouteKindConflicts)
}

var ListenerSetProtocolAndRouteKindConflicts = confsuite.ConformanceTest{
	ShortName:   "ListenerSetProtocolAndRouteKindConflicts",
	Description: "ListenerSet listeners allow specific route kinds and honor mixed-protocol listener precedence",
	Features: []features.FeatureName{
		features.SupportGateway,
		features.SupportListenerSet,
		features.SupportTLSRoute,
	},
	Manifests: []string{
		"tests/listenerset-protocol-and-route-kind-conflicts.yaml",
	},
	Parallel: true,
	Test: func(t *testing.T, suite *confsuite.ConformanceTestSuite) {
		ns := confsuite.InfrastructureNamespace
		kubernetes.NamespacesMustBeReady(t, suite.Client, suite.TimeoutConfig, []string{ns})

		gwNN := types.NamespacedName{Name: "gateway-with-listener-sets-test-supported-route-kinds", Namespace: ns}
		t.Run("Gateway is accepted", func(t *testing.T) {
			kubernetes.GatewayMustHaveCondition(t, suite.Client, suite.TimeoutConfig, gwNN, metav1.Condition{
				Type:   string(gatewayv1.GatewayConditionAccepted),
				Status: metav1.ConditionTrue,
			})
		})

		lsNN := types.NamespacedName{Name: "listenerset-test-allowed-routes-supported-kinds", Namespace: ns}
		t.Run("ListenerSet listener has the expected conditions", func(t *testing.T) {
			kubernetes.ListenerSetListenersMustHaveConditions(t, suite.Client, suite.TimeoutConfig, lsNN,
				[]metav1.Condition{
					{
						Type:   string(gatewayv1.ListenerConditionResolvedRefs),
						Status: metav1.ConditionFalse,
						Reason: string(gatewayv1.ListenerReasonInvalidRouteKinds),
					},
				}, "listener-set-listener-allowed-routes-tls-only")
		})

		controlRouteNN := types.NamespacedName{Name: "listenerset-supported-kinds-control", Namespace: ns}
		kubernetes.TLSRouteMustHaveCondition(t, suite.Client, suite.TimeoutConfig, controlRouteNN, lsNN, metav1.Condition{
			Type:   string(gatewayv1.RouteConditionAccepted),
			Status: metav1.ConditionTrue,
			Reason: string(gatewayv1.RouteReasonAccepted),
		})
		kubernetes.TLSRouteMustHaveResolvedRefsConditionsTrue(t, suite.Client, suite.TimeoutConfig, controlRouteNN, lsNN)
		kubernetes.ListenerSetListenersMustHaveConditions(t, suite.Client, suite.TimeoutConfig, lsNN, generateAcceptedListenerConditions(), "tls-control")

		kubernetes.ListenerSetMustHaveCondition(t, suite.Client, suite.TimeoutConfig, lsNN, metav1.Condition{
			Type:   string(gatewayv1.ListenerSetConditionAccepted),
			Status: metav1.ConditionTrue,
		})
		kubernetes.ListenerSetMustHaveCondition(t, suite.Client, suite.TimeoutConfig, lsNN, metav1.Condition{
			Type:   string(gatewayv1.ListenerSetConditionProgrammed),
			Status: metav1.ConditionTrue,
		})

		conflictedConditions := []metav1.Condition{
			{
				Type:   string(gatewayv1.ListenerConditionAccepted),
				Status: metav1.ConditionFalse,
				Reason: string(gatewayv1.ListenerReasonProtocolConflict),
			},
			{
				Type:   string(gatewayv1.ListenerConditionProgrammed),
				Status: metav1.ConditionFalse,
				Reason: string(gatewayv1.ListenerReasonProtocolConflict),
			},
			{
				Type:   string(gatewayv1.ListenerConditionConflicted),
				Status: metav1.ConditionTrue,
				Reason: string(gatewayv1.ListenerReasonProtocolConflict),
			},
		}
		kubernetes.ListenerSetListenersMustHaveConditions(t, suite.Client, suite.TimeoutConfig, lsNN, conflictedConditions, "tls-conflict")
		kubernetes.GatewayListenersMustHaveConditions(t, suite.Client, suite.TimeoutConfig, gwNN, generateAcceptedListenerConditions(), "https-conflict")

		gwAddr, err := kubernetes.WaitForGatewayAddress(t, suite.Client, suite.TimeoutConfig, kubernetes.NewGatewayRef(gwNN, "https-conflict"))
		require.NoError(t, err, "timed out waiting for Gateway address to be assigned")
		gwIP, _, err := net.SplitHostPort(gwAddr)
		require.NoErrorf(t, err, "failed to split Gateway address %q", gwAddr)

		caCertNN := types.NamespacedName{Name: "tls-checks-ca-certificate", Namespace: ns}
		caConfigMap, err := kubernetes.GetConfigMapData(suite.Client, suite.TimeoutConfig, caCertNN)
		require.NoError(t, err, "failed to get TLS backend CA certificate")
		caCert, ok := caConfigMap["ca.crt"]
		require.True(t, ok, "ca.crt not found in TLS backend CA ConfigMap")

		t.Run("ListenerSet TLS control listener serves traffic", func(t *testing.T) {
			tcp.MakeTCPRequestAndExpectEventuallyValidResponse(t, suite.TimeoutConfig, net.JoinHostPort(gwIP, "9443"), []byte(caCert), "abc.example.com", true, tcp.ExpectedResponse{
				BackendIsTLS: true,
				Backend:      "tcp-backend",
				Namespace:    ns,
				Hostname:     "abc.example.com",
			})
		})

		conflictedAddr := net.JoinHostPort(gwIP, "8443")

		t.Run("Conflicted ListenerSet TLS listener does not reach its backend", func(t *testing.T) {
			tls.MakeTLSConnectionAndExpectEventuallyConsistentFailure(t, suite.TimeoutConfig, conflictedAddr, []byte(caCert), "abc.example.com")
		})
	},
}
