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
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

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
