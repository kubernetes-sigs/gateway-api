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

package gateway

import (
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"

	gatewayv1 "sigs.k8s.io/gateway-api/apis/v1"
	"sigs.k8s.io/gateway-api/gwctl/pkg/common"
	"sigs.k8s.io/gateway-api/gwctl/pkg/topology"
)

var (
	AllRelations = []*topology.Relation{
		GatewayParentGatewayClassRelation,
		HTTPRouteParentGatewaysRelation,
		HTTPRouteChildBackendRefsRelation,
		GatewayNamespace,
		HTTPRouteNamespace,
		BackendNamespace,
	}

	// GatewayParentGatewayClassRelation returns GatewayClass for the Gateway.
	GatewayParentGatewayClassRelation = &topology.Relation{
		From: common.GatewayGK,
		To:   common.GatewayClassGK,
		Name: "GatewayClass",
		NeighborFunc: func(u *unstructured.Unstructured) []common.GKNN {
			gateway := &gatewayv1.Gateway{}
			if err := runtime.DefaultUnstructuredConverter.FromUnstructured(u.UnstructuredContent(), gateway); err != nil {
				panic(fmt.Sprintf("failed to convert unstructured Gateway to structured: %v", err))
			}
			return []common.GKNN{{
				Group: common.GatewayClassGK.Group,
				Kind:  common.GatewayClassGK.Kind,
				Name:  string(gateway.Spec.GatewayClassName),
			}}
		},
	}

	// HTTPRouteParentGatewayRelation returns Gateways which the HTTPRoute is
	// attached to.
	HTTPRouteParentGatewaysRelation = &topology.Relation{
		From: common.HTTPRouteGK,
		To:   common.GatewayGK,
		Name: "ParentRef",
		NeighborFunc: func(u *unstructured.Unstructured) []common.GKNN {
			httpRoute := &gatewayv1.HTTPRoute{}
			if err := runtime.DefaultUnstructuredConverter.FromUnstructured(u.UnstructuredContent(), httpRoute); err != nil {
				panic(fmt.Sprintf("failed to convert unstructured HTTPRoute to structured: %v", err))
			}
			result := []common.GKNN{}
			for _, gatewayRef := range httpRoute.Spec.ParentRefs {
				namespace := httpRoute.GetNamespace()
				if namespace == "" {
					namespace = metav1.NamespaceDefault
				}
				if gatewayRef.Namespace != nil {
					namespace = string(*gatewayRef.Namespace)
				}

				result = append(result, common.GKNN{
					Group:     common.GatewayGK.Group,
					Kind:      common.GatewayGK.Kind,
					Namespace: namespace,
					Name:      string(gatewayRef.Name),
				})
			}
			return result
		},
	}

	// HTTPRouteChildBackendRefsRelation returns Backends which the HTTPRoute
	// references.
	HTTPRouteChildBackendRefsRelation = &topology.Relation{
		From: common.HTTPRouteGK,
		To:   common.ServiceGK,
		Name: "BackendRef",
		NeighborFunc: func(u *unstructured.Unstructured) []common.GKNN {
			httpRoute := &gatewayv1.HTTPRoute{}
			if err := runtime.DefaultUnstructuredConverter.FromUnstructured(u.UnstructuredContent(), httpRoute); err != nil {
				panic(fmt.Sprintf("failed to convert unstructured HTTPRoute to structured: %v", err))
			}
			// Aggregate all BackendRefs
			var backendRefs []gatewayv1.BackendObjectReference
			for _, rule := range httpRoute.Spec.Rules {
				for _, backendRef := range rule.BackendRefs {
					backendRefs = append(backendRefs, backendRef.BackendObjectReference)
				}
				for _, filter := range rule.Filters {
					if filter.Type != gatewayv1.HTTPRouteFilterRequestMirror {
						continue
					}
					if filter.RequestMirror == nil {
						continue
					}
					backendRefs = append(backendRefs, filter.RequestMirror.BackendRef)
				}
			}

			// Convert each BackendRef to GKNN. GNKK does not use pointers and
			// thus is easily comparable.
			resultSet := make(map[common.GKNN]bool)
			for _, backendRef := range backendRefs {
				objRef := common.GKNN{
					Name: string(backendRef.Name),
					// Assume namespace is unspecified in the backendRef and
					// check later to override the default value.
					Namespace: httpRoute.GetNamespace(),
				}
				if backendRef.Group != nil {
					objRef.Group = string(*backendRef.Group)
				}
				if backendRef.Kind != nil {
					objRef.Kind = string(*backendRef.Kind)
				} else {
					// Although for resources existing on the server, this value
					// should have received a default before getting persisted.
					// We still explicitly set this for the local analysis when
					// the defaults do not get set automatically.
					objRef.Kind = common.ServiceGK.Kind
				}
				if backendRef.Namespace != nil {
					objRef.Namespace = string(*backendRef.Namespace)
				}
				resultSet[objRef] = true
			}

			// Return unique objRefs
			var result []common.GKNN
			for objRef := range resultSet {
				result = append(result, objRef)
			}
			return result
		},
	}

	// GatewayNamespace returns the Namespace for the Gateway.
	GatewayNamespace = &topology.Relation{
		From: common.GatewayGK,
		To:   common.NamespaceGK,
		Name: "Namespace",
		NeighborFunc: func(u *unstructured.Unstructured) []common.GKNN {
			return []common.GKNN{{
				Group: common.NamespaceGK.Group,
				Kind:  common.NamespaceGK.Kind,
				Name:  u.GetNamespace(),
			}}
		},
	}

	// HTTPRouteNamespace returns the Namespace for the HTTPRoute.
	HTTPRouteNamespace = &topology.Relation{
		From: common.HTTPRouteGK,
		To:   common.NamespaceGK,
		Name: "Namespace",
		NeighborFunc: func(u *unstructured.Unstructured) []common.GKNN {
			return []common.GKNN{{
				Group: common.NamespaceGK.Group,
				Kind:  common.NamespaceGK.Kind,
				Name:  u.GetNamespace(),
			}}
		},
	}

	// BackendNamespace returns the Namespace for the Gateway.
	BackendNamespace = &topology.Relation{
		From: common.ServiceGK,
		To:   common.NamespaceGK,
		Name: "Namespace",
		NeighborFunc: func(u *unstructured.Unstructured) []common.GKNN {
			return []common.GKNN{{
				Group: common.NamespaceGK.Group,
				Kind:  common.NamespaceGK.Kind,
				Name:  u.GetNamespace(),
			}}
		},
	}
)

