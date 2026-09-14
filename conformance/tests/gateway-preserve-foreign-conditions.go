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
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/util/retry"

	gatewayv1 "sigs.k8s.io/gateway-api/apis/v1"
	"sigs.k8s.io/gateway-api/conformance/utils/kubernetes"
	confsuite "sigs.k8s.io/gateway-api/conformance/utils/suite"
	"sigs.k8s.io/gateway-api/pkg/features"
)

var errForeignConditionNotStored = errors.New("the seeded condition is missing from the update response")

func init() {
	ConformanceTests = append(ConformanceTests, GatewayPreserveForeignConditions)
}

var GatewayPreserveForeignConditions = confsuite.ConformanceTest{
	ShortName: "GatewayPreserveForeignConditions",
	Description: "An implementation must not remove or modify Gateway and Listener status conditions " +
		"whose type it is not responsible for.",
	Features: []features.FeatureName{
		features.SupportGateway,
	},
	Manifests:   []string{"tests/gateway-preserve-foreign-conditions.yaml"},
	Provisional: true,
	Test: func(t *testing.T, suite *confsuite.ConformanceTestSuite) {
		ns := confsuite.InfrastructureNamespace
		gwNN := types.NamespacedName{Name: "gateway-preserve-foreign-conditions", Namespace: ns}
		listenerName := gatewayv1.SectionName("http")

		kubernetes.NamespacesMustBeReady(t, suite.Client, suite.TimeoutConfig, []string{ns})
		kubernetes.GatewayMustHaveLatestConditions(t, suite.Client, suite.TimeoutConfig, gwNN)

		// Listener status is where the second seeded condition goes, so its
		// entry has to exist first.
		waitErr := wait.PollUntilContextTimeout(context.Background(), suite.TimeoutConfig.DefaultPollInterval, suite.TimeoutConfig.GatewayStatusMustHaveListeners, true, func(ctx context.Context) (bool, error) {
			gw := &gatewayv1.Gateway{}
			if err := suite.Client.Get(ctx, gwNN, gw); err != nil {
				return false, err
			}
			return listenerStatus(gw, listenerName) != nil, nil
		})
		require.NoErrorf(t, waitErr, "error waiting for Gateway %s to report status for listener %s", gwNN, listenerName)

		foreignCondition := func(observedGeneration int64) metav1.Condition {
			return metav1.Condition{
				Type:               kubernetes.StaleConditionType,
				Status:             metav1.ConditionTrue,
				Reason:             "Seeded",
				ObservedGeneration: observedGeneration,
				LastTransitionTime: metav1.Now(),
			}
		}

		var seededGw, seededListener metav1.Condition
		err := retry.RetryOnConflict(retry.DefaultRetry, func() error {
			gw := &gatewayv1.Gateway{}
			if err := suite.Client.Get(t.Context(), gwNN, gw); err != nil {
				return err
			}
			ls := listenerStatus(gw, listenerName)
			if ls == nil {
				return errForeignConditionNotStored
			}
			gw.Status.Conditions = append(gw.Status.Conditions, foreignCondition(gw.Generation))
			ls.Conditions = append(ls.Conditions, foreignCondition(gw.Generation))
			if err := suite.Client.Status().Update(t.Context(), gw); err != nil {
				return err
			}
			// Read the conditions back off the update response rather than
			// reusing the values written above: the API server truncates
			// lastTransitionTime to second precision.
			storedGw := foreignConditionIn(gw.Status.Conditions)
			ls = listenerStatus(gw, listenerName)
			if storedGw == nil || ls == nil {
				return errForeignConditionNotStored
			}
			storedListener := foreignConditionIn(ls.Conditions)
			if storedListener == nil {
				return errForeignConditionNotStored
			}
			seededGw = *storedGw
			seededListener = *storedListener
			return nil
		})
		require.NoError(t, err, "error seeding foreign conditions on Gateway %s", gwNN)

		// Bump the generation so the implementation has to run a full
		// read-modify-write cycle on a status that already holds the seeded
		// conditions.
		err = retry.RetryOnConflict(retry.DefaultRetry, func() error {
			gw := &gatewayv1.Gateway{}
			if getErr := suite.Client.Get(t.Context(), gwNN, gw); getErr != nil {
				return getErr
			}
			gw.Spec.Listeners[0].Hostname = new(gatewayv1.Hostname("second.example.com"))
			return suite.Client.Update(t.Context(), gw)
		})
		require.NoError(t, err, "error updating Gateway %s", gwNN)

		// The implementation's own conditions catching up to the bumped
		// generation, on the Gateway and on every listener, is what proves it
		// wrote status while the seeded conditions were there.
		waitErr = wait.PollUntilContextTimeout(context.Background(), suite.TimeoutConfig.DefaultPollInterval, suite.TimeoutConfig.LatestObservedGenerationSet, true, func(ctx context.Context) (bool, error) {
			gw := &gatewayv1.Gateway{}
			if err := suite.Client.Get(ctx, gwNN, gw); err != nil {
				return false, err
			}
			if err := kubernetes.ConditionsHaveLatestObservedGeneration(gw, gw.Status.Conditions); err != nil {
				return false, nil
			}
			for _, ls := range gw.Status.Listeners {
				if err := kubernetes.ConditionsHaveLatestObservedGeneration(gw, ls.Conditions); err != nil {
					return false, nil
				}
			}
			return true, nil
		})
		require.NoErrorf(t, waitErr, "error waiting for Gateway %s conditions to catch up with the bumped generation", gwNN)

		gw := &gatewayv1.Gateway{}
		require.NoError(t, suite.Client.Get(t.Context(), gwNN, gw), "error fetching Gateway %s", gwNN)

		stored := foreignConditionIn(gw.Status.Conditions)
		require.NotNilf(t, stored, "Gateway %s: the %s condition was removed", gwNN, kubernetes.StaleConditionType)
		require.Equalf(t, seededGw, *stored, "Gateway %s: the %s condition was modified", gwNN, kubernetes.StaleConditionType)

		ls := listenerStatus(gw, listenerName)
		require.NotNilf(t, ls, "Gateway %s: status for listener %s is missing", gwNN, listenerName)
		stored = foreignConditionIn(ls.Conditions)
		require.NotNilf(t, stored, "Gateway %s listener %s: the %s condition was removed", gwNN, listenerName, kubernetes.StaleConditionType)
		require.Equalf(t, seededListener, *stored, "Gateway %s listener %s: the %s condition was modified", gwNN, listenerName, kubernetes.StaleConditionType)
	},
}

func listenerStatus(gw *gatewayv1.Gateway, name gatewayv1.SectionName) *gatewayv1.ListenerStatus {
	for i := range gw.Status.Listeners {
		if gw.Status.Listeners[i].Name == name {
			return &gw.Status.Listeners[i]
		}
	}

	return nil
}

func foreignConditionIn(conditions []metav1.Condition) *metav1.Condition {
	for i := range conditions {
		if conditions[i].Type == kubernetes.StaleConditionType {
			return &conditions[i]
		}
	}

	return nil
}
