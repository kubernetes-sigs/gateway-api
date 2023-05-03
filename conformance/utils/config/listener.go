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

package config

// DynamicPortRange specifies the starting and ending port of a
// range
type DynamicPortRange struct {
	Start int
	End   int
}

// ListenerConfig allow conformance test runners to configure
type ListenerConfig struct {
	DynamicPortRange DynamicPortRange
}

// DefaultListenerConfig returns a [ListenerConfig] where the [DynamicPortRange]
// defaults to the suggested IANA [dynamic port range] (49152-65535)
//
// [dynamic port range]: https://www.iana.org/assignments/service-names-port-numbers/service-names-port-numbers.xhtml
func DefaultListenerConfig() ListenerConfig {
	return ListenerConfig{
		DynamicPortRange: DynamicPortRange{
			Start: 49152,
			End:   65535,
		},
	}
}

// SetupListenerConfig will apply defaults to the passed [ListenerConfig]
func SetupListenerConfig(c *ListenerConfig) {
	defaultConfig := DefaultListenerConfig()

	if c.DynamicPortRange.Start == 0 {
		c.DynamicPortRange.Start = defaultConfig.DynamicPortRange.Start
	}
	if c.DynamicPortRange.End == 0 {
		c.DynamicPortRange.End = defaultConfig.DynamicPortRange.End
	}
}
