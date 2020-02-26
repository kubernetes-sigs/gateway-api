/*

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

import (
	core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +kubebuilder:object:root=true
// +kubebuilder:resource:scope=Cluster
// +genclient
// +genclient:nonNamespaced
// +k8s:openapi-gen=true
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// GatewayClass describes a class of Gateways available to the user
// for creating Gateway resources.
//
// GatewayClass is a Cluster level resource.
//
// Support: Core.
type GatewayClass struct {
	metav1.TypeMeta   `json:",inline" protobuf:"bytes,4,opt,name=typeMeta"`
	metav1.ObjectMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`

	// Spec for this GatewayClass.
	Spec GatewayClassSpec `json:"spec,omitempty" protobuf:"bytes,2,opt,name=spec"`
	// Status of the GatewayClass.
	Status GatewayClassStatus `json:"status,omitempty" protobuf:"bytes,3,opt,name=status"`
}

// GatewayClassSpec reflects the configuration of a class of Gateways.
type GatewayClassSpec struct {
	// Controller is a domain/path string that indicates the
	// controller that managing Gateways of this class.
	//
	// Example: "acme.io/gateway-controller".
	//
	// This field is not mutable and cannot be empty.
	//
	// The format of this field is DOMAIN "/" PATH, where DOMAIN
	// and PATH are valid Kubernetes names
	// (https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#names).
	//
	// Support: Core
	//
	// +required
	Controller string `json:"controller" protobuf:"bytes,1,opt,name=controller"`

	// AllowedGatewayNamespaces is a selector of namespaces that Gateways can
	// use this GatewayClass from. This is a standard Kubernetes LabelSelector,
	// a label query over a set of resources. The result of matchLabels and
	// matchExpressions are ANDed. Controllers must not support Gateways in
	// namespaces outside this selector.
	//
	// An empty selector (default) indicates that Gateways can use this
	// GatewayClass from any namespace. This field is intentionally not a
	// pointer because the nil behavior (no namespaces) is undesirable here.
	//
	// Support: Core
	//
	// +optional
	AllowedGatewayNamespaces metav1.LabelSelector `json:"allowedGatewayNamespaces" protobuf:"bytes,2,opt,name=allowedGatewayNamespaces"`

	// AllowedRouteNamespaces is a selector of namespaces that Gateways of this
	// class can reference Routes in. This is a standard Kubernetes
	// LabelSelector, a label query over a set of resources. The result of
	// matchLabels and matchExpressions are ANDed. Controllers must not support
	// Routes in namespaces outside this selector.
	//
	// A nil selector (default) indicates that Gateways of this class can
	// reference Routes within the same namespace. An empty selector indicates
	// that Gateways can reference Routes in any namespace. This field is
	// intentionally a pointer to support the nil behavior (only local Routes
	// allowed).
	//
	// Support: Core
	//
	// +optional
	AllowedRouteNamespaces *metav1.LabelSelector `json:"allowedRouteNamespaces,omitempty" protobuf:"bytes,3,opt,name=allowedRouteNamespaces"`

	// ParametersRef is a controller specific resource containing
	// the configuration parameters corresponding to this
	// class. This is optional if the controller does not require
	// any additional configuration.
	//
	// Valid resources for reference are up to the Controller. Examples
	// include "configmaps" (omit or specify the empty string for the group
	// to indicate the core API group) or a custom resource (CRD).  Omitting
	// or specifying the empty string for both the resource and group
	// indicates that the resource is "configmaps".
	//
	// Support: Custom
	//
	// +optional
	// +protobuf=false
	ParametersRef *GatewayClassParametersObjectReference `json:"parameters,omitempty" protobuf:"bytes,4,opt,name=parametersRef"`
}

// GatewayClassParametersObjectReference identifies a parameters object for a
// gateway class within a known namespace.
//
// +k8s:deepcopy-gen=false
type GatewayClassParametersObjectReference = LocalObjectReference

// GatewayClassConditionType is the type of status conditions.
type GatewayClassConditionType string

const (
	// GatewayClassConditionStatusInvalidParameters indicates the
	// validity of the Parameters set for a given controller. This
	// will initially start off as "Unknown".
	GatewayClassConditionStatusInvalidParameters GatewayClassConditionType = "InvalidParameters"
)

// GatewayClassConditionStatus is the status for a condition.
type GatewayClassConditionStatus = core.ConditionStatus

// GatewayClassStatus is the current status for the GatewayClass.
//
// +kubebuilder:subresource:status
type GatewayClassStatus struct {
	// Conditions is the current status from the controller for
	// this GatewayClass.
	Conditions []GatewayClassCondition `json:"conditions,omitempty" protobuf:"bytes,1,rep,name=conditions"`
}

// GatewayClassCondition contains the details for the current
// condition of this GatewayClass.
//
// Support: Core, unless otherwise specified.
type GatewayClassCondition struct {
	// Type of this condition.
	Type GatewayClassConditionType `json:"type" protobuf:"bytes,1,opt,name=type"`
	// Status of this condition.
	Status GatewayClassConditionStatus `json:"status" protobuf:"bytes,2,opt,name=status"`

	// Reason is a machine consumable string for the last
	// transition. It should be a one-word, CamelCase
	// string. Reason will be defined by the controller.
	//
	// Support: Custom; values will be controller-specific.
	//
	// +optional
	Reason *string `json:"reason,omitempty" protobuf:"bytes,3,opt,name=reason"`
	// Message is a human readable reason for last transition.
	//
	// +optional
	Message *string `json:"message,omitempty" protobuf:"bytes,4,opt,name=message"`
	// LastTransitionTime is the time of the last change to this condition.
	//
	// +optional
	LastTransitionTime *metav1.Time `json:"lastTransitionTime,omitempty" protobuf:"bytes,5,opt,name=lastTransitionTime"`
}

// +kubebuilder:object:root=true

// GatewayClassList contains a list of GatewayClass
type GatewayClassList struct {
	metav1.TypeMeta `json:",inline" protobuf:"bytes,3,opt,name=typeMeta"`
	metav1.ListMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`
	Items           []GatewayClass `json:"items" protobuf:"bytes,2,rep,name=items"`
}

func init() {
	SchemeBuilder.Register(&GatewayClass{}, &GatewayClassList{})
}
