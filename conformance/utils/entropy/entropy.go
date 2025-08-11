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

package entropy

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"time"
)

// randomNumber generates a random number between 0 and limit-1
func randomNumber(limit int64) (*int64, error) {
	number, err := rand.Int(rand.Reader, big.NewInt(limit))
	if err != nil {
		return nil, err
	}
	n := number.Int64()
	return &n, nil
}

// AddDelay adds a random delay up to the specified limit in milliseconds
func AddDelay(limit int64) error {
	randomSleepDuration, err := randomNumber(limit)
	if err != nil {
		return err
	}
	time.Sleep(time.Duration(*randomSleepDuration) * time.Millisecond)
	return nil
}

// GenerateRandomValue generates a random value as a string for use in headers/metadata
func GenerateRandomValue(limit int64) (string, error) {
	randomVal, err := randomNumber(limit)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%d", *randomVal), nil
}

// AddRandomEntropy randomly chooses to add delay, random value, or both
// The addRandomValue function should be provided by the caller to handle
// protocol-specific ways of adding the random value (HTTP headers, gRPC metadata, etc.)
func AddRandomEntropy(addRandomValue func(string) error) error {
	random, err := randomNumber(3)
	if err != nil {
		return err
	}

	switch *random {
	case 0:
		return AddDelay(1000)
	case 1:
		randomValue, err := GenerateRandomValue(10000)
		if err != nil {
			return err
		}
		return addRandomValue(randomValue)
	case 2:
		if err := AddDelay(1000); err != nil {
			return err
		}
		randomValue, err := GenerateRandomValue(10000)
		if err != nil {
			return err
		}
		return addRandomValue(randomValue)
	default:
		return fmt.Errorf("invalid random value: %d", *random)
	}
}