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

type DynamicPortRange struct {
	Start int
	End   int
}

type ListenerConfig struct {
	DynamicPortRange DynamicPortRange
}

func DefaultListenerConfig() ListenerConfig {
	return ListenerConfig{
		DynamicPortRange: DynamicPortRange{
			Start: 49152,
			End:   65535,
		},
	}
}
func SetupListenerConfig(c *ListenerConfig) {
	defaultConfig := DefaultListenerConfig()

	if c.DynamicPortRange.Start == 0 {
		c.DynamicPortRange.Start = defaultConfig.DynamicPortRange.Start
	}
	if c.DynamicPortRange.End == 0 {
		c.DynamicPortRange.End = defaultConfig.DynamicPortRange.End
	}
}
