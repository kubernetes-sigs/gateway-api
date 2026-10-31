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
	"io/fs"
	"os"
	"path"
	"strings"

	"golang.org/x/mod/semver"
	"sigs.k8s.io/yaml"

	v1 "sigs.k8s.io/gateway-api/conformance/apis/v1"
)

// Row is one profile from one report, flattened for the tables.
type Row struct {
	Organization string
	Project      string
	Version      string
	Mode         string
	// ReleaseDir is the directory under conformance/reports, e.g. "v1.4.1".
	ReleaseDir string
	// File is the report path relative to the reports dir, used for
	// deterministic tie-breaks.
	File    string
	Profile v1.ProfileReport
}

// LoadRows reads every report at or above v1.0.0 and returns one Row per profile.
func LoadRows(reportsDir string) ([]Row, error) {
	var rows []Row
	rootFS := os.DirFS(reportsDir)
	err := fs.WalkDir(rootFS, ".", func(rel string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || path.Ext(d.Name()) != ".yaml" {
			return nil
		}
		dir, _, found := strings.Cut(rel, "/")
		if !found || !semver.IsValid(dir) || semver.Compare(dir, "v1.0.0") < 0 {
			return nil
		}
		content, err := fs.ReadFile(rootFS, rel)
		if err != nil {
			return err
		}
		var report v1.ConformanceReport
		if err := yaml.Unmarshal(content, &report); err != nil {
			return fmt.Errorf("%s: %w", rel, err)
		}
		for _, p := range report.ProfileReports {
			rows = append(rows, Row{
				Organization: report.Organization,
				Project:      report.Project,
				Version:      report.Version,
				Mode:         report.Mode,
				ReleaseDir:   dir,
				File:         rel,
				Profile:      p,
			})
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return rows, nil
}

// GroupByMinor groups rows by the minor release of their directory
// (v1.5, v1.5.0 and v1.5.1 all land in "v1.5").
func GroupByMinor(rows []Row) map[string][]Row {
	out := map[string][]Row{}
	for _, r := range rows {
		m := semver.MajorMinor(r.ReleaseDir)
		out[m] = append(out[m], r)
	}
	return out
}

// LatestPerProject keeps, for each organization/project, only the rows from
// that project's newest release directory.
func LatestPerProject(rows []Row) []Row {
	latest := map[string]string{}
	for _, r := range rows {
		k := r.Organization + "/" + r.Project
		if cur, ok := latest[k]; !ok || semver.Compare(r.ReleaseDir, cur) > 0 {
			latest[k] = r.ReleaseDir
		}
	}
	var out []Row
	for _, r := range rows {
		if semver.Compare(r.ReleaseDir, latest[r.Organization+"/"+r.Project]) == 0 {
			out = append(out, r)
		}
	}
	return out
}

func supportedCount(r Row) int {
	if r.Profile.Extended == nil {
		return 0
	}
	return len(r.Profile.Extended.SupportedFeatures)
}

// Dedupe keeps one row per organization, project, version, profile and mode.
// On a collision the row with more supported features wins, then the
// lexicographically smaller File, so the result does not depend on walk order.
func Dedupe(rows []Row) []Row {
	idx := map[string]int{}
	var out []Row
	for _, r := range rows {
		k := strings.Join([]string{r.Organization, r.Project, r.Version, r.Profile.Name, r.Mode}, "\x00")
		i, ok := idx[k]
		if !ok {
			idx[k] = len(out)
			out = append(out, r)
			continue
		}
		cur := out[i]
		if supportedCount(r) > supportedCount(cur) ||
			(supportedCount(r) == supportedCount(cur) && r.File < cur.File) {
			out[i] = r
		}
	}
	return out
}
