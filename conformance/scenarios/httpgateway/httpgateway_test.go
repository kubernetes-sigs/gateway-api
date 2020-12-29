/*
Copyright 2020 The Kubernetes Authors.

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

	"sigs.k8s.io/gateway-api/apis/v1alpha1"
	"sigs.k8s.io/gateway-api/conformance/utils"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
)

func TestHttpGateway(t *testing.T) {
	// Test Setup
	c, sc, err := utils.LoadClientset()
	if err != nil {
		t.Fatalf("error loading clientset: %v", err)
	}
	err = utils.DynamicApply(utils.DynamicParams{Path: "../../../config/crd/bases"})
	if err != nil {
		t.Fatalf("error installing gateway-api CRDs: %v", err)
	}
	ns, err := utils.NewNamespace(c)
	if err != nil {
		t.Fatalf("error creating namespace: %v", err)
	}

	dp := utils.DynamicParams{
		Path:      "httpgateway.yaml",
		Namespace: ns,
	}
	err = utils.DynamicApply(dp)
	if err != nil {
		t.Fatalf("error installing gateway-api examples: %v", err)
	}

	t.Run("Gateway.Status.Conditions Scheduled condition should be set to true within 3 minutes", func(t *testing.T) {
		t.Parallel()
		waitErr := wait.PollImmediate(5*time.Second, 3*time.Minute, func() (bool, error) {
			gwc, getErr := sc.NetworkingV1alpha1().Gateways(ns).Get(context.TODO(), "my-gateway", metav1.GetOptions{})
			if err != nil {
				return false, fmt.Errorf("error fetching Gateway: %w", getErr)
			}

			for _, cond := range gwc.Status.Conditions {
				if cond.Type == string(v1alpha1.GatewayConditionScheduled) && cond.Status == "True" {
					return true, nil
				}
			}

			return false, nil
		})
		if err != nil {
			t.Errorf("Gateway.Status.Conditions Scheduled condition was not set to true: %v", waitErr)
		}
	})

	t.Run("Gateway.Status.Conditions Ready condition should be set to true within 3 minutes", func(t *testing.T) {
		t.Parallel()
		waitErr := wait.PollImmediate(5*time.Second, 3*time.Minute, func() (bool, error) {
			gwc, getErr := sc.NetworkingV1alpha1().Gateways(ns).Get(context.TODO(), "my-gateway", metav1.GetOptions{})
			if err != nil {
				return false, fmt.Errorf("error fetching Gateway: %w", getErr)
			}

			for _, cond := range gwc.Status.Conditions {
				if cond.Type == string(v1alpha1.GatewayConditionReady) && cond.Status == "True" {
					return true, nil
				}
			}

			return false, nil
		})
		if err != nil {
			t.Errorf("Gateway.Status.Conditions Ready condition was not set to true: %v", waitErr)
		}
	})

	t.Run("Gateway.Status.Listeners[0].Conditions Ready condition should be set to true within 3 minutes", func(t *testing.T) {
		t.Parallel()
		waitErr := wait.PollImmediate(5*time.Second, 3*time.Minute, func() (bool, error) {
			gwc, getErr := sc.NetworkingV1alpha1().Gateways(ns).Get(context.TODO(), "my-gateway", metav1.GetOptions{})
			if err != nil {
				return false, fmt.Errorf("error fetching Gateway: %w", getErr)
			}

			if len(gwc.Status.Listeners) == 0 {
				return false, nil
			}

			if len(gwc.Status.Listeners) > 1 {
				return false, fmt.Errorf("Expected 1 listener entry in status, got %d", len(gwc.Status.Listeners))
			}

			for _, cond := range gwc.Status.Listeners[0].Conditions {
				if cond.Type == string(v1alpha1.ListenerConditionReady) && cond.Status == "True" {
					return true, nil
				}
			}

			return false, nil
		})
		if waitErr != nil {
			t.Errorf("Gateway.Status.Listeners[0].Conditions Ready was not set to true: %v", waitErr)
		}
	})

	// Cleanup
	t.Cleanup(func() {
		dp.Delete = true
		err = utils.DynamicApply(dp)
		if err != nil {
			t.Fatalf("error deleting gateway-api examples: %v", err)
		}
		err = utils.CleanupNamespaces(c)
		if err != nil {
			t.Fatalf("error cleaning up namespaces: %v", err)
		}
	})
}
