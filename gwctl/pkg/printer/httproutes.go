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

	"k8s.io/apimachinery/pkg/util/duration"

	gatewayv1 "sigs.k8s.io/gateway-api/apis/v1"
	"sigs.k8s.io/gateway-api/gwctl/pkg/extension/directlyattachedpolicy"
	"sigs.k8s.io/gateway-api/gwctl/pkg/extension/gatewayeffectivepolicy"
	extensionutils "sigs.k8s.io/gateway-api/gwctl/pkg/extension/utils"
	"sigs.k8s.io/gateway-api/gwctl/pkg/policymanager"
	"sigs.k8s.io/gateway-api/gwctl/pkg/topology"
)

func (p *TablePrinter) printHTTPRoute(httpRouteNode *topology.Node, w io.Writer) error {
	if err := p.checkTypeChange("HTTPRoute", w); err != nil {
		return err
	}

	if p.table == nil {
		var columnNames []string
		if p.OutputFormat == OutputFormatWide {
			columnNames = []string{"NAMESPACE", "NAME", "HOSTNAMES", "PARENT REFS", "AGE", "POLICIES"}
		} else {
			columnNames = []string{"NAMESPACE", "NAME", "HOSTNAMES", "PARENT REFS", "AGE"}
		}

		p.table = &Table{
			ColumnNames:  columnNames,
			UseSeparator: false,
		}
	}

	httpRoute := topology.MustAccessObject(httpRouteNode, &gatewayv1.HTTPRoute{})

	var hostNames []string
	for _, hostName := range httpRoute.Spec.Hostnames {
		hostNames = append(hostNames, string(hostName))
	}
	hostNamesOutput := "None"
	if hostNamesCount := len(hostNames); hostNamesCount > 0 {
		if hostNamesCount > 2 {
			hostNamesOutput = fmt.Sprintf("%v + %v more", strings.Join(hostNames[:2], ","), hostNamesCount-2)
		} else {
			hostNamesOutput = strings.Join(hostNames, ",")
		}
	}

	parentRefsCount := fmt.Sprintf("%d", len(httpRoute.Spec.ParentRefs))

	age := "<unknown>"
	creationTimestamp := httpRoute.GetCreationTimestamp()
	if !creationTimestamp.IsZero() {
		age = duration.HumanDuration(p.Clock.Since(creationTimestamp.Time))
	}

	row := []string{
		httpRoute.GetNamespace(),
		httpRoute.GetName(),
		hostNamesOutput,
		parentRefsCount,
		age,
	}
	if p.OutputFormat == OutputFormatWide {
		policiesMap, err := directlyattachedpolicy.Access(httpRouteNode)
		if err != nil {
			return err
		}
		policiesCount := fmt.Sprintf("%d", len(policiesMap))
		row = append(row, policiesCount)
	}
	p.table.Rows = append(p.table.Rows, row)
	return nil
}

func (p *DescriptionPrinter) printHTTPRoute(httpRouteNode *topology.Node, w io.Writer) error {
	if p.printSeparator {
		fmt.Fprintf(w, "\n\n")
	}
	p.printSeparator = true

	httpRoute := topology.MustAccessObject(httpRouteNode, &gatewayv1.HTTPRoute{})

	metadata := httpRoute.ObjectMeta.DeepCopy()
	metadata.Labels = nil
	metadata.Annotations = nil
	metadata.Name = ""
	metadata.Namespace = ""
	metadata.ManagedFields = nil

	pairs := []*DescriberKV{
		{"Name", httpRoute.GetName()},
		{"Namespace", httpRoute.Namespace},
		{"Label", httpRoute.Labels},
		{"Annotations", httpRoute.Annotations},
		{"APIVersion", httpRoute.APIVersion},
		{"Kind", httpRoute.Kind},
		{"Metadata", metadata},
		{"Spec", httpRoute.Spec},
		{"Status", httpRoute.Status},
	}

	// DirectlyAttachedPolicies
	policiesMap, err := directlyattachedpolicy.Access(httpRouteNode)
	if err != nil {
		return err
	}
	policies := policymanager.ConvertPoliciesMapToSlice(policiesMap)
	pairs = append(pairs, &DescriberKV{Key: "DirectlyAttachedPolicies", Value: convertPoliciesToRefsTable(policies, false)})

	// InheritedPolicies
	effectivePolicies, err := gatewayeffectivepolicy.Access(httpRouteNode)
	if err != nil {
		return err
	}
	policies = policymanager.ConvertPoliciesMapToSlice(effectivePolicies.HTTPRouteInheritedPolicies)
	pairs = append(pairs, &DescriberKV{Key: "InheritedPolicies", Value: convertPoliciesToRefsTable(policies, true)})

	// EffectivePolicies
	if err != nil {
		return err
	}
	if len(effectivePolicies.HTTPRouteEffectivePolicies) != 0 {
		pairs = append(pairs, &DescriberKV{Key: "EffectivePolicies", Value: effectivePolicies.HTTPRouteEffectivePolicies})
	}

	// Analysis
	analysisErrors, err := extensionutils.AggregateAnalysisErrors(httpRouteNode)
	if err != nil {
		return err
	}
	if len(analysisErrors) != 0 {
		pairs = append(pairs, &DescriberKV{Key: "Analysis", Value: convertErrorsToString(analysisErrors)})
	}

	// Events
	events, err := p.EventFetcher.FetchEventsFor(httpRoute)
	if err != nil {
		return err
	}
	pairs = append(pairs, &DescriberKV{Key: "Events", Value: convertEventsSliceToTable(events, p.Clock)})

	Describe(w, pairs)
	return nil
}
