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
	"sigs.k8s.io/gateway-api/conformance/utils/kubernetes"
	"sigs.k8s.io/gateway-api/conformance/utils/suite"
)

func init() {
	ConformanceTests = append(ConformanceTests, HTTPRouteInvalidReferencePolicy)
}

var HTTPRouteInvalidReferencePolicy = suite.ConformanceTest{
	ShortName:   "HTTPRouteInvalidReferencePolicy",
	Description: "A single HTTPRoute in the gateway-conformance-infra namespace should fail to attach to a Gateway in the same namespace if the route has a backendRef Service in the gateway-conformance-app-backend namespace and a ReferencePolicy exists but does not grant permission to route to that specific Service",
	Manifests:   []string{"tests/httproute-invalid-reference-policy.yaml"},
	Test: func(t *testing.T, suite *suite.ConformanceTestSuite) {
		routeNN := types.NamespacedName{Name: "invalid-reference-policy", Namespace: "gateway-conformance-infra"}
		gwNN := types.NamespacedName{Name: "same-namespace", Namespace: "gateway-conformance-infra"}

		ns := v1alpha2.Namespace(gwNN.Namespace)
		kind := v1alpha2.Kind("Gateway")

		t.Run("Route status should have a route parent status with an Accepted condition set to False", func(t *testing.T) {
			parents := []v1alpha2.RouteParentStatus{{
				ParentRef: v1alpha2.ParentReference{
					Group:     (*v1alpha2.Group)(&v1alpha2.GroupVersion.Group),
					Kind:      &kind,
					Name:      v1alpha2.ObjectName(gwNN.Name),
					Namespace: &ns,
				},
				ControllerName: v1alpha2.GatewayController(suite.ControllerName),
				Conditions: []metav1.Condition{{
					Type:   string(v1alpha2.ConditionRouteAccepted),
					Status: metav1.ConditionFalse,
				}},
			}}

			kubernetes.HTTPRouteMustHaveParents(t, suite.Client, routeNN, parents, true, 60)
		})

		t.Run("Gateway should have 0 Routes attached", func(t *testing.T) {
			gw := &v1alpha2.Gateway{}
			err := suite.Client.Get(context.TODO(), gwNN, gw)
			require.NoError(t, err, "error fetching Gateway")
			// There are two valid ways to represent this:
			// 1. No listeners in status
			// 2. One listener in status with 0 attached routes
			if len(gw.Status.Listeners) == 0 {
				// No listeners in status.
			} else if len(gw.Status.Listeners) == 1 {
				require.Equal(t, int32(0), gw.Status.Listeners[0].AttachedRoutes)
			} else {
				t.Errorf("Expected no more than 1 listener in status, got %d", len(gw.Status.Listeners))
			}
		})
	},
}
