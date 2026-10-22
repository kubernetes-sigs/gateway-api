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

	"k8s.io/apimachinery/pkg/util/sets"
)

func TestReleaseFeaturesNamesCoversBothChannels(t *testing.T) {
	t.Parallel()

	rf := ReleaseFeatures{Standard: []string{"A", "B"}, Experimental: []string{"C"}}
	names := rf.Names()

	for _, want := range []string{"A", "B", "C"} {
		if !names.Has(want) {
			t.Errorf("Names() is missing %q", want)
		}
	}
	if names.Len() != 3 {
		t.Errorf("Names() has %d entries, want 3", names.Len())
	}
}

func TestInvalidNames(t *testing.T) {
	t.Parallel()

	got := invalidNames([]string{"B", "X", "A", "X", "Y"}, sets.New("A", "B"))
	want := []string{"X", "Y"}
	if !slices.Equal(got, want) {
		t.Errorf("invalidNames() = %v, want %v", got, want)
	}

	if got := invalidNames([]string{"A"}, sets.New("A")); len(got) != 0 {
		t.Errorf("invalidNames() = %v, want none", got)
	}
}
