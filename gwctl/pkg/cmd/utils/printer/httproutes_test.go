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
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"sigs.k8s.io/gateway-api/gwctl/pkg/cmd/utils"
	"sigs.k8s.io/gateway-api/gwctl/pkg/common"
	"sigs.k8s.io/gateway-api/gwctl/pkg/common/resourcehelpers"
	"sigs.k8s.io/gateway-api/gwctl/pkg/effectivepolicy"

	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	gatewayv1alpha2 "sigs.k8s.io/gateway-api/apis/v1alpha2"
	gatewayv1beta1 "sigs.k8s.io/gateway-api/apis/v1beta1"
)

func TestHTTPRoutesPrinter_PrintDescribeView(t *testing.T) {
	objects := []runtime.Object{
		&gatewayv1beta1.GatewayClass{
			ObjectMeta: metav1.ObjectMeta{
				Name: "foo-gatewayclass",
			},
			Spec: gatewayv1beta1.GatewayClassSpec{
				ControllerName: "example.net/gateway-controller",
				Description:    common.PtrTo("random"),
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

		&gatewayv1beta1.Gateway{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "foo-gateway",
				Namespace: "default",
			},
			Spec: gatewayv1beta1.GatewaySpec{
				GatewayClassName: "foo-gatewayclass",
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

		&gatewayv1beta1.HTTPRoute{
			ObjectMeta: metav1.ObjectMeta{
				Name: "foo-httproute",
			},
			Spec: gatewayv1beta1.HTTPRouteSpec{
				CommonRouteSpec: gatewayv1beta1.CommonRouteSpec{
					ParentRefs: []gatewayv1beta1.ParentReference{{
						Kind:  common.PtrTo(gatewayv1beta1.Kind("Gateway")),
						Group: common.PtrTo(gatewayv1beta1.Group("gateway.networking.k8s.io")),
						Name:  "foo-gateway",
					}},
				},
			},
		},
		&unstructured.Unstructured{
			Object: map[string]interface{}{
				"apiVersion": "bar.com/v1",
				"kind":       "TimeoutPolicy",
				"metadata": map[string]interface{}{
					"name": "timeout-policy-httproute",
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
	}

	params := utils.MustParamsForTest(t, common.MustClientsForTest(t, objects...))
	httpRoutes, err := resourcehelpers.ListHTTPRoutes(context.Background(), params.K8sClients, "")
	if err != nil {
		t.Fatalf("Failed to List HTTPRoutespkg/resources/httproutes/httproutes_test.go: %v", err)
	}

	hp := &HTTPRoutesPrinter{
		Out: params.Out,
		EPC: effectivepolicy.NewCalculator(params.K8sClients, params.PolicyManager),
	}
	hp.PrintDescribeView(context.Background(), httpRoutes)

	got := params.Out.(*bytes.Buffer).String()
	want := `
Name: foo-httproute
ParentRefs:
- group: gateway.networking.k8s.io
  kind: Gateway
  name: foo-gateway
DirectlyAttachedPolicies:
- Group: bar.com
  Kind: TimeoutPolicy
  Name: timeout-policy-httproute
EffectivePolicies:
  default/foo-gateway:
    HealthCheckPolicy.foo.com:
      key1: value-parent-1
      key2: value-child-2
      key3: value-parent-3
      key4: value-parent-4
      key5: value-parent-5
    TimeoutPolicy.bar.com:
      condition: path=/def
      seconds: 60
`
	if diff := cmp.Diff(common.YamlString(want), common.YamlString(got), common.YamlStringTransformer); diff != "" {
		t.Errorf("Unexpected diff\ngot=\n%v\nwant=\n%v\ndiff (-want +got)=\n%v", got, want, diff)
	}
}
