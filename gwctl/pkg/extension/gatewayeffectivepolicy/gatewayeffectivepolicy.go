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

package gatewayeffectivepolicy

import (
	"fmt"
	"maps"

	"k8s.io/klog/v2"

	"sigs.k8s.io/gateway-api/gwctl/pkg/common"
	"sigs.k8s.io/gateway-api/gwctl/pkg/extension/directlyattachedpolicy"
	"sigs.k8s.io/gateway-api/gwctl/pkg/policymanager"
	"sigs.k8s.io/gateway-api/gwctl/pkg/topology"
	topologygw "sigs.k8s.io/gateway-api/gwctl/pkg/topology/gateway"
)

const (
	extensionName = "InheritedPolicy"
)

type Extension struct{}

func NewExtension() *Extension {
	return &Extension{}
}

// Extension calculates the effective policies for all Gateways, HTTPRoutes, and
// Backends in the Graph.
func (a *Extension) Execute(graph *topology.Graph) error {
	graph.RemoveMetadata(extensionName)
	if err := a.calculateInheritedPolicies(graph); err != nil {
		return err
	}
	return a.calculateEffectivePolicies(graph)
}

// calculateInheritedPolicies calculates the inherited polices for all Gateways,
// HTTRoutes, and Backends in the Graph.
func (a *Extension) calculateInheritedPolicies(graph *topology.Graph) error {
	if err := a.calculateInheritedPoliciesForGateways(graph); err != nil {
		return err
	}
	if err := a.calculateInheritedPoliciesForHTTPRoutes(graph); err != nil {
		return err
	}
	if err := a.calculateInheritedPoliciesForBackends(graph); err != nil {
		return err
	}
	return nil
}

// calculateInheritedPoliciesForGateways calculates the inherited policies for
// all Gateways present in the Graph.
func (a *Extension) calculateInheritedPoliciesForGateways(graph *topology.Graph) error {
	for _, gatewayNode := range graph.Nodes[common.GatewayGK] {
		result := make(map[common.GKNN]*policymanager.Policy)

		// Policies inherited from Gateway's namespace.
		namespaceNode := topologygw.GatewayNode(gatewayNode).Namespace()
		if namespaceNode != nil {
			namespacePoliciesMap, err := directlyattachedpolicy.Access(namespaceNode)
			if err != nil {
				return err
			}
			maps.Copy(result, filterInheritablePolicies(namespacePoliciesMap))
		}

		// Policies inherited from GatewayClass.
		gatewayClassNode := topologygw.GatewayNode(gatewayNode).GatewayClass()
		if gatewayClassNode != nil {
			gatewayClassPoliciesMap, err := directlyattachedpolicy.Access(gatewayClassNode)
			if err != nil {
				return err
			}
			maps.Copy(result, filterInheritablePolicies(gatewayClassPoliciesMap))
		}

		gatewayNode.Metadata[extensionName] = &NodeMetadata{GatewayInheritedPolicies: result}
	}
	return nil
}

// calculateInheritedPoliciesForHTTPRoutes calculates the inherited policies for
// all HTTPRoutes present in the Graph.
func (a *Extension) calculateInheritedPoliciesForHTTPRoutes(graph *topology.Graph) error {
	for _, httpRouteNode := range graph.Nodes[common.HTTPRouteGK] {
		result := make(map[common.GKNN]*policymanager.Policy)

		// Policies inherited from HTTPRoute's namespace.
		namespaceNode := topologygw.HTTPRouteNode(httpRouteNode).Namespace()
		if namespaceNode != nil {
			namespacePoliciesMap, err := directlyattachedpolicy.Access(namespaceNode)
			if err != nil {
				return err
			}
			maps.Copy(result, filterInheritablePolicies(namespacePoliciesMap))
		}

		// Policies inherited from Gateways.
		gatewayNodes := topologygw.HTTPRouteNode(httpRouteNode).Gateways()
		if gatewayNodes != nil {
			for _, gatewayNode := range gatewayNodes {
				// Add policies inherited by GatewayNode.
				effPolicyMetadata, err := Access(gatewayNode)
				if err != nil {
					return err
				}
				if effPolicyMetadata != nil {
					maps.Copy(result, effPolicyMetadata.GatewayInheritedPolicies)
				}

				// Add inheritable policies directly applied to GatewayNode.
				gatewayPoliciesMap, err := directlyattachedpolicy.Access(gatewayNode)
				if err != nil {
					return err
				}
				maps.Copy(result, filterInheritablePolicies(gatewayPoliciesMap))
			}
		}

		httpRouteNode.Metadata[extensionName] = &NodeMetadata{HTTPRouteInheritedPolicies: result}
	}
	return nil
}

