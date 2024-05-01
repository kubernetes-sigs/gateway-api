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
	"os"
	"sort"
	"strings"
	"text/tabwriter"

	"sigs.k8s.io/gateway-api/gwctl/pkg/resourcediscovery"

	"k8s.io/apimachinery/pkg/util/duration"
	"k8s.io/utils/clock"
)

type GatewaysPrinter struct {
	Out   io.Writer
	Clock clock.Clock
}

func (gp *GatewaysPrinter) Print(resourceModel *resourcediscovery.ResourceModel) {
	tw := tabwriter.NewWriter(gp.Out, 0, 0, 2, ' ', 0)
	row := []string{"NAME", "CLASS", "ADDRESSES", "PORTS", "PROGRAMMED", "AGE"}
	_, err := tw.Write([]byte(strings.Join(row, "\t") + "\n"))
	if err != nil {
		fmt.Fprint(os.Stderr, err)
		os.Exit(1)
	}

	gatewayNodes := make([]*resourcediscovery.GatewayNode, 0, len(resourceModel.Gateways))
	for _, gatewayNode := range resourceModel.Gateways {
		gatewayNodes = append(gatewayNodes, gatewayNode)
	}

	sort.Slice(gatewayNodes, func(i, j int) bool {
		if gatewayNodes[i].Gateway.GetName() != gatewayNodes[j].Gateway.GetName() {
			return gatewayNodes[i].Gateway.GetName() < gatewayNodes[j].Gateway.GetName()
		}
		return gatewayNodes[i].Gateway.Spec.GatewayClassName < gatewayNodes[j].Gateway.Spec.GatewayClassName
	})

	for _, gatewayNode := range gatewayNodes {
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
		_, err := tw.Write([]byte(strings.Join(row, "\t") + "\n"))
		if err != nil {
			fmt.Fprint(os.Stderr, err)
			os.Exit(1)
		}
	}
	tw.Flush()
}

func (gp *GatewaysPrinter) PrintDescribeView(resourceModel *resourcediscovery.ResourceModel) {
	index := 0
	for _, gatewayNode := range resourceModel.Gateways {
		index++

		metadata := gatewayNode.Gateway.ObjectMeta.DeepCopy()
		metadata.Labels = nil
		metadata.Annotations = nil
		metadata.Name = ""
		metadata.Namespace = ""

		pairs := []*DescriberKV{
			{Key: "Name", Value: gatewayNode.Gateway.GetName()},
			{Key: "Namespace", Value: gatewayNode.Gateway.GetNamespace()},
			{Key: "Labels", Value: gatewayNode.Gateway.Labels},
			{Key: "Annotations", Value: gatewayNode.Gateway.Annotations},
			{Key: "APIVersion", Value: gatewayNode.Gateway.APIVersion},
			{Key: "Kind", Value: gatewayNode.Gateway.Kind},
			{Key: "Metadata", Value: metadata},
			{Key: "Spec", Value: &gatewayNode.Gateway.Spec},
			{Key: "Status", Value: &gatewayNode.Gateway.Status},
		}

		// AttachedRoutes
		attachedRoutes := &Table{
			ColumnNames:  []string{"Kind", "Name"},
			UseSeparator: true,
		}
		for _, httpRouteNode := range gatewayNode.HTTPRoutes {
			row := []string{
				httpRouteNode.HTTPRoute.Kind, // Kind
				fmt.Sprintf("%v/%v", httpRouteNode.HTTPRoute.Namespace, httpRouteNode.HTTPRoute.Name), // Name
			}
			attachedRoutes.Rows = append(attachedRoutes.Rows, row)
		}
		pairs = append(pairs, &DescriberKV{Key: "AttachedRoutes", Value: attachedRoutes})

		// DirectlyAttachedPolicies
		if policyRefs := resourcediscovery.ConvertPoliciesMapToPolicyRefs(gatewayNode.Policies); len(policyRefs) != 0 {
			directlyAttachedPolicies := &Table{
				ColumnNames:  []string{"Type", "Name"},
				UseSeparator: true,
			}
			for _, policyRef := range policyRefs {
				row := []string{
					fmt.Sprintf("%v.%v", policyRef.Kind, policyRef.Group),     // Type
					fmt.Sprintf("%v/%v", policyRef.Namespace, policyRef.Name), // Name
				}
				directlyAttachedPolicies.Rows = append(directlyAttachedPolicies.Rows, row)
			}
			pairs = append(pairs, &DescriberKV{Key: "DirectlyAttachedPolicies", Value: directlyAttachedPolicies})
		}

		// EffectivePolicies
		if len(gatewayNode.EffectivePolicies) != 0 {
			pairs = append(pairs, &DescriberKV{Key: "EffectivePolicies", Value: gatewayNode.EffectivePolicies})
		}

		// Events
		pairs = append(pairs, &DescriberKV{Key: "Events", Value: convertEventsSliceToTable(gatewayNode.Events, gp.Clock)})

		Describe(gp.Out, pairs)

		if index+1 <= len(resourceModel.Gateways) {
			fmt.Fprintf(gp.Out, "\n\n")
		}
	}
}
