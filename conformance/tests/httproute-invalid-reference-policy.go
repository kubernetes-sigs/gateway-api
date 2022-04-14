/*
Copyright 2022 The Kubernetes Authors.

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

	"sigs.k8s.io/gateway-api/apis/v1alpha2"
	"sigs.k8s.io/gateway-api/conformance/utils/http"
	"sigs.k8s.io/gateway-api/conformance/utils/kubernetes"
	"sigs.k8s.io/gateway-api/conformance/utils/suite"
)

func init() {
	ConformanceTests = append(ConformanceTests, HTTPRouteInvalidReferencePolicy)
}

var HTTPRouteInvalidReferencePolicy = suite.ConformanceTest{
	ShortName:   "HTTPRouteInvalidReferencePolicy",
	Description: "A single HTTPRoute in the gateway-conformance-infra namespace should fail to attach to a Gateway in the same namespace if the route has a backendRef Service in the gateway-conformance-app-backend namespace and a ReferencePolicy exists but does not grant permission to route to that specific Service",
	Features: []suite.SupportedFeature{
		suite.SupportReferencePolicy,
	},
	Manifests: []string{"tests/httproute-invalid-reference-policy.yaml"},
	Test: func(t *testing.T, s *suite.ConformanceTestSuite) {
		routeNN := types.NamespacedName{Name: "invalid-reference-policy", Namespace: "gateway-conformance-infra"}
		gwNN := types.NamespacedName{Name: "same-namespace", Namespace: "gateway-conformance-infra"}

		ns := v1alpha2.Namespace(gwNN.Namespace)
		gwKind := v1alpha2.Kind("Gateway")

		// TODO(mikemorris): Add check for Accepted condition once
		// https://github.com/kubernetes-sigs/gateway-api/issues/1112
		// has been resolved
		t.Run("Route status should have a route parent status with a ResolvedRefs condition with status False and reason RefNotPermitted", func(t *testing.T) {
			parents := []v1alpha2.RouteParentStatus{{
				ParentRef: v1alpha2.ParentReference{
					Group:     (*v1alpha2.Group)(&v1alpha2.GroupVersion.Group),
					Kind:      &gwKind,
					Name:      v1alpha2.ObjectName(gwNN.Name),
					Namespace: &ns,
				},
				ControllerName: v1alpha2.GatewayController(s.ControllerName),
				Conditions: []metav1.Condition{{
					Type:   string(v1alpha2.RouteConditionResolvedRefs),
					Status: metav1.ConditionFalse,
					Reason: string(v1alpha2.RouteReasonRefNotPermitted),
				}},
			}}

			kubernetes.HTTPRouteMustHaveParents(t, s.Client, routeNN, parents, false, 60)
		})

		t.Run("Gateway listener should have a ResolvedRefs condition with status False and reason RefNotPermitted", func(t *testing.T) {
			listeners := []v1alpha2.ListenerStatus{{
				Name: v1alpha2.SectionName("http"),
				SupportedKinds: []v1alpha2.RouteGroupKind{{
					Group: (*v1alpha2.Group)(&v1alpha2.GroupVersion.Group),
					Kind:  v1alpha2.Kind("HTTPRoute"),
				}},
				Conditions: []metav1.Condition{{
					Type:   string(v1alpha2.RouteConditionResolvedRefs),
					Status: metav1.ConditionFalse,
					Reason: string(v1alpha2.RouteReasonRefNotPermitted),
				}},
			}}

			kubernetes.GatewayStatusMustHaveListeners(t, s.Client, gwNN, listeners, 60)
		})

		gwAddr := kubernetes.GatewayAndHTTPRoutesMustBeReady(t, s.Client, s.ControllerName, gwNN, routeNN)

		t.Run("Simple HTTP request should not reach app-backend-v2", func(t *testing.T) {
			// This requires consecutive successes, so we only configure a single backend
			http.MakeRequestAndExpectEventuallyConsistentResponse(t, s.RoundTripper, gwAddr, http.ExpectedResponse{
				Request: http.ExpectedRequest{
					Method: "GET",
					Path:   "/",
				},
				StatusCode: 503,
				// TODO: should these fields be populated when the BackendRef is invalid?
				Backend:   "app-backend-v2",
				Namespace: "gateway-conformance-app-backend",
			})
		})
	},
}
