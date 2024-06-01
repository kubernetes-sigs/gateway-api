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
	"sort"
	"strings"

	"golang.org/x/exp/maps"
	"k8s.io/apimachinery/pkg/util/duration"
	"k8s.io/utils/clock"

	"sigs.k8s.io/gateway-api/gwctl/pkg/resourcediscovery"
)

var _ Printer = (*HTTPRoutesPrinter)(nil)

type HTTPRoutesPrinter struct {
	io.Writer
	Clock clock.Clock
}

func (hp *HTTPRoutesPrinter) GetPrintableNodes(resourceModel *resourcediscovery.ResourceModel) []NodeResource {
	return NodeResources(maps.Values(resourceModel.HTTPRoutes))
}

func (hp *HTTPRoutesPrinter) PrintTable(resourceModel *resourcediscovery.ResourceModel, wide bool) {
	var columnNames []string
	if wide {
		columnNames = []string{"NAMESPACE", "NAME", "HOSTNAMES", "PARENT REFS", "AGE", "POLICIES"}
	} else {
		columnNames = []string{"NAMESPACE", "NAME", "HOSTNAMES", "PARENT REFS", "AGE"}
	}

	table := &Table{
		ColumnNames:  columnNames,
		UseSeparator: false,
	}
	httpRouteNodes := maps.Values(resourceModel.HTTPRoutes)

	for _, httpRouteNode := range SortByString(httpRouteNodes) {
		var hostNames []string
		for _, hostName := range httpRouteNode.HTTPRoute.Spec.Hostnames {
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

		parentRefsCount := fmt.Sprintf("%d", len(httpRouteNode.HTTPRoute.Spec.ParentRefs))

		age := duration.HumanDuration(hp.Clock.Since(httpRouteNode.HTTPRoute.GetCreationTimestamp().Time))

		row := []string{
			httpRouteNode.HTTPRoute.GetNamespace(),
			httpRouteNode.HTTPRoute.GetName(),
			hostNamesOutput,
			parentRefsCount,
			age,
		}
		if wide {
			policiesCount := fmt.Sprintf("%d", len(httpRouteNode.Policies))
			row = append(row, policiesCount)
		}
		table.Rows = append(table.Rows, row)
	}
	table.Write(hp, 0)
}

func (hp *HTTPRoutesPrinter) PrintDescribeView(resourceModel *resourcediscovery.ResourceModel) {
	index := 0

	for _, httpRouteNode := range resourceModel.HTTPRoutes {
		index++

		metadata := httpRouteNode.HTTPRoute.ObjectMeta.DeepCopy()
		resetMetadataFields(metadata)

		namespace := handleDefaultNamespace(httpRouteNode.HTTPRoute.Namespace)

		pairs := []*DescriberKV{
			{"Name", httpRouteNode.HTTPRoute.GetName()},
			{"Namespace", namespace},
			{"Label", httpRouteNode.HTTPRoute.Labels},
			{"Annotations", httpRouteNode.HTTPRoute.Annotations},
			{"APIVersion", httpRouteNode.HTTPRoute.APIVersion},
			{"Kind", httpRouteNode.HTTPRoute.Kind},
			{"Metadata", metadata},
			{"Spec", httpRouteNode.HTTPRoute.Spec},
			{"Status", httpRouteNode.HTTPRoute.Status},
		}

		// DirectlyAttachedPolicies
		directlyAttachedPolicies := &Table{
			ColumnNames:  []string{"Type", "Name"},
			UseSeparator: true,
		}

		for _, policyNode := range httpRouteNode.Policies {
			if policyNode.Policy.IsDirect() {
				policyNamespace := handleDefaultNamespace(policyNode.Policy.TargetRef().Namespace)

				row := []string{
					// Type
					fmt.Sprintf("%v.%v", policyNode.Policy.Unstructured().GroupVersionKind().Kind, policyNode.Policy.Unstructured().GroupVersionKind().Group),
					// Name
					fmt.Sprintf("%v/%v", policyNamespace, policyNode.Policy.Unstructured().GetName()),
				}

				directlyAttachedPolicies.Rows = append(directlyAttachedPolicies.Rows, row)
			}
		}

		if len(directlyAttachedPolicies.Rows) != 0 {
			pairs = append(pairs, &DescriberKV{Key: "DirectlyAttachedPolicies", Value: directlyAttachedPolicies})
		}

		// InheritedPolicies
		if len(httpRouteNode.InheritedPolicies) != 0 {
			inheritedPolicies := &Table{
				ColumnNames:  []string{"Type", "Name", "Target Kind", "Target Name"},
				UseSeparator: true,
			}

			for _, policyNode := range httpRouteNode.InheritedPolicies {
				policyNamespace := handleDefaultNamespace(policyNode.Policy.Unstructured().GetNamespace())

				row := []string{
					// Type
					fmt.Sprintf(
						"%v.%v",
						policyNode.Policy.Unstructured().GroupVersionKind().Kind,
						policyNode.Policy.Unstructured().GroupVersionKind().Group,
					),
					// Name
					fmt.Sprintf("%v/%v", policyNamespace, policyNode.Policy.Unstructured().GetName()),
					// Target Kind
					policyNode.Policy.TargetRef().Kind,
				}

				// Target Name
				switch policyNode.Policy.TargetRef().Kind {

				case "Namespace":
					row = append(row, handleDefaultNamespace(policyNode.Policy.TargetRef().Name))

				case "GatewayClass":
					row = append(row, policyNode.Policy.TargetRef().Name)

				default:
					// handle namespaced objects
					targetRefNamespace := handleDefaultNamespace(policyNode.Policy.TargetRef().Namespace)
					name := fmt.Sprintf("%v/%v", targetRefNamespace, policyNode.Policy.TargetRef().Name)

					row = append(row, name)
				}

				// Sort inheritedPolices on the basis of Type and Name
				sort.Slice(inheritedPolicies.Rows, func(i, j int) bool {
					// Compare the Type of inheritedPolicies
					if inheritedPolicies.Rows[i][0] != inheritedPolicies.Rows[j][0] {
						return inheritedPolicies.Rows[i][0] < inheritedPolicies.Rows[j][0]
					}
					// If inheritedPolicies are of same Type, compare Names
					return inheritedPolicies.Rows[i][1] < inheritedPolicies.Rows[j][1]
				})

				inheritedPolicies.Rows = append(inheritedPolicies.Rows, row)
			}

			pairs = append(pairs, &DescriberKV{Key: "InheritedPolicies", Value: inheritedPolicies})
		}

		// EffectivePolices
		if len(httpRouteNode.EffectivePolicies) != 0 {
			pairs = append(pairs, &DescriberKV{Key: "EffectivePolicies", Value: httpRouteNode.EffectivePolicies})
		}

		// Analysis
		if len(httpRouteNode.Errors) != 0 {
			pairs = append(pairs, &DescriberKV{Key: "Analysis", Value: convertErrorsToString(httpRouteNode.Errors)})
		}

		// Events
		pairs = append(pairs, &DescriberKV{Key: "Events", Value: convertEventsSliceToTable(httpRouteNode.Events, hp.Clock)})

		Describe(hp, pairs)

		if index+1 <= len(resourceModel.HTTPRoutes) {
			fmt.Fprintf(hp, "\n\n")
		}
	}
}
