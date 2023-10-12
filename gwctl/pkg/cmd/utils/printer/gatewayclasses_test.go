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

func TestGatewayClassesPrinter_PrintDescribeView(t *testing.T) {
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
	}

	params := utils.MustParamsForTest(t, common.MustClientsForTest(t, objects...))
	gws, err := resourcehelpers.ListGatewayClasses(context.Background(), params.K8sClients)
	if err != nil {
		t.Fatalf("Failed to List GatewayClasses: %v", err)
	}

	gcp := &GatewayClassesPrinter{
		Out: params.Out,
		EPC: effectivepolicy.NewCalculator(params.K8sClients, params.PolicyManager),
	}
	gcp.PrintDescribeView(context.Background(), gws)

	got := params.Out.(*bytes.Buffer).String()
	want := `
Name: foo-gatewayclass
ControllerName: example.net/gateway-controller
Description: random
DirectlyAttachedPolicies:
- Group: foo.com
  Kind: HealthCheckPolicy
  Name: policy-name
`
	if diff := cmp.Diff(common.YamlString(want), common.YamlString(got), common.YamlStringTransformer); diff != "" {
		t.Errorf("Unexpected diff\ngot=\n%v\nwant=\n%v\ndiff (-want +got)=\n%v", got, want, diff)
	}
}
