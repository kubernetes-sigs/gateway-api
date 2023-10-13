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

	"sigs.k8s.io/gateway-api/gwctl/pkg/policymanager"

	gatewayv1beta1 "sigs.k8s.io/gateway-api/apis/v1beta1"
)

type gatewayClasses struct {
	epc *Calculator
}

func (g *gatewayClasses) GetDirectlyAttachedPolicies(ctx context.Context, name string) ([]policymanager.Policy, error) {
	gw := &gatewayv1beta1.GatewayClass{}
	gvks, _, err := g.epc.k8sClients.Client.Scheme().ObjectKinds(gw)
	if err != nil {
		return []policymanager.Policy{}, err
	}

	objRef := policymanager.ObjRef{
		Group: gvks[0].Group,
		Kind:  gvks[0].Kind,
		Name:  name,
	}
	return g.epc.policyManager.PoliciesAttachedTo(objRef), nil
}