// calculateInheritedPoliciesForBackends calculates the inherited policies for
// all Backends present in ResourceModel.
func (a *Extension) calculateInheritedPoliciesForBackends(graph *topology.Graph) error {
	for _, backendNode := range graph.Nodes[common.ServiceGK] {
		result := make(map[common.GKNN]*policymanager.Policy)

		// Policies inherited from Backend's namespace.
		namespaceNode := topologygw.BackendNode(backendNode).Namespace()
		if namespaceNode != nil {
			namespacePoliciesMap, err := directlyattachedpolicy.Access(namespaceNode)
			if err != nil {
				return err
			}
			maps.Copy(result, filterInheritablePolicies(namespacePoliciesMap))
		}

		// Policies inherited from HTTPRoutes.
		httpRouteNodes := topologygw.BackendNode(backendNode).HTTPRoutes()
		if httpRouteNodes != nil {
			for _, httpRouteNode := range httpRouteNodes {
				// Add policies inherited by HTTPRouteNode.
				effPolicyMetadata, err := Access(httpRouteNode)
				if err != nil {
					return err
				}
				if effPolicyMetadata != nil {
					maps.Copy(result, effPolicyMetadata.HTTPRouteInheritedPolicies)
				}

				// Add inheritable policies directly applied to HTTPRouteNode.
				httpRoutePoliciesMap, err := directlyattachedpolicy.Access(httpRouteNode)
				if err != nil {
					return err
				}
				maps.Copy(result, filterInheritablePolicies(httpRoutePoliciesMap))
			}
		}

		backendNode.Metadata[extensionName] = &NodeMetadata{BackendInheritedPolicies: result}
	}
	return nil
}

// filterInheritablePolicies filters and returns policies which can be inherited.
func filterInheritablePolicies(policies map[common.GKNN]*policymanager.Policy) map[common.GKNN]*policymanager.Policy {
	inheritablePolicies := make(map[common.GKNN]*policymanager.Policy)

	for gknn, policy := range policies {
		if policy.IsInheritable() {
			inheritablePolicies[gknn] = policy
		}
	}

	return inheritablePolicies
}

func (a *Extension) calculateEffectivePolicies(graph *topology.Graph) error {
	if err := a.calculateEffectivePoliciesForGateways(graph); err != nil {
		return err
	}
	if err := a.calculateEffectivePoliciesForHTTPRoutes(graph); err != nil {
		return err
	}
	if err := a.calculateEffectivePoliciesForBackends(graph); err != nil {
		return err
	}
	return nil
}

// calculateEffectivePoliciesForGateways calculates the effective policies for
// each Gateway by merging policies from different hierarchies (GatewayClass,
// Namespace, and Gateway).
func (a *Extension) calculateEffectivePoliciesForGateways(graph *topology.Graph) error {
	for _, gatewayNode := range graph.Nodes[common.GatewayGK] {
		if gatewayNode.Depth > graph.MaxDepth {
			continue
		}

		gatewayClassNode := topologygw.GatewayNode(gatewayNode).GatewayClass()
		if gatewayClassNode == nil {
			klog.V(3).InfoS("No GatewayClass node found for Gateway, skipping effective policy calculation", "gateway", gatewayNode.GKNN())
			continue
		}
		namespaceNode := topologygw.GatewayNode(gatewayNode).Namespace()
		if namespaceNode == nil {
			klog.V(3).InfoS("No Namespace node found for Gateway, skipping effective policy calculation", "gateway", gatewayNode.GKNN())
			continue
		}

		gatewayClassPoliciesMap, err := directlyattachedpolicy.Access(gatewayClassNode)
		if err != nil {
			return err
		}
		namespacePoliciesMap, err := directlyattachedpolicy.Access(namespaceNode)
		if err != nil {
			return err
		}
		gatewayPoliciesMap, err := directlyattachedpolicy.Access(gatewayNode)
		if err != nil {
			return err
		}

		// Do not calculate effective policy for the Gateway if the referenced
		// GatewayClass does not exist. For now, we only calculate effective policy
		// once the references are corrected.
		if gatewayClassNode == nil {
			continue
		}

		// Fetch all policies.
		gatewayClassPolicies := policymanager.ConvertPoliciesMapToSlice(filterInheritablePolicies(gatewayClassPoliciesMap))
		gatewayNamespacePolicies := policymanager.ConvertPoliciesMapToSlice(filterInheritablePolicies(namespacePoliciesMap))
		gatewayPolicies := policymanager.ConvertPoliciesMapToSlice(filterInheritablePolicies(gatewayPoliciesMap))

		// Merge policies by their kind.
		gatewayClassPoliciesByKind, err := policymanager.MergePoliciesOfSimilarKind(gatewayClassPolicies)
		if err != nil {
			return err
		}
		gatewayNamespacePoliciesByKind, err := policymanager.MergePoliciesOfSimilarKind(gatewayNamespacePolicies)
		if err != nil {
			return err
		}
		gatewayPoliciesByKind, err := policymanager.MergePoliciesOfSimilarKind(gatewayPolicies)
		if err != nil {
			return err
		}

		// Merge all hierarchial policies.
		result, err := policymanager.MergePoliciesOfDifferentHierarchy(gatewayClassPoliciesByKind, gatewayNamespacePoliciesByKind)
		if err != nil {
			return err
		}

		result, err = policymanager.MergePoliciesOfDifferentHierarchy(result, gatewayPoliciesByKind)
		if err != nil {
			return err
		}

		gatewayNodeMetadata, err := Access(gatewayNode)
		if err != nil {
			return err
		}
		if gatewayNodeMetadata == nil {
			gatewayNodeMetadata = &NodeMetadata{}
			gatewayNode.Metadata[extensionName] = gatewayNodeMetadata
		}
		gatewayNodeMetadata.GatewayEffectivePolicies = result
	}
	return nil
}

