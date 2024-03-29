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

// Package v1beta1 contains API Schema definitions for the
// gateway.networking.k8s.io API group.
//
// +kubebuilder:object:generate=true
// +groupName=gateway.networking.x-k8s.io
package v1beta1

import (
	v1 "sigs.k8s.io/gateway-api/apis/v1"
	v1beta1 "sigs.k8s.io/gateway-api/apis/v1beta1"
)

// +kubebuilder:object:root=true
type XGateway v1.Gateway

// +kubebuilder:object:root=true
type XGatewayList v1.GatewayList

// +kubebuilder:object:root=true
type XGatewayClass v1.GatewayClass

// +kubebuilder:object:root=true
type XGatewayClassList v1.GatewayClassList

// +kubebuilder:object:root=true
type XHTTPRoute v1.HTTPRoute

// +kubebuilder:object:root=true
type XHTTPRouteList v1.HTTPRouteList

// +kubebuilder:object:root=true
type XReferenceGrant v1beta1.ReferenceGrant

// +kubebuilder:object:root=true
type XReferenceGrantList v1beta1.ReferenceGrantList
