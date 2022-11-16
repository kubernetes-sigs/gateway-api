/*
Copyright 2020 The Kubernetes Authors.

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

package v1alpha2

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"sigs.k8s.io/gateway-api/apis/v1beta1"
)

// +genclient
// +genclient:nonNamespaced
// +kubebuilder:object:root=true
// +kubebuilder:resource:categories=gateway-api,scope=Cluster,shortName=gc
// +kubebuilder:subresource:status
// +kubebuilder:deprecatedversion:warning="The v1alpha2 version of GatewayClass has been deprecated and will be removed in a future release of the API. Please upgrade to v1beta1."
// +kubebuilder:printcolumn:name="Controller",type=string,JSONPath=`.spec.controllerName`
// +kubebuilder:printcolumn:name="Accepted",type=string,JSONPath=`.status.conditions[?(@.type=="Accepted")].status`
// +kubebuilder:printcolumn:name="Age",type=date,JSONPath=`.metadata.creationTimestamp`
// +kubebuilder:printcolumn:name="Description",type=string,JSONPath=`.spec.description`,priority=1

// GatewayClass describes a class of Gateways available to the user for creating
// Gateway resources.
//
// It is recommended that this resource be used as a template for Gateways. This
// means that a Gateway is based on the state of the GatewayClass at the time it
// was created and changes to the GatewayClass or associated parameters are not
// propagated down to existing Gateways. This recommendation is intended to
// limit the blast radius of changes to GatewayClass or associated parameters.
// If implementations choose to propagate GatewayClass changes to existing
// Gateways, that MUST be clearly documented by the implementation.
//
// Whenever one or more Gateways are using a GatewayClass, implementations MUST
// add the `gateway-exists-finalizer.gateway.networking.k8s.io` finalizer on the
// associated GatewayClass. This ensures that a GatewayClass associated with a
// Gateway is not deleted while in use.
//
// GatewayClass is a Cluster level resource.
type GatewayClass v1beta1.GatewayClass

const (
	// GatewayClassFinalizerGatewaysExist should be added as a finalizer to the
	// GatewayClass whenever there are provisioned Gateways using a
	// GatewayClass.
	GatewayClassFinalizerGatewaysExist = "gateway-exists-finalizer.gateway.networking.k8s.io"
)

// GatewayClassSpec reflects the configuration of a class of Gateways.
// +k8s:deepcopy-gen=false
type GatewayClassSpec = v1beta1.GatewayClassSpec

// ParametersReference identifies an API object containing controller-specific
// configuration resource within the cluster.
// +k8s:deepcopy-gen=false
type ParametersReference = v1beta1.ParametersReference

// GatewayClassConditionType is the type for status conditions on
// Gateway resources. This type should be used with the
// GatewayClassStatus.Conditions field.
// +k8s:deepcopy-gen=false
type GatewayClassConditionType = v1beta1.GatewayClassConditionType

// GatewayClassConditionReason defines the set of reasons that explain why a
// particular GatewayClass condition type has been raised.
// +k8s:deepcopy-gen=false
type GatewayClassConditionReason = v1beta1.GatewayClassConditionReason

const (
	// This condition indicates whether the GatewayClass has been accepted by
	// the controller requested in the `spec.controller` field.
	//
	// This condition defaults to Unknown, and MUST be set by a controller when
	// it sees a GatewayClass using its controller string. The status of this
	// condition MUST be set to True if the controller will support provisioning
	// Gateways using this class. Otherwise, this status MUST be set to False.
	// If the status is set to False, the controller SHOULD set a Message and
	// Reason as an explanation.
	//
	// Possible reasons for this condition to be true are:
	//
	// * "Accepted"
	//
	// Possible reasons for this condition to be False are:
	//
	// * "InvalidParameters"
	//
	// Possible reasons for this condition to be Unknown are:
	//
	// * "Pending"
	//
	// Controllers should prefer to use the values of GatewayClassConditionReason
	// for the corresponding Reason, where appropriate.
	GatewayClassConditionStatusAccepted GatewayClassConditionType = "Accepted"

	// This reason is used with the "Accepted" condition when the condition is
	// true.
	GatewayClassReasonAccepted GatewayClassConditionReason = "Accepted"

	// This reason is used with the "Accepted" condition when the
	// GatewayClass was not accepted because the parametersRef field
	// was invalid, with more detail in the message.
	GatewayClassReasonInvalidParameters GatewayClassConditionReason = "InvalidParameters"

	// This reason is used with the "Accepted" condition when the
	// requested controller has not yet made a decision about whether
	// to admit the GatewayClass. It is the default Reason on a new
	// GatewayClass.
	GatewayClassReasonPending GatewayClassConditionReason = "Pending"

	// Deprecated: Use "Pending" instead.
	GatewayClassReasonWaiting GatewayClassConditionReason = "Waiting"
)

// GatewayClassStatus is the current status for the GatewayClass.
// +k8s:deepcopy-gen=false
type GatewayClassStatus = v1beta1.GatewayClassStatus

// +kubebuilder:object:root=true

// GatewayClassList contains a list of GatewayClass
type GatewayClassList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []GatewayClass `json:"items"`
}
