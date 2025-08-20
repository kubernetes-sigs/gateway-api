/*
Copyright 2025 The Kubernetes Authors.

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
	"cmp"
	"crypto/rand"
	"errors"
	"fmt"
	"math"
	"math/big"
	"slices"
	"strconv"
	"strings"
	"sync"
	"time"

	"golang.org/x/sync/errgroup"
)

// RequestSender defines an interface for sending requests (HTTP, gRPC, or mesh)
type RequestSender interface {
	SendRequest() (podName string, err error)
}

// TestWeightedDistribution tests that requests are distributed according to expected weights
func TestWeightedDistribution(sender RequestSender, expectedWeights map[string]float64) error {
	const (
		concurrentRequests  = 10
		tolerancePercentage = 0.05
		totalRequests       = 500.0
	)

	var (
		g         errgroup.Group
		seenMutex sync.Mutex
		seen      = make(map[string]float64, len(expectedWeights))
	)

	g.SetLimit(concurrentRequests)
	for i := 0.0; i < totalRequests; i++ {
		g.Go(func() error {
			podName, err := sender.SendRequest()
			if err != nil {
				return err
			}

			seenMutex.Lock()
			defer seenMutex.Unlock()

			for expectedBackend := range expectedWeights {
				if strings.HasPrefix(podName, expectedBackend) {
					seen[expectedBackend]++
					return nil
				}
			}

			return fmt.Errorf("request was handled by an unexpected pod %q", podName)
		})
	}

	if err := g.Wait(); err != nil {
		return fmt.Errorf("error while sending requests: %w", err)
	}

	// Count how many backends should receive traffic (weight > 0)
	expectedActiveBackends := 0
	for _, weight := range expectedWeights {
		if weight > 0.0 {
			expectedActiveBackends++
		}
	}

	var errs []error
	if len(seen) != expectedActiveBackends {
		errs = append(errs, fmt.Errorf("expected %d backends to receive traffic, but got %d", expectedActiveBackends, len(seen)))
	}

	for wantBackend, wantPercent := range expectedWeights {
		gotCount, ok := seen[wantBackend]

		if !ok && wantPercent != 0.0 {
			errs = append(errs, fmt.Errorf("expect traffic to hit backend %q - but none was received", wantBackend))
			continue
		}

		gotPercent := gotCount / totalRequests

		if math.Abs(gotPercent-wantPercent) > tolerancePercentage {
			errs = append(errs, fmt.Errorf("backend %q weighted traffic of %v not within tolerance %v (+/-%f)",
				wantBackend,
				gotPercent,
				wantPercent,
				tolerancePercentage,
			))
		}
	}

	slices.SortFunc(errs, func(a, b error) int {
		return cmp.Compare(a.Error(), b.Error())
	})
	return errors.Join(errs...)
}

// Entropy utilities

// randomNumber generates a random number between 0 and limit-1
func randomNumber(limit int) (int, error) {
	number, err := rand.Int(rand.Reader, big.NewInt(int64(limit)))
	if err != nil {
		return 0, err
	}
	return int(number.Int64()), nil
}

// AddDelay adds a random delay up to the specified limit in milliseconds
func AddDelay(limit int) error {
	randomSleepDuration, err := randomNumber(limit)
	if err != nil {
		return err
	}
	time.Sleep(time.Duration(randomSleepDuration) * time.Millisecond)
	return nil
}

// AddRandomEntropy randomly chooses to add delay, random value, or both
// The addRandomValue function should be provided by the caller to handle
// protocol-specific ways of adding the random value (HTTP headers, gRPC metadata, etc.)
func AddRandomEntropy(addRandomValue func(string) error) error {
	random, err := randomNumber(3)
	if err != nil {
		return err
	}

	switch random {
	case 0:
		return AddDelay(1000)
	case 1:
		randomValue, err := randomNumber(10000)
		if err != nil {
			return err
		}
		return addRandomValue(strconv.Itoa(randomValue))
	case 2:
		if err := AddDelay(1000); err != nil {
			return err
		}
		randomValue, err := randomNumber(10000)
		if err != nil {
			return err
		}
		return addRandomValue(strconv.Itoa(randomValue))
	default:
		return fmt.Errorf("invalid random value: %d", random)
	}
}
