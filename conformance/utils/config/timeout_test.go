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

package config

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSetupTimeoutConfigAppliesAllDefaults(t *testing.T) {
	cfg := TimeoutConfig{}

	require.NoError(t, json.Unmarshal([]byte(`{"createTimeout":"5s"}`), &cfg))

	SetupTimeoutConfig(&cfg)

	defaults := DefaultTimeoutConfig()

	assert.Equal(t, 5, int(cfg.CreateTimeout.Seconds()))
	assert.Equal(t, defaults.ListenerSetMustHaveCondition, cfg.ListenerSetMustHaveCondition)
	assert.Equal(t, defaults.ListenerSetListenersMustHaveConditions, cfg.ListenerSetListenersMustHaveConditions)
	assert.Equal(t, defaults.RequiredConsecutiveSuccesses, cfg.RequiredConsecutiveSuccesses)
}
