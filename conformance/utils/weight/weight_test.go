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

package weight

import (
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"sigs.k8s.io/gateway-api/conformance/utils/config"
)

func TestExtractBackendName(t *testing.T) {
	tests := []struct {
		name     string
		podName  string
		expected string
	}{
		{
			name:     "standard pod name with deployment and pod hash",
			podName:  "backend-v1-5746b847bc-4x9zq",
			expected: "backend-v1",
		},
		{
			name:     "pod name with single component backend",
			podName:  "backend-5746b847bc-4x9zq",
			expected: "backend",
		},
		{
			name:     "pod name with multi-part backend name",
			podName:  "my-awesome-service-v2-abc1234-xyz89",
			expected: "my-awesome-service-v2",
		},
		{
			name:     "short name without sufficient dash components",
			podName:  "singleword",
			expected: "singleword",
		},
		{
			name:     "two part name without sufficient dash components",
			podName:  "two-parts",
			expected: "two-parts",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := extractBackendName(tc.podName)
			assert.Equal(t, tc.expected, got)
		})
	}
}

func TestTestWeightedDistribution(t *testing.T) {
	t.Run("matching distribution succeeds", func(t *testing.T) {
		var count atomic.Int32
		sender := NewFunctionBasedSender(func() (string, error) {
			idx := count.Add(1)
			if idx <= 350 {
				return "backend-v1-5746b847bc-pod1", nil
			}
			return "backend-v2-5746b847bc-pod2", nil
		})

		expectedWeights := map[string]float64{
			"backend-v1": 0.7,
			"backend-v2": 0.3,
			"backend-v3": 0.0,
		}

		err := TestWeightedDistribution(sender, expectedWeights)
		require.NoError(t, err)
	})

	t.Run("unexpected backend returns error", func(t *testing.T) {
		sender := NewFunctionBasedSender(func() (string, error) {
			return "unknown-backend-5746b847bc-pod1", nil
		})

		expectedWeights := map[string]float64{
			"backend-v1": 1.0,
		}

		err := TestWeightedDistribution(sender, expectedWeights)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "unexpected pod")
	})

	t.Run("out of tolerance distribution returns error", func(t *testing.T) {
		sender := NewFunctionBasedSender(func() (string, error) {
			return "backend-v1-5746b847bc-pod1", nil
		})

		expectedWeights := map[string]float64{
			"backend-v1": 0.5,
			"backend-v2": 0.5,
		}

		err := TestWeightedDistribution(sender, expectedWeights)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "not within tolerance")
	})
}

func TestTestWeightedDistributionBatch(t *testing.T) {
	t.Run("matching batch distribution succeeds", func(t *testing.T) {
		sender := NewBatchFunctionBasedSender(func(count int) ([]string, error) {
			res := make([]string, count)
			for i := range count {
				if i < 350 {
					res[i] = "backend-v1-5746b847bc-pod1"
				} else {
					res[i] = "backend-v2-5746b847bc-pod2"
				}
			}
			return res, nil
		})

		expectedWeights := map[string]float64{
			"backend-v1": 0.7,
			"backend-v2": 0.3,
		}

		err := TestWeightedDistributionBatch(sender, expectedWeights)
		require.NoError(t, err)
	})

	t.Run("unexpected backend in batch returns error", func(t *testing.T) {
		sender := NewBatchFunctionBasedSender(func(count int) ([]string, error) {
			res := make([]string, count)
			for i := range count {
				res[i] = "rogue-backend-5746b847bc-pod1"
			}
			return res, nil
		})

		expectedWeights := map[string]float64{
			"backend-v1": 1.0,
		}

		err := TestWeightedDistributionBatch(sender, expectedWeights)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "unexpected pod")
	})
}

func TestExpectWeightedDistribution(t *testing.T) {
	timeoutConfig := config.TimeoutConfig{
		MaxTimeToConsistency: 2 * time.Second,
		RequestTimeout:       100 * time.Millisecond,
		DefaultPollInterval:  10 * time.Millisecond,
	}

	t.Run("succeeds when distribution matches immediately", func(t *testing.T) {
		var reqCount atomic.Int32
		sender := NewFunctionBasedSender(func() (string, error) {
			idx := reqCount.Add(1) % 500
			if idx < 350 {
				return "backend-v1-5746b847bc-pod1", nil
			}
			return "backend-v2-5746b847bc-pod2", nil
		})

		expectedWeights := map[string]float64{
			"backend-v1": 0.7,
			"backend-v2": 0.3,
		}

		ExpectWeightedDistribution(t, timeoutConfig, sender, expectedWeights)
	})

	t.Run("succeeds after initial mismatch converges", func(t *testing.T) {
		var round atomic.Int32
		var reqCount atomic.Int32

		sender := NewFunctionBasedSender(func() (string, error) {
			currReq := reqCount.Add(1)
			if currReq%500 == 1 {
				round.Add(1)
			}

			// In round 1, send 100% to backend-v1 (mismatch)
			if round.Load() == 1 {
				return "backend-v1-5746b847bc-pod1", nil
			}

			// In round 2+, converge to 70% v1 / 30% v2
			if (currReq % 500) < 350 {
				return "backend-v1-5746b847bc-pod1", nil
			}
			return "backend-v2-5746b847bc-pod2", nil
		})

		expectedWeights := map[string]float64{
			"backend-v1": 0.7,
			"backend-v2": 0.3,
		}

		ExpectWeightedDistribution(t, timeoutConfig, sender, expectedWeights)
		assert.GreaterOrEqual(t, round.Load(), int32(2))
	})
}

func TestExpectWeightedDistributionBatch(t *testing.T) {
	timeoutConfig := config.TimeoutConfig{
		MaxTimeToConsistency: 2 * time.Second,
		RequestTimeout:       100 * time.Millisecond,
		DefaultPollInterval:  10 * time.Millisecond,
	}

	t.Run("succeeds when batch distribution matches immediately", func(t *testing.T) {
		sender := NewBatchFunctionBasedSender(func(count int) ([]string, error) {
			res := make([]string, count)
			for i := range count {
				if i < 350 {
					res[i] = "backend-v1-5746b847bc-pod1"
				} else {
					res[i] = "backend-v2-5746b847bc-pod2"
				}
			}
			return res, nil
		})

		expectedWeights := map[string]float64{
			"backend-v1": 0.7,
			"backend-v2": 0.3,
		}

		ExpectWeightedDistributionBatch(t, timeoutConfig, sender, expectedWeights)
	})

	t.Run("succeeds after initial batch mismatch converges", func(t *testing.T) {
		var attempt atomic.Int32
		sender := NewBatchFunctionBasedSender(func(count int) ([]string, error) {
			res := make([]string, count)
			att := attempt.Add(1)
			for i := range count {
				if att == 1 {
					// Round 1: mismatch
					res[i] = "backend-v1-5746b847bc-pod1"
				} else {
					// Round 2+: converged
					if i < 350 {
						res[i] = "backend-v1-5746b847bc-pod1"
					} else {
						res[i] = "backend-v2-5746b847bc-pod2"
					}
				}
			}
			return res, nil
		})

		expectedWeights := map[string]float64{
			"backend-v1": 0.7,
			"backend-v2": 0.3,
		}

		ExpectWeightedDistributionBatch(t, timeoutConfig, sender, expectedWeights)
		assert.GreaterOrEqual(t, attempt.Load(), int32(2))
	})
}
