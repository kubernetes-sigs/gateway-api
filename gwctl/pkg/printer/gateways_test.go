/*
Copyright 2023 The Kubernetes Authors.

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
	"fmt"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"

	corev1 "k8s.io/api/core/v1"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	testingclock "k8s.io/utils/clock/testing"

	apisv1beta1 "sigs.k8s.io/gateway-api/apis/applyconfiguration/apis/v1beta1"
	gatewayv1 "sigs.k8s.io/gateway-api/apis/v1"
	gatewayv1alpha2 "sigs.k8s.io/gateway-api/apis/v1alpha2"
	"sigs.k8s.io/gateway-api/gwctl/pkg/common"
	"sigs.k8s.io/gateway-api/gwctl/pkg/resourcediscovery"
	"sigs.k8s.io/gateway-api/gwctl/pkg/utils"
)

func TestGatewaysPrinter_PrintTable(t *testing.T) {
	fakeClock := testingclock.NewFakeClock(time.Now())
	objects := []runtime.Object{
		&gatewayv1.GatewayClass{
			ObjectMeta: metav1.ObjectMeta{
				Name: "internal-class",
			},
			Spec: gatewayv1.GatewayClassSpec{
				ControllerName: "example.net/gateway-controller",
				Description:    common.PtrTo("random"),
			},
		},
		&gatewayv1.GatewayClass{
			ObjectMeta: metav1.ObjectMeta{
				Name: "external-class",
			},
			Spec: gatewayv1.GatewayClassSpec{
				ControllerName: "example.net/gateway-controller",
				Description:    common.PtrTo("random"),
			},
		},
		&gatewayv1.GatewayClass{
			ObjectMeta: metav1.ObjectMeta{
				Name: "regional-internal-class",
			},
			Spec: gatewayv1.GatewayClassSpec{
				ControllerName: "example.net/gateway-controller",
				Description:    common.PtrTo("random"),
			},
		},
		&gatewayv1.Gateway{
			ObjectMeta: metav1.ObjectMeta{
				Name: "abc-gateway-12345",
				CreationTimestamp: metav1.Time{
					Time: fakeClock.Now().Add(-20 * 24 * time.Hour),
				},
			},
			Spec: gatewayv1.GatewaySpec{
				GatewayClassName: "internal-class",
				Listeners: []gatewayv1.Listener{
					{
						Name:     gatewayv1.SectionName("https-443"),
						Protocol: gatewayv1.HTTPSProtocolType,
						Port:     gatewayv1.PortNumber(443),
					},
					{
						Name:     gatewayv1.SectionName("http-8080"),
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
		},
		&gatewayv1.Gateway{
			ObjectMeta: metav1.ObjectMeta{
				Name: "demo-gateway-2",
				CreationTimestamp: metav1.Time{
					Time: fakeClock.Now().Add(-5 * 24 * time.Hour),
				},
			},
			Spec: gatewayv1.GatewaySpec{
				GatewayClassName: "external-class",
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
		&gatewayv1.Gateway{
			ObjectMeta: metav1.ObjectMeta{
				Name: "random-gateway",
				CreationTimestamp: metav1.Time{
					Time: fakeClock.Now().Add(-3 * time.Second),
				},
			},
			Spec: gatewayv1.GatewaySpec{
				GatewayClassName: "regional-internal-class",
				Listeners: []gatewayv1.Listener{
					{
						Name:     gatewayv1.SectionName("http-8443"),
						Protocol: gatewayv1.HTTPProtocolType,
						Port:     gatewayv1.PortNumber(8443),
					},
				},
			},
			Status: gatewayv1.GatewayStatus{
				Addresses: []gatewayv1.GatewayStatusAddress{
					{
						Value: "10.11.12.13",
					},
				},
				Conditions: []metav1.Condition{
					{
						Type:   "Programmed",
						Status: "Unknown",
					},
				},
			},
		},
	}

	params := utils.MustParamsForTest(t, common.MustClientsForTest(t, objects...))
	discoverer := resourcediscovery.Discoverer{
		K8sClients:    params.K8sClients,
		PolicyManager: params.PolicyManager,
	}
	resourceModel, err := discoverer.DiscoverResourcesForGateway(resourcediscovery.Filter{})
	if err != nil {
		t.Fatalf("Failed to construct resourceModel: %v", err)
	}

	gp := &GatewaysPrinter{
		Writer: params.Out,
		Clock:  fakeClock,
	}
	gp.PrintTable(resourceModel)

	got := params.Out.(*bytes.Buffer).String()
	want := `
NAME               CLASS                    ADDRESSES                   PORTS     PROGRAMMED  AGE
abc-gateway-12345  internal-class           192.168.100.5               443,8080  False       20d
demo-gateway-2     external-class           10.0.0.1,10.0.0.2 + 1 more  80        True        5d
random-gateway     regional-internal-class  10.11.12.13                 8443      Unknown     3s
`

	if diff := cmp.Diff(common.YamlString(want), common.YamlString(got), common.YamlStringTransformer); diff != "" {
		t.Errorf("Unexpected diff\ngot=\n%v\nwant=\n%v\ndiff (-want +got)=\n%v", got, want, diff)
	}
}

func TestGatewaysPrinter_PrintDescribeView(t *testing.T) {
	fakeClock := testingclock.NewFakeClock(time.Now())
	objects := []runtime.Object{
		&gatewayv1.GatewayClass{
			ObjectMeta: metav1.ObjectMeta{
				Name: "foo-gatewayclass",
			},
			Spec: gatewayv1.GatewayClassSpec{
				ControllerName: "example.net/gateway-controller",
				Description:    common.PtrTo("random"),
			},
		},

		&gatewayv1.Gateway{
			ObjectMeta: metav1.ObjectMeta{
				Name: "foo-gateway",
				UID:  "00000000-0000-0000-0000-000000000001",
			},
			Spec: gatewayv1.GatewaySpec{
				GatewayClassName: "foo-gatewayclass",
			},
		},

		&gatewayv1.HTTPRoute{
			TypeMeta: metav1.TypeMeta{
				Kind: "HTTPRoute",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name: "foo-httproute",
			},
			Spec: gatewayv1.HTTPRouteSpec{
				CommonRouteSpec: gatewayv1.CommonRouteSpec{
					ParentRefs: []gatewayv1.ParentReference{{
						Kind:  common.PtrTo(gatewayv1.Kind("Gateway")),
						Group: common.PtrTo(gatewayv1.Group("gateway.networking.k8s.io")),
						Name:  "foo-gateway",
					}},
				},
			},
		},

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
					"name": "health-check-gatewayclass",
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
						"group": "gateway.networking.k8s.io",
						"kind":  "GatewayClass",
						"name":  "foo-gatewayclass",
					},
				},
			},
		},
		&unstructured.Unstructured{
			Object: map[string]interface{}{
				"apiVersion": "foo.com/v1",
				"kind":       "HealthCheckPolicy",
				"metadata": map[string]interface{}{
					"name": "health-check-gateway",
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
						"group":     "gateway.networking.k8s.io",
						"kind":      "Gateway",
						"name":      "foo-gateway",
						"namespace": "default",
					},
				},
			},
		},

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
					"name": "timeout-policy-namespace",
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

		&corev1.Event{
			ObjectMeta: metav1.ObjectMeta{
				Name: "event-1",
			},
			Type:   corev1.EventTypeNormal,
			Reason: "SYNC",
			Source: corev1.EventSource{
				Component: "my-gateway-controller",
			},
			InvolvedObject: corev1.ObjectReference{
				Kind: "Gateway",
				Name: "foo-gateway",
				UID:  "00000000-0000-0000-0000-000000000001",
			},
			Message: "some random message",
		},
	}

	params := utils.MustParamsForTest(t, common.MustClientsForTest(t, objects...))
	discoverer := resourcediscovery.Discoverer{
		K8sClients:    params.K8sClients,
		PolicyManager: params.PolicyManager,
	}
	resourceModel, err := discoverer.DiscoverResourcesForGateway(resourcediscovery.Filter{})
	if err != nil {
		t.Fatalf("Failed to construct resourceModel: %v", err)
	}

	gp := &GatewaysPrinter{
		Writer: params.Out,
		Clock:  fakeClock,
	}
	gp.PrintDescribeView(resourceModel)

	got := params.Out.(*bytes.Buffer).String()
	want := `
Name: foo-gateway
Namespace: ""
Labels: null
Annotations: null
APIVersion: ""
Kind: ""
Metadata:
  creationTimestamp: null
  resourceVersion: "999"
  uid: 00000000-0000-0000-0000-000000000001
Spec:
  gatewayClassName: foo-gatewayclass
  listeners: null
Status: {}
AttachedRoutes:
  Kind       Name
  ----       ----
  HTTPRoute  /foo-httproute
DirectlyAttachedPolicies:
  Type                       Name
  ----                       ----
  HealthCheckPolicy.foo.com  /health-check-gateway
EffectivePolicies:
  HealthCheckPolicy.foo.com:
    key1: value-parent-1
    key2: value-child-2
    key3: value-parent-3
    key4: value-parent-4
    key5: value-parent-5
  TimeoutPolicy.bar.com:
    condition: path=/abc
    seconds: 30
Events:
  Type    Reason  Age      From                   Message
  ----    ------  ---      ----                   -------
  Normal  SYNC    Unknown  my-gateway-controller  some random message
`
	if diff := cmp.Diff(common.YamlString(want), common.YamlString(got), common.YamlStringTransformer); diff != "" {
		t.Errorf("Unexpected diff\ngot=\n%v\nwant=\n%v\ndiff (-want +got)=\n%v", got, want, diff)
	}
}

// TestGatewaysPrinter_PrintJsonYaml tests the -o json/yaml output of the `get` subcommand
func TestGatewaysPrinter_PrintJsonYaml(t *testing.T) {
	fakeClock := testingclock.NewFakeClock(time.Now())
	creationTime := fakeClock.Now().Add(-5 * 24 * time.Hour).UTC() // UTC being necessary for consistently handling the time while marshaling/unmarshaling its JSON
	gcName := "gateway-1"
	gcApplyConfig := apisv1beta1.Gateway(gcName, "")

	gcObject := &gatewayv1.Gateway{
		TypeMeta: metav1.TypeMeta{
			APIVersion: *gcApplyConfig.APIVersion,
			Kind:       *gcApplyConfig.Kind,
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:   gcName,
			Labels: map[string]string{"app": "foo", "env": "internal"},
			CreationTimestamp: metav1.Time{
				Time: creationTime,
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
		gcObject,
	}

	params := utils.MustParamsForTest(t, common.MustClientsForTest(t, objects...))
	discoverer := resourcediscovery.Discoverer{
		K8sClients:    params.K8sClients,
		PolicyManager: params.PolicyManager,
	}
	resourceModel, err := discoverer.DiscoverResourcesForGateway(resourcediscovery.Filter{})
	if err != nil {
		t.Fatalf("Failed to construct resourceModel: %v", resourceModel)
	}

	gp := &GatewaysPrinter{
		Writer: params.Out,
		Clock:  fakeClock,
	}
	Print(gp, resourceModel, utils.OutputFormatJSON)

	gotJSON := common.JSONString(params.Out.(*bytes.Buffer).String())
	wantJSON := common.JSONString(fmt.Sprintf(`
        {
          "apiVersion": "v1",
          "items": [
            {
              "apiVersion": "gateway.networking.k8s.io/v1beta1",
              "kind": "Gateway",
              "metadata": {
                "creationTimestamp": "%s",
                "labels": {
                  "app": "foo",
                  "env": "internal"
                },
                "name": "gateway-1",
                "resourceVersion": "999"
              },
              "spec": {
                "gatewayClassName": "gatewayclass-1",
                "listeners": [
                  {
                    "name": "http-8080",
                    "port": 8080,
                    "protocol": "HTTP"
                  }
                ]
              },
              "status": {
                "addresses": [
                  {
                    "value": "192.168.100.5"
                  }
                ],
                "conditions": [
                  {
                    "lastTransitionTime": null,
                    "message": "",
                    "reason": "",
                    "status": "False",
                    "type": "Programmed"
                  }
                ]
              }
            }
          ],
          "kind": "List"
        }`, creationTime.Format(time.RFC3339)))
	diff, err := wantJSON.CmpDiff(gotJSON)
	if err != nil {
		t.Fatalf("Failed to compare the json diffs: %v", diff)
	}
	if diff != "" {
		t.Errorf("Unexpected diff\ngot=\n%v\nwant=\n%v\ndiff (-want +got)=\n%v", gotJSON, wantJSON, diff)
	}

	gp.Writer = &bytes.Buffer{}

	Print(gp, resourceModel, utils.OutputFormatYAML)

	gotYaml := common.YamlString(gp.Writer.(*bytes.Buffer).String())
	wantYaml := common.YamlString(fmt.Sprintf(`
apiVersion: v1
items:
- apiVersion: gateway.networking.k8s.io/v1beta1
  kind: Gateway
  metadata:
    creationTimestamp: "%s"
    labels:
      app: foo
      env: internal
    name: gateway-1
    resourceVersion: "999"
  spec:
    gatewayClassName: gatewayclass-1
    listeners:
    - name: http-8080
      port: 8080
      protocol: HTTP
  status:
    addresses:
    - value: 192.168.100.5
    conditions:
    - lastTransitionTime: null
      message: ""
      reason: ""
      status: "False"
      type: Programmed
kind: List`, creationTime.Format(time.RFC3339)))
	if diff := cmp.Diff(wantYaml, gotYaml, common.YamlStringTransformer); diff != "" {
		t.Errorf("Unexpected diff\ngot=\n%v\nwant=\n%v\ndiff (-want +got)=\n%v", gotYaml, wantYaml, diff)
	}
}

// TestGatewaysPrinter_PrintYaml tests the -o yaml output of the `get` subcommand
func TestGatewaysPrinter_PrintYaml(t *testing.T) {
	fakeClock := testingclock.NewFakeClock(time.Now())
	creationTime := fakeClock.Now().Add(-5 * 24 * time.Hour).UTC() // UTC being necessary for consistently handling the time while marshaling/unmarshaling its JSON
	gcName := "gateway-1"
	gcApplyConfig := apisv1beta1.Gateway(gcName, "")

	gcObject := &gatewayv1.Gateway{
		TypeMeta: metav1.TypeMeta{
			APIVersion: *gcApplyConfig.APIVersion,
			Kind:       *gcApplyConfig.Kind,
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:   gcName,
			Labels: map[string]string{"app": "foo", "env": "internal"},
			CreationTimestamp: metav1.Time{
				Time: creationTime,
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
		gcObject,
	}

	params := utils.MustParamsForTest(t, common.MustClientsForTest(t, objects...))
	discoverer := resourcediscovery.Discoverer{
		K8sClients:    params.K8sClients,
		PolicyManager: params.PolicyManager,
	}
	resourceModel, err := discoverer.DiscoverResourcesForGateway(resourcediscovery.Filter{})
	if err != nil {
		t.Fatalf("Failed to construct resourceModel: %v", resourceModel)
	}

	gp := &GatewaysPrinter{
		Writer: params.Out,
		Clock:  fakeClock,
	}
	Print(gp, resourceModel, utils.OutputFormatYAML)

	got := common.YamlString(params.Out.(*bytes.Buffer).String())
	want := common.YamlString(fmt.Sprintf(`
apiVersion: v1
items:
- apiVersion: gateway.networking.k8s.io/v1beta1
  kind: Gateway
  metadata:
    creationTimestamp: "%s"
    labels:
      app: foo
      env: internal
    name: gateway-1
    resourceVersion: "999"
  spec:
    gatewayClassName: gatewayclass-1
    listeners:
    - name: http-8080
      port: 8080
      protocol: HTTP
  status:
    addresses:
    - value: 192.168.100.5
    conditions:
    - lastTransitionTime: null
      message: ""
      reason: ""
      status: "False"
      type: Programmed
kind: List`, creationTime.Format(time.RFC3339)))
	if diff := cmp.Diff(want, got, common.YamlStringTransformer); diff != "" {
		t.Errorf("Unexpected diff\ngot=\n%v\nwant=\n%v\ndiff (-want +got)=\n%v", got, want, diff)
	}
}
