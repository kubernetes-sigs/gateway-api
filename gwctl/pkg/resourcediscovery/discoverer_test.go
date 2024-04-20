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
	corev1 "k8s.io/api/core/v1"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	testingclock "k8s.io/utils/clock/testing"

	gatewayv1 "sigs.k8s.io/gateway-api/apis/v1"
	"sigs.k8s.io/gateway-api/gwctl/pkg/common"
	"sigs.k8s.io/gateway-api/gwctl/pkg/utils"
)

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
	params := utils.MustParamsForTest(t, common.MustClientsForTest(t, objects...))
	discoverer := Discoverer{
		K8sClients:    params.K8sClients,
		PolicyManager: params.PolicyManager,
	}
	labelSelector := "env=internal"
	selector, err := labels.Parse(labelSelector)
	if err != nil {
		t.Errorf("Unable to find resources that match the label selector \"%s\": %v\n", labelSelector, err)
	}
	resourceModel, err := discoverer.DiscoverResourcesForGatewayClass(Filter{Labels: selector})
	if err != nil {
		t.Fatalf("Failed to construct resourceModel: %v", resourceModel)
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

	params := utils.MustParamsForTest(t, common.MustClientsForTest(t, objects...))
	discoverer := Discoverer{
		K8sClients:    params.K8sClients,
		PolicyManager: params.PolicyManager,
	}
	labelSelector := "env=internal"
	selector, err := labels.Parse(labelSelector)
	if err != nil {
		t.Errorf("Unable to find resources that match the label selector \"%s\": %v\n", labelSelector, err)
	}
	resourceModel, err := discoverer.DiscoverResourcesForGateway(Filter{Labels: selector})
	if err != nil {
		t.Fatalf("Failed to construct resourceModel: %v", resourceModel)
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

	params := utils.MustParamsForTest(t, common.MustClientsForTest(t, objects...))
	discoverer := Discoverer{
		K8sClients:    params.K8sClients,
		PolicyManager: params.PolicyManager,
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

	params := utils.MustParamsForTest(t, common.MustClientsForTest(t, objects...))
	discoverer := Discoverer{
		K8sClients:    params.K8sClients,
		PolicyManager: params.PolicyManager,
	}
	labelSelector := "env=internal"
	selector, err := labels.Parse(labelSelector)
	if err != nil {
		t.Errorf("Unable to find resources that match the label selector \"%s\": %v\n", labelSelector, err)
	}
	resourceModel, err := discoverer.DiscoverResourcesForNamespace(Filter{Labels: selector})
	if err != nil {
		t.Fatalf("Failed to construct resourceModel: %v", resourceModel)
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
