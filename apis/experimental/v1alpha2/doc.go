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
package experimental

import (
	v1alpha2 "sigs.k8s.io/gateway-api/apis/v1alpha2"
)

// +kubebuilder:object:root=true
type XGRPCRoute v1alpha2.GRPCRoute

// +kubebuilder:object:root=true
type XGRPCRouteList v1alpha2.GRPCRouteList

// +kubebuilder:object:root=true
type XTCPRoute v1alpha2.TCPRoute

// +kubebuilder:object:root=true
type XTCPRouteList v1alpha2.TCPRouteList

// +kubebuilder:object:root=true
type XTLSRoute v1alpha2.TLSRoute

// +kubebuilder:object:root=true
type XTLSRouteList v1alpha2.TLSRouteList

// +kubebuilder:object:root=true
type XUDPRoute v1alpha2.UDPRoute

// +kubebuilder:object:root=true
type XUDPRouteList v1alpha2.UDPRouteList
