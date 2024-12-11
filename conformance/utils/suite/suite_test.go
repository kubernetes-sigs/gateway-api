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

package suite

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/sets"

	confv1 "sigs.k8s.io/gateway-api/conformance/apis/v1"
	"sigs.k8s.io/gateway-api/pkg/consts"
	"sigs.k8s.io/gateway-api/pkg/features"
)

func TestGetAPIVersionAndChannel(t *testing.T) {
	testCases := []struct {
		name            string
		crds            []apiextensionsv1.CustomResourceDefinition
		expectedVersion string
		expectedChannel string
		err             error
	}{
		{
			name: "no Gateway API CRDs",
			err:  errors.New("no Gateway API CRDs with the proper annotations found in the cluster"),
		},
		{
			name: "properly installed Gateway API CRDs",
			crds: []apiextensionsv1.CustomResourceDefinition{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "gateways.gateway.networking.k8s.io",
						Annotations: map[string]string{
							consts.BundleVersionAnnotation: consts.BundleVersion,
							consts.ChannelAnnotation:       "standard",
						},
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "httproutes.gateway.networking.k8s.io",
						Annotations: map[string]string{
							consts.BundleVersionAnnotation: consts.BundleVersion,
							consts.ChannelAnnotation:       "standard",
						},
					},
				},
			},
			expectedVersion: consts.BundleVersion,
			expectedChannel: "standard",
		},
		{
			name: "properly installed Gateway API CRDs, with additional CRDs",
			crds: []apiextensionsv1.CustomResourceDefinition{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "gateways.gateway.networking.k8s.io",
						Annotations: map[string]string{
							consts.BundleVersionAnnotation: consts.BundleVersion,
							consts.ChannelAnnotation:       "standard",
						},
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "httproutes.gateway.networking.k8s.io",
						Annotations: map[string]string{
							consts.BundleVersionAnnotation: consts.BundleVersion,
							consts.ChannelAnnotation:       "standard",
						},
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "crd.fake.group.k8s.io",
					},
				},
			},
			expectedVersion: consts.BundleVersion,
			expectedChannel: "standard",
		},
		{
			name: "installed Gateway API CRDs having multiple versions",
			crds: []apiextensionsv1.CustomResourceDefinition{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "gateways.gateway.networking.k8s.io",
						Annotations: map[string]string{
							consts.BundleVersionAnnotation: consts.BundleVersion,
							consts.ChannelAnnotation:       "standard",
						},
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "httproutes.gateway.networking.k8s.io",
						Annotations: map[string]string{
							consts.BundleVersionAnnotation: "v2.0.0",
							consts.ChannelAnnotation:       "standard",
						},
					},
				},
			},
			err: errors.New("multiple gateway API CRDs versions detected"),
		},
		{
			name: "installed Gateway API  CRDs having multiple channels",
			crds: []apiextensionsv1.CustomResourceDefinition{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "gateways.gateway.networking.k8s.io",
						Annotations: map[string]string{
							consts.BundleVersionAnnotation: consts.BundleVersion,
							consts.ChannelAnnotation:       "standard",
						},
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "httproutes.gateway.networking.k8s.io",
						Annotations: map[string]string{
							consts.BundleVersionAnnotation: consts.BundleVersion,
							consts.ChannelAnnotation:       "experimental",
						},
					},
				},
			},
			err: errors.New("multiple gateway API CRDs channels detected"),
		},
		{
			name: "installed Gateway API CRDs having partial annotations",
			crds: []apiextensionsv1.CustomResourceDefinition{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "gateways.gateway.networking.k8s.io",
						Annotations: map[string]string{
							consts.BundleVersionAnnotation: consts.BundleVersion,
							consts.ChannelAnnotation:       "standard",
						},
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "httproutes.gateway.networking.k8s.io",
						Annotations: map[string]string{
							consts.BundleVersionAnnotation: consts.BundleVersion,
						},
					},
				},
			},
			err: errors.New("detected CRDs with partial version and channel annotations"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			version, channel, err := getAPIVersionAndChannel(tc.crds)
			assert.Equal(t, tc.expectedVersion, version)
			assert.Equal(t, tc.expectedChannel, channel)
			assert.Equal(t, tc.err, err)
		})
	}
}

const (
	coreFeature     features.FeatureName = "coreFature"
	extendedFeature features.FeatureName = "extendedFeature"

	testProfileName ConformanceProfileName = "testProfile"
)

var testProfile = ConformanceProfile{
	Name:             "testProfile",
	CoreFeatures:     sets.New(coreFeature),
	ExtendedFeatures: sets.New(extendedFeature),
}

var (
	coreTest = ConformanceTest{
		ShortName: "coreTest",
		Features:  []features.FeatureName{coreFeature},
	}
	extendedTest = ConformanceTest{
		ShortName: "extendedTest",
		Features:  []features.FeatureName{extendedFeature},
	}
	coreProvisionalTest = ConformanceTest{
		ShortName:   "coreProvisionalTest",
		Features:    []features.FeatureName{coreFeature},
		Provisional: true,
	}
	extendedProvisionalTest = ConformanceTest{
		ShortName:   "extendedProvisionalTest",
		Features:    []features.FeatureName{extendedFeature},
		Provisional: true,
	}
)

