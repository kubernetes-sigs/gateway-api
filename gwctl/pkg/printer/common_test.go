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
	"bytes"
	"testing"

	"github.com/google/go-cmp/cmp"

	"sigs.k8s.io/gateway-api/gwctl/pkg/common"
)

func TestDescribe(t *testing.T) {
	pairs := []*DescriberKV{
		{Key: "Key1", Value: "string"},
		{Key: "Key2", Value: []any{"list-string1", 1234, "list-string-3"}},
		{
			Key: "Key3-with-nested-structures",
			Value: map[string]any{
				"a": "b",
				"d": map[string]any{
					"e": []string{"v1", "v2", "v3"},
				},
				"c": 123,
			},
		},
		{
			Key: "Key4-table",
			Value: &Table{
				ColumnNames: []string{"col1", "col2", "col3"},
				Rows: [][]string{
					{"row1-a", "row1-b", "row1-c"},
					{"row2-a", "row2-b", "row2-c"},
					{"row3-a", "row3-b", "row3-c"},
				},
				UseSeparator: true,
			},
		},
	}

	writable := &bytes.Buffer{}
	Describe(writable, pairs)

	got := writable.String()
	want := `
Key1: string
Key2:
- list-string1
- 1234
- list-string-3
Key3-with-nested-structures:
  a: b
  c: 123
  d:
    e:
    - v1
    - v2
    - v3
Key4-table:
  col1    col2    col3
  ----    ----    ----
  row1-a  row1-b  row1-c
  row2-a  row2-b  row2-c
  row3-a  row3-b  row3-c
`
	if diff := cmp.Diff(common.YamlString(want), common.YamlString(got), common.YamlStringTransformer); diff != "" {
		t.Errorf("Unexpected diff\ngot=\n%v\nwant=\n%v\ndiff (-want +got)=\n%v", got, want, diff)
	}
}

func TestTable_writeTable(t *testing.T) {
	testcases := []struct {
		name   string
		table  *Table
		indent int
		want   string
	}{
		{
			name: "without separator",
			table: &Table{
				ColumnNames: []string{"Kind", "Name"},
				Rows: [][]string{
					{"HTTPRoute", "default/my-httproute"},
					{"TCPRoute", "ns2/my-tcproute"},
				},
			},
			indent: 0,
			want: `
Kind       Name
HTTPRoute  default/my-httproute
TCPRoute   ns2/my-tcproute
`,
		},
		{
			name: "with separator",
			table: &Table{
				ColumnNames: []string{"Kind", "Name"},
				Rows: [][]string{
					{"HTTPRoute", "default/my-httproute"},
					{"TCPRoute", "ns2/my-tcproute"},
				},
				UseSeparator: true,
			},
			indent: 0,
			want: `
Kind       Name
----       ----
HTTPRoute  default/my-httproute
TCPRoute   ns2/my-tcproute
`,
		},
		{
			name: "with indent and separator",
			table: &Table{
				ColumnNames: []string{"Kind", "Name"},
				Rows: [][]string{
					{"HTTPRoute", "default/my-httproute"},
					{"TCPRoute", "ns2/my-tcproute"},
				},
				UseSeparator: true,
			},
			indent: 3, // We want 3 spaces at the start of each row.
			want: `
   Kind       Name
   ----       ----
   HTTPRoute  default/my-httproute
   TCPRoute   ns2/my-tcproute
`,
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			writable := &bytes.Buffer{}
			tc.table.Write(writable, tc.indent)

			got := writable.String()
			if diff := cmp.Diff(common.YamlString(tc.want), common.YamlString(got), common.YamlStringTransformer); diff != "" {
				t.Errorf("Unexpected diff\ngot=\n%v\nwant=\n%v\ndiff (-want +got)=\n%v", got, tc.want, diff)
			}
		})
	}
}
