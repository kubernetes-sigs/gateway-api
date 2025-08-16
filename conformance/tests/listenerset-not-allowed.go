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

	gatewayv1 "sigs.k8s.io/gateway-api/apis/v1"
	gatewayxv1a1 "sigs.k8s.io/gateway-api/apisx/v1alpha1"
	"sigs.k8s.io/gateway-api/conformance/utils/http"
	"sigs.k8s.io/gateway-api/conformance/utils/kubernetes"
	"sigs.k8s.io/gateway-api/conformance/utils/suite"
	"sigs.k8s.io/gateway-api/pkg/features"
)

func init() {
	ConformanceTests = append(ConformanceTests, ListenerSetNotAllowed)
}

var ListenerSetNotAllowed = suite.ConformanceTest{
	ShortName:   "ListenerSetNotAllowed",
	Description: "Listener Set not allowed on the Gateway",
	Features: []features.FeatureName{
		features.SupportGateway,
		features.SupportGatewayListenerSet,
		features.SupportHTTPRoute,
	},
	Manifests: []string{
		"tests/listenerset-not-allowed.yaml",
	},
	Test: func(t *testing.T, suite *suite.ConformanceTestSuite) {
		ns := "gateway-conformance-infra"

		kubernetes.NamespacesMustBeReady(t, suite.Client, suite.TimeoutConfig, []string{ns})

		testCases := []http.ExpectedResponse{
			// Requests to the listener defined on the gateway should work
			{
				Request:   http.Request{Host: "example.com", Path: "/gateway-route"},
				Backend:   "infra-backend-v1",
				Namespace: ns,
			},
			// Requests to the listenerset listeners should fail
			{
				Request:  http.Request{Host: "foo.com", Path: "/gateway-route"},
				Response: http.Response{StatusCode: 404},
			},
			{
				Request:  http.Request{Host: "bar.com", Path: "/gateway-route"},
				Response: http.Response{StatusCode: 404},
			},
		}

		gwNN := types.NamespacedName{Name: "gateway-does-not-allow-listenerset", Namespace: ns}
		gwRoutes := []types.NamespacedName{
			{Name: "attaches-to-all-listeners", Namespace: ns},
		}

		gwAddr := kubernetes.GatewayAndHTTPRoutesMustBeAccepted(t, suite.Client, suite.TimeoutConfig, suite.ControllerName, kubernetes.NewGatewayRef(gwNN), gwRoutes...)
		for _, routeNN := range gwRoutes {
			kubernetes.HTTPRouteMustHaveResolvedRefsConditionsTrue(t, suite.Client, suite.TimeoutConfig, routeNN, gwNN)
		}

		kubernetes.GatewayMustHaveCondition(t, suite.Client, suite.TimeoutConfig, gwNN, metav1.Condition{
			Type:   string(gatewayv1.GatewayConditionAttachedListenerSets),
			Status: metav1.ConditionUnknown,
			Reason: string(gatewayv1.GatewayReasonListenerSetsNotAllowed),
		})
		disallowedLsNN := types.NamespacedName{Name: "listenerset-not-allowed", Namespace: ns}
		kubernetes.ListenerSetMustHaveCondition(t, suite.Client, suite.TimeoutConfig, disallowedLsNN, metav1.Condition{
			Type:   string(gatewayxv1a1.ListenerSetConditionAccepted),
			Status: metav1.ConditionFalse,
			Reason: string(gatewayxv1a1.ListenerSetReasonNotAllowed),
		})
		kubernetes.ListenerSetMustHaveCondition(t, suite.Client, suite.TimeoutConfig, disallowedLsNN, metav1.Condition{
			Type:   string(gatewayxv1a1.ListenerSetConditionProgrammed),
			Status: metav1.ConditionFalse,
			Reason: string(gatewayxv1a1.ListenerSetReasonNotAllowed),
		})

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
