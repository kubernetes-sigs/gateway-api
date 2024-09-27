/*
Copyright 2024 The Kubernetes Authors.

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
	apiextv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apimachinery/pkg/types"

	"sigs.k8s.io/gateway-api/conformance/utils/kubernetes"
	"sigs.k8s.io/gateway-api/conformance/utils/suite"
	"sigs.k8s.io/gateway-api/pkg/consts"
	"sigs.k8s.io/gateway-api/pkg/features"
)

func init() {
	ConformanceTests = append(ConformanceTests, GatewayClassSupportedVersionCondition)
}

var GatewayClassSupportedVersionCondition = suite.ConformanceTest{
	ShortName: "GatewayClassSupportedVersionCondition",
	Features: []features.FeatureName{
		features.SupportGateway,
	},
	Description: "A GatewayClass should set the SupportedVersion condition based on the presence and version of Gateway API CRDs in the cluster",
	Manifests:   []string{"tests/gatewayclass-supported-version-condition.yaml"},
	Test: func(t *testing.T, s *suite.ConformanceTestSuite) {
		gwc := types.NamespacedName{Name: "gatewayclass-supported-version-condition"}

		t.Run("SupportedVersion condition should be set correctly", func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), s.TimeoutConfig.DefaultTestTimeout)
			defer cancel()

			// Ensure GatewayClass conditions are as expected before proceeding
			kubernetes.GWCMustHaveAcceptedConditionTrue(t, s.Client, s.TimeoutConfig, gwc.Name)
			kubernetes.GWCMustHaveSupportedVersionConditionAny(t, s.Client, s.TimeoutConfig, gwc.Name)

			// Retrieve the GatewayClass CRD
			crd := &apiextv1.CustomResourceDefinition{}
			crdName := types.NamespacedName{Name: "gatewayclasses.gateway.networking.k8s.io"}
			err := s.Client.Get(ctx, crdName, crd)
			require.NoErrorf(t, err, "error getting GatewayClass CRD: %v", err)

			if crd.Annotations != nil {
				// Remove the bundle version annotation if it exists
				if _, exists := crd.Annotations[consts.BundleVersionAnnotation]; !exists {
					t.Fatalf("Annotation %q does not exist on CRD %s", consts.BundleVersionAnnotation, crdName)
				}
				delete(crd.Annotations, consts.BundleVersionAnnotation)
				if err := s.Client.Update(ctx, crd); err != nil {
					t.Fatalf("Failed to update CRD %s: %v", crdName, err)
				}
			}

			// Ensure the SupportedVersion status condition is false
			kubernetes.GWCMustHaveSupportedVersionConditionFalse(t, s.Client, s.TimeoutConfig, gwc.Name)

			// Add the bundle version annotation
			crd.Annotations[consts.BundleVersionAnnotation] = consts.BundleVersion
			if err := s.Client.Update(ctx, crd); err != nil {
				t.Fatalf("Failed to update CRD %s: %v", crdName, err)
			}

			// Ensure the SupportedVersion status condition is true
			kubernetes.GWCMustHaveSupportedVersionConditionTrue(t, s.Client, s.TimeoutConfig, gwc.Name)

			// Set the bundle version annotation to an unsupported version
			crd.Annotations[consts.BundleVersionAnnotation] = "v0.0.0"
			if err := s.Client.Update(ctx, crd); err != nil {
				t.Fatalf("Failed to update CRD %s: %v", crdName, err)
			}

			// Ensure the SupportedVersion is false
			kubernetes.GWCMustHaveSupportedVersionConditionFalse(t, s.Client, s.TimeoutConfig, gwc.Name)

			// Add the bundle version annotation back
			crd.Annotations[consts.BundleVersionAnnotation] = consts.BundleVersion
			if err := s.Client.Update(ctx, crd); err != nil {
				t.Fatalf("Failed to update CRD %s: %v", crdName, err)
			}

			// Ensure the SupportedVersion is true
			kubernetes.GWCMustHaveSupportedVersionConditionTrue(t, s.Client, s.TimeoutConfig, gwc.Name)
		})
	},
}
