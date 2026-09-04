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

func init() {
	ConformanceTests = append(ConformanceTests, BackendTLSPolicyPreserveForeignStatus)
}

var BackendTLSPolicyPreserveForeignStatus = confsuite.ConformanceTest{
	ShortName: "BackendTLSPolicyPreserveForeignStatus",
	Description: "An implementation must not remove or modify BackendTLSPolicy status.ancestors " +
		"entries whose controllerName belongs to another implementation.",
	Features: []features.FeatureName{
		features.SupportGateway,
		features.SupportHTTPRoute,
		features.SupportBackendTLSPolicy,
	},
	Manifests:   []string{"tests/backendtlspolicy-preserve-foreign-status.yaml"},
	Provisional: true,
	Test: func(t *testing.T, suite *confsuite.ConformanceTestSuite) {
		ns := confsuite.InfrastructureNamespace
		routeNN := types.NamespacedName{Name: "backendtlspolicy-preserve-foreign-status", Namespace: ns}
		gwNN := types.NamespacedName{Name: "same-namespace", Namespace: ns}
		policyNN := types.NamespacedName{Name: "backendtlspolicy-preserve-foreign-status", Namespace: ns}

		kubernetes.NamespacesMustBeReady(t, suite.Client, suite.TimeoutConfig, []string{ns})
		kubernetes.GatewayAndHTTPRoutesMustBeAccepted(t, suite.Client, suite.TimeoutConfig, suite.ControllerName, kubernetes.NewGatewayRef(gwNN), routeNN)
		kubernetes.BackendTLSPolicyMustHaveAcceptedConditionTrue(t, suite.Client, suite.TimeoutConfig, policyNN, gwNN)

		var seeded gatewayv1.PolicyAncestorStatus
		err := retry.RetryOnConflict(retry.DefaultRetry, func() error {
			policy := &gatewayv1.BackendTLSPolicy{}
			if err := suite.Client.Get(t.Context(), policyNN, policy); err != nil {
				return err
			}
			policy.Status.Ancestors = append(policy.Status.Ancestors, gatewayv1.PolicyAncestorStatus{
				AncestorRef: gatewayv1.ParentReference{
					Group:     new(gatewayv1.Group(gatewayv1.GroupVersion.Group)),
					Kind:      new(gatewayv1.Kind("Gateway")),
					Name:      "unmanaged-gateway",
					Namespace: new(gatewayv1.Namespace(ns)),
				},
				ControllerName: kubernetes.StaleControllerName,
				Conditions: []metav1.Condition{{
					Type:               string(gatewayv1.PolicyConditionAccepted),
					Status:             metav1.ConditionTrue,
					Reason:             string(gatewayv1.PolicyReasonAccepted),
					ObservedGeneration: policy.Generation,
					LastTransitionTime: metav1.Now(),
				}},
			})
			if err := suite.Client.Status().Update(t.Context(), policy); err != nil {
				return err
			}
			// Read the entry back off the update response rather than reusing
			// the value written above: the API server truncates
			// lastTransitionTime to second precision.
			stored := foreignAncestorStatus(policy.Status.Ancestors)
			if stored == nil {
				return errForeignStatusNotStored
			}
			seeded = *stored
			return nil
		})
		require.NoError(t, err, "error seeding a foreign controller's status entry on BackendTLSPolicy %s", policyNN)

		// Bump the generation so the implementation has to run a full
		// read-modify-write cycle on a status that already holds the seeded entry.
		err = retry.RetryOnConflict(retry.DefaultRetry, func() error {
			policy := &gatewayv1.BackendTLSPolicy{}
			if getErr := suite.Client.Get(t.Context(), policyNN, policy); getErr != nil {
				return getErr
			}
			policy.Spec.Validation.Hostname = "second.example.com"
			return suite.Client.Update(t.Context(), policy)
		})
		require.NoError(t, err, "error updating BackendTLSPolicy %s", policyNN)

		// The implementation's own entry catching up to the bumped generation is
		// what proves it wrote status while the seeded entry was there.
		kubernetes.BackendTLSPolicyMustHaveAcceptedConditionTrue(t, suite.Client, suite.TimeoutConfig, policyNN, gwNN)

		policy := &gatewayv1.BackendTLSPolicy{}
		require.NoError(t, suite.Client.Get(t.Context(), policyNN, policy), "error fetching BackendTLSPolicy %s", policyNN)
		kubernetes.BackendTLSPolicyMustHaveLatestConditions(t, policy)

		stored := foreignAncestorStatus(policy.Status.Ancestors)
		require.NotNilf(t, stored, "BackendTLSPolicy %s: status entry owned by %s was removed", policyNN, kubernetes.StaleControllerName)
		require.Equalf(t, seeded, *stored, "BackendTLSPolicy %s: status entry owned by %s was modified", policyNN, kubernetes.StaleControllerName)
	},
}

func foreignAncestorStatus(ancestors []gatewayv1.PolicyAncestorStatus) *gatewayv1.PolicyAncestorStatus {
	for i := range ancestors {
		if ancestors[i].ControllerName == kubernetes.StaleControllerName {
			return &ancestors[i]
		}
	}

	return nil
}
