/*
Copyright 2025 The Kubernetes Authors.

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

package v1alpha1

import v1 "sigs.k8s.io/gateway-api/apis/v1"

type (
	// +k8s:deepcopy-gen=false
	AllowedRoutes = v1.AllowedRoutes
	// +k8s:deepcopy-gen=false
	GatewayTLSConfig = v1.GatewayTLSConfig
	// +k8s:deepcopy-gen=false
	Group = v1.Group
	// +k8s:deepcopy-gen=false
	Hostname = v1.Hostname
	// +k8s:deepcopy-gen=false
	Kind = v1.Kind
	// +k8s:deepcopy-gen=false
	ObjectName = v1.ObjectName
	// +k8s:deepcopy-gen=false
	PortNumber = v1.PortNumber
	// +k8s:deepcopy-gen=false
	ProtocolType = v1.ProtocolType
	// +k8s:deepcopy-gen=false
	RouteGroupKind = v1.RouteGroupKind
	// +k8s:deepcopy-gen=false
	SectionName = v1.SectionName
)

// ParentGatewayReference identifies an API object including its namespace,
// defaulting to Gateway.
type ParentGatewayReference struct {
	// Group is the group of the referent.
	//
	// +optional
	// +kubebuilder:default="gateway.networking.k8s.io"
	Group *Group `json:"group"`

	// Kind is kind of the referent. For example "Gateway".
	//
	// +optional
	// +kubebuilder:default=Gateway
	Kind *Kind `json:"kind"`

	// Name is the name of the referent.
	Name ObjectName `json:"name"`
}
