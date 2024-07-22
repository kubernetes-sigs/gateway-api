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
	"strings"

	"golang.org/x/exp/maps"
	"k8s.io/apimachinery/pkg/util/duration"
	"k8s.io/utils/clock"

	"sigs.k8s.io/gateway-api/gwctl/pkg/resourcediscovery"
)

var _ Printer = (*HTTPRoutesPrinter)(nil)

type HTTPRoutesPrinter struct {
	io.Writer
	Clock        clock.Clock
	EventFetcher eventFetcher
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
		metadata.Labels = nil
		metadata.Annotations = nil
		metadata.Name = ""
		metadata.Namespace = ""
		metadata.ManagedFields = nil

		pairs := []*DescriberKV{
			{"Name", httpRouteNode.HTTPRoute.GetName()},
			{"Namespace", httpRouteNode.HTTPRoute.Namespace},
			{"Label", httpRouteNode.HTTPRoute.Labels},
			{"Annotations", httpRouteNode.HTTPRoute.Annotations},
			{"APIVersion", httpRouteNode.HTTPRoute.APIVersion},
			{"Kind", httpRouteNode.HTTPRoute.Kind},
			{"Metadata", metadata},
			{"Spec", httpRouteNode.HTTPRoute.Spec},
			{"Status", httpRouteNode.HTTPRoute.Status},
		}

		// DirectlyAttachedPolicies
		policies := SortByString(maps.Values(httpRouteNode.Policies))
		pairs = append(pairs, &DescriberKV{Key: "DirectlyAttachedPolicies", Value: convertPoliciesToRefsTable(policies, false)})

		// InheritedPolicies
		inheritedPolicies := SortByString(maps.Values(httpRouteNode.InheritedPolicies))
		pairs = append(pairs, &DescriberKV{Key: "InheritedPolicies", Value: convertPoliciesToRefsTable(inheritedPolicies, true)})

		// EffectivePolices
		if len(httpRouteNode.EffectivePolicies) != 0 {
			pairs = append(pairs, &DescriberKV{Key: "EffectivePolicies", Value: httpRouteNode.EffectivePolicies})
		}

		// Analysis
		if len(httpRouteNode.Errors) != 0 {
			pairs = append(pairs, &DescriberKV{Key: "Analysis", Value: convertErrorsToString(httpRouteNode.Errors)})
		}

		// Events
		eventList := hp.EventFetcher.FetchEventsFor(context.Background(), httpRouteNode.HTTPRoute)
		pairs = append(pairs, &DescriberKV{Key: "Events", Value: convertEventsSliceToTable(eventList.Items, hp.Clock)})

		Describe(hp, pairs)

		if index+1 <= len(resourceModel.HTTPRoutes) {
			fmt.Fprintf(hp, "\n\n")
		}
	}
}
