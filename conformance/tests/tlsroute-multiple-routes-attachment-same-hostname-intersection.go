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
	"sigs.k8s.io/gateway-api/conformance/utils/tls"
	"sigs.k8s.io/gateway-api/pkg/features"
)

func init() {
	ConformanceTests = append(ConformanceTests, TLSRouteMultipleRoutesAttachmentSameHostnameIntersection)
}

var TLSRouteMultipleRoutesAttachmentSameHostnameIntersection = confsuite.ConformanceTest{
	ShortName:   "TLSRouteMultipleRoutesAttachmentSameHostnameIntersection",
	Description: "TLSRoutes attached to the same listener should resolve conflicts using hostname specificity over creation timestamp",
	Features: []features.FeatureName{
		features.SupportGateway,
		features.SupportTLSRoute,
	},
	Manifests: []string{"tests/tlsroute-multiple-routes-attachment-same-hostname-intersection.yaml"},
	Test: func(t *testing.T, suite *confsuite.ConformanceTestSuite) {
		ns := confsuite.InfrastructureNamespace
		certNN := types.NamespacedName{Name: "tls-checks-certificate", Namespace: ns}

		// This test creates an additional Gateway in the gateway-conformance-infra
		// namespace so we have to wait for it to be ready.
		kubernetes.NamespacesMustBeReady(t, suite.Client, suite.TimeoutConfig, []string{ns})

		serverCertPem, _, err := kubernetes.GetTLSSecret(suite.Client, certNN)
		if err != nil {
			t.Fatalf("unexpected error finding TLS secret: %v", err)
		}
		if len(serverCertPem) == 0 {
			t.Fatal("missing required server certificate pem for the test")
		}

		t.Run("TLSRoute exact hostname match takes precedence over TLSRoute wildcard hostname match despite having newer creation timestamp", func(t *testing.T) {
			gwNN := types.NamespacedName{Name: "gw-tlsroute-exact-hostname-x", Namespace: ns}
			olderRouteNN := types.NamespacedName{Name: "tlsroute-wc-hostname-x-older", Namespace: ns}
			newerRouteNN := types.NamespacedName{Name: "tlsroute-exact-hostname-x-newer", Namespace: ns}

			gwAddr, _ := kubernetes.GatewayAndTLSRoutesMustBeAccepted(t, suite.Client, suite.TimeoutConfig, suite.ControllerName, kubernetes.NewGatewayRef(gwNN), olderRouteNN)

			// CreationTimestamp has second-level precision; sleep ensures the second route is strictly newer than the first.
			time.Sleep(time.Second)

			newerRoute := &gatewayv1.TLSRoute{
				ObjectMeta: metav1.ObjectMeta{
					Name:      newerRouteNN.Name,
					Namespace: newerRouteNN.Namespace,
				},
				Spec: gatewayv1.TLSRouteSpec{
					CommonRouteSpec: gatewayv1.CommonRouteSpec{
						ParentRefs: []gatewayv1.ParentReference{{
							Name: gatewayv1.ObjectName(gwNN.Name),
						}},
					},
					Hostnames: []gatewayv1.Hostname{"abc.example.com"},
					Rules: []gatewayv1.TLSRouteRule{{
						BackendRefs: []gatewayv1.BackendRef{{
							BackendObjectReference: gatewayv1.BackendObjectReference{
								Name: gatewayv1.ObjectName("tls-backend-2"),
								Port: ptr.To(gatewayv1.PortNumber(443)),
							},
						}},
					}},
				},
			}
			suite.Applier.MustApplyObjectsWithCleanup(t, suite.Client, suite.TimeoutConfig, []client.Object{newerRoute}, suite.CleanupTestResources)

			kubernetes.TLSRouteMustHaveResolvedRefsConditionsTrue(t, suite.Client, suite.TimeoutConfig, olderRouteNN, gwNN)
			kubernetes.TLSRouteMustHaveResolvedRefsConditionsTrue(t, suite.Client, suite.TimeoutConfig, newerRouteNN, gwNN)

			tls.MakeTLSRequestAndExpectEventuallyConsistentResponse(t, suite.RoundTripper, suite.TimeoutConfig, gwAddr, serverCertPem, nil, nil, "abc.example.com",
				http.ExpectedResponse{
					Request:   http.Request{Host: "abc.example.com", Path: "/"},
					Backend:   "tls-backend-2",
					Namespace: ns,
				})
		})

		t.Run("TLSRoute more specific wildcard hostname match takes precedence over less specific wildcard hostname match despite having newer creation timestamp", func(t *testing.T) {
			gwNN := types.NamespacedName{Name: "gw-tlsroute-wc-hostname-x", Namespace: ns}
			olderRouteNN := types.NamespacedName{Name: "tlsroute-less-specific-wc-hostname-x-older", Namespace: ns}
			newerRouteNN := types.NamespacedName{Name: "tlsroute-more-specific-wc-hostname-x-newer", Namespace: ns}

			gwAddr, _ := kubernetes.GatewayAndTLSRoutesMustBeAccepted(t, suite.Client, suite.TimeoutConfig, suite.ControllerName, kubernetes.NewGatewayRef(gwNN), olderRouteNN)

			// CreationTimestamp has second-level precision; sleep ensures the second route is strictly newer than the first.
			time.Sleep(time.Second)

			newerRoute := &gatewayv1.TLSRoute{
				ObjectMeta: metav1.ObjectMeta{
					Name:      newerRouteNN.Name,
					Namespace: newerRouteNN.Namespace,
				},
				Spec: gatewayv1.TLSRouteSpec{
					CommonRouteSpec: gatewayv1.CommonRouteSpec{
						ParentRefs: []gatewayv1.ParentReference{{
							Name: gatewayv1.ObjectName(gwNN.Name),
						}},
					},
					Hostnames: []gatewayv1.Hostname{"*.example.com"},
					Rules: []gatewayv1.TLSRouteRule{{
						BackendRefs: []gatewayv1.BackendRef{{
							BackendObjectReference: gatewayv1.BackendObjectReference{
								Name: gatewayv1.ObjectName("tls-backend-2"),
								Port: ptr.To(gatewayv1.PortNumber(443)),
							},
						}},
					}},
				},
			}
			suite.Applier.MustApplyObjectsWithCleanup(t, suite.Client, suite.TimeoutConfig, []client.Object{newerRoute}, suite.CleanupTestResources)

			kubernetes.TLSRouteMustHaveResolvedRefsConditionsTrue(t, suite.Client, suite.TimeoutConfig, olderRouteNN, gwNN)
			kubernetes.TLSRouteMustHaveResolvedRefsConditionsTrue(t, suite.Client, suite.TimeoutConfig, newerRouteNN, gwNN)

			tls.MakeTLSRequestAndExpectEventuallyConsistentResponse(t, suite.RoundTripper, suite.TimeoutConfig, gwAddr, serverCertPem, nil, nil, "abc.example.com",
				http.ExpectedResponse{
					Request:   http.Request{Host: "abc.example.com", Path: "/"},
					Backend:   "tls-backend-2",
					Namespace: ns,
				})
		})
	},
}
