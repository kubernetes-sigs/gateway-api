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
	ConformanceTests = append(ConformanceTests, HTTPRouteExternalAuthBackendRefNotFound)
}

var HTTPRouteExternalAuthBackendRefNotFound = confsuite.ConformanceTest{
	ShortName:   "HTTPRouteExternalAuthBackendRefNotFound",
	Description: "An HTTPRoute with an ExternalAuth filter whose backendRef cannot be resolved must have ResolvedRefs=False/BackendNotFound and must fail closed",
	Manifests:   []string{"tests/httproute-external-auth-backendref-not-found.yaml"},
	Features: []features.FeatureName{
		features.SupportGateway,
		features.SupportHTTPRoute,
		features.SupportHTTPRouteExternalAuth,
		features.SupportHTTPRouteExternalAuthHTTP,
	},
	Test: func(t *testing.T, suite *confsuite.ConformanceTestSuite) {
		ns := confsuite.InfrastructureNamespace
		gwNN := types.NamespacedName{Name: "same-namespace", Namespace: ns}
		unknownNN := types.NamespacedName{Name: "external-auth-unknown-backendref", Namespace: ns}
		invalidPortNN := types.NamespacedName{Name: "external-auth-backendref-invalid-port", Namespace: ns}

		gwAddr := kubernetes.GatewayAndHTTPRoutesMustBeAccepted(t, suite.Client, suite.TimeoutConfig, suite.ControllerName, kubernetes.NewGatewayRef(gwNN), unknownNN, invalidPortNN)

		resolvedRefsFalse := metav1.Condition{
			Type:   string(v1.RouteConditionResolvedRefs),
			Status: metav1.ConditionFalse,
			Reason: string(v1.RouteReasonBackendNotFound),
		}

		t.Run("route with nonexistent ExternalAuth Service has ResolvedRefs=False/BackendNotFound", func(t *testing.T) {
			kubernetes.HTTPRouteMustHaveCondition(t, suite.Client, suite.TimeoutConfig, unknownNN, gwNN, resolvedRefsFalse)
		})

		t.Run("route with ExternalAuth Service on a nonexistent port has ResolvedRefs=False/BackendNotFound", func(t *testing.T) {
			kubernetes.HTTPRouteMustHaveCondition(t, suite.Client, suite.TimeoutConfig, invalidPortNN, gwNN, resolvedRefsFalse)
		})

		t.Run("requests to a route with a nonexistent ExternalAuth Service are denied", func(t *testing.T) {
			http.MakeRequestAndExpectEventuallyConsistentResponse(t, suite.RoundTripper, suite.TimeoutConfig, gwAddr, http.ExpectedResponse{
				Request:  http.Request{Path: "/external-auth/unknown-backendref"},
				Response: http.Response{StatusCodes: []int{403, 500, 502, 503}},
			})
		})

		t.Run("requests to a route with an ExternalAuth Service on a nonexistent port are denied", func(t *testing.T) {
			http.MakeRequestAndExpectEventuallyConsistentResponse(t, suite.RoundTripper, suite.TimeoutConfig, gwAddr, http.ExpectedResponse{
				Request:  http.Request{Path: "/external-auth/invalid-port"},
				Response: http.Response{StatusCodes: []int{403, 500, 502, 503}},
			})
		})
	},
}
