/*
Copyright The Kubernetes Authors.

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

package main

import (
	"strings"
	"testing"

	v1 "sigs.k8s.io/gateway-api/conformance/apis/v1"
)

func mkRow(org, project, version string, p v1.ProfileReport) Row {
	return Row{Organization: org, Project: project, Version: version, Mode: "default", Profile: p}
}

func TestFeaturesCell(t *testing.T) {
	tests := []struct {
		name string
		p    v1.ProfileReport
		want string
	}{
		{"nil extended", v1.ProfileReport{}, "0/0 features"},
		{
			"plain",
			v1.ProfileReport{Extended: &v1.ExtendedStatus{
				SupportedFeatures:   []string{"A", "B"},
				UnsupportedFeatures: []string{"C"},
			}},
			"2/3 features",
		},
		{
			"partial",
			v1.ProfileReport{Extended: &v1.ExtendedStatus{
				Result: v1.Partial, Statistics: v1.Statistics{Passed: 16, Skipped: 1},
				SupportedFeatures: []string{"A"},
			}},
			"1/1 features<br><b>Partially conformant; 1/17 tests skipped</b>",
		},
		{
			"failing core",
			v1.ProfileReport{
				Core: v1.Status{Result: v1.Failure, Statistics: v1.Statistics{Passed: 8, Failed: 2}},
			},
			"<b>Failing 2/10 core tests</b><br>0/0 features",
		},
		{
			"failing extended",
			v1.ProfileReport{Extended: &v1.ExtendedStatus{
				Result: v1.Failure, Statistics: v1.Statistics{Passed: 3, Failed: 1},
			}},
			"0/0 features<br><b>Failing 1/4 tests</b>",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := featuresCell(tt.p); got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}
}

func TestBuildTable(t *testing.T) {
	ok := v1.Status{Result: v1.Success}
	rows := []Row{
		mkRow("zeta", "z", "1.0", v1.ProfileReport{
			Core: ok,
			Extended: &v1.ExtendedStatus{
				SupportedFeatures:   []string{"HTTPRouteCORS"},
				UnsupportedFeatures: []string{"GatewayPort8080"},
			},
		}),
		mkRow("Alpha", "a", "1.0", v1.ProfileReport{Core: v1.Status{Result: v1.Partial}}),
	}
	tbl := BuildTable(rows, false, false)

	wantHeader := "Organization|Project|Version|Mode|Extended Features|Core|Gateway Port 8080|HTTPRoute CORS"
	if got := strings.Join(tbl.Header, "|"); got != wantHeader {
		t.Fatalf("header = %q, want %q", got, wantHeader)
	}
	if len(tbl.Rows) != 2 || tbl.Rows[0][0] != "Alpha" || tbl.Rows[1][0] != "zeta" {
		t.Fatalf("rows not sorted case-insensitively: %v", tbl.Rows)
	}
	// Alpha: partial core -> cross, nothing supported.
	if tbl.Rows[0][5] != no || tbl.Rows[0][6] != no || tbl.Rows[0][7] != no {
		t.Errorf("unexpected Alpha row: %v", tbl.Rows[0])
	}
	// zeta: core ok, CORS supported, Port8080 unsupported.
	if tbl.Rows[1][5] != yes || tbl.Rows[1][6] != no || tbl.Rows[1][7] != yes {
		t.Errorf("unexpected zeta row: %v", tbl.Rows[1])
	}
}

func TestBuildTableDropColumns(t *testing.T) {
	rows := []Row{mkRow("a", "p", "1.0", v1.ProfileReport{})}
	got := strings.Join(BuildTable(rows, true, true).Header, "|")
	if got != "Organization|Project|Version|Extended Features" {
		t.Errorf("header = %q", got)
	}
}

func TestSortRowsVersion(t *testing.T) {
	rows := []Row{
		mkRow("a", "p", "1.10", v1.ProfileReport{}),
		mkRow("a", "p", "1.9", v1.ProfileReport{}),
	}
	SortRows(rows)
	if rows[0].Version != "1.9" || rows[1].Version != "1.10" {
		t.Errorf("want 1.9 before 1.10, got %s, %s", rows[0].Version, rows[1].Version)
	}
}

func TestMarkdown(t *testing.T) {
	tbl := Table{Header: []string{"A", "B"}, Rows: [][]string{{"1", "2"}}}
	want := "| A | B |\n| :--- | :--- |\n| 1 | 2 |\n"
	if got := tbl.Markdown(); got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}