// calculateEffectivePoliciesForHTTPRoutes calculates the effective policies for
// each HTTPRoute, taking into account policies from different hierarchies
// (GatewayClass, Namespace, Gateway, and HTTPRoute).
func (a *Extension) calculateEffectivePoliciesForHTTPRoutes(graph *topology.Graph) error {
	for _, httpRouteNode := range graph.Nodes[common.HTTPRouteGK] {
		result := make(map[common.GKNN]map[policymanager.PolicyCrdID]*policymanager.Policy)

		namespaceNode := topologygw.HTTPRouteNode(httpRouteNode).Namespace()
		if namespaceNode == nil {
			klog.V(3).InfoS("No Namespace node found for HTTPRoute, skipping effective policy calculation", "httpRoute", httpRouteNode.GKNN())
			continue
		}

		httpRoutePoliciesMap, err := directlyattachedpolicy.Access(httpRouteNode)
		if err != nil {
			return err
		}
		namespacePoliciesMap, err := directlyattachedpolicy.Access(namespaceNode)
		if err != nil {
			return err
		}

		// Step 1: Aggregate all policies of the HTTPRoute and the
		// HTTPRoute-namespace.
		httpRoutePolicies := policymanager.ConvertPoliciesMapToSlice(filterInheritablePolicies(httpRoutePoliciesMap))
		httpRouteNamespacePolicies := policymanager.ConvertPoliciesMapToSlice(filterInheritablePolicies(namespacePoliciesMap))

		// Step 2: Merge HTTPRoute and HTTPRoute-namespace policies by their kind.
		httpRoutePoliciesByKind, err := policymanager.MergePoliciesOfSimilarKind(httpRoutePolicies)
		if err != nil {
			return err
		}
		httpRouteNamespacePoliciesByKind, err := policymanager.MergePoliciesOfSimilarKind(httpRouteNamespacePolicies)
		if err != nil {
			return err
		}

		// Step 3: Loop through all Gateways and merge policies for each Gateway.
		// End result is we get policies partitioned by each Gateway.
		for gatewayGKNN, gatewayNode := range topologygw.HTTPRouteNode(httpRouteNode).Gateways() {
			gatewayNodeMetadata, err := Access(gatewayNode) //nolint:govet
			if err != nil {
				return err
			}
			gatewayPoliciesByKind := gatewayNodeMetadata.GatewayEffectivePolicies

			// Merge all hierarchial policies.
			mergedPolicies, err := policymanager.MergePoliciesOfDifferentHierarchy(gatewayPoliciesByKind, httpRouteNamespacePoliciesByKind)
			if err != nil {
				return err
			}

			mergedPolicies, err = policymanager.MergePoliciesOfDifferentHierarchy(mergedPolicies, httpRoutePoliciesByKind)
			if err != nil {
				return err
			}

			result[gatewayGKNN] = mergedPolicies
		}

		httpRouteNodeMetadata, err := Access(httpRouteNode)
		if err != nil {
			return err
		}
		if httpRouteNodeMetadata == nil {
			httpRouteNodeMetadata = &NodeMetadata{}
			httpRouteNode.Metadata[extensionName] = httpRouteNodeMetadata
		}
		httpRouteNodeMetadata.HTTPRouteEffectivePolicies = result
	}
	return nil
}

