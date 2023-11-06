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

	gatewayv1beta1 "sigs.k8s.io/gateway-api/apis/v1beta1"
	"sigs.k8s.io/yaml"

	"sigs.k8s.io/gateway-api/gwctl/pkg/effectivepolicy"
	"sigs.k8s.io/gateway-api/gwctl/pkg/policymanager"
)

type GatewaysPrinter struct {
	Out io.Writer
	EPC *effectivepolicy.Calculator
}

type gatewayDescribeView struct {
	// Gateway name
	Name string `json:",omitempty"`
	// Gateway namespace
	Namespace                string                                             `json:",omitempty"`
	GatewayClass             string                                             `json:",omitempty"`
	DirectlyAttachedPolicies []policymanager.ObjRef                             `json:",omitempty"`
	EffectivePolicies        map[policymanager.PolicyCrdID]policymanager.Policy `json:",omitempty"`
}

func (gp *GatewaysPrinter) PrintDescribeView(ctx context.Context, gws []gatewayv1beta1.Gateway) {
	for i, gw := range gws {
		allPolicies, err := gp.EPC.Gateways.GetDirectlyAttachedPolicies(ctx, gw.Namespace, gw.Name)
		if err != nil {
			panic(err)
		}
		effectivePolicies, err := gp.EPC.Gateways.GetEffectivePolicies(ctx, gw.Namespace, gw.Name)
		if err != nil {
			panic(err)
		}

		views := []gatewayDescribeView{
			{
				Name:      gw.GetName(),
				Namespace: gw.GetNamespace(),
			},
			{
				GatewayClass: string(gw.Spec.GatewayClassName),
			},
		}
		if policyRefs := policymanager.ToPolicyRefs(allPolicies); len(policyRefs) != 0 {
			views = append(views, gatewayDescribeView{
				DirectlyAttachedPolicies: policyRefs,
			})
		}
		if len(effectivePolicies) != 0 {
			views = append(views, gatewayDescribeView{
				EffectivePolicies: effectivePolicies,
			})
		}

		for _, view := range views {
			b, err := yaml.Marshal(view)
			if err != nil {
				panic(err)
			}
			fmt.Fprint(gp.Out, string(b))
		}

		if i+1 != len(gws) {
			fmt.Fprintf(gp.Out, "\n\n")
		}
	}
}
