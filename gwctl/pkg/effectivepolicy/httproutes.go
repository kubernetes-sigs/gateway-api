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
	"fmt"

	gatewayv1beta1 "sigs.k8s.io/gateway-api/apis/v1beta1"

	"sigs.k8s.io/gateway-api/gwctl/pkg/common/resourcehelpers"
	"sigs.k8s.io/gateway-api/gwctl/pkg/policymanager"
)

type httpRoutes struct {
	epc *Calculator
}

func (h *httpRoutes) GetDirectlyAttachedPolicies(ctx context.Context, namespace, name string) ([]policymanager.Policy, error) {
	httpRoute := &gatewayv1beta1.HTTPRoute{}
	gvks, _, err := h.epc.k8sClients.Client.Scheme().ObjectKinds(httpRoute)
	if err != nil {
		return []policymanager.Policy{}, err
	}

	objRef := policymanager.ObjRef{
		Group:     gvks[0].Group,
		Kind:      gvks[0].Kind,
		Name:      name,
		Namespace: namespace,
	}
	return h.epc.policyManager.PoliciesAttachedTo(objRef), nil
}

func (h *httpRoutes) GetEffectivePolicies(ctx context.Context, namespace, name string) (map[string]map[policymanager.PolicyCrdID]policymanager.Policy, error) {
	result := make(map[string]map[policymanager.PolicyCrdID]policymanager.Policy)

	// Step 1: Aggregate all policies of the HTTPRoute and the HTTPRoute-namespace.
	httpRoutePolicies, err := h.GetDirectlyAttachedPolicies(ctx, namespace, name)
	if err != nil {
		return nil, err
	}
	httpRouteNamespacePolicies, err := h.epc.Namespaces.GetDirectlyAttachedPolicies(ctx, namespace)
	if err != nil {
		return nil, err
	}

	// Step 2: Merge HTTPRoute and HTTPRoute-namespace policies by their kind.
	httpRoutePoliciesByKind, err := policymanager.MergePoliciesOfSimilarKind(httpRoutePolicies)
	if err != nil {
		return nil, err
	}
	httpRouteNamespacePoliciesByKind, err := policymanager.MergePoliciesOfSimilarKind(httpRouteNamespacePolicies)
	if err != nil {
		return nil, err
	}

	// Step 3: Fetch the HTTPRoute to identify the Gateways it is attached to.
	httpRoute, err := resourcehelpers.GetHTTPRoute(ctx, h.epc.k8sClients, namespace, name)
	if err != nil {
		return result, err
	}

	// Step 4: Loop through all Gateways and merge policies for each Gateway. End
	// result is we get policies partitioned by each Gateway.
	for _, gatewayRef := range httpRoute.Spec.ParentRefs {
		ns := namespace
		if gatewayRef.Namespace != nil {
			ns = string(*gatewayRef.Namespace)
		}
		if ns == "" {
			ns = "default"
		}

		gatewayPoliciesByKind, err := h.epc.Gateways.GetEffectivePolicies(ctx, string(ns), string(gatewayRef.Name))
		if err != nil {
			return result, err
		}

		// Merge all hierarchial policies.
		mergedPolicies, err := policymanager.MergePoliciesOfDifferentHierarchy(gatewayPoliciesByKind, httpRouteNamespacePoliciesByKind)
		if err != nil {
			return nil, err
		}

		mergedPolicies, err = policymanager.MergePoliciesOfDifferentHierarchy(mergedPolicies, httpRoutePoliciesByKind)
		if err != nil {
			return nil, err
		}

		gatewayID := fmt.Sprintf("%v/%v", ns, gatewayRef.Name)
		result[gatewayID] = mergedPolicies
	}

	return result, nil
}
