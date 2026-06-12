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
	ConformanceTests = append(ConformanceTests, GRPCRouteSessionPersistenceCookiePermanentLifetime)
}

var GRPCRouteSessionPersistenceCookiePermanentLifetime = confsuite.ConformanceTest{
	ShortName:   "GRPCRouteSessionPersistenceCookiePermanentLifetime",
	Description: "A GRPCRoute with permanent cookie lifetime sets a cookie with an Expires or Max-Age attribute",
	Manifests:   []string{"tests/grpcroute-session-persistence-cookie-permanent-lifetime.yaml"},
	Provisional: true,
	Features: []features.FeatureName{
		features.SupportGateway,
		features.SupportGRPCRoute,
		features.SupportCookieSessionPersistence,
		features.SupportCookieSessionPersistencePermanentLifetime,
	},
	Test: func(t *testing.T, suite *confsuite.ConformanceTestSuite) {
		ns := confsuite.InfrastructureNamespace
		routeNN := types.NamespacedName{Name: "grpcroute-session-persistence-cookie-permanent-lifetime", Namespace: ns}
		gwNN := types.NamespacedName{Name: "same-namespace", Namespace: ns}
		gwAddr := kubernetes.GatewayAndRoutesMustBeAccepted(t, suite.Client, suite.TimeoutConfig, suite.ControllerName, kubernetes.NewGatewayRef(gwNN), &v1.GRPCRoute{}, true, routeNN)

		req := grpc.ExpectedResponse{
			EchoRequest: &pb.EchoRequest{},
		}

		// Wait for the route to stabilize and capture the permanent cookie.
		var cookie *httputils.CookieInfo
		httputils.AwaitConvergence(t, 1, suite.TimeoutConfig.MaxTimeToConsistency, func(elapsed time.Duration) bool {
			client := &grpc.DefaultClient{}
			defer client.Close()
			resp, err := client.SendRPC(t, gwAddr, req, suite.TimeoutConfig.RequestTimeout)
			if err != nil {
				tlog.Logf(t, "request failed: %v (after %v)", err, elapsed)
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
			tlog.Logf(t, "received set-cookie: %v attributes=%v", resp.Headers.Get("set-cookie"), cookie.Attributes)
			return true
		})

		// Permanent cookies must carry either Expires or Max-Age to communicate the expiry time to clients.
		if !cookie.HasAttribute("Expires") && !cookie.HasAttribute("Max-Age") {
			t.Errorf("permanent cookie must have an Expires or Max-Age attribute, got attributes: %v", cookie.Attributes)
		}
	},
}
