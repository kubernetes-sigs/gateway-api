package main

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func makeDuration(h, m, s, ms int) time.Duration {
	duration := h*int(time.Hour) + m*int(time.Minute) + s*int(time.Second) + ms*int(time.Millisecond)
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
			expected: makeDuration(0, 0, 0, 0),
		},
		{
			name:     "0s should be 0s",
			args:     "0s",
			expected: makeDuration(0, 0, 0, 0),
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
			name:     "1h30m10s shoulw be 1h30m10s",
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
		args     string
		expected string
	}{
		{
			name:     "0 should be 0s",
			args:     "0",
			expected: "0s",
		},
		{
			name:     "1h should be 1h",
			args:     "1h",
			expected: "1h",
		},
		{
			name:     "30m should be 30m",
			args:     "30m",
			expected: "30m",
		},
		{
			name:     "10s should be 10s",
			args:     "10s",
			expected: "10s",
		},
		{
			name:     "500ms should be 500ms",
			args:     "500ms",
			expected: "500ms",
		},
		{
			name:     "2h30m should be 2h30m",
			args:     "2h30m",
			expected: "2h30m",
		},
		{
			name:     "1h30m10s should be 1h30m10s",
			args:     "1h30m10s",
			expected: "1h30m10s",
		},
		{
			name:     "600ms should be 600ms",
			args:     "600ms",
			expected: "600ms",
		},
		{
			name:     "2h600ms should be 2h600ms",
			args:     "2h600ms",
			expected: "2h600ms",
		},
		{
			name:     "2h30m600ms should be 2h30m600ms",
			args:     "2h30m600ms",
			expected: "2h30m600ms",
		},
		{
			name:     "2h30m10s600ms should be 2h30m10s600ms",
			args:     "2h30m10s600ms",
			expected: "2h30m10s600ms",
		},
		{
			name:     "0.5m should be 30s",
			args:     "0.5m",
			expected: "30s",
		},
		{
			name:     "0.5s should be 500ms",
			args:     "0.5s",
			expected: "500ms",
		},
	}

	// invalid test cases
	invalidTestCases := []struct {
		name string
		args string
	}{
		{
			name: "Sub-milliseconds not allowed (100us)",
			args: "100us",
		},
		{
			name: "Sub-milliseconds not allowed (0.5ms)",
			args: "0.5ms",
		},
		{
			name: "Out of range (greater than 99999 hours)",
			args: "100000h",
		},
		{
			name: "Negative duration not supported",
			args: "-10h",
		},
	}
	// Valid test cases

	for _, tc := range validTestCases {
		t.Run(tc.name, func(t *testing.T) {
			a, _ := time.ParseDuration(tc.args)
			arg, _ := FormatDuration(a)
			assert.Equal(t, tc.expected, arg)
		})
	}

	// Invalid test cases
	for _, tc := range invalidTestCases {
		t.Run(tc.name, func(t *testing.T) {
			_, errArg := ParseDuration(tc.args)
			assert.Error(t, errArg)
		})
	}
}
