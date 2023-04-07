//go:build experimental
// +build experimental

/*
Copyright 2022 The Kubernetes Authors.

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
	confv1a1 "sigs.k8s.io/gateway-api/conformance/apis/v1alpha1"
)

// -----------------------------------------------------------------------------
// ConformanceReport - Private Types
// -----------------------------------------------------------------------------

type testResult struct {
	test      ConformanceTest
	succeeded bool
}

type profileReportsMap map[ConformanceProfileName]confv1a1.ProfileReport

func newReports() profileReportsMap {
	return make(profileReportsMap)
}

func (p profileReportsMap) addTestResults(testResults ...testResult) error {
	for _, testResult := range testResults {
		conformanceProfile, err := getConformanceProfileForTest(testResult.test.ShortName)
		if err != nil {
			return err
		}

		// initialize the profile report if not already initialized
		if _, ok := p[conformanceProfile.Name]; !ok {
			p[conformanceProfile.Name] = confv1a1.ProfileReport{
				Name: string(conformanceProfile.Name),
			}
		}

		// TODO: refactor and clean this up later
		testIsExtended := isTestExtended(conformanceProfile, testResult.test)
		if testResult.succeeded {
			if testIsExtended {
				extended := p[conformanceProfile.Name].Extended
				extended.Statistics.Passed++
			} else {
				core := p[conformanceProfile.Name].Core
				core.Statistics.Passed++
			}
		} else {
			if testIsExtended {
				extended := p[conformanceProfile.Name].Extended
				extended.Statistics.Failed++
			} else {
				core := p[conformanceProfile.Name].Core
				core.Statistics.Failed++
			}
		}
	}

	return nil
}

func (p profileReportsMap) list() (profileReports []confv1a1.ProfileReport) {
	for _, profileReport := range p {
		profileReports = append(profileReports, profileReport)
	}
	return
}

func (p profileReportsMap) compileResults() {
	for _, report := range p {
		// report the overall result for core features
		if report.Core.Passed == 0 || report.Core.Failed > 0 {
			report.Core.Result = confv1a1.Failure
		} else if report.Core.Skipped > 0 {
			report.Core.Result = confv1a1.Partial
		} else {
			report.Core.Result = confv1a1.Success
		}

		// report the overall result for extended features
		if report.Extended.Passed == 0 || report.Extended.Failed > 0 {
			report.Extended.Result = confv1a1.Failure
		} else if report.Extended.Skipped > 0 {
			report.Extended.Result = confv1a1.Partial
		} else {
			report.Extended.Result = confv1a1.Success
		}
	}
}

// -----------------------------------------------------------------------------
// ConformanceReport - Private Helper Functions
// -----------------------------------------------------------------------------

// isTestExtended determines if a provided test is considered to be supported
// at an extended level of support given the provided conformance profile.
//
// TODO: right now the tests themselves don't indicate the conformance
// support level associated with them. The only way we have right now
// in this prototype to know whether a test belongs to any particular
// conformance level is to compare the features needed for the test to
// the conformance profiles known list of core vs extended features.
// Later if we move out of Prototyping/Provisional it would probably
// be best to indicate the conformance support level of each test, but
// for now this hack works.
func isTestExtended(profile ConformanceProfile, test ConformanceTest) bool {
	for _, supportedFeature := range test.Features {
		// if ANY of the features needed for the test are extended features,
		// then we consider the entire test extended level support.
		if profile.ExtendedFeatures.Has(supportedFeature) {
			return true
		}
	}
	return false
}
