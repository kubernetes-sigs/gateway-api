/*
Copyright The Kubernetes Authors.

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
	"flag"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/yaml"

	v1 "sigs.k8s.io/gateway-api/apis/v1"
	conformanceconfig "sigs.k8s.io/gateway-api/conformance/utils/config"
	"sigs.k8s.io/gateway-api/conformance/utils/flags"
	"sigs.k8s.io/gateway-api/conformance/utils/suite"
	"sigs.k8s.io/gateway-api/pkg/features"
)

func TestConformanceOptions(t *testing.T) {
	t.Skip("this test replaces real test flags, skipping for now")
	// Ensure that conformance options provided via yaml are read from specified file.
	// Flags should take precedence over yaml file options.
	*flags.ConformanceOptionsFile = "data/test-conformance-options.yaml"

	flag.CommandLine.Set("report-output", "test-output/override.yaml")
	flag.CommandLine.Set("timeout-config-overrides", "GetTimeout:40;DefaultTestTimeout:45")

	options := DefaultOptions(t)

	// Overwritten in yaml file.
	assert.Equal(t, "testmesh", options.MeshName)
	assert.Equal(t, "placeholder", options.Mode)
	assert.Equal(t, "testproject", options.Implementation.Project)
	// Use default value.
	assert.Equal(t, "gateway-conformance", options.GatewayClassName)
	assert.Empty(t, options.RunTest)
	// Specified in yaml file, but overwritten by flag.
	assert.Equal(t, "test-output/override.yaml", options.ReportOutputPath)

	// Overwritten in yaml file.
	assert.Equal(t, 30*time.Second, options.TimeoutConfig.DeleteTimeout)
	// Use default value.
	assert.Equal(t, 60*time.Second, options.TimeoutConfig.CreateTimeout)
	assert.Equal(t, 60*time.Second, options.TimeoutConfig.RouteMustHaveParents)
	// Specified in yaml file, but overwritten by flag.
	assert.Equal(t, 40*time.Second, options.TimeoutConfig.GetTimeout)
	assert.Equal(t, 45*time.Second, options.TimeoutConfig.DefaultTestTimeout)

	// Verify SupportedFeatures unmarshalled correctly.
	assert.ElementsMatch(t, []features.FeatureName{"HTTPRouteHostRewrite", "HTTPRouteMethodMatching"}, options.SupportedFeatures)

	// Verify ExemptFeatures unmarshalled correctly.
	assert.ElementsMatch(t, []features.FeatureName{"GatewayPort8080"}, options.ExemptFeatures)

	// Verify UsableNetworkAddresses unmarshalled correctly.
	expectedUsable := []v1.GatewaySpecAddress{
		{Type: ptr.To(v1.IPAddressType), Value: "192.168.1.1"},
		{Type: ptr.To(v1.HostnameAddressType), Value: "example.com"},
	}
	assert.Equal(t, expectedUsable, options.UsableNetworkAddresses)

	// Verify UnusableNetworkAddresses unmarshalled correctly.
	expectedUnusable := []v1.GatewaySpecAddress{
		{Type: ptr.To(v1.IPAddressType), Value: "10.0.0.1"},
	}
	assert.Equal(t, expectedUnusable, options.UnusableNetworkAddresses)
}

// loadConfigurableOptionsFromYAML mirrors the YAML-loading step performed by
// DefaultOptions: defaults are seeded first, then the YAML file (if any) is
// applied on top. It deliberately omits the command-line flag overrides so
// tests can exercise the YAML path without mutating shared flag state.
func loadConfigurableOptionsFromYAML(t *testing.T, path string) *suite.ConfigurableOptions {
	t.Helper()

	opts := &suite.ConfigurableOptions{
		CleanupBaseResources: flags.DefaultCleanupBaseResources,
		CleanupTestResources: flags.DefaultCleanupTestResources,
		GatewayClassName:     flags.DefaultGatewayClassName,
		Mode:                 flags.DefaultMode,
		TimeoutConfig:        conformanceconfig.DefaultTimeoutConfig(),
	}

	if path != "" {
		data, err := os.ReadFile(path)
		require.NoError(t, err, "error reading conformance options file")
		require.NoError(t, yaml.Unmarshal(data, opts), "error unmarshalling conformance options file")
	}
	return opts
}

func TestConformanceOptions_EmptyYAML(t *testing.T) {
	// An empty YAML file must unmarshal successfully and leave all default
	// values intact — no field should be silently zeroed.
	opts := loadConfigurableOptionsFromYAML(t, "data/test-conformance-options-empty.yaml")

	defaults := conformanceconfig.DefaultTimeoutConfig()
	assert.Equal(t, flags.DefaultGatewayClassName, opts.GatewayClassName)
	assert.Equal(t, flags.DefaultMode, opts.Mode)
	assert.Equal(t, flags.DefaultCleanupBaseResources, opts.CleanupBaseResources)
	assert.Equal(t, flags.DefaultCleanupTestResources, opts.CleanupTestResources)
	assert.Equal(t, defaults, opts.TimeoutConfig)
	assert.Empty(t, opts.MeshName)
	assert.Empty(t, opts.SupportedFeatures)
	assert.Empty(t, opts.ExemptFeatures)
	assert.Empty(t, opts.UsableNetworkAddresses)
	assert.Empty(t, opts.UnusableNetworkAddresses)
}

func TestConformanceOptions_PartialYAML(t *testing.T) {
	// A YAML file that sets only a subset of fields (here meshName and a
	// single timeoutConfig entry) must override exactly those fields and
	// preserve defaults for everything else, including sibling fields
	// inside timeoutConfig.
	opts := loadConfigurableOptionsFromYAML(t, "data/test-conformance-options-partial.yaml")

	defaults := conformanceconfig.DefaultTimeoutConfig()

	assert.Equal(t, "partial-only", opts.MeshName)
	assert.Equal(t, 5*time.Second, opts.TimeoutConfig.DeleteTimeout)
	// Sibling timeout fields keep their default values.
	assert.Equal(t, defaults.CreateTimeout, opts.TimeoutConfig.CreateTimeout)
	assert.Equal(t, defaults.GetTimeout, opts.TimeoutConfig.GetTimeout)
	// Unrelated top-level fields keep their defaults.
	assert.Equal(t, flags.DefaultGatewayClassName, opts.GatewayClassName)
	assert.Equal(t, flags.DefaultMode, opts.Mode)
	assert.Equal(t, flags.DefaultCleanupBaseResources, opts.CleanupBaseResources)
	assert.Equal(t, flags.DefaultCleanupTestResources, opts.CleanupTestResources)
}
