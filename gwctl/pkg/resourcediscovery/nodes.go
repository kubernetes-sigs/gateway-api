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

package resourcediscovery

import (
	"fmt"
	"strings"

	gatewayv1 "sigs.k8s.io/gateway-api/apis/v1"
	"sigs.k8s.io/gateway-api/gwctl/pkg/policymanager"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/klog/v2"
)

// resourceID defines a type to represent unique IDs for a resource.
type resourceID struct {
	Group     string
	Kind      string
	Namespace string
	Name      string
}

type (
	gatewayClassID resourceID
	namespaceID    resourceID
	gatewayID      resourceID
	httpRouteID    resourceID
	backendID      resourceID
	policyID       resourceID
)

// GatewayClassID returns an ID for a GatewayClass.
func GatewayClassID(gatewayClassName string) gatewayClassID {
	return gatewayClassID(resourceID{Name: gatewayClassName})
}

// NamespaceID returns an ID for a Namespace.
func NamespaceID(namespaceName string) namespaceID {
	if namespaceName == "" {
		namespaceName = metav1.NamespaceDefault
	}
	return namespaceID(resourceID{Name: namespaceName})
}

// GatewayID returns an ID for a Gateway.
func GatewayID(namespace, name string) gatewayID {
	if namespace == "" {
		namespace = metav1.NamespaceDefault
	}
	return gatewayID(resourceID{Namespace: namespace, Name: name})
}

// HTTPRouteID returns an ID for a HTTPRoute.
func HTTPRouteID(namespace, name string) httpRouteID {
	if namespace == "" {
		namespace = metav1.NamespaceDefault
	}
	return httpRouteID(resourceID{Namespace: namespace, Name: name})
}

// BackendID returns an ID for a Backend.
func BackendID(group, kind, namespace, name string) backendID {
	return backendID(resourceID{
		Group:     strings.ToLower(group),
		Kind:      strings.ToLower(kind),
		Namespace: namespace,
		Name:      name,
	})
}

// BackendIDForService returns an ID for a Backend which contains an underlying
// Service type.
func BackendIDForService(namespace, name string) backendID {
	return BackendID("", "service", namespace, name)
}

// PolicyID returns an ID for a Policy.
func PolicyID(group, kind, namespace, name string) policyID {
	return policyID(resourceID{
		Group:     strings.ToLower(group),
		Kind:      strings.ToLower(kind),
		Namespace: namespace,
		Name:      name,
	})
}

// MarshalText is used to implement encoding.TextMarshaler interface for
// gatewayID.
func (g gatewayID) MarshalText() ([]byte, error) {
	return []byte(fmt.Sprintf("%v/%v", g.Namespace, g.Name)), nil
}

// GatewayClassNode models the relationships and dependencies of a GatewayClass
// resource.
type GatewayClassNode struct {
	// GatewayClass references the actual GatewayClass resource.
	GatewayClass *gatewayv1.GatewayClass

	// Gateways tracks Gateways that are configured to use this GatewayClass.
	Gateways map[gatewayID]*GatewayNode
	// Policies stores Policies that directly apply to this GatewayClass.
	Policies map[policyID]*PolicyNode
}

func NewGatewayClassNode(gatewayClass *gatewayv1.GatewayClass) *GatewayClassNode {
	return &GatewayClassNode{
		GatewayClass: gatewayClass,
		Gateways:     make(map[gatewayID]*GatewayNode),
		Policies:     make(map[policyID]*PolicyNode),
	}
}

func (g *GatewayClassNode) ID() gatewayClassID {
	if g.GatewayClass == nil {
		klog.V(0).ErrorS(nil, "returning empty ID since GatewayClass is nil")
		return gatewayClassID(resourceID{})
	}
	return GatewayClassID(g.GatewayClass.GetName())
}

