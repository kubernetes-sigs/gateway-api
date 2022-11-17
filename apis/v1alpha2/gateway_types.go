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
// +kubebuilder:object:root=true
// +kubebuilder:resource:categories=gateway-api,shortName=gtw
// +kubebuilder:subresource:status
// +kubebuilder:deprecatedversion:warning="The v1alpha2 version of Gateway has been deprecated and will be removed in a future release of the API. Please upgrade to v1beta1."
// +kubebuilder:printcolumn:name="Class",type=string,JSONPath=`.spec.gatewayClassName`
// +kubebuilder:printcolumn:name="Address",type=string,JSONPath=`.status.addresses[*].value`
// +kubebuilder:printcolumn:name="Ready",type=string,JSONPath=`.status.conditions[?(@.type=="Ready")].status`
// +kubebuilder:printcolumn:name="Age",type=date,JSONPath=`.metadata.creationTimestamp`

// Gateway represents an instance of a service-traffic handling infrastructure
// by binding Listeners to a set of IP addresses.
type Gateway v1beta1.Gateway

// +kubebuilder:object:root=true

// GatewayList contains a list of Gateways.
type GatewayList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Gateway `json:"items"`
}

// GatewaySpec defines the desired state of Gateway.
//
// Not all possible combinations of options specified in the Spec are
// valid. Some invalid configurations can be caught synchronously via a
// webhook, but there are many cases that will require asynchronous
// signaling via the GatewayStatus block.
// +k8s:deepcopy-gen=false
type GatewaySpec = v1beta1.GatewaySpec

// Listener embodies the concept of a logical endpoint where a Gateway accepts
// network connections.
// +k8s:deepcopy-gen=false
type Listener = v1beta1.Listener

// ProtocolType defines the application protocol accepted by a Listener.
// Implementations are not required to accept all the defined protocols. If an
// implementation does not support a specified protocol, it MUST set the
// "Accepted" condition to False for the affected Listener with a reason of
// "UnsupportedProtocol".
//
// Core ProtocolType values are listed in the table below.
//
// Implementations can define their own protocols if a core ProtocolType does not
// exist. Such definitions must use prefixed name, such as
// `mycompany.com/my-custom-protocol`. Un-prefixed names are reserved for core
// protocols. Any protocol defined by implementations will fall under
// implementation-specific conformance.
//
// Valid values include:
//
// * "HTTP" - Core support
// * "example.com/bar" - Implementation-specific support
//
// Invalid values include:
//
// * "example.com" - must include path if domain is used
// * "foo.example.com" - must include path if domain is used
//
// +kubebuilder:validation:MinLength=1
// +kubebuilder:validation:MaxLength=255
// +kubebuilder:validation:Pattern=`^[a-zA-Z0-9]([-a-zSA-Z0-9]*[a-zA-Z0-9])?$|[a-z0-9]([-a-z0-9]*[a-z0-9])?(\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*\/[A-Za-z0-9]+$`
// +k8s:deepcopy-gen=false
type ProtocolType = v1beta1.ProtocolType

const (
	// Accepts cleartext HTTP/1.1 sessions over TCP. Implementations MAY also
	// support HTTP/2 over cleartext. If implementations support HTTP/2 over
	// cleartext on "HTTP" listeners, that MUST be clearly documented by the
	// implementation.
	HTTPProtocolType ProtocolType = "HTTP"

	// Accepts HTTP/1.1 or HTTP/2 sessions over TLS.
	HTTPSProtocolType ProtocolType = "HTTPS"

	// Accepts TLS sessions over TCP.
	TLSProtocolType ProtocolType = "TLS"

	// Accepts TCP sessions.
	TCPProtocolType ProtocolType = "TCP"

	// Accepts UDP packets.
	UDPProtocolType ProtocolType = "UDP"
)

// GatewayTLSConfig describes a TLS configuration.
// +k8s:deepcopy-gen=false
type GatewayTLSConfig = v1beta1.GatewayTLSConfig

