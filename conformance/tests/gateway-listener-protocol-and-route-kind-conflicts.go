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

	v1 "sigs.k8s.io/gateway-api/apis/v1"
	"sigs.k8s.io/gateway-api/conformance/utils/kubernetes"
	confsuite "sigs.k8s.io/gateway-api/conformance/utils/suite"
	"sigs.k8s.io/gateway-api/conformance/utils/tcp"
	"sigs.k8s.io/gateway-api/conformance/utils/tls"
	"sigs.k8s.io/gateway-api/pkg/features"
)

func init() {
	ConformanceTests = append(ConformanceTests, GatewayListenerProtocolAndRouteKindConflicts)
}

var GatewayListenerProtocolAndRouteKindConflicts = confsuite.ConformanceTest{
	ShortName:   "GatewayListenerProtocolAndRouteKindConflicts",
	Description: "An HTTPRoute must not attach to a listener that only allows TLSRoutes, and conflicting HTTPS and TLS listeners must not serve traffic",
	Features: []features.FeatureName{
		features.SupportGateway,
		features.SupportHTTPRoute,
		features.SupportTLSRoute,
	},
	Manifests: []string{"tests/gateway-listener-protocol-and-route-kind-conflicts.yaml"},
	Test: func(t *testing.T, suite *confsuite.ConformanceTestSuite) {
		kubernetes.NamespacesMustBeReady(t, suite.Client, suite.TimeoutConfig, []string{confsuite.InfrastructureNamespace})

		ns := confsuite.InfrastructureNamespace
		gwNN := types.NamespacedName{Name: "tlsroutes-only", Namespace: ns}
		routeNN := types.NamespacedName{Name: "disallowed-kind", Namespace: ns}
		kubernetes.HTTPRouteMustHaveResolvedRefsConditionsTrue(t, suite.Client, suite.TimeoutConfig, routeNN, gwNN)

		t.Run("Route should not have been accepted with reason NotAllowedByListeners", func(t *testing.T) {
			kubernetes.HTTPRouteMustHaveCondition(t, suite.Client, suite.TimeoutConfig, routeNN, gwNN, metav1.Condition{
				Type:   string(v1.RouteConditionAccepted),
				Status: metav1.ConditionFalse,
				Reason: string(v1.RouteReasonNotAllowedByListeners),
			})
		})
		t.Run("Route should not have Parents set in status", func(t *testing.T) {
			kubernetes.HTTPRouteMustHaveNoAcceptedParents(t, suite.Client, suite.TimeoutConfig, routeNN)
		})

		controlRouteNN := types.NamespacedName{Name: "mixed-protocol-conflict-control", Namespace: ns}
		gwAddr, hostnames := kubernetes.GatewayAndTLSRoutesMustBeAccepted(t, suite.Client, suite.TimeoutConfig, suite.ControllerName,
			kubernetes.NewGatewayRef(gwNN, "tls"), controlRouteNN)
		require.Len(t, hostnames, 1, "expected exactly one control TLSRoute hostname")

		caCertNN := types.NamespacedName{Name: "tls-checks-ca-certificate", Namespace: ns}
		caConfigMap, err := kubernetes.GetConfigMapData(suite.Client, suite.TimeoutConfig, caCertNN)
		require.NoError(t, err, "failed to get TLS backend CA certificate")
		caCert, ok := caConfigMap["ca.crt"]
		require.True(t, ok, "ca.crt not found in TLS backend CA ConfigMap")

		t.Run("Allowed TLSRoute serves traffic", func(t *testing.T) {
			tcp.MakeTCPRequestAndExpectEventuallyValidResponse(t, suite.TimeoutConfig, gwAddr, []byte(caCert), string(hostnames[0]), true, tcp.ExpectedResponse{
				BackendIsTLS: true,
				Backend:      "tcp-backend",
				Namespace:    ns,
				Hostname:     string(hostnames[0]),
			})
		})

		conflictedConditions := []metav1.Condition{
			{
				Type:   string(v1.ListenerConditionAccepted),
				Status: metav1.ConditionFalse,
				Reason: "Invalid",
			},
			{
				Type:   string(v1.ListenerConditionProgrammed),
				Status: metav1.ConditionFalse,
				Reason: "",
			},
			{
				Type:   string(v1.ListenerConditionConflicted),
				Status: metav1.ConditionTrue,
				Reason: string(v1.ListenerReasonProtocolConflict),
			},
		}
		t.Run("Gateway status reports conflicting listeners separately", func(t *testing.T) {
			kubernetes.GatewayListenerMustHaveConditions(t, suite.Client, suite.TimeoutConfig, gwNN, "conflicted-https", conflictedConditions)
			kubernetes.GatewayListenerMustHaveConditions(t, suite.Client, suite.TimeoutConfig, gwNN, "conflicted-tls", conflictedConditions)
			kubernetes.GatewayListenerMustHaveAttachedRoutes(t, suite.Client, suite.TimeoutConfig, gwNN, "tls", 1)
			kubernetes.GatewayListenerMustHaveAttachedRoutes(t, suite.Client, suite.TimeoutConfig, gwNN, "conflicted-https", 0)
			kubernetes.GatewayListenerMustHaveAttachedRoutes(t, suite.Client, suite.TimeoutConfig, gwNN, "conflicted-tls", 0)
		})

		gwIP, _, err := net.SplitHostPort(gwAddr)
		require.NoErrorf(t, err, "failed to split Gateway address %q", gwAddr)
		conflictedAddr := net.JoinHostPort(gwIP, "8443")

		httpsCertNN := types.NamespacedName{Name: "tls-validity-checks-certificate", Namespace: ns}
		httpsCert, _, err := kubernetes.GetTLSSecret(suite.Client, httpsCertNN)
		require.NoError(t, err, "failed to get HTTPS listener certificate")

		t.Run("Conflicting HTTPS listener rejects traffic", func(t *testing.T) {
			tls.MakeTLSConnectionAndExpectEventuallyConsistentFailure(t, suite.TimeoutConfig, conflictedAddr, httpsCert, "https.example.org")
		})
		t.Run("Conflicting TLS listener rejects traffic", func(t *testing.T) {
			tls.MakeTLSConnectionAndExpectEventuallyConsistentFailure(t, suite.TimeoutConfig, conflictedAddr, []byte(caCert), "abc.example.com")
		})
	},
}
