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
	"errors"
	"fmt"
	"regexp"
	"time"
)

var re = regexp.MustCompile(`^([0-9]{1,5}(h|m|s|ms)){1,4}$`)

func ParseDuration(s string) (*time.Duration, error) {
	/*
		parseDuration parses a GEP-2257 Duration format to a time object
		Valid date units in time.Duration are "ns", "us" (or "µs"), "ms", "s", "m", "h"
		Valid date units according to GEP-2257 are "h", "m", "s", "ms"

		input: string

		output: time.Duration

		See https://gateway-api.sigs.k8s.io/geps/gep-2257/ for more details.
	*/
	if !re.MatchString(s) {
		return nil, errors.New("Invalid duration format")
	}
	parsedTime, err := time.ParseDuration(s)
	if err != nil {
		return nil, err
	}

	return &parsedTime, nil
}

const maxDuration = 99999*time.Hour + 59*time.Minute + 59*time.Second + 999*time.Millisecond

func FormatDuration(duration time.Duration) (string, error) {
	/*
		formatDuration formats a time object to GEP-2257 Duration format to a GEP-2257 Duration Format
		The time format from GEP-2257 must match the regex
		"^([1-9]{1,5}(h|m|s|ms)){1,4}$"

		A time.Duration allows for negative time, floating points, and allow for zero units
		For example, -4h, 4.5h, and 4h0m0s are all valid in the golang time package

		See https://gateway-api.sigs.k8s.io/geps/gep-2257/ for more details.

		Input: time.Duration

		Returns: string or error if duration cannot be expressed as a GEP-2257 Duration format.
	*/

	if duration == 0 {
		return "0s", nil
	}

	// check if a negative value
	if duration < 0 {
		return "", errors.New("Invalid duration format. Cannot have negative durations")
	}
	// check for the maximum value allowed to be expressed
	if duration > maxDuration {
		return "", errors.New("Invalid duration format. Duration larger than maximum expression allowed in GEP-2257")
	}
	// time.Duration allows for floating point ms, which is not allowed in GEP-2257
	durationMicroseconds := duration.Microseconds()

	if durationMicroseconds%1000 != 0 {
		return "", errors.New("Cannot express sub-milliseconds precision in GEP-2257")
	}

	output := ""
	seconds := int(duration.Seconds())

	// calculating the hours
	hours := seconds / 3600

	if hours > 0 {
		output += fmt.Sprintf("%dh", hours)
		seconds -= hours * 3600
	}

	// calculating the minutes
	minutes := seconds / 60

	if minutes > 0 {
		output += fmt.Sprintf("%dm", minutes)
		seconds -= minutes * 60
	}

	// calculating the seconds
	if seconds > 0 {
		output += fmt.Sprintf("%ds", seconds)
	}

	// calculating the milliseconds
	durationMilliseconds := durationMicroseconds / 1000

	ms := durationMilliseconds % 1000
	if ms != 0 {
		output += fmt.Sprintf("%dms", ms)
	}

	return output, nil
}
