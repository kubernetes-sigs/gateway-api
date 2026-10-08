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
	"regexp"
	"slices"
	"strings"
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
func (r ReleaseFeatures) Names() map[string]struct{} {
	names := make(map[string]struct{}, len(r.Standard)+len(r.Experimental))
	for _, name := range r.Standard {
		names[name] = struct{}{}
	}
	for _, name := range r.Experimental {
		names[name] = struct{}{}
	}
	return names
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
func invalidNames(names []string, valid map[string]struct{}) []string {
	var invalid []string
	for _, name := range names {
		if _, ok := valid[name]; !ok {
			invalid = append(invalid, name)
		}
	}
	return sortedUnique(invalid)
}

var (
	// featureListKeyRE matches the key of a list of features in a report,
	// for example "    supportedFeatures:".
	featureListKeyRE = regexp.MustCompile(`^(\s*)(?:supportedFeatures|unsupportedFeatures):\s*$`)
	// listItemRE matches a block sequence item holding a single word, for
	// example "    - HTTPRouteCORS".
	listItemRE = regexp.MustCompile(`^(\s*)-\s+(\S+)\s*$`)
)

// pruneFeatures removes the features in invalid from the supportedFeatures
// and unsupportedFeatures lists of a conformance report.
//
// The report is edited as text rather than being unmarshaled and marshaled
// again, so the rest of the file is left byte-for-byte as it was and the
// resulting diff only contains the removed lines. A list that ends up empty is
// removed together with its key, which matches how reports are generated.
//
// It returns the new content and the sorted, de-duplicated names that were
// actually removed.
func pruneFeatures(data []byte, invalid map[string]struct{}) ([]byte, []string) {
	lines := strings.Split(string(data), "\n")
	out := make([]string, 0, len(lines))
	var removed []string

	for i := 0; i < len(lines); i++ {
		keyMatch := featureListKeyRE.FindStringSubmatch(lines[i])
		if keyMatch == nil {
			out = append(out, lines[i])
			continue
		}
		keyIndent := len(keyMatch[1])

		// The list items belong to the key until the first line that is
		// not a list item. YAML allows the items to be indented the same as
		// the key, or deeper.
		var kept []string
		removedBefore := len(removed)
		j := i + 1
		for ; j < len(lines); j++ {
			itemMatch := listItemRE.FindStringSubmatch(lines[j])
			if itemMatch == nil || len(itemMatch[1]) < keyIndent {
				break
			}
			name := strings.Trim(itemMatch[2], `"'`)
			if _, drop := invalid[name]; drop {
				removed = append(removed, name)
				continue
			}
			kept = append(kept, lines[j])
		}

		// Only drop the key when this list was emptied by the pruning.
		if len(kept) > 0 || len(removed) == removedBefore {
			out = append(out, lines[i])
			out = append(out, kept...)
		}
		i = j - 1
	}

	return []byte(strings.Join(out, "\n")), sortedUnique(removed)
}
