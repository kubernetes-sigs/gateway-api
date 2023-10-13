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
	_ "embed"

	gatewayv1beta1 "sigs.k8s.io/gateway-api/apis/v1beta1"

	"sigs.k8s.io/gateway-api/gwctl/pkg/common/resourcehelpers"
	"sigs.k8s.io/gateway-api/gwctl/pkg/policymanager"
)

type gateways struct {
	epc *Calculator
}

func (g *gateways) GetDirectlyAttachedPolicies(ctx context.Context, namespace, name string) ([]policymanager.Policy, error) {
	gw := &gatewayv1beta1.Gateway{}
	gvks, _, err := g.epc.k8sClients.Client.Scheme().ObjectKinds(gw)
	if err != nil {
		return []policymanager.Policy{}, err
	}

	objRef := policymanager.ObjRef{
		Group:     gvks[0].Group,
		Kind:      gvks[0].Kind,
		Name:      name,
		Namespace: namespace,
	}
	return g.epc.policyManager.PoliciesAttachedTo(objRef), nil
}

// getGatewayClassPolicies will get the policies attached to the GatewayClass of the given Gateway.
func (g *gateways) getGatewayClassPolicies(ctx context.Context, namespace, name string) ([]policymanager.Policy, error) {
	gw, err := resourcehelpers.GetGateways(ctx, g.epc.k8sClients, namespace, name)
	if err != nil {
		return []policymanager.Policy{}, err
	}

	return g.epc.GatewayClasses.GetDirectlyAttachedPolicies(ctx, string(gw.Spec.GatewayClassName))
}

func (g *gateways) GetEffectivePolicies(ctx context.Context, namespace, name string) (map[policymanager.PolicyCrdID]policymanager.Policy, error) {
	// Fetch all policies.
	gatewayClassPolicies, err := g.getGatewayClassPolicies(ctx, namespace, name)
	if err != nil {
		return nil, err
	}
	gatewayNamespacePolicies, err := g.epc.Namespaces.GetDirectlyAttachedPolicies(ctx, namespace)
	if err != nil {
		return nil, err
	}
	gatewayPolicies, err := g.GetDirectlyAttachedPolicies(ctx, namespace, name)
	if err != nil {
		return nil, err
	}

	// Merge policies by their kind.
	gatewayClassPoliciesByKind, err := policymanager.MergePoliciesOfSimilarKind(gatewayClassPolicies)
	if err != nil {
		return nil, err
	}
	gatewayNamespacePoliciesByKind, err := policymanager.MergePoliciesOfSimilarKind(gatewayNamespacePolicies)
	if err != nil {
		return nil, err
	}
	gatewayPoliciesByKind, err := policymanager.MergePoliciesOfSimilarKind(gatewayPolicies)
	if err != nil {
		return nil, err
	}

	// Merge all hierarchial policies.
	result, err := policymanager.MergePoliciesOfDifferentHierarchy(gatewayClassPoliciesByKind, gatewayNamespacePoliciesByKind)
	if err != nil {
		return nil, err
	}

	result, err = policymanager.MergePoliciesOfDifferentHierarchy(result, gatewayPoliciesByKind)
	if err != nil {
		return nil, err
	}

	return result, nil
}