// TLSModeType type defines how a Gateway handles TLS sessions.
//
// Note that values may be added to this enum, implementations
// must ensure that unknown values will not cause a crash.
//
// Unknown values here must result in the implementation setting the
// Ready Condition for the Listener to `status: False`, with a
// Reason of `Invalid`.
//
// +kubebuilder:validation:Enum=Terminate;Passthrough
// +k8s:deepcopy-gen=false
type TLSModeType = v1beta1.TLSModeType

const (
	// In this mode, TLS session between the downstream client
	// and the Gateway is terminated at the Gateway.
	TLSModeTerminate TLSModeType = "Terminate"

	// In this mode, the TLS session is NOT terminated by the Gateway. This
	// implies that the Gateway can't decipher the TLS stream except for
	// the ClientHello message of the TLS protocol.
	//
	// Note that SSL passthrough is only supported by TLSRoute.
	TLSModePassthrough TLSModeType = "Passthrough"
)

// AllowedRoutes defines which Routes may be attached to this Listener.
// +k8s:deepcopy-gen=false
type AllowedRoutes = v1beta1.AllowedRoutes

// FromNamespaces specifies namespace from which Routes may be attached to a
// Gateway.
//
// Note that values may be added to this enum, implementations
// must ensure that unknown values will not cause a crash.
//
// Unknown values here must result in the implementation setting the
// Ready Condition for the Listener to `status: False`, with a
// Reason of `Invalid`.
//
// +kubebuilder:validation:Enum=All;Selector;Same
// +k8s:deepcopy-gen=false
type FromNamespaces = v1beta1.FromNamespaces

const (
	// Routes in all namespaces may be attached to this Gateway.
	NamespacesFromAll FromNamespaces = "All"
	// Only Routes in namespaces selected by the selector may be attached to
	// this Gateway.
	NamespacesFromSelector FromNamespaces = "Selector"
	// Only Routes in the same namespace as the Gateway may be attached to this
	// Gateway.
	NamespacesFromSame FromNamespaces = "Same"
)

// RouteNamespaces indicate which namespaces Routes should be selected from.
// +k8s:deepcopy-gen=false
type RouteNamespaces = v1beta1.RouteNamespaces

// RouteGroupKind indicates the group and kind of a Route resource.
// +k8s:deepcopy-gen=false
type RouteGroupKind = v1beta1.RouteGroupKind

// GatewayAddress describes an address that can be bound to a Gateway.
// +k8s:deepcopy-gen=false
type GatewayAddress = v1beta1.GatewayAddress

// GatewayStatus defines the observed state of Gateway.
// +k8s:deepcopy-gen=false
type GatewayStatus = v1beta1.GatewayStatus

// GatewayConditionType is a type of condition associated with a
// Gateway. This type should be used with the GatewayStatus.Conditions
// field.
// +k8s:deepcopy-gen=false
type GatewayConditionType = v1beta1.GatewayConditionType

// GatewayConditionReason defines the set of reasons that explain why a
// particular Gateway condition type has been raised.
// +k8s:deepcopy-gen=false
type GatewayConditionReason = v1beta1.GatewayConditionReason

const (
	// This condition indicates whether a Gateway has generated some
	// configuration that will soon be ready in the underlying data plane.
	//
	// It is a positive-polarity summary condition, and so should always be
	// present on the resource with ObservedGeneration set.
	//
	// It should be set to Unknown if the controller performs updates to the
	// status before it has all the information it needs to be able to determine
	// if the condition is true.
	//
	// Possible reasons for this condition to be True are:
	//
	// * "Programmed"
	//
	// Possible reasons for this condition to be False are:
	//
	// * "Invalid"
	// * "Pending"
	//
	// Possible reasons for this condition to be Unknown are:
	//
	// * "Pending"
	//
	// Controllers may raise this condition with other reasons,
	// but should prefer to use the reasons listed above to improve
	// interoperability.
	GatewayConditionProgrammed GatewayConditionType = "Programmed"

	// This reason is used with the "Programmed" condition when the condition is
	// true.
	GatewayReasonProgrammed GatewayConditionReason = "Programmed"

	// This reason is used with the "Programmed" condition when the Listener is
	// syntactically or semantically invalid.
	GatewayReasonInvalid GatewayConditionReason = "Invalid"
)

