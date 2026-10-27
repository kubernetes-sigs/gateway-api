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

	"k8s.io/apimachinery/pkg/util/sets"
)

// ReleaseFeatures lists the names of the features that are available in a
// single Gateway API release, grouped by release channel.
type ReleaseFeatures struct {
	Standard     []string `json:"standard"`
	Experimental []string `json:"experimental"`
}

// Names returns the set of all the feature names available in the release,
// across both channels.
//
// Conformance reports for the standard channel routinely list experimental
// features (mostly as unsupported), so a report is validated against both channels
// rather than only against the channel it was generated for.
func (r ReleaseFeatures) Names() sets.Set[string] {
	return sets.New(r.Standard...).Insert(r.Experimental...)
}

// sortedUnique returns a sorted copy of in without duplicates. The result is
// never nil so that it is always serialized as a list.
func sortedUnique(in []string) []string {
	out := make([]string, len(in))
	copy(out, in)
	slices.Sort(out)
	return slices.Compact(out)
}

// invalidNames returns the sorted, de-duplicated names in names that are not
// part of valid.
func invalidNames(names []string, valid sets.Set[string]) []string {
	var invalid []string
	for _, name := range names {
		if !valid.Has(name) {
			invalid = append(invalid, name)
		}
	}
	return sortedUnique(invalid)
}
