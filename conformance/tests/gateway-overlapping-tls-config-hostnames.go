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
	"sigs.k8s.io/gateway-api/conformance/utils/kubernetes"
	"sigs.k8s.io/gateway-api/conformance/utils/suite"
	"sigs.k8s.io/gateway-api/pkg/features"
)

func init() {
	ConformanceTests = append(ConformanceTests, GatewayOverlappingTLSConfigHostnames)
}

var GatewayOverlappingTLSConfigHostnames = suite.ConformanceTest{
	ShortName:   "GatewayOverlappingTLSConfigHostnames",
	Description: "A Gateway with HTTPS listeners that have overlapping hostnames must have the OverlappingTLSConfig condition set True with reason OverlappingHostnames on the affected listeners.",
	Features: []features.FeatureName{
		features.SupportGateway,
	},
	Manifests: []string{"tests/gateway-overlapping-tls-config-hostnames.yaml"},
	Test: func(t *testing.T, s *suite.ConformanceTestSuite) {
		ns := "gateway-conformance-infra"
		gwNN := types.NamespacedName{Name: "gateway-overlapping-hostnames", Namespace: ns}

		kubernetes.NamespacesMustBeReady(t, s.Client, s.TimeoutConfig, []string{ns})
		kubernetes.GatewayListenersMustHaveConditions(t, s.Client, s.TimeoutConfig, gwNN,
			[]metav1.Condition{
				{
					Type:   string(gatewayv1.ListenerSetConditionAccepted),
					Status: metav1.ConditionTrue,
					Reason: string(gatewayv1.ListenerConditionAccepted),
				},
			})

		t.Run("Gateway listeners with overlapping hostnames on the same port must have OverlappingTLSConfig condition set", func(t *testing.T) {
			kubernetes.GatewayListenersMustHaveConditions(t, s.Client, s.TimeoutConfig, gwNN,
				[]metav1.Condition{
					{
						Type:   string(gatewayv1.ListenerConditionOverlappingTLSConfig),
						Status: metav1.ConditionTrue,
						Reason: string(gatewayv1.ListenerReasonOverlappingHostnames),
					},
				},
				"foo-example-https", "bar-example-https", "wildcard-example-https", "wildcard-other-https", "wildcard-bar-other-https",
			)
		})
		t.Run("Gateway listeners without overlapping hostnames must not have OverlappingTLSConfig condition set", func(t *testing.T) {
			kubernetes.GatewayListenerMustNotHaveCondition(t, s.Client, s.TimeoutConfig, gwNN, "not-overlapping-https", gatewayv1.ListenerConditionOverlappingTLSConfig)
			kubernetes.GatewayListenerMustNotHaveCondition(t, s.Client, s.TimeoutConfig, gwNN, "not-overlapping-wildcard-https", gatewayv1.ListenerConditionOverlappingTLSConfig)
		})
		t.Run("Gateway listeners with overlapping hostnames on different ports must not have OverlappingTLSConfig condition set", func(t *testing.T) {
			kubernetes.GatewayListenerMustNotHaveCondition(t, s.Client, s.TimeoutConfig, gwNN, "not-overlapping-other-port", gatewayv1.ListenerConditionOverlappingTLSConfig)
		})
	},
}
