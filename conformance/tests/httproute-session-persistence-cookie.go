/*
Copyright 2026 The Kubernetes Authors.

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
	"net/http"
	"net/url"
	"testing"
	"time"

	"k8s.io/apimachinery/pkg/types"

	httputils "sigs.k8s.io/gateway-api/conformance/utils/http"
	"sigs.k8s.io/gateway-api/conformance/utils/kubernetes"
	"sigs.k8s.io/gateway-api/conformance/utils/roundtripper"
	"sigs.k8s.io/gateway-api/conformance/utils/tlog"
	confsuite "sigs.k8s.io/gateway-api/conformance/utils/suite"
	"sigs.k8s.io/gateway-api/pkg/features"
)

func init() {
	ConformanceTests = append(ConformanceTests, HTTPRouteSessionPersistenceCookie)
}

var HTTPRouteSessionPersistenceCookie = confsuite.ConformanceTest{
	ShortName:   "HTTPRouteSessionPersistenceCookie",
	Description: "An HTTPRoute with cookie-based session persistence sets a session cookie (no Expires or Max-Age) and routes all subsequent requests carrying that cookie to the same backend pod",
	Manifests:   []string{"tests/httproute-session-persistence-cookie.yaml"},
	Provisional: true,
	Features: []features.FeatureName{
		features.SupportGateway,
		features.SupportHTTPRoute,
		features.SupportCookieSessionPersistence,
	},
	Test: func(t *testing.T, suite *confsuite.ConformanceTestSuite) {
		routeNN := types.NamespacedName{Name: "session-persistence-cookie", Namespace: confsuite.InfrastructureNamespace}
		gwNN := types.NamespacedName{Name: "same-namespace", Namespace: confsuite.InfrastructureNamespace}
		gwAddr := kubernetes.GatewayAndHTTPRoutesMustBeAccepted(t, suite.Client, suite.TimeoutConfig, suite.ControllerName, kubernetes.NewGatewayRef(gwNN), routeNN)
		kubernetes.HTTPRouteMustHaveResolvedRefsConditionsTrue(t, suite.Client, suite.TimeoutConfig, routeNN, gwNN)

		baseReq := roundtripper.Request{
			T:      t,
			Method: "GET",
			URL: url.URL{
				Scheme: "http",
				Host:   httputils.CalculateHost(t, gwAddr, "http"),
				Path:   "/session-persistence",
			},
			Headers: map[string][]string{},
		}

		// Wait for the route to be data-plane ready and capture the first session cookie.
		var cookie *httputils.CookieInfo
		var firstPod string
		httputils.AwaitConvergence(t, 1, suite.TimeoutConfig.MaxTimeToConsistency, func(elapsed time.Duration) bool {
			cReq, cRes, err := suite.RoundTripper.CaptureRoundTrip(baseReq)
			if err != nil {
				tlog.Logf(t, "request failed: %v (after %v)", err, elapsed)
				return false
			}
			if cRes.StatusCode != 200 {
				tlog.Logf(t, "expected status 200, got %d (after %v)", cRes.StatusCode, elapsed)
				return false
			}
			c, err := httputils.ExtractResponseCookie(http.Header(cRes.Headers))
			if err != nil {
				tlog.Logf(t, "no session cookie in response (after %v): %v", elapsed, err)
				return false
			}
			cookie = c
			firstPod = cReq.Pod
			tlog.Logf(t, "session established: Set-Cookie=%q pod=%s attributes=%v", cRes.Headers["Set-Cookie"], firstPod, cookie.Attributes)
			return true
		})

		if cookie.HasAttribute("Expires") {
			t.Errorf("session cookie must not have an Expires attribute, got attributes: %v", cookie.Attributes)
		}
		if cookie.HasAttribute("Max-Age") {
			t.Errorf("session cookie must not have a Max-Age attribute, got attributes: %v", cookie.Attributes)
		}

		// Verify stickiness: all subsequent requests carrying the session cookie must reach the same pod.
		stickyReq := baseReq
		stickyReq.Headers = map[string][]string{
			"Cookie": {cookie.Name + "=" + cookie.Value},
		}
		for i := 0; i < 10; i++ {
			cReq, cRes, err := suite.RoundTripper.CaptureRoundTrip(stickyReq)
			if err != nil {
				t.Errorf("sticky request %d failed: %v", i+1, err)
				continue
			}
			tlog.Logf(t, "sticky request %d: status=%d pod=%s", i+1, cRes.StatusCode, cReq.Pod)
			if cRes.StatusCode != 200 {
				t.Errorf("sticky request %d: expected status 200, got %d", i+1, cRes.StatusCode)
				continue
			}
			if cReq.Pod != firstPod {
				t.Errorf("sticky request %d: expected pod %q, got %q (session cookie not honored)", i+1, firstPod, cReq.Pod)
			}
		}
	},
}
