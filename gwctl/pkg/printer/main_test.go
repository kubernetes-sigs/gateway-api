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
	"flag"
	"os"
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/utils/ptr"

	"k8s.io/klog/v2"

	gatewayv1 "sigs.k8s.io/gateway-api/apis/v1"
	"sigs.k8s.io/gateway-api/gwctl/pkg/common"
	"sigs.k8s.io/gateway-api/gwctl/pkg/policymanager"
	"sigs.k8s.io/gateway-api/gwctl/pkg/topology"
)

func TestMain(m *testing.M) {
	fs := flag.NewFlagSet("mock-flags", flag.PanicOnError)
	klog.InitFlags(fs)
	fs.Set("v", "3") // Set klog verbosity.

	os.Exit(m.Run())
}

func testData(t *testing.T) map[schema.GroupKind][]*topology.Node {
	ns1 := mustNewNode(t, &corev1.Namespace{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "Namespace",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: "ns-1",
		},
		Status: corev1.NamespaceStatus{
			Phase: corev1.NamespaceActive,
		},
	})

	gatewayClass1 := mustNewNode(t, &gatewayv1.GatewayClass{
		TypeMeta: metav1.TypeMeta{
			APIVersion: gatewayv1.GroupVersion.String(),
			Kind:       "GatewayClass",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: "gateway-class-1",
		},
		Spec: gatewayv1.GatewayClassSpec{
			ControllerName: "foo.com/external-gateway-class",
		},
		Status: gatewayv1.GatewayClassStatus{
			Conditions: []metav1.Condition{
				{
					Type:   "Accepted",
					Status: "True",
				},
			},
		},
	})

	gateway1 := mustNewNode(t,
		&gatewayv1.Gateway{
			TypeMeta: metav1.TypeMeta{
				APIVersion: gatewayv1.GroupVersion.String(),
				Kind:       "Gateway",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:      "gateway-1",
				Namespace: "ns-1",
			},
			Spec: gatewayv1.GatewaySpec{
				GatewayClassName: "gateway-class-1",
				Listeners: []gatewayv1.Listener{
					{
						Name:     gatewayv1.SectionName("http-80"),
						Protocol: gatewayv1.HTTPProtocolType,
						Port:     gatewayv1.PortNumber(80),
					},
				},
			},
			Status: gatewayv1.GatewayStatus{
				Addresses: []gatewayv1.GatewayStatusAddress{
					{
						Value: "10.0.0.1",
					},
					{
						Value: "10.0.0.2",
					},
					{
						Value: "10.0.0.3",
					},
				},
				Conditions: []metav1.Condition{
					{
						Type:   "Programmed",
						Status: "True",
					},
				},
			},
		},
	)

	httpRoute1 := mustNewNode(t,
		&gatewayv1.HTTPRoute{
			TypeMeta: metav1.TypeMeta{
				APIVersion: gatewayv1.GroupVersion.String(),
				Kind:       "HTTPRoute",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:      "http-route-1",
				Namespace: "ns-1",
			},
			Spec: gatewayv1.HTTPRouteSpec{
				Hostnames: []gatewayv1.Hostname{"foo.com", "bar.com", "example.com", "example2.com", "example3.com", "example4.com", "example5.com"},
				CommonRouteSpec: gatewayv1.CommonRouteSpec{
					ParentRefs: []gatewayv1.ParentReference{
						{
							Kind:      ptr.To(gatewayv1.Kind("Gateway")),
							Group:     ptr.To(gatewayv1.Group("gateway.networking.k8s.io")),
							Namespace: ptr.To(gatewayv1.Namespace("ns-1")),
							Name:      "gateway-1",
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
										Name:      gatewayv1.ObjectName("service-1"),
										Kind:      ptr.To(gatewayv1.Kind("Service")),
										Namespace: ptr.To(gatewayv1.Namespace("ns-1")),
									},
								},
							},
						},
					},
				},
			},
		},
	)

	service1 := mustNewNode(t, &corev1.Service{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Service",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "svc-1",
			Namespace: "ns-1",
		},
	})

	graph := &topology.Graph{}
	graph.AddNode(ns1)
	graph.AddNode(gatewayClass1)
	graph.AddNode(gateway1)
	graph.AddNode(httpRoute1)
	graph.AddNode(service1)

	result := map[schema.GroupKind][]*topology.Node{}
	for gk, nodes := range graph.Nodes {
		for _, node := range nodes {
			result[gk] = append(result[gk], node)
		}
	}
	return result
}

func testPoliciesData(t *testing.T) map[schema.GroupKind][]*topology.Node {
	return map[schema.GroupKind][]*topology.Node{
		common.PolicyGK: {
			mustNewPolicyNode(t, &unstructured.Unstructured{
				Object: map[string]interface{}{
					"apiVersion": "bar.com/v1",
					"kind":       "TimeoutPolicy",
					"metadata": map[string]interface{}{
						"name":      "policy-1",
						"namespace": "ns-1",
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
			}, false),
		},
	}
}

func mustNewNode(t *testing.T, obj runtime.Object) *topology.Node {
	u, err := runtime.DefaultUnstructuredConverter.ToUnstructured(obj)
	if err != nil {
		t.Fatal(err)
	}
	return &topology.Node{Object: &unstructured.Unstructured{Object: u}}
}

func mustNewPolicyNode(t *testing.T, u *unstructured.Unstructured, inherited bool) *topology.Node {
	policy, err := policymanager.ConstructPolicy(u, inherited)
	if err != nil {
		t.Fatal(err)
	}

	return &topology.Node{
		Object: policy.Unstructured,
		Metadata: map[string]any{
			common.PolicyGK.String(): policy,
		},
	}
}
