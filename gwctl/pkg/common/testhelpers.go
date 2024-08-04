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
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// MultiLine defines a custom type for wrapping texts spanning multilpe lines.
// It makes use of MultiLineTransformer to generate slightly better diffing
// output from cmp.Diff() for multi-line texts.
type MultiLine string

// MultiLineTransformer transforms a MultiLine into a slice of strings by
// splitting on each new line. This allows the diffing function (used in tests)
// to compare each line independently. The result is that the diff output marks
// each line where a diff was observed.
var MultiLineTransformer = cmp.Transformer("MultiLine", func(m MultiLine) []string {
	return strings.Split(string(m), "\n")
})

const (
	beginMarker = "#################################### BEGIN #####################################"
	endMarker   = "##################################### END ######################################"
)

func (m MultiLine) String() string {
	return fmt.Sprintf("%v\n%v%v", beginMarker, string(m), endMarker)
}

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

func NamespaceForTest(name string) *corev1.Namespace {
	return &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		Status: corev1.NamespaceStatus{
			Phase: corev1.NamespaceActive,
		},
	}
}

func MustPrettyPrint(data any) {
	b, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		panic(err)
	}
	fmt.Println(string(b))
}
