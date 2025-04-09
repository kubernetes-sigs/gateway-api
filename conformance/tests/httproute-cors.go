package tests

import (
	"testing"

	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/gateway-api/conformance/utils/http"
	"sigs.k8s.io/gateway-api/conformance/utils/kubernetes"
	"sigs.k8s.io/gateway-api/conformance/utils/suite"
	"sigs.k8s.io/gateway-api/pkg/features"
)

func init() {
	ConformanceTests = append(ConformanceTests, HTTPRouteCORS)
}

var HTTPRouteCORS = suite.ConformanceTest{
	ShortName:   "HTTPRouteCORS",
	Description: "An HTTPRoute with cors filter",
	Features: []features.FeatureName{
		features.SupportGateway,
		features.SupportHTTPRoute,
		features.SupportHTTPRouteCORS,
	},
	Manifests: []string{"tests/httproute-cors.yaml"},
	Test: func(t *testing.T, suite *suite.ConformanceTestSuite) {
		ns := "gateway-conformance-infra"
		routeNN := types.NamespacedName{Name: "cors", Namespace: ns}
		gwNN := types.NamespacedName{Name: "same-namespace", Namespace: ns}
		gwAddr := kubernetes.GatewayAndHTTPRoutesMustBeAccepted(t, suite.Client, suite.TimeoutConfig, suite.ControllerName, kubernetes.NewGatewayRef(gwNN), routeNN)
		kubernetes.HTTPRouteMustHaveResolvedRefsConditionsTrue(t, suite.Client, suite.TimeoutConfig, routeNN, gwNN)

		testCases := []http.ExpectedResponse{{
			TestCaseName: "GET request",
			Request: http.Request{
				Path: "/one",
				Headers: map[string]string{
					"Origin": "http://example.com",
				},
			},
			Response: http.Response{
				StatusCode: 200,
				Headers: map[string]string{
					"Access-Control-Allow-Origin": "http://example.com",
				},
			},
			Backend:   "infra-backend-v1",
			Namespace: ns,
		}, {
			TestCaseName: "CORS preflight request with request method and headers",
			Request: http.Request{
				Method: "OPTIONS",
				Path:   "/one",
				Headers: map[string]string{
					"Origin":                         "http://one.example.com",
					"Access-Control-Request-Method":  "POST",
					"Access-Control-Request-Headers": "Accept,Content-Type",
				},
			},
			ExpectedRequest: &http.ExpectedRequest{
				Request: http.Request{
					Method: "OPTIONS",
				},
			},
			Response: http.Response{
				StatusCode: 200,
				Headers: map[string]string{
					"Access-Control-Allow-Origin":  "http://one.example.com",
					"Access-Control-Allow-Methods": "GET,HEAD,POST",
					"Access-Control-Allow-Headers": "Accept,Accept-Language,Content-Language,Content-Type,Range",
					"Access-Control-Max-Age":       "5",
				},
				AbsentHeaders: []string{
					"Access-Control-Allow-Credentials",
					"Access-Control-Expose-Headers",
				},
			},
		}, {
			TestCaseName: "CORS preflight request with request method",
			Request: http.Request{
				Method: "OPTIONS",
				Path:   "/two",
				Headers: map[string]string{
					"Origin":                        "http://two.example.com",
					"Access-Control-Request-Method": "DELETE",
				},
			},
			ExpectedRequest: &http.ExpectedRequest{
				Request: http.Request{
					Method: "OPTIONS",
				},
			},
			Response: http.Response{
				StatusCode: 200,
				Headers: map[string]string{
					"Access-Control-Allow-Origin":      "http://two.example.com",
					"Access-Control-Allow-Methods":     "GET,HEAD,POST,DELETE",
					"Access-Control-Max-Age":           "1728000",
					"Access-Control-Allow-Credentials": "true",
					"Access-Control-Expose-Headers":    "Content-Security-Policy",
				},
				AbsentHeaders: []string{
					"Access-Control-Allow-Headers",
				},
			},
		}}

		for i := range testCases {
			// Declare tc here to avoid loop variable
			// reuse issues across parallel tests.
			tc := testCases[i]
			t.Run(tc.GetTestCaseName(i), func(t *testing.T) {
				t.Parallel()
				http.MakeRequestAndExpectEventuallyConsistentResponse(t, suite.RoundTripper, suite.TimeoutConfig, gwAddr, tc)
			})
		}
	},
}
