/*
Copyright 2022 The Kubernetes Authors.

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

package tls

import (
	"net/url"
	"strings"
	"testing"
	"time"

	"sigs.k8s.io/gateway-api/conformance/utils/config"
	"sigs.k8s.io/gateway-api/conformance/utils/http"
	"sigs.k8s.io/gateway-api/conformance/utils/roundtripper"
)

// requiredConsecutiveSuccesses is the number of requests that must succeed in a row
// for MakeRequestAndExpectEventuallyConsistentResponse to consider the response "consistent"
// before making additional assertions on the response body. If this number is not reached within
// maxTimeToConsistency, the test will fail.
const requiredConsecutiveSuccesses = 3

// MakeTLSRequestAndExpectEventuallyConsistentResponse makes a request with the given parameters,
// understanding that the request may fail for some amount of time.
//
// Once the request succeeds consistently with the response having the expected status code, make
// additional assertions on the response body using the provided ExpectedResponse.
func MakeTLSRequestAndExpectEventuallyConsistentResponse(t *testing.T, r roundtripper.RoundTripper, timeoutConfig config.TimeoutConfig, gwAddr string, cPem, kPem []byte, server string, expected http.ExpectedResponse) {
	t.Helper()

	protocol := "HTTPS"
	scheme := "https"

	if expected.Request.Method == "" {
		expected.Request.Method = "GET"
	}

	if expected.Response.StatusCode == 0 {
		expected.Response.StatusCode = 200
	}

	t.Logf("Making %s request to %s://%s%s", expected.Request.Method, scheme, gwAddr, expected.Request.Path)

	path, query, _ := strings.Cut(expected.Request.Path, "?")

	req := roundtripper.Request{
		Method:   expected.Request.Method,
		Host:     expected.Request.Host,
		URL:      url.URL{Scheme: scheme, Host: gwAddr, Path: path, RawQuery: query},
		Protocol: protocol,
		Headers:  map[string][]string{},
	}

	if expected.Request.Headers != nil {
		for name, value := range expected.Request.Headers {
			req.Headers[name] = []string{value}
		}
	}

	backendSetHeaders := []string{}
	for name, val := range expected.BackendSetResponseHeaders {
		backendSetHeaders = append(backendSetHeaders, name+":"+val)
	}
	req.Headers["X-Echo-Set-Header"] = []string{strings.Join(backendSetHeaders, ",")}

	WaitForConsistentTLSResponse(t, r, req, expected, requiredConsecutiveSuccesses, timeoutConfig.MaxTimeToConsistency, cPem, kPem, server)
}

// WaitForConsistentTLSResponse - repeats the provided request until it completes with a response having
// the expected response consistently. The provided threshold determines how many times in
// a row this must occur to be considered "consistent".
func WaitForConsistentTLSResponse(t *testing.T, r roundtripper.RoundTripper, req roundtripper.Request, expected http.ExpectedResponse, threshold int, maxTimeToConsistency time.Duration, cPem, kPem []byte, server string) {
	http.AwaitConvergence(t, threshold, maxTimeToConsistency, func(elapsed time.Duration) bool {
		cReq, cRes, err := r.CaptureTLSRoundTrip(req, cPem, kPem, server)
		if err != nil {
			t.Logf("Request failed, not ready yet: %v (after %v)", err.Error(), elapsed)
			return false
		}

		if err := http.CompareRequest(cReq, cRes, expected); err != nil {
			t.Logf("Response expectation failed for request: %v  not ready yet: %v (after %v)", req, err, elapsed)
			return false
		}

		return true
	})
	t.Logf("Request passed")
}
