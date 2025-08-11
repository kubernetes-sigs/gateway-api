/*
Copyright 2025 The Kubernetes Authors.

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
	"time"

	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc/codes"

	"sigs.k8s.io/gateway-api/conformance/utils/grpc"
	"sigs.k8s.io/gateway-api/conformance/utils/http"
	"sigs.k8s.io/gateway-api/conformance/utils/roundtripper"
	"sigs.k8s.io/gateway-api/conformance/utils/suite"
)

// RequestSender defines an interface for sending requests (HTTP or gRPC)
type RequestSender interface {
	SendRequest() (podName string, err error)
}

// HTTPRequestSender implements RequestSender for HTTP requests
type HTTPRequestSender struct {
	t            *testing.T
	roundTripper roundtripper.RoundTripper
	gwAddr       string
	expected     http.ExpectedResponse
}

func (s *HTTPRequestSender) SendRequest() (string, error) {
	uniqueExpected := s.expected
	if err := http.AddEntropy(&uniqueExpected); err != nil {
		return "", fmt.Errorf("error adding entropy: %w", err)
	}
	req := http.MakeRequest(s.t, &uniqueExpected, s.gwAddr, "HTTP", "http")
	cReq, cRes, err := s.roundTripper.CaptureRoundTrip(req)
	if err != nil {
		return "", fmt.Errorf("failed to roundtrip request: %w", err)
	}
	if err := http.CompareRoundTrip(s.t, &req, cReq, cRes, s.expected); err != nil {
		return "", fmt.Errorf("response expectation failed for request: %w", err)
	}
	return cReq.Pod, nil
}

// GRPCRequestSender implements RequestSender for gRPC requests
type GRPCRequestSender struct {
	t        *testing.T
	client   grpc.Client
	gwAddr   string
	expected grpc.ExpectedResponse
	timeout  time.Duration
}

func (s *GRPCRequestSender) SendRequest() (string, error) {
	uniqueExpected := s.expected
	if err := grpc.AddEntropy(&uniqueExpected); err != nil {
		return "", fmt.Errorf("error adding entropy: %w", err)
	}
	resp, err := s.client.SendRPC(s.t, s.gwAddr, uniqueExpected, s.timeout)
	if err != nil {
		return "", fmt.Errorf("failed to send gRPC request: %w", err)
	}
	if resp.Code != codes.OK {
		return "", fmt.Errorf("expected OK response, got %v", resp.Code)
	}
	return resp.Response.GetAssertions().GetContext().GetPod(), nil
}

// testWeightedDistribution tests that requests are distributed according to expected weights
func testWeightedDistribution(sender RequestSender, expectedWeights map[string]float64) error {
	const (
		concurrentRequests  = 10
		tolerancePercentage = 0.05
		totalRequests       = 500.0
	)

	var (
		g         errgroup.Group
		seenMutex sync.Mutex
		seen      = make(map[string]float64, len(expectedWeights))
	)

	g.SetLimit(concurrentRequests)
	for i := 0.0; i < totalRequests; i++ {
		g.Go(func() error {
			podName, err := sender.SendRequest()
			if err != nil {
				return err
			}

			seenMutex.Lock()
			defer seenMutex.Unlock()

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

	// Count how many backends should receive traffic (weight > 0)
	expectedActiveBackends := 0
	for _, weight := range expectedWeights {
		if weight > 0.0 {
			expectedActiveBackends++
		}
	}

	var errs []error
	if len(seen) != expectedActiveBackends {
		errs = append(errs, fmt.Errorf("expected %d backends to receive traffic, but got %d", expectedActiveBackends, len(seen)))
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

// newHTTPRequestSender creates a new HTTPRequestSender
func newHTTPRequestSender(t *testing.T, suite *suite.ConformanceTestSuite, gwAddr string, expected http.ExpectedResponse) *HTTPRequestSender {
	return &HTTPRequestSender{
		t:            t,
		roundTripper: suite.RoundTripper,
		gwAddr:       gwAddr,
		expected:     expected,
	}
}

// newGRPCRequestSender creates a new GRPCRequestSender
func newGRPCRequestSender(t *testing.T, suite *suite.ConformanceTestSuite, gwAddr string, expected grpc.ExpectedResponse) *GRPCRequestSender {
	return &GRPCRequestSender{
		t:        t,
		client:   &grpc.DefaultClient{},
		gwAddr:   gwAddr,
		expected: expected,
		timeout:  suite.TimeoutConfig.MaxTimeToConsistency,
	}
}
