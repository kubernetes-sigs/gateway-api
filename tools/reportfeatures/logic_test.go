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
	"slices"
	"testing"
)

func setOf(names ...string) map[string]struct{} {
	set := make(map[string]struct{}, len(names))
	for _, name := range names {
		set[name] = struct{}{}
	}
	return set
}

func TestReleaseFeaturesNamesCoversBothChannels(t *testing.T) {
	t.Parallel()

	rf := ReleaseFeatures{Standard: []string{"A", "B"}, Experimental: []string{"C"}}
	names := rf.Names()
	for _, want := range []string{"A", "B", "C"} {
		if _, ok := names[want]; !ok {
			t.Errorf("Names() is missing %q", want)
		}
	}
	if len(names) != 3 {
		t.Errorf("Names() has %d entries, want 3", len(names))
	}
}

func TestInvalidNames(t *testing.T) {
	t.Parallel()

	got := invalidNames([]string{"B", "X", "A", "X", "Y"}, setOf("A", "B"))
	want := []string{"X", "Y"}
	if !slices.Equal(got, want) {
		t.Errorf("invalidNames() = %v, want %v", got, want)
	}
	if got := invalidNames([]string{"A"}, setOf("A")); len(got) != 0 {
		t.Errorf("invalidNames() = %v, want none", got)
	}
}

// report mirrors the layout of the v1.3.0 Linkerd report, which lists features
// that were only introduced in v1.4.0 as both supported and unsupported.
const report = `apiVersion: gateway.networking.k8s.io/v1
kind: ConformanceReport
profiles:
- core:
    result: success
    statistics:
      Failed: 0
      Passed: 12
      Skipped: 0
  extended:
    result: success
    statistics:
      Failed: 0
      Passed: 8
      Skipped: 0
    supportedFeatures:
    - HTTPRouteSchemeRedirect
    - MeshClusterIPMatching
    - MeshHTTPRouteQueryParamMatching
    - MeshHTTPRouteRedirectPort
    unsupportedFeatures:
    - MeshHTTPRouteRedirectPath
    - MeshHTTPRouteRewritePath
  name: MESH-HTTP
  summary: Core tests succeeded. Extended tests succeeded.
- core:
    result: success
    statistics:
      Failed: 0
      Passed: 3
      Skipped: 0
    skippedTests:
    - MeshHTTPRouteRedirectPath
  extended:
    result: success
    statistics:
      Failed: 0
      Passed: 1
      Skipped: 0
    supportedFeatures:
    - MeshHTTPRouteRedirectPath
    - MeshClusterIPMatching
  name: MESH-GRPC
  summary: Core tests succeeded. Extended tests succeeded.
`

func TestPruneFeaturesRemovesOnlyTheInvalidOnes(t *testing.T) {
	t.Parallel()

	invalid := setOf("MeshHTTPRouteQueryParamMatching", "MeshHTTPRouteRedirectPort")
	got, removed := pruneFeatures([]byte(report), invalid)

	wantRemoved := []string{"MeshHTTPRouteQueryParamMatching", "MeshHTTPRouteRedirectPort"}
	if !slices.Equal(removed, wantRemoved) {
		t.Errorf("removed = %v, want %v", removed, wantRemoved)
	}

	want := `apiVersion: gateway.networking.k8s.io/v1
kind: ConformanceReport
profiles:
- core:
    result: success
    statistics:
      Failed: 0
      Passed: 12
      Skipped: 0
  extended:
    result: success
    statistics:
      Failed: 0
      Passed: 8
      Skipped: 0
    supportedFeatures:
    - HTTPRouteSchemeRedirect
    - MeshClusterIPMatching
    unsupportedFeatures:
    - MeshHTTPRouteRedirectPath
    - MeshHTTPRouteRewritePath
  name: MESH-HTTP
  summary: Core tests succeeded. Extended tests succeeded.
- core:
    result: success
    statistics:
      Failed: 0
      Passed: 3
      Skipped: 0
    skippedTests:
    - MeshHTTPRouteRedirectPath
  extended:
    result: success
    statistics:
      Failed: 0
      Passed: 1
      Skipped: 0
    supportedFeatures:
    - MeshHTTPRouteRedirectPath
    - MeshClusterIPMatching
  name: MESH-GRPC
  summary: Core tests succeeded. Extended tests succeeded.
`
	if string(got) != want {
		t.Errorf("unexpected pruned report:\n%s", got)
	}
}

func TestPruneFeaturesDropsAnEmptiedListAndItsKey(t *testing.T) {
	t.Parallel()

	invalid := setOf("MeshHTTPRouteRedirectPath", "MeshHTTPRouteRewritePath")
	got, removed := pruneFeatures([]byte(report), invalid)

	wantRemoved := []string{"MeshHTTPRouteRedirectPath", "MeshHTTPRouteRewritePath"}
	if !slices.Equal(removed, wantRemoved) {
		t.Errorf("removed = %v, want %v", removed, wantRemoved)
	}

	// The unsupportedFeatures list of MESH-HTTP is now empty, so its key is
	// gone, and the same name in the other profile's supportedFeatures list is
	// removed as well. The skippedTests entry of the same name is a test name,
	// not a feature, so it must stay.
	want := `apiVersion: gateway.networking.k8s.io/v1
kind: ConformanceReport
profiles:
- core:
    result: success
    statistics:
      Failed: 0
      Passed: 12
      Skipped: 0
  extended:
    result: success
    statistics:
      Failed: 0
      Passed: 8
      Skipped: 0
    supportedFeatures:
    - HTTPRouteSchemeRedirect
    - MeshClusterIPMatching
    - MeshHTTPRouteQueryParamMatching
    - MeshHTTPRouteRedirectPort
  name: MESH-HTTP
  summary: Core tests succeeded. Extended tests succeeded.
- core:
    result: success
    statistics:
      Failed: 0
      Passed: 3
      Skipped: 0
    skippedTests:
    - MeshHTTPRouteRedirectPath
  extended:
    result: success
    statistics:
      Failed: 0
      Passed: 1
      Skipped: 0
    supportedFeatures:
    - MeshClusterIPMatching
  name: MESH-GRPC
  summary: Core tests succeeded. Extended tests succeeded.
`
	if string(got) != want {
		t.Errorf("unexpected pruned report:\n%s", got)
	}
}

func TestPruneFeaturesWithNothingToRemoveLeavesTheReportUntouched(t *testing.T) {
	t.Parallel()

	got, removed := pruneFeatures([]byte(report), setOf("DoesNotAppear"))
	if string(got) != report {
		t.Errorf("the report was modified:\n%s", got)
	}
	if len(removed) != 0 {
		t.Errorf("removed = %v, want none", removed)
	}
}

func TestPruneFeaturesKeepsAnAlreadyEmptyKey(t *testing.T) {
	t.Parallel()

	in := "  extended:\n    supportedFeatures:\n    - A\n    unsupportedFeatures:\n  name: X\n"
	got, removed := pruneFeatures([]byte(in), setOf("A"))
	want := "  extended:\n    unsupportedFeatures:\n  name: X\n"
	if string(got) != want {
		t.Errorf("got %q, want %q", got, want)
	}
	if !slices.Equal(removed, []string{"A"}) {
		t.Errorf("removed = %v, want [A]", removed)
	}
}

func TestPruneFeaturesHandlesIndentedItemsAndQuotes(t *testing.T) {
	t.Parallel()

	in := "    supportedFeatures:\n      - \"A\"\n      - B\n    name: X\n"
	got, _ := pruneFeatures([]byte(in), setOf("A"))
	want := "    supportedFeatures:\n      - B\n    name: X\n"
	if string(got) != want {
		t.Errorf("got %q, want %q", got, want)
	}
}
