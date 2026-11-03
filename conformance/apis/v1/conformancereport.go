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
	"net/url"
	"strings"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ConformanceReport is a report of conformance testing results including the
// specific conformance profiles that were tested and the results of the tests
// with summaries and statistics.
type ConformanceReport struct {
	metav1.TypeMeta `json:",inline"`
	Implementation  `json:"implementation"`

	// Date indicates the date that this report was generated.
	Date string `json:"date"`

	// GatewayAPIVersion indicates which release version of Gateway API this
	// test report was made for.
	GatewayAPIVersion string `json:"gatewayAPIVersion"`

	// Mode is the operating mode the implementation used to run conformance tests.
	Mode string `json:"mode"`

	// GatewayAPIChannel indicates which release channel of Gateway API this
	// test report was made for.
	GatewayAPIChannel string `json:"gatewayAPIChannel"`

	// ProfileReports is a list of the individual reports for each conformance
	// profile that was enabled for a test run.
	ProfileReports []ProfileReport `json:"profiles"`

	// SucceededProvisionalTests is a list of the names of the provisional tests that
	// have been successfully run.
	SucceededProvisionalTests []string `json:"succeededProvisionalTests,omitempty"`
}

// Implementation provides metadata information on the downstream
// implementation of Gateway API which ran conformance tests.
type Implementation struct {
	// Organization refers to the company, group or individual which maintains
	// the named implementation. Organizations can provide reports for any
	// number of distinct Gateway API implementations they maintain, but need
	// to identify themselves using this organization field for grouping.
	Organization string `json:"organization"`

	// Project indicates the name of the project or repository for a Gateway API
	// implementation.
	Project string `json:"project"`

	// URL indicates a human-usable URL where more information about the
	// implementation can be found. For open source projects this should
	// generally link to the code repository.
	URL string `json:"url"`

	// Version indicates the version of the implementation that was used for
	// testing. This should generally be a semver version when applicable.
	Version string `json:"version"`

	// Contact is contact information for the maintainers so that Gateway API
	// maintainers can get ahold of them as needed. Ideally this should be
	// GitHub usernames (in the form of `@<username>`) or team names (in the
	// form of `@<team>/<name>`), but when that's not possible it can be email
	// addresses.
	// Rather than GitHub usernames or email addresses you can provide a URL to the relevant
	// support pages for the project. Ideally this would be something like the issue creation page
	// on a repository, but for projects without a publicly exposed repository a general support
	// page URL can be provided.
	Contact []string `json:"contact"`
}

// Validate ensures that the Implementation struct has valid fields set.
// See https://github.com/kubernetes-sigs/gateway-api/issues/2178.
func (i *Implementation) Validate() error {
	if i.Organization == "" {
		return errors.New("implementation's organization cannot be empty")
	}
	if i.Project == "" {
		return errors.New("implementation's project cannot be empty")
	}
	if i.URL == "" {
		return errors.New("implementation's url cannot be empty")
	}
	if err := validateURL(i.URL); err != nil {
		return fmt.Errorf("implementation's url is invalid: %w", err)
	}
	if i.Version == "" {
		return errors.New("implementation's version cannot be empty")
	}
	if len(i.Contact) == 0 {
		return errors.New("implementation's contact cannot be empty")
	}
	return nil
}

// validateURL ensures that rawURL is a well-formed, absolute URL with a
// recognized http(s) scheme and a non-empty host, so that it can reliably
// point maintainers to more information about an implementation.
func validateURL(rawURL string) error {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return fmt.Errorf("could not parse url: %w", err)
	}
	if !parsed.IsAbs() {
		return fmt.Errorf("url must be absolute (missing scheme) %s", rawURL)
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return fmt.Errorf("url scheme must be http or https, got %q", parsed.Scheme)
	}
	if parsed.Host == "" {
		return errors.New("url must include a host")
	}
	return nil
}

// IsConformant returns Result values for Gateway, Mesh, and a slice of errors that describe
// if something went wrong.
//
// Note that this does not use errors.Join because we need the errors to be presented without
// newlines in eventual reports, and that's surprisingly difficult with Join-ed errors.
func (c *ConformanceReport) IsConformant(version string) (*Result, *Result, []error, []error) {
	// All profiles must be conformant for the report to be conformant.
	// For results, Failure > Partial > Success, and the greatest
	// value will be used.
	//
	// So, if one profile is Partial, then Partial is the best you get,
	// per traffic direction. So for Gateway and Mesh, Success or Partial are
	// orthogonal. You can be Conformant for Gateway, and Partial for Mesh,
	// or any other combination.

	// Because of the above, we need to track the result per direction, as well
	// as if the report contains a result at all for each direction.
	// (otherwise we can't distinguish between "has Gateway, failed" and "does not
	// have Gateway")
	gatewayResult := Success
	meshResult := Success

	hasGateway := false
	hasMesh := false

	var gatewayErrs []error
	var meshErrs []error

	var totalSkips int

	for _, profile := range c.ProfileReports {
		profileResult, skips, errs := profile.IsConformant(version)
		totalSkips += skips

		if strings.HasPrefix(profile.Name, "MESH") {
			hasMesh = true
			if len(errs) > 0 {
				meshErrs = append(meshErrs, errs...)
			}
			meshResult = updateResult(meshResult, profileResult)
		} else {
			hasGateway = true
			if len(errs) > 0 {
				gatewayErrs = append(gatewayErrs, errs...)
			}
			gatewayResult = updateResult(gatewayResult, profileResult)
		}

	}

	// The Implementation struct must be valid.
	if err := c.Validate(); err != nil {
		if hasGateway {
			gatewayErrs = append(gatewayErrs, fmt.Errorf("Implementation details were not valid: %s", err))
			gatewayResult = updateResult(gatewayResult, Failure)
		}
		if hasMesh {
			meshErrs = append(meshErrs, fmt.Errorf("Implementation details were not valid: %s", err))
			meshResult = updateResult(meshResult, Failure)
		}

	}

	// When we've checked all the profiles in a report, check to see how many test skips in total
	// there were.
	// As of v1.7, if there are more than 5 test skips in total, then the report is _not_ conformant,
	// and will fail _both_ Mesh and Gateway conformance directions.
	if totalSkips > 5 {
		if hasGateway {
			gatewayErrs = append(gatewayErrs, fmt.Errorf("Too many tests were skipped across the report: %d ", totalSkips))
			gatewayResult = updateResult(gatewayResult, Failure)
		}
		if hasMesh {
			meshErrs = append(meshErrs, fmt.Errorf("Too many tests were skipped across the report: %d ", totalSkips))
			meshResult = updateResult(meshResult, Failure)
		}
	}

	retGatewayResult := &gatewayResult

	if !hasGateway {
		retGatewayResult = nil
	}

	retMeshResult := &meshResult

	if !hasMesh {
		retMeshResult = nil
	}

	return retGatewayResult, retMeshResult, gatewayErrs, meshErrs
}

// updateResult updates the Result values after comparing them
// For results, Failure > Partial > Success
func updateResult(oldResult, newResult Result) Result {
	if oldResult == newResult {
		return oldResult
	}

	if oldResult == Failure {
		return oldResult
	}

	if oldResult == Partial {
		if newResult == Success {
			return oldResult
		}
		if newResult == Failure {
			return newResult
		}
	}

	return newResult
}
