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
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	v1 "sigs.k8s.io/gateway-api/apis/v1"
	"sigs.k8s.io/gateway-api/conformance/utils/kubernetes"
	"sigs.k8s.io/gateway-api/conformance/utils/suite"
	"sigs.k8s.io/gateway-api/pkg/features"
)

func init() {
	ConformanceTests = append(ConformanceTests, GatewayUpdateListenerHostname)
}

var GatewayUpdateListenerHostname = suite.ConformanceTest{
	ShortName:   "GatewayUpdateListenerHostname",
	Description: "An HTTPRoute rejected with NoMatchingListenerHostname must have its status updated to Accepted=True after the Gateway listener hostname is changed to match the route's hostname.",
	Features: []features.FeatureName{
		features.SupportGateway,
		features.SupportHTTPRoute,
	},
	Manifests: []string{"tests/gateway-update-listener-hostname.yaml"},
	Test: func(t *testing.T, s *suite.ConformanceTestSuite) {
		t.Run("should update HTTPRoute Accepted condition after Gateway listener hostname is changed to match", func(t *testing.T) {
			gwNN := types.NamespacedName{Name: "gateway-update-listener-hostname", Namespace: suite.InfrastructureNamespace}
			routeNN := types.NamespacedName{Name: "httproute-update-listener-hostname", Namespace: suite.InfrastructureNamespace}
			namespaces := []string{suite.InfrastructureNamespace}
			kubernetes.NamespacesMustBeReady(t, s.Client, s.TimeoutConfig, namespaces)

			kubernetes.GatewayMustHaveLatestConditions(t, s.Client, s.TimeoutConfig, gwNN)

			kubernetes.HTTPRouteMustHaveCondition(t, s.Client, s.TimeoutConfig, routeNN, gwNN, metav1.Condition{
				Type:   string(v1.RouteConditionAccepted),
				Status: metav1.ConditionFalse,
				Reason: string(v1.RouteReasonNoMatchingListenerHostname),
			})

			ctx, cancel := context.WithTimeout(context.Background(), s.TimeoutConfig.DefaultTestTimeout)
			defer cancel()

			original := &v1.Gateway{}
			err := s.Client.Get(ctx, gwNN, original)
			require.NoErrorf(t, err, "error getting Gateway: %v", err)

			mutate := original.DeepCopy()

			// update the listener hostname to match the HTTPRoute's hostname.
			matchingHostname := v1.Hostname("accepted.example.com")
			for i, listener := range mutate.Spec.Listeners {
				if listener.Name == "http" {
					mutate.Spec.Listeners[i].Hostname = &matchingHostname
					break
				}
			}

			err = s.Client.Patch(ctx, mutate, client.MergeFrom(original))
			require.NoErrorf(t, err, "error patching the Gateway: %v", err)

			kubernetes.NamespacesMustBeReady(t, s.Client, s.TimeoutConfig, namespaces)

			kubernetes.HTTPRouteMustHaveCondition(t, s.Client, s.TimeoutConfig, routeNN, gwNN, metav1.Condition{
				Type:   string(v1.RouteConditionAccepted),
				Status: metav1.ConditionTrue,
				Reason: string(v1.RouteReasonAccepted),
			})

			kubernetes.GatewayStatusMustHaveListeners(t, s.Client, s.TimeoutConfig, gwNN, []v1.ListenerStatus{
				{
					Name: v1.SectionName("http"),
					SupportedKinds: []v1.RouteGroupKind{{
						Group: (*v1.Group)(&v1.GroupVersion.Group),
						Kind:  v1.Kind("HTTPRoute"),
					}},
					Conditions: []metav1.Condition{
						{
							Type:   string(v1.ListenerConditionAccepted),
							Status: metav1.ConditionTrue,
							Reason: "",
						},
						{
							Type:   string(v1.ListenerConditionResolvedRefs),
							Status: metav1.ConditionTrue,
							Reason: "",
						},
					},
					AttachedRoutes: 1,
				},
			})

			kubernetes.GatewayMustHaveLatestConditions(t, s.Client, s.TimeoutConfig, gwNN)
		})
	},
}
