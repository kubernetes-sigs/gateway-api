/*
Copyright 2026 The Kubernetes Authors.

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
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/utils/ptr"
	v1 "sigs.k8s.io/gateway-api/apis/v1"
	"sigs.k8s.io/gateway-api/conformance/utils/flags"
	"sigs.k8s.io/gateway-api/pkg/features"
)

func TestConformanceOptions(t *testing.T) {
	// Ensure that conformance options provided via yaml are read from specified file.
	// Flags should take precedence over yaml file options.
	*flags.ConformanceOptionsFile = "data/test-conformance-options.yaml"

	flag.CommandLine.Set("report-output", "test-output/override.yaml")
	flag.CommandLine.Set("timeout-config-overrides", "GetTimeout:40;DefaultTestTimeout:45")

	options := DefaultOptions(t)

	// Overwritten in yaml file.
	assert.Equal(t, "istio", options.MeshName)
	assert.Equal(t, "placeholder", options.Mode)
	assert.Equal(t, "istio", options.Implementation.Project)
	// Use default value.
	assert.Equal(t, "gateway-conformance", options.GatewayClassName)
	assert.Equal(t, "", options.RunTest)
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
	expectedSupported := sets.New[features.FeatureName]("HTTPRouteHostRewrite", "HTTPRouteMethodMatching")
	assert.True(t, options.SupportedFeatures.Equal(expectedSupported), "SupportedFeatures mismatch: got %v", options.SupportedFeatures.UnsortedList())

	// Verify ExemptFeatures unmarshalled correctly.
	expectedExempt := sets.New[features.FeatureName]("GatewayPort8080")
	assert.True(t, options.ExemptFeatures.Equal(expectedExempt), "ExemptFeatures mismatch: got %v", options.ExemptFeatures.UnsortedList())

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
