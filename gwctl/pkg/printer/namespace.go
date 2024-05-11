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
	"fmt"
	"io"
	"os"
	"strings"
	"text/tabwriter"

	"k8s.io/apimachinery/pkg/util/duration"
	"k8s.io/utils/clock"

	"sigs.k8s.io/gateway-api/gwctl/pkg/common"
	"sigs.k8s.io/gateway-api/gwctl/pkg/policymanager"
	"sigs.k8s.io/gateway-api/gwctl/pkg/resourcediscovery"

	"sigs.k8s.io/yaml"
)

var _ Printer = (*NamespacesPrinter)(nil)

type NamespacesPrinter struct {
	io.Writer
	Clock clock.Clock
}

type namespaceDescribeView struct {
	Name                     string                 `json:",omitempty"`
	Labels                   map[string]string      `json:",omitempty"`
	Annotations              map[string]string      `json:",omitempty"`
	Status                   string                 `json:",omitempty"`
	DirectlyAttachedPolicies []policymanager.ObjRef `json:",omitempty"`
}

func (nsp *NamespacesPrinter) GetPrintableNodes(resourceModel *resourcediscovery.ResourceModel) []NodeResource {
	return NodeResources(common.MapToValues(resourceModel.Namespaces))
}

func (nsp *NamespacesPrinter) PrintTable(resourceModel *resourcediscovery.ResourceModel) {
	tw := tabwriter.NewWriter(nsp, 0, 0, 2, ' ', 0)
	row := []string{"NAME", "STATUS", "AGE"}
	_, err := tw.Write([]byte(strings.Join(row, "\t") + "\n"))
	if err != nil {
		fmt.Fprint(os.Stderr, err)
		os.Exit(1)
	}

	namespaceNodes := common.MapToValues(resourceModel.Namespaces)
	for _, namespaceNode := range SortByString(namespaceNodes) {
		age := duration.HumanDuration(nsp.Clock.Since(namespaceNode.Namespace.CreationTimestamp.Time))
		row := []string{
			namespaceNode.Namespace.Name,
			string(namespaceNode.Namespace.Status.Phase),
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

func (nsp *NamespacesPrinter) PrintDescribeView(resourceModel *resourcediscovery.ResourceModel) {
	namespaceNodes := common.MapToValues(resourceModel.Namespaces)
	index := 0
	for _, namespaceNode := range SortByString(namespaceNodes) {
		index++

		views := []namespaceDescribeView{
			{
				Name: namespaceNode.Namespace.Name,
			},
			{
				Annotations: namespaceNode.Namespace.Annotations,
				Labels:      namespaceNode.Namespace.Labels,
			},
			{
				Status: string(namespaceNode.Namespace.Status.Phase),
			},
		}

		if policyRefs := resourcediscovery.ConvertPoliciesMapToPolicyRefs(namespaceNode.Policies); len(policyRefs) != 0 {
			views = append(views, namespaceDescribeView{
				DirectlyAttachedPolicies: policyRefs,
			})
		}

		for _, view := range views {
			b, err := yaml.Marshal(view)
			if err != nil {
				fmt.Fprintf(os.Stderr, "failed to marshal to yaml: %v\n", err)
				os.Exit(1)
			}
			fmt.Fprint(nsp, string(b))
		}

		if index+1 <= len(resourceModel.Namespaces) {
			fmt.Fprintf(nsp, "\n\n")
		}
	}
}
