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

func httpRows(profile string, projects ...string) []Row {
	var rows []Row
	for _, p := range projects {
		rows = append(rows, Row{
			Organization: "o", Project: p, Version: "1", Mode: "default",
			Profile: v1.ProfileReport{Name: profile},
		})
	}
	return rows
}

func TestRenderPage(t *testing.T) {
	g := group{Minor: "v1.6", Latest: "v1.6.0", Rows: httpRows("GATEWAY-HTTP", "a", "b", "c")}
	page, ok, err := RenderPage(g, 6, true)
	if err != nil || !ok {
		t.Fatalf("ok=%v err=%v", ok, err)
	}
	for _, want := range []string{"title: \"v1.6\"", "weight: 10", "toc_hide: true", "### GRPCRoute", "## Mesh Profile", "{{% alert"} {
		if !strings.Contains(page, want) {
			t.Errorf("page is missing %q", want)
		}
	}
}

func TestRenderPageTooFewProjects(t *testing.T) {
	g := group{Minor: "v1.6", Latest: "v1.6.0", Rows: httpRows("GATEWAY-HTTP", "a", "b")}
	if _, ok, _ := RenderPage(g, 6, false); ok {
		t.Error("expected the page to be skipped with only 2 projects")
	}
}

func TestRenderPageV10(t *testing.T) {
	g := group{Minor: "v1.0", Latest: "v1.0.0", Rows: httpRows("HTTP", "a", "b", "c")}
	page, ok, err := RenderPage(g, 6, false)
	if err != nil || !ok {
		t.Fatalf("ok=%v err=%v", ok, err)
	}
	if strings.Contains(page, "GRPCRoute") || strings.Contains(page, "toc_hide") {
		t.Error("v1.0 page should have no GRPC/TLS sections and should not be hidden")
	}
	if !strings.Contains(page, "weight: 70") {
		t.Error("expected weight 70 for v1.0 when the newest minor is 6")
	}
	if strings.Contains(page, "| Mode ") {
		t.Error("Mode column should be dropped for v1.0.0")
	}
}
