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
	"os"
	"path/filepath"
	"strings"
	"testing"

	v1 "sigs.k8s.io/gateway-api/conformance/apis/v1"
)

func writeReport(t *testing.T, root, rel, org, project, version, mode, profile string, supported ...string) {
	t.Helper()
	var sb strings.Builder
	fmt.Fprintf(&sb, "implementation:\n  organization: %s\n  project: %s\n  version: %s\nmode: %s\nprofiles:\n- name: %s\n",
		org, project, version, mode, profile)
	if len(supported) > 0 {
		sb.WriteString("  extended:\n    supportedFeatures:\n")
		for _, f := range supported {
			fmt.Fprintf(&sb, "    - %s\n", f)
		}
	}
	p := filepath.Join(root, rel)
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(p, []byte(sb.String()), 0o600); err != nil {
		t.Fatal(err)
	}
}

func TestLoadRows(t *testing.T) {
	root := t.TempDir()
	writeReport(t, root, "v1.4.0/acme/a.yaml", "acme", "gw", "1.0", "default", "GATEWAY-HTTP", "FeatureA")
	writeReport(t, root, "v1.4.1/acme/b.yaml", "acme", "gw", "1.1", "default", "GATEWAY-HTTP")
	writeReport(t, root, "v0.8.0/acme/old.yaml", "acme", "gw", "0.1", "default", "HTTP")
	if err := os.WriteFile(filepath.Join(root, "v1.4.0", "README.md"), []byte("x"), 0o600); err != nil {
		t.Fatal(err)
	}

	rows, err := LoadRows(root)
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 2 {
		t.Fatalf("got %d rows, want 2: %+v", len(rows), rows)
	}
	r := rows[0]
	if r.Organization != "acme" || r.Project != "gw" || r.ReleaseDir != "v1.4.0" ||
		r.Mode != "default" || r.Profile.Name != "GATEWAY-HTTP" {
		t.Errorf("unexpected first row: %+v", r)
	}
}

func TestGroupByMinor(t *testing.T) {
	rows := []Row{{ReleaseDir: "v1.5"}, {ReleaseDir: "v1.5.0"}, {ReleaseDir: "v1.5.1"}, {ReleaseDir: "v1.6"}}
	got := GroupByMinor(rows)
	if len(got) != 2 || len(got["v1.5"]) != 3 || len(got["v1.6"]) != 1 {
		t.Errorf("unexpected grouping: %+v", got)
	}
}

func TestLatestPerProject(t *testing.T) {
	rows := []Row{
		{Organization: "a", Project: "x", ReleaseDir: "v1.4.0"},
		{Organization: "a", Project: "x", ReleaseDir: "v1.4.1"},
		{Organization: "b", Project: "y", ReleaseDir: "v1.4.0"},
		{Organization: "c", Project: "z", ReleaseDir: "v1.5"},
		{Organization: "c", Project: "z", ReleaseDir: "v1.5.0"},
	}
	got := LatestPerProject(rows)
	if len(got) != 4 || got[0].ReleaseDir != "v1.4.1" || got[1].Project != "y" {
		t.Errorf("unexpected result: %+v", got)
	}
}

func TestDedupe(t *testing.T) {
	mk := func(file string, features ...string) Row {
		return Row{
			Organization: "a", Project: "x", Version: "1", Mode: "default", File: file,
			Profile: v1.ProfileReport{
				Name:     "GATEWAY-HTTP",
				Extended: &v1.ExtendedStatus{SupportedFeatures: features},
			},
		}
	}
	got := Dedupe([]Row{
		mk("v1.5/a/standard.yaml"),
		mk("v1.5/a/experimental.yaml", "F1", "F2"),
	})
	if len(got) != 1 || got[0].File != "v1.5/a/experimental.yaml" {
		t.Errorf("more features should win: %+v", got)
	}
	got = Dedupe([]Row{mk("v1.5/a/z.yaml", "F1"), mk("v1.5/a/b.yaml", "F1")})
	if len(got) != 1 || got[0].File != "v1.5/a/b.yaml" {
		t.Errorf("tie should go to smaller File: %+v", got)
	}
}
