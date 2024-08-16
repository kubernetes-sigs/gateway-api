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

package main

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func makeDuration(h, m, s, ms float64) time.Duration {
	duration := h*float64(time.Hour) + m*float64(time.Minute) + s*float64(time.Second) + ms*float64(time.Millisecond)
	return time.Duration(duration)
}

func TestParseDuration(t *testing.T) {
	// valid durations
	validTestCases := []struct {
		name     string
		args     string
		expected time.Duration
	}{
		{
			name:     "0h to timeDuration of 0s",
			args:     "0h",
			expected: time.Hour * 0,
		},
		{
			name:     "0s should be 0s",
			args:     "0s",
			expected: time.Second * 0,
		},
		{
			name:     "0h0m0s should be 0s",
			args:     "0h0m0s",
			expected: makeDuration(0, 0, 0, 0),
		},
		{
			name:     "1h should be 1h",
			args:     "1h",
			expected: makeDuration(1, 0, 0, 0),
		},
		{
			name:     "30m should be 30m",
			args:     "30m",
			expected: makeDuration(0, 30, 0, 0),
		},
		{
			name:     "10s should be 10s",
			args:     "10s",
			expected: makeDuration(0, 0, 10, 0),
		},
		{
			name:     "500ms should be 500ms",
			args:     "500ms",
			expected: makeDuration(0, 0, 0, 500),
		},
		{
			name:     "2h30m should be 2h30m",
			args:     "2h30m",
			expected: makeDuration(2, 30, 0, 0),
		},
		{
			name:     "150m should be 2h30m",
			args:     "150m",
			expected: makeDuration(0, 150, 0, 0),
		},
		{
			name:     "7320s should be 2h30m",
			args:     "7320s",
			expected: makeDuration(0, 0, 7320, 0),
		},
		{
			name:     "1h30m10s should be 1h30m10s",
			args:     "1h30m10s",
			expected: makeDuration(1, 30, 10, 0),
		},
		{
			name:     "10s30m1h should be 1h30m10s",
			args:     "10s30m1h",
			expected: makeDuration(1, 30, 10, 0),
		},
		{
			name:     "100ms200ms300ms should be 600ms",
			args:     "100ms200ms300ms",
			expected: makeDuration(0, 0, 0, 600),
		},
	}

	invalidTestCases := []struct {
		name string
		args string
	}{
		{
			name: "Missing unit",
			args: "1",
		},
		{
			name: "Missing unit in 1h1",
			args: "1h1",
		},
		{
			name: "Too many units/components",
			args: "1h30m10s20ms50h",
		},
		{
			name: "Too many digits",
			args: "999999h",
		},
		{
			name: "No floating points allowed",
			args: "1.5h",
		},
		{
			name: "No floating points for seconds allowed",
			args: "0.5s",
		},
		{
			name: "Negative numbers not allowed",
			args: "-15m",
		},
	}

	// Running valid test cases
	for _, tc := range validTestCases {
		t.Run(tc.name, func(t *testing.T) {
			arg, _ := ParseDuration(tc.args)
			assert.Equal(t, tc.expected, *arg)
		})
	}

	// Running invalid test cases
	for _, tc := range invalidTestCases {
		t.Run(tc.name, func(t *testing.T) {
			_, errArg := ParseDuration(tc.args)
			assert.Error(t, errArg)
		})
	}
}

func TestFormatDuration(t *testing.T) {
	validTestCases := []struct {
		name     string
		args     time.Duration
		expected string
	}{
		{
			name:     "0 should be 0s",
			args:     makeDuration(0, 0, 0, 0),
			expected: "0s",
		},
		{
			name:     "1h should be 1h",
			args:     makeDuration(1, 0, 0, 0),
			expected: "1h",
		},
		{
			name:     "30m should be 30m",
			args:     makeDuration(0, 30, 0, 0),
			expected: "30m",
		},
		{
			name:     "10s should be 10s",
			args:     makeDuration(0, 0, 10, 0),
			expected: "10s",
		},
		{
			name:     "500ms should be 500ms",
			args:     makeDuration(0, 0, 0, 500),
			expected: "500ms",
		},
		{
			name:     "2h30m should be 2h30m",
			args:     makeDuration(2, 30, 0, 0),
			expected: "2h30m",
		},
		{
			name:     "1h30m10s should be 1h30m10s",
			args:     makeDuration(1, 30, 10, 0),
			expected: "1h30m10s",
		},
		{
			name:     "600ms should be 600ms",
			args:     makeDuration(0, 0, 0, 600),
			expected: "600ms",
		},
		{
			name:     "2h600ms should be 2h600ms",
			args:     makeDuration(2, 0, 0, 600),
			expected: "2h600ms",
		},
		{
			name:     "2h30m600ms should be 2h30m600ms",
			args:     makeDuration(2, 30, 0, 600),
			expected: "2h30m600ms",
		},
		{
			name:     "2h30m10s600ms should be 2h30m10s600ms",
			args:     makeDuration(2, 30, 10, 600),
			expected: "2h30m10s600ms",
		},
		{
			name:     "0.5m should be 30s",
			args:     makeDuration(0, 0.5, 0, 0),
			expected: "30s",
		},
		{
			name:     "0.5s should be 500ms",
			args:     makeDuration(0, 0, 0.5, 0),
			expected: "500ms",
		},
	}
	// invalid test cases
	invalidTestCases := []struct {
		name string
		args time.Duration
	}{
		{
			name: "Sub-milliseconds not allowed (100us)",
			args: 100 * time.Microsecond,
		},
		{
			name: "Sub-milliseconds not allowed (0.5ms)",
			args: makeDuration(0, 0, 0, 0.5),
		},
		{
			name: "Out of range (greater than 99999 hours)",
			args: makeDuration(100000, 0, 0, 0),
		},
		{
			name: "Negative duration not supported",
			args: -15 * time.Hour,
		},
		{
			name: "Max duration represented in GEP",
			args: makeDuration(99999, 59, 59, 999) + makeDuration(0, 0, 0, 1),
		},
	}
	// Valid test cases

	for _, tc := range validTestCases {
		t.Run(tc.name, func(t *testing.T) {
			arg, _ := FormatDuration(tc.args)
			assert.Equal(t, tc.expected, arg)
		})
	}

	// Invalid test cases
	for _, tc := range invalidTestCases {
		t.Run(tc.name, func(t *testing.T) {
			_, errArg := FormatDuration(tc.args)
			assert.Error(t, errArg)
		})
	}
}
