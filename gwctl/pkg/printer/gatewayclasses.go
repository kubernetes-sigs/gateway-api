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

type GatewayClassesPrinter struct {
	Out io.Writer
}

type gatewayClassDescribeView struct {
	// GatewayClass name
	Name           string `json:",omitempty"`
	ControllerName string `json:",omitempty"`
	// GatewayClass description
	Description              string                 `json:",omitempty"`
	DirectlyAttachedPolicies []policymanager.ObjRef `json:",omitempty"`
}

func (gcp *GatewayClassesPrinter) PrintDescribeView(resourceModel *resourcediscovery.ResourceModel) {
	index := 0
	for _, gatewayClassNode := range resourceModel.GatewayClasses {
		index++

		views := []gatewayClassDescribeView{
			{
				Name: gatewayClassNode.GatewayClass.GetName(),
			},
			{
				ControllerName: string(gatewayClassNode.GatewayClass.Spec.ControllerName),
				Description:    *gatewayClassNode.GatewayClass.Spec.Description,
			},
		}
		if policyRefs := resourcediscovery.ConvertPoliciesMapToPolicyRefs(gatewayClassNode.Policies); len(policyRefs) != 0 {
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

		if index+1 <= len(resourceModel.GatewayClasses) {
			fmt.Fprintf(gcp.Out, "\n\n")
		}
	}
}
