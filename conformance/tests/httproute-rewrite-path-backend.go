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

	"sigs.k8s.io/gateway-api/conformance/utils/http"
	"sigs.k8s.io/gateway-api/conformance/utils/suite"
	"sigs.k8s.io/gateway-api/pkg/features"
)

func init() {
	ConformanceTests = append(ConformanceTests, HTTPRouteRewritePathBackend)
}

var HTTPRouteRewritePathBackend = suite.ConformanceTest{
	ShortName:   "HTTPRouteRewritePathBackend",
	Description: "An HTTPRoute with path rewrite filter on the backend ref",
	Manifests:   []string{"tests/httproute-rewrite-path-backend.yaml"},
	Features: []features.FeatureName{
		features.SupportGateway,
		features.SupportHTTPRoute,
		features.SupportHTTPRoutePathRewrite,
	},
	Provisional: true,
	Test: func(t *testing.T, s *suite.ConformanceTestSuite) {
		ns := "gateway-conformance-infra"
		gwAddr := suite.DefaultConformanceTestBoilerplate(t, s, ns, "rewrite-path-backend", "same-namespace")
		testCases := []http.ExpectedResponse{
			{
				Request: http.Request{
					Path: "/prefix/one/two",
				},
				ExpectedRequest: &http.ExpectedRequest{
					Request: http.Request{
						Path: "/one/two",
					},
				},
				Backend:   "infra-backend-v1",
				Namespace: ns,
			},
			{
				Request: http.Request{
					Path: "/strip-prefix/three",
				},
				ExpectedRequest: &http.ExpectedRequest{
					Request: http.Request{
						Path: "/three",
					},
				},
				Backend:   "infra-backend-v1",
				Namespace: ns,
			},
			{
				Request: http.Request{
					Path: "/strip-prefix",
				},
				ExpectedRequest: &http.ExpectedRequest{
					Request: http.Request{
						Path: "/",
					},
				},
				Backend:   "infra-backend-v1",
				Namespace: ns,
			},
			{
				Request: http.Request{
					Path: "/full/one/two",
				},
				ExpectedRequest: &http.ExpectedRequest{
					Request: http.Request{
						Path: "/one",
					},
				},
				Backend:   "infra-backend-v1",
				Namespace: ns,
			},
			// {
			// 	Request: http.Request{
			// 		Path: "/full/rewrite-path-and-modify-headers/test",
			// 		Headers: map[string]string{
			// 			"X-Header-Remove":     "remove-val",
			// 			"X-Header-Add-Append": "append-val-1",
			// 			"X-Header-Set":        "set-val",
			// 		},
			// 	},
			// 	ExpectedRequest: &http.ExpectedRequest{
			// 		Request: http.Request{
			// 			Path: "/test",
			// 			Headers: map[string]string{
			// 				"X-Header-Add":        "header-val-1",
			// 				"X-Header-Add-Append": "append-val-1,header-val-2",
			// 				"X-Header-Set":        "set-overwrites-values",
			// 			},
			// 		},
			// 		AbsentHeaders: []string{"X-Header-Remove"},
			// 	},
			// 	Backend:   "infra-backend-v1",
			// 	Namespace: ns,
			// },
			// {
			// 	Request: http.Request{
			// 		Path: "/prefix/rewrite-path-and-modify-headers/one",
			// 		Headers: map[string]string{
			// 			"X-Header-Remove":     "remove-val",
			// 			"X-Header-Add-Append": "append-val-1",
			// 			"X-Header-Set":        "set-val",
			// 		},
			// 	},
			// 	ExpectedRequest: &http.ExpectedRequest{
			// 		Request: http.Request{
			// 			Path: "/prefix/one",
			// 			Headers: map[string]string{
			// 				"X-Header-Add":        "header-val-1",
			// 				"X-Header-Add-Append": "append-val-1,header-val-2",
			// 				"X-Header-Set":        "set-overwrites-values",
			// 			},
			// 		},
			// 		AbsentHeaders: []string{"X-Header-Remove"},
			// 	},
			// 	Backend:   "infra-backend-v1",
			// 	Namespace: ns,
			// },
		}
		for i := range testCases {
			// Declare tc here to avoid loop variable
			// reuse issues across parallel tests.
			tc := testCases[i]
			t.Run(tc.GetTestCaseName(i), func(t *testing.T) {
				t.Parallel()
				http.MakeRequestAndExpectEventuallyConsistentResponse(t, s.RoundTripper, s.TimeoutConfig, gwAddr, tc)
			})
		}
	},
}
