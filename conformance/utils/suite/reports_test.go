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
	"testing"

	"github.com/stretchr/testify/require"

	confv1 "sigs.k8s.io/gateway-api/conformance/apis/v1"
)

func TestBuildSummary(t *testing.T) {
	testCases := []struct {
		name            string
		report          confv1.ProfileReport
		expectedSummary string
	}{
		{
			name: "core tests failed, no extended tests",
			report: confv1.ProfileReport{
				Name: string(GatewayHTTPConformanceProfileName),
				Core: confv1.Status{
					Result: confv1.Failure,
					Statistics: confv1.Statistics{
						Passed: 5,
						Failed: 3,
					},
				},
			},
			expectedSummary: "Core tests failed with 3 test failures.",
		},
		{
			name: "core tests succeeded, extended tests failed",
			report: confv1.ProfileReport{
				Name: string(GatewayHTTPConformanceProfileName),
				Core: confv1.Status{
					Result: confv1.Success,
					Statistics: confv1.Statistics{
						Passed: 8,
					},
				},
				Extended: &confv1.ExtendedStatus{
					Status: confv1.Status{
						Result: confv1.Failure,
						Statistics: confv1.Statistics{
							Passed: 2,
							Failed: 1,
						},
					},
				},
			},
			expectedSummary: "Core tests succeeded. Extended tests failed with 1 test failures.",
		},
		{
			name: "core tests partially succeeded, extended tests succeeded",
			report: confv1.ProfileReport{
				Name: string(GatewayHTTPConformanceProfileName),
				Core: confv1.Status{
					Result: confv1.Partial,
					Statistics: confv1.Statistics{
						Passed:  6,
						Skipped: 2,
					},
				},
				Extended: &confv1.ExtendedStatus{
					Status: confv1.Status{
						Result: confv1.Success,
						Statistics: confv1.Statistics{
							Passed: 2,
						},
					},
				},
			},
			expectedSummary: "Core tests partially succeeded with 2 test skips. Extended tests succeeded.",
		},
		{
			name: "core tests succeeded, extended tests partially succeeded",
			report: confv1.ProfileReport{
				Name: string(GatewayHTTPConformanceProfileName),
				Core: confv1.Status{
					Result: confv1.Success,
					Statistics: confv1.Statistics{
						Passed: 8,
					},
				},
				Extended: &confv1.ExtendedStatus{
					Status: confv1.Status{
						Result: confv1.Partial,
						Statistics: confv1.Statistics{
							Passed:  2,
							Skipped: 1,
						},
					},
				},
			},
			expectedSummary: "Core tests succeeded. Extended tests partially succeeded with 1 test skips.",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			summary := buildSummary(tc.report)
			require.Equal(t, tc.expectedSummary, summary)
		})
	}
}
