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

	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	v1 "sigs.k8s.io/gateway-api/apis/v1"
	"sigs.k8s.io/gateway-api/conformance/utils/http"
	"sigs.k8s.io/gateway-api/conformance/utils/kubernetes"
	confsuite "sigs.k8s.io/gateway-api/conformance/utils/suite"
	"sigs.k8s.io/gateway-api/pkg/features"
)

func init() {
	ConformanceTests = append(ConformanceTests, HTTPRouteNoBackendRefs)
}

var HTTPRouteNoBackendRefs = confsuite.ConformanceTest{
	ShortName:   "HTTPRouteNoBackendRefs",
	Description: "HTTPRoute rules with omitted or empty backendRefs explicitly respond with 500",
	Features: []features.FeatureName{
		features.SupportGateway,
		features.SupportHTTPRoute,
	},
	Manifests: []string{"tests/httproute-omitted-backendrefs.yaml"},
	Test: func(t *testing.T, suite *confsuite.ConformanceTestSuite) {
		ns := confsuite.InfrastructureNamespace
		routeNN := types.NamespacedName{Name: "omitted-backendrefs", Namespace: ns}
		gwNN := types.NamespacedName{Name: "same-namespace", Namespace: ns}
		gwAddr, err := kubernetes.WaitForGatewayAddress(t, suite.Client, suite.TimeoutConfig, kubernetes.NewGatewayRef(gwNN))
		require.NoErrorf(t, err, "timed out waiting for Gateway address to be assigned")
		kubernetes.HTTPRouteMustHaveResolvedRefsConditionsTrue(t, suite.Client, suite.TimeoutConfig, routeNN, gwNN)

		parentRef := parentRefTo(gwNN)
		// Set the namespace to nil since it is in the same namespace as the parent
		parentRef.Namespace = nil
		parents := []v1.RouteParentStatus{{
			ParentRef:      parentRef,
			ControllerName: v1.GatewayController(suite.ControllerName),
			Conditions: []metav1.Condition{
				{
					Type:   string(v1.RouteConditionAccepted),
					Status: metav1.ConditionTrue,
					Reason: "", // any reason
				},
			},
		}}
		kubernetes.HTTPRouteMustHaveParents(t, suite.Client, suite.TimeoutConfig, routeNN, parents, true)

		testCases := []http.ExpectedResponse{
			{
				Request:   http.Request{Path: "/forward"},
				Backend:   confsuite.InfraBackendServiceNameV1,
				Namespace: ns,
			}, {
				Request:  http.Request{Path: "/omitted-no-forward"},
				Response: http.Response{StatusCode: 500},
			}, {
				Request:  http.Request{Path: "/empty-no-forward"},
				Response: http.Response{StatusCode: 500},
			},
		}

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
