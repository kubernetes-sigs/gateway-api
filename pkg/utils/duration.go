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
	if matched := re.MatchString(s); matched == false {
		return nil, errors.New("Invalid duration format")
	}
	parsedTime, err := time.ParseDuration(s)
	if err != nil {
		return nil, err
	}

	return &parsedTime, nil

}

var reSplit = regexp.MustCompile(`[0-9]{1,5}(h|ms|s|m)`)

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
	m, _ := time.ParseDuration("0s")
	if duration == m {
		return "0s", nil
	}
	// time.Duration allows for floating point ms, which is not allowed in GEP-2257
	durationMicroseconds := duration.Microseconds()

	if durationMicroseconds%1000 != 0 {
		return "", errors.New("Cannot express sub-milliseconds precision in GEP-2257")
	}

	//Golang's time.Duration allows for floating point seconds instead of converting to ms
	durationMilliseconds := duration.Milliseconds()

	var ms int64
	if durationMilliseconds%1000 != 0 {
		ms = durationMilliseconds % 1000
		durationMilliseconds -= ms
		duration = time.Millisecond * time.Duration(durationMilliseconds)
	}

	durationString := duration.String()
	if ms > 0 {
		durationString += fmt.Sprintf("%dms", ms)
	}

	// check if a negative value
	if duration < 0 {
		return "", errors.New("Invalid duration format. Cannot have negative durations")
	}

	// trim the 0 values from the string (for example, 30m0s should result in 30m)
	// going to have a regexp that finds the index of the time units with 0, then appropriately trim those away
	temp := reSplit.FindAll([]byte(durationString), -1)
	res := ""
	for _, t := range temp {
		if t[0] != '0' {
			res += string(t)
		} else {
			continue
		}
	}

	// check if there are floating number points
	if matched := re.MatchString(res); matched == false {
		return "", errors.New("Invalid duration format")
	}

	return res, nil
}
