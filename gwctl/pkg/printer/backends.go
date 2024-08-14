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

	"sigs.k8s.io/gateway-api/gwctl/pkg/extension/directlyattachedpolicy"
	"sigs.k8s.io/gateway-api/gwctl/pkg/extension/gatewayeffectivepolicy"
	"sigs.k8s.io/gateway-api/gwctl/pkg/extension/refgrantvalidator"
	extensionutils "sigs.k8s.io/gateway-api/gwctl/pkg/extension/utils"
	"sigs.k8s.io/gateway-api/gwctl/pkg/policymanager"
	"sigs.k8s.io/gateway-api/gwctl/pkg/topology"
	topologygw "sigs.k8s.io/gateway-api/gwctl/pkg/topology/gateway"
)

func (p *TablePrinter) printBackend(backendNode *topology.Node, w io.Writer) error {
	if err := p.checkTypeChange("Backend", w); err != nil {
		return err
	}

	if p.table == nil {
		var columnNames []string
		if p.OutputFormat == OutputFormatWide {
			columnNames = []string{"NAMESPACE", "NAME", "TYPE", "AGE", "REFERRED BY ROUTES", "POLICIES"}
		} else {
			columnNames = []string{"NAMESPACE", "NAME", "TYPE", "AGE"}
		}
		p.table = &Table{
			ColumnNames:  columnNames,
			UseSeparator: false,
		}
	}

	backend := backendNode.Object

	namespace := backend.GetNamespace()
	name := backend.GetName()
	backendType := backend.GetKind()

	age := "<unknown>"
	creationTimestamp := backend.GetCreationTimestamp()
	if !creationTimestamp.IsZero() {
		age = duration.HumanDuration(p.Clock.Since(creationTimestamp.Time))
	}

	row := []string{
		namespace,
		name,
		backendType,
		age,
	}
	if p.OutputFormat == OutputFormatWide {
		httpRouteNodes := maps.Values(topologygw.BackendNode(backendNode).HTTPRoutes())
		sortedHTTPRouteNodes := topology.SortedNodes(httpRouteNodes)
		totalRoutes := len(sortedHTTPRouteNodes)
		var referredByRoutes string
		if totalRoutes == 0 {
			referredByRoutes = "None"
		} else {
			var routes []string
			for i, httpRouteNode := range sortedHTTPRouteNodes {
				if i < 2 {
					namespacedName := httpRouteNode.GKNN().NamespacedName().String()
					routes = append(routes, namespacedName)
				} else {
					break
				}
			}
			referredByRoutes = strings.Join(routes, ", ")
			if totalRoutes > 2 {
				referredByRoutes += fmt.Sprintf(" + %d more", totalRoutes-2)
			}
		}
		policiesMap, err := directlyattachedpolicy.Access(backendNode)
		if err != nil {
			return err
		}
		policiesCount := fmt.Sprintf("%d", len(policiesMap))
		row = append(row, referredByRoutes, policiesCount)
	}
	p.table.Rows = append(p.table.Rows, row)

	return nil
}

func (p *DescriptionPrinter) printBackend(backendNode *topology.Node, w io.Writer) error {
	if p.printSeparator {
		fmt.Fprintf(w, "\n\n")
	}
	p.printSeparator = true

	backend := backendNode.Object.DeepCopy()
	backend.SetLabels(nil)
	backend.SetAnnotations(nil)

	pairs := []*DescriberKV{
		{Key: "Name", Value: backendNode.Object.GetName()},
		{Key: "Namespace", Value: backendNode.Object.GetNamespace()},
		{Key: "Labels", Value: backendNode.Object.GetLabels()},
		{Key: "Annotations", Value: backendNode.Object.GetAnnotations()},
		{Key: "Backend", Value: backend},
	}

	// ReferencedByRoutes
	routes := &Table{
		ColumnNames:  []string{"Kind", "Name"},
		UseSeparator: true,
	}
	for _, httpRouteNode := range topologygw.BackendNode(backendNode).HTTPRoutes() {
		row := []string{
			httpRouteNode.GKNN().Kind,                      // Kind
			httpRouteNode.GKNN().NamespacedName().String(), // Name
		}
		routes.Rows = append(routes.Rows, row)
	}
	pairs = append(pairs, &DescriberKV{Key: "ReferencedByRoutes", Value: routes})

	// DirectlyAttachedPolicies
	policiesMap, err := directlyattachedpolicy.Access(backendNode)
	if err != nil {
		return err
	}
	policies := policymanager.ConvertPoliciesMapToSlice(policiesMap)
	pairs = append(pairs, &DescriberKV{Key: "DirectlyAttachedPolicies", Value: convertPoliciesToRefsTable(policies, false)})

	// InheritedPolicies
	effectivePolicies, err := gatewayeffectivepolicy.Access(backendNode)
	if err != nil {
		return err
	}
	policies = policymanager.ConvertPoliciesMapToSlice(effectivePolicies.BackendInheritedPolicies)
	pairs = append(pairs, &DescriberKV{Key: "InheritedPolicies", Value: convertPoliciesToRefsTable(policies, true)})

	// EffectivePolicies
	if err != nil {
		return err
	}
	if len(effectivePolicies.BackendEffectivePolicies) != 0 {
		pairs = append(pairs, &DescriberKV{Key: "EffectivePolicies", Value: effectivePolicies.BackendEffectivePolicies})
	}

	// ReferenceGrants
	referenceGrantsMetadata, err := refgrantvalidator.Access(backendNode)
	if err != nil {
		return err
	}
	if referenceGrantsMetadata != nil && len(referenceGrantsMetadata.ReferenceGrants) != 0 {
		var names []string
		for _, refGrantNode := range referenceGrantsMetadata.ReferenceGrants {
			names = append(names, refGrantNode.GetName())
		}
		pairs = append(pairs, &DescriberKV{Key: "ReferenceGrants", Value: names})
	}

	// Analysis
	analysisErrors, err := extensionutils.AggregateAnalysisErrors(backendNode)
	if err != nil {
		return err
	}
	if len(analysisErrors) != 0 {
		pairs = append(pairs, &DescriberKV{Key: "Analysis", Value: convertErrorsToString(analysisErrors)})
	}

	// Events
	events, err := p.EventFetcher.FetchEventsFor(backendNode.Object)
	if err != nil {
		return err
	}
	pairs = append(pairs, &DescriberKV{Key: "Events", Value: convertEventsSliceToTable(events, p.Clock)})

	Describe(w, pairs)
	return nil
}
