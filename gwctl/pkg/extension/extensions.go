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

package extension

import "sigs.k8s.io/gateway-api/gwctl/pkg/topology"

type Extension interface {
	Execute(*topology.Graph) error
}

// TODO: Scope of improvement in the future involves:
//   - Making executions parallel, when there are blocking operations.
//   - Defining dependent extensions to determine their relative order.
func ExecuteAll(graph *topology.Graph, extensions ...Extension) error {
	for _, nodes := range graph.Nodes {
		for _, node := range nodes {
			if node.Metadata == nil {
				node.Metadata = make(map[string]any)
			}
		}
	}

	for _, extension := range extensions {
		err := extension.Execute(graph)
		if err != nil {
			return err
		}
	}
	return nil
}
