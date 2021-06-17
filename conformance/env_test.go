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

package conformance

import (
	"context"
	"fmt"
	"testing"
	"time"

	"sigs.k8s.io/gateway-api/conformance/utils"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
)

func TestEnvironments(t *testing.T) {
	utils.LoadClientsets(t)

	t.Run("The gateway should have an Admitted condition set to True within 3 minutes", func(t *testing.T) {
		err := conditionMustBeTrueWithin(t, "gateway-conformance", "Admitted", "True", 3)
		if err != nil {
			t.Errorf("Timed out waiting for Admitted condition to be set to True: %v", err)
		}
	})
}

func conditionMustBeTrueWithin(t *testing.T, gwcName, condName, condValue string, minutes int) error {
	waitFor := time.Duration(minutes) * time.Minute
	waitErr := wait.PollImmediate(5*time.Second, waitFor, func() (bool, error) {
		gwc, getErr := utils.GWClient.NetworkingV1alpha1().GatewayClasses().Get(context.TODO(), gwcName, metav1.GetOptions{})
		if getErr != nil {
			return false, fmt.Errorf("error fetching GatewayClass: %w", getErr)
		}

		condFound := false
		for _, cond := range gwc.Status.Conditions {
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
