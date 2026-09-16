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

func TestImplementation_Validate(t *testing.T) {
	validImpl := Implementation{
		Organization: "sigs-k8s",
		Project:      "gateway-api",
		URL:          "https://github.com/kubernetes-sigs/gateway-api",
		Version:      "v1.0.0",
		Contact:      []string{"@kubernetes-sigs/gateway-api-maintainers"},
	}

	tests := []struct {
		name    string
		mutate  func(i Implementation) Implementation
		wantErr bool
	}{
		{
			name:    "valid implementation",
			mutate:  func(i Implementation) Implementation { return i },
			wantErr: false,
		},
		{
			name: "empty organization",
			mutate: func(i Implementation) Implementation {
				i.Organization = ""
				return i
			},
			wantErr: true,
		},
		{
			name: "empty project",
			mutate: func(i Implementation) Implementation {
				i.Project = ""
				return i
			},
			wantErr: true,
		},
		{
			name: "empty url",
			mutate: func(i Implementation) Implementation {
				i.URL = ""
				return i
			},
			wantErr: true,
		},
		{
			name: "url missing scheme",
			mutate: func(i Implementation) Implementation {
				i.URL = "github.com/kubernetes-sigs/gateway-api"
				return i
			},
			wantErr: true,
		},
		{
			name: "url with unsupported scheme",
			mutate: func(i Implementation) Implementation {
				i.URL = "ftp://github.com/kubernetes-sigs/gateway-api"
				return i
			},
			wantErr: true,
		},
		{
			name: "url missing host",
			mutate: func(i Implementation) Implementation {
				i.URL = "https:///path-only"
				return i
			},
			wantErr: true,
		},
		{
			name: "malformed url",
			mutate: func(i Implementation) Implementation {
				i.URL = "https://exa mple.com"
				return i
			},
			wantErr: true,
		},
		{
			name: "empty version",
			mutate: func(i Implementation) Implementation {
				i.Version = ""
				return i
			},
			wantErr: true,
		},
		{
			name: "empty contact",
			mutate: func(i Implementation) Implementation {
				i.Contact = nil
				return i
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			impl := tt.mutate(validImpl)
			err := impl.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

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
