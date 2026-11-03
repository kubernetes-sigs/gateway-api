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

package v1_test

import (
	"testing"

	v1 "sigs.k8s.io/gateway-api/conformance/apis/v1"
)

func TestProfileReport_IsConformant(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		// Receiver type
		coreResult     v1.Result
		extendedResult *v1.Result
		coreSkips      int
		extendedSkips  int
		// Named input parameters for target function.
		version    string
		wantResult v1.Result
		wantErrs   bool
	}{
		{
			name:           "Success in both",
			coreResult:     v1.Success,
			extendedResult: &v1.Success,
			coreSkips:      0,
			extendedSkips:  0,
			wantResult:     v1.Success,
			wantErrs:       false,
		},
		{
			name:           "Failure in Core",
			coreResult:     v1.Failure,
			extendedResult: &v1.Success,
			coreSkips:      0,
			extendedSkips:  0,
			wantResult:     v1.Failure,
			wantErrs:       true,
		},
		{
			name:           "Failure in Extended",
			coreResult:     v1.Success,
			extendedResult: &v1.Failure,
			coreSkips:      0,
			extendedSkips:  0,
			wantResult:     v1.Failure,
			wantErrs:       true,
		},
		{
			name:           "Failure in both",
			coreResult:     v1.Failure,
			extendedResult: &v1.Failure,
			coreSkips:      0,
			extendedSkips:  0,
			wantResult:     v1.Failure,
			wantErrs:       true,
		},
		{
			name:           "Partial in Extended",
			coreResult:     v1.Success,
			extendedResult: &v1.Partial,
			coreSkips:      0,
			extendedSkips:  0,
			wantResult:     v1.Partial,
			wantErrs:       false,
		},
		{
			name:           "Partial in both",
			coreResult:     v1.Partial,
			extendedResult: &v1.Partial,
			coreSkips:      0,
			extendedSkips:  0,
			wantResult:     v1.Partial,
			wantErrs:       false,
		},
		{
			name:           "Core Skips, Partial",
			coreResult:     v1.Partial,
			extendedResult: &v1.Partial,
			coreSkips:      2,
			extendedSkips:  0,
			wantResult:     v1.Partial,
			wantErrs:       true,
		},
		{
			name:           "Success, too many core skips",
			coreResult:     v1.Success,
			extendedResult: &v1.Success,
			coreSkips:      2,
			extendedSkips:  0,
			wantResult:     v1.Partial,
			wantErrs:       true,
		},
		{
			name:           "Success, too many skips overall",
			coreResult:     v1.Success,
			extendedResult: &v1.Success,
			coreSkips:      0,
			extendedSkips:  6,
			wantResult:     v1.Partial,
			wantErrs:       true,
		},
		{
			name:           "Success, too many skips overall, with core skips",
			coreResult:     v1.Success,
			extendedResult: &v1.Success,
			coreSkips:      2,
			extendedSkips:  6,
			wantResult:     v1.Partial,
			wantErrs:       true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Small hack, we use the test name as the profile name to make checking results easier.
			p := makeProfileReport(tt.name, tt.coreResult, tt.extendedResult, tt.coreSkips, tt.extendedSkips)
			got, got2, gotErrs := p.IsConformant("")
			if len(gotErrs) > 0 {
				if !tt.wantErrs {
					t.Errorf("IsConformant() failed, got unexpected errors: %v", gotErrs)
				}
				return
			}
			if tt.wantErrs {
				t.Fatal("IsConformant() succeeded unexpectedly")
			}
			// TODO: update the condition below to compare got with tt.want.
			if tt.wantResult != got {
				t.Errorf("Conformance result = %v, want %v", got, tt.wantResult)
			}
			if got2 != tt.coreSkips+tt.extendedSkips {
				t.Errorf("Got unexpected number of total skips:  %v, want %v", got2, tt.coreSkips+tt.extendedSkips)
			}
		})
	}
}

func makeProfileReport(name string, coreResult v1.Result, extendedResult *v1.Result, coreSkips, extendedSkips int) *v1.ProfileReport {

	returnProfile := v1.ProfileReport{
		Name:    name,
		Summary: "something",
		Core: v1.Status{
			Result: coreResult,
			Statistics: v1.Statistics{
				Skipped: uint32(coreSkips),
			},
		},
	}

	if extendedResult != nil {
		returnProfile.Extended = &v1.ExtendedStatus{
			Status: v1.Status{
				Result: *extendedResult,
				Statistics: v1.Statistics{
					Skipped: uint32(extendedSkips),
				},
			},
		}
	}

	return &returnProfile
}
