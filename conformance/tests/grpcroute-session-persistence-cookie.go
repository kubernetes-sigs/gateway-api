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
	"testing"
	"time"

	"k8s.io/apimachinery/pkg/types"

	v1 "sigs.k8s.io/gateway-api/apis/v1"
	pb "sigs.k8s.io/gateway-api/conformance/echo-basic/grpcechoserver"
	"sigs.k8s.io/gateway-api/conformance/utils/grpc"
	httputils "sigs.k8s.io/gateway-api/conformance/utils/http"
	"sigs.k8s.io/gateway-api/conformance/utils/kubernetes"
	"sigs.k8s.io/gateway-api/conformance/utils/tlog"
	confsuite "sigs.k8s.io/gateway-api/conformance/utils/suite"
	"sigs.k8s.io/gateway-api/pkg/features"
)

func init() {
	ConformanceTests = append(ConformanceTests, GRPCRouteSessionPersistenceCookie)
}

var GRPCRouteSessionPersistenceCookie = confsuite.ConformanceTest{
	ShortName:   "GRPCRouteSessionPersistenceCookie",
	Description: "A GRPCRoute with cookie-based session persistence sets a session cookie (no Expires or Max-Age) and routes all subsequent requests carrying that cookie to the same backend pod",
	Manifests:   []string{"tests/grpcroute-session-persistence-cookie.yaml"},
	Provisional: true,
	Features: []features.FeatureName{
		features.SupportGateway,
		features.SupportGRPCRoute,
		features.SupportCookieSessionPersistence,
	},
	Test: func(t *testing.T, suite *confsuite.ConformanceTestSuite) {
		ns := confsuite.InfrastructureNamespace
		routeNN := types.NamespacedName{Name: "grpcroute-session-persistence-cookie", Namespace: ns}
		gwNN := types.NamespacedName{Name: "same-namespace", Namespace: ns}
		gwAddr := kubernetes.GatewayAndRoutesMustBeAccepted(t, suite.Client, suite.TimeoutConfig, suite.ControllerName, kubernetes.NewGatewayRef(gwNN), &v1.GRPCRoute{}, true, routeNN)

		baseReq := grpc.ExpectedResponse{
			EchoRequest: &pb.EchoRequest{},
		}

		// Wait for the route to be data-plane ready and capture the first session cookie.
		var cookie *httputils.CookieInfo
		var firstPod string
		httputils.AwaitConvergence(t, 1, suite.TimeoutConfig.MaxTimeToConsistency, func(elapsed time.Duration) bool {
			client := &grpc.DefaultClient{}
			defer client.Close()
			resp, err := client.SendRPC(t, gwAddr, baseReq, suite.TimeoutConfig.RequestTimeout)
			if err != nil {
				tlog.Logf(t, "request failed: %v (after %v)", err, elapsed)
				return false
			}
			if resp.Response == nil || resp.Response.Assertions == nil || resp.Response.Assertions.Context == nil {
				tlog.Logf(t, "empty response assertions (after %v)", elapsed)
				return false
			}
			if resp.Headers == nil {
				tlog.Logf(t, "nil response headers (after %v)", elapsed)
				return false
			}
			h := make(http.Header)
			for k, v := range *resp.Headers {
				h[http.CanonicalHeaderKey(k)] = v
			}
			c, err := httputils.ExtractResponseCookie(h)
			if err != nil {
				tlog.Logf(t, "no session cookie in response (after %v): %v", elapsed, err)
				return false
			}
			cookie = c
			firstPod = resp.Response.Assertions.Context.Pod
			tlog.Logf(t, "session established: set-cookie=%v pod=%s attributes=%v", resp.Headers.Get("set-cookie"), firstPod, cookie.Attributes)
			return true
		})

		if cookie.HasAttribute("Expires") {
			t.Errorf("session cookie must not have an Expires attribute, got attributes: %v", cookie.Attributes)
		}
		if cookie.HasAttribute("Max-Age") {
			t.Errorf("session cookie must not have a Max-Age attribute, got attributes: %v", cookie.Attributes)
		}

		// Verify stickiness: each sticky request uses a fresh connection so routing is driven
		// by the cookie, not connection-level affinity.
		stickyReq := grpc.ExpectedResponse{
			EchoRequest: &pb.EchoRequest{},
			RequestMetadata: &grpc.RequestMetadata{
				Metadata: map[string]string{
					"cookie": cookie.Name + "=" + cookie.Value,
				},
			},
		}
		for i := 0; i < 10; i++ {
			client := &grpc.DefaultClient{}
			resp, err := client.SendRPC(t, gwAddr, stickyReq, suite.TimeoutConfig.RequestTimeout)
			client.Close()
			if err != nil {
				t.Errorf("sticky request %d failed: %v", i+1, err)
				continue
			}
			if resp.Response == nil || resp.Response.Assertions == nil || resp.Response.Assertions.Context == nil {
				t.Errorf("sticky request %d: empty response assertions", i+1)
				continue
			}
			pod := resp.Response.Assertions.Context.Pod
			tlog.Logf(t, "sticky request %d: pod=%s", i+1, pod)
			if pod != firstPod {
				t.Errorf("sticky request %d: expected pod %q, got %q (session cookie not honored)", i+1, firstPod, pod)
			}
		}
	},
}