const (
	// This condition is true when the controller managing the Gateway is
	// syntactically and semantically valid enough to produce some configuration
	// in the underlying data plane, though it has not necessarily configured it yet.
	//
	// Possible reasons for this condition to be True are:
	//
	// * "Accepted"
	//
	// Possible reasons for this condition to be False are:
	//
	// * "NotReconciled"
	// * "NoResources"
	//
	// Possible reasons for this condition to be Unknown are:
	//
	// * "Pending"
	//
	// Controllers may raise this condition with other reasons,
	// but should prefer to use the reasons listed above to improve
	// interoperability.
	GatewayConditionAccepted GatewayConditionType = "Accepted"

	// Deprecated: use "Accepted" instead.
	GatewayConditionScheduled GatewayConditionType = "Scheduled"

	// This reason is used with the "Accepted" condition when the condition is
	// True.
	GatewayReasonAccepted GatewayConditionReason = "Accepted"

	// This reason is used with the "Scheduled" condition when the condition is
	// True.
	//
	// Deprecated: use the "Accepted" condition with reason "Accepted" instead.
	GatewayReasonScheduled GatewayConditionReason = "Scheduled"

	// This reason is used with the "Accepted", "Programmed" and "Ready"
	// conditions when the status is "Unknown" and no controller has reconciled
	// the Gateway.
	GatewayReasonPending GatewayConditionReason = "Pending"

	// Deprecated: Use "Pending" instead.
	GatewayReasonNotReconciled GatewayConditionReason = "NotReconciled"

	// This reason is used with the "Accepted" condition when the
	// Gateway is not scheduled because insufficient infrastructure
	// resources are available.
	GatewayReasonNoResources GatewayConditionReason = "NoResources"
)

const (
	// Ready is an optional Condition that has Extended support. When it's set,
	// the condition indicates whether the Gateway has been completely configured
	// and traffic is ready to flow through the data plane immediately.
	//
	// If both the "ListenersNotValid" and "ListenersNotReady"
	// reasons are true, the Gateway controller should prefer the
	// "ListenersNotValid" reason.
	//
	// Possible reasons for this condition to be true are:
	//
	// * "Ready"
	//
	// Possible reasons for this condition to be False are:
	//
	// * "ListenersNotValid"
	// * "ListenersNotReady"
	// * "AddressNotAssigned"
	//
	// Controllers may raise this condition with other reasons,
	// but should prefer to use the reasons listed above to improve
	// interoperability.
	GatewayConditionReady GatewayConditionType = "Ready"

	// This reason is used with the "Ready" condition when the condition is
	// true.
	GatewayReasonReady GatewayConditionReason = "Ready"

	// This reason is used with the "Ready" condition when one or
	// more Listeners have an invalid or unsupported configuration
	// and cannot be configured on the Gateway.
	GatewayReasonListenersNotValid GatewayConditionReason = "ListenersNotValid"

	// This reason is used with the "Ready" condition when one or
	// more Listeners are not ready to serve traffic.
	GatewayReasonListenersNotReady GatewayConditionReason = "ListenersNotReady"

	// This reason is used with the "Ready" condition when none of the requested
	// addresses have been assigned to the Gateway. This reason can be used to
	// express a range of circumstances, including (but not limited to) IPAM
	// address exhaustion, invalid or unsupported address requests, or a named
	// address not being found.
	GatewayReasonAddressNotAssigned GatewayConditionReason = "AddressNotAssigned"
)

// ListenerStatus is the status associated with a Listener.
// +k8s:deepcopy-gen=false
type ListenerStatus = v1beta1.ListenerStatus

// ListenerConditionType is a type of condition associated with the
// listener. This type should be used with the ListenerStatus.Conditions
// field.
// +k8s:deepcopy-gen=false
type ListenerConditionType = v1beta1.ListenerConditionType

