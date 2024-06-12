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

	gatewayv1alpha2 "sigs.k8s.io/gateway-api/apis/v1alpha2"
	"sigs.k8s.io/gateway-api/gwctl/pkg/common"
	"sigs.k8s.io/gateway-api/gwctl/pkg/resourcediscovery"
	"sigs.k8s.io/gateway-api/gwctl/pkg/utils"
)

func TestNamespacePrinter_PrintTable(t *testing.T) {
	fakeClock := testingclock.NewFakeClock(time.Now())
	objects := []runtime.Object{
		&corev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name: "default",
				CreationTimestamp: metav1.Time{
					Time: fakeClock.Now().Add(-46 * 24 * time.Hour),
				},
			},
			Status: corev1.NamespaceStatus{
				Phase: corev1.NamespaceActive,
			},
		},
		&corev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name: "kube-system",
				CreationTimestamp: metav1.Time{
					Time: fakeClock.Now().Add(-46 * 24 * time.Hour),
				},
			},
			Status: corev1.NamespaceStatus{
				Phase: corev1.NamespaceActive,
			},
		},
		&corev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name: "ns1",
				CreationTimestamp: metav1.Time{
					Time: fakeClock.Now().Add(-10 * time.Minute),
				},
			},
			Status: corev1.NamespaceStatus{
				Phase: corev1.NamespaceTerminating,
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
						"name": "ns1",
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
	resourceModel, err := discoverer.DiscoverResourcesForNamespace(resourcediscovery.Filter{})
	if err != nil {
		t.Fatalf("Failed to construct resourceModel: %v", err)
	}

	nsp := &NamespacesPrinter{
		Writer: buff,
		Clock:  fakeClock,
	}
	nsp.PrintTable(resourceModel, false)

	got := buff.String()
	want := `
NAME         STATUS       AGE
default      Active       46d
kube-system  Active       46d
ns1          Terminating  10m
`
	if diff := cmp.Diff(common.YamlString(want), common.YamlString(got), common.YamlStringTransformer); diff != "" {
		t.Errorf("Unexpected diff\ngot=\n%v\nwant=\n%v\ndiff (-want +got)=\n%v", got, want, diff)
	}

	buff.Reset()
	nsp2 := &NamespacesPrinter{
		Writer: buff,
		Clock:  fakeClock,
	}
	nsp2.PrintTable(resourceModel, true)

	got2 := buff.String()
	want2 := `
NAME         STATUS       AGE  POLICIES
default      Active       46d  0
kube-system  Active       46d  0
ns1          Terminating  10m  1
`
	if diff := cmp.Diff(common.YamlString(want2), common.YamlString(got2), common.YamlStringTransformer); diff != "" {
		t.Errorf("Unexpected diff\ngot=\n%v\nwant=\n%v\ndiff (-want +got)=\n%v", got2, want2, diff)
	}
}

func TestNamespacePrinter_PrintDescribeView(t *testing.T) {
	fakeClock := testingclock.NewFakeClock(time.Now())
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

	k8sClients := common.MustClientsForTest(t, objects...)
	policyManager := utils.MustPolicyManagerForTest(t, k8sClients)
	buff := &bytes.Buffer{}
	discoverer := resourcediscovery.Discoverer{
		K8sClients:    k8sClients,
		PolicyManager: policyManager,
	}
	resourceModel, err := discoverer.DiscoverResourcesForNamespace(resourcediscovery.Filter{})
	if err != nil {
		t.Fatalf("Failed to construct resourceModel: %v", err)
	}

	nsp := &NamespacesPrinter{
		Writer:       buff,
		Clock:        fakeClock,
		EventFetcher: discoverer,
	}
	nsp.PrintDescribeView(resourceModel)

	got := buff.String()
	want := `
Name: development
Labels:
  type: test-namespace
Annotations:
  test-annotation: development-annotation
Status:
  phase: Active
DirectlyAttachedPolicies:
  Type                       Name
  ----                       ----
  HealthCheckPolicy.foo.com  health-check-gatewayclass
Events: <none>


Name: production
Labels:
  type: production-namespace
Annotations: null
Status:
  phase: Active
DirectlyAttachedPolicies:
  Type                   Name
  ----                   ----
  TimeoutPolicy.bar.com  timeout-policy-namespace
Events: <none>
`
	if diff := cmp.Diff(common.YamlString(want), common.YamlString(got), common.YamlStringTransformer); diff != "" {
		t.Errorf("Unexpected diff\ngot=\n%v\nwant=\n%v\ndiff (-want +got)=\n%v", got, want, diff)
	}
}