type gatewayClassNode interface {
	Gateways() map[common.GKNN]*topology.Node
}

type gatewayNodeClassImpl struct {
	node *topology.Node
}

func GatewayClassNode(node *topology.Node) gatewayClassNode { //nolint:revive
	return &gatewayNodeClassImpl{node: node}
}

func (n *gatewayNodeClassImpl) Gateways() map[common.GKNN]*topology.Node {
	return n.node.InNeighbors[GatewayParentGatewayClassRelation]
}

type gatewayNode interface {
	Namespace() *topology.Node
	GatewayClass() *topology.Node
	HTTPRoutes() map[common.GKNN]*topology.Node
}

type gatewayNodeImpl struct {
	node *topology.Node
}

func GatewayNode(node *topology.Node) gatewayNode { //nolint:revive
	return &gatewayNodeImpl{node: node}
}

func (n *gatewayNodeImpl) Namespace() *topology.Node {
	for _, namespaceNode := range n.node.OutNeighbors[GatewayNamespace] {
		return namespaceNode
	}
	return nil
}

func (n *gatewayNodeImpl) GatewayClass() *topology.Node {
	for _, gatewayClassNode := range n.node.OutNeighbors[GatewayParentGatewayClassRelation] {
		return gatewayClassNode
	}
	return nil
}

func (n *gatewayNodeImpl) HTTPRoutes() map[common.GKNN]*topology.Node {
	return n.node.InNeighbors[HTTPRouteParentGatewaysRelation]
}

type httpRouteNode interface {
	Namespace() *topology.Node
	Gateways() map[common.GKNN]*topology.Node
	Backends() map[common.GKNN]*topology.Node
}

type httpRouteNodeImpl struct {
	node *topology.Node
}

func HTTPRouteNode(node *topology.Node) httpRouteNode {
	return &httpRouteNodeImpl{node: node}
}

func (n *httpRouteNodeImpl) Namespace() *topology.Node {
	for _, namespaceNode := range n.node.OutNeighbors[HTTPRouteNamespace] {
		return namespaceNode
	}
	return nil
}

func (n *httpRouteNodeImpl) Gateways() map[common.GKNN]*topology.Node {
	return n.node.OutNeighbors[HTTPRouteParentGatewaysRelation]
}

func (n *httpRouteNodeImpl) Backends() map[common.GKNN]*topology.Node {
	return n.node.OutNeighbors[HTTPRouteChildBackendRefsRelation]
}

type backendNode interface {
	Namespace() *topology.Node
	HTTPRoutes() map[common.GKNN]*topology.Node
}

type backendNodeImpl struct {
	node *topology.Node
}

func BackendNode(node *topology.Node) backendNode {
	return &backendNodeImpl{node: node}
}

func (n *backendNodeImpl) Namespace() *topology.Node {
	for _, namespaceNode := range n.node.OutNeighbors[BackendNamespace] {
		return namespaceNode
	}
	return nil
}

func (n *backendNodeImpl) HTTPRoutes() map[common.GKNN]*topology.Node {
	return n.node.InNeighbors[HTTPRouteChildBackendRefsRelation]
}
