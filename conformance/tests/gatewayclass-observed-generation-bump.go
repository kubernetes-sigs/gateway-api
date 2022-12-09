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
	"time"

	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/types"

	"sigs.k8s.io/gateway-api/apis/v1beta1"
	"sigs.k8s.io/gateway-api/conformance/utils/kubernetes"
	"sigs.k8s.io/gateway-api/conformance/utils/suite"
)

func init() {
	ConformanceTests = append(ConformanceTests, GatewayClassObservedGenerationBump)
}

var GatewayClassObservedGenerationBump = suite.ConformanceTest{
	ShortName:   "GatewayClassObservedGenerationBump",
	Description: "A GatewayClass should update the observedGeneration in all of it's Status.Conditions after an update to the spec",
	Manifests:   []string{"tests/gatewayclass-observed-generation-bump.yaml"},
	Test: func(t *testing.T, s *suite.ConformanceTestSuite) {

		gwc := types.NamespacedName{Name: "gatewayclass-observed-generation-bump"}

		t.Run("observedGeneration should increment", func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
			defer cancel()

			kubernetes.GWCMustBeAccepted(t, s.Client, s.TimeoutConfig, gwc.Name)

			existing := &v1beta1.GatewayClass{}
			err := s.Client.Get(ctx, gwc, existing)
			require.NoErrorf(t, err, "error getting GatewayClass: %v", err)

			// Sanity check
			if kubernetes.GatewayClassConditionsHaveLatestObservedGeneration(existing) {
				t.Fatal("Not all the condition's observedGeneration were updated")
			}

			desc := "new"
			existing.Spec.Description = &desc

			err = s.Client.Update(ctx, existing)
			require.NoErrorf(t, err, "error updating the GatewayClass: %v", err)

			// Ensure the generation and observedGeneration sync up
			kubernetes.GWCMustBeAccepted(t, s.Client, s.TimeoutConfig, gwc.Name)

			updated := &v1beta1.GatewayClass{}
			err = s.Client.Get(ctx, gwc, updated)
			require.NoErrorf(t, err, "error getting GatewayClass: %v", err)

			// Sanity check
			if kubernetes.GatewayClassConditionsHaveLatestObservedGeneration(updated) {
				t.Fatal("Not all the condition's observedGeneration were updated")
			}

			if existing.Generation == updated.Generation {
				t.Errorf("Expected generation to change because of spec change - remained at %v", updated.Generation)
			}

			for _, uc := range updated.Status.Conditions {
				for _, ec := range existing.Status.Conditions {
					if ec.Type == uc.Type && ec.ObservedGeneration == uc.ObservedGeneration {
						t.Errorf("Expected status condition %q observedGeneration to change - remained at %v", uc.Type, uc.ObservedGeneration)
					}
				}
			}
		})
	},
}
