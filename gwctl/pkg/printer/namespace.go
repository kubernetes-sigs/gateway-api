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

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/duration"

	"sigs.k8s.io/gateway-api/gwctl/pkg/extension/directlyattachedpolicy"
	"sigs.k8s.io/gateway-api/gwctl/pkg/policymanager"
	"sigs.k8s.io/gateway-api/gwctl/pkg/topology"
)

func (p *TablePrinter) printNamespace(namespaceNode *topology.Node, w io.Writer) error {
	if err := p.checkTypeChange("Namespace", w); err != nil {
		return err
	}

	if p.table == nil {
		var columnNames []string
		if p.OutputFormat == OutputFormatWide {
			columnNames = []string{"NAME", "STATUS", "AGE", "POLICIES"}
		} else {
			columnNames = []string{"NAME", "STATUS", "AGE"}
		}
		p.table = &Table{
			ColumnNames:  columnNames,
			UseSeparator: false,
		}
	}

	ns := topology.MustAccessObject(namespaceNode, &corev1.Namespace{})

	age := "<unknown>"
	creationTimestamp := ns.GetCreationTimestamp()
	if !creationTimestamp.IsZero() {
		age = duration.HumanDuration(p.Clock.Since(creationTimestamp.Time))
	}

	row := []string{
		ns.Name,
		string(ns.Status.Phase),
		age,
	}
	if p.OutputFormat == OutputFormatWide {
		policiesMap, err := directlyattachedpolicy.Access(namespaceNode)
		if err != nil {
			return err
		}
		policiesCount := fmt.Sprintf("%d", len(policiesMap))
		row = append(row, policiesCount)
	}
	p.table.Rows = append(p.table.Rows, row)
	return nil
}

func (p *DescriptionPrinter) printNamespace(namespaceNode *topology.Node, w io.Writer) error {
	if p.printSeparator {
		fmt.Fprintf(w, "\n\n")
	}
	p.printSeparator = true

	namespace := topology.MustAccessObject(namespaceNode, &corev1.Namespace{})

	metadata := namespace.ObjectMeta.DeepCopy()
	metadata.Labels = nil
	metadata.Annotations = nil
	metadata.Name = ""
	metadata.Namespace = ""
	metadata.ManagedFields = nil

	pairs := []*DescriberKV{
		{Key: "Name", Value: namespace.GetName()},
		{Key: "Labels", Value: namespace.Labels},
		{Key: "Annotations", Value: namespace.Annotations},
		{Key: "Status", Value: &namespace.Status},
	}

	// DirectlyAttachedPolicies
	policiesMap, err := directlyattachedpolicy.Access(namespaceNode)
	if err != nil {
		return err
	}
	policies := policymanager.ConvertPoliciesMapToSlice(policiesMap)
	pairs = append(pairs, &DescriberKV{Key: "DirectlyAttachedPolicies", Value: convertPoliciesToRefsTable(policies, false)})

	// Events
	events, err := p.EventFetcher.FetchEventsFor(namespace)
	if err != nil {
		return err
	}
	pairs = append(pairs, &DescriberKV{Key: "Events", Value: convertEventsSliceToTable(events, p.Clock)})

	Describe(w, pairs)
	return nil
}
