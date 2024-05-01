/*
Copyright 2023 The Kubernetes Authors.

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

package common

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/google/go-cmp/cmp"
)

// YamlString defines a custom type for wrapping yaml texts. It makes use of
// YamlStringTransformer to generate slightly better diffing output from
// cmp.Diff() for multi-line yaml texts.
type YamlString string

// YamlStringTransformer transforms a YamlString into a slice of strings by
// splitting on each new line. This allows the diffing function (used in tests)
// to compare each line independently. The result is that the diff output marks
// each line where a diff was observed.
var YamlStringTransformer = cmp.Transformer("YamlLines", func(s YamlString) []string {
	// Split string on each new line.
	lines := strings.Split(string(s), "\n")

	// Remove any empty lines from the start and end.
	var start, end int
	for i := range lines {
		if lines[i] != "" {
			start = i
			break
		}
	}
	for i := len(lines) - 1; i >= 0; i-- {
		if lines[i] != "" {
			end = i
			break
		}
	}
	return lines[start : end+1]
})

type JSONString string

func (src JSONString) CmpDiff(tgt JSONString) (diff string, err error) {
	var srcMap, targetMap map[string]interface{}
	err = json.Unmarshal([]byte(src), &srcMap)
	if err != nil {
		err = fmt.Errorf("failed to unmarshal the source json: %w", err)
		return
	}
	err = json.Unmarshal([]byte(tgt), &targetMap)
	if err != nil {
		err = fmt.Errorf("failed to unmarshal the target json: %w", err)
		return
	}

	return cmp.Diff(srcMap, targetMap), nil
}
