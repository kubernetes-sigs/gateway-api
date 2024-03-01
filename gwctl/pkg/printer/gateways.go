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
	"strings"
	"text/tabwriter"

	"sigs.k8s.io/gateway-api/gwctl/pkg/policymanager"
	"sigs.k8s.io/gateway-api/gwctl/pkg/resourcediscovery"
	"sigs.k8s.io/yaml"
	
	"k8s.io/apimachinery/pkg/util/duration"
	"k8s.io/utils/clock"
)

type GatewaysPrinter struct {
	Out   io.Writer
	Clock clock.Clock
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

func (gp *GatewaysPrinter) Print(resourceModel *resourcediscovery.ResourceModel) {
	tw := tabwriter.NewWriter(gp.Out, 0, 0, 2, ' ', 0)
	row := []string{"NAME", "CLASS", "ADDRESSES", "PORTS", "PROGRAMMED", "AGE"}
	tw.Write([]byte(strings.Join(row, "\t") + "\n"))

	for _, gatewayNode := range resourceModel.Gateways {
		var addresses []string
		for _, address := range gatewayNode.Gateway.Status.Addresses {
			addresses = append(addresses, address.Value)
		}
		addressesOutput := strings.Join(addresses, ",")
		if cnt := len(addresses); cnt > 2 {
			addressesOutput = fmt.Sprintf("%v + %v more", strings.Join(addresses[:2], ","), cnt-2)
		}

		var ports []string
		for _, listener := range gatewayNode.Gateway.Spec.Listeners {
			ports = append(ports, fmt.Sprintf("%d", int(listener.Port)))
		}
		portsOutput := strings.Join(ports, ",")

		programmedStatus := "Unknown"
		for _, condition := range gatewayNode.Gateway.Status.Conditions {
			if condition.Type == "Programmed" {
				programmedStatus = string(condition.Status)
				break
			}
		}

		age := duration.HumanDuration(gp.Clock.Since(gatewayNode.Gateway.GetCreationTimestamp().Time))

		row := []string{
			gatewayNode.Gateway.GetName(),
			string(gatewayNode.Gateway.Spec.GatewayClassName),
			addressesOutput,
			portsOutput,
			programmedStatus,
			age,
		}
		tw.Write([]byte(strings.Join(row, "\t") + "\n"))
	}
	tw.Flush()
}

func (gp *GatewaysPrinter) PrintDescribeView(resourceModel *resourcediscovery.ResourceModel) {
	index := 0
	for _, gatewayNode := range resourceModel.Gateways {
		index++
		views := []gatewayDescribeView{
			{
				Name:      gatewayNode.Gateway.GetName(),
				Namespace: gatewayNode.Gateway.GetNamespace(),
			},
			{
				GatewayClass: string(gatewayNode.Gateway.Spec.GatewayClassName),
			},
		}
		if policyRefs := resourcediscovery.ConvertPoliciesMapToPolicyRefs(gatewayNode.Policies); len(policyRefs) != 0 {
			views = append(views, gatewayDescribeView{
				DirectlyAttachedPolicies: policyRefs,
			})
		}
		if len(gatewayNode.EffectivePolicies) != 0 {
			views = append(views, gatewayDescribeView{
				EffectivePolicies: gatewayNode.EffectivePolicies,
			})
		}

		for _, view := range views {
			b, err := yaml.Marshal(view)
			if err != nil {
				panic(err)
			}
			fmt.Fprint(gp.Out, string(b))
		}

		if index+1 <= len(resourceModel.Gateways) {
			fmt.Fprintf(gp.Out, "\n\n")
		}
	}
}
