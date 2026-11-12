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
	"fmt"
	"sort"
	"strings"

	"golang.org/x/mod/semver"

	v1 "sigs.k8s.io/gateway-api/conformance/apis/v1"
)

const (
	yes = "✅"
	no  = "❌"
)

// Table is a rendered markdown table: a header and rows of equal width.
type Table struct {
	Header []string
	Rows   [][]string
}

// SelectProfile returns the rows whose profile has the given name.
func SelectProfile(rows []Row, name string) []Row {
	var out []Row
	for _, r := range rows {
		if r.Profile.Name == name {
			out = append(out, r)
		}
	}
	return out
}

func versionLess(a, b string) bool {
	if semver.IsValid(a) && semver.IsValid(b) {
		if c := semver.Compare(a, b); c != 0 {
			return c < 0
		}
		return false
	}
	return a < b
}

// SortRows orders rows by organization (case-insensitive), version, mode, file.
func SortRows(rows []Row) {
	sort.SliceStable(rows, func(i, j int) bool {
		a, b := rows[i], rows[j]
		if x, y := strings.ToLower(a.Organization), strings.ToLower(b.Organization); x != y {
			return x < y
		}
		if a.Organization != b.Organization {
			return a.Organization < b.Organization
		}
		if a.Project != b.Project {
			return a.Project < b.Project
		}
		if a.Version != b.Version {
			return versionLess(canonVersion(a.Version), canonVersion(b.Version))
		}
		if a.Mode != b.Mode {
			return a.Mode < b.Mode
		}
		return a.File < b.File
	})
}

// canonVersion prefixes "v" so implementation versions like "1.31" can be
// compared as semver.
func canonVersion(v string) string {
	if strings.HasPrefix(v, "v") {
		return v
	}
	return "v" + v
}

func coreCell(s v1.Status) string {
	if s.Result == v1.Success {
		return yes
	}
	return no
}

func total(s v1.Statistics) uint32 { return s.Passed + s.Failed + s.Skipped }

// featuresCell builds the "Extended Features" cell.
func featuresCell(p v1.ProfileReport) string {
	var lines []string
	if p.Core.Failed > 0 && total(p.Core.Statistics) > 0 {
		lines = append(lines, fmt.Sprintf("<b>Failing %d/%d core tests</b>", p.Core.Failed, total(p.Core.Statistics)))
	}
	var supported, unsupported int
	var extTotal, extFailed, extSkipped uint32
	var extResult v1.Result
	if p.Extended != nil {
		supported = len(p.Extended.SupportedFeatures)
		unsupported = len(p.Extended.UnsupportedFeatures)
		extTotal = total(p.Extended.Statistics)
		extFailed = p.Extended.Failed
		extSkipped = p.Extended.Skipped
		extResult = p.Extended.Result
	}
	lines = append(lines, fmt.Sprintf("%d/%d features", supported, supported+unsupported))
	if extResult == v1.Partial && extTotal > 0 {
		lines = append(lines, fmt.Sprintf("<b>Partially conformant; %d/%d tests skipped</b>", extSkipped, extTotal))
	} else if extResult == v1.Failure && extTotal > 0 {
		lines = append(lines, fmt.Sprintf("<b>Failing %d/%d tests</b>", extFailed, extTotal))
	}
	return strings.Join(lines, "<br>")
}

// BuildTable renders already-selected rows (one profile) as a Table.
func BuildTable(rows []Row, dropCore, dropMode bool) Table {
	rows = append([]Row(nil), rows...)
	SortRows(rows)

	names := map[string]bool{}
	for _, r := range rows {
		if r.Profile.Extended == nil {
			continue
		}
		for _, f := range r.Profile.Extended.SupportedFeatures {
			names[ProcessFeatureName(f)] = true
		}
		for _, f := range r.Profile.Extended.UnsupportedFeatures {
			names[ProcessFeatureName(f)] = true
		}
	}
	var features []string
	for n := range names {
		features = append(features, n)
	}
	sort.Strings(features)

	header := []string{"Organization", "Project", "Version"}
	if !dropMode {
		header = append(header, "Mode")
	}
	header = append(header, "Extended Features")
	if !dropCore {
		header = append(header, "Core")
	}
	header = append(header, features...)

	t := Table{Header: header}
	for _, r := range rows {
		supported := map[string]bool{}
		if r.Profile.Extended != nil {
			for _, f := range r.Profile.Extended.SupportedFeatures {
				supported[ProcessFeatureName(f)] = true
			}
		}
		row := []string{r.Organization, r.Project, r.Version}
		if !dropMode {
			row = append(row, r.Mode)
		}
		row = append(row, featuresCell(r.Profile))
		if !dropCore {
			row = append(row, coreCell(r.Profile.Core))
		}
		for _, f := range features {
			if supported[f] {
				row = append(row, yes)
			} else {
				row = append(row, no)
			}
		}
		t.Rows = append(t.Rows, row)
	}
	return t
}

// Markdown renders the table as a pipe table.
func (t Table) Markdown() string {
	var sb strings.Builder
	writeRow := func(cells []string) {
		sb.WriteString("| " + strings.Join(cells, " | ") + " |\n")
	}
	writeRow(t.Header)
	sep := make([]string, len(t.Header))
	for i := range sep {
		sep[i] = ":---"
	}
	writeRow(sep)
	for _, r := range t.Rows {
		writeRow(r)
	}
	return sb.String()
}
