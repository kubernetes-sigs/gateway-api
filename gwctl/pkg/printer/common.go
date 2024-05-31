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
	"sort"
	"strings"
	"text/tabwriter"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/duration"
	"k8s.io/utils/clock"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/yaml"

	"sigs.k8s.io/gateway-api/gwctl/pkg/common"
)

// DescriberKV stores key-value pairs that are used with Describing a resource.
type DescriberKV struct {
	Key   string
	Value any
}

const (
	// Default indentation for Tables that are printed in the Describe view.
	defaultDescribeTableIndentSpaces = 2
)

// Describe writes the key-value paris to the writer. It handles things like
// properly writing special data types like Tables.
func Describe(w io.Writer, pairs []*DescriberKV) {
	for _, pair := range pairs {
		// If the Value is of type Table, it needs special handling.
		if table, ok := pair.Value.(*Table); ok {
			if len(table.Rows) == 0 {
				fmt.Fprintf(w, "%v: <none>\n", pair.Key)
			} else {
				fmt.Fprintf(w, "%v:\n", pair.Key)
				table.Write(w, defaultDescribeTableIndentSpaces)
			}
			continue
		}

		// If Value is NOT a Table, it can be handled through the yaml Marshaller.
		data := map[string]any{pair.Key: pair.Value}
		b, err := yaml.Marshal(data)
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to marshal to yaml: %v\n", err)
			os.Exit(1)
		}
		fmt.Fprint(w, string(b))
	}
}

type Table struct {
	ColumnNames []string
	Rows        [][]string
	// UseSeparator indicates whether the header row and data rows will be
	// separated through a separator.
	UseSeparator bool
}

// Write will write a formatted table to the writer. indent controls the
// number of spaces at the beginning of each row.
func (t *Table) Write(w io.Writer, indent int) {
	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)

	// Print column names.
	if len(t.ColumnNames) > 0 {
		row := t.indentRow(t.ColumnNames, indent)
		_, err := tw.Write([]byte(strings.Join(row, "\t") + "\n"))
		if err != nil {
			fmt.Fprint(os.Stderr, err)
			os.Exit(1)
		}
	}

	// Optionally print a separator between header row and data rows.
	if t.UseSeparator {
		row := make([]string, len(t.ColumnNames))
		for i, value := range t.ColumnNames {
			row[i] = strings.Repeat("-", len(value))
		}
		row = t.indentRow(row, indent)
		_, err := tw.Write([]byte(strings.Join(row, "\t") + "\n"))
		if err != nil {
			fmt.Fprint(os.Stderr, err)
			os.Exit(1)
		}
	}

	// Print data rows.
	for _, row := range t.Rows {
		row = t.indentRow(row, indent)
		_, err := tw.Write([]byte(strings.Join(row, "\t") + "\n"))
		if err != nil {
			fmt.Fprint(os.Stderr, err)
			os.Exit(1)
		}
	}
	tw.Flush()
}

// indentRow will add 'indent' spaces to the beginning of the row.
func (t *Table) indentRow(row []string, indent int) []string {
	if len(row) == 0 {
		return row
	}

	newRow := append([]string{}, row...)
	newRow[0] = fmt.Sprintf("%s%s", strings.Repeat(" ", indent), newRow[0])
	return newRow
}

func convertEventsSliceToTable(events []corev1.Event, clock clock.Clock) *Table {
	table := &Table{
		ColumnNames:  []string{"Type", "Reason", "Age", "From", "Message"},
		UseSeparator: true,
	}
	for _, event := range events {
		age := "Unknown"
		if !event.FirstTimestamp.IsZero() {
			age = duration.HumanDuration(clock.Since(event.FirstTimestamp.Time))
		}

		row := []string{
			event.Type,             // Type
			event.Reason,           // Reason
			age,                    // Age
			event.Source.Component, // From
			event.Message,          // Message
		}
		table.Rows = append(table.Rows, row)
	}
	return table
}

func convertPolicyRefsToTable(policyRefs []common.ObjRef) *Table {
	table := &Table{
		ColumnNames:  []string{"Type", "Name"},
		UseSeparator: true,
	}
	for _, policyRef := range policyRefs {
		name := policyRef.Name
		if policyRef.Namespace != "" {
			name = fmt.Sprintf("%v/%v", policyRef.Namespace, name)
		}
		row := []string{
			fmt.Sprintf("%v.%v", policyRef.Kind, policyRef.Group), // Type
			name, // Name
		}
		table.Rows = append(table.Rows, row)
	}
	return table
}

func convertErrorsToString(errors []error) []string {
	var result []string
	for _, err := range errors {
		result = append(result, err.Error())
	}
	return result
}

type NodeResource interface {
	ClientObject() client.Object
}

func ClientObjects[K NodeResource](nodes []K) []client.Object {
	clientObjects := make([]client.Object, len(nodes))
	for i, node := range nodes {
		clientObjects[i] = node.ClientObject()
	}
	return clientObjects
}

func SortByString[K NodeResource](items []K) []K {
	sort.Slice(items, func(i, j int) bool {
		a := client.ObjectKeyFromObject(items[i].ClientObject()).String()
		b := client.ObjectKeyFromObject(items[j].ClientObject()).String()
		return a < b
	})
	return items
}

func NodeResources[K NodeResource](items []K) []NodeResource {
	output := make([]NodeResource, len(items))
	for i, item := range items {
		output[i] = item
	}
	return output
}
