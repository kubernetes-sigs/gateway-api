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

	gatewayv1beta1 "sigs.k8s.io/gateway-api/apis/v1beta1"
	"sigs.k8s.io/yaml"
)

type GatewayClassesPrinter struct {
	Out io.Writer
	EPC *effectivepolicy.Calculator
}

type gatewayClassDescribeView struct {
	// GatewayClass name
	Name           string `json:",omitempty"`
	ControllerName string `json:",omitempty"`
	// GatewayClass description
	Description              string                 `json:",omitempty"`
	DirectlyAttachedPolicies []policymanager.ObjRef `json:",omitempty"`
}

func (gcp *GatewayClassesPrinter) PrintDescribeView(ctx context.Context, gwClasses []gatewayv1beta1.GatewayClass) {
	for i, gwc := range gwClasses {
		directlyAttachedPolicies, err := gcp.EPC.GatewayClasses.GetDirectlyAttachedPolicies(ctx, gwc.Name)
		if err != nil {
			panic(err)
		}

		policyRefs := policymanager.ToPolicyRefs(directlyAttachedPolicies)

		views := []gatewayClassDescribeView{
			{
				Name: gwc.GetName(),
			},
			{
				ControllerName: string(gwc.Spec.ControllerName),
				Description:    *gwc.Spec.Description,
			},
		}
		if len(policyRefs) != 0 {
			views = append(views, gatewayClassDescribeView{
				DirectlyAttachedPolicies: policyRefs,
			})
		}

		for _, view := range views {
			b, err := yaml.Marshal(view)
			if err != nil {
				panic(err)
			}
			fmt.Fprint(gcp.Out, string(b))
		}

		if i+1 != len(gwClasses) {
			fmt.Fprintf(gcp.Out, "\n\n")
		}
	}
}
