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

	gatewayv1 "sigs.k8s.io/gateway-api/apis/v1"
	"sigs.k8s.io/gateway-api/gwctl/pkg/extension/directlyattachedpolicy"
	"sigs.k8s.io/gateway-api/gwctl/pkg/extension/gatewayeffectivepolicy"
	extensionutils "sigs.k8s.io/gateway-api/gwctl/pkg/extension/utils"
	"sigs.k8s.io/gateway-api/gwctl/pkg/policymanager"
	"sigs.k8s.io/gateway-api/gwctl/pkg/topology"
	topologygw "sigs.k8s.io/gateway-api/gwctl/pkg/topology/gateway"
)

func (p *TablePrinter) printGateway(gatewayNode *topology.Node, w io.Writer) error {
	if err := p.checkTypeChange("Gateway", w); err != nil {
		return err
	}

	if p.table == nil {
		var columnNames []string
		if p.OutputFormat == OutputFormatWide {
			columnNames = []string{"NAMESPACE", "NAME", "CLASS", "ADDRESSES", "PORTS", "PROGRAMMED", "AGE", "POLICIES", "HTTPROUTES"}
		} else {
			columnNames = []string{"NAMESPACE", "NAME", "CLASS", "ADDRESSES", "PORTS", "PROGRAMMED", "AGE"}
		}
		p.table = &Table{
			ColumnNames:  columnNames,
			UseSeparator: false,
		}
	}

	gateway := topology.MustAccessObject(gatewayNode, &gatewayv1.Gateway{})

	var addresses []string
	for _, address := range gateway.Status.Addresses {
		addresses = append(addresses, address.Value)
	}
	addressesOutput := strings.Join(addresses, ",")
	if cnt := len(addresses); cnt > 2 {
		addressesOutput = fmt.Sprintf("%v + %v more", strings.Join(addresses[:2], ","), cnt-2)
	}

	var ports []string
	for _, listener := range gateway.Spec.Listeners {
		ports = append(ports, fmt.Sprintf("%d", int(listener.Port)))
	}
	portsOutput := strings.Join(ports, ",")

	programmedStatus := "Unknown"
	for _, condition := range gateway.Status.Conditions {
		if condition.Type == "Programmed" {
			programmedStatus = string(condition.Status)
			break
		}
	}

	age := "<unknown>"
	creationTimestamp := gateway.GetCreationTimestamp()
	if !creationTimestamp.IsZero() {
		age = duration.HumanDuration(p.Clock.Since(creationTimestamp.Time))
	}

	row := []string{
		gateway.GetNamespace(),
		gateway.GetName(),
		string(gateway.Spec.GatewayClassName),
		addressesOutput,
		portsOutput,
		programmedStatus,
		age,
	}
	if p.OutputFormat == OutputFormatWide {
		policiesMap, err := directlyattachedpolicy.Access(gatewayNode)
		if err != nil {
			return err
		}
		policiesCount := fmt.Sprintf("%d", len(policiesMap))

		httpRoutesCount := fmt.Sprintf("%d", len(topologygw.GatewayNode(gatewayNode).HTTPRoutes()))

		row = append(row, policiesCount, httpRoutesCount)
	}
	p.table.Rows = append(p.table.Rows, row)

	return nil
}

