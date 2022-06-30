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
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	"sigs.k8s.io/gateway-api/apis/v1alpha2"
	"sigs.k8s.io/gateway-api/conformance/utils/http"
	"sigs.k8s.io/gateway-api/conformance/utils/kubernetes"
	"sigs.k8s.io/gateway-api/conformance/utils/suite"
)

func init() {
	ConformanceTests = append(ConformanceTests, HTTPRouteInvalidCrossNamespaceBackendRef)
}

var HTTPRouteInvalidBackendRefUnknownKind = suite.ConformanceTest{
	ShortName:   "HTTPRouteInvalidBackendRefUnknownKind",
	Description: "A single HTTPRoute in the gateway-conformance-infra namespace should set a ResolvedRefs status False with reason RefNotPermitted when attempting to bind to a Gateway in the same namespace if the route has a BackendRef that points to an unknown Kind.",
	Exemptions: []suite.ExemptFeature{
		suite.ExemptReferenceGrant,
	},
	Manifests: []string{"tests/httproute-invalid-cross-namespace-backend-ref.yaml"},
	Test: func(t *testing.T, suite *suite.ConformanceTestSuite) {
		routeNN := types.NamespacedName{Name: "invalid-cross-namespace-backend-ref", Namespace: "gateway-conformance-infra"}
		gwNN := types.NamespacedName{Name: "same-namespace", Namespace: "gateway-conformance-infra"}

		ns := v1alpha2.Namespace(gwNN.Namespace)
		kind := v1alpha2.Kind("Gateway")

		// TODO(mikemorris): Add check for Accepted condition once
		// https://github.com/kubernetes-sigs/gateway-api/issues/1112
		// has been resolved
		t.Run("Route status should have a route parent status with a ResolvedRefs condition with status False and reason RefNotPermitted", func(t *testing.T) {
			parents := []v1alpha2.RouteParentStatus{{
				ParentRef: v1alpha2.ParentReference{
					Group:     (*v1alpha2.Group)(&v1alpha2.GroupVersion.Group),
					Kind:      &kind,
					Name:      v1alpha2.ObjectName(gwNN.Name),
					Namespace: &ns,
				},
				ControllerName: v1alpha2.GatewayController(suite.ControllerName),
				Conditions: []metav1.Condition{{
					Type:   string(v1alpha2.RouteConditionResolvedRefs),
					Status: metav1.ConditionFalse,
					Reason: string(v1alpha2.RouteReasonInvalidKind),
				}},
			}}

			kubernetes.HTTPRouteMustHaveParents(t, suite.Client, routeNN, parents, false, 60)
		})

		// TODO(mikemorris): Add check for Listener attached routes or
		// Listener ResolvedRefs RefNotPermitted condition once
		// https://github.com/kubernetes-sigs/gateway-api/issues/1112
		// has been resolved

		gwAddr := kubernetes.GatewayAndHTTPRoutesMustBeReady(t, suite.Client, suite.ControllerName, kubernetes.NewGatewayRef(gwNN))
		t.Run("HTTP Request to invalid cross-namespace backend receive a 500", func(t *testing.T) {
			http.MakeRequestAndExpectEventuallyConsistentResponse(t, suite.RoundTripper, gwAddr, http.ExpectedResponse{
				Request: http.ExpectedRequest{
					Method: "GET",
					Path:   "/v2",
				},
				StatusCode: 500,
			})
		})

	},
}
