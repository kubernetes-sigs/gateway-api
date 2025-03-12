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
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"testing"
	"time"

	"k8s.io/apimachinery/pkg/types"

	h "sigs.k8s.io/gateway-api/conformance/utils/http"
	"sigs.k8s.io/gateway-api/conformance/utils/kubernetes"
	"sigs.k8s.io/gateway-api/conformance/utils/roundtripper"
	"sigs.k8s.io/gateway-api/conformance/utils/suite"
	"sigs.k8s.io/gateway-api/pkg/features"
)

func init() {
	ConformanceTests = append(ConformanceTests, BackendTLSPolicy)
}

var BackendTLSPolicy = suite.ConformanceTest{
	ShortName:   "BackendTLSPolicy",
	Description: "A single service that is targeted by a BackendTLSPolicy must successfully complete TLS termination",
	Features: []features.FeatureName{
		features.SupportGateway,
		features.SupportBackendTLSPolicy,
	},
	Manifests: []string{"tests/backendtlspolicy.yaml"},
	Test: func(t *testing.T, suite *suite.ConformanceTestSuite) {
		ns := "gateway-conformance-infra"
		routeNN := types.NamespacedName{Name: "gateway-conformance-infra-test", Namespace: ns}
		gwNN := types.NamespacedName{Name: "gateway-backendtlspolicy", Namespace: ns}

		kubernetes.NamespacesMustBeReady(t, suite.Client, suite.TimeoutConfig, []string{ns})
		gwAddr := kubernetes.GatewayAndHTTPRoutesMustBeAccepted(t, suite.Client, suite.TimeoutConfig, suite.ControllerName, kubernetes.NewGatewayRef(gwNN), routeNN)
		kubernetes.HTTPRouteMustHaveResolvedRefsConditionsTrue(t, suite.Client, suite.TimeoutConfig, routeNN, gwNN)

		serverStr := "abc.example.com"
		headers := make(map[string]string)
		headers["Host"] = serverStr

		// Verify that the response to a call to /backendTLS will return the matching SNI.
		t.Run("Simple request targeting BackendTLSPolicy should reach infra-backend", func(t *testing.T) {
			expected := h.ExpectedResponse{
				Request: h.Request{
					Headers: headers,
					Host:    serverStr,
					Path:    "/backendTLS",
				},
				Response: h.Response{StatusCode: 200},
			}
			req := h.MakeRequest(t, &expected, gwAddr, "HTTPS", "https")

			if found, err := sameSNI(req, serverStr); err != nil || !found {
				t.Errorf("no SNI found for request: %v", err)
			}
		})
	},
}

func sameSNI(request roundtripper.Request, sni string) (bool, error) {
	client := &http.Client{}
	assertions := &kubernetes.RequestAssertions{}

	method := "GET"
	if request.Method != "" {
		method = request.Method
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, method, request.URL.String(), nil)
	if err != nil {
		return false, fmt.Errorf("error creating request: %v", err)
	}

	if request.Host != "" {
		req.Host = request.Host
	}

	if request.Headers != nil {
		for name, value := range request.Headers {
			req.Header.Set(name, value[0])
		}
	}

	resp, err := client.Do(req)
	if err != nil {
		return false, fmt.Errorf("request %v caused an error %v", req, err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return false, fmt.Errorf("failed to read response body: %v", err)
	}

	err = json.Unmarshal(body, assertions)
	if err != nil {
		return false, fmt.Errorf("unexpected error reading response: %w", err)
	}

	return assertions.SNI == sni, nil
}
