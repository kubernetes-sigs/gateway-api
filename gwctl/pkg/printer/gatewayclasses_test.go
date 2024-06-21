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

func TestGatewayClassesPrinter_PrintTable(t *testing.T) {
	fakeClock := testingclock.NewFakeClock(time.Now())
	objects := []runtime.Object{
		&gatewayv1.GatewayClass{
			ObjectMeta: metav1.ObjectMeta{
				Name: "bar-com-internal-gateway-class",
				CreationTimestamp: metav1.Time{
					Time: fakeClock.Now().Add(-365 * 24 * time.Hour),
				},
			},
			Spec: gatewayv1.GatewayClassSpec{
				ControllerName: "bar.baz/internal-gateway-class",
			},
			Status: gatewayv1.GatewayClassStatus{
				Conditions: []metav1.Condition{
					{
						Type:   "Accepted",
						Status: "True",
					},
				},
			},
		},
		&gatewayv1.GatewayClass{
			ObjectMeta: metav1.ObjectMeta{
				Name: "foo-com-external-gateway-class",
				CreationTimestamp: metav1.Time{
					Time: fakeClock.Now().Add(-100 * 24 * time.Hour),
				},
			},
			Spec: gatewayv1.GatewayClassSpec{
				ControllerName: "foo.com/external-gateway-class",
			},
			Status: gatewayv1.GatewayClassStatus{
				Conditions: []metav1.Condition{
					{
						Type:   "Accepted",
						Status: "False",
					},
				},
			},
		},
		&gatewayv1.GatewayClass{
			ObjectMeta: metav1.ObjectMeta{
				Name: "foo-com-internal-gateway-class",
				CreationTimestamp: metav1.Time{
					Time: fakeClock.Now().Add(-24 * time.Minute),
				},
			},
			Spec: gatewayv1.GatewayClassSpec{
				ControllerName: "foo.com/internal-gateway-class",
			},
			Status: gatewayv1.GatewayClassStatus{
				Conditions: []metav1.Condition{
					{
						Type:   "Accepted",
						Status: "Unknown",
					},
				},
			},
		},
		&gatewayv1.Gateway{
			ObjectMeta: metav1.ObjectMeta{
				Name: "bar-gateway",
				CreationTimestamp: metav1.Time{
					Time: fakeClock.Now().Add(-3 * time.Second),
				},
			},
			Spec: gatewayv1.GatewaySpec{
				GatewayClassName: "bar-com-internal-gateway-class",
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

	k8sClients := common.MustClientsForTest(t, objects...)
	policyManager := utils.MustPolicyManagerForTest(t, k8sClients)
	buff := &bytes.Buffer{}
	discoverer := resourcediscovery.Discoverer{
		K8sClients:    k8sClients,
		PolicyManager: policyManager,
	}
	resourceModel, err := discoverer.DiscoverResourcesForGatewayClass(resourcediscovery.Filter{})
	if err != nil {
		t.Fatalf("Failed to construct resourceModel: %v", err)
	}

	gcp := &GatewayClassesPrinter{
		Writer: buff,
		Clock:  fakeClock,
	}
	gcp.PrintTable(resourceModel, false)

	got := buff.String()
	want := `
NAME                            CONTROLLER                      ACCEPTED  AGE
bar-com-internal-gateway-class  bar.baz/internal-gateway-class  True      365d
foo-com-external-gateway-class  foo.com/external-gateway-class  False     100d
foo-com-internal-gateway-class  foo.com/internal-gateway-class  Unknown   24m
`
	if diff := cmp.Diff(common.YamlString(want), common.YamlString(got), common.YamlStringTransformer); diff != "" {
		t.Errorf("Unexpected diff\ngot=\n%v\nwant=\n%v\ndiff (-want +got)=\n%v", got, want, diff)
	}
	buff.Reset()
	gcp2 := &GatewayClassesPrinter{
		Writer: buff,
		Clock:  fakeClock,
	}
	gcp2.PrintTable(resourceModel, true)

	got2 := buff.String()
	want2 := `
NAME                            CONTROLLER                      ACCEPTED  AGE   GATEWAYS
bar-com-internal-gateway-class  bar.baz/internal-gateway-class  True      365d  1
foo-com-external-gateway-class  foo.com/external-gateway-class  False     100d  0
foo-com-internal-gateway-class  foo.com/internal-gateway-class  Unknown   24m   0
`
	if diff := cmp.Diff(common.YamlString(want2), common.YamlString(got2), common.YamlStringTransformer); diff != "" {
		t.Errorf("Unexpected diff\ngot=\n%v\nwant=\n%v\ndiff (-want +got)=\n%v", got2, want2, diff)
	}
}

func TestGatewayClassesPrinter_PrintDescribeView(t *testing.T) {
	fakeClock := testingclock.NewFakeClock(time.Now())

	testcases := []struct {
		name    string
		objects []runtime.Object
		want    string
	}{
		{
			name: "GatewayClass with description and policy",
			objects: []runtime.Object{
				&gatewayv1.GatewayClass{
					TypeMeta: metav1.TypeMeta{
						APIVersion: gatewayv1.GroupVersion.String(),
						Kind:       "GatewayClass",
					},
					ObjectMeta: metav1.ObjectMeta{
						Name: "foo-gatewayclass",
						UID:  "00000000-0000-0000-0000-000000000001",
					},
					Spec: gatewayv1.GatewayClassSpec{
						ControllerName: "example.net/gateway-controller",
						Description:    common.PtrTo("random"),
					},
				},
				&apiextensionsv1.CustomResourceDefinition{
					ObjectMeta: metav1.ObjectMeta{
						Name: "healthcheckpolicies.foo.com",
						Labels: map[string]string{
							gatewayv1alpha2.PolicyLabelKey: "true",
						},
					},
					Spec: apiextensionsv1.CustomResourceDefinitionSpec{
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
							"name": "policy-name",
						},
						"spec": map[string]interface{}{
							"targetRef": map[string]interface{}{
								"group": "gateway.networking.k8s.io",
								"kind":  "GatewayClass",
								"name":  "foo-gatewayclass",
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
						Kind: "GatewayClass",
						Name: "foo-gatewayclass",
						UID:  "00000000-0000-0000-0000-000000000001",
					},
					Message: "some random message",
				},
			},
			want: `
Name: foo-gatewayclass
Labels: null
Annotations: null
APIVersion: gateway.networking.k8s.io/v1
Kind: GatewayClass
Metadata:
  creationTimestamp: null
  resourceVersion: "999"
  uid: 00000000-0000-0000-0000-000000000001
Spec:
  controllerName: example.net/gateway-controller
  description: random
Status: {}
DirectlyAttachedPolicies:
  Type                       Name
  ----                       ----
  HealthCheckPolicy.foo.com  policy-name
Events:
  Type    Reason  Age      From                   Message
  ----    ------  ---      ----                   -------
  Normal  SYNC    Unknown  my-gateway-controller  some random message
`,
		},
		{
			name: "GatewayClass with no description",
			objects: []runtime.Object{
				&gatewayv1.GatewayClass{
					TypeMeta: metav1.TypeMeta{
						APIVersion: gatewayv1.GroupVersion.String(),
						Kind:       "GatewayClass",
					},
					ObjectMeta: metav1.ObjectMeta{
						Name: "foo-gatewayclass",
						Labels: map[string]string{
							"foo": "bar",
						},
					},
					Spec: gatewayv1.GatewayClassSpec{
						ControllerName: "example.net/gateway-controller",
					},
				},
			},
			want: `
Name: foo-gatewayclass
Labels:
  foo: bar
Annotations: null
APIVersion: gateway.networking.k8s.io/v1
Kind: GatewayClass
Metadata:
  creationTimestamp: null
  resourceVersion: "999"
Spec:
  controllerName: example.net/gateway-controller
Status: {}
DirectlyAttachedPolicies: <none>
Events: <none>
`,
		},
	}

	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			k8sClients := common.MustClientsForTest(t, tc.objects...)
			policyManager := utils.MustPolicyManagerForTest(t, k8sClients)
			buff := &bytes.Buffer{}
			discoverer := resourcediscovery.Discoverer{
				K8sClients:    k8sClients,
				PolicyManager: policyManager,
			}
			resourceModel, err := discoverer.DiscoverResourcesForGatewayClass(resourcediscovery.Filter{})
			if err != nil {
				t.Fatalf("Failed to construct resourceModel: %v", err)
			}

			gcp := &GatewayClassesPrinter{
				Writer:       buff,
				Clock:        fakeClock,
				EventFetcher: discoverer,
			}
			gcp.PrintDescribeView(resourceModel)

			got := buff.String()
			if diff := cmp.Diff(common.YamlString(tc.want), common.YamlString(got), common.YamlStringTransformer); diff != "" {
				t.Errorf("Unexpected diff\ngot=\n%v\nwant=\n%v\ndiff (-want +got)=\n%v", got, tc.want, diff)
			}
		})
	}
}

// TestGatewayClassesPrinter_PrintJsonYaml tests the -o json/yaml output of the `get` subcommand
func TestGatewayClassesPrinter_PrintJsonYaml(t *testing.T) {
	fakeClock := testingclock.NewFakeClock(time.Now())
	creationTime := fakeClock.Now().Add(-365 * 24 * time.Hour).UTC() // UTC being necessary for consistently handling the time while marshaling/unmarshaling its JSON

	gtwName := "foo-com-internal-gateway-class"
	gtwApplyConfig := apisv1beta1.GatewayClass(gtwName)

	gtwObject := &gatewayv1.GatewayClass{
		TypeMeta: metav1.TypeMeta{
			APIVersion: *gtwApplyConfig.APIVersion,
			Kind:       *gtwApplyConfig.Kind,
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:   "foo-com-internal-gateway-class",
			Labels: map[string]string{"app": "foo", "env": "internal"},
			CreationTimestamp: metav1.Time{
				Time: creationTime,
			},
		},
		Spec: gatewayv1.GatewayClassSpec{
			ControllerName: gatewayv1.GatewayController(gtwName + "/controller"),
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
	gtwObject.APIVersion = *gtwApplyConfig.APIVersion
	gtwObject.Kind = *gtwApplyConfig.Kind

	k8sClients := common.MustClientsForTest(t, gtwObject)
	policyManager := utils.MustPolicyManagerForTest(t, k8sClients)
	buff := &bytes.Buffer{}
	discoverer := resourcediscovery.Discoverer{
		K8sClients:    k8sClients,
		PolicyManager: policyManager,
	}
	resourceModel, err := discoverer.DiscoverResourcesForGatewayClass(resourcediscovery.Filter{})
	if err != nil {
		t.Fatalf("Failed to construct resourceModel: %v", resourceModel)
	}

	gcp := &GatewayClassesPrinter{
		Writer: buff,
		Clock:  fakeClock,
	}
	Print(gcp, resourceModel, utils.OutputFormatJSON)

	gotJSON := common.JSONString(buff.String())
	wantJSON := common.JSONString(fmt.Sprintf(`
        {
          "apiVersion": "v1",
          "items": [
            {
              "apiVersion": "gateway.networking.k8s.io/v1beta1",
              "kind": "GatewayClass",
              "metadata": {
                "creationTimestamp": "%s",
                "labels": {
                  "app": "foo",
                  "env": "internal"
                },
                "name": "foo-com-internal-gateway-class",
                "resourceVersion": "999"
              },
              "spec": {
                "controllerName": "foo-com-internal-gateway-class/controller"
              },
              "status": {
                "conditions": [
                  {
                    "lastTransitionTime": null,
                    "message": "",
                    "reason": "",
                    "status": "True",
                    "type": "Accepted"
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

	gcp.Writer = &bytes.Buffer{}

	Print(gcp, resourceModel, utils.OutputFormatYAML)

	gotYaml := common.YamlString(gcp.Writer.(*bytes.Buffer).String())
	wantYaml := common.YamlString(fmt.Sprintf(`
apiVersion: v1
items:
- apiVersion: gateway.networking.k8s.io/v1beta1
  kind: GatewayClass
  metadata:
    creationTimestamp: "%s"
    labels:
      app: foo
      env: internal
    name: foo-com-internal-gateway-class
    resourceVersion: "999"
  spec:
    controllerName: foo-com-internal-gateway-class/controller
  status:
    conditions:
    - lastTransitionTime: null
      message: ""
      reason: ""
      status: "True"
      type: Accepted
kind: List`, creationTime.Format(time.RFC3339)))
	if diff := cmp.Diff(wantYaml, gotYaml, common.YamlStringTransformer); diff != "" {
		t.Errorf("Unexpected diff\ngot=\n%v\nwant=\n%v\ndiff (-want +got)=\n%v", gotYaml, wantYaml, diff)
	}
}
