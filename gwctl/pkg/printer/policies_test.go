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
	"k8s.io/apimachinery/pkg/runtime"
	testingclock "k8s.io/utils/clock/testing"

	gatewayv1alpha2 "sigs.k8s.io/gateway-api/apis/v1alpha2"
	"sigs.k8s.io/gateway-api/gwctl/pkg/cmd/utils"
	"sigs.k8s.io/gateway-api/gwctl/pkg/common"
)

func TestPoliciesPrinter_Print_And_PrintDescribeView(t *testing.T) {
	fakeClock := testingclock.NewFakeClock(time.Now())
	objects := []runtime.Object{
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
						"group": "gateway.networking.k8s.io",
						"kind":  "HTTPRoute",
						"name":  "foo-httproute",
					},
				},
			},
		},
	}

	params := utils.MustParamsForTest(t, common.MustClientsForTest(t, objects...))

	pp := &PoliciesPrinter{
		Out:   &bytes.Buffer{},
		Clock: fakeClock,
	}

	pp.Print(params.PolicyManager.GetPolicies())
	got := pp.Out.(*bytes.Buffer).String()
	want := `
NAME                       KIND                       TARGET NAME       TARGET KIND   POLICY TYPE  AGE
health-check-gateway       HealthCheckPolicy.foo.com  foo-gateway       Gateway       Inherited    20d
health-check-gatewayclass  HealthCheckPolicy.foo.com  foo-gatewayclass  GatewayClass  Inherited    6d
timeout-policy-httproute   TimeoutPolicy.bar.com      foo-httproute     HTTPRoute     Direct       13m
timeout-policy-namespace   TimeoutPolicy.bar.com      default           Namespace     Direct       5m
`
	if diff := cmp.Diff(common.YamlString(want), common.YamlString(got), common.YamlStringTransformer); diff != "" {
		t.Errorf("Print: Unexpected diff\ngot=\n%v\nwant=\n%v\ndiff (-want +got)=\n%v", got, want, diff)
	}

	pp.Out = &bytes.Buffer{}
	pp.PrintDescribeView(params.PolicyManager.GetPolicies())
	got = pp.Out.(*bytes.Buffer).String()
	want = `
Name: health-check-gateway
Group: foo.com
Kind: HealthCheckPolicy
Inherited: "true"
Spec:
  default:
    key2: value-child-2
    key5: value-child-5
  override:
    key1: value-child-1
  targetRef:
    group: gateway.networking.k8s.io
    kind: Gateway
    name: foo-gateway
    namespace: default


Name: health-check-gatewayclass
Group: foo.com
Kind: HealthCheckPolicy
Inherited: "true"
Spec:
  default:
    key2: value-parent-2
    key4: value-parent-4
  override:
    key1: value-parent-1
    key3: value-parent-3
    key5: value-parent-5
  targetRef:
    group: gateway.networking.k8s.io
    kind: GatewayClass
    name: foo-gatewayclass


Name: timeout-policy-httproute
Group: bar.com
Kind: TimeoutPolicy
Inherited: "false"
Spec:
  condition: path=/def
  seconds: 60
  targetRef:
    group: gateway.networking.k8s.io
    kind: HTTPRoute
    name: foo-httproute


Name: timeout-policy-namespace
Group: bar.com
Kind: TimeoutPolicy
Inherited: "false"
Spec:
  condition: path=/abc
  seconds: 30
  targetRef:
    kind: Namespace
    name: default
`
	if diff := cmp.Diff(common.YamlString(want), common.YamlString(got), common.YamlStringTransformer); diff != "" {
		t.Errorf("PrintDescribeView: Unexpected diff\ngot=\n%v\nwant=\n%v\ndiff (-want +got)=\n%v", got, want, diff)
	}
}

func TestPoliciesPrinter_PrintCRDs(t *testing.T) {
	fakeClock := testingclock.NewFakeClock(time.Now())
	objects := []runtime.Object{
		&apiextensionsv1.CustomResourceDefinition{
			ObjectMeta: metav1.ObjectMeta{
				Name: "healthcheckpolicies.foo.com",
				Labels: map[string]string{
					gatewayv1alpha2.PolicyLabelKey: "inherited",
				},
				CreationTimestamp: metav1.Time{
					Time: fakeClock.Now().Add(-24 * 24 * time.Hour),
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
				CreationTimestamp: metav1.Time{
					Time: fakeClock.Now().Add(-5 * time.Minute),
				},
			},
			Spec: apiextensionsv1.CustomResourceDefinitionSpec{
				Scope:    apiextensionsv1.NamespaceScoped,
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
	}

	params := utils.MustParamsForTest(t, common.MustClientsForTest(t, objects...))
	pp := &PoliciesPrinter{
		Out:   &bytes.Buffer{},
		Clock: fakeClock,
	}
	pp.PrintCRDs(params.PolicyManager.GetCRDs())

	got := pp.Out.(*bytes.Buffer).String()
	want := `
NAME                         POLICY TYPE  SCOPE       AGE
healthcheckpolicies.foo.com  Inherited    Cluster     24d
timeoutpolicies.bar.com      Direct       Namespaced  5m
`
	if diff := cmp.Diff(common.YamlString(want), common.YamlString(got), common.YamlStringTransformer); diff != "" {
		t.Errorf("Unexpected diff\ngot=\n%v\nwant=\n%v\ndiff (-want +got)=\n%v", got, want, diff)
	}
}
