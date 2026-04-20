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

	//"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/assert"

	"sigs.k8s.io/gateway-api/conformance/utils/flags"
)

func TestConformanceOptions(t *testing.T) {
	// Ensure that conformance options provided via yaml are read from specified file.
	// Flags should take precedence over yaml file options.
	*flags.ConformanceOptionsFile = "data/conformance-options.yaml"

	flag.CommandLine.Set("report-output", "test-output/override.yaml")
	flag.CommandLine.Set("timeout-config-overrides", "GetTimeout:40;DefaultTestTimeout:45")

	options := DefaultOptions(t)

	// Overwritten in yaml file.
	assert.Equal(t, "istio", options.MeshName)
	assert.Equal(t, "placeholder", options.Mode)
	// Use default value.
	assert.Equal(t, "gateway-conformance", options.GatewayClassName)
	// Specified in yaml file, but overwritten by flag.
	assert.Equal(t, "test-output/override.yaml", options.ReportOutputPath)

	// Overwritten in yaml file.
	assert.Equal(t, 30 * time.Second, options.TimeoutConfig.DeleteTimeout)
	// Use default value.
	assert.Equal(t, 60 * time.Second, options.TimeoutConfig.CreateTimeout)
	assert.Equal(t, 60 * time.Second, options.TimeoutConfig.RouteMustHaveParents)
	// Specified in yaml file, but overwritten by flag.
	assert.Equal(t, 40 * time.Second, options.TimeoutConfig.GetTimeout)
	assert.Equal(t, 45 * time.Second, options.TimeoutConfig.DefaultTestTimeout)
}
