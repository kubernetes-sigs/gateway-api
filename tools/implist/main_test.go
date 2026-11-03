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
	"os"
	"path/filepath"
	"slices"
	"testing"
)

func TestReadReportsFiltersUnsupportedVersions(t *testing.T) {
	t.Parallel()

	// Derive the versions to check from the provided versions.
	allVersions := ConformantVersions
	allVersions = append(allVersions, StaleVersions...)
	allVersions = append(allVersions, NonConformantVersion)

	slices.Reverse(allVersions)

	reportPassVersions := ConformantVersions
	reportPassVersions = append(reportPassVersions, StaleVersions...)

	slices.Reverse(reportPassVersions)

	reportsDir := t.TempDir()
	for _, version := range allVersions {
		versionDir := filepath.Join(reportsDir, version)
		if err := os.Mkdir(versionDir, 0o755); err != nil {
			t.Fatalf("failed to create report directory: %v", err)
		}
		if err := os.WriteFile(filepath.Join(versionDir, "report.yaml"), []byte("{}"), 0o600); err != nil {
			t.Fatalf("failed to write report: %v", err)
		}
	}

	reports, err := readReports(reportsDir)
	if err != nil {
		t.Fatalf("readReports returned error: %v", err)
	}

	if got, want := len(reports), 3; got != want {
		t.Fatalf("readReports returned %d reports, want %d", got, want)
	}
	if _, ok := reports[filepath.Join(NonConformantVersion, "report.yaml")]; ok {
		t.Error("readReports included a report from an unsupported version")
	}
	for _, version := range reportPassVersions {
		if _, ok := reports[filepath.Join(version, "report.yaml")]; !ok {
			t.Errorf("readReports omitted report from %s", version)
		}
	}
}
