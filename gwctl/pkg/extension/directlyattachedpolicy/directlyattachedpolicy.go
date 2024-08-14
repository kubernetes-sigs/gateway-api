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

package directlyattachedpolicy

import (
	"fmt"

	"k8s.io/klog/v2"

	"sigs.k8s.io/gateway-api/gwctl/pkg/common"
	"sigs.k8s.io/gateway-api/gwctl/pkg/policymanager"
	"sigs.k8s.io/gateway-api/gwctl/pkg/topology"
)

const (
	extensionName = "DirectlyAttachedPolicy"
)

type Extension struct {
	policyManager *policymanager.PolicyManager
}

func NewExtension(policyManager *policymanager.PolicyManager) *Extension {
	return &Extension{policyManager: policyManager}
}

func (a *Extension) Execute(graph *topology.Graph) error {
	graph.RemoveMetadata(extensionName)
	for _, policy := range a.policyManager.GetPolicies() {
		gk := policy.TargetRef.GroupKind()
		nn := policy.TargetRef.NamespacedName()

		if graph.Nodes[gk] == nil || graph.Nodes[gk][nn] == nil {
			// This target doesn't exist in the graph, so skip the policy.
			continue
		}

		node := graph.Nodes[gk][nn]
		if node.Metadata == nil {
			node.Metadata = map[string]any{}
		}
		if node.Metadata[extensionName] == nil {
			node.Metadata[extensionName] = map[common.GKNN]*policymanager.Policy{}
		}

		data, err := Access(node)
		if err != nil {
			return err
		}
		data[policy.GKNN()] = policy
	}
	return nil
}

func Access(node *topology.Node) (map[common.GKNN]*policymanager.Policy, error) {
	rawData, ok := node.Metadata[extensionName]
	if !ok || rawData == nil {
		klog.V(3).InfoS(fmt.Sprintf("no data found in node for %v", extensionName), "node", node.GKNN())
		return nil, nil
	}
	data, ok := rawData.(map[common.GKNN]*policymanager.Policy)
	if !ok {
		return nil, fmt.Errorf("unable to perform type assertion for %v in node %v", extensionName, node.GKNN())
	}
	return data, nil
}
