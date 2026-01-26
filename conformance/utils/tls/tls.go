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
	"testing"
	"time"

	"sigs.k8s.io/gateway-api/conformance/utils/config"
	"sigs.k8s.io/gateway-api/conformance/utils/http"
	"sigs.k8s.io/gateway-api/conformance/utils/roundtripper"
	"sigs.k8s.io/gateway-api/conformance/utils/tlog"
)

// MakeTLSRequestAndExpectEventuallyConsistentResponse makes a request with the given parameters,
// understanding that the request may fail for some amount of time.
//
// Once the request succeeds consistently with the response having the expected status code, make
// additional assertions on the response body using the provided ExpectedResponse.
func MakeTLSRequestAndExpectEventuallyConsistentResponse(t *testing.T, r roundtripper.RoundTripper, timeoutConfig config.TimeoutConfig, gwAddr string, serverCertificate, clientCertificate, clientCertificateKey []byte, serverName string, expected http.ExpectedResponse) {
	t.Helper()

	req := http.MakeRequest(t, &expected, gwAddr, roundtripper.HTTPSProtocol, "https")
	req.ServerName = serverName
	req.ServerCertificate = serverCertificate
	req.ClientCertificate = clientCertificate
	req.ClientCertificateKey = clientCertificateKey

	WaitForConsistentTLSResponse(t, r, req, expected, timeoutConfig.RequiredConsecutiveSuccesses, timeoutConfig.MaxTimeToConsistency)
}

// WaitForConsistentTLSResponse - repeats the provided request until it completes with a response having
// the expected response consistently. The provided threshold determines how many times in
// a row this must occur to be considered "consistent".
func WaitForConsistentTLSResponse(t *testing.T, r roundtripper.RoundTripper, req roundtripper.Request, expected http.ExpectedResponse, threshold int, maxTimeToConsistency time.Duration) {
	http.AwaitConvergence(t, threshold, maxTimeToConsistency, func(elapsed time.Duration) bool {
		cReq, cRes, err := r.CaptureRoundTrip(req)
		if err != nil {
			tlog.Logf(t, "Request failed, not ready yet: %v (after %v)", err.Error(), elapsed)
			return false
		}

		if err := http.CompareRoundTrip(t, &req, cReq, cRes, expected); err != nil {
			tlog.Logf(t, "Response expectation failed for request: %+v  not ready yet: %v (after %v)", req, err, elapsed)
			tlog.Logf(t, "Full response: %+v", cReq)
			return false
		}

		return true
	})
	tlog.Logf(t, "Request passed")
}

// MakeTLSRequestAndExpectFailureResponse makes one shot request. This function fails
// when HTTP Status OK (200) is returned.
func MakeTLSRequestAndExpectFailureResponse(t *testing.T, r roundtripper.RoundTripper, gwAddr string, serverCertificate, clientCertificate, clientCertificateKey []byte, serverName string, expected http.ExpectedResponse) {
	t.Helper()

	req := http.MakeRequest(t, &expected, gwAddr, roundtripper.HTTPSProtocol, "https")
	req.ServerName = serverName
	req.ServerCertificate = serverCertificate
	req.ClientCertificate = clientCertificate
	req.ClientCertificateKey = clientCertificateKey

	_, _, err := r.CaptureRoundTrip(req)
	if err == nil {
		t.Fatalf("Request should fail")
	}
}