// ListenerConditionReason defines the set of reasons that explain
// why a particular Listener condition type has been raised.
// +k8s:deepcopy-gen=false
type ListenerConditionReason = v1beta1.ListenerConditionReason

const (
	// This condition indicates that the controller was unable to resolve
	// conflicting specification requirements for this Listener. If a
	// Listener is conflicted, its network port should not be configured
	// on any network elements.
	//
	// Possible reasons for this condition to be true are:
	//
	// * "HostnameConflict"
	// * "ProtocolConflict"
	//
	// Possible reasons for this condition to be False are:
	//
	// * "NoConflicts"
	//
	// Controllers may raise this condition with other reasons,
	// but should prefer to use the reasons listed above to improve
	// interoperability.
	ListenerConditionConflicted ListenerConditionType = "Conflicted"

	// This reason is used with the "Conflicted" condition when
	// the Listener conflicts with hostnames in other Listeners. For
	// example, this reason would be used when multiple Listeners on
	// the same port use `example.com` in the hostname field.
	ListenerReasonHostnameConflict ListenerConditionReason = "HostnameConflict"

	// This reason is used with the "Conflicted" condition when
	// multiple Listeners are specified with the same Listener port
	// number, but have conflicting protocol specifications.
	ListenerReasonProtocolConflict ListenerConditionReason = "ProtocolConflict"

	// This reason is used with the "Conflicted" condition when the condition
	// is False.
	ListenerReasonNoConflicts ListenerConditionReason = "NoConflicts"
)

const (
	// This condition indicates that, even though the listener is
	// syntactically and semantically valid, the controller is not able
	// to configure it on the underlying Gateway infrastructure.
	//
	// A Listener is specified as a logical requirement, but needs to be
	// configured on a network endpoint (i.e. address and port) by a
	// controller. The controller may be unable to attach the Listener
	// if it specifies an unsupported requirement, or prerequisite
	// resources are not available.
	//
	// Possible reasons for this condition to be True are:
	//
	// * "Accepted"
	//
	// Possible reasons for this condition to be False are:
	//
	// * "PortUnavailable"
	// * "UnsupportedProtocol"
	// * "UnsupportedAddress"
	//
	// Possible reasons for this condition to be Unknown are:
	//
	// * "Pending"
	//
	// Controllers may raise this condition with other reasons,
	// but should prefer to use the reasons listed above to improve
	// interoperability.
	ListenerConditionAccepted ListenerConditionType = "Accepted"

	// Deprecated: use "Accepted" instead.
	ListenerConditionDetached ListenerConditionType = "Detached"

	// This reason is used with the "Accepted" condition when the condition is
	// True.
	ListenerReasonAccepted ListenerConditionReason = "Accepted"

	// This reason is used with the "Detached" condition when the condition is
	// False.
	//
	// Deprecated: use the "Accepted" condition with reason "Accepted" instead.
	ListenerReasonAttached ListenerConditionReason = "Attached"

	// This reason is used with the "Accepted" condition when the Listener
	// requests a port that cannot be used on the Gateway. This reason could be
	// used in a number of instances, including:
	//
	// * The port is already in use.
	// * The port is not supported by the implementation.
	ListenerReasonPortUnavailable ListenerConditionReason = "PortUnavailable"

	// This reason is used with the "Accepted" condition when the
	// Listener could not be attached to be Gateway because its
	// protocol type is not supported.
	ListenerReasonUnsupportedProtocol ListenerConditionReason = "UnsupportedProtocol"

	// This reason is used with the "Accepted" condition when the Listener could
	// not be attached to the Gateway because the requested address is not
	// supported. This reason could be used in a number of instances, including:
	//
	// * The address is already in use.
	// * The type of address is not supported by the implementation.
	ListenerReasonUnsupportedAddress ListenerConditionReason = "UnsupportedAddress"
)

