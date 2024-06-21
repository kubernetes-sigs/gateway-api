/*
Copyright 2024 The Kubernetes Authors.

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

package resourcediscovery

import (
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	apimachinerytypes "k8s.io/apimachinery/pkg/types"
	testingclock "k8s.io/utils/clock/testing"

	gatewayv1 "sigs.k8s.io/gateway-api/apis/v1"
	gatewayv1beta1 "sigs.k8s.io/gateway-api/apis/v1beta1"
	"sigs.k8s.io/gateway-api/gwctl/pkg/common"
	"sigs.k8s.io/gateway-api/gwctl/pkg/utils"
)

func TestDiscoverResourcesForGatewayClasses(t *testing.T) {
	testcases := []struct {
		name    string
		objects []runtime.Object
		filter  Filter

		wantGatewayClasses []string
		wantGateways       []apimachinerytypes.NamespacedName
	}{
		{
			name:   "normal",
			filter: Filter{Labels: labels.Everything()},
			objects: []runtime.Object{
				common.NamespaceForTest("default"),
				common.NamespaceForTest("baz"),
				&gatewayv1.GatewayClass{
					ObjectMeta: metav1.ObjectMeta{
						Name: "foo-gatewayclass",
					},
				},
				&gatewayv1.Gateway{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "bar-gateway",
						Namespace: "default",
					},
					Spec: gatewayv1.GatewaySpec{
						GatewayClassName: "bar-gatewayclass",
					},
				},
				&gatewayv1.Gateway{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "baz-gateway",
						Namespace: "baz",
					},
					Spec: gatewayv1.GatewaySpec{
						GatewayClassName: "foo-gatewayclass",
					},
				},
			},
			wantGatewayClasses: []string{"foo-gatewayclass"},
			wantGateways: []apimachinerytypes.NamespacedName{
				{Namespace: "baz", Name: "baz-gateway"},
			},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			k8sClients := common.MustClientsForTest(t, tc.objects...)
			policyManager := utils.MustPolicyManagerForTest(t, k8sClients)
			discoverer := Discoverer{
				K8sClients:    k8sClients,
				PolicyManager: policyManager,
			}

			resourceModel, err := discoverer.DiscoverResourcesForGatewayClass(tc.filter)
			if err != nil {
				t.Fatalf("Failed to construct resourceModel: %v", resourceModel)
			}

			gotGatewayClasses := gatewayClassNamesFromResourceModel(resourceModel)
			gotGateways := namespacedGatewaysFromResourceModel(resourceModel)

			if tc.wantGatewayClasses != nil {
				if diff := cmp.Diff(tc.wantGatewayClasses, gotGatewayClasses, cmpopts.EquateEmpty()); diff != "" {
					t.Errorf("Unexpected diff in GatewayClasses; got=%v, want=%v;\ndiff (-want +got)=\n%v", gotGatewayClasses, tc.wantGatewayClasses, diff)
				}
			}
			if tc.wantGateways != nil {
				if diff := cmp.Diff(tc.wantGateways, gotGateways, cmpopts.EquateEmpty()); diff != "" {
					t.Errorf("Unexpected diff in Gateways; got=%v, want=%v;\ndiff (-want +got)=\n%v", gotGateways, tc.wantGateways, diff)
				}
			}
		})
	}
}

func TestDiscoverResourcesForGateway(t *testing.T) {
	testcases := []struct {
		name    string
		objects []runtime.Object
		filter  Filter

		wantGateways []apimachinerytypes.NamespacedName
		// wantGatewayErrors maps a Gateway to the list of errors that Gateway has.
		wantGatewayErrors map[apimachinerytypes.NamespacedName][]error
	}{
		{
			name:   "normal",
			filter: Filter{Labels: labels.Everything()},
			objects: []runtime.Object{
				common.NamespaceForTest("default"),
				&gatewayv1.GatewayClass{
					ObjectMeta: metav1.ObjectMeta{
						Name: "foo-gatewayclass",
					},
				},
				&gatewayv1.Gateway{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "foo-gateway",
						Namespace: "default",
					},
					Spec: gatewayv1.GatewaySpec{
						GatewayClassName: "foo-gatewayclass",
					},
				},
			},
			wantGateways: []apimachinerytypes.NamespacedName{
				{Namespace: "default", Name: "foo-gateway"},
			},
			wantGatewayErrors: map[apimachinerytypes.NamespacedName][]error{
				{Namespace: "default", Name: "foo-gateway"}: nil, // Want no errors.
			},
		},
		{
			name:   "gateway should have error if it references a non-existent gatewayclass",
			filter: Filter{Labels: labels.Everything()},
			objects: []runtime.Object{
				common.NamespaceForTest("default"),
				&gatewayv1.Gateway{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "foo-gateway",
						Namespace: "default",
					},
					Spec: gatewayv1.GatewaySpec{
						GatewayClassName: "foo-gatewayclass", // GatewayClass does not exist.
					},
				},
			},
			wantGateways: []apimachinerytypes.NamespacedName{
				{Namespace: "default", Name: "foo-gateway"},
			},
			wantGatewayErrors: map[apimachinerytypes.NamespacedName][]error{
				{Namespace: "default", Name: "foo-gateway"}: {
					ReferenceToNonExistentResourceError{ReferenceFromTo: ReferenceFromTo{
						ReferringObject: common.ObjRef{Kind: "Gateway", Name: "foo-gateway", Namespace: "default"},
						ReferredObject:  common.ObjRef{Kind: "GatewayClass", Name: "foo-gatewayclass"},
					}},
				},
			},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			k8sClients := common.MustClientsForTest(t, tc.objects...)
			policyManager := utils.MustPolicyManagerForTest(t, k8sClients)
			discoverer := Discoverer{
				K8sClients:    k8sClients,
				PolicyManager: policyManager,
			}

			resourceModel, err := discoverer.DiscoverResourcesForGateway(tc.filter)
			if err != nil {
				t.Fatalf("Failed to construct resourceModel: %v", resourceModel)
			}

			gotGateways := namespacedGatewaysFromResourceModel(resourceModel)
			gotGatewayErrors := gatewayErrorsFromResourceModel(resourceModel)

			if tc.wantGateways != nil {
				if diff := cmp.Diff(tc.wantGateways, gotGateways, cmpopts.EquateEmpty()); diff != "" {
					t.Errorf("Unexpected diff in Gateways; got=%v, want=%v;\ndiff (-want +got)=\n%v", gotGateways, tc.wantGateways, diff)
				}
			}
			if tc.wantGatewayErrors != nil {
				if diff := cmp.Diff(tc.wantGatewayErrors, gotGatewayErrors, cmpopts.EquateEmpty()); diff != "" {
					t.Errorf("Unexpected diff in Gateway errors; got=%v, want=%v;\ndiff (-want +got)=\n%v", gotGatewayErrors, tc.wantGatewayErrors, diff)
				}
			}
		})
	}
}

func TestDiscoverResourcesForHTTPRoute(t *testing.T) {
	testcases := []struct {
		name    string
		objects []runtime.Object
		filter  Filter

		wantHTTPRoutes []apimachinerytypes.NamespacedName
		wantBackends   []apimachinerytypes.NamespacedName
	}{
		{
			name:   "normal",
			filter: Filter{Labels: labels.Everything()},
			objects: []runtime.Object{
				common.NamespaceForTest("default"),
				&gatewayv1.HTTPRoute{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "foo-httproute",
						Namespace: "default",
					},
					Spec: gatewayv1.HTTPRouteSpec{
						CommonRouteSpec: gatewayv1.CommonRouteSpec{},
						Rules: []gatewayv1.HTTPRouteRule{
							{
								BackendRefs: []gatewayv1.HTTPBackendRef{
									{
										BackendRef: gatewayv1.BackendRef{
											BackendObjectReference: gatewayv1.BackendObjectReference{
												Kind: common.PtrTo(gatewayv1.Kind("Service")),
												Name: "foo-svc",
												Port: common.PtrTo(gatewayv1.PortNumber(80)),
											},
										},
									},
								},
							},
						},
					},
				},
				&corev1.Service{
					TypeMeta: metav1.TypeMeta{
						Kind:       "Service",
						APIVersion: "v1",
					},
					ObjectMeta: metav1.ObjectMeta{
						Name:      "foo-svc",
						Namespace: "default",
					},
				},
			},
			wantHTTPRoutes: []apimachinerytypes.NamespacedName{
				{Namespace: "default", Name: "foo-httproute"},
			},
			wantBackends: []apimachinerytypes.NamespacedName{
				{Namespace: "default", Name: "foo-svc"},
			},
		},
		{
			name:   "backendref from different namespace should require referencegrant",
			filter: Filter{Labels: labels.Everything()},
			objects: []runtime.Object{
				common.NamespaceForTest("default"),
				common.NamespaceForTest("bar"),
				&gatewayv1.HTTPRoute{
					TypeMeta: metav1.TypeMeta{
						APIVersion: gatewayv1.GroupVersion.String(),
						Kind:       "HTTPRoute",
					},
					ObjectMeta: metav1.ObjectMeta{
						Name:      "foo-httproute",
						Namespace: "default",
					},
					Spec: gatewayv1.HTTPRouteSpec{
						CommonRouteSpec: gatewayv1.CommonRouteSpec{},
						Rules: []gatewayv1.HTTPRouteRule{
							{
								BackendRefs: []gatewayv1.HTTPBackendRef{
									{
										BackendRef: gatewayv1.BackendRef{
											BackendObjectReference: gatewayv1.BackendObjectReference{
												Kind:      common.PtrTo(gatewayv1.Kind("Service")),
												Name:      "bar-svc",
												Namespace: common.PtrTo(gatewayv1.Namespace("bar")), // Different namespace than HTTPRoute.
												Port:      common.PtrTo(gatewayv1.PortNumber(80)),
											},
										},
									},
								},
							},
						},
					},
				},
				&corev1.Service{
					TypeMeta: metav1.TypeMeta{
						Kind:       "Service",
						APIVersion: "v1",
					},
					ObjectMeta: metav1.ObjectMeta{
						Name:      "bar-svc",
						Namespace: "bar",
					},
				},
			},
			wantHTTPRoutes: []apimachinerytypes.NamespacedName{
				{Namespace: "default", Name: "foo-httproute"},
			},
			wantBackends: []apimachinerytypes.NamespacedName{},
		},
		{
			name:   "backendref from different namespace should get allowed with referencegrant",
			filter: Filter{Labels: labels.Everything()},
			objects: []runtime.Object{
				common.NamespaceForTest("default"),
				common.NamespaceForTest("bar"),
				&gatewayv1.HTTPRoute{
					TypeMeta: metav1.TypeMeta{
						APIVersion: gatewayv1.GroupVersion.String(),
						Kind:       "HTTPRoute",
					},
					ObjectMeta: metav1.ObjectMeta{
						Name:      "foo-httproute",
						Namespace: "default",
					},
					Spec: gatewayv1.HTTPRouteSpec{
						CommonRouteSpec: gatewayv1.CommonRouteSpec{},
						Rules: []gatewayv1.HTTPRouteRule{
							{
								BackendRefs: []gatewayv1.HTTPBackendRef{
									{
										BackendRef: gatewayv1.BackendRef{
											BackendObjectReference: gatewayv1.BackendObjectReference{
												Kind:      common.PtrTo(gatewayv1.Kind("Service")),
												Name:      "bar-svc",
												Namespace: common.PtrTo(gatewayv1.Namespace("bar")), // Different namespace than HTTPRoute.
												Port:      common.PtrTo(gatewayv1.PortNumber(80)),
											},
										},
									},
								},
							},
						},
					},
				},
				&corev1.Service{
					TypeMeta: metav1.TypeMeta{
						Kind:       "Service",
						APIVersion: "v1",
					},
					ObjectMeta: metav1.ObjectMeta{
						Name:      "bar-svc",
						Namespace: "bar",
					},
				},
				&gatewayv1beta1.ReferenceGrant{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "bar-reference-grant",
						Namespace: "bar",
					},
					Spec: gatewayv1beta1.ReferenceGrantSpec{
						From: []gatewayv1beta1.ReferenceGrantFrom{{
							Group:     gatewayv1.Group(gatewayv1.GroupVersion.Group),
							Kind:      "HTTPRoute",
							Namespace: "default",
						}},
						To: []gatewayv1beta1.ReferenceGrantTo{{
							Kind: "Service",
						}},
					},
				},
			},
			wantHTTPRoutes: []apimachinerytypes.NamespacedName{
				{Namespace: "default", Name: "foo-httproute"},
			},
			wantBackends: []apimachinerytypes.NamespacedName{
				{Namespace: "bar", Name: "bar-svc"},
			},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			k8sClients := common.MustClientsForTest(t, tc.objects...)
			policyManager := utils.MustPolicyManagerForTest(t, k8sClients)
			discoverer := Discoverer{
				K8sClients:    k8sClients,
				PolicyManager: policyManager,
			}

			resourceModel, err := discoverer.DiscoverResourcesForHTTPRoute(tc.filter)
			if err != nil {
				t.Fatalf("Failed to construct resourceModel: %v", resourceModel)
			}

			gotHTTPRoutes := namespacedHTTPRoutesFromResourceModel(resourceModel)
			gotBackends := namespacedBackendsFromResourceModel(resourceModel)

			if tc.wantHTTPRoutes != nil {
				if diff := cmp.Diff(tc.wantHTTPRoutes, gotHTTPRoutes, cmpopts.EquateEmpty()); diff != "" {
					t.Errorf("Unexpected diff in HTTPRoutes; got=%v, want=%v;\ndiff (-want +got)=\n%v", gotHTTPRoutes, tc.wantHTTPRoutes, diff)
				}
			}
			if tc.wantBackends != nil {
				if diff := cmp.Diff(tc.wantBackends, gotBackends, cmpopts.EquateEmpty()); diff != "" {
					t.Errorf("Unexpected diff in Backends; got=%v, want=%v;\ndiff (-want +got)=\n%v", gotBackends, tc.wantBackends, diff)
				}
			}
		})
	}
}

func TestDiscoverResourcesForBackend(t *testing.T) {
	testcases := []struct {
		name    string
		objects []runtime.Object
		filter  Filter

		wantBackends   []apimachinerytypes.NamespacedName
		wantHTTPRoutes []apimachinerytypes.NamespacedName
	}{
		{
			name:   "normal",
			filter: Filter{Labels: labels.Everything()},
			objects: []runtime.Object{
				common.NamespaceForTest("default"),
				&corev1.Service{
					TypeMeta: metav1.TypeMeta{
						Kind:       "Service",
						APIVersion: "v1",
					},
					ObjectMeta: metav1.ObjectMeta{
						Name:      "foo-svc",
						Namespace: "default",
					},
				},
				&gatewayv1.HTTPRoute{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "foo-httproute",
						Namespace: "default",
					},
					Spec: gatewayv1.HTTPRouteSpec{
						CommonRouteSpec: gatewayv1.CommonRouteSpec{},
						Rules: []gatewayv1.HTTPRouteRule{
							{
								BackendRefs: []gatewayv1.HTTPBackendRef{
									{
										BackendRef: gatewayv1.BackendRef{
											BackendObjectReference: gatewayv1.BackendObjectReference{
												Kind: common.PtrTo(gatewayv1.Kind("Service")),
												Name: "foo-svc",
												Port: common.PtrTo(gatewayv1.PortNumber(80)),
											},
										},
									},
								},
							},
						},
					},
				},
			},
			wantBackends: []apimachinerytypes.NamespacedName{
				{Namespace: "default", Name: "foo-svc"},
			},
			wantHTTPRoutes: []apimachinerytypes.NamespacedName{
				{Namespace: "default", Name: "foo-httproute"},
			},
		},
		{
			name:   "httproute from different namespace should require referencegrant",
			filter: Filter{Labels: labels.Everything()},
			objects: []runtime.Object{
				common.NamespaceForTest("default"),
				common.NamespaceForTest("bar"),
				&corev1.Service{
					TypeMeta: metav1.TypeMeta{
						Kind:       "Service",
						APIVersion: "v1",
					},
					ObjectMeta: metav1.ObjectMeta{
						Name:      "foo-svc",
						Namespace: "default",
					},
				},
				&gatewayv1.HTTPRoute{
					TypeMeta: metav1.TypeMeta{
						APIVersion: gatewayv1.GroupVersion.String(),
						Kind:       "HTTPRoute",
					},
					ObjectMeta: metav1.ObjectMeta{
						Name:      "bar-httproute",
						Namespace: "bar", // Different namespace than Service.
					},
					Spec: gatewayv1.HTTPRouteSpec{
						CommonRouteSpec: gatewayv1.CommonRouteSpec{},
						Rules: []gatewayv1.HTTPRouteRule{
							{
								BackendRefs: []gatewayv1.HTTPBackendRef{
									{
										BackendRef: gatewayv1.BackendRef{
											BackendObjectReference: gatewayv1.BackendObjectReference{
												Kind:      common.PtrTo(gatewayv1.Kind("Service")),
												Name:      "foo-svc",
												Namespace: common.PtrTo(gatewayv1.Namespace("default")),
												Port:      common.PtrTo(gatewayv1.PortNumber(80)),
											},
										},
									},
								},
							},
						},
					},
				},
			},
			wantBackends: []apimachinerytypes.NamespacedName{
				{Namespace: "default", Name: "foo-svc"},
			},
			wantHTTPRoutes: []apimachinerytypes.NamespacedName{},
		},
		{
			name:   "httproute from different namespace should get allowed with referencegrant",
			filter: Filter{Labels: labels.Everything()},
			objects: []runtime.Object{
				common.NamespaceForTest("default"),
				common.NamespaceForTest("bar"),
				&corev1.Service{
					TypeMeta: metav1.TypeMeta{
						Kind:       "Service",
						APIVersion: "v1",
					},
					ObjectMeta: metav1.ObjectMeta{
						Name:      "foo-svc",
						Namespace: "default",
					},
				},
				&gatewayv1.HTTPRoute{
					TypeMeta: metav1.TypeMeta{
						APIVersion: gatewayv1.GroupVersion.String(),
						Kind:       "HTTPRoute",
					},
					ObjectMeta: metav1.ObjectMeta{
						Name:      "bar-httproute",
						Namespace: "bar", // Different namespace than Service.
					},
					Spec: gatewayv1.HTTPRouteSpec{
						CommonRouteSpec: gatewayv1.CommonRouteSpec{},
						Rules: []gatewayv1.HTTPRouteRule{
							{
								BackendRefs: []gatewayv1.HTTPBackendRef{
									{
										BackendRef: gatewayv1.BackendRef{
											BackendObjectReference: gatewayv1.BackendObjectReference{
												Kind:      common.PtrTo(gatewayv1.Kind("Service")),
												Name:      "foo-svc",
												Namespace: common.PtrTo(gatewayv1.Namespace("default")),
												Port:      common.PtrTo(gatewayv1.PortNumber(80)),
											},
										},
									},
								},
							},
						},
					},
				},
				&gatewayv1beta1.ReferenceGrant{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "foo-reference-grant",
						Namespace: "default",
					},
					Spec: gatewayv1beta1.ReferenceGrantSpec{
						From: []gatewayv1beta1.ReferenceGrantFrom{{
							Group:     gatewayv1.Group(gatewayv1.GroupVersion.Group),
							Kind:      "HTTPRoute",
							Namespace: "bar",
						}},
						To: []gatewayv1beta1.ReferenceGrantTo{{
							Kind: "Service",
						}},
					},
				},
			},
			wantBackends: []apimachinerytypes.NamespacedName{
				{Namespace: "default", Name: "foo-svc"},
			},
			wantHTTPRoutes: []apimachinerytypes.NamespacedName{
				{Namespace: "bar", Name: "bar-httproute"},
			},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			k8sClients := common.MustClientsForTest(t, tc.objects...)
			policyManager := utils.MustPolicyManagerForTest(t, k8sClients)
			discoverer := Discoverer{
				K8sClients:    k8sClients,
				PolicyManager: policyManager,
			}

			resourceModel, err := discoverer.DiscoverResourcesForBackend(tc.filter)
			if err != nil {
				t.Fatalf("Failed to construct resourceModel: %v", resourceModel)
			}

			gotBackends := namespacedBackendsFromResourceModel(resourceModel)
			gotHTTPRoutes := namespacedHTTPRoutesFromResourceModel(resourceModel)

			if tc.wantBackends != nil {
				if diff := cmp.Diff(tc.wantBackends, gotBackends, cmpopts.EquateEmpty()); diff != "" {
					t.Errorf("Unexpected diff in Backends; got=%v, want=%v;\ndiff (-want +got)=\n%v", gotBackends, tc.wantBackends, diff)
				}
			}
			if tc.wantHTTPRoutes != nil {
				if diff := cmp.Diff(tc.wantHTTPRoutes, gotHTTPRoutes, cmpopts.EquateEmpty()); diff != "" {
					t.Errorf("Unexpected diff in HTTPRoutes; got=%v, want=%v;\ndiff (-want +got)=\n%v", gotHTTPRoutes, tc.wantHTTPRoutes, diff)
				}
			}
		})
	}
}

// TestDiscoverResourcesForGatewayClass_LabelSelector Tests label selector filtering for GatewayClasses.
func TestDiscoverResourcesForGatewayClass_LabelSelector(t *testing.T) {
	fakeClock := testingclock.NewFakeClock(time.Now())

	gatewayClass := func(name string, labels map[string]string) *gatewayv1.GatewayClass {
		return &gatewayv1.GatewayClass{
			ObjectMeta: metav1.ObjectMeta{
				Name:   name,
				Labels: labels,
				CreationTimestamp: metav1.Time{
					Time: fakeClock.Now().Add(-365 * 24 * time.Hour),
				},
			},
			Spec: gatewayv1.GatewayClassSpec{
				ControllerName: gatewayv1.GatewayController(name + "/controller"),
			},
			Status: gatewayv1.GatewayClassStatus{
				Conditions: []metav1.Condition{
					{
						Type:   "Accepted",
						Status: metav1.ConditionTrue,
					},
				},
			},
		}
	}
	objects := []runtime.Object{
		gatewayClass("foo-com-external-gateway-class", map[string]string{"app": "foo"}),
		gatewayClass("foo-com-internal-gateway-class", map[string]string{"app": "foo", "env": "internal"}),
	}
	k8sClients := common.MustClientsForTest(t, objects...)
	policyManager := utils.MustPolicyManagerForTest(t, k8sClients)
	discoverer := Discoverer{
		K8sClients:    k8sClients,
		PolicyManager: policyManager,
	}
	labelSelector := "env=internal"
	selector, err := labels.Parse(labelSelector)
	if err != nil {
		t.Errorf("Unable to find resources that match the label selector \"%s\": %v\n", labelSelector, err)
	}
	resourceModel, err := discoverer.DiscoverResourcesForGatewayClass(Filter{Labels: selector})
	if err != nil {
		t.Fatalf("Failed to construct resourceModel: %v", err)
	}

	expectedGatewayClassNames := []string{"foo-com-internal-gateway-class"}
	gatewayClassNames := make([]string, 0, len(resourceModel.GatewayClasses))
	for _, gatewayClassNode := range resourceModel.GatewayClasses {
		gatewayClassNames = append(gatewayClassNames, gatewayClassNode.GatewayClass.GetName())
	}
	if diff := cmp.Diff(expectedGatewayClassNames, gatewayClassNames); diff != "" {
		t.Errorf("Unexpected diff\ngot=\n%v\nwant=\n%v\ndiff (-want +got)=\n%v", gatewayClassNames, expectedGatewayClassNames, diff)
	}
}

// TestDiscoverResourcesForGateway_LabelSelector tests label selector filtering for Gateways.
func TestDiscoverResourcesForGateway_LabelSelector(t *testing.T) {
	fakeClock := testingclock.NewFakeClock(time.Now())
	gateway := func(name string, labels map[string]string) *gatewayv1.Gateway {
		return &gatewayv1.Gateway{
			ObjectMeta: metav1.ObjectMeta{
				Name:   name,
				Labels: labels,
				CreationTimestamp: metav1.Time{
					Time: fakeClock.Now().Add(-5 * 24 * time.Hour),
				},
			},
			Spec: gatewayv1.GatewaySpec{
				GatewayClassName: "gatewayclass-1",
				Listeners: []gatewayv1.Listener{
					{
						Name:     "http-8080",
						Protocol: gatewayv1.HTTPProtocolType,
						Port:     gatewayv1.PortNumber(8080),
					},
				},
			},
			Status: gatewayv1.GatewayStatus{
				Addresses: []gatewayv1.GatewayStatusAddress{
					{
						Value: "192.168.100.5",
					},
				},
				Conditions: []metav1.Condition{
					{
						Type:   "Programmed",
						Status: "False",
					},
				},
			},
		}
	}

	objects := []runtime.Object{
		&gatewayv1.GatewayClass{
			ObjectMeta: metav1.ObjectMeta{
				Name: "gatewayclass-1",
			},
			Spec: gatewayv1.GatewayClassSpec{
				ControllerName: "example.net/gateway-controller",
				Description:    common.PtrTo("random"),
			},
		},
		gateway("gateway-1", map[string]string{"app": "foo"}),
		gateway("gateway-2", map[string]string{"app": "foo", "env": "internal"}),
	}

	k8sClients := common.MustClientsForTest(t, objects...)
	policyManager := utils.MustPolicyManagerForTest(t, k8sClients)
	discoverer := Discoverer{
		K8sClients:    k8sClients,
		PolicyManager: policyManager,
	}
	labelSelector := "env=internal"
	selector, err := labels.Parse(labelSelector)
	if err != nil {
		t.Errorf("Unable to find resources that match the label selector \"%s\": %v\n", labelSelector, err)
	}
	resourceModel, err := discoverer.DiscoverResourcesForGateway(Filter{Labels: selector})
	if err != nil {
		t.Fatalf("Failed to construct resourceModel: %v", err)
	}

	expectedGatewayNames := []string{"gateway-2"}
	gatewayNames := make([]string, 0, len(resourceModel.Gateways))
	for _, gatewayNode := range resourceModel.Gateways {
		gatewayNames = append(gatewayNames, gatewayNode.Gateway.GetName())
	}

	if diff := cmp.Diff(expectedGatewayNames, gatewayNames); diff != "" {
		t.Errorf("Unexpected diff\ngot=\n%v\nwant=\n%v\ndiff (-want +got)=\n%v", gatewayNames, expectedGatewayNames, diff)
	}
}

// TestDiscoverResourcesForHTTPRoute_LabelSelector tests label selector filtering for HTTPRoute.
func TestDiscoverResourcesForHTTPRoute_LabelSelector(t *testing.T) {
	fakeClock := testingclock.NewFakeClock(time.Now())
	httpRoute := func(name string, labels map[string]string) *gatewayv1.HTTPRoute {
		return &gatewayv1.HTTPRoute{
			ObjectMeta: metav1.ObjectMeta{
				Name:      name,
				Namespace: "default",
				CreationTimestamp: metav1.Time{
					Time: fakeClock.Now().Add(-24 * time.Hour),
				},
				Labels: labels,
			},
			Spec: gatewayv1.HTTPRouteSpec{
				Hostnames: []gatewayv1.Hostname{"example.com"},
				CommonRouteSpec: gatewayv1.CommonRouteSpec{
					ParentRefs: []gatewayv1.ParentReference{
						{
							Name: "gateway-1",
						},
					},
				},
			},
		}
	}

	objects := []runtime.Object{
		&gatewayv1.GatewayClass{
			ObjectMeta: metav1.ObjectMeta{
				Name: "gatewayclass-1",
			},
			Spec: gatewayv1.GatewayClassSpec{
				ControllerName: "example.net/gateway-controller",
				Description:    common.PtrTo("random"),
			},
		},

		&gatewayv1.Gateway{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "gateway-1",
				Namespace: "default",
			},
			Spec: gatewayv1.GatewaySpec{
				GatewayClassName: "gatewayclass-1",
			},
		},
		httpRoute("httproute-1", map[string]string{"app": "foo"}),
		httpRoute("httproute-2", map[string]string{"app": "foo", "env": "internal"}),
	}

	k8sClients := common.MustClientsForTest(t, objects...)
	policyManager := utils.MustPolicyManagerForTest(t, k8sClients)
	discoverer := Discoverer{
		K8sClients:    k8sClients,
		PolicyManager: policyManager,
	}

	labelSelector := "env=internal"
	selector, err := labels.Parse(labelSelector)
	if err != nil {
		t.Errorf("Unable to find resources that match the label selector \"%s\": %v\n", labelSelector, err)
	}
	resourceModel, err := discoverer.DiscoverResourcesForHTTPRoute(Filter{Labels: selector})
	if err != nil {
		t.Fatalf("Failed to discover resources: %v", err)
	}

	expectedHTTPRouteNames := []string{"httproute-2"}
	HTTPRouteNames := make([]string, 0, len(resourceModel.HTTPRoutes))
	for _, HTTPRouteNode := range resourceModel.HTTPRoutes {
		HTTPRouteNames = append(HTTPRouteNames, HTTPRouteNode.HTTPRoute.GetName())
	}

	if diff := cmp.Diff(expectedHTTPRouteNames, HTTPRouteNames); diff != "" {
		t.Errorf("Unexpected diff\ngot=\n%v\nwant=\n%v\ndiff (-want +got)=\n%v", expectedHTTPRouteNames, HTTPRouteNames, diff)
	}
}

// TestDiscoverResourcesForNamespace_LabelSelector tests label selector filtering for Namespaces.
func TestDiscoverResourcesForNamespace_LabelSelector(t *testing.T) {
	fakeClock := testingclock.NewFakeClock(time.Now())
	namespace := func(name string, labels map[string]string) *corev1.Namespace {
		return &corev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name: name,
				CreationTimestamp: metav1.Time{
					Time: fakeClock.Now().Add(-46 * 24 * time.Hour),
				},
				Labels: labels,
			},
			Status: corev1.NamespaceStatus{
				Phase: corev1.NamespaceActive,
			},
		}
	}

	objects := []runtime.Object{
		namespace("namespace-1", map[string]string{"app": "foo"}),
		namespace("namespace-2", map[string]string{"app": "foo", "env": "internal"}),
	}

	k8sClients := common.MustClientsForTest(t, objects...)
	policyManager := utils.MustPolicyManagerForTest(t, k8sClients)
	discoverer := Discoverer{
		K8sClients:    k8sClients,
		PolicyManager: policyManager,
	}
	labelSelector := "env=internal"
	selector, err := labels.Parse(labelSelector)
	if err != nil {
		t.Errorf("Unable to find resources that match the label selector \"%s\": %v\n", labelSelector, err)
	}
	resourceModel, err := discoverer.DiscoverResourcesForNamespace(Filter{Labels: selector})
	if err != nil {
		t.Fatalf("Failed to construct resourceModel: %v", err)
	}

	expectedNamespaceNames := []string{"namespace-2"}
	namespaceNames := make([]string, 0, len(resourceModel.Namespaces))
	for _, namespaceNode := range resourceModel.Namespaces {
		namespaceNames = append(namespaceNames, namespaceNode.Namespace.Name)
	}

	if diff := cmp.Diff(expectedNamespaceNames, namespaceNames); diff != "" {
		t.Errorf("Unexpected diff\ngot=\n%v\nwant=\n%v\ndiff (-want +got)=\n%v", expectedNamespaceNames, namespaceNames, diff)
	}
}

func gatewayClassNamesFromResourceModel(r *ResourceModel) []string {
	var gatewayClassNames []string
	for _, gatewayClassNode := range r.GatewayClasses {
		gatewayClassNames = append(gatewayClassNames, gatewayClassNode.GatewayClass.GetName())
	}
	return gatewayClassNames
}

func namespacedGatewaysFromResourceModel(r *ResourceModel) []apimachinerytypes.NamespacedName {
	var gateways []apimachinerytypes.NamespacedName
	for _, gatewayNode := range r.Gateways {
		gateways = append(gateways, apimachinerytypes.NamespacedName{
			Namespace: gatewayNode.Gateway.GetNamespace(),
			Name:      gatewayNode.Gateway.GetName(),
		})
	}
	return gateways
}

func namespacedHTTPRoutesFromResourceModel(r *ResourceModel) []apimachinerytypes.NamespacedName {
	var httpRoutes []apimachinerytypes.NamespacedName
	for _, httpRouteNode := range r.HTTPRoutes {
		httpRoutes = append(httpRoutes, apimachinerytypes.NamespacedName{
			Namespace: httpRouteNode.HTTPRoute.GetNamespace(),
			Name:      httpRouteNode.HTTPRoute.GetName(),
		})
	}
	return httpRoutes
}

func namespacedBackendsFromResourceModel(r *ResourceModel) []apimachinerytypes.NamespacedName {
	var backends []apimachinerytypes.NamespacedName
	for _, backendNode := range r.Backends {
		backends = append(backends, apimachinerytypes.NamespacedName{
			Namespace: backendNode.Backend.GetNamespace(),
			Name:      backendNode.Backend.GetName(),
		})
	}
	return backends
}

func gatewayErrorsFromResourceModel(r *ResourceModel) map[apimachinerytypes.NamespacedName][]error {
	result := make(map[apimachinerytypes.NamespacedName][]error)
	for _, gatewayNode := range r.Gateways {
		gatewayNN := apimachinerytypes.NamespacedName{
			Namespace: gatewayNode.Gateway.GetNamespace(),
			Name:      gatewayNode.Gateway.GetName(),
		}
		result[gatewayNN] = gatewayNode.Errors
	}
	return result
}
