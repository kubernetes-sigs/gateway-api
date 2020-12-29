/*
Copyright 2020 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License"); you may not use
this file except in compliance with the License. You may obtain a copy of the
License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software distributed
under the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR
CONDITIONS OF ANY KIND, either express or implied. See the License for the
specific language governing permissions and limitations under the License.
*/

package state

import (
	"fmt"
	"strings"

	"sigs.k8s.io/gateway-api/conformance/http"
)

// Scenario holds state for a test scenario
type Scenario struct {
	Namespace string

	CapturedRequest  *http.CapturedRequest
	CapturedResponse *http.CapturedResponse

	IPOrFQDN string
}

// New creates a new state to use in a test Scenario
func New() *Scenario {
	return &Scenario{}
}

// CaptureRoundTrip will perform an HTTP request and return the CapturedRequest
// and CapturedResponse tuple
func (s *Scenario) CaptureRoundTrip(method, scheme, hostname, path string) error {
	capturedRequest, capturedResponse, err := http.CaptureRoundTrip(method, scheme, hostname, path, s.IPOrFQDN)
	if err != nil {
		return err
	}

	s.CapturedRequest = capturedRequest
	s.CapturedResponse = capturedResponse

	return nil
}

// AssertStatusCode returns an error if the captured response status code does
// not match the expected value
func (s *Scenario) AssertStatusCode(statusCode int) error {
	if s.CapturedResponse.StatusCode != statusCode {
		return fmt.Errorf("expected status code %v but %v was returned", statusCode, s.CapturedResponse.StatusCode)
	}

	return nil
}

// AssertServedBy returns an error if the captured request was not served by the
// expected service
func (s *Scenario) AssertServedBy(service string) error {
	if s.CapturedRequest.Service != service {
		return fmt.Errorf("expected the request to be served by %v but it was served by %v", service, s.CapturedRequest.Service)
	}

	return nil
}

// AssertRequestHost returns an error if the captured request host does not
// match the expected value
func (s *Scenario) AssertRequestHost(host string) error {
	if s.CapturedRequest.Host != host {
		return fmt.Errorf("expected the request host to be %v but was %v", host, s.CapturedRequest.Host)
	}

	return nil
}

// AssertTLSHostname returns an error if the captured TLS response hostname does
// not match the expected value
func (s *Scenario) AssertTLSHostname(hostname string) error {
	if s.CapturedResponse.TLSHostname != hostname {
		return fmt.Errorf("expected the response TLS hostname to be %v but was %v", hostname, s.CapturedResponse.TLSHostname)
	}

	return nil
}

// AssertResponseProto returns an error if the captured response proto does not
// match the expected value
func (s *Scenario) AssertResponseProto(proto string) error {
	if s.CapturedResponse.Proto != proto {
		return fmt.Errorf("expected the response protocol to be %v but it was %v", proto, s.CapturedResponse.Proto)
	}

	return nil
}

// AssertRequestProto returns an error if the captured request proto does not
// match the expected value
func (s *Scenario) AssertRequestProto(proto string) error {
	if s.CapturedRequest.Proto != proto {
		return fmt.Errorf("expected the request protocol to be %v but it was %v", proto, s.CapturedRequest.Proto)
	}

	return nil
}

// AssertMethod returns an error if the captured request method does not match
// the expected value
func (s *Scenario) AssertMethod(method string) error {
	if s.CapturedRequest.Method != method {
		return fmt.Errorf("expected the request method to be %v but it was %v", method, s.CapturedRequest.Method)
	}

	return nil
}

// AssertRequestPath returns an error if the captured request path does not
// match the expected value
func (s *Scenario) AssertRequestPath(path string) error {
	if !strings.HasPrefix(path, "/") {
		path = fmt.Sprintf("/%s", path)
	}

	if s.CapturedRequest.Path != path {
		return fmt.Errorf("expected the request path to be %v but it was %v", path, s.CapturedRequest.Path)
	}

	return nil
}

// AssertResponseHeader returns an error if the captured response headers do not
// contain the expected headerKey, or if the matching response header value does
// not match the expected headerValue. If the headerValue string equals `*`, the
// header value check is ignored.
func (s *Scenario) AssertResponseHeader(headerKey string, headerValue string) error {
	if headerValues := s.CapturedResponse.Headers[headerKey]; headerValues == nil {
		return fmt.Errorf("expected response headers to contain %v but it only contained %v", headerKey, s.CapturedResponse.Headers)
	} else if headerValue != "*" {
		for _, value := range headerValues {
			if value == headerValue {
				return nil
			}
		}

		return fmt.Errorf("expected response headers %v to contain a %v value but it contained %v", headerKey, headerValue, headerValues)
	}

	return nil
}

// AssertRequestHeader returns an error if the captured request headers do not
// contain the expected headerKey, or if the matching request header value does
// not match the expected headerValue. If the headerValue string equals `*`, the
// header value check is ignored.
func (s *Scenario) AssertRequestHeader(headerKey string, headerValue string) error {
	if headerValues := s.CapturedRequest.Headers[headerKey]; headerValues == nil {
		return fmt.Errorf("expected request headers to contain %v but it only contained %v", headerKey, s.CapturedRequest.Headers)
	} else if headerValue != "*" {
		for _, value := range headerValues {
			if value == headerValue {
				return nil
			}
		}

		return fmt.Errorf("expected request headers %v to contain a %v value but it contained %v", headerKey, headerValue, headerValues)
	}

	return nil
}

// AssertResponseCertificate returns nil if the captured certificate for the
// named host is valid. Otherwise it returns an error describing the mismatch.
func (s *Scenario) AssertResponseCertificate(hostname string) error {
	if s.CapturedResponse == nil || s.CapturedResponse.Certificate == nil {
		return fmt.Errorf("hostname verification requires executing a request and also target an HTTPS URL")
	}

	return s.CapturedResponse.Certificate.VerifyHostname(hostname)
}
