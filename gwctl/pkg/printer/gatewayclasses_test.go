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
	}

	params := utils.MustParamsForTest(t, common.MustClientsForTest(t, objects...))
	discoverer := resourcediscovery.Discoverer{
		K8sClients:    params.K8sClients,
		PolicyManager: params.PolicyManager,
	}
	resourceModel, err := discoverer.DiscoverResourcesForGatewayClass(resourcediscovery.Filter{})
	if err != nil {
		t.Fatalf("Failed to construct resourceModel: %v", resourceModel)
	}

	gcp := &GatewayClassesPrinter{
		Writer: params.Out,
		Clock:  fakeClock,
	}
	Print(gcp, resourceModel, utils.OutputFormatTable)

	got := params.Out.(*bytes.Buffer).String()
	want := `
NAME                            CONTROLLER                      ACCEPTED  AGE
bar-com-internal-gateway-class  bar.baz/internal-gateway-class  True      365d
foo-com-external-gateway-class  foo.com/external-gateway-class  False     100d
foo-com-internal-gateway-class  foo.com/internal-gateway-class  Unknown   24m
`
	if diff := cmp.Diff(common.YamlString(want), common.YamlString(got), common.YamlStringTransformer); diff != "" {
		t.Errorf("Unexpected diff\ngot=\n%v\nwant=\n%v\ndiff (-want +got)=\n%v", got, want, diff)
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
					ObjectMeta: metav1.ObjectMeta{
						Name: "foo-gatewayclass",
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
			},
			want: `
Name: foo-gatewayclass
Labels: null
Annotations: null
Metadata:
  creationTimestamp: null
  resourceVersion: "999"
ControllerName: example.net/gateway-controller
Description: random
Status: {}
DirectlyAttachedPolicies:
- Group: foo.com
  Kind: HealthCheckPolicy
  Name: policy-name
`,
		},
		{
			name: "GatewayClass with no description",
			objects: []runtime.Object{
				&gatewayv1.GatewayClass{
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
Metadata:
  creationTimestamp: null
  resourceVersion: "999"
ControllerName: example.net/gateway-controller
Status: {}
`,
		},
	}

	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			params := utils.MustParamsForTest(t, common.MustClientsForTest(t, tc.objects...))
			discoverer := resourcediscovery.Discoverer{
				K8sClients:    params.K8sClients,
				PolicyManager: params.PolicyManager,
			}
			resourceModel, err := discoverer.DiscoverResourcesForGatewayClass(resourcediscovery.Filter{})
			if err != nil {
				t.Fatalf("Failed to construct resourceModel: %v", resourceModel)
			}

			gcp := &GatewayClassesPrinter{
				Writer: params.Out,
				Clock:  fakeClock,
			}
			gcp.PrintDescribeView(resourceModel)

			got := params.Out.(*bytes.Buffer).String()
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

	params := utils.MustParamsForTest(t, common.MustClientsForTest(t, gtwObject))
	discoverer := resourcediscovery.Discoverer{
		K8sClients:    params.K8sClients,
		PolicyManager: params.PolicyManager,
	}
	resourceModel, err := discoverer.DiscoverResourcesForGatewayClass(resourcediscovery.Filter{})
	if err != nil {
		t.Fatalf("Failed to construct resourceModel: %v", resourceModel)
	}

	gcp := &GatewayClassesPrinter{
		Writer: params.Out,
		Clock:  fakeClock,
	}
	Print(gcp, resourceModel, utils.OutputFormatJSON)

	gotJSON := common.JSONString(params.Out.(*bytes.Buffer).String())
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
