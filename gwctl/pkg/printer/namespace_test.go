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

	"github.com/google/go-cmp/cmp"
	corev1 "k8s.io/api/core/v1"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	gatewayv1alpha2 "sigs.k8s.io/gateway-api/apis/v1alpha2"
	"sigs.k8s.io/gateway-api/gwctl/pkg/utils"
	"sigs.k8s.io/gateway-api/gwctl/pkg/common"
	"sigs.k8s.io/gateway-api/gwctl/pkg/resourcediscovery"
)

func TestNamespacePrinter_PrintDescribeView(t *testing.T) {
	objects := []runtime.Object{
		// Defining a Namespace called development
		&corev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name: "development",
				Labels: map[string]string{
					"type": "test-namespace",
				},
				Annotations: map[string]string{
					"test-annotation": "development-annotation",
				},
			},
			Status: corev1.NamespaceStatus{
				Phase: corev1.NamespaceActive,
			},
		},

		// CRD and definition for HealthCheckPolicy attached to development namespace
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
						"kind": "Namespace",
						"name": "development",
					},
				},
			},
		},

		// Defining a Namespace called production
		&corev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name: "production",
				Labels: map[string]string{
					"type": "production-namespace",
				},
			},
			Status: corev1.NamespaceStatus{
				Phase: corev1.NamespaceActive,
			},
		},
		// CRD and definition for TimeoutPolicy attached to default namespace
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
						"name": "production",
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
	resourceModel, err := discoverer.DiscoverResourcesForNamespace(resourcediscovery.Filter{})
	if err != nil {
		t.Fatalf("Failed to construct resourceModel: %v", resourceModel)
	}

	nsp := &NamespacesPrinter{
		Out: params.Out,
	}
	nsp.PrintDescribeView(resourceModel)

	got := params.Out.(*bytes.Buffer).String()
	want := `
Name: development
Annotations:
  test-annotation: development-annotation
Labels:
  type: test-namespace
Status: Active
DirectlyAttachedPolicies:
- Group: foo.com
  Kind: HealthCheckPolicy
  Name: health-check-gatewayclass


Name: production
Labels:
  type: production-namespace
Status: Active
DirectlyAttachedPolicies:
- Group: bar.com
  Kind: TimeoutPolicy
  Name: timeout-policy-namespace
`
	if diff := cmp.Diff(common.YamlString(want), common.YamlString(got), common.YamlStringTransformer); diff != "" {
		t.Errorf("Unexpected diff\ngot=\n%v\nwant=\n%v\ndiff (-want +got)=\n%v", got, want, diff)
	}

}