func TestSuiteReport(t *testing.T) {
	testCases := []struct {
		name                      string
		features                  sets.Set[features.FeatureName]
		extendedSupportedFeatures map[ConformanceProfileName]sets.Set[features.FeatureName]
		profiles                  sets.Set[ConformanceProfileName]
		skipProvisionalTests      bool
		results                   map[string]testResult
		expectedReport            confv1.ConformanceReport
		expectedError             error
	}{
		{
			name:     "all tests succeeded",
			features: sets.New(coreFeature, extendedFeature),
			extendedSupportedFeatures: map[ConformanceProfileName]sets.Set[features.FeatureName]{
				testProfileName: sets.New(extendedFeature),
			},
			profiles: sets.New(testProfileName),
			results: map[string]testResult{
				coreTest.ShortName: {
					result: testSucceeded,
					test:   coreTest,
				},
				extendedTest.ShortName: {
					result: testSucceeded,
					test:   extendedTest,
				},
				coreProvisionalTest.ShortName: {
					result: testSucceeded,
					test:   coreProvisionalTest,
				},
				extendedProvisionalTest.ShortName: {
					result: testSucceeded,
					test:   extendedProvisionalTest,
				},
			},
			expectedReport: confv1.ConformanceReport{
				ProfileReports: []confv1.ProfileReport{
					{
						Name:    string(testProfileName),
						Summary: "Core tests succeeded. Extended tests succeeded.",
						Core: confv1.Status{
							Result: confv1.Success,
							Statistics: confv1.Statistics{
								Passed: 2,
							},
						},
						Extended: &confv1.ExtendedStatus{
							Status: confv1.Status{
								Result: confv1.Success,
								Statistics: confv1.Statistics{
									Passed: 2,
								},
							},
							SupportedFeatures: []string{string(extendedFeature)},
						},
					},
				},
				SucceededProvisionalTests: []string{
					coreProvisionalTest.ShortName,
					extendedProvisionalTest.ShortName,
				},
			},
		},
		{
			name:     "mixed results",
			features: sets.New(coreFeature, extendedFeature),
			extendedSupportedFeatures: map[ConformanceProfileName]sets.Set[features.FeatureName]{
				testProfileName: sets.New(extendedFeature),
			},
			profiles: sets.New(testProfileName),
			results: map[string]testResult{
				coreTest.ShortName: {
					result: testFailed,
					test:   coreTest,
				},
				extendedTest.ShortName: {
					result: testSkipped,
					test:   extendedTest,
				},
				coreProvisionalTest.ShortName: {
					result: testSucceeded,
					test:   coreProvisionalTest,
				},
				extendedProvisionalTest.ShortName: {
					result: testProvisionalSkipped,
					test:   extendedProvisionalTest,
				},
			},
			expectedReport: confv1.ConformanceReport{
				ProfileReports: []confv1.ProfileReport{
					{
						Name:    string(testProfileName),
						Summary: "Core tests failed with 1 test failures. Extended tests partially succeeded with 1 test skips.",
						Core: confv1.Status{
							Result: confv1.Failure,
							Statistics: confv1.Statistics{
								Passed: 1,
								Failed: 1,
							},
							FailedTests: []string{
								coreTest.ShortName,
							},
						},
						Extended: &confv1.ExtendedStatus{
							Status: confv1.Status{
								Result: confv1.Partial,
								Statistics: confv1.Statistics{
									Skipped: 1,
								},
								SkippedTests: []string{
									extendedTest.ShortName,
								},
							},
							SupportedFeatures: []string{string(extendedFeature)},
						},
					},
				},
				SucceededProvisionalTests: []string{
					coreProvisionalTest.ShortName,
				},
			},
		},
		{
			name:     "skip provisional tests",
			features: sets.New(coreFeature, extendedFeature),
			extendedSupportedFeatures: map[ConformanceProfileName]sets.Set[features.FeatureName]{
				testProfileName: sets.New(extendedFeature),
			},
			profiles:             sets.New(testProfileName),
			skipProvisionalTests: true,
			results: map[string]testResult{
				coreTest.ShortName: {
					result: testSucceeded,
					test:   coreTest,
				},
				extendedTest.ShortName: {
					result: testSucceeded,
					test:   extendedTest,
				},
				coreProvisionalTest.ShortName: {
					result: testProvisionalSkipped,
					test:   coreProvisionalTest,
				},
				extendedProvisionalTest.ShortName: {
					result: testProvisionalSkipped,
					test:   extendedProvisionalTest,
				},
			},
			expectedReport: confv1.ConformanceReport{
				ProfileReports: []confv1.ProfileReport{
					{
						Name:    string(testProfileName),
						Summary: "Core tests succeeded. Extended tests succeeded.",
						Core: confv1.Status{
							Result: confv1.Success,
							Statistics: confv1.Statistics{
								Passed: 1,
							},
						},
						Extended: &confv1.ExtendedStatus{
							Status: confv1.Status{
								Result: confv1.Success,
								Statistics: confv1.Statistics{
									Passed: 1,
								},
							},
							SupportedFeatures: []string{string(extendedFeature)},
						},
					},
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			conformanceProfileMap[testProfileName] = testProfile

			suite := ConformanceTestSuite{
				conformanceProfiles:       tc.profiles,
				SupportedFeatures:         tc.features,
				extendedSupportedFeatures: tc.extendedSupportedFeatures,
				results:                   tc.results,
				SkipProvisionalTests:      tc.skipProvisionalTests,
			}
			report, err := suite.Report()
			assert.Equal(t, tc.expectedReport.ProfileReports, report.ProfileReports)
			assert.Equal(t, tc.expectedReport.SucceededProvisionalTests, report.SucceededProvisionalTests)
			assert.Equal(t, tc.expectedError, err)
		})
	}
}
