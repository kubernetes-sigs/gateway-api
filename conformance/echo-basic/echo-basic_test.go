/*
Copyright 2023 The Kubernetes Authors.

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

package main

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestHealthHandler(t *testing.T) {
	// Test a valid health check
	req, err := http.NewRequest("GET", "/health", nil) //nolint:noctx
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(healthHandler)
	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Expected status code %d, but got %d", http.StatusOK, status)
	}

	expectedResponse := "OK"
	if rr.Body.String() != expectedResponse {
		t.Errorf("Expected response body %s, but got %s", expectedResponse, rr.Body.String())
	}
}

func TestDelayResponse(t *testing.T) {
	// Test with a valid delay integer value
	req, err := http.NewRequest("GET", "/?delay=1s", nil) //nolint:noctx
	if err != nil {
		t.Fatal(err)
	}

	err = delayResponse(req)
	if err != nil {
		t.Errorf("Expected no error, but got %v", err)
	}

	// Test with a valid delay decimal value
	req, err = http.NewRequest("GET", "/?delay=0.1s", nil) //nolint:noctx
	if err != nil {
		t.Fatal(err)
	}

	err = delayResponse(req)
	if err != nil {
		t.Errorf("Expected no error, but got %v", err)
	}

	// Test with an invalid delay value
	req, err = http.NewRequest("GET", "/?delay=invalid", nil) //nolint:noctx
	if err != nil {
		t.Fatal(err)
	}

	err = delayResponse(req)
	if err == nil {
		t.Errorf("Expected an error, but got nil")
	}
}

func TestStatusHandler(t *testing.T) {
	// Test with a valid status code (200)
	req := httptest.NewRequest("GET", "/status/200", nil)
	rr := httptest.NewRecorder()

	statusHandler(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Expected status code %d, but got %d", http.StatusOK, status)
	}

	// Test with a valid status code (404)
	req = httptest.NewRequest("GET", "/status/404", nil)
	rr = httptest.NewRecorder()

	statusHandler(rr, req)

	if status := rr.Code; status != http.StatusNotFound {
		t.Errorf("Expected status code %d, but got %d", http.StatusNotFound, status)
	}

	// Test with a valid status (305)
	req = httptest.NewRequest("GET", "/status/305", nil)
	rr = httptest.NewRecorder()

	statusHandler(rr, req)

	if status := rr.Code; status != http.StatusUseProxy {
		t.Errorf("Expected status code %d, but got %d", http.StatusUseProxy, status)
	}

	// Test with an invalid status code (string)
	req = httptest.NewRequest("GET", "/status/invalid", nil)
	rr = httptest.NewRecorder()

	statusHandler(rr, req)

	if status := rr.Code; status != http.StatusBadRequest {
		t.Errorf("Expected status code %d, but got %d", http.StatusBadRequest, status)
	}
}

func TestEchoHandler(t *testing.T) {
	// Create an HTTP request to the / endpoint
	req := httptest.NewRequest("GET", "/", nil)
	rr := httptest.NewRecorder()

	// Set up the context
	context = Context{
		Namespace: "testNamespace",
		Ingress:   "testIngress",
		Service:   "testService",
		Pod:       "testPod",
	}

	echoHandler(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Expected status code %d, but got %d", http.StatusOK, status)
	}

	// Test response headers have correct ContentType
	expectedContentType := "application/json"
	if contentType := rr.Header().Get("Content-Type"); contentType != expectedContentType {
		t.Errorf("Expected Content-Type header %s, but got %s", expectedContentType, contentType)
	}

	// Test response headers have correct ContentTypeOptions
	expectedXContentTypeOptions := "nosniff"
	if xContentTypeOptions := rr.Header().Get("X-Content-Type-Options"); xContentTypeOptions != expectedXContentTypeOptions {
		t.Errorf("Expected X-Content-Type-Options header %s, but got %s", expectedXContentTypeOptions, xContentTypeOptions)
	}

	// Test the response body by unmarshalling it into a RequestAssertions struct
	var responseAssertions RequestAssertions
	err := json.Unmarshal(rr.Body.Bytes(), &responseAssertions)
	if err != nil {
		t.Errorf("Error unmarshalling response body: %v", err)
	}

	// Test RequestAssertions struct contains expected path
	expectedPath := "/"
	if responseAssertions.Path != expectedPath {
		t.Errorf("Expected Path %s, but got %s", expectedPath, responseAssertions.Path)
	}

	// Test RequestAssertions struct contains expected context namespace
	expectedNamespace := context.Namespace
	if responseAssertions.Context.Namespace != expectedNamespace {
		t.Errorf("Expected X-Content-Type-Options header %s, but got %s", expectedNamespace, responseAssertions.Context.Namespace)
	}
}

func TestWriteEchoResponseHeaders(t *testing.T) {
	// Create a response recorder to capture the response
	rr := httptest.NewRecorder()
	headers := http.Header{}

	// Test with a single header
	headers.Add("X-Echo-Set-Header", "Header1:Value1")

	writeEchoResponseHeaders(rr, headers)

	// Test response headers, check header is correct
	expectedHeaderName := "Header1"
	expectedHeaderValue := "Value1"
	if headerValues := rr.Header().Values(expectedHeaderName); len(headerValues) != 1 || headerValues[0] != expectedHeaderValue {
		t.Errorf("Expected header %s:%s, but got %s:%s", expectedHeaderName, expectedHeaderValue, headerValues[0], headerValues[1])
	}

	// Test with multiple headers separated by commas
	rr = httptest.NewRecorder()
	headers = http.Header{}
	headers.Add("X-Echo-Set-Header", "Header2:Value2, Header3:Value3")

	writeEchoResponseHeaders(rr, headers)

	// Test response headers, check headers are correct
	expectedHeaderName2 := "Header2"
	expectedHeaderValue2 := "Value2"
	expectedHeaderName3 := "Header3"
	expectedHeaderValue3 := "Value3"
	headerValues2 := rr.Header().Values(expectedHeaderName2)
	headerValues3 := rr.Header().Values(expectedHeaderName3)
	if len(headerValues2) != 1 || headerValues2[0] != expectedHeaderValue2 ||
		len(headerValues3) != 1 || headerValues3[0] != expectedHeaderValue3 {
		t.Errorf("Expected header %s:%s and header %s:%s, but got %s:%s and %s:%s",
			expectedHeaderName2, expectedHeaderValue2, expectedHeaderName3, expectedHeaderValue3,
			headerValues2[0], headerValues2[1], headerValues3[0], headerValues3[1])
	}

	// Test empty header
	rr = httptest.NewRecorder()
	headers = http.Header{}
	headers.Add("X-Echo-Set-Header", "")

	writeEchoResponseHeaders(rr, headers)

	// Test response headers is not set
	headerValuesEmpty := rr.Header().Values(expectedHeaderName)
	if len(headerValuesEmpty) != 0 {
		t.Errorf("Expected no header %s, but got %s:%s", expectedHeaderName, expectedHeaderName, headerValuesEmpty[0])
	}
}

func TestProcessError(t *testing.T) {
	// Create a response recorder to capture the response
	rr := httptest.NewRecorder()
	err := fmt.Errorf("Test error message")
	code := http.StatusInternalServerError

	processError(rr, err, code)

	// Test response status code is StatusInternalServerError
	if status := rr.Code; status != code {
		t.Errorf("Expected status code %d, but got %d", code, status)
	}

	// Test Content-Type header
	expectedContentType := "application/json"
	if contentType := rr.Header().Get("Content-Type"); contentType != expectedContentType {
		t.Errorf("Expected Content-Type header %s, but got %s", expectedContentType, contentType)
	}

	// Test X-Content-Type-Options header
	expectedXContentTypeOptions := "nosniff"
	if xContentTypeOptions := rr.Header().Get("X-Content-Type-Options"); xContentTypeOptions != expectedXContentTypeOptions {
		t.Errorf("Expected X-Content-Type-Options header %s, but got %s", expectedXContentTypeOptions, xContentTypeOptions)
	}

	// Test the response body by unmarshalling it into a responseBody struct
	var responseBody struct {
		Message string `json:"message"`
	}
	err = json.Unmarshal(rr.Body.Bytes(), &responseBody)
	if err != nil {
		t.Errorf("Error unmarshalling response body: %v", err)
	}

	// Test response body error message
	expectedErrorMessage := "Test error message"
	if responseBody.Message != expectedErrorMessage {
		t.Errorf("Expected error message '%s', but got '%s'", expectedErrorMessage, responseBody.Message)
	}
}

func TestProcessErrorWithJSONError(t *testing.T) {
	// Create a response recorder to capture the response
	rr := httptest.NewRecorder()
	err := json.Unmarshal([]byte(`testing invalid JSON`), new(interface{}))
	code := http.StatusInternalServerError

	processError(rr, err, code)

	// Test response status code is StatusInternalServerError
	if status := rr.Code; status != code {
		t.Errorf("Expected status code %d, but got %d", code, status)
	}

	// Test Content-Type header
	expectedContentType := "application/json"
	if contentType := rr.Header().Get("Content-Type"); contentType != expectedContentType {
		t.Errorf("Expected Content-Type header %s, but got %s", expectedContentType, contentType)
	}

	// Test X-Content-Type-Options header
	expectedXContentTypeOptions := "nosniff"
	if xContentTypeOptions := rr.Header().Get("X-Content-Type-Options"); xContentTypeOptions != expectedXContentTypeOptions {
		t.Errorf("Expected X-Content-Type-Options header %s, but got %s", expectedXContentTypeOptions, xContentTypeOptions)
	}

	// Test the response body by unmarshalling it into a responseBody struct
	var responseBody struct {
		Message string `json:"message"`
	}
	err = json.Unmarshal(rr.Body.Bytes(), &responseBody)
	if err != nil {
		t.Errorf("Error unmarshalling response body: %v", err)
	}

	// Test response body error message
	if !strings.Contains(responseBody.Message, "invalid character") {
		t.Errorf("Expected error message to contain 'invalid character', but got '%s'", responseBody.Message)
	}
}

func TestTLSStateToAssertions(t *testing.T) {
	// Create a mock TLS connection state for testing
	mockState := &tls.ConnectionState{
		Version:            tls.VersionTLS13,
		NegotiatedProtocol: "http/2",
		ServerName:         "test-example.com",
		CipherSuite:        tls.TLS_AES_256_GCM_SHA384,
		PeerCertificates:   []*x509.Certificate{},
	}

	// Call the function to convert the mock state
	result := tlsStateToAssertions(mockState)

	// Define expected TLS state
	expected := &TLSAssertions{
		Version:            "TLSv1.3",
		NegotiatedProtocol: "http/2",
		ServerName:         "test-example.com",
		CipherSuite:        "TLS_AES_256_GCM_SHA384",
		PeerCertificates:   []string{},
	}

	// Test the converted result is matching the expected
	if !compareTLSAssertions(result, expected) {
		t.Errorf("Result does not match expected values.\nGot: %+v\nExpected: %+v", result, expected)
	}
}

func compareTLSAssertions(a, b *TLSAssertions) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	return a.Version == b.Version &&
		a.NegotiatedProtocol == b.NegotiatedProtocol &&
		a.ServerName == b.ServerName &&
		a.CipherSuite == b.CipherSuite &&
		strings.Join(a.PeerCertificates, "") == strings.Join(b.PeerCertificates, "")
}
