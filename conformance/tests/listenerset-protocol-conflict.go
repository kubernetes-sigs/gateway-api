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

	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	gatewayv1 "sigs.k8s.io/gateway-api/apis/v1"
	gatewayxv1a1 "sigs.k8s.io/gateway-api/apisx/v1alpha1"
	"sigs.k8s.io/gateway-api/conformance/utils/http"
	"sigs.k8s.io/gateway-api/conformance/utils/kubernetes"
	"sigs.k8s.io/gateway-api/conformance/utils/suite"
	"sigs.k8s.io/gateway-api/pkg/features"
)

func init() {
	ConformanceTests = append(ConformanceTests, ListenerSetProtocolConflict)
}

var ListenerSetProtocolConflict = suite.ConformanceTest{
	ShortName:   "ListenerSetProtocolConflict",
	Description: "Listener Set listener with protocol conflict with a Gateway listener",
	Features: []features.FeatureName{
		features.SupportGateway,
		features.SupportGatewayListenerSet,
		features.SupportHTTPRoute,
	},
	Manifests: []string{
		"tests/listenerset-protocol-conflict.yaml",
	},
	Test: func(t *testing.T, suite *suite.ConformanceTestSuite) {
		ns := "gateway-conformance-infra"

		kubernetes.NamespacesMustBeReady(t, suite.Client, suite.TimeoutConfig, []string{ns})

		testCases := []http.ExpectedResponse{
			// Requests to the listeners without protocol conflict should work
			{
				Request:   http.Request{Host: "gateway.com", Path: "/gateway-route"},
				Backend:   "infra-backend-v1",
				Namespace: ns,
			},
			{
				Request:   http.Request{Host: "listenerset.com", Path: "/gateway-route"},
				Backend:   "infra-backend-v1",
				Namespace: ns,
			},
		}

		gwNN := types.NamespacedName{Name: "gateway-with-listenerset-http-listener", Namespace: ns}
		gwRoutes := []types.NamespacedName{
			{Name: "attaches-to-all-listeners", Namespace: ns},
		}

		gwAddr, err := kubernetes.WaitForGatewayAddress(t, suite.Client, suite.TimeoutConfig, kubernetes.NewGatewayRef(gwNN))
		require.NoErrorf(t, err, "timed out waiting for Gateway address to be assigned")
		for _, routeNN := range gwRoutes {
			kubernetes.HTTPRouteMustHaveResolvedRefsConditionsTrue(t, suite.Client, suite.TimeoutConfig, routeNN, gwNN)
		}

		acceptedListenerConditions := []metav1.Condition{
			{
				Type:   string(gatewayv1.ListenerConditionResolvedRefs),
				Status: metav1.ConditionTrue,
				Reason: "", // any reason
			},
			{
				Type:   string(gatewayv1.ListenerConditionAccepted),
				Status: metav1.ConditionTrue,
				Reason: "", // any reason
			},
			{
				Type:   string(gatewayv1.ListenerConditionProgrammed),
				Status: metav1.ConditionTrue,
				Reason: "", // any reason
			},
		}
		conflictedListenerConditions := []metav1.Condition{
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

		kubernetes.GatewayMustHaveCondition(t, suite.Client, suite.TimeoutConfig, gwNN, metav1.Condition{
			Type:   string(gatewayv1.GatewayConditionAttachedListenerSets),
			Status: metav1.ConditionTrue,
			Reason: string(gatewayv1.GatewayReasonListenerSetsAttached),
		})
		kubernetes.GatewayListenerMustHaveConditions(t, suite.Client, suite.TimeoutConfig, gwNN, "gateway-com", acceptedListenerConditions)
		kubernetes.GatewayListenerMustHaveConditions(t, suite.Client, suite.TimeoutConfig, gwNN, "protocol-conflict", conflictedListenerConditions)

		lsNN := types.NamespacedName{Name: "listenerset-with-http-listener", Namespace: ns}
		kubernetes.ListenerSetMustHaveCondition(t, suite.Client, suite.TimeoutConfig, lsNN, metav1.Condition{
			Type:   string(gatewayxv1a1.ListenerSetConditionAccepted),
			Status: metav1.ConditionTrue,
			Reason: string(gatewayxv1a1.ListenerSetReasonListenersNotValid),
		})
		kubernetes.ListenerSetMustHaveCondition(t, suite.Client, suite.TimeoutConfig, lsNN, metav1.Condition{
			Type:   string(gatewayxv1a1.ListenerSetConditionProgrammed),
			Status: metav1.ConditionTrue,
			Reason: string(gatewayxv1a1.ListenerSetReasonListenersNotValid),
		})
		kubernetes.ListenerSetListenerMustHaveConditions(t, suite.Client, suite.TimeoutConfig, lsNN, "listenerset-com", acceptedListenerConditions)
		kubernetes.ListenerSetListenerMustHaveConditions(t, suite.Client, suite.TimeoutConfig, lsNN, "protocol-conflict", conflictedListenerConditions)

		for i := range testCases {
			// Declare tc here to avoid loop variable
			// reuse issues across parallel tests.
			tc := testCases[i]
			t.Run(tc.GetTestCaseName(i), func(t *testing.T) {
				t.Parallel()
				http.MakeRequestAndExpectEventuallyConsistentResponse(t, suite.RoundTripper, suite.TimeoutConfig, gwAddr, tc)
			})
		}
	},
}