// GatewayNode models the relationships and dependencies of a Gateway resource.
type GatewayNode struct {
	// Gateway references the actual Gateway resource.
	Gateway *gatewayv1.Gateway

	// Namespace is the namespace of the Gateway.
	Namespace *NamespaceNode
	// GatewayClass tracks the GatewayClass for this Gateway.
	GatewayClass *GatewayClassNode
	// HTTPRoutes stores HTTPRoutes attached to this Gateway.
	HTTPRoutes map[httpRouteID]*HTTPRouteNode
	// Policies stores Policies directly applied to the Gateway.
	Policies map[policyID]*PolicyNode
	// EffectivePolicies reflects the effective policies applicable to this Gateway,
	// considering inheritance and hierarchy.
	EffectivePolicies map[policymanager.PolicyCrdID]policymanager.Policy
}

func NewGatewayNode(gateway *gatewayv1.Gateway) *GatewayNode {
	return &GatewayNode{
		Gateway:           gateway,
		HTTPRoutes:        make(map[httpRouteID]*HTTPRouteNode),
		Policies:          make(map[policyID]*PolicyNode),
		EffectivePolicies: make(map[policymanager.PolicyCrdID]policymanager.Policy),
	}
}

func (g *GatewayNode) ID() gatewayID {
	if g.Gateway == nil {
		klog.V(0).ErrorS(nil, "returning empty ID since Gateway is nil")
		return gatewayID(resourceID{})
	}
	return GatewayID(g.Gateway.GetNamespace(), g.Gateway.GetName())
}

// HTTPRouteNode models the relationships and dependencies of an HTTPRoute
// resource.
type HTTPRouteNode struct {
	// HTTPRoute references the actual HTTPRoute resource.
	HTTPRoute *gatewayv1.HTTPRoute

	// Namespace is the namespace of the HTTPRoute.
	Namespace *NamespaceNode
	// Gateways stores Gateways whhich this HTTPRoute is attached to.
	Gateways map[gatewayID]*GatewayNode
	// Backends lists Backends serving as target endpoints for traffic through
	// this route.
	Backends map[backendID]*BackendNode
	// Policies stores Policies directly applied to the HTTPRoute.
	Policies map[policyID]*PolicyNode
	// EffectivePolicies reflects the effective policies applicable to this
	// HTTPRoute, mapped per Gateway for context-specific enforcement.
	EffectivePolicies map[gatewayID]map[policymanager.PolicyCrdID]policymanager.Policy
}

func NewHTTPRouteNode(httpRoute *gatewayv1.HTTPRoute) *HTTPRouteNode {
	return &HTTPRouteNode{
		HTTPRoute:         httpRoute,
		Gateways:          make(map[gatewayID]*GatewayNode),
		Backends:          make(map[backendID]*BackendNode),
		Policies:          make(map[policyID]*PolicyNode),
		EffectivePolicies: make(map[gatewayID]map[policymanager.PolicyCrdID]policymanager.Policy),
	}
}

func (h *HTTPRouteNode) ID() httpRouteID {
	if h.HTTPRoute == nil {
		klog.V(0).ErrorS(nil, "returning empty ID since HTTPRoute is nil")
		return httpRouteID(resourceID{})
	}
	return HTTPRouteID(h.HTTPRoute.GetNamespace(), h.HTTPRoute.GetName())
}

// BackendNode models the relationships and dependencies of a Backend resource,
// representing the ultimate destination for traffic directed by HTTPRoutes. It
// serves as a generic abstraction, encompassing various underlying resource
// types that can act as traffic targets, such as Services, ServiceImports, etc.
type BackendNode struct {
	// Backend references the actual Backend resource.
	Backend *unstructured.Unstructured

	// Namespace is the namespace of the Backend.
	Namespace *NamespaceNode
	// HTTPRoutes lists HTTPRoutes that reference this Backend as a target.
	HTTPRoutes map[httpRouteID]*HTTPRouteNode
	// Policies stores Policies directly applied to the Backend.
	Policies map[policyID]*PolicyNode
	// EffectivePolicies reflects the effective policies applicable to this
	// Backend, mapped per Gateway for context-specific enforcement.
	EffectivePolicies map[gatewayID]map[policymanager.PolicyCrdID]policymanager.Policy
}

