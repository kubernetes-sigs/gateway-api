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
	"context"
	"fmt"
	"io"

	"sigs.k8s.io/gateway-api/gwctl/pkg/effectivepolicy"
	"sigs.k8s.io/gateway-api/gwctl/pkg/policymanager"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"sigs.k8s.io/yaml"
)

type BackendsPrinter struct {
	Out io.Writer
	EPC *effectivepolicy.Calculator
}

type backendDescribeView struct {
	Group                    string                                                        `json:",omitempty"`
	Kind                     string                                                        `json:",omitempty"`
	Name                     string                                                        `json:",omitempty"`
	Namespace                string                                                        `json:",omitempty"`
	DirectlyAttachedPolicies []policymanager.ObjRef                                        `json:",omitempty"`
	EffectivePolicies        map[string]map[policymanager.PolicyCrdID]policymanager.Policy `json:",omitempty"`
}

func (bp *BackendsPrinter) PrintDescribeView(ctx context.Context, backendsList []unstructured.Unstructured) {
	for i, backend := range backendsList {
		directlyAttachedPolicies, err := bp.EPC.Backends.GetDirectlyAttachedPolicies(ctx, backend)
		if err != nil {
			panic(err)
		}
		effectivePolicies, err := bp.EPC.Backends.GetEffectivePolicies(ctx, backend)
		if err != nil {
			panic(err)
		}

		views := []backendDescribeView{
			{
				Group:     backend.GroupVersionKind().Group,
				Kind:      backend.GroupVersionKind().Kind,
				Name:      backend.GetName(),
				Namespace: backend.GetNamespace(),
			},
		}
		if policyRefs := policymanager.ToPolicyRefs(directlyAttachedPolicies); len(policyRefs) != 0 {
			views = append(views, backendDescribeView{
				DirectlyAttachedPolicies: policyRefs,
			})
		}
		if len(effectivePolicies) != 0 {
			views = append(views, backendDescribeView{
				EffectivePolicies: effectivePolicies,
			})
		}

		for _, view := range views {
			b, err := yaml.Marshal(view)
			if err != nil {
				panic(err)
			}
			fmt.Fprint(bp.Out, string(b))
		}

		if i+1 != len(backendsList) {
			fmt.Fprintf(bp.Out, "\n\n")
		}
	}
}
