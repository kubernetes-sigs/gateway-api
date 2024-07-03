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
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	corev1 "k8s.io/api/core/v1"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	apimachinerytypes "k8s.io/apimachinery/pkg/types"

	gatewayv1 "sigs.k8s.io/gateway-api/apis/v1"
	gatewayv1alpha2 "sigs.k8s.io/gateway-api/apis/v1alpha2"
	"sigs.k8s.io/gateway-api/gwctl/pkg/common"
	"sigs.k8s.io/gateway-api/gwctl/pkg/utils"
)

func TestResourceModel_calculateInheritedPolicies(t *testing.T) {
	testcases := []struct {
		name    string
		objects []runtime.Object

		wantInheritedPoliciesForGateways   []apimachinerytypes.NamespacedName
		wantInheritedPoliciesForHTTPRoutes []apimachinerytypes.NamespacedName
		wantInheritedPoliciesForBackends   []apimachinerytypes.NamespacedName
	}{
		{
			name: "normal",
			objects: []runtime.Object{
				&gatewayv1.GatewayClass{
					ObjectMeta: metav1.ObjectMeta{
						Name: "foo-gatewayclass",
					},
				},
				common.NamespaceForTest("default"),
				&gatewayv1.Gateway{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "foo-gateway",
						Namespace: "default",
					},
					Spec: gatewayv1.GatewaySpec{
						GatewayClassName: "foo-gatewayclass",
					},
				},
				&gatewayv1.HTTPRoute{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "foo-httproute",
						Namespace: "default",
					},
					Spec: gatewayv1.HTTPRouteSpec{
						CommonRouteSpec: gatewayv1.CommonRouteSpec{
							ParentRefs: []gatewayv1.ParentReference{{
								Kind:  common.PtrTo(gatewayv1.Kind("Gateway")),
								Group: common.PtrTo(gatewayv1.Group("gateway.networking.k8s.io")),
								Name:  "foo-gateway",
							}},
						},
						Rules: []gatewayv1.HTTPRouteRule{
							{
								BackendRefs: []gatewayv1.HTTPBackendRef{
									{
										BackendRef: gatewayv1.BackendRef{
											BackendObjectReference: gatewayv1.BackendObjectReference{
												Kind: common.PtrTo(gatewayv1.Kind("Service")),
												Name: "foo-svc",
												Port: common.PtrTo(gatewayv1.PortNumber(80)),
											},
										},
									},
								},
							},
						},
					},
				},
				&corev1.Service{
					TypeMeta: metav1.TypeMeta{
						Kind:       "Service",
						APIVersion: "v1",
					},
					ObjectMeta: metav1.ObjectMeta{
						Name:      "foo-svc",
						Namespace: "default",
					},
				},

				&apiextensionsv1.CustomResourceDefinition{
					ObjectMeta: metav1.ObjectMeta{
						Name: "timeoutpolicies.bar.com",
						Labels: map[string]string{
							gatewayv1alpha2.PolicyLabelKey: "Inherited",
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
							"name": "timeout-policy-on-namespace",
						},
						"spec": map[string]interface{}{
							"targetRef": map[string]interface{}{
								"kind": "Namespace",
								"name": "default",
							},
						},
					},
				},

				&apiextensionsv1.CustomResourceDefinition{
					ObjectMeta: metav1.ObjectMeta{
						Name: "healthcheckpolicies.bar.com",
						Labels: map[string]string{
							gatewayv1alpha2.PolicyLabelKey: "Inherited",
						},
					},
					Spec: apiextensionsv1.CustomResourceDefinitionSpec{
						Scope:    apiextensionsv1.NamespaceScoped,
						Group:    "bar.com",
						Versions: []apiextensionsv1.CustomResourceDefinitionVersion{{Name: "v1"}},
						Names: apiextensionsv1.CustomResourceDefinitionNames{
							Plural: "healthcheckpolicies",
							Kind:   "HealthCheckPolicy",
						},
					},
				},
				&unstructured.Unstructured{
					Object: map[string]interface{}{
						"apiVersion": "bar.com/v1",
						"kind":       "HealthCheckPolicy",
						"metadata": map[string]interface{}{
							"name":      "health-check-policy-on-httproute",
							"namespace": "default",
						},
						"spec": map[string]interface{}{
							"targetRef": map[string]interface{}{
								"group": "gateway.networking.k8s.io",
								"kind":  "HTTPRoute",
								"name":  "foo-httproute",
							},
						},
					},
				},

				// Direct Policies should not appear in inherited policies.
				&apiextensionsv1.CustomResourceDefinition{
					ObjectMeta: metav1.ObjectMeta{
						Name: "tlspolicies.bar.com",
						Labels: map[string]string{
							gatewayv1alpha2.PolicyLabelKey: "Direct",
						},
					},
					Spec: apiextensionsv1.CustomResourceDefinitionSpec{
						Scope:    apiextensionsv1.NamespaceScoped,
						Group:    "bar.com",
						Versions: []apiextensionsv1.CustomResourceDefinitionVersion{{Name: "v1"}},
						Names: apiextensionsv1.CustomResourceDefinitionNames{
							Plural: "tlspolicies",
							Kind:   "TLSPolicy",
						},
					},
				},
				&unstructured.Unstructured{
					Object: map[string]interface{}{
						"apiVersion": "bar.com/v1",
						"kind":       "TLSPolicy",
						"metadata": map[string]interface{}{
							"name":      "tls-policy-on-httproute",
							"namespace": "default",
						},
						"spec": map[string]interface{}{
							"targetRef": map[string]interface{}{
								"group": "gateway.networking.k8s.io",
								"kind":  "HTTPRoute",
								"name":  "foo-httproute",
							},
						},
					},
				},
			},
			wantInheritedPoliciesForGateways: []apimachinerytypes.NamespacedName{
				{Name: "timeout-policy-on-namespace"},
			},
			wantInheritedPoliciesForHTTPRoutes: []apimachinerytypes.NamespacedName{
				{Name: "timeout-policy-on-namespace"},
			},
			wantInheritedPoliciesForBackends: []apimachinerytypes.NamespacedName{
				{Name: "timeout-policy-on-namespace"},
				{Namespace: "default", Name: "health-check-policy-on-httproute"},
			},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			k8sClients := common.MustClientsForTest(t, tc.objects...)
			policyManager := utils.MustPolicyManagerForTest(t, k8sClients)
			discoverer := Discoverer{
				K8sClients:    k8sClients,
				PolicyManager: policyManager,
			}

			// TODO: Decouple this test from dependency on
			// DiscoverResourcesForBackend() and only invoke the function under
			// test viz. calculateInheritedPolicies()
			resourceModel, err := discoverer.DiscoverResourcesForBackend(Filter{})
			if err != nil {
				t.Fatalf("Failed to construct resourceModel: %v", resourceModel)
			}

			var gotInheritedPoliciesForGateways []apimachinerytypes.NamespacedName
			for _, gatewayNode := range resourceModel.Gateways {
				for _, policyNode := range gatewayNode.InheritedPolicies {
					gotInheritedPoliciesForGateways = append(gotInheritedPoliciesForGateways, apimachinerytypes.NamespacedName{
						Namespace: policyNode.Policy.Unstructured().GetNamespace(),
						Name:      policyNode.Policy.Unstructured().GetName(),
					})
				}
			}
			var gotInheritedPoliciesForHTTPRoutes []apimachinerytypes.NamespacedName
			for _, httpRouteNode := range resourceModel.HTTPRoutes {
				for _, policyNode := range httpRouteNode.InheritedPolicies {
					gotInheritedPoliciesForHTTPRoutes = append(gotInheritedPoliciesForHTTPRoutes, apimachinerytypes.NamespacedName{
						Namespace: policyNode.Policy.Unstructured().GetNamespace(),
						Name:      policyNode.Policy.Unstructured().GetName(),
					})
				}
			}
			var gotInheritedPoliciesForBackends []apimachinerytypes.NamespacedName
			for _, backendNode := range resourceModel.Backends {
				for _, policyNode := range backendNode.InheritedPolicies {
					gotInheritedPoliciesForBackends = append(gotInheritedPoliciesForBackends, apimachinerytypes.NamespacedName{
						Namespace: policyNode.Policy.Unstructured().GetNamespace(),
						Name:      policyNode.Policy.Unstructured().GetName(),
					})
				}
			}

			lessFunc := func(a, b apimachinerytypes.NamespacedName) bool {
				return fmt.Sprintf("%s/%s", a.Namespace, a.Name) < fmt.Sprintf("%s/%s", b.Namespace, b.Name)
			}

			if diff := cmp.Diff(tc.wantInheritedPoliciesForGateways, gotInheritedPoliciesForGateways, cmpopts.SortSlices(lessFunc)); diff != "" {
				t.Errorf("Unexpected diff in inheritedPoliciesForGateways: (-want, +got):\n%v", diff)
			}
			if diff := cmp.Diff(tc.wantInheritedPoliciesForHTTPRoutes, gotInheritedPoliciesForHTTPRoutes, cmpopts.SortSlices(lessFunc)); diff != "" {
				t.Errorf("Unexpected diff in inheritedPoliciesForHTTPRoutes: (-want, +got):\n%v", diff)
			}
			if diff := cmp.Diff(tc.wantInheritedPoliciesForBackends, gotInheritedPoliciesForBackends, cmpopts.SortSlices(lessFunc)); diff != "" {
				t.Errorf("Unexpected diff in inheritedPoliciesForBackends: (-want, +got):\n%v", diff)
			}
		})
	}
}
