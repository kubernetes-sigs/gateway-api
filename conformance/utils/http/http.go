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

package http

import (
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"sigs.k8s.io/gateway-api/conformance/utils/roundtripper"
)

// ExpectedResponse defines the response expected for a given request.
type ExpectedResponse struct {
	Request    ExpectedRequest
	StatusCode int
	Backend    string
	Namespace  string
}

// ExpectedRequest can be used as both the request to make and a means to verify
// that echoserver received the expected request.
type ExpectedRequest struct {
	Host    string
	Method  string
	Path    string
	Headers map[string]string
}

// MakeRequestAndExpectResponse makes a request with the given parameters and
// verifies the response matches the provided ExpectedResponse.
func MakeRequestAndExpectResponse(t *testing.T, r roundtripper.RoundTripper, gwAddr string, expected ExpectedResponse) {
	t.Helper()

	if expected.Request.Method == "" {
		expected.Request.Method = "GET"
	}

	if expected.StatusCode == 0 {
		expected.StatusCode = 200
	}

	t.Logf("Making %s request to http://%s%s", expected.Request.Method, gwAddr, expected.Request.Path)

	req := roundtripper.Request{
		Method:   expected.Request.Method,
		Host:     expected.Request.Host,
		URL:      url.URL{Scheme: "http", Host: gwAddr, Path: expected.Request.Path},
		Protocol: "HTTP",
	}

	if expected.Request.Headers != nil {
		req.Headers = map[string][]string{}
		for name, value := range expected.Request.Headers {
			req.Headers[name] = []string{value}
		}
	}

	var cReq *roundtripper.CapturedRequest
	var cRes *roundtripper.CapturedResponse

	// We expect eventual consistency once the resources are ready, so wait until
	// we've made a successful request and gotten a response with desired status
	// before making assertions on the response body.
	require.Eventually(t, func() bool {
		var err error

		cReq, cRes, err = r.CaptureRoundTrip(req)
		if err != nil {
			t.Logf("Request failed, not ready yet: %v", err.Error())
			return false
		}

		if cRes.StatusCode != expected.StatusCode {
			t.Logf("Expected response to have status %d but got %d, not ready yet", expected.StatusCode, cRes.StatusCode)
			return false
		}

		return true
	}, 60*time.Second, 1*time.Second, "error making request, never got expected status")

	ExpectResponse(t, cReq, cRes, expected)
}

// ExpectResponse verifies that a captured request and response match the
// provided ExpectedResponse.
func ExpectResponse(t *testing.T, cReq *roundtripper.CapturedRequest, cRes *roundtripper.CapturedResponse, expected ExpectedResponse) {
	t.Helper()
	assert.Equal(t, expected.StatusCode, cRes.StatusCode, "expected status code to be %d, got %d", expected.StatusCode, cRes.StatusCode)
	if cRes.StatusCode == 200 {
		assert.Equal(t, expected.Request.Path, cReq.Path, "expected path to be %s, got %s", expected.Request.Path, cReq.Path)
		assert.Equal(t, expected.Request.Method, cReq.Method, "expected method to be %s, got %s", expected.Request.Method, cReq.Method)
		assert.Equal(t, expected.Namespace, cReq.Namespace, "expected namespace to be %s, got %s", expected.Namespace, cReq.Namespace)
		if expected.Request.Headers != nil {
			if cReq.Headers == nil {
				t.Error("No headers captured")
			} else {
				for name, val := range cReq.Headers {
					cReq.Headers[strings.ToLower(name)] = val
				}
				for name, expectedVal := range expected.Request.Headers {
					actualVal, ok := cReq.Headers[strings.ToLower(name)]
					if !ok {
						t.Errorf("Expected %s header to be set, actual headers: %v", name, cReq.Headers)
					} else if actualVal[0] != expectedVal {
						t.Errorf("Expected %s header to be set to %s, got %s", name, expectedVal, actualVal[0])
					}
				}
			}
		}
		if !strings.HasPrefix(cReq.Pod, expected.Backend) {
			t.Errorf("Expected pod name to start with %s, got %s", expected.Backend, cReq.Pod)
		}
	}
}