const (
	// This condition indicates whether the controller was able to
	// resolve all the object references for the Listener.
	//
	// Possible reasons for this condition to be true are:
	//
	// * "ResolvedRefs"
	//
	// Possible reasons for this condition to be False are:
	//
	// * "InvalidCertificateRef"
	// * "InvalidRouteKinds"
	// * "RefNotPermitted"
	//
	// Controllers may raise this condition with other reasons,
	// but should prefer to use the reasons listed above to improve
	// interoperability.
	ListenerConditionResolvedRefs ListenerConditionType = "ResolvedRefs"

	// This reason is used with the "ResolvedRefs" condition when the condition
	// is true.
	ListenerReasonResolvedRefs ListenerConditionReason = "ResolvedRefs"

	// This reason is used with the "ResolvedRefs" condition when the
	// Listener has a TLS configuration with at least one TLS CertificateRef
	// that is invalid or does not exist.
	//
	// A CertificateRef is considered invalid when it refers to a nonexistent
	// or unsupported resource or kind, or when the data within that resource
	// is malformed.
	//
	// This reason must be used only when the reference is allowed, either by
	// referencing an object in the same namespace as the Gateway, or when
	// a cross-namespace reference has been explicitly allowed by a ReferenceGrant.
	// If the reference is not allowed, the reason RefNotPermitted must be used
	// instead.
	ListenerReasonInvalidCertificateRef ListenerConditionReason = "InvalidCertificateRef"

	// This reason is used with the "ResolvedRefs" condition when an invalid or
	// unsupported Route kind is specified by the Listener.
	ListenerReasonInvalidRouteKinds ListenerConditionReason = "InvalidRouteKinds"

	// This reason is used with the "ResolvedRefs" condition when the
	// Listener has a TLS configuration that references an object in another
	// namespace, where the object in the other namespace does not have a
	// ReferenceGrant explicitly allowing the reference.
	ListenerReasonRefNotPermitted ListenerConditionReason = "RefNotPermitted"
)

const (
	// This condition indicates whether a Listener has generated some
	// configuration that will soon be ready in the underlying data plane.
	//
	// It is a positive-polarity summary condition, and so should always be
	// present on the resource with ObservedGeneration set.
	//
	// It should be set to Unknown if the controller performs updates to the
	// status before it has all the information it needs to be able to determine
	// if the condition is true.
	//
	// Possible reasons for this condition to be True are:
	//
	// * "Programmed"
	//
	// Possible reasons for this condition to be False are:
	//
	// * "Invalid"
	// * "Pending"
	//
	// Possible reasons for this condition to be Unknown are:
	//
	// * "Pending"
	//
	// Controllers may raise this condition with other reasons,
	// but should prefer to use the reasons listed above to improve
	// interoperability.
	ListenerConditionProgrammed ListenerConditionType = "Programmed"

	// This reason is used with the "Programmed" condition when the condition is
	// true.
	ListenerReasonProgrammed ListenerConditionReason = "Programmed"
)

const (
	// Ready is an optional Condition that has Extended support. When it's set,
	// the condition indicates whether the Listener has been configured on the
	// Gateway and traffic is ready to flow through the data plane immediately.
	//
	// Possible reasons for this condition to be True are:
	//
	// * "Ready"
	//
	// Possible reasons for this condition to be False are:
	//
	// * "Invalid"
	// * "Pending"
	//
	// Possible reasons for this condition to be Unknown are:
	//
	// * "Pending"
	//
	// Controllers may raise this condition with other reasons,
	// but should prefer to use the reasons listed above to improve
	// interoperability.
	ListenerConditionReady ListenerConditionType = "Ready"

	// This reason is used with the "Ready" condition when the condition is
	// true.
	ListenerReasonReady ListenerConditionReason = "Ready"

	// This reason is used with the "Ready" and "Programmed" conditions when the
	// Listener is syntactically or semantically invalid.
	ListenerReasonInvalid ListenerConditionReason = "Invalid"

	// This reason is used with the "Accepted", "Ready" and "Programmed"
	// conditions when the Listener is either not yet reconciled or not yet not
	// online and ready to accept client traffic.
	ListenerReasonPending ListenerConditionReason = "Pending"
)
