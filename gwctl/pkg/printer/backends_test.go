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

	"github.com/google/go-cmp/cmp"
	corev1 "k8s.io/api/core/v1"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	testingclock "k8s.io/utils/clock/testing"
	"k8s.io/utils/ptr"

	gatewayv1 "sigs.k8s.io/gateway-api/apis/v1"
	gatewayv1alpha2 "sigs.k8s.io/gateway-api/apis/v1alpha2"
	"sigs.k8s.io/gateway-api/gwctl/pkg/common"
	"sigs.k8s.io/gateway-api/gwctl/pkg/resourcediscovery"
	"sigs.k8s.io/gateway-api/gwctl/pkg/utils"
)

func TestBackendsPrinter_Print(t *testing.T) {
	fakeClock := testingclock.NewFakeClock(time.Now())

	httpRoute := func(namespace, name, serviceName, gatewayName string) *gatewayv1.HTTPRoute {
		return &gatewayv1.HTTPRoute{
			ObjectMeta: metav1.ObjectMeta{
				Name:      name,
				Namespace: namespace,
				CreationTimestamp: metav1.Time{
					Time: fakeClock.Now().Add(-24 * time.Hour),
				},
			},
			Spec: gatewayv1.HTTPRouteSpec{
				CommonRouteSpec: gatewayv1.CommonRouteSpec{
					ParentRefs: []gatewayv1.ParentReference{
						{
							Kind:      common.PtrTo(gatewayv1.Kind("Gateway")),
							Group:     common.PtrTo(gatewayv1.Group("gateway.networking.k8s.io")),
							Namespace: common.PtrTo(gatewayv1.Namespace(namespace)),
							Name:      gatewayv1.ObjectName(gatewayName),
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
										Name:      gatewayv1.ObjectName(serviceName),
										Kind:      ptr.To(gatewayv1.Kind("Service")),
										Namespace: ptr.To(gatewayv1.Namespace(namespace)),
									},
								},
							},
						},
					},
				},
			},
		}
	}

	objects := []runtime.Object{
		common.NamespaceForTest("ns1"),
		&gatewayv1.GatewayClass{
			ObjectMeta: metav1.ObjectMeta{
				Name: "foo-gatewayclass-1",
			},
			Spec: gatewayv1.GatewayClassSpec{
				ControllerName: "example.net/gateway-controller",
				Description:    common.PtrTo("random"),
			},
		},
		&gatewayv1.Gateway{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "foo-gateway-1",
				Namespace: "ns1",
			},
			Spec: gatewayv1.GatewaySpec{
				GatewayClassName: "foo-gatewayclass-1",
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
				Name:      "foo-svc-2",
				Namespace: "ns1",
				CreationTimestamp: metav1.Time{
					Time: fakeClock.Now().Add(-48 * time.Hour),
				},
			},
		},
		httpRoute("ns1", "foo-httproute-1", "foo-svc-1", "foo-gateway-1"),
		httpRoute("ns1", "foo-httproute-2", "foo-svc-2", "foo-gateway-1"),
		httpRoute("ns1", "foo-httproute-3", "foo-svc-2", "foo-gateway-1"),
		httpRoute("ns1", "foo-httproute-4", "foo-svc-2", "foo-gateway-1"),
		httpRoute("ns1", "foo-httproute-5", "foo-svc-2", "foo-gateway-1"),
	}

	backendPolicies := []runtime.Object{
		&apiextensionsv1.CustomResourceDefinition{
			ObjectMeta: metav1.ObjectMeta{
				Name: "healthcheckpolicies.foo.com",
				Labels: map[string]string{
					gatewayv1alpha2.PolicyLabelKey: "Direct",
				},
			},
			Spec: apiextensionsv1.CustomResourceDefinitionSpec{
				Scope:    apiextensionsv1.NamespaceScoped,
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
					"namespace":         "default",
					"creationTimestamp": fakeClock.Now().Add(-6 * 24 * time.Hour).Format(time.RFC3339),
				},
				"spec": map[string]interface{}{
					"default": map[string]interface{}{
						"key2": "value-parent-2",
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

	var finalObjects []runtime.Object
	finalObjects = append(finalObjects, objects...)
	finalObjects = append(finalObjects, backendPolicies...)

	k8sClients := common.MustClientsForTest(t, finalObjects...)
	policyManager := utils.MustPolicyManagerForTest(t, k8sClients)
	buff := &bytes.Buffer{}
	discoverer := resourcediscovery.Discoverer{
		K8sClients:    k8sClients,
		PolicyManager: policyManager,
	}
	resourceModel, err := discoverer.DiscoverResourcesForBackend(resourcediscovery.Filter{})
	if err != nil {
		t.Fatalf("Failed to construct resourceModel %v: %v", resourceModel, err)
	}

	bp := &BackendsPrinter{
		Writer: buff,
		Clock:  fakeClock,
	}

	bp.PrintTable(resourceModel, false)

	got := buff.String()
	want := `
NAMESPACE  NAME       TYPE     AGE
ns1        foo-svc-1  Service  3d
ns1        foo-svc-2  Service  2d
`
	if diff := cmp.Diff(common.YamlString(want), common.YamlString(got), common.YamlStringTransformer); diff != "" {
		t.Errorf("Unexpected diff\ngot=\n%v\nwant=\n%v\ndiff (-want +got)=\n%v", got, want, diff)
	}

	buff.Reset()
	nsp2 := &BackendsPrinter{
		Writer: buff,
		Clock:  fakeClock,
	}
	nsp2.PrintTable(resourceModel, true)

	got2 := buff.String()
	want2 := `
NAMESPACE  NAME       TYPE     AGE  REFERRED BY ROUTES                                 POLICIES
ns1        foo-svc-1  Service  3d   ns1/foo-httproute-1                                1
ns1        foo-svc-2  Service  2d   ns1/foo-httproute-2, ns1/foo-httproute-3 + 2 more  0
`
	if diff := cmp.Diff(common.YamlString(want2), common.YamlString(got2), common.YamlStringTransformer); diff != "" {
		t.Errorf("Unexpected diff\ngot=\n%v\nwant=\n%v\ndiff (-want +got)=\n%v", got2, want2, diff)
	}
}
