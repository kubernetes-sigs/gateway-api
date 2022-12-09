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
	"reflect"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	"sigs.k8s.io/gateway-api/apis/v1beta1"
	"sigs.k8s.io/gateway-api/conformance/utils/kubernetes"
	"sigs.k8s.io/gateway-api/conformance/utils/suite"
)

func init() {
	ConformanceTests = append(ConformanceTests, HTTPRouteObservedGenerationBump)
}

var HTTPRouteObservedGenerationBump = suite.ConformanceTest{
	ShortName:   "HTTPRouteObservedGenerationBump",
	Description: "A HTTPRoute in the gateway-conformance-infra namespace should update the observedGeneration in all of it's Status.Conditions after an update to the spec",
	Manifests:   []string{"tests/httproute-observed-generation-bump.yaml"},
	Test: func(t *testing.T, s *suite.ConformanceTestSuite) {

		routeNN := types.NamespacedName{Name: "observed-generation-bump", Namespace: "gateway-conformance-infra"}
		gwNN := types.NamespacedName{Name: "same-namespace", Namespace: "gateway-conformance-infra"}

		acceptedCondition := metav1.Condition{
			Type:   string(v1beta1.RouteConditionAccepted),
			Status: metav1.ConditionTrue,
			Reason: "", // any reason
		}

		t.Run("observedGeneration should increment", func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
			defer cancel()

			namespaces := []string{"gateway-conformance-infra"}
			kubernetes.NamespacesMustBeAccepted(t, s.Client, s.TimeoutConfig, namespaces)

			existing := &v1beta1.HTTPRoute{}
			err := s.Client.Get(ctx, routeNN, existing)
			require.NoErrorf(t, err, "error getting HTTPRoute: %v", err)

			// Sanity check
			if kubernetes.HTTPRouteConditionsHaveLatestObservedGeneration(existing) {
				t.Fatal("Not all the condition's observedGeneration were updated")
			}

			existing.Spec.Rules[0].BackendRefs[0].Name = "infra-backend-new"
			err = s.Client.Update(ctx, existing)
			require.NoErrorf(t, err, "error updating the HTTPRoute: %v", err)

			kubernetes.HTTPRouteMustHaveCondition(t, s.Client, s.TimeoutConfig, routeNN, gwNN, acceptedCondition)

			updated := &v1beta1.HTTPRoute{}
			err = s.Client.Get(ctx, routeNN, updated)
			require.NoErrorf(t, err, "error getting Gateway: %v", err)

			// Sanity check
			if kubernetes.HTTPRouteConditionsHaveLatestObservedGeneration(updated) {
				t.Fatal("Not all the condition's observedGeneration were updated")
			}

			if existing.Generation == updated.Generation {
				t.Errorf("Expected generation to change because of spec change - remained at %v", updated.Generation)
			}

			for _, up := range updated.Status.Parents {
				existing := parentStatusForRef(existing.Status.Parents, up.ParentRef)
				if existing == nil {
					t.Logf("Observed unexpected new parent ref %#v", up.ParentRef)
					continue
				}
				for _, uc := range up.Conditions {
					for _, ec := range existing.Conditions {
						if ec.Type == uc.Type && ec.ObservedGeneration == uc.ObservedGeneration {
							t.Errorf("Expected status condition %q observedGeneration to change - remained at %v", uc.Type, uc.ObservedGeneration)
						}
					}
				}
			}
		})
	},
}

func parentStatusForRef(statuses []v1beta1.RouteParentStatus, ref v1beta1.ParentReference) *v1beta1.RouteParentStatus {
	for _, status := range statuses {
		if reflect.DeepEqual(status.ParentRef, ref) {
			return &status
		}
	}
	return nil

}
