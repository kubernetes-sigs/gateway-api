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
	"strings"
	"testing"
)

const snapshotV13 = `standard:
- GatewayPort8080
- HTTPRouteSchemeRedirect
experimental:
- HTTPRouteBackendTimeout
`

const snapshotV14 = `standard:
- GatewayPort8080
- HTTPRouteSchemeRedirect
- MeshHTTPRouteRedirectPort
experimental:
- HTTPRouteBackendTimeout
`

func reportWith(supported, unsupported string) string {
	return `apiVersion: gateway.networking.k8s.io/v1
kind: ConformanceReport
gatewayAPIVersion: v1.3.0
profiles:
- core:
    result: success
    statistics:
      Failed: 0
      Passed: 1
      Skipped: 0
  extended:
    result: success
    statistics:
      Failed: 0
      Passed: 1
      Skipped: 0
    supportedFeatures:
` + supported + `    unsupportedFeatures:
` + unsupported + `  name: MESH-HTTP
  summary: Core tests succeeded. Extended tests succeeded.
`
}

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o750); err != nil {
		t.Fatalf("failed to create directory: %v", err)
	}
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("failed to write %s: %v", path, err)
	}
}

// setup creates snapshots for v1.3 and v1.4, and returns the config to use.
func setup(t *testing.T) config {
	t.Helper()

	dir := t.TempDir()
	cfg := config{
		ReportsDir:   filepath.Join(dir, "reports"),
		SnapshotsDir: filepath.Join(dir, "release-features"),
	}
	writeFile(t, filepath.Join(cfg.SnapshotsDir, "v1.3.yaml"), snapshotV13)
	writeFile(t, filepath.Join(cfg.SnapshotsDir, "v1.4.yaml"), snapshotV14)
	return cfg
}

func TestVerifyReportsFlagsFeaturesFromALaterRelease(t *testing.T) {
	t.Parallel()
	cfg := setup(t)

	// MeshHTTPRouteRedirectPort was introduced in v1.4, so it is not valid in
	// a v1.3 report, whichever list it is in. Experimental features are valid
	// in any report of the release.
	writeFile(t, filepath.Join(cfg.ReportsDir, "v1.3.0", "impl", "bad.yaml"),
		reportWith("    - HTTPRouteSchemeRedirect\n    - MeshHTTPRouteRedirectPort\n", "    - HTTPRouteBackendTimeout\n    - Nonexistent\n"))
	writeFile(t, filepath.Join(cfg.ReportsDir, "v1.3.0", "impl", "good.yaml"),
		reportWith("    - HTTPRouteSchemeRedirect\n", "    - HTTPRouteBackendTimeout\n"))
	writeFile(t, filepath.Join(cfg.ReportsDir, "v1.4.0", "impl", "good.yaml"),
		reportWith("    - MeshHTTPRouteRedirectPort\n", "    - HTTPRouteBackendTimeout\n"))

	problems, err := verifyReports(cfg)
	if err != nil {
		t.Fatalf("verifyReports returned error: %v", err)
	}
	if len(problems) != 1 {
		t.Fatalf("verifyReports returned %d problems, want 1: %v", len(problems), problems)
	}
	for _, want := range []string{"v1.3.0/impl/bad.yaml", "v1.3", "MeshHTTPRouteRedirectPort, Nonexistent"} {
		if !strings.Contains(filepath.ToSlash(problems[0]), want) {
			t.Errorf("problem %q does not mention %q", problems[0], want)
		}
	}
}

func TestVerifyReportsSkipsSymlinkedReleaseDirectories(t *testing.T) {
	t.Parallel()
	cfg := setup(t)

	writeFile(t, filepath.Join(cfg.ReportsDir, "v1.3", "impl", "bad.yaml"),
		reportWith("    - Nonexistent\n", "    - HTTPRouteBackendTimeout\n"))
	if err := os.Symlink("v1.3", filepath.Join(cfg.ReportsDir, "v1.3.0")); err != nil {
		t.Skipf("symlinks are not supported: %v", err)
	}

	problems, err := verifyReports(cfg)
	if err != nil {
		t.Fatalf("verifyReports returned error: %v", err)
	}
	if len(problems) != 1 {
		t.Fatalf("verifyReports returned %d problems, want 1 (the link must not be verified twice): %v", len(problems), problems)
	}
}

