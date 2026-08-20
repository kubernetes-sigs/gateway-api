/*
Copyright The Kubernetes Authors.

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
	"strings"
	"testing"

	"k8s.io/apimachinery/pkg/types"

	"sigs.k8s.io/gateway-api/conformance/utils/http"
	"sigs.k8s.io/gateway-api/conformance/utils/kubernetes"
	confsuite "sigs.k8s.io/gateway-api/conformance/utils/suite"
	"sigs.k8s.io/gateway-api/pkg/features"
)

func init() {
	ConformanceTests = append(ConformanceTests, HTTPRouteExternalAuthHTTPForwardBody)
}

// NOTE: GEP-1494 has an unresolved inconsistency between the ExternalAuthFilter.ForwardBody
// field comment (which says oversized bodies are rejected with 4xx) and the
// ForwardBodyConfig.MaxSize comment (which says the body is truncated). The over-size test
// therefore accepts both 413 and 403. Implementations that truncate will return 200 and fail
// that case; that is expected until the spec ambiguity is resolved.
var HTTPRouteExternalAuthHTTPForwardBody = confsuite.ConformanceTest{
	ShortName:   "HTTPRouteExternalAuthHTTPForwardBody",
	Description: "An HTTPRoute ExternalAuth HTTP filter honors forwardBody.maxSize: bodies within the limit are forwarded and confirmed via X-Auth-Received-Body-Size; bodies over the limit are rejected with 4xx; maxSize=0 withholds the body entirely",
	Manifests:   []string{"tests/httproute-external-auth-http-forward-body.yaml"},
	Features: []features.FeatureName{
		features.SupportGateway,
		features.SupportHTTPRoute,
		features.SupportHTTPRouteExternalAuth,
		features.SupportHTTPRouteExternalAuthHTTP,
		features.SupportHTTPRouteExternalAuthForwardBody,
	},
	Test: func(t *testing.T, suite *confsuite.ConformanceTestSuite) {
		ns := confsuite.InfrastructureNamespace
		routeNN := types.NamespacedName{Name: "external-auth-http-forward-body", Namespace: ns}
		gwNN := types.NamespacedName{Name: "same-namespace", Namespace: ns}
		gwAddr := kubernetes.GatewayAndHTTPRoutesMustBeAccepted(t, suite.Client, suite.TimeoutConfig, suite.ControllerName, kubernetes.NewGatewayRef(gwNN), routeNN)
		kubernetes.HTTPRouteMustHaveResolvedRefsConditionsTrue(t, suite.Client, suite.TimeoutConfig, routeNN, gwNN)

		testCases := []http.ExpectedResponse{
			{
				TestCaseName: "maxSize=100: body within limit (10 bytes) should be forwarded to the auth server",
				Request: http.Request{
					Method: "POST",
					Path:   "/http/allowed/forward-body",
					Body:   "0123456789",
				},
				ExpectedRequest: &http.ExpectedRequest{
					Request: http.Request{
						Method: "POST",
						Path:   "/http/allowed/forward-body",
						Headers: map[string]string{
							"X-Auth-Received-Body-Size": "10",
						},
					},
				},
				Backend:   confsuite.InfraBackendServiceNameV1,
				Namespace: ns,
				Response:  http.Response{StatusCode: 200},
			},
			{
				TestCaseName: "maxSize=100: body exceeding limit (200 bytes) should be rejected with 4xx",
				Request: http.Request{
					Method: "POST",
					Path:   "/http/allowed/forward-body",
					Body:   strings.Repeat("x", 200),
				},
				Response: http.Response{StatusCodes: []int{413, 403}},
			},
			{
				TestCaseName: "maxSize=0: allowed request with body must not be rejected and auth server must receive zero body bytes",
				Request: http.Request{
					Method: "POST",
					Path:   "/http/allowed/forward-body-zero",
					Body:   strings.Repeat("x", 200),
				},
				ExpectedRequest: &http.ExpectedRequest{
					Request: http.Request{
						Method: "POST",
						Path:   "/http/allowed/forward-body-zero",
						Headers: map[string]string{
							"X-Auth-Received-Body-Size": "0",
						},
					},
				},
				Backend:   confsuite.InfraBackendServiceNameV1,
				Namespace: ns,
				Response:  http.Response{StatusCode: 200},
			},
			{
				TestCaseName: "maxSize=0: denied request with body must still be denied and auth server must receive zero body bytes",
				Request: http.Request{
					Method: "POST",
					Path:   "/http/denied/forward-body-zero",
					Body:   strings.Repeat("x", 200),
				},
				Response: http.Response{
					StatusCode: 403,
					Headers: map[string]string{
						"X-Auth-Received-Body-Size": "0",
					},
				},
			},
		}
		for i := range testCases {
			tc := testCases[i]
			t.Run(tc.GetTestCaseName(i), func(t *testing.T) {
				t.Parallel()
				http.MakeRequestAndExpectEventuallyConsistentResponse(t, suite.RoundTripper, suite.TimeoutConfig, gwAddr, tc)
			})
		}
	},
}
