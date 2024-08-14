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

	"golang.org/x/exp/maps"
	"k8s.io/apimachinery/pkg/util/duration"

	gatewayv1 "sigs.k8s.io/gateway-api/apis/v1"
	"sigs.k8s.io/gateway-api/gwctl/pkg/extension/directlyattachedpolicy"
	"sigs.k8s.io/gateway-api/gwctl/pkg/policymanager"
	"sigs.k8s.io/gateway-api/gwctl/pkg/topology"
	topologygw "sigs.k8s.io/gateway-api/gwctl/pkg/topology/gateway"
)

func (p *TablePrinter) printGatewayClass(gatewayClassNode *topology.Node, w io.Writer) error {
	if err := p.checkTypeChange("GatewayClass", w); err != nil {
		return err
	}

	if p.table == nil {
		var columnNames []string
		if p.OutputFormat == OutputFormatWide {
			columnNames = []string{"NAME", "CONTROLLER", "ACCEPTED", "AGE", "GATEWAYS"}
		} else {
			columnNames = []string{"NAME", "CONTROLLER", "ACCEPTED", "AGE"}
		}
		p.table = &Table{
			ColumnNames:  columnNames,
			UseSeparator: false,
		}
	}

	gatewayClass := topology.MustAccessObject(gatewayClassNode, &gatewayv1.GatewayClass{})

	accepted := "Unknown"
	for _, condition := range gatewayClass.Status.Conditions {
		if condition.Type == "Accepted" {
			accepted = string(condition.Status)
		}
	}

	age := "<unknown>"
	creationTimestamp := gatewayClass.GetCreationTimestamp()
	if !creationTimestamp.IsZero() {
		age = duration.HumanDuration(p.Clock.Since(creationTimestamp.Time))
	}

	row := []string{
		gatewayClass.GetName(),
		string(gatewayClass.Spec.ControllerName),
		accepted,
		age,
	}
	if p.OutputFormat == OutputFormatWide {
		gatewayCount := fmt.Sprintf("%d", len(topologygw.GatewayClassNode(gatewayClassNode).Gateways()))
		row = append(row, gatewayCount)
	}
	p.table.Rows = append(p.table.Rows, row)
	return nil
}

func (p *DescriptionPrinter) printGatewayClass(gatewayClassNode *topology.Node, w io.Writer) error {
	if p.printSeparator {
		fmt.Fprintf(w, "\n\n")
	}
	p.printSeparator = true

	gatewayClass := topology.MustAccessObject(gatewayClassNode, &gatewayv1.GatewayClass{})

	metadata := gatewayClass.ObjectMeta.DeepCopy()
	metadata.Labels = nil
	metadata.Annotations = nil
	metadata.Name = ""
	metadata.Namespace = ""
	metadata.ManagedFields = nil

	pairs := []*DescriberKV{
		{Key: "Name", Value: gatewayClass.GetName()},
		{Key: "Labels", Value: gatewayClass.GetLabels()},
		{Key: "Annotations", Value: gatewayClass.GetAnnotations()},
		{Key: "APIVersion", Value: gatewayClass.APIVersion},
		{Key: "Kind", Value: gatewayClass.Kind},
		{Key: "Metadata", Value: metadata},
		{Key: "Spec", Value: &gatewayClass.Spec},
		{Key: "Status", Value: &gatewayClass.Status},
	}

	const (
		maxGateways = 10
	)

	// AttachedGateways
	attachedGateways := &Table{
		ColumnNames:  []string{"Kind", "Name"},
		UseSeparator: true,
	}
	gatewaysCount := 0
	gatewayNodes := maps.Values(topologygw.GatewayClassNode(gatewayClassNode).Gateways())
	for _, gatewayNode := range topology.SortedNodes(gatewayNodes) {
		gatewaysCount++
		if gatewaysCount > maxGateways {
			attachedGateways.Rows = append(attachedGateways.Rows, []string{fmt.Sprintf("(Truncated)")})
			break
		}
		row := []string{
			gatewayNode.GKNN().Kind,                      // Kind
			gatewayNode.GKNN().NamespacedName().String(), // Name
		}
		attachedGateways.Rows = append(attachedGateways.Rows, row)
	}
	pairs = append(pairs, &DescriberKV{Key: "AttachedGateways", Value: attachedGateways})

	// DirectlyAttachedPolicies
	policiesMap, err := directlyattachedpolicy.Access(gatewayClassNode)
	if err != nil {
		return err
	}
	policies := policymanager.ConvertPoliciesMapToSlice(policiesMap)
	pairs = append(pairs, &DescriberKV{Key: "DirectlyAttachedPolicies", Value: convertPoliciesToRefsTable(policies, false)})

	// Events
	events, err := p.EventFetcher.FetchEventsFor(gatewayClass)
	if err != nil {
		return err
	}
	pairs = append(pairs, &DescriberKV{Key: "Events", Value: convertEventsSliceToTable(events, p.Clock)})

	Describe(w, pairs)
	return nil
}