func TestVerifyReportsSkipsReleasesOlderThanTheOldestSnapshot(t *testing.T) {
	t.Parallel()
	cfg := setup(t)

	writeFile(t, filepath.Join(cfg.ReportsDir, "v1.2.0", "impl", "old.yaml"),
		reportWith("    - Nonexistent\n", "    - HTTPRouteBackendTimeout\n"))

	problems, err := verifyReports(cfg)
	if err != nil {
		t.Fatalf("verifyReports returned error: %v", err)
	}
	if len(problems) != 0 {
		t.Fatalf("verifyReports returned problems for an unverifiable release: %v", problems)
	}
}

func TestVerifyReportsFailsWhenASnapshotIsMissing(t *testing.T) {
	t.Parallel()
	cfg := setup(t)

	// v1.3 and v1.5 have snapshots, but v1.4 does not.
	if err := os.Remove(filepath.Join(cfg.SnapshotsDir, "v1.4.yaml")); err != nil {
		t.Fatal(err)
	}
	writeFile(t, filepath.Join(cfg.SnapshotsDir, "v1.5.yaml"), snapshotV14)
	writeFile(t, filepath.Join(cfg.ReportsDir, "v1.4.0", "impl", "report.yaml"),
		reportWith("    - GatewayPort8080\n", "    - HTTPRouteBackendTimeout\n"))

	if _, err := verifyReports(cfg); err == nil {
		t.Fatal("verifyReports succeeded, want an error about the missing v1.4 snapshot")
	} else if !strings.Contains(err.Error(), "v1.4") {
		t.Errorf("error %q does not mention v1.4", err)
	}
}

func TestPruneRemovesInvalidFeaturesAndLeavesAValidTree(t *testing.T) {
	t.Parallel()
	cfg := setup(t)
	cfg.Prune = true

	path := filepath.Join(cfg.ReportsDir, "v1.3.0", "impl", "bad.yaml")
	writeFile(t, path, reportWith("    - HTTPRouteSchemeRedirect\n    - MeshHTTPRouteRedirectPort\n", "    - Nonexistent\n"))

	problems, err := verifyReports(cfg)
	if err != nil {
		t.Fatalf("verifyReports returned error: %v", err)
	}
	if len(problems) != 0 {
		t.Fatalf("prune mode must not report problems, got %v", problems)
	}

	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	want := reportWith("    - HTTPRouteSchemeRedirect\n", "")
	want = strings.Replace(want, "    unsupportedFeatures:\n", "", 1)
	if string(got) != want {
		t.Errorf("unexpected pruned report:\n%s\nwant:\n%s", got, want)
	}

	cfg.Prune = false
	problems, err = verifyReports(cfg)
	if err != nil || len(problems) != 0 {
		t.Fatalf("the pruned tree should verify cleanly, got problems=%v err=%v", problems, err)
	}
}

func TestValidFeaturesUsesCurrentFeaturesForANewerRelease(t *testing.T) {
	t.Parallel()
	cfg := setup(t)

	snapshots, err := loadSnapshots(cfg.SnapshotsDir)
	if err != nil {
		t.Fatal(err)
	}
	names, verifiable, err := validFeatures("v1.9", snapshots)
	if err != nil {
		t.Fatal(err)
	}
	if !verifiable || len(names) == 0 {
		t.Errorf("a release newer than every snapshot should be verified against pkg/features, got verifiable=%v with %d names", verifiable, len(names))
	}
}

func TestGenerateSnapshotRoundTrips(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()

	if err := generateSnapshot(dir, "v1.6.0"); err == nil {
		t.Error("generateSnapshot accepted v1.6.0, want an error since only vMAJOR.MINOR is valid")
	}
	if err := generateSnapshot(dir, "v1.6"); err != nil {
		t.Fatalf("generateSnapshot returned error: %v", err)
	}

	content, err := os.ReadFile(filepath.Join(dir, "v1.6.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(string(content), "# This file lists the features available in Gateway API v1.6.") {
		t.Errorf("the snapshot has no header:\n%s", content)
	}

	snapshots, err := loadSnapshots(dir)
	if err != nil {
		t.Fatal(err)
	}
	current, err := currentFeatures()
	if err != nil {
		t.Fatal(err)
	}
	got := snapshots["v1.6"]
	if len(got.Standard) != len(current.Standard) || len(got.Experimental) != len(current.Experimental) {
		t.Errorf("the snapshot has %d/%d features, pkg/features has %d/%d",
			len(got.Standard), len(got.Experimental), len(current.Standard), len(current.Experimental))
	}
}
