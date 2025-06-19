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
	"strings"
	"testing"

	"sigs.k8s.io/gateway-api/conformance/utils/http"
	"sigs.k8s.io/gateway-api/conformance/utils/suite"
	"sigs.k8s.io/gateway-api/pkg/features"
)

func init() {
	ConformanceTests = append(ConformanceTests,
		HTTPRouteRewritePathBackendWeights,
	)
}

var HTTPRouteRewritePathBackendWeights = suite.ConformanceTest{
	ShortName:   "HTTPRouteRewritePathBackendWeights",
	Description: "An HTTPRoute with backend URL filter filter sends traffic to the correct backends.",
	Features: []features.FeatureName{
		features.SupportGateway,
		features.SupportHTTPRoute,
	},
	Manifests: []string{"tests/httproute-rewrite-path-backend-weights.yaml"},
	Test: func(t *testing.T, s *suite.ConformanceTestSuite) {
		ns := "gateway-conformance-infra"
		gwAddr := suite.DefaultConformanceTestBoilerplate(t, s, ns, "rewrite-path-backend-weights", "same-namespace")

		roundTripper := s.RoundTripper

		expected := http.ExpectedResponse{
			Request:   http.Request{Path: "/prefix/test"},
			Response:  http.Response{StatusCode: 200},
			Namespace: "gateway-conformance-infra",
		}

		req := http.MakeRequest(t, &expected, gwAddr, "HTTP", "http")

		// Assert request succeeds before checking traffic
		http.MakeRequestAndExpectEventuallyConsistentResponse(t, s.RoundTripper, s.TimeoutConfig, gwAddr, expected)

		for range 100 {
			cReq, _, err := roundTripper.CaptureRoundTrip(req)
			if err != nil {
				t.Fatalf("failed to roundtrip request: %v", err)
			}

			backend, found := strings.CutSuffix(cReq.Path, "/test")
			if !found {
				t.Fatalf("expected to have sufix \"/test\": %v", cReq.Path)
			}
			backend, found = strings.CutPrefix(backend, "/")
			if !found {
				t.Fatalf("expected to have prefix \"/\": %v", backend)
			}

			if !strings.Contains(cReq.Pod, backend) {
				t.Fatalf(
					"expected %q to be subset of %q and sent the request to the correct pod",
					cReq.Pod,
					backend,
				)
			}
		}

		expected = http.ExpectedResponse{
			Request:   http.Request{Path: "/"},
			Response:  http.Response{StatusCode: 200},
			Namespace: "gateway-conformance-infra",
		}

		req = http.MakeRequest(t, &expected, gwAddr, "HTTP", "http")

		// Assert request succeeds before checking traffic
		http.MakeRequestAndExpectEventuallyConsistentResponse(t, s.RoundTripper, s.TimeoutConfig, gwAddr, expected)

		for range 100 {
			cReq, _, err := roundTripper.CaptureRoundTrip(req)
			if err != nil {
				t.Fatalf("failed to roundtrip request: %v", err)
			}

			if !strings.HasPrefix(cReq.Path, "/infra-backend") {
				t.Fatalf("expected to have prefix \"/infra-backend\": %v", cReq.Path)
			}

			backend, _ := strings.CutPrefix(cReq.Path, "/")

			if !strings.Contains(cReq.Pod, backend) {
				t.Fatalf(
					"expected %q to be subset of %q and sent the request to the correct pod",
					backend,
					cReq.Pod,
				)
			}

		}
	},
}
