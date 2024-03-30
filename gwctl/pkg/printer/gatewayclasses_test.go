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
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"

	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	testingclock "k8s.io/utils/clock/testing"

	gatewayv1 "sigs.k8s.io/gateway-api/apis/v1"
	gatewayv1alpha2 "sigs.k8s.io/gateway-api/apis/v1alpha2"
	"sigs.k8s.io/gateway-api/gwctl/pkg/common"
	"sigs.k8s.io/gateway-api/gwctl/pkg/resourcediscovery"
	"sigs.k8s.io/gateway-api/gwctl/pkg/utils"
)

func TestGatewayClassesPrinter_Print(t *testing.T) {
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
		Out:   params.Out,
		Clock: fakeClock,
	}
	gcp.Print(resourceModel)

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
ControllerName: example.net/gateway-controller
Description: random
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
					},
					Spec: gatewayv1.GatewayClassSpec{
						ControllerName: "example.net/gateway-controller",
					},
				},
			},
			want: `
Name: foo-gatewayclass
ControllerName: example.net/gateway-controller
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
				Out:   params.Out,
				Clock: fakeClock,
			}
			gcp.PrintDescribeView(resourceModel)

			got := params.Out.(*bytes.Buffer).String()
			if diff := cmp.Diff(common.YamlString(tc.want), common.YamlString(got), common.YamlStringTransformer); diff != "" {
				t.Errorf("Unexpected diff\ngot=\n%v\nwant=\n%v\ndiff (-want +got)=\n%v", got, tc.want, diff)
			}
		})
	}
}

// TestGatewayClassesPrinter_LabelSelector Tests label selector filtering for GatewayClasses in 'get' command.
func TestGatewayClassesPrinter_LabelSelector(t *testing.T) {
	fakeClock := testingclock.NewFakeClock(time.Now())

	gatewayClass := func(name string, labels map[string]string) *gatewayv1.GatewayClass {
		return &gatewayv1.GatewayClass{
			ObjectMeta: metav1.ObjectMeta{
				Name:   name,
				Labels: labels,
				CreationTimestamp: metav1.Time{
					Time: fakeClock.Now().Add(-365 * 24 * time.Hour),
				}},
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
	discoverer := resourcediscovery.Discoverer{
		K8sClients:    params.K8sClients,
		PolicyManager: params.PolicyManager,
	}
	labelSelector := "env=internal"
	selector, err := labels.Parse(labelSelector)
	if err != nil {
		t.Errorf("Unable to find resources that match the label selector \"%s\": %v\n", labelSelector, err)
	}
	resourceModel, err := discoverer.DiscoverResourcesForGatewayClass(resourcediscovery.Filter{Labels: selector})
	if err != nil {
		t.Fatalf("Failed to construct resourceModel: %v", resourceModel)
	}

	gcp := &GatewayClassesPrinter{
		Out:   params.Out,
		Clock: fakeClock,
	}
	gcp.Print(resourceModel)

	got := params.Out.(*bytes.Buffer).String()
	want := `
NAME                            CONTROLLER                                 ACCEPTED  AGE
foo-com-internal-gateway-class  foo-com-internal-gateway-class/controller  True      365d
`
	if diff := cmp.Diff(common.YamlString(want), common.YamlString(got), common.YamlStringTransformer); diff != "" {
		t.Errorf("Unexpected diff\ngot=\n%v\nwant=\n%v\ndiff (-want +got)=\n%v", got, want, diff)
	}
}
