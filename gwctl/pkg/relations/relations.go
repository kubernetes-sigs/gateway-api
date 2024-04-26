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

// Package relations provides functions for navigating relationships between
// Gateway API resources.
package relations

import (
	gatewayv1 "sigs.k8s.io/gateway-api/apis/v1"
	gatewayv1beta1 "sigs.k8s.io/gateway-api/apis/v1beta1"
	"sigs.k8s.io/gateway-api/gwctl/pkg/common"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	types "k8s.io/apimachinery/pkg/types"
)

// FindGatewayRefsForHTTPRoute returns Gateways which the HTTPRoute is attached
// to.
func FindGatewayRefsForHTTPRoute(httpRoute gatewayv1.HTTPRoute) []types.NamespacedName {
	result := []types.NamespacedName{}
	for _, gatewayRef := range httpRoute.Spec.ParentRefs {
		namespace := httpRoute.GetNamespace()
		if namespace == "" {
			namespace = metav1.NamespaceDefault
		}
		if gatewayRef.Namespace != nil {
			namespace = string(*gatewayRef.Namespace)
		}

		result = append(result, types.NamespacedName{
			Namespace: namespace,
			Name:      string(gatewayRef.Name),
		})
	}
	return result
}

// FindGatewayClassNameForGateway returns GatewayClass for the Gateway.
func FindGatewayClassNameForGateway(gateway gatewayv1.Gateway) string {
	return string(gateway.Spec.GatewayClassName)
}

// FindBackendRefsForHTTPRoute returns Backends which the HTTPRoute references.
func FindBackendRefsForHTTPRoute(httpRoute gatewayv1.HTTPRoute) []common.ObjRef {
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

	// Convert each BackendRef to ObjRef. ObjRef does not use pointers and thus is
	// easily comparable.
	resultSet := make(map[common.ObjRef]bool)
	for _, backendRef := range backendRefs {
		objRef := common.ObjRef{
			Name: string(backendRef.Name),
			// Assume namespace is unspecified in the backendRef and check later to
			// override the default value.
			Namespace: httpRoute.GetNamespace(),
		}
		if backendRef.Group != nil {
			objRef.Group = string(*backendRef.Group)
		}
		if backendRef.Kind != nil {
			objRef.Kind = string(*backendRef.Kind)
		}
		if backendRef.Namespace != nil {
			objRef.Namespace = string(*backendRef.Namespace)
		}
		resultSet[objRef] = true
	}

	// Return unique objRefs
	var result []common.ObjRef
	for objRef := range resultSet {
		result = append(result, objRef)
	}
	return result
}

// ReferenceGrantExposes returns true if the provided reference grant "exposes"
// the given resource. "Exposes" means that the resource is part of the "To"
// fields within the ReferenceGrant.
func ReferenceGrantExposes(referenceGrant gatewayv1beta1.ReferenceGrant, resource common.ObjRef) bool {
	if referenceGrant.GetNamespace() != resource.Namespace {
		return false
	}
	for _, to := range referenceGrant.Spec.To {
		if to.Group != gatewayv1.Group(resource.Group) {
			continue
		}
		if to.Kind != gatewayv1.Kind(resource.Kind) {
			continue
		}
		if to.Name == nil || len(*to.Name) == 0 || *to.Name == gatewayv1.ObjectName(resource.Name) {
			return true
		}
	}
	return false
}

// ReferenceGrantAccepts returns true if the provided reference grant "accepts"
// references from the given resource. "Accepts" means that the resource is part
// of the "From" fields within the ReferenceGrant.
func ReferenceGrantAccepts(referenceGrant gatewayv1beta1.ReferenceGrant, resource common.ObjRef) bool {
	resource.Name = ""
	for _, from := range referenceGrant.Spec.From {
		fromRef := common.ObjRef{
			Group:     string(from.Group),
			Kind:      string(from.Kind),
			Namespace: string(from.Namespace),
		}
		if fromRef == resource {
			return true
		}
	}
	return false
}
