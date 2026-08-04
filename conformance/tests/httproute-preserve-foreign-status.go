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
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/util/retry"

	gatewayv1 "sigs.k8s.io/gateway-api/apis/v1"
	"sigs.k8s.io/gateway-api/conformance/utils/kubernetes"
	confsuite "sigs.k8s.io/gateway-api/conformance/utils/suite"
	"sigs.k8s.io/gateway-api/pkg/features"
)

var errForeignStatusNotStored = errors.New("the seeded status entry is missing from the update response")

func init() {
	ConformanceTests = append(ConformanceTests, HTTPRoutePreserveForeignStatus)
}

var HTTPRoutePreserveForeignStatus = confsuite.ConformanceTest{
	ShortName: "HTTPRoutePreserveForeignStatus",
	Description: "An implementation must not remove or modify HTTPRoute status.parents entries " +
		"whose controllerName belongs to another implementation.",
	Features: []features.FeatureName{
		features.SupportGateway,
		features.SupportHTTPRoute,
	},
	Manifests:   []string{"tests/httproute-preserve-foreign-status.yaml"},
	Provisional: true,
	Test: func(t *testing.T, suite *confsuite.ConformanceTestSuite) {
		ns := confsuite.InfrastructureNamespace
		routeNN := types.NamespacedName{Name: "preserve-foreign-status", Namespace: ns}
		gwNN := types.NamespacedName{Name: "same-namespace", Namespace: ns}

		kubernetes.GatewayAndHTTPRoutesMustBeAccepted(t, suite.Client, suite.TimeoutConfig, suite.ControllerName, kubernetes.NewGatewayRef(gwNN), routeNN)

		var seeded gatewayv1.RouteParentStatus
		err := retry.RetryOnConflict(retry.DefaultRetry, func() error {
			route := &gatewayv1.HTTPRoute{}
			if err := suite.Client.Get(t.Context(), routeNN, route); err != nil {
				return err
			}
			route.Status.Parents = append(route.Status.Parents, gatewayv1.RouteParentStatus{
				ParentRef: gatewayv1.ParentReference{
					Group:     new(gatewayv1.Group(gatewayv1.GroupVersion.Group)),
					Kind:      new(gatewayv1.Kind("Gateway")),
					Name:      "unmanaged-gateway",
					Namespace: new(gatewayv1.Namespace(ns)),
				},
				ControllerName: kubernetes.StaleControllerName,
				Conditions: []metav1.Condition{{
					Type:               string(gatewayv1.RouteConditionAccepted),
					Status:             metav1.ConditionTrue,
					Reason:             string(gatewayv1.RouteReasonAccepted),
					ObservedGeneration: route.Generation,
					LastTransitionTime: metav1.Now(),
				}},
			})
			if err := suite.Client.Status().Update(t.Context(), route); err != nil {
				return err
			}
			// Read the entry back off the update response rather than reusing
			// the value written above: the API server truncates
			// lastTransitionTime to second precision.
			stored := foreignParentStatus(route.Status.Parents)
			if stored == nil {
				return errForeignStatusNotStored
			}
			seeded = *stored
			return nil
		})
		require.NoError(t, err, "error seeding a foreign controller's status entry on HTTPRoute %s", routeNN)

		// Bump the generation so the implementation has to run a full
		// read-modify-write cycle on a status that already holds the seeded entry.
		err = retry.RetryOnConflict(retry.DefaultRetry, func() error {
			route := &gatewayv1.HTTPRoute{}
			if getErr := suite.Client.Get(t.Context(), routeNN, route); getErr != nil {
				return getErr
			}
			route.Spec.Rules[0].Matches[0].Path.Value = new("/preserve-foreign-status-updated")
			return suite.Client.Update(t.Context(), route)
		})
		require.NoError(t, err, "error updating HTTPRoute %s", routeNN)

		// The implementation's own entry catching up to the bumped generation is
		// what proves it wrote status while the seeded entry was there.
		kubernetes.HTTPRouteMustHaveParents(t, suite.Client, suite.TimeoutConfig, routeNN,
			[]gatewayv1.RouteParentStatus{
				{
					ParentRef: gatewayv1.ParentReference{
						Group:     new(gatewayv1.Group(gatewayv1.GroupVersion.Group)),
						Kind:      new(gatewayv1.Kind("Gateway")),
						Name:      gatewayv1.ObjectName(gwNN.Name),
						Namespace: new(gatewayv1.Namespace(ns)),
					},
					ControllerName: gatewayv1.GatewayController(suite.ControllerName),
					Conditions: []metav1.Condition{
						{
							Type:   string(gatewayv1.RouteConditionAccepted),
							Status: metav1.ConditionTrue,
						},
					},
				},
				{
					ParentRef:      seeded.ParentRef,
					ControllerName: kubernetes.StaleControllerName,
					Conditions: []metav1.Condition{
						{
							Type:   string(gatewayv1.RouteConditionAccepted),
							Status: metav1.ConditionTrue,
							Reason: string(gatewayv1.RouteReasonAccepted),
						},
					},
				},
			},
			// The Route and the Gateway share a namespace, so an implementation
			// that leaves parentRef.namespace unset still matches.
			false,
		)

		route := &gatewayv1.HTTPRoute{}
		require.NoError(t, suite.Client.Get(t.Context(), routeNN, route), "error fetching HTTPRoute %s", routeNN)

		// HTTPRouteMustHaveParents compares conditions by type, status and reason,
		// so it cannot see a rewrite that keeps the entry Accepted but restamps
		// observedGeneration or lastTransitionTime.
		stored := foreignParentStatus(route.Status.Parents)
		require.NotNilf(t, stored, "HTTPRoute %s: status entry owned by %s was removed", routeNN, kubernetes.StaleControllerName)
		require.Equalf(t, seeded, *stored, "HTTPRoute %s: status entry owned by %s was modified", routeNN, kubernetes.StaleControllerName)
	},
}

func foreignParentStatus(parents []gatewayv1.RouteParentStatus) *gatewayv1.RouteParentStatus {
	for i := range parents {
		if parents[i].ControllerName == kubernetes.StaleControllerName {
			return &parents[i]
		}
	}

	return nil
}