func NewBackendNode(backend *unstructured.Unstructured) *BackendNode {
	return &BackendNode{
		Backend:           backend,
		HTTPRoutes:        make(map[httpRouteID]*HTTPRouteNode),
		Policies:          make(map[policyID]*PolicyNode),
		EffectivePolicies: make(map[gatewayID]map[policymanager.PolicyCrdID]policymanager.Policy),
	}
}

func (b *BackendNode) ID() backendID {
	if b.Backend == nil {
		klog.V(0).ErrorS(nil, "returning empty ID since Backend is empty")
		return backendID(resourceID{})
	}
	return BackendID(
		b.Backend.GroupVersionKind().Group,
		b.Backend.GroupVersionKind().Kind,
		b.Backend.GetNamespace(),
		b.Backend.GetName(),
	)
}

// HTTPRouteNode models the relationships and dependencies of a Namespace.
type NamespaceNode struct {
	// NamespaceName identifies the Namespace.
	NamespaceName string

	// Gateways lists Gateways deployed within the Namespace.
	Gateways map[gatewayID]*GatewayNode
	// HTTPRoutes lists HTTPRoutes configured within the Namespace.
	HTTPRoutes map[httpRouteID]*HTTPRouteNode
	// Backends lists Backends residing within the Namespace.
	Backends map[backendID]*BackendNode
	// Policies stores Policies directly applied to the Namespace.
	Policies map[policyID]*PolicyNode
}

func NewNamespaceNode(namespaceName string) *NamespaceNode {
	if namespaceName == "" {
		namespaceName = metav1.NamespaceDefault
	}
	return &NamespaceNode{
		NamespaceName: namespaceName,
		Gateways:      make(map[gatewayID]*GatewayNode),
		HTTPRoutes:    make(map[httpRouteID]*HTTPRouteNode),
		Backends:      make(map[backendID]*BackendNode),
		Policies:      make(map[policyID]*PolicyNode),
	}
}

func (n *NamespaceNode) ID() namespaceID {
	if n.NamespaceName == "" {
		klog.V(0).ErrorS(nil, "returning empty ID since Namespace is empty")
		return namespaceID(resourceID{})
	}
	return NamespaceID(n.NamespaceName)
}

// PolicyNode models the relationships and dependencies of a Policy resource
type PolicyNode struct {
	// Policy references the actual Policy resource.
	Policy *policymanager.Policy

	// Namespace references the Namespace to which the policy is directly
	// attached. It's nil if the policy is not associated with a specific
	// namespace.
	Namespace *NamespaceNode
	// GatewayClass references the GatewayClassNode to which the policy is
	// directly attached. It's nil if the policy is not associated with a specific
	// GatewayClass.
	GatewayClass *GatewayClassNode
	// Gateway references the GatewayNode to which the policy is directly
	// attached. It's nil if the policy is not associated with a specific Gateway.
	Gateway *GatewayNode
	// HTTPRoute references the HTTPRouteNode to which the policy is directly
	// attached. It's nil if the policy is not associated with a specific
	// HTTPRoute.
	HTTPRoute *HTTPRouteNode
	// Backend references the BackendNode to which the policy is directly
	// attached. It's nil if the policy is not associated with a specific Backend.
	Backend *BackendNode
}

func NewPolicyNode(policy *policymanager.Policy) *PolicyNode {
	return &PolicyNode{
		Policy: policy,
	}
}

func (p *PolicyNode) ID() policyID {
	if p.Policy == nil {
		klog.V(0).ErrorS(nil, "returning empty ID since Policy is empty")
		return policyID(resourceID{})
	}
	return PolicyID(
		p.Policy.Unstructured().GroupVersionKind().Group,
		p.Policy.Unstructured().GetKind(),
		p.Policy.Unstructured().GetNamespace(),
		p.Policy.Unstructured().GetName(),
	)
}
