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

package v1

import "testing"

func Test_updateResult(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		// Named input parameters for target function.
		old  Result
		new  Result
		want Result
	}{
		{
			name: "Old Failure, new Success",
			old:  Failure,
			new:  Success,
			want: Success,
		},
		{
			name: "Old Success, new Failure",
			old:  Success,
			new:  Failure,
			want: Success,
		},
		{
			name: "Old Partial, new Success",
			old:  Partial,
			new:  Success,
			want: Success,
		},
		{
			name: "Old Success, new Partial",
			old:  Success,
			new:  Partial,
			want: Success,
		},
		{
			name: "Old Failure, new Partial",
			old:  Failure,
			new:  Partial,
			want: Partial,
		},
		{
			name: "Old Partial, new Failure",
			old:  Partial,
			new:  Failure,
			want: Partial,
		},
		{
			name: "Both Failure",
			old:  Failure,
			new:  Failure,
			want: Failure,
		},
		{
			name: "Both Partial",
			old:  Partial,
			new:  Partial,
			want: Partial,
		},
		{
			name: "Both Success",
			old:  Success,
			new:  Success,
			want: Success,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := updateResult(tt.old, tt.new)
			if got != tt.want {
				t.Errorf("updateResult() = %v, want %v", got, tt.want)
			}
		})
	}
}
