/*
Copyright 2024 The Kubernetes Authors.

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

package suite

import (
	"reflect"
	"testing"

	"sigs.k8s.io/gateway-api/pkg/features"
)

func TestParseSupportedFeatures(t *testing.T) {
	flags := []string{
		"",
		"a",
		"b,c,d",
	}

	expected := [][]features.FeatureName{
		nil,
		{"a"},
		{"b", "c", "d"},
	}

	for i, f := range flags {
		expect := expected[i]
		got := ParseSupportedFeatures(f)
		if !reflect.DeepEqual(got, expect) {
			t.Errorf("Unexpected features from flags '%s', expected: %v, got: %v", f, expect, got)
		}
	}
}

func TestParseKeyValuePairs(t *testing.T) {
	flags := []string{
		"",
		"a=b",
		"b=c,d=e,f=g",
	}
	labels := []map[string]string{
		nil,
		{"a": "b"},
		{"b": "c", "d": "e", "f": "g"},
	}

	for i, f := range flags {
		expect := labels[i]
		got := ParseKeyValuePairs(f)
		if !reflect.DeepEqual(got, expect) {
			t.Errorf("Unexpected labels from flags '%s', expected: %v, got: %v", f, expect, got)
		}
	}
}
