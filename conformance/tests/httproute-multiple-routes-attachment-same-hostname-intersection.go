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
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"

	gatewayv1 "sigs.k8s.io/gateway-api/apis/v1"
	"sigs.k8s.io/gateway-api/conformance/utils/http"
	"sigs.k8s.io/gateway-api/conformance/utils/kubernetes"
	confsuite "sigs.k8s.io/gateway-api/conformance/utils/suite"
	"sigs.k8s.io/gateway-api/pkg/features"
)

func init() {
	ConformanceTests = append(ConformanceTests, HTTPRouteMultipleRoutesAttachmentSameHostnameIntersection)
}

var HTTPRouteMultipleRoutesAttachmentSameHostnameIntersection = confsuite.ConformanceTest{
	ShortName:   "HTTPRouteMultipleRoutesAttachmentSameHostnameIntersection",
	Description: "HTTPRoutes attached to the same listener should resolve conflicts using original hostnames over hostname intersections",
	Features: []features.FeatureName{
		features.SupportGateway,
		features.SupportHTTPRoute,
	},
	Manifests: []string{"tests/httproute-multiple-routes-attachment-same-hostname-intersection.yaml"},
	Test: func(t *testing.T, suite *confsuite.ConformanceTestSuite) {
		ns := confsuite.InfrastructureNamespace

		kubernetes.NamespacesMustBeReady(t, suite.Client, suite.TimeoutConfig, []string{ns})

		t.Run("HTTPRoute exact hostname match takes precedence over HTTPRoute wildcard hostname match despite having newer creation timestamp", func(t *testing.T) {
			gwNN := types.NamespacedName{Name: "gw-httproute-exact-hostname-x", Namespace: ns}
			olderRouteNN := types.NamespacedName{Name: "httproute-wc-hostname-x-older", Namespace: ns}
			newerRouteNN := types.NamespacedName{Name: "httproute-exact-hostname-x-newer", Namespace: ns}

			// This test creates an additional Gateway in the gateway-conformance-infra namespace so we have to wait for it to be ready.
			gwAddr := kubernetes.GatewayAndHTTPRoutesMustBeAccepted(t, suite.Client, suite.TimeoutConfig, suite.ControllerName, kubernetes.NewGatewayRef(gwNN), olderRouteNN)

			// CreationTimestamp has second-level precision; sleep ensures the second route is strictly newer than the first.
			time.Sleep(time.Second)

			newerRoute := &gatewayv1.HTTPRoute{
				ObjectMeta: metav1.ObjectMeta{
					Name:      newerRouteNN.Name,
					Namespace: newerRouteNN.Namespace,
				},
				Spec: gatewayv1.HTTPRouteSpec{
					CommonRouteSpec: gatewayv1.CommonRouteSpec{
						ParentRefs: []gatewayv1.ParentReference{{
							Name: gatewayv1.ObjectName(gwNN.Name),
						}},
					},
					Hostnames: []gatewayv1.Hostname{"abc.example.com"},
					Rules: []gatewayv1.HTTPRouteRule{{
						BackendRefs: []gatewayv1.HTTPBackendRef{{
							BackendRef: gatewayv1.BackendRef{
								BackendObjectReference: gatewayv1.BackendObjectReference{
									Name: gatewayv1.ObjectName(confsuite.InfraBackendServiceNameV2),
									Port: ptr.To(gatewayv1.PortNumber(8080)),
								},
							},
						}},
					}},
				},
			}
			suite.Applier.MustApplyObjectsWithCleanup(t, suite.Client, suite.TimeoutConfig, []client.Object{newerRoute}, suite.CleanupTestResources)

			kubernetes.HTTPRouteMustHaveResolvedRefsConditionsTrue(t, suite.Client, suite.TimeoutConfig, olderRouteNN, gwNN)
			kubernetes.HTTPRouteMustHaveResolvedRefsConditionsTrue(t, suite.Client, suite.TimeoutConfig, newerRouteNN, gwNN)

			http.MakeRequestAndExpectEventuallyConsistentResponse(t, suite.RoundTripper, suite.TimeoutConfig, gwAddr,
				http.ExpectedResponse{
					Request:   http.Request{Host: "abc.example.com", Path: "/"},
					Backend:   confsuite.InfraBackendServiceNameV2,
					Namespace: ns,
				})
		})

		t.Run("HTTPRoute more specific wildcard hostname match takes precedence over less specific wildcard hostname match despite having newer creation timestamp", func(t *testing.T) {
			gwNN := types.NamespacedName{Name: "gw-httproute-wc-hostname-x", Namespace: ns}
			olderRouteNN := types.NamespacedName{Name: "httproute-less-specific-wc-hostname-x-older", Namespace: ns}
			newerRouteNN := types.NamespacedName{Name: "httproute-more-specific-wc-hostname-x-newer", Namespace: ns}

			// This test creates an additional Gateway in the gateway-conformance-infra namespace so we have to wait for it to be ready.
			gwAddr := kubernetes.GatewayAndHTTPRoutesMustBeAccepted(t, suite.Client, suite.TimeoutConfig, suite.ControllerName, kubernetes.NewGatewayRef(gwNN), olderRouteNN)

			// CreationTimestamp has second-level precision; sleep ensures the second route is strictly newer than the first.
			time.Sleep(time.Second)

			newerRoute := &gatewayv1.HTTPRoute{
				ObjectMeta: metav1.ObjectMeta{
					Name:      newerRouteNN.Name,
					Namespace: newerRouteNN.Namespace,
				},
				Spec: gatewayv1.HTTPRouteSpec{
					CommonRouteSpec: gatewayv1.CommonRouteSpec{
						ParentRefs: []gatewayv1.ParentReference{{
							Name: gatewayv1.ObjectName(gwNN.Name),
						}},
					},
					Hostnames: []gatewayv1.Hostname{"*.example.com"},
					Rules: []gatewayv1.HTTPRouteRule{{
						BackendRefs: []gatewayv1.HTTPBackendRef{{
							BackendRef: gatewayv1.BackendRef{
								BackendObjectReference: gatewayv1.BackendObjectReference{
									Name: gatewayv1.ObjectName(confsuite.InfraBackendServiceNameV2),
									Port: ptr.To(gatewayv1.PortNumber(8080)),
								},
							},
						}},
					}},
				},
			}
			suite.Applier.MustApplyObjectsWithCleanup(t, suite.Client, suite.TimeoutConfig, []client.Object{newerRoute}, suite.CleanupTestResources)

			kubernetes.HTTPRouteMustHaveResolvedRefsConditionsTrue(t, suite.Client, suite.TimeoutConfig, olderRouteNN, gwNN)
			kubernetes.HTTPRouteMustHaveResolvedRefsConditionsTrue(t, suite.Client, suite.TimeoutConfig, newerRouteNN, gwNN)

			http.MakeRequestAndExpectEventuallyConsistentResponse(t, suite.RoundTripper, suite.TimeoutConfig, gwAddr,
				http.ExpectedResponse{
					Request:   http.Request{Host: "abc.example.com", Path: "/"},
					Backend:   confsuite.InfraBackendServiceNameV2,
					Namespace: ns,
				})
		})
	},
}
