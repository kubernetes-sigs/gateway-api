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
	"bytes"
	"embed"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"text/template"

	"golang.org/x/mod/semver"
)

//go:embed templates/page.md.tmpl
var templateFS embed.FS

// group is one minor release and the rows that belong on its page.
type group struct {
	Minor  string // e.g. "v1.6"
	Latest string // newest release dir in the group, e.g. "v1.6.0"
	Rows   []Row
}

type pageData struct {
	Title       string
	Weight      int
	Hidden      bool
	HasGRPCTLS  bool
	GatewayHTTP string
	GatewayGRPC string
	GatewayTLS  string
	MeshHTTP    string
}

// BuildGroups groups rows by minor release, keeps the newest report per
// project and removes duplicates. Groups are ordered oldest to newest.
func BuildGroups(rows []Row) []group {
	var out []group
	for minor, rs := range GroupByMinor(rows) {
		latest := rs[0].ReleaseDir
		for _, r := range rs {
			if semver.Compare(r.ReleaseDir, latest) > 0 {
				latest = r.ReleaseDir
			}
		}
		out = append(out, group{Minor: minor, Latest: latest, Rows: Dedupe(LatestPerProject(rs))})
	}
	sort.Slice(out, func(i, j int) bool {
		return semver.Compare(out[i].Latest, out[j].Latest) < 0
	})
	return out
}

func minorNumber(minor string) int {
	parts := strings.Split(strings.TrimPrefix(minor, "v"), ".")
	if len(parts) < 2 {
		return 0
	}
	n, _ := strconv.Atoi(parts[1])
	return n
}

// RenderPage renders one release page. The bool is false when the page
// should be skipped because fewer than 3 projects reported HTTP results.
func RenderPage(g group, maxMinor int, hidden bool) (string, bool, error) {
	isV10 := semver.Compare(g.Latest, "v1.1.0") < 0
	httpName, meshName := "GATEWAY-HTTP", "MESH-HTTP"
	if isV10 {
		httpName, meshName = "HTTP", "MESH"
	}
	projects := map[string]bool{}
	for _, r := range SelectProfile(g.Rows, httpName) {
		projects[r.Project] = true
	}
	if len(projects) < 3 {
		return "", false, nil
	}

	dropCore := semver.Compare(g.Latest, "v1.4.0") < 0
	dropMode := g.Latest == "v1.0.0"
	build := func(name string) string {
		return BuildTable(SelectProfile(g.Rows, name), dropCore, dropMode).Markdown()
	}
	data := pageData{
		Title:       semver.MajorMinor(g.Latest),
		Weight:      (maxMinor + 1 - minorNumber(g.Minor)) * 10,
		Hidden:      hidden,
		HasGRPCTLS:  !isV10,
		GatewayHTTP: build(httpName),
		MeshHTTP:    build(meshName),
	}
	if !isV10 {
		data.GatewayGRPC = build("GATEWAY-GRPC")
		data.GatewayTLS = build("GATEWAY-TLS")
	}

	tmpl, err := template.New("page.md.tmpl").Delims("[[", "]]").ParseFS(templateFS, "templates/page.md.tmpl")
	if err != nil {
		return "", false, fmt.Errorf("parsing template: %w", err)
	}
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", false, fmt.Errorf("rendering %s: %w", g.Minor, err)
	}
	return buf.String(), true, nil
}

// WritePages reads reportsDir and writes one page per release into outDir.
func WritePages(reportsDir, outDir string) error {
	rows, err := LoadRows(reportsDir)
	if err != nil {
		return err
	}
	groups := BuildGroups(rows)
	if len(groups) == 0 {
		return fmt.Errorf("no conformance reports found in %s", reportsDir)
	}
	if err := os.MkdirAll(outDir, 0o755); err != nil {
		return err
	}
	index := filepath.Join(outDir, "_index.md")
	if _, err := os.Stat(index); os.IsNotExist(err) {
		content := "---\ntitle: \"Implementations\"\nweight: 40\n---\nConformance data for Gateway API Implementations.\n"
		if err := os.WriteFile(index, []byte(content), 0o600); err != nil {
			return err
		}
	}
	maxMinor := minorNumber(groups[len(groups)-1].Minor)
	for i, g := range groups {
		page, ok, err := RenderPage(g, maxMinor, i < len(groups)-4)
		if err != nil {
			return err
		}
		if !ok {
			log.Printf("skipping %s: fewer than 3 projects", g.Minor)
			continue
		}
		path := filepath.Join(outDir, g.Minor+".md")
		if err := os.WriteFile(path, []byte(page), 0o600); err != nil {
			return err
		}
		log.Printf("generated %s", path)
	}
	return nil
}
