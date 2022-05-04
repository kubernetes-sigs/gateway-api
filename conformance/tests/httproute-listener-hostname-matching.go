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

package tests

import (
	"testing"

	"k8s.io/apimachinery/pkg/types"

	"sigs.k8s.io/gateway-api/conformance/utils/http"
	"sigs.k8s.io/gateway-api/conformance/utils/kubernetes"
	"sigs.k8s.io/gateway-api/conformance/utils/suite"
)

func init() {
	ConformanceTests = append(ConformanceTests, HTTPRouteListenerHostnameMatching)
}

var HTTPRouteListenerHostnameMatching = suite.ConformanceTest{
	ShortName:   "HTTPRouteListenerHostnameMatching",
	Description: "Multiple HTTP listeners with the same port and different hostnames, each with a different HTTPRoute",
	Manifests:   []string{"tests/httproute-listener-hostname-matching.yaml"},
	Test: func(t *testing.T, suite *suite.ConformanceTestSuite) {
		ns := "gateway-conformance-infra"

		// This test creates an additional Gateway in the gateway-conformance-infra
		// namespace so we have to wait for it to be ready.
		kubernetes.NamespacesMustBeReady(t, suite.Client, []string{ns}, 300)

		gwNN := types.NamespacedName{Name: "httproute-listener-hostname-matching", Namespace: ns}
		routes := []types.NamespacedName{
			{Namespace: ns, Name: "backend-v1"},
			{Namespace: ns, Name: "backend-v2"},
			{Namespace: ns, Name: "backend-v3"},
		}
		gwAddr := kubernetes.GatewayAndHTTPRoutesMustBeReady(t, suite.Client, suite.ControllerName, gwNN, routes...)

		testCases := []http.ExpectedResponse{{
			Request:   http.ExpectedRequest{Host: "bar.com", Path: "/"},
			Backend:   "infra-backend-v1",
			Namespace: ns,
		}, {
			Request:   http.ExpectedRequest{Host: "foo.bar.com", Path: "/"},
			Backend:   "infra-backend-v2",
			Namespace: ns,
		}, {
			Request:   http.ExpectedRequest{Host: "baz.bar.com", Path: "/"},
			Backend:   "infra-backend-v3",
			Namespace: ns,
		}, {
			Request:   http.ExpectedRequest{Host: "boo.bar.com", Path: "/"},
			Backend:   "infra-backend-v3",
			Namespace: ns,
		}, {
			Request:    http.ExpectedRequest{Host: "too.many.prefixes.bar.com", Path: "/"},
			StatusCode: 404,
		}, {
			Request:    http.ExpectedRequest{Host: "no.matching.host", Path: "/"},
			StatusCode: 404,
		}}

		for i := range testCases {
			// Declare tc here to avoid loop variable
			// reuse issues across parallel tests.
			tc := testCases[i]
			t.Run(testName(tc, i), func(t *testing.T) {
				t.Parallel()
				http.MakeRequestAndExpectEventuallyConsistentResponse(t, suite.RoundTripper, gwAddr, tc)
			})
		}
	},
}
