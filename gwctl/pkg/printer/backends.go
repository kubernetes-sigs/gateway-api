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
	"fmt"
	"io"

	"sigs.k8s.io/yaml"

	"sigs.k8s.io/gateway-api/gwctl/pkg/policymanager"
	"sigs.k8s.io/gateway-api/gwctl/pkg/resourcediscovery"
)

type BackendsPrinter struct {
	Out io.Writer
}

type backendDescribeView struct {
	Group                    string                 `json:",omitempty"`
	Kind                     string                 `json:",omitempty"`
	Name                     string                 `json:",omitempty"`
	Namespace                string                 `json:",omitempty"`
	DirectlyAttachedPolicies []policymanager.ObjRef `json:",omitempty"`
	EffectivePolicies        any                    `json:",omitempty"`
}

func (bp *BackendsPrinter) PrintDescribeView(resourceModel *resourcediscovery.ResourceModel) {
	index := 0
	for _, backendNode := range resourceModel.Backends {
		index++

		views := []backendDescribeView{
			{
				Group:     backendNode.Backend.GroupVersionKind().Group,
				Kind:      backendNode.Backend.GroupVersionKind().Kind,
				Name:      backendNode.Backend.GetName(),
				Namespace: backendNode.Backend.GetNamespace(),
			},
		}
		if policyRefs := resourcediscovery.ConvertPoliciesMapToPolicyRefs(backendNode.Policies); len(policyRefs) != 0 {
			views = append(views, backendDescribeView{
				DirectlyAttachedPolicies: policyRefs,
			})
		}
		if len(backendNode.EffectivePolicies) != 0 {
			views = append(views, backendDescribeView{
				EffectivePolicies: backendNode.EffectivePolicies,
			})
		}

		for _, view := range views {
			b, err := yaml.Marshal(view)
			if err != nil {
				panic(err)
			}
			fmt.Fprint(bp.Out, string(b))
		}

		if index+1 <= len(resourceModel.Backends) {
			fmt.Fprintf(bp.Out, "\n\n")
		}
	}
}