func (p *DescriptionPrinter) printGateway(gatewayNode *topology.Node, w io.Writer) error {
	if p.printSeparator {
		fmt.Fprintf(w, "\n\n")
	}
	p.printSeparator = true

	gateway := topology.MustAccessObject(gatewayNode, &gatewayv1.Gateway{})

	metadata := gateway.ObjectMeta.DeepCopy()
	metadata.Labels = nil
	metadata.Annotations = nil
	metadata.Name = ""
	metadata.Namespace = ""
	metadata.ManagedFields = nil

	pairs := []*DescriberKV{
		{Key: "Name", Value: gateway.GetName()},
		{Key: "Namespace", Value: gateway.GetNamespace()},
		{Key: "Labels", Value: gateway.Labels},
		{Key: "Annotations", Value: gateway.Annotations},
		{Key: "APIVersion", Value: gateway.APIVersion},
		{Key: "Kind", Value: gateway.Kind},
		{Key: "Metadata", Value: metadata},
		{Key: "Spec", Value: &gateway.Spec},
		{Key: "Status", Value: &gateway.Status},
	}

	const (
		maxHTTPRoutes = 10
		maxBackends   = 10
	)

	// AttachedRoutes
	attachedRoutes := &Table{
		ColumnNames:  []string{"Kind", "Name"},
		UseSeparator: true,
	}
	// Backends
	backends := &Table{
		ColumnNames:  []string{"Kind", "Name"},
		UseSeparator: true,
	}
	httpRouteCount, backendsCount := 0, 0
	httpRouteNodes := maps.Values(topologygw.GatewayNode(gatewayNode).HTTPRoutes())
	for _, httpRouteNode := range topology.SortedNodes(httpRouteNodes) {
		httpRouteCount++
		if httpRouteCount > maxHTTPRoutes {
			attachedRoutes.Rows = append(attachedRoutes.Rows, []string{fmt.Sprintf("(Truncated)")})
			break
		}
		row := []string{
			httpRouteNode.GKNN().Kind,                      // Kind
			httpRouteNode.GKNN().NamespacedName().String(), // Name
		}
		attachedRoutes.Rows = append(attachedRoutes.Rows, row)

		backendNodes := maps.Values(topologygw.HTTPRouteNode(httpRouteNode).Backends())
		for _, backendNode := range topology.SortedNodes(backendNodes) {
			backendsCount++
			if backendsCount > maxBackends {
				backends.Rows = append(backends.Rows, []string{fmt.Sprintf("(Truncated)")})
				break
			}
			row := []string{
				backendNode.GKNN().Kind,                      // Kind
				backendNode.GKNN().NamespacedName().String(), // Name
			}
			backends.Rows = append(backends.Rows, row)
		}
	}
	pairs = append(pairs, &DescriberKV{Key: "AttachedRoutes", Value: attachedRoutes})
	pairs = append(pairs, &DescriberKV{Key: "Backends", Value: backends})

	// DirectlyAttachedPolicies
	policiesMap, err := directlyattachedpolicy.Access(gatewayNode)
	if err != nil {
		return err
	}
	policies := policymanager.ConvertPoliciesMapToSlice(policiesMap)
	pairs = append(pairs, &DescriberKV{Key: "DirectlyAttachedPolicies", Value: convertPoliciesToRefsTable(policies, false)})

	// InheritedPolicies
	effectivePolicies, err := gatewayeffectivepolicy.Access(gatewayNode)
	if err != nil {
		return err
	}
	policies = policymanager.ConvertPoliciesMapToSlice(effectivePolicies.GatewayInheritedPolicies)
	pairs = append(pairs, &DescriberKV{Key: "InheritedPolicies", Value: convertPoliciesToRefsTable(policies, true)})

	// EffectivePolicies``
	if len(effectivePolicies.GatewayEffectivePolicies) != 0 {
		pairs = append(pairs, &DescriberKV{Key: "EffectivePolicies", Value: effectivePolicies.GatewayEffectivePolicies})
	}

	// // Analysis
	analysisErrors, err := extensionutils.AggregateAnalysisErrors(gatewayNode)
	if err != nil {
		return err
	}
	if len(analysisErrors) != 0 {
		pairs = append(pairs, &DescriberKV{Key: "Analysis", Value: convertErrorsToString(analysisErrors)})
	}

	// Events
	events, err := p.EventFetcher.FetchEventsFor(gateway)
	if err != nil {
		return err
	}
	pairs = append(pairs, &DescriberKV{Key: "Events", Value: convertEventsSliceToTable(events, p.Clock)})

	Describe(w, pairs)
	return nil
}
