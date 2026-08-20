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
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	v1 "sigs.k8s.io/gateway-api/apis/v1"
	"sigs.k8s.io/gateway-api/conformance/utils/http"
	"sigs.k8s.io/gateway-api/conformance/utils/kubernetes"
	confsuite "sigs.k8s.io/gateway-api/conformance/utils/suite"
	"sigs.k8s.io/gateway-api/pkg/features"
)

func init() {
	ConformanceTests = append(ConformanceTests, HTTPRouteExternalAuthPartiallyInvalid)
}

var HTTPRouteExternalAuthPartiallyInvalid = confsuite.ConformanceTest{
	ShortName:   "HTTPRouteExternalAuthPartiallyInvalid",
	Description: "An HTTPRoute with two rules where one rule has a valid ExternalAuth backendRef and another rule has an unresolvable ExternalAuth backendRef should have PartiallyInvalid=True and ResolvedRefs=False",
	Manifests:   []string{"tests/httproute-external-auth-partially-invalid.yaml"},
	Features: []features.FeatureName{
		features.SupportGateway,
		features.SupportHTTPRoute,
		features.SupportHTTPRouteExternalAuth,
		features.SupportHTTPRouteExternalAuthHTTP,
	},
	Test: func(t *testing.T, suite *confsuite.ConformanceTestSuite) {
		ns := confsuite.InfrastructureNamespace
		routeNN := types.NamespacedName{Name: "external-auth-partially-invalid", Namespace: ns}
		gwNN := types.NamespacedName{Name: "same-namespace", Namespace: ns}

		gwAddr := kubernetes.GatewayAndHTTPRoutesMustBeAccepted(t, suite.Client, suite.TimeoutConfig, suite.ControllerName, kubernetes.NewGatewayRef(gwNN), routeNN)

		t.Run("HTTPRoute with one invalid ExternalAuth rule has PartiallyInvalid=True", func(t *testing.T) {
			// The spec lists "UnsupportedValue" as the canonical reason but permits other
			// reasons. Implementations resolving ExternalAuth backendRefs typically set
			// "BackendNotFound". We accept any reason by passing an empty string.
			kubernetes.HTTPRouteMustHaveCondition(t, suite.Client, suite.TimeoutConfig, routeNN, gwNN, metav1.Condition{
				Type:   string(v1.RouteConditionPartiallyInvalid),
				Status: metav1.ConditionTrue,
				Reason: "",
			})
		})

		t.Run("HTTPRoute with one invalid ExternalAuth rule has ResolvedRefs=False/BackendNotFound", func(t *testing.T) {
			kubernetes.HTTPRouteMustHaveCondition(t, suite.Client, suite.TimeoutConfig, routeNN, gwNN, metav1.Condition{
				Type:   string(v1.RouteConditionResolvedRefs),
				Status: metav1.ConditionFalse,
				Reason: string(v1.RouteReasonBackendNotFound),
			})
		})

		t.Run("requests to the invalid ExternalAuth rule return 403, or 5xx", func(t *testing.T) {
			http.MakeRequestAndExpectEventuallyConsistentResponse(t, suite.RoundTripper, suite.TimeoutConfig, gwAddr, http.ExpectedResponse{
				Request:  http.Request{Path: "/external-auth/invalid"},
				Response: http.Response{StatusCodes: []int{403, 500, 502, 503}},
			})
		})
	},
}
