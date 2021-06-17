/*
Copyright 2021 The Kubernetes Authors.

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

package httpgateway

import (
	"context"
	"fmt"
	"testing"
	"time"

	tstate "sigs.k8s.io/gateway-api/conformance/state"
	"sigs.k8s.io/gateway-api/conformance/utils"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
)

var (
	state *tstate.Scenario
)

func TestHTTPGateway(t *testing.T) {
	state = tstate.New("gateway-conformance")
	t.Cleanup(func() {
		utils.CleanupScenario(t, state.Namespace, "httpgateway")
	})

	utils.SetupScenario(t, state.Namespace, "httpgateway")

	t.Run("The gateway should have Scheduled condition set to True within 3 minutes", func(t *testing.T) {
		err := conditionMustBeTrueWithin(t, "gateway-conformance", "Scheduled", "True", 3)
		if err != nil {
			t.Errorf("Timed out waiting for Scheduled condition to be set to True: %v", err)
		}
	})

	t.Run("The gateway should have Ready condition set to True within 3 minutes", func(t *testing.T) {
		err := conditionMustBeTrueWithin(t, "gateway-conformance", "Ready", "True", 3)
		if err != nil {
			t.Errorf("Timed out waiting for Ready condition to be set to True: %v", err)
		}
	})

	t.Run("The gateway should have at least one address within 3 minutes", func(t *testing.T) {
		err := mustHaveAddressWithin(t, "gateway-conformance", 3)
		if err != nil {
			t.Errorf("Timed out waiting for Gateway to have at least one address: %v", err)
		}
	})
}

func conditionMustBeTrueWithin(t *testing.T, gwName, condName, condValue string, minutes int) error {
	waitFor := time.Duration(minutes) * time.Minute
	waitErr := wait.PollImmediate(5*time.Second, waitFor, func() (bool, error) {
		gw, getErr := utils.GWClient.NetworkingV1alpha1().Gateways(state.Namespace).Get(context.TODO(), gwName, metav1.GetOptions{})
		if getErr != nil {
			return false, fmt.Errorf("error fetching Gateway: %w", getErr)
		}

		condFound := false
		for _, cond := range gw.Status.Conditions {
			if cond.Type == condName {
				condFound = true
				if cond.Status == metav1.ConditionStatus(condValue) {
					return true, nil
				}
				t.Logf("%s condition set to %s, expected %s", condName, cond.Status, condValue)
			}
		}

		if !condFound {
			t.Logf("%s was not in conditions list", condName)
		}

		return false, nil
	})
	return waitErr
}

func mustHaveAddressWithin(t *testing.T, gwName string, minutes int) error {
	waitFor := time.Duration(minutes) * time.Minute
	waitErr := wait.PollImmediate(5*time.Second, waitFor, func() (bool, error) {
		gw, getErr := utils.GWClient.NetworkingV1alpha1().Gateways(state.Namespace).Get(context.TODO(), gwName, metav1.GetOptions{})
		if getErr != nil {
			return false, fmt.Errorf("error fetching Gateway: %w", getErr)
		}

		if len(gw.Status.Addresses) > 0 {
			return true, nil
		}

		t.Logf("Gateway.Status.Addresses empty")

		return false, nil
	})
	return waitErr
}
