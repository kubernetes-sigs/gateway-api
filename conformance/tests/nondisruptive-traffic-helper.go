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
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"sigs.k8s.io/gateway-api/conformance/utils/http"
	"sigs.k8s.io/gateway-api/conformance/utils/roundtripper"
	"sigs.k8s.io/gateway-api/conformance/utils/tlog"
)

// trafficResult holds the results of a continuous traffic test.
type trafficResult struct {
	TotalRequests  int64
	FailedRequests int64
}

// continuousTraffic starts a goroutine that sends HTTP requests at the given interval.
// Returns a stop function; calling it stops the goroutine and returns the results.
// Counts any err != nil or StatusCode != 200 as a failed request.
func continuousTraffic(
	t *testing.T,
	rt roundtripper.RoundTripper,
	gwAddr, host, path string,
) (stop func() trafficResult) {
	t.Helper()

	var totalRequests atomic.Int64
	var failedRequests atomic.Int64
	stopCh := make(chan struct{})
	var once sync.Once
	var res trafficResult

	expected := http.ExpectedResponse{
		Request: http.Request{
			Host: host,
			Path: path,
		},
		Response: http.Response{
			StatusCode: 200,
		},
	}
	req := http.MakeRequest(t, &expected, gwAddr, "HTTP", "http")

	const interval = 100 * time.Millisecond

	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for {
			select {
			case <-stopCh:
				return
			case <-ticker.C:
				totalRequests.Add(1)
				_, cRes, err := rt.CaptureRoundTrip(req)
				if err != nil {
					failedRequests.Add(1)
					tlog.Logf(t, "continuous traffic request failed: %v", err)
					continue
				}
				if cRes.StatusCode != 200 {
					failedRequests.Add(1)
					tlog.Logf(t, "continuous traffic request returned status %d", cRes.StatusCode)
				}
			}
		}
	}()

	stopFn := func() trafficResult {
		once.Do(func() {
			close(stopCh)
			// Wait briefly for any in-flight request to complete.
			time.Sleep(interval * 2)
			res = trafficResult{
				TotalRequests:  totalRequests.Load(),
				FailedRequests: failedRequests.Load(),
			}
		})
		return res
	}

	// Ensure we never leak the goroutine even if the test bails out early.
	t.Cleanup(func() { stopFn() })
	return stopFn
}
