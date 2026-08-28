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

package v1

import (
	"errors"
	"fmt"

	"golang.org/x/mod/semver"
)

// ProfileReport is the generated report for the test results of a specific
// named conformance profile.
type ProfileReport struct {
	// Name indicates the name of the conformance profile (e.g. "HTTP",
	// "TLS", "Mesh", e.t.c.).
	Name string `json:"name"`

	// Summary is a human-readable message intended for end-users to understand
	// the overall status at a glance.
	Summary string `json:"summary"`

	// Core indicates the core support level which includes the set of tests
	// which are the minimum the implementation must pass to be considered at
	// all conformant.
	Core Status `json:"core"`

	// Extended indicates the extended support level which includes additional
	// optional features which the implementation may choose to implement
	// support for, but are not required.
	Extended *ExtendedStatus `json:"extended,omitempty"`
}

// ExtendedStatus shows the testing results for the extended support level.
type ExtendedStatus struct {
	Status `json:",inline"`

	// SupportedFeatures indicates which extended features were flagged as
	// supported by the implementation and tests will be attempted for.
	SupportedFeatures []string `json:"supportedFeatures,omitempty"`

	// UnsupportedFeatures indicates which extended features the implementation
	// does not have support for and therefore will not attempt to test.
	UnsupportedFeatures []string `json:"unsupportedFeatures,omitempty"`
}

// Status includes details on the results of a test.
type Status struct {
	Result `json:"result"`

	// Statistics includes numerical statistics on the result of the test run.
	Statistics `json:"statistics"`

	// SkippedTests indicates which tests were explicitly disabled in the test
	// suite. Skipping tests for Core level support implicitly identifies the
	// results as being partial and the implementation will not be considered
	// conformant at any level.
	SkippedTests []string `json:"skippedTests,omitempty"`

	// FailedTests indicates which tests were failing during the execution of
	// test suite.
	FailedTests []string `json:"failedTests,omitempty"`
}

func (p *ProfileReport) IsConformant(version string) error {
	if p.Name == "" {
		return errors.New("the ProfileReport must have a non-empty name")
	}

	if p.Summary == "" {
		return errors.New("the ProfileReport must have a non-empty summary")
	}

	if err := p.Core.IsConformant(version); err != nil {
		return fmt.Errorf("core feature error: %s", err)
	}

	if p.Extended != nil {
		if err := p.Extended.IsConformant(version); err != nil {
			return fmt.Errorf("extended features error: %s", err)
		}
	}

	return nil
}

func (s *Status) IsConformant(version string) error {
	if s.Result == Failure {
		return errors.New("Status showed a failure")
	}

	if s.Result == Partial {
		// TODO(youngnick): Update this to be zero skipped tests later.
		// Check how many skipped tests there were
		if semver.MajorMinor(version) == "v1.7" {
			if s.Skipped > 5 {
				return errors.New("Status showed too many skipped tests")
			}
		}
	}

	// This is a success, so we are good.
	return nil
}
