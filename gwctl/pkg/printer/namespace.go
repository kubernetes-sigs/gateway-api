/*
Copyright 2024 The Kubernetes Authors.

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

var _ Printer = (*NamespacesPrinter)(nil)

type NamespacesPrinter struct {
	io.Writer
	Clock        clock.Clock
	EventFetcher eventFetcher
}

func (nsp *NamespacesPrinter) GetPrintableNodes(resourceModel *resourcediscovery.ResourceModel) []NodeResource {
	return NodeResources(maps.Values(resourceModel.Namespaces))
}

func (nsp *NamespacesPrinter) PrintTable(resourceModel *resourcediscovery.ResourceModel, wide bool) {
	var columnNames []string
	if wide {
		columnNames = []string{"NAME", "STATUS", "AGE", "POLICIES"}
	} else {
		columnNames = []string{"NAME", "STATUS", "AGE"}
	}

	table := &Table{
		ColumnNames:  columnNames,
		UseSeparator: false,
	}

	namespaceNodes := maps.Values(resourceModel.Namespaces)
	for _, namespaceNode := range SortByString(namespaceNodes) {
		age := duration.HumanDuration(nsp.Clock.Since(namespaceNode.Namespace.CreationTimestamp.Time))
		row := []string{
			namespaceNode.Namespace.Name,
			string(namespaceNode.Namespace.Status.Phase),
			age,
		}
		if wide {
			policiesCount := fmt.Sprintf("%d", len(namespaceNode.Policies))
			row = append(row, policiesCount)
		}
		table.Rows = append(table.Rows, row)
	}

	table.Write(nsp, 0)
}

func (nsp *NamespacesPrinter) PrintDescribeView(resourceModel *resourcediscovery.ResourceModel) {
	namespaceNodes := maps.Values(resourceModel.Namespaces)
	index := 0
	for _, namespaceNode := range SortByString(namespaceNodes) {
		index++

		metadata := namespaceNode.Namespace.ObjectMeta.DeepCopy()
		metadata.Labels = nil
		metadata.Annotations = nil
		metadata.Name = ""
		metadata.Namespace = ""
		metadata.ManagedFields = nil

		pairs := []*DescriberKV{
			{Key: "Name", Value: namespaceNode.Namespace.GetName()},
			{Key: "Labels", Value: namespaceNode.Namespace.Labels},
			{Key: "Annotations", Value: namespaceNode.Namespace.Annotations},
			{Key: "Status", Value: &namespaceNode.Namespace.Status},
		}

		// DirectlyAttachedPolicies
		policies := SortByString(maps.Values(namespaceNode.Policies))
		pairs = append(pairs, &DescriberKV{Key: "DirectlyAttachedPolicies", Value: convertPoliciesToRefsTable(policies, false)})

		// Events
		eventList := nsp.EventFetcher.FetchEventsFor(context.Background(), namespaceNode.Namespace)
		pairs = append(pairs, &DescriberKV{Key: "Events", Value: convertEventsSliceToTable(eventList.Items, nsp.Clock)})

		Describe(nsp, pairs)

		if index+1 <= len(resourceModel.Namespaces) {
			fmt.Fprintf(nsp, "\n\n")
		}
	}
}
