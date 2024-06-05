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

	"golang.org/x/exp/maps"
	"k8s.io/apimachinery/pkg/util/duration"
	"k8s.io/utils/clock"

	"sigs.k8s.io/gateway-api/gwctl/pkg/resourcediscovery"
)

var _ Printer = (*GatewaysPrinter)(nil)

type GatewaysPrinter struct {
	io.Writer
	Clock clock.Clock
}

func (gp *GatewaysPrinter) GetPrintableNodes(resourceModel *resourcediscovery.ResourceModel) []NodeResource {
	return NodeResources(maps.Values(resourceModel.Gateways))
}

func (gp *GatewaysPrinter) PrintTable(resourceModel *resourcediscovery.ResourceModel) {
	table := &Table{
		ColumnNames:  []string{"NAMESPACE", "NAME", "CLASS", "ADDRESSES", "PORTS", "PROGRAMMED", "AGE"},
		UseSeparator: false,
	}

	gatewayNodes := maps.Values(resourceModel.Gateways)

	for _, gatewayNode := range SortByString(gatewayNodes) {
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
			gatewayNode.Gateway.GetNamespace(),
			gatewayNode.Gateway.GetName(),
			string(gatewayNode.Gateway.Spec.GatewayClassName),
			addressesOutput,
			portsOutput,
			programmedStatus,
			age,
		}
		table.Rows = append(table.Rows, row)
	}

	table.Write(gp, 0)
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
		metadata.ManagedFields = nil

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
		policyRefs := resourcediscovery.ConvertPoliciesMapToPolicyRefs(gatewayNode.Policies)
		pairs = append(pairs, &DescriberKV{Key: "DirectlyAttachedPolicies", Value: convertPolicyRefsToTable(policyRefs)})

		// EffectivePolicies
		if len(gatewayNode.EffectivePolicies) != 0 {
			pairs = append(pairs, &DescriberKV{Key: "EffectivePolicies", Value: gatewayNode.EffectivePolicies})
		}

		// Analysis
		if len(gatewayNode.Errors) != 0 {
			pairs = append(pairs, &DescriberKV{Key: "Analysis", Value: convertErrorsToString(gatewayNode.Errors)})
		}

		// Events
		pairs = append(pairs, &DescriberKV{Key: "Events", Value: convertEventsSliceToTable(gatewayNode.Events, gp.Clock)})

		Describe(gp, pairs)

		if index+1 <= len(resourceModel.Gateways) {
			fmt.Fprintf(gp, "\n\n")
		}
	}
}