// calculateEffectivePoliciesForBackends calculates the effective policies for
// each Backend, considering policies from different hierarchies (GatewayClass,
// Namespace, Gateway, HTTPRoute, and Backend).
func (a *Extension) calculateEffectivePoliciesForBackends(graph *topology.Graph) error {
	for _, backendNode := range graph.Nodes[common.ServiceGK] {
		result := make(map[common.GKNN]map[policymanager.PolicyCrdID]*policymanager.Policy)

		namespaceNode := topologygw.BackendNode(backendNode).Namespace()
		if namespaceNode == nil {
			klog.V(3).InfoS("No Namespace node found for Backend, skipping effective policy calculation", "backend", backendNode.GKNN())
			continue
		}

		backendPoliciesMap, err := directlyattachedpolicy.Access(backendNode)
		if err != nil {
			return err
		}
		namespacePoliciesMap, err := directlyattachedpolicy.Access(namespaceNode)
		if err != nil {
			return err
		}

		// Step 1: Aggregate all policies of the Backend and the Backend-namespace.
		backendPolicies := policymanager.ConvertPoliciesMapToSlice(filterInheritablePolicies(backendPoliciesMap))
		backendNamespacePolicies := policymanager.ConvertPoliciesMapToSlice(filterInheritablePolicies(namespacePoliciesMap))

		// Step 2: Merge Backend and Backend-namespace policies by their kind.
		backendPoliciesByKind, err := policymanager.MergePoliciesOfSimilarKind(backendPolicies)
		if err != nil {
			return err
		}
		backendNamespacePoliciesByKind, err := policymanager.MergePoliciesOfSimilarKind(backendNamespacePolicies)
		if err != nil {
			return err
		}

		// Step 3: Loop through all HTTPRoutes and get their effective policies. Merge
		// effective policies such that we get policies partitioned by Gateway.
		for _, httpRouteNode := range topologygw.BackendNode(backendNode).HTTPRoutes() {
			httpRouteNodeMetadata, err := Access(httpRouteNode) //nolint:govet
			if err != nil {
				return err
			}
			httpRoutePoliciesByGateway := httpRouteNodeMetadata.HTTPRouteEffectivePolicies

			for gatewayID, policies := range httpRoutePoliciesByGateway {
				result[gatewayID], err = policymanager.MergePoliciesOfSameHierarchy(result[gatewayID], policies)
				if err != nil {
					return err
				}
			}
		}

		// Step 4: Loop through all Gateways and merge the Backend and
		// Backend-namespace specific policies. Note that this needs to be done
		// separately from Step 4 i.e. we can't have this loop within Step 4 itself.
		// This is because we first want to merge all policies of the same-hierarchy
		// together and then move to the next hierarchy of Backend and
		// Backend-namespace.
		for gatewayID := range result {
			// Merge all hierarchial policies.
			result[gatewayID], err = policymanager.MergePoliciesOfDifferentHierarchy(result[gatewayID], backendNamespacePoliciesByKind)
			if err != nil {
				return err
			}

			result[gatewayID], err = policymanager.MergePoliciesOfDifferentHierarchy(result[gatewayID], backendPoliciesByKind)
			if err != nil {
				return err
			}
		}

		backendNodeMetadata, err := Access(backendNode)
		if err != nil {
			return err
		}
		if backendNodeMetadata == nil {
			backendNodeMetadata = &NodeMetadata{}
			backendNode.Metadata[extensionName] = backendNodeMetadata
		}
		backendNodeMetadata.BackendEffectivePolicies = result
	}
	return nil
}

type NodeMetadata struct {
	GatewayInheritedPolicies   map[common.GKNN]*policymanager.Policy
	HTTPRouteInheritedPolicies map[common.GKNN]*policymanager.Policy
	BackendInheritedPolicies   map[common.GKNN]*policymanager.Policy

	GatewayEffectivePolicies   map[policymanager.PolicyCrdID]*policymanager.Policy
	HTTPRouteEffectivePolicies map[common.GKNN]map[policymanager.PolicyCrdID]*policymanager.Policy
	BackendEffectivePolicies   map[common.GKNN]map[policymanager.PolicyCrdID]*policymanager.Policy
}

func Access(node *topology.Node) (*NodeMetadata, error) {
	rawData, ok := node.Metadata[extensionName]
	if !ok || rawData == nil {
		klog.V(3).InfoS(fmt.Sprintf("no data found in node for %v", extensionName), "node", node.GKNN())
		return nil, nil
	}
	data, ok := rawData.(*NodeMetadata)
	if !ok {
		return nil, fmt.Errorf("unable to perform type assertion for %v in node %v", extensionName, node.GKNN())
	}
	return data, nil
}
