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

package effectivepolicy

import (
	"context"

	"sigs.k8s.io/gateway-api/gwctl/pkg/common"
	"sigs.k8s.io/gateway-api/gwctl/pkg/common/resourcehelpers"
	"sigs.k8s.io/gateway-api/gwctl/pkg/policymanager"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	gatewayv1beta1 "sigs.k8s.io/gateway-api/apis/v1beta1"
)

type backends struct {
	epc *Calculator
}

func (b *backends) GetDirectlyAttachedPolicies(ctx context.Context, backend unstructured.Unstructured) ([]policymanager.Policy, error) {
	objRef := policymanager.ObjRef{
		Group:     backend.GroupVersionKind().Group,
		Kind:      backend.GroupVersionKind().Kind,
		Name:      backend.GetName(),
		Namespace: backend.GetNamespace(),
	}
	return b.epc.policyManager.PoliciesAttachedTo(objRef), nil
}

func (b *backends) GetEffectivePolicies(ctx context.Context, backend unstructured.Unstructured) (map[string]map[policymanager.PolicyCrdID]policymanager.Policy, error) {
	result := make(map[string]map[policymanager.PolicyCrdID]policymanager.Policy)

	// Step 1: Aggregate all policies of the Backend and the Backend-namespace.
	backendPolicies, err := b.GetDirectlyAttachedPolicies(ctx, backend)
	if err != nil {
		return nil, err
	}
	backendNamespacePolicies, err := b.epc.Namespaces.GetDirectlyAttachedPolicies(ctx, backend.GetNamespace())
	if err != nil {
		return nil, err
	}

	// Step 2: Merge Backend and Backend-namespace policies by their kind.
	backendPoliciesByKind, err := policymanager.MergePoliciesOfSimilarKind(backendPolicies)
	if err != nil {
		return nil, err
	}
	backendNamespacePoliciesByKind, err := policymanager.MergePoliciesOfSimilarKind(backendNamespacePolicies)
	if err != nil {
		return nil, err
	}

	// Step 3: Find all HTTPRoutes which reference this Backend.
	httpRoutes, err := httpRoutesForBackend(ctx, b.epc.k8sClients, backend)
	if err != nil {
		return nil, err
	}

	// Step 4: Loop through all HTTPRoutes and get their effective policies. Merge
	// effective policies such that we get policies partitioned by Gateway.
	for _, httpRoute := range httpRoutes {
		httpRoutePoliciesByGateway, err := b.epc.HTTPRoutes.GetEffectivePolicies(ctx, httpRoute.GetNamespace(), httpRoute.GetName())
		if err != nil {
			return nil, err
		}

		for gatewayRef, policies := range httpRoutePoliciesByGateway {
			result[gatewayRef], err = policymanager.MergePoliciesOfSameHierarchy(result[gatewayRef], policies)
			if err != nil {
				return nil, err
			}
		}
	}

	// Step 5: Loop through all Gateways and merge the Backend and
	// Backend-namespace specific policies. Note that this needs to be done
	// separately from Step 4 i.e. we can't have this loop within Step 4 itself.
	// This is because we first want to merge all policies of the same-hierarchy
	// together and then move to the next hierarchy of Backend and
	// Backend-namespace.
	for gatewayRef := range result {
		// Merge all hierarchial policies.
		result[gatewayRef], err = policymanager.MergePoliciesOfDifferentHierarchy(result[gatewayRef], backendNamespacePoliciesByKind)
		if err != nil {
			return nil, err
		}

		result[gatewayRef], err = policymanager.MergePoliciesOfDifferentHierarchy(result[gatewayRef], backendPoliciesByKind)
		if err != nil {
			return nil, err
		}
	}

	return result, nil
}

func httpRoutesForBackend(ctx context.Context, k8sClients *common.K8sClients, backend unstructured.Unstructured) ([]gatewayv1beta1.HTTPRoute, error) {
	allHTTPRoutes, err := resourcehelpers.ListHTTPRoutes(ctx, k8sClients, "")
	if err != nil {
		return nil, err
	}

	var filteredHTTPRoutes []gatewayv1beta1.HTTPRoute
	for _, httpRoute := range allHTTPRoutes {
		found := false

		for _, rule := range httpRoute.Spec.Rules {
			for _, backendRef := range rule.BackendRefs {
				if *backendRef.Group != gatewayv1beta1.Group(backend.GroupVersionKind().Group) {
					continue
				}
				if *backendRef.Kind != gatewayv1beta1.Kind(backend.GroupVersionKind().Kind) {
					continue
				}
				if backendRef.Name != gatewayv1beta1.ObjectName(backend.GetName()) {
					continue
				}
				var ns string
				if backendRef.Namespace != nil {
					ns = string(*backendRef.Namespace)
				}
				if ns == "" {
					ns = httpRoute.GetNamespace()
				}
				if ns != backend.GetNamespace() {
					continue
				}
				found = true
				break
			}
			if found {
				break
			}
		}

		if found {
			filteredHTTPRoutes = append(filteredHTTPRoutes, httpRoute)
		}
	}

	return filteredHTTPRoutes, nil
}
