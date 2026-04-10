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
	"fmt"
	stdhttp "net/http"
	"testing"

	"k8s.io/apimachinery/pkg/types"

	"sigs.k8s.io/gateway-api/conformance/utils/http"
	"sigs.k8s.io/gateway-api/conformance/utils/kubernetes"
	"sigs.k8s.io/gateway-api/conformance/utils/suite"
	"sigs.k8s.io/gateway-api/pkg/features"
)

func init() {
	ConformanceTests = append(ConformanceTests, HTTPRouteSessionPersistenceCookie)
}

var HTTPRouteSessionPersistenceCookie = suite.ConformanceTest{
	ShortName:   "HTTPRouteSessionPersistenceCookie",
	Description: "An HTTPRoute with cookie-based session persistence routes requests with the same cookie to the same backend",
	Manifests:   []string{"tests/httproute-session-persistence-cookie.yaml"},
	Features: []features.FeatureName{
		features.SupportGateway,
		features.SupportHTTPRoute,
		features.SupportHTTPRouteSessionPersistenceCookie,
	},
	Test: func(t *testing.T, suite *suite.ConformanceTestSuite) {
		const (
			ns   = "gateway-conformance-infra"
			path = "/session-persistence"
		)
		routeNN := types.NamespacedName{Name: "session-persistence-cookie", Namespace: ns}
		gwNN := types.NamespacedName{Name: "same-namespace", Namespace: ns}
		gwAddr := kubernetes.GatewayAndHTTPRoutesMustBeAccepted(t, suite.Client, suite.TimeoutConfig, suite.ControllerName, kubernetes.NewGatewayRef(gwNN), routeNN)
		kubernetes.HTTPRouteMustHaveResolvedRefsConditionsTrue(t, suite.Client, suite.TimeoutConfig, routeNN, gwNN)

		expected := http.ExpectedResponse{
			Request: http.Request{
				Path: path,
			},
			Response: http.Response{
				StatusCode: 200,
			},
			Namespace: ns,
		}

		http.MakeRequestAndExpectEventuallyConsistentResponse(t, suite.RoundTripper, suite.TimeoutConfig, gwAddr, expected)

		req := http.MakeRequest(t, &expected, gwAddr, "HTTP", "http")
		initialPod := ""
		for i := 0; i < 10; i++ {
			cReq, cRes, err := suite.RoundTripper.CaptureRoundTrip(req)
			if err != nil {
				t.Fatalf("request %d with cookie failed: %v", i+1, err)
			}
			if err := http.CompareRoundTrip(t, &req, cReq, cRes, expected); err != nil {
				t.Fatalf("request %d with cookie failed expectations: %v", i+1, err)
			}

			if i == 0 {
				if cReq.Pod == "" {
					t.Fatalf("expected pod to be set")
				}
				cookie, err := parseCookie(cRes.Headers)
				if err != nil {
					t.Fatalf("failed to parse session persistence cookie: %v", err)
				}
				t.Logf("session persistence cookie: %s=%s", cookie.Name, cookie.Value)
				req.Headers["Cookie"] = []string{fmt.Sprintf("%s=%s", cookie.Name, cookie.Value)}
				initialPod = cReq.Pod
				continue
			}
			if cReq.Pod != initialPod {
				t.Fatalf("expected session persistence to keep routing to pod %q, got %q", initialPod, cReq.Pod)
			}
		}
	},
}

func parseCookie(headers map[string][]string) (*stdhttp.Cookie, error) {
	parser := &stdhttp.Response{Header: stdhttp.Header(headers)}
	cookies := parser.Cookies()
	if len(cookies) == 0 {
		return nil, fmt.Errorf("cookie not found: headers: %v", headers)
	}
	return cookies[0], nil
}
