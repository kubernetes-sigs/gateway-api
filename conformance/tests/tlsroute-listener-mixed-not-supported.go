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
	ConformanceTests = append(ConformanceTests, TLSRouteListenerMixedNotSupported)
}

var TLSRouteListenerMixedNotSupported = suite.ConformanceTest{
	ShortName:   "TLSRouteListenerMixedNotSupported",
	Description: "When TLSRoute mixed termination/passthrough listener is NOT supported, a Gateway Listener with 2 distinct TLS modes on the same port MUST have Accepted=False with Reason=ProtocolConflict",
	Features: []features.FeatureName{
		features.SupportGateway,
		features.SupportTLSRoute,
		features.SupportTLSRouteModeTerminate,
	},
	Manifests: []string{"tests/tlsroute-listener-mixed-not-supported.yaml"},
	Test: func(t *testing.T, suite *suite.ConformanceTestSuite) {
		if suite.SupportedFeatures.Has(features.SupportTLSRouteModeMixed) {
			return
		}
		gwNN := types.NamespacedName{Name: "gateway-tlsroute-mixed-unsupported", Namespace: "gateway-conformance-infra"}

		t.Run("Listener with unsupported mode Terminate must have Accepted=False with Reason=UnsupportedValue", func(t *testing.T) {
			listeners := []v1.ListenerStatus{
				{
					Name: v1.SectionName("tls-terminate"),
					Conditions: []metav1.Condition{{
						Type:   string(v1.ListenerConditionAccepted),
						Status: metav1.ConditionFalse,
						Reason: string(v1.ListenerReasonProtocolConflict),
					}},
					AttachedRoutes: 0,
				},
				{
					Name: v1.SectionName("tls-passthrough"),
					Conditions: []metav1.Condition{{
						Type:   string(v1.ListenerConditionAccepted),
						Status: metav1.ConditionFalse,
						Reason: string(v1.ListenerReasonProtocolConflict),
					}},
					AttachedRoutes: 0,
				},
			}
			kubernetes.GatewayStatusMustHaveListeners(t, suite.Client, suite.TimeoutConfig, gwNN, listeners)
		})
	},
}
