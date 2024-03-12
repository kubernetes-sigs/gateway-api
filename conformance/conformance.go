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

package conformance

import (
	"os"
	"testing"

	confv1a1 "sigs.k8s.io/gateway-api/conformance/apis/v1alpha1"
	"sigs.k8s.io/gateway-api/conformance/tests"
	"sigs.k8s.io/gateway-api/conformance/utils/flags"
	"sigs.k8s.io/gateway-api/conformance/utils/suite"

	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

// DefaultOptions will parse the test binary flags and create a configuration
// that can be used to run Gateway API conformance tests.
//
// End-users downstream can override options after instantiation
func DefaultOptions(t *testing.T) *suite.ExperimentalConformanceOptions {
	return &suite.ExperimentalConformanceOptions{
		Options:             *DefaultLegacyOptions(t),
		Mode:                *flags.Mode,
		AllowCRDsMismatch:   *flags.AllowCRDsMismatch,
		ConformanceProfiles: suite.ParseConformanceProfiles(*flags.ConformanceProfiles),
		ReportOutputPath:    *flags.ReportOutput,
		Implementation: suite.ParseImplementation(
			*flags.ImplementationOrganization,
			*flags.ImplementationProject,
			*flags.ImplementationURL,
			*flags.ImplementationVersion,
			*flags.ImplementationContact,
		),
	}
}

// RunConformance will run the Gateway API conformance tests given the supplied options.
// If options is nil then the DefaultOptions will be instantiated
func RunConformance(t *testing.T, opts *suite.ExperimentalConformanceOptions) {
	if opts == nil {
		opts = DefaultOptions(t)
	}

	if opts.ConformanceProfiles.Len() == 0 {
		RunLegacyConformance(t, &opts.Options)
		return
	}

	// Validate implementation if we are not an end user
	if opts.ReportOutputPath != "" {
		err := suite.ValidateImplementation(opts.Implementation)
		require.NoError(t, err, "Error parsing implementation's details")
	}

	t.Log("Running experimental conformance tests with:")
	logOptions(t, &opts.Options)

	cSuite, err := suite.NewExperimentalConformanceTestSuite(*opts)
	if err != nil {
		t.Fatalf("error creating experimental conformance test suite: %v", err)
	}

	cSuite.Setup(t)
	require.NoError(t, cSuite.Run(t, tests.ConformanceTests))

	report, err := cSuite.Report()
	require.NoError(t, err, "error generating conformance profile report")
	require.NoError(t, writeReport(t.Logf, *report, opts.ReportOutputPath), "error writing report")
}

func writeReport(logf func(string, ...any), report confv1a1.ConformanceReport, output string) error {
	//nolint:musttag // the linter is complaining report doesn't have yaml tags - but it has json ones
	rawReport, err := yaml.Marshal(report)
	if err != nil {
		return err
	}

	if output != "" {
		if err = os.WriteFile(output, rawReport, 0o600); err != nil {
			return err
		}
	}
	logf("Conformance report:\n%s", string(rawReport))

	return nil
}
