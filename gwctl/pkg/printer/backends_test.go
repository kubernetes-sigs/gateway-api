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

package printer

import (
	"bytes"
	"testing"
	"time"

	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/utils/ptr"

	gatewayv1alpha2 "sigs.k8s.io/gateway-api/apis/v1alpha2"

	"github.com/google/go-cmp/cmp"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	testingclock "k8s.io/utils/clock/testing"

	gatewayv1 "sigs.k8s.io/gateway-api/apis/v1"

	"sigs.k8s.io/gateway-api/gwctl/pkg/common"
	"sigs.k8s.io/gateway-api/gwctl/pkg/resourcediscovery"
	"sigs.k8s.io/gateway-api/gwctl/pkg/utils"
)

func TestBackendsPrinter_Print(t *testing.T) {
	fakeClock := testingclock.NewFakeClock(time.Now())

	healthCheckPolicies := []runtime.Object{
		&apiextensionsv1.CustomResourceDefinition{
			ObjectMeta: metav1.ObjectMeta{
				Name: "healthcheckpolicies.foo.com",
				Labels: map[string]string{
					gatewayv1alpha2.PolicyLabelKey: "inherited",
				},
			},
			Spec: apiextensionsv1.CustomResourceDefinitionSpec{
				Scope:    apiextensionsv1.ClusterScoped,
				Group:    "foo.com",
				Versions: []apiextensionsv1.CustomResourceDefinitionVersion{{Name: "v1"}},
				Names: apiextensionsv1.CustomResourceDefinitionNames{
					Plural: "healthcheckpolicies",
					Kind:   "HealthCheckPolicy",
				},
			},
		},
		&unstructured.Unstructured{
			Object: map[string]interface{}{
				"apiVersion": "foo.com/v1",
				"kind":       "HealthCheckPolicy",
				"metadata": map[string]interface{}{
					"name":              "health-check-gatewayclass",
					"creationTimestamp": fakeClock.Now().Add(-6 * 24 * time.Hour).Format(time.RFC3339),
				},
				"spec": map[string]interface{}{
					"override": map[string]interface{}{
						"key1": "value-parent-1",
						"key3": "value-parent-3",
						"key5": "value-parent-5",
					},
					"default": map[string]interface{}{
						"key2": "value-parent-2",
						"key4": "value-parent-4",
					},
					"targetRef": map[string]interface{}{
						"group":     "",
						"kind":      "Service",
						"name":      "foo-svc-0",
						"namespace": "default",
					},
				},
			},
		},
		&unstructured.Unstructured{
			Object: map[string]interface{}{
				"apiVersion": "foo.com/v1",
				"kind":       "HealthCheckPolicy",
				"metadata": map[string]interface{}{
					"name":              "health-check-gateway",
					"creationTimestamp": fakeClock.Now().Add(-20 * 24 * time.Hour).Format(time.RFC3339),
				},
				"spec": map[string]interface{}{
					"override": map[string]interface{}{
						"key1": "value-child-1",
					},
					"default": map[string]interface{}{
						"key2": "value-child-2",
						"key5": "value-child-5",
					},
					"targetRef": map[string]interface{}{
						"group":     "",
						"kind":      "Service",
						"name":      "foo-svc-1",
						"namespace": "ns1",
					},
				},
			},
		},
	}

	timeoutPolicies := []runtime.Object{
		&apiextensionsv1.CustomResourceDefinition{
			ObjectMeta: metav1.ObjectMeta{
				Name: "timeoutpolicies.bar.com",
				Labels: map[string]string{
					gatewayv1alpha2.PolicyLabelKey: "direct",
				},
			},
			Spec: apiextensionsv1.CustomResourceDefinitionSpec{
				Scope:    apiextensionsv1.ClusterScoped,
				Group:    "bar.com",
				Versions: []apiextensionsv1.CustomResourceDefinitionVersion{{Name: "v1"}},
				Names: apiextensionsv1.CustomResourceDefinitionNames{
					Plural: "timeoutpolicies",
					Kind:   "TimeoutPolicy",
				},
			},
		},
		&unstructured.Unstructured{
			Object: map[string]interface{}{
				"apiVersion": "bar.com/v1",
				"kind":       "TimeoutPolicy",
				"metadata": map[string]interface{}{
					"name":              "timeout-policy-namespace",
					"creationTimestamp": fakeClock.Now().Add(-5 * time.Minute).Format(time.RFC3339),
				},
				"spec": map[string]interface{}{
					"condition": "path=/abc",
					"seconds":   int64(30),
					"targetRef": map[string]interface{}{
						"kind": "Namespace",
						"name": "default",
					},
				},
			},
		},
		&unstructured.Unstructured{
			Object: map[string]interface{}{
				"apiVersion": "bar.com/v1",
				"kind":       "TimeoutPolicy",
				"metadata": map[string]interface{}{
					"name":              "timeout-policy-httproute",
					"creationTimestamp": fakeClock.Now().Add(-13 * time.Minute).Format(time.RFC3339),
				},
				"spec": map[string]interface{}{
					"condition": "path=/def",
					"seconds":   int64(60),
					"targetRef": map[string]interface{}{
						"group":     "gateway.networking.k8s.io",
						"kind":      "HTTPRoute",
						"name":      "bar-route-21",
						"namespace": "ns1",
					},
				},
			},
		},
	}

	objects := []runtime.Object{
		&gatewayv1.GatewayClass{
			ObjectMeta: metav1.ObjectMeta{
				Name: "demo-gatewayclass-1",
			},
			Spec: gatewayv1.GatewayClassSpec{
				ControllerName: "example.net/gateway-controller",
				Description:    common.PtrTo("random"),
			},
		},
		&corev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name: "ns1",
			},
			Status: corev1.NamespaceStatus{
				Phase: corev1.NamespaceActive,
			},
		},
		&corev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name: "ns2",
			},
			Status: corev1.NamespaceStatus{
				Phase: corev1.NamespaceActive,
			},
		},
		&corev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name: "ns3",
			},
			Status: corev1.NamespaceStatus{
				Phase: corev1.NamespaceActive,
			},
		},
		&corev1.Service{
			TypeMeta: metav1.TypeMeta{
				Kind:       "Service",
				APIVersion: "v1",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:      "foo-svc-0",
				Namespace: "default",
				CreationTimestamp: metav1.Time{
					Time: fakeClock.Now().Add(-72 * time.Hour),
				},
			},
		},
		&corev1.Service{
			TypeMeta: metav1.TypeMeta{
				Kind:       "Service",
				APIVersion: "v1",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:      "foo-svc-1",
				Namespace: "ns1",
				CreationTimestamp: metav1.Time{
					Time: fakeClock.Now().Add(-48 * time.Hour),
				},
			},
		},
		&corev1.Service{
			TypeMeta: metav1.TypeMeta{
				Kind:       "Service",
				APIVersion: "v1",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:      "foo-svc-2",
				Namespace: "ns2",
				CreationTimestamp: metav1.Time{
					Time: fakeClock.Now().Add(-36 * time.Hour),
				},
			},
		},
		&corev1.Service{
			TypeMeta: metav1.TypeMeta{
				Kind:       "Service",
				APIVersion: "v1",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:      "foo-svc-3",
				Namespace: "ns3",
				CreationTimestamp: metav1.Time{
					Time: fakeClock.Now().Add(-24 * time.Hour),
				},
			},
		},
		&corev1.Service{
			TypeMeta: metav1.TypeMeta{
				Kind:       "Service",
				APIVersion: "v1",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:      "foo-svc-4",
				Namespace: "ns3",
				CreationTimestamp: metav1.Time{
					Time: fakeClock.Now().Add(-128 * time.Hour),
				},
			},
		},
		&gatewayv1.Gateway{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "demo-gateway-1",
				Namespace: "default",
			},
			Spec: gatewayv1.GatewaySpec{
				GatewayClassName: "demo-gatewayclass-1",
			},
		},
		&gatewayv1.Gateway{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "demo-gateway-2",
				Namespace: "ns2",
			},
			Spec: gatewayv1.GatewaySpec{
				GatewayClassName: "demo-gatewayclass-1",
			},
		},
		&gatewayv1.Gateway{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "demo-gateway-345",
				Namespace: "ns1",
			},
			Spec: gatewayv1.GatewaySpec{
				GatewayClassName: "demo-gatewayclass-1",
			},
		},
		&gatewayv1.HTTPRoute{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "foo-httproute-1",
				Namespace: "default",
				CreationTimestamp: metav1.Time{
					Time: fakeClock.Now().Add(-24 * time.Hour),
				},
			},
			Spec: gatewayv1.HTTPRouteSpec{
				Hostnames: []gatewayv1.Hostname{"example.com", "example2.com", "example3.com"},
				CommonRouteSpec: gatewayv1.CommonRouteSpec{
					ParentRefs: []gatewayv1.ParentReference{
						{
							Kind:      common.PtrTo(gatewayv1.Kind("Gateway")),
							Group:     common.PtrTo(gatewayv1.Group("gateway.networking.k8s.io")),
							Namespace: common.PtrTo(gatewayv1.Namespace("ns2")),
							Name:      "demo-gateway-2",
						},
					},
				},
				Rules: []gatewayv1.HTTPRouteRule{
					{
						BackendRefs: []gatewayv1.HTTPBackendRef{
							{
								BackendRef: gatewayv1.BackendRef{
									BackendObjectReference: gatewayv1.BackendObjectReference{
										Port:      ptr.To(gatewayv1.PortNumber(8080)),
										Name:      "foo-svc-0",
										Kind:      ptr.To(gatewayv1.Kind("Service")),
										Namespace: ptr.To(gatewayv1.Namespace("default")),
									},
								},
							},
							{
								BackendRef: gatewayv1.BackendRef{
									BackendObjectReference: gatewayv1.BackendObjectReference{
										Port:      ptr.To(gatewayv1.PortNumber(8081)),
										Name:      "foo-svc-1",
										Kind:      ptr.To(gatewayv1.Kind("Service")),
										Namespace: ptr.To(gatewayv1.Namespace("ns1")),
									},
								},
							},
							{
								BackendRef: gatewayv1.BackendRef{
									BackendObjectReference: gatewayv1.BackendObjectReference{
										Port:      ptr.To(gatewayv1.PortNumber(8082)),
										Name:      "foo-svc-2",
										Kind:      ptr.To(gatewayv1.Kind("Service")),
										Namespace: ptr.To(gatewayv1.Namespace("ns2")),
									},
								},
							},
						},
					},
				},
			},
		},
		&gatewayv1.HTTPRoute{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "qmn-httproute-100",
				Namespace: "default",
				CreationTimestamp: metav1.Time{
					Time: fakeClock.Now().Add(-11 * time.Hour),
				},
			},
			Spec: gatewayv1.HTTPRouteSpec{
				Hostnames: []gatewayv1.Hostname{"example.com"},
				CommonRouteSpec: gatewayv1.CommonRouteSpec{
					ParentRefs: []gatewayv1.ParentReference{
						{
							Kind:  common.PtrTo(gatewayv1.Kind("Gateway")),
							Group: common.PtrTo(gatewayv1.Group("gateway.networking.k8s.io")),
							Name:  "demo-gateway-1",
						},
						{
							Kind:  common.PtrTo(gatewayv1.Kind("Gateway")),
							Group: common.PtrTo(gatewayv1.Group("gateway.networking.k8s.io")),
							Name:  "demo-gateway-345",
						},
					},
				},
				Rules: []gatewayv1.HTTPRouteRule{
					{
						BackendRefs: []gatewayv1.HTTPBackendRef{
							{
								BackendRef: gatewayv1.BackendRef{
									BackendObjectReference: gatewayv1.BackendObjectReference{
										Port:      ptr.To(gatewayv1.PortNumber(8081)),
										Name:      "foo-svc-1",
										Kind:      ptr.To(gatewayv1.Kind("Service")),
										Namespace: ptr.To(gatewayv1.Namespace("ns1")),
									},
								},
							},
							{
								BackendRef: gatewayv1.BackendRef{
									BackendObjectReference: gatewayv1.BackendObjectReference{
										Port:      ptr.To(gatewayv1.PortNumber(8082)),
										Name:      "foo-svc-2",
										Kind:      ptr.To(gatewayv1.Kind("Service")),
										Namespace: ptr.To(gatewayv1.Namespace("ns2")),
									},
								},
							},
							{
								BackendRef: gatewayv1.BackendRef{
									BackendObjectReference: gatewayv1.BackendObjectReference{
										Port:      ptr.To(gatewayv1.PortNumber(8083)),
										Name:      "foo-svc-3",
										Kind:      ptr.To(gatewayv1.Kind("Service")),
										Namespace: ptr.To(gatewayv1.Namespace("ns3")),
									},
								},
							},
						},
					},
				},
			},
		},
		&gatewayv1.HTTPRoute{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "bar-route-21",
				Namespace: "ns1",
				CreationTimestamp: metav1.Time{
					Time: fakeClock.Now().Add(-9 * time.Hour),
				},
			},
			Spec: gatewayv1.HTTPRouteSpec{
				Hostnames: []gatewayv1.Hostname{"foo.com", "bar.com", "example.com", "example2.com", "example3.com", "example4.com", "example5.com"},
				CommonRouteSpec: gatewayv1.CommonRouteSpec{
					ParentRefs: []gatewayv1.ParentReference{
						{
							Kind:      common.PtrTo(gatewayv1.Kind("Gateway")),
							Group:     common.PtrTo(gatewayv1.Group("gateway.networking.k8s.io")),
							Namespace: common.PtrTo(gatewayv1.Namespace("default")),
							Name:      "demo-gateway-2",
						},
					},
				},
				Rules: []gatewayv1.HTTPRouteRule{
					{
						BackendRefs: []gatewayv1.HTTPBackendRef{
							{
								BackendRef: gatewayv1.BackendRef{
									BackendObjectReference: gatewayv1.BackendObjectReference{
										Port:      ptr.To(gatewayv1.PortNumber(8082)),
										Name:      "foo-svc-2",
										Kind:      ptr.To(gatewayv1.Kind("Service")),
										Namespace: ptr.To(gatewayv1.Namespace("ns2")),
									},
								},
							},
							{
								BackendRef: gatewayv1.BackendRef{
									BackendObjectReference: gatewayv1.BackendObjectReference{
										Port:      ptr.To(gatewayv1.PortNumber(8083)),
										Name:      "foo-svc-3",
										Kind:      ptr.To(gatewayv1.Kind("Service")),
										Namespace: ptr.To(gatewayv1.Namespace("ns3")),
									},
								},
							},
						},
					},
				},
			},
		},
	}

	finalObjects := []runtime.Object{}
	finalObjects = append(finalObjects, healthCheckPolicies...)
	finalObjects = append(finalObjects, timeoutPolicies...)
	finalObjects = append(finalObjects, objects...)

	params := utils.MustParamsForTest(t, common.MustClientsForTest(t, finalObjects...))
	discoverer := resourcediscovery.Discoverer{
		K8sClients:    params.K8sClients,
		PolicyManager: params.PolicyManager,
	}
	resourceModel, err := discoverer.DiscoverResourcesForBackend(resourcediscovery.Filter{})
	if err != nil {
		t.Fatalf("Failed to construct resourceModel %v: %v", resourceModel, err)
	}

	bp := &BackendsPrinter{
		Writer: params.Out,
		Clock:  fakeClock,
	}

	bp.Print(resourceModel)

	got := params.Out.(*bytes.Buffer).String()
	want := `
NAMESPACE  NAME       TYPE     REFERRED BY ROUTES                                          AGE   POLICIES
default    foo-svc-0  Service  default/foo-httproute-1                                     3d    1
ns1        foo-svc-1  Service  default/foo-httproute-1,default/qmn-httproute-100           2d    1
ns2        foo-svc-2  Service  default/foo-httproute-1,default/qmn-httproute-100 + 1 more  36h   0
ns3        foo-svc-3  Service  default/qmn-httproute-100,ns1/bar-route-21                  24h   0
ns3        foo-svc-4  Service  None                                                        5d8h  0
`
	if diff := cmp.Diff(common.YamlString(want), common.YamlString(got), common.YamlStringTransformer); diff != "" {
		t.Errorf("Unexpected diff\ngot=\n%v\nwant=\n%v\ndiff (-want +got)=\n%v", got, want, diff)
	}
}
