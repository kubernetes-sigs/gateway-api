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
	ConformanceTests = append(ConformanceTests, ListenerSetUpdateListenerHostname)
}

var ListenerSetUpdateListenerHostname = suite.ConformanceTest{
	ShortName:   "ListenerSetUpdateListenerHostname",
	Description: "An HTTPRoute rejected with NoMatchingListenerHostname must have its status updated to Accepted=True after the ListenerSet listener hostname is changed to match the route's hostname.",
	Features: []features.FeatureName{
		features.SupportGateway,
		features.SupportListenerSet,
		features.SupportHTTPRoute,
	},
	Manifests: []string{"tests/listenerset-update-listener-hostname.yaml"},
	Test: func(t *testing.T, s *suite.ConformanceTestSuite) {
		t.Run("should update HTTPRoute Accepted condition after ListenerSet listener hostname is changed to match", func(t *testing.T) {
			lsNN := types.NamespacedName{Name: "listenerset-update-listener-hostname", Namespace: suite.InfrastructureNamespace}
			routeNN := types.NamespacedName{Name: "httproute-listenerset-update-listener-hostname", Namespace: suite.InfrastructureNamespace}
			kubernetes.NamespacesMustBeReady(t, s.Client, s.TimeoutConfig, []string{suite.InfrastructureNamespace})

			kubernetes.HTTPRouteMustHaveCondition(t, s.Client, s.TimeoutConfig, routeNN, lsNN, metav1.Condition{
				Type:   string(v1.RouteConditionAccepted),
				Status: metav1.ConditionFalse,
				Reason: string(v1.RouteReasonNoMatchingListenerHostname),
			})

			ctx, cancel := context.WithTimeout(context.Background(), s.TimeoutConfig.DefaultTestTimeout)
			defer cancel()

			original := &v1.ListenerSet{}
			err := s.Client.Get(ctx, lsNN, original)
			require.NoErrorf(t, err, "error getting ListenerSet: %v", err)

			mutate := original.DeepCopy()

			// update the listener hostname to match the HTTPRoute's hostname.
			matchingHostname := v1.Hostname("accepted.example.com")
			mutate.Spec.Listeners[0].Hostname = &matchingHostname

			err = s.Client.Patch(ctx, mutate, client.MergeFrom(original))
			require.NoErrorf(t, err, "error patching the ListenerSet: %v", err)

			kubernetes.HTTPRouteMustHaveCondition(t, s.Client, s.TimeoutConfig, routeNN, lsNN, metav1.Condition{
				Type:   string(v1.RouteConditionAccepted),
				Status: metav1.ConditionTrue,
				Reason: string(v1.RouteReasonAccepted),
			})

			kubernetes.ListenerSetStatusMustHaveListeners(t, s.Client, s.TimeoutConfig, lsNN, []v1.ListenerEntryStatus{
				{
					Name: v1.SectionName("http"),
					SupportedKinds: []v1.RouteGroupKind{{
						Group: (*v1.Group)(&v1.GroupVersion.Group),
						Kind:  v1.Kind("HTTPRoute"),
					}},
					Conditions:     generateAcceptedListenerConditions(),
					AttachedRoutes: 1,
				},
			})
		})
	},
}
