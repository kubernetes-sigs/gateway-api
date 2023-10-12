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

	corev1 "k8s.io/api/core/v1"

	"sigs.k8s.io/gateway-api/gwctl/pkg/policymanager"
)

type namespaces struct {
	epc *Calculator
}

func (n *namespaces) GetDirectlyAttachedPolicies(ctx context.Context, name string) ([]policymanager.Policy, error) {
	ns := &corev1.Namespace{}
	gvks, _, err := n.epc.k8sClients.Client.Scheme().ObjectKinds(ns)
	if err != nil {
		return []policymanager.Policy{}, err
	}

	objRef := policymanager.ObjRef{
		Group: gvks[0].Group,
		Kind:  gvks[0].Kind,
		Name:  name,
	}
	return n.epc.policyManager.PoliciesAttachedTo(objRef), nil
}
