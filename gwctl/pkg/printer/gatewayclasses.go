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

	"golang.org/x/exp/maps"
	"k8s.io/apimachinery/pkg/util/duration"
	"k8s.io/utils/clock"

	"sigs.k8s.io/gateway-api/gwctl/pkg/resourcediscovery"
)

var _ Printer = (*GatewayClassesPrinter)(nil)

type GatewayClassesPrinter struct {
	io.Writer
	Clock        clock.Clock
	EventFetcher eventFetcher
}

func (gcp *GatewayClassesPrinter) GetPrintableNodes(resourceModel *resourcediscovery.ResourceModel) []NodeResource {
	return NodeResources(maps.Values(resourceModel.GatewayClasses))
}

func (gcp *GatewayClassesPrinter) PrintTable(resourceModel *resourcediscovery.ResourceModel, wide bool) {
	var columnNames []string
	if wide {
		columnNames = []string{"NAME", "CONTROLLER", "ACCEPTED", "AGE", "GATEWAYS"}
	} else {
		columnNames = []string{"NAME", "CONTROLLER", "ACCEPTED", "AGE"}
	}
	table := &Table{
		ColumnNames:  columnNames,
		UseSeparator: false,
	}

	gatewayClassNodes := maps.Values(resourceModel.GatewayClasses)

	for _, gatewayClassNode := range SortByString(gatewayClassNodes) {
		accepted := "Unknown"
		for _, condition := range gatewayClassNode.GatewayClass.Status.Conditions {
			if condition.Type == "Accepted" {
				accepted = string(condition.Status)
			}
		}

		age := duration.HumanDuration(gcp.Clock.Since(gatewayClassNode.GatewayClass.GetCreationTimestamp().Time))

		row := []string{
			gatewayClassNode.GatewayClass.GetName(),
			string(gatewayClassNode.GatewayClass.Spec.ControllerName),
			accepted,
			age,
		}
		if wide {
			gatewayCount := fmt.Sprintf("%d", len(gatewayClassNode.Gateways))
			row = append(row, gatewayCount)
		}
		table.Rows = append(table.Rows, row)
	}

	table.Write(gcp, 0)
}

func (gcp *GatewayClassesPrinter) PrintDescribeView(resourceModel *resourcediscovery.ResourceModel) {
	index := 0
	for _, gatewayClassNode := range resourceModel.GatewayClasses {
		index++

		metadata := gatewayClassNode.GatewayClass.ObjectMeta.DeepCopy()
		metadata.Labels = nil
		metadata.Annotations = nil
		metadata.Name = ""
		metadata.Namespace = ""
		metadata.ManagedFields = nil

		pairs := []*DescriberKV{
			{Key: "Name", Value: gatewayClassNode.GatewayClass.GetName()},
			{Key: "Labels", Value: gatewayClassNode.GatewayClass.GetLabels()},
			{Key: "Annotations", Value: gatewayClassNode.GatewayClass.GetAnnotations()},
			{Key: "APIVersion", Value: gatewayClassNode.GatewayClass.APIVersion},
			{Key: "Kind", Value: gatewayClassNode.GatewayClass.Kind},
			{Key: "Metadata", Value: metadata},
			{Key: "Spec", Value: &gatewayClassNode.GatewayClass.Spec},
			{Key: "Status", Value: &gatewayClassNode.GatewayClass.Status},
		}

		// DirectlyAttachedPolicies
		policyRefs := resourcediscovery.ConvertPoliciesMapToPolicyRefs(gatewayClassNode.Policies)
		pairs = append(pairs, &DescriberKV{Key: "DirectlyAttachedPolicies", Value: convertPolicyRefsToTable(policyRefs)})

		// Events
		eventList := gcp.EventFetcher.FetchEventsFor(context.Background(), gatewayClassNode.GatewayClass)
		pairs = append(pairs, &DescriberKV{Key: "Events", Value: convertEventsSliceToTable(eventList.Items, gcp.Clock)})

		Describe(gcp, pairs)

		if index+1 <= len(resourceModel.GatewayClasses) {
			fmt.Fprintf(gcp, "\n\n")
		}
	}
}