// TestNamespacesPrinter_PrintJsonYaml tests the correctness of JSON/YAML output associated with -o json/yaml of `get` subcommand
func TestNamespacesPrinter_PrintJsonYaml(t *testing.T) {
	fakeClock := testingclock.NewFakeClock(time.Now())
	creationTime := fakeClock.Now().Add(-46 * 24 * time.Hour).UTC() // UTC being necessary for consistently handling the time while marshaling/unmarshaling its JSON

	nsObject := &corev1.Namespace{
		TypeMeta: metav1.TypeMeta{
			Kind:       "v1",
			APIVersion: "Namespace",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: "namespace-1",
			CreationTimestamp: metav1.Time{
				Time: creationTime,
			},
			Labels: map[string]string{"app": "foo", "env": "internal"},
		},
		Spec: corev1.NamespaceSpec{
			Finalizers: []corev1.FinalizerName{"kubernetes"},
		},
		Status: corev1.NamespaceStatus{
			Phase: corev1.NamespaceActive,
		},
	}

	k8sClients := common.MustClientsForTest(t, nsObject)
	policyManager := utils.MustPolicyManagerForTest(t, k8sClients)
	buff := &bytes.Buffer{}
	discoverer := resourcediscovery.Discoverer{
		K8sClients:    k8sClients,
		PolicyManager: policyManager,
	}
	resourceModel, err := discoverer.DiscoverResourcesForNamespace(resourcediscovery.Filter{})
	if err != nil {
		t.Fatalf("Failed to construct resourceModel: %v", resourceModel)
	}

	nsp := &NamespacesPrinter{
		Writer: buff,
		Clock:  fakeClock,
	}
	Print(nsp, resourceModel, utils.OutputFormatJSON)

	gotJSON := common.JSONString(buff.String())
	wantJSON := common.JSONString(fmt.Sprintf(`
        {
          "apiVersion": "v1",
          "items": [
            {
              "apiVersion": "Namespace",
              "kind": "v1",
              "metadata": {
                "creationTimestamp": "%s",
                "labels": {
                  "app": "foo",
                  "env": "internal"
                },
                "name": "namespace-1",
                "resourceVersion": "999"
              },
              "spec": {
                "finalizers": [
                  "kubernetes"
                ]
              },
              "status": {
                "phase": "Active"
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

	nsp.Writer = &bytes.Buffer{}
	Print(nsp, resourceModel, utils.OutputFormatYAML)

	gotYaml := common.YamlString(nsp.Writer.(*bytes.Buffer).String())
	wantYaml := common.YamlString(fmt.Sprintf(`
apiVersion: v1
items:
- apiVersion: Namespace
  kind: v1
  metadata:
    creationTimestamp: "%s"
    labels:
      app: foo
      env: internal
    name: namespace-1
    resourceVersion: "999"
  spec:
    finalizers:
    - kubernetes
  status:
    phase: Active
kind: List
`, creationTime.Format(time.RFC3339)))
	if diff := cmp.Diff(wantYaml, gotYaml, common.YamlStringTransformer); diff != "" {
		t.Errorf("Unexpected diff\ngot=\n%v\nwant=\n%v\ndiff (-want +got)=\n%v", gotYaml, wantYaml, diff)
	}
}
