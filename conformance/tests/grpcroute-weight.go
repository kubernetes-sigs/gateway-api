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
	"cmp"
	"errors"
	"fmt"
	"math"
	"slices"
	"strings"
	"sync"
	"testing"

	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc/codes"
	"k8s.io/apimachinery/pkg/types"
	v1 "sigs.k8s.io/gateway-api/apis/v1"
	pb "sigs.k8s.io/gateway-api/conformance/echo-basic/grpcechoserver"

	"sigs.k8s.io/gateway-api/conformance/utils/grpc"
	"sigs.k8s.io/gateway-api/conformance/utils/kubernetes"
	"sigs.k8s.io/gateway-api/conformance/utils/suite"
	"sigs.k8s.io/gateway-api/pkg/features"
)

func init() {
	ConformanceTests = append(ConformanceTests, GRPCRouteWeight)
}

var GRPCRouteWeight = suite.ConformanceTest{
	ShortName:   "GRPCRouteWeight",
	Description: "An GRPCRoute with weighted backends",
	Manifests:   []string{"tests/grpcroute-weight.yaml"},
	Features: []features.FeatureName{
		features.SupportGateway,
		features.SupportGRPCRoute,
	},
	Test: func(t *testing.T, suite *suite.ConformanceTestSuite) {
		var (
			ns      = "gateway-conformance-infra"
			routeNN = types.NamespacedName{Name: "weighted-backends", Namespace: ns}
			gwNN    = types.NamespacedName{Name: "same-namespace", Namespace: ns}
			gwAddr  = kubernetes.GatewayAndRoutesMustBeAccepted(t, suite.Client, suite.TimeoutConfig, suite.ControllerName, kubernetes.NewGatewayRef(gwNN), &v1.GRPCRoute{}, true, routeNN)
		)

		t.Run("Requests should have a distribution that matches the weight", func(t *testing.T) {
			expected := grpc.ExpectedResponse{
				EchoRequest: &pb.EchoRequest{},
				Response:    grpc.Response{Code: codes.OK},
				Namespace:   "gateway-conformance-infra",
			}

			// Assert request succeeds before doing 	our distribution check
			grpc.MakeRequestAndExpectEventuallyConsistentResponse(t, suite.GRPCClient, suite.TimeoutConfig, gwAddr, expected)

			for i := 0; i < 10; i++ {
				if err := testGRPCDistribution(t, suite, gwAddr, expected); err != nil {
					t.Logf("Traffic distribution test failed (%d/10): %s", i+1, err)
				} else {
					return
				}
			}
			t.Fatal("Weighted distribution tests failed")
		})
	},
}

func testGRPCDistribution(t *testing.T, suite *suite.ConformanceTestSuite, gwAddr string, expected grpc.ExpectedResponse) error {
	const (
		concurrentRequests  = 10
		tolerancePercentage = 0.05
		totalRequests       = 500.0
	)
	var (
		g               errgroup.Group
		seenMutex       sync.Mutex
		seen            = make(map[string]float64, 3 /* number of backends */)
		expectedWeights = map[string]float64{
			"grpc-infra-backend-v1": 0.7,
			"grpc-infra-backend-v2": 0.3,
			"grpc-infra-backend-v3": 0.0,
		}
		grpcClient = &grpc.DefaultClient{}
	)
	g.SetLimit(concurrentRequests)
	for i := 0.0; i < totalRequests; i++ {
		g.Go(func() error {
			resp, err := grpcClient.SendRPC(t, gwAddr, expected, suite.TimeoutConfig.MaxTimeToConsistency)
			if err != nil {
				return fmt.Errorf("failed to send gRPC request: %w", err)
			}
			if resp.Code != codes.OK {
				return fmt.Errorf("expected OK response, got %v", resp.Code)
			}

			seenMutex.Lock()
			defer seenMutex.Unlock()

			podName := resp.Response.GetAssertions().GetContext().GetPod()
			for expectedBackend := range expectedWeights {
				if strings.HasPrefix(podName, expectedBackend) {
					seen[expectedBackend]++
					return nil
				}
			}

			return fmt.Errorf("request was handled by an unexpected pod %q", podName)
		})
	}

	if err := g.Wait(); err != nil {
		return fmt.Errorf("error while sending requests: %w", err)
	}

	var errs []error
	if len(seen) != 2 {
		errs = append(errs, fmt.Errorf("expected only two backends to receive traffic"))
	}

	for wantBackend, wantPercent := range expectedWeights {
		gotCount, ok := seen[wantBackend]

		if !ok && wantPercent != 0.0 {
			errs = append(errs, fmt.Errorf("expect traffic to hit backend %q - but none was received", wantBackend))
			continue
		}

		gotPercent := gotCount / totalRequests

		if math.Abs(gotPercent-wantPercent) > tolerancePercentage {
			errs = append(errs, fmt.Errorf("backend %q weighted traffic of %v not within tolerance %v (+/-%f)",
				wantBackend,
				gotPercent,
				wantPercent,
				tolerancePercentage,
			))
		}
	}
	slices.SortFunc(errs, func(a, b error) int {
		return cmp.Compare(a.Error(), b.Error())
	})
	return errors.Join(errs...)
}
