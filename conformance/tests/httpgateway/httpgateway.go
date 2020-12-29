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
	"time"

	"github.com/cucumber/godog"
	"github.com/cucumber/messages-go/v10"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/klog"
	tstate "sigs.k8s.io/gateway-api/conformance/state"
	"sigs.k8s.io/gateway-api/conformance/utils"
)

var (
	state *tstate.Scenario
)

// InitializeScenario configures the Feature to test
func InitializeScenario(ctx *godog.ScenarioContext) {
	ctx.Step(`^a new random namespace$`, aNewRandomNamespace)
	ctx.Step(`^the "([^"]*)" scenario$`, theScenario)
	ctx.Step(`^Gateway "([^"]*)" should have "([^"]*)" condition should be set to "([^"]*)" within (\d+) minutes$`, conditionMustBeTrueWithin)
	ctx.Step(`^Gateway "([^"]*)" should have an address in status within (\d+) minutes`, mustHaveAddressWithin)

	ctx.BeforeScenario(func(*godog.Scenario) {
		state = tstate.New()
	})

	ctx.AfterScenario(func(*messages.Pickle, error) {
		// delete namespace an all the content
		_ = utils.DeleteNamespace(state.Namespace)
	})
}

func aNewRandomNamespace() error {
	ns, err := utils.NewNamespace()
	if err != nil {
		return err
	}

	state.Namespace = ns
	return nil
}

func theScenario(name string) error {
	dp := utils.DynamicParams{
		Path:      fmt.Sprintf("features/%s/%s.yaml", name, name),
		Namespace: state.Namespace,
	}
	return utils.DynamicApply(dp)
}

func conditionMustBeTrueWithin(gwName, condName, condValue string, minutes int) error {
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
				klog.Warningf("%s condition set to %s, expected %s", condName, cond.Status, condValue)
			}
		}

		if !condFound {
			klog.Warningf("%s was not in conditions list", condName)
		}

		return false, nil
	})
	return waitErr
}

func mustHaveAddressWithin(gwName string, minutes int) error {
	waitFor := time.Duration(minutes) * time.Minute
	waitErr := wait.PollImmediate(5*time.Second, waitFor, func() (bool, error) {
		gw, getErr := utils.GWClient.NetworkingV1alpha1().Gateways(state.Namespace).Get(context.TODO(), gwName, metav1.GetOptions{})
		if getErr != nil {
			return false, fmt.Errorf("error fetching Gateway: %w", getErr)
		}

		if len(gw.Status.Addresses) > 0 {
			return true, nil
		}

		klog.Warningf("Gateway.Status.Addresses empty")

		return false, nil
	})
	return waitErr
}
