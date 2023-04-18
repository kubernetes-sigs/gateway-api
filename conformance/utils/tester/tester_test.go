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

package tester

import (
	"testing"
)

// TestTester is intended to demonstrate implemementation and usage of the Tester interface / struct and not test correctness.
func TestTester(t *testing.T) {
	ti := New(t)
	ti.Log("this should have a timestamp header")
	ti.Logf("%s %s", "this should have", "a timestamp header")
	ti.Run("subtest-1", func(ti Tester) {
		ti.Parallel()
		ti.Log("Log from subtest-1")
	})
	ti.Run("subtest-2", func(ti Tester) {
		ti.Parallel()
		ti.Log("Log from subtest-2")
	})
}
