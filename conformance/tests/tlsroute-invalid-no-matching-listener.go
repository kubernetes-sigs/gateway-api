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

	v1 "sigs.k8s.io/gateway-api/apis/v1"
	"sigs.k8s.io/gateway-api/conformance/utils/kubernetes"
	"sigs.k8s.io/gateway-api/conformance/utils/suite"
	"sigs.k8s.io/gateway-api/pkg/features"
)

func init() {
	ConformanceTests = append(ConformanceTests, TLSRouteInvalidNoMatchingListener)
}

var TLSRouteInvalidNoMatchingListener = suite.ConformanceTest{
	ShortName:   "TLSRouteInvalidNoMatchingListener",
	Description: "A TLSRoute should set Accepted=False when attaching to a Gateway with no compatible TLS listener",
	Features: []features.FeatureName{
		features.SupportGateway,
		features.SupportTLSRoute,
	},
	Manifests: []string{"tests/tlsroute-invalid-no-matching-listener.yaml"},
	Test: func(t *testing.T, suite *suite.ConformanceTestSuite) {
		routeNoMatchingPortNN := types.NamespacedName{Name: "tlsroute-no-matching-listener", Namespace: "gateway-conformance-infra"}
		routeNotAllowedKindNN := types.NamespacedName{Name: "tlsroute-not-allowed-kind", Namespace: "gateway-conformance-infra"}
		routeWrongProtocolNN := types.NamespacedName{Name: "tlsroute-wrong-protocol", Namespace: "gateway-conformance-infra"}
		routeWrongSectionNN := types.NamespacedName{Name: "tlsroute-wrong-section-name", Namespace: "gateway-conformance-infra"}

		gwHTTPOnlyNN := types.NamespacedName{Name: "gateway-http-only", Namespace: "gateway-conformance-infra"}
		gwTLSHTTPRouteOnlyNN := types.NamespacedName{Name: "gateway-tls-httproute-only", Namespace: "gateway-conformance-infra"}
		gwHTTPSOnlyNN := types.NamespacedName{Name: "gateway-https-only", Namespace: "gateway-conformance-infra"}

		acceptedCondNoMatchingParent := metav1.Condition{
			Type:   string(v1.RouteConditionAccepted),
			Status: metav1.ConditionFalse,
			Reason: string(v1.RouteReasonNoMatchingParent),
		}
		acceptedCondNotAllowed := metav1.Condition{
			Type:   string(v1.RouteConditionAccepted),
			Status: metav1.ConditionFalse,
			Reason: string(v1.RouteReasonNotAllowedByListeners),
		}
		t.Run("TLSRoute rejected when listener protocol is HTTP", func(t *testing.T) {
			kubernetes.TLSRouteMustHaveCondition(t, suite.Client, suite.TimeoutConfig, routeNoMatchingPortNN, gwHTTPOnlyNN, acceptedCondNotAllowed)
		})
		t.Run("TLSRoute rejected when kind not allowed", func(t *testing.T) {
			kubernetes.TLSRouteMustHaveCondition(t, suite.Client, suite.TimeoutConfig, routeNotAllowedKindNN, gwTLSHTTPRouteOnlyNN, acceptedCondNotAllowed)
		})
		t.Run("TLSRoute rejected when listener protocol is HTTPS", func(t *testing.T) {
			kubernetes.TLSRouteMustHaveCondition(t, suite.Client, suite.TimeoutConfig, routeWrongProtocolNN, gwHTTPSOnlyNN, acceptedCondNotAllowed)
		})
		t.Run("TLSRoute rejected when sectionName not found", func(t *testing.T) {
			kubernetes.TLSRouteMustHaveCondition(t, suite.Client, suite.TimeoutConfig, routeWrongSectionNN, gwHTTPOnlyNN, acceptedCondNoMatchingParent)
		})
		t.Run("TLSRoute (HTTP listener) should not have Parents accepted", func(t *testing.T) {
			kubernetes.TLSRouteMustHaveNoAcceptedParents(t, suite.Client, suite.TimeoutConfig, routeNoMatchingPortNN)
		})
		t.Run("TLSRoute (kind not allowed) should not have Parents accepted", func(t *testing.T) {
			kubernetes.TLSRouteMustHaveNoAcceptedParents(t, suite.Client, suite.TimeoutConfig, routeNotAllowedKindNN)
		})
		t.Run("TLSRoute (HTTPS listener) should not have Parents accepted", func(t *testing.T) {
			kubernetes.TLSRouteMustHaveNoAcceptedParents(t, suite.Client, suite.TimeoutConfig, routeWrongProtocolNN)
		})
		t.Run("TLSRoute (wrong sectionName) should not have Parents accepted", func(t *testing.T) {
			kubernetes.TLSRouteMustHaveNoAcceptedParents(t, suite.Client, suite.TimeoutConfig, routeWrongSectionNN)
		})
		t.Run("Gateway HTTP-only should have 0 Routes attached", func(t *testing.T) {
			kubernetes.GatewayMustHaveZeroRoutes(t, suite.Client, suite.TimeoutConfig, gwHTTPOnlyNN)
		})
		t.Run("Gateway TLS-HTTPRoute-only should have 0 Routes attached", func(t *testing.T) {
			kubernetes.GatewayMustHaveZeroRoutes(t, suite.Client, suite.TimeoutConfig, gwTLSHTTPRouteOnlyNN)
		})
		t.Run("Gateway HTTPS-only should have 0 Routes attached", func(t *testing.T) {
			kubernetes.GatewayMustHaveZeroRoutes(t, suite.Client, suite.TimeoutConfig, gwHTTPSOnlyNN)
		})
	},
}
