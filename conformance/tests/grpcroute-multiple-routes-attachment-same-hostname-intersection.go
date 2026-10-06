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

	"k8s.io/apimachinery/pkg/types"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"

	pb "sigs.k8s.io/gateway-api-conformance-images/echo-basic/grpcechoserver"
	gatewayv1 "sigs.k8s.io/gateway-api/apis/v1"
	"sigs.k8s.io/gateway-api/conformance/utils/grpc"
	"sigs.k8s.io/gateway-api/conformance/utils/kubernetes"
	confsuite "sigs.k8s.io/gateway-api/conformance/utils/suite"
	"sigs.k8s.io/gateway-api/pkg/features"
)

func init() {
	ConformanceTests = append(ConformanceTests, GRPCRouteMultipleRoutesAttachmentSameHostnameIntersection)
}

var GRPCRouteMultipleRoutesAttachmentSameHostnameIntersection = confsuite.ConformanceTest{
	ShortName:   "GRPCRouteMultipleRoutesAttachmentSameHostnameIntersection",
	Description: "GRPCRoutes attached to the same listener should resolve conflicts using original hostnames over hostname intersections",
	Features: []features.FeatureName{
		features.SupportGateway,
		features.SupportGatewayRouteHostnameIntersectionPrecedence,
		features.SupportGRPCRoute,
	},
	Manifests: []string{"tests/grpcroute-multiple-routes-attachment-same-hostname-intersection.yaml"},
	Test: func(t *testing.T, suite *confsuite.ConformanceTestSuite) {
		ns := confsuite.InfrastructureNamespace

		kubernetes.NamespacesMustBeReady(t, suite.Client, suite.TimeoutConfig, []string{ns})

		t.Run("GRPCRoute exact hostname match takes precedence over GRPCRoute wildcard hostname match despite having newer creation timestamp", func(t *testing.T) {
			gwNN := types.NamespacedName{Name: "gw-grpcroute-exact-hostname-x", Namespace: ns}
			olderRouteNN := types.NamespacedName{Name: "grpcroute-wc-hostname-x-older", Namespace: ns}
			newerRouteNN := types.NamespacedName{Name: "grpcroute-exact-hostname-x-newer", Namespace: ns}

			// This test creates an additional Gateway in the gateway-conformance-infra namespace so we have to wait for it to be ready.
			gwAddr := kubernetes.GatewayAndRoutesMustBeAccepted(t, suite.Client, suite.TimeoutConfig, suite.ControllerName, kubernetes.NewGatewayRef(gwNN), &gatewayv1.GRPCRoute{}, true, olderRouteNN)

			// CreationTimestamp has second-level precision; sleep ensures the second route is strictly newer than the first.
			time.Sleep(time.Second)

			newerRoute := &gatewayv1.GRPCRoute{
				Name:      newerRouteNN.Name,
				Namespace: newerRouteNN.Namespace,
				Spec: gatewayv1.GRPCRouteSpec{
					CommonRouteSpec: gatewayv1.CommonRouteSpec{
						ParentRefs: []gatewayv1.ParentReference{{
							Name: gatewayv1.ObjectName(gwNN.Name),
						}},
					},
					Hostnames: []gatewayv1.Hostname{"abc.example.com"},
					Rules: []gatewayv1.GRPCRouteRule{{
						BackendRefs: []gatewayv1.GRPCBackendRef{{
							Name: gatewayv1.ObjectName("grpc-infra-backend-v2"),
							Port: ptr.To(gatewayv1.PortNumber(8080)),
						}},
					}},
				},
			}
			suite.Applier.MustApplyObjectsWithCleanup(t, suite.Client, suite.TimeoutConfig, []client.Object{newerRoute}, suite.CleanupTestResources)

			kubernetes.GatewayAndRoutesMustBeAccepted(t, suite.Client, suite.TimeoutConfig, suite.ControllerName, kubernetes.NewGatewayRef(gwNN), &gatewayv1.GRPCRoute{}, true, olderRouteNN, newerRouteNN)

			grpc.MakeRequestAndExpectEventuallyConsistentResponse(t, suite.GRPCClient, suite.TimeoutConfig, gwAddr,
				grpc.ExpectedResponse{
					EchoRequest: &pb.EchoRequest{},
					RequestMetadata: &grpc.RequestMetadata{
						Authority: "abc.example.com",
					},
					Backend:   "grpc-infra-backend-v2",
					Namespace: ns,
				})
		})

		t.Run("GRPCRoute more specific wildcard hostname match takes precedence over less specific wildcard hostname match despite having newer creation timestamp", func(t *testing.T) {
			gwNN := types.NamespacedName{Name: "gw-grpcroute-wc-hostname-x", Namespace: ns}
			olderRouteNN := types.NamespacedName{Name: "grpcroute-less-specific-wc-hostname-x-older", Namespace: ns}
			newerRouteNN := types.NamespacedName{Name: "grpcroute-more-specific-wc-hostname-x-newer", Namespace: ns}

			// This test creates an additional Gateway in the gateway-conformance-infra namespace so we have to wait for it to be ready.
			gwAddr := kubernetes.GatewayAndRoutesMustBeAccepted(t, suite.Client, suite.TimeoutConfig, suite.ControllerName, kubernetes.NewGatewayRef(gwNN), &gatewayv1.GRPCRoute{}, true, olderRouteNN)

			// CreationTimestamp has second-level precision; sleep ensures the second route is strictly newer than the first.
			time.Sleep(time.Second)

			newerRoute := &gatewayv1.GRPCRoute{
				Name:      newerRouteNN.Name,
				Namespace: newerRouteNN.Namespace,
				Spec: gatewayv1.GRPCRouteSpec{
					CommonRouteSpec: gatewayv1.CommonRouteSpec{
						ParentRefs: []gatewayv1.ParentReference{{
							Name: gatewayv1.ObjectName(gwNN.Name),
						}},
					},
					Hostnames: []gatewayv1.Hostname{"*.example.com"},
					Rules: []gatewayv1.GRPCRouteRule{{
						BackendRefs: []gatewayv1.GRPCBackendRef{{
							Name: gatewayv1.ObjectName("grpc-infra-backend-v2"),
							Port: ptr.To(gatewayv1.PortNumber(8080)),
						}},
					}},
				},
			}
			suite.Applier.MustApplyObjectsWithCleanup(t, suite.Client, suite.TimeoutConfig, []client.Object{newerRoute}, suite.CleanupTestResources)

			kubernetes.GatewayAndRoutesMustBeAccepted(t, suite.Client, suite.TimeoutConfig, suite.ControllerName, kubernetes.NewGatewayRef(gwNN), &gatewayv1.GRPCRoute{}, true, olderRouteNN, newerRouteNN)

			grpc.MakeRequestAndExpectEventuallyConsistentResponse(t, suite.GRPCClient, suite.TimeoutConfig, gwAddr,
				grpc.ExpectedResponse{
					EchoRequest: &pb.EchoRequest{},
					RequestMetadata: &grpc.RequestMetadata{
						Authority: "abc.example.com",
					},
					Backend:   "grpc-infra-backend-v2",
					Namespace: ns,
				})
		})
	},
}
