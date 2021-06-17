/*
Copyright 2020 The Kubernetes Authors.

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
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httputil"
	"regexp"
	"strings"
	"time"
)

var (
	// HTTPClientTimeout specifies a time limit for requests made by a client
	HTTPClientTimeout = 10 * time.Second
	// EnableDebug enable dump of requests and responses of HTTP requests
	// (useful for debug)
	EnableDebug = false
)

// CapturedRequest contains the original HTTP request metadata as received by
// the echoserver handling the test request.
type CapturedRequest struct {
	Path    string              `json:"path"`
	Host    string              `json:"host"`
	Method  string              `json:"method"`
	Proto   string              `json:"proto"`
	Headers map[string][]string `json:"headers"`

	Namespace string `json:"namespace"`
	Service   string `json:"service"`
	Pod       string `json:"pod"`
}

// CapturedResponse contains the HTTP response metadata from the echoserver.
type CapturedResponse struct {
	StatusCode    int
	ContentLength int64
	Proto         string
	Headers       map[string][]string
	TLSHostname   string

	Certificate *x509.Certificate
}

// CaptureRoundTrip will perform an HTTP request and return the CapturedRequest
// and CapturedResponse tuple
func CaptureRoundTrip(method, scheme, hostname, path, location string) (*CapturedRequest, *CapturedResponse, error) {
	var capturedTLSHostname string
	var certificate *x509.Certificate

	tr := &http.Transport{
		DisableCompression: true,
		TLSClientConfig: &tls.Config{
			// Skip all usual TLS verifications, since we are using self-signed
			// certificates.
			// nolint:gosec
			InsecureSkipVerify: true,
			VerifyPeerCertificate: func(certificates [][]byte, _ [][]*x509.Certificate) error {
				certs := make([]*x509.Certificate, len(certificates))
				for i, asn1Data := range certificates {
					cert, err := x509.ParseCertificate(asn1Data)
					if err != nil {
						return fmt.Errorf("tls: failed to parse certificate from server: " + err.Error())
					}
					certs[i] = cert
				}

				capturedTLSHostname = certs[0].DNSNames[0]
				certificate = certs[0]
				return nil
			},
		},
	}

	if scheme == "https" && hostname != "" {
		tr.TLSClientConfig.ServerName = hostname
	}

	client := &http.Client{
		Transport: tr,
		Timeout:   HTTPClientTimeout,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	url := fmt.Sprintf("%s://%s/%s", scheme, location, strings.TrimPrefix(path, "/"))

	req, reqErr := http.NewRequestWithContext(context.TODO(), method, url, nil)
	if reqErr != nil {
		return nil, nil, reqErr
	}

	if hostname != "" {
		req.Host = hostname
	}

	if EnableDebug {
		dump, dumpErr := httputil.DumpRequestOut(req, true)
		if dumpErr != nil {
			return nil, nil, dumpErr
		}

		fmt.Printf("Sending request:\n%s\n\n", formatDump(dump, "> "))
	}

	resp, respErr := client.Do(req)
	if respErr != nil {
		return nil, nil, respErr
	}
	defer resp.Body.Close()

	if EnableDebug {
		dump, dumpErr := httputil.DumpResponse(resp, true)
		if dumpErr != nil {
			return nil, nil, dumpErr
		}

		fmt.Printf("Received response:\n%s\n\n", formatDump(dump, "< "))
	}

	// check if the result is a redirect and return a new request this avoids
	// the issue of URLs without valid DNS names and also sends the traffic to
	// the ingress controller IP address or FQDN
	if isRedirect(resp.StatusCode) {
		redirectURL, err := resp.Location()
		if err != nil {
			return nil, nil, err
		}

		return CaptureRoundTrip(method, redirectURL.Scheme, redirectURL.Hostname(), redirectURL.Path, location)
	}

	capReq := CapturedRequest{}
	body, _ := ioutil.ReadAll(resp.Body)

	// we cannot assume the response is JSON
	if isJSON(body) {
		jsonErr := json.Unmarshal(body, &capReq)
		if jsonErr != nil {
			return nil, nil, fmt.Errorf("unexpected error reading response: %w", jsonErr)
		}
	}

	capRes := &CapturedResponse{
		resp.StatusCode,
		resp.ContentLength,
		resp.Proto,
		resp.Header,
		capturedTLSHostname,
		certificate,
	}

	return &capReq, capRes, nil
}

func isJSON(content []byte) bool {
	var js map[string]interface{}
	return json.Unmarshal(content, &js) == nil
}

func isRedirect(statusCode int) bool {
	switch statusCode {
	case http.StatusMovedPermanently,
		http.StatusFound,
		http.StatusSeeOther,
		http.StatusTemporaryRedirect,
		http.StatusPermanentRedirect:
		return true
	}

	return false
}

var startLineRegex = regexp.MustCompile(`(?m)^`)

func formatDump(data []byte, prefix string) string {
	data = startLineRegex.ReplaceAllLiteral(data, []byte(prefix))
	return string(data)
}
