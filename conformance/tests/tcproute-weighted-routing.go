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
	"fmt"
	"math"
	"strings"
	"sync"
	"testing"
	"time"

	"golang.org/x/sync/errgroup"
	"k8s.io/apimachinery/pkg/types"

	"sigs.k8s.io/gateway-api/conformance/utils/kubernetes"
	confsuite "sigs.k8s.io/gateway-api/conformance/utils/suite"
	"sigs.k8s.io/gateway-api/conformance/utils/tlog"
	"sigs.k8s.io/gateway-api/conformance/utils/weight"
	"sigs.k8s.io/gateway-api/pkg/features"
)

func init() {
	ConformanceTests = append(ConformanceTests, TCPRouteWeightedRouting)
}

var TCPRouteWeightedRouting = confsuite.ConformanceTest{
	ShortName:   "TCPRouteWeightedRouting",
	Description: "A TCPRoute with multiple weighted backends should distribute TCP traffic across the backends in proportion to the configured weights.",
	Manifests:   []string{"tests/tcproute-weighted-routing.yaml"},
	Features: []features.FeatureName{
		features.SupportGateway,
		features.SupportTCPRoute,
	},
	Provisional: true,
	Test: func(t *testing.T, suite *confsuite.ConformanceTestSuite) {
		ns := confsuite.InfrastructureNamespace
		gwNN := types.NamespacedName{Name: "tcp-weighted-gateway", Namespace: ns}
		routeNN := types.NamespacedName{Name: "tcp-weighted-route", Namespace: ns}

		// The test creates an additional Gateway in the gateway-conformance-infra
		// namespace so we have to wait for it to be ready.
		kubernetes.NamespacesMustBeReady(t, suite.Client, suite.TimeoutConfig, []string{ns})

		gwAddr := kubernetes.GatewayAndTCPRoutesMustBeAccepted(t, suite.Client, suite.TimeoutConfig, suite.ControllerName,
			kubernetes.NewGatewayRef(gwNN, "tcp"), routeNN)

		t.Run("TCP traffic should be distributed across the weighted backends", func(t *testing.T) {
			expectedWeights := map[string]float64{
				"tcp-backend-v1": 0.7,
				"tcp-backend-v2": 0.3,
			}

			for i := range weight.MaxTestRetries {
				err := assertTCPWeightedDistribution(t.Context(), gwAddr, expectedWeights, 0.03)
				if err == nil {
					return
				}
				tlog.Logf(t, "TCP weighted distribution attempt %d/%d failed: %s", i+1, weight.MaxTestRetries, err)
			}
			t.Fatal("TCP weighted distribution did not converge within tolerance")
		})
	},
}

// assertTCPWeightedDistribution opens a fixed number of TCP connections to
// gwAddr in parallel, classifies each response by the backend Deployment that
// produced it (extracted from the tcpserver TCPAssertions JSON pod name), and
// returns nil if the observed distribution is within tolerance of
// expectedWeights for every backend.
func assertTCPWeightedDistribution(ctx context.Context, gwAddr string, expectedWeights map[string]float64, tolerance float64) error {
	const (
		concurrentRequests = 10
		totalRequests      = 500
		probeTimeout       = 5 * time.Second
	)

	var (
		mu   sync.Mutex
		seen = make(map[string]float64, len(expectedWeights))
		g    errgroup.Group
	)
	g.SetLimit(concurrentRequests)

	for range totalRequests {
		g.Go(func() error {
			pod, err := tcpEchoSendOnce(ctx, gwAddr, probeTimeout)
			if err != nil {
				return err
			}
			backend := extractTCPBackendName(pod)

			mu.Lock()
			defer mu.Unlock()
			if _, ok := expectedWeights[backend]; !ok {
				return fmt.Errorf("response from unexpected backend %q (pod %q)", backend, pod)
			}
			seen[backend]++
			return nil
		})
	}
	if err := g.Wait(); err != nil {
		return fmt.Errorf("error while sending TCP probes: %w", err)
	}

	var errs []error
	for backend, want := range expectedWeights {
		got := seen[backend] / float64(totalRequests)
		if math.Abs(got-want) > tolerance {
			errs = append(errs, fmt.Errorf("backend %q got %.2f%% of traffic; expected %.2f%% (+/- %.2f%%)",
				backend, got*100, want*100, tolerance*100))
		}
	}
	return errors.Join(errs...)
}

// extractTCPBackendName trims the {deployment-hash}-{pod-hash} suffix from a
// pod name to recover the Deployment name. Pod names follow the pattern
// {deployment}-{rs-hash}-{pod-hash}; if the input doesn't match the pattern
// the original name is returned.
func extractTCPBackendName(podName string) string {
	parts := strings.Split(podName, "-")
	if len(parts) < 3 {
		return podName
	}
	return strings.Join(parts[:len(parts)-2], "-")
}
