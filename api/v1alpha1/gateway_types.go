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

// +genclient
// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// Gateway represents an instantiation of a service-traffic handling
// infrastructure by binding Listeners to a set of IP addresses.
type Gateway struct {
	metav1.TypeMeta   `json:",inline" protobuf:"bytes,1,opt,name=typeMeta"`
	metav1.ObjectMeta `json:"metadata,omitempty" protobuf:"bytes,2,opt,name=metadata"`

	Spec   GatewaySpec   `json:"spec,omitempty" protobuf:"bytes,3,opt,name=spec"`
	Status GatewayStatus `json:"status,omitempty" protobuf:"bytes,4,opt,name=status"`
}

// +kubebuilder:object:root=true

// GatewayList contains a list of Gateway
type GatewayList struct {
	metav1.TypeMeta `json:",inline" protobuf:"bytes,1,opt,name=typeMeta"`
	metav1.ListMeta `json:"metadata,omitempty" protobuf:"bytes,2,opt,name=metadata"`
	Items           []Gateway `json:"items" protobuf:"bytes,3,rep,name=items"`
}

// GatewaySpec defines the desired state of Gateway.
//
// The Spec is split into two major pieces: listeners describing
// client-facing properties and routes that describe application-level
// routing.
//
// Not all possible combinations of options specified in the Spec are
// valid. Some invalid configurations can be caught synchronously via a
// webhook, but there are many cases that will require asynchronous
// signaling via the GatewayStatus block.
type GatewaySpec struct {
	// Class used for this Gateway. This is the name of a GatewayClass resource.
	Class string `json:"class" protobuf:"bytes,1,opt,name=class"`

	// Listeners associated with this Gateway. Listeners define
	// logical endpoints that are bound on this Gateway's addresses.
	// At least one Listener MUST be specified.
	//
	// Each Listener in this array must have a unique Port field,
	// however a GatewayClass may collapse compatible Listener
	// definitions into single implementation-defined acceptor
	// configuration even if their Port fields would otherwise conflict.
	//
	// Listeners are compatible if all of the following conditions are true:
	//
	// 1. all their Protocol fields are "HTTP", or all their Protocol fields are "HTTPS" or TLS"
	// 2. their Hostname fields are specified with a match type other than "Any"
	// 3. their Hostname fields are not an exact match for any other Listener
	//
	// As a special case, each group of compatible listeners
	// may contain exactly one Listener with a match type of "Any".
	//
	// If the GatewayClass collapses compatible Listeners, the
	// host name provided in the incoming client request MUST be
	// matched to a Listener to find the correct set of Routes.
	// The incoming host name MUST be matched using the Hostname
	// field for each Listener in order of most to least specific.
	// That is, "Exact" matches must be processed before "Domain"
	// matches, which must be processed before "Any" matches.
	//
	// If this field specifies multiple Listeners that have the same
	// Port value but are not compatible, the GatewayClass must raise
	// a "PortConflict" condition on the Gateway.
	//
	// Support: Core
	//
	// +required
	// +kubebuilder:validation:MinItems=1
	Listeners []Listener `json:"listeners" protobuf:"bytes,2,rep,name=listeners"`

	// Addresses requested for this gateway. This is optional and
	// behavior can depend on the GatewayClass. If a value is set
	// in the spec and the requested address is invalid, the
	// GatewayClass MUST indicate this in the associated entry in
	// GatewayStatus.Listeners.
	//
	// If no ListenerAddresses are specified, the GatewayClass may
	// schedule the Gateway in an implementation-defined manner,
	// assigning an appropriate set of ListenerAddresses.
	//
	// The GatewayClass MUST bind all Listeners to every
	// ListenerAddress that it assigns to the Gateway.
	//
	// Support: Core
	//
	// +optional
	Addresses []ListenerAddress `json:"addresses" protobuf:"bytes,3,rep,name=addresses"`
}

// ProtocolType defines the application protocol accepted by a Listener.
type ProtocolType string

const (
	// HTTPProtocolType accepts cleartext HTTP/1.1 sessions over TCP.
	HTTPProtocolType ProtocolType = "HTTP"

	// HTTPSProtocolType accepts HTTP/1.1 or HTTP/2 sessions over TLS.
	HTTPSProtocolType ProtocolType = "HTTPS"

	// TLSProtocolType accepts TLS sessions over TCP.
	TLSProtocolType ProtocolType = "TLS"

	// TCPProtocolType accepts TCP sessions.
	TCPProtocolType ProtocolType = "TCP"
)

// HostnameMatchType specifies the types of matches that are valid
// for host names.
type HostnameMatchType string

const (
	// HostnameMatchDomain specifies that the host name provided
	// by the client should be matched against a DNS domain value.
	// The domain match removes the leftmost DNS label from the
	// host name provided by the client and compares the resulting
	// value.
	//
	// For example, "example.com" is a "Domain" match for the host
	// name "foo.example.com", but not for "foo.bar.example.com"
	// or for "example.foo.com".
	//
	// This match type MUST be case-insensitive.
	HostnameMatchDomain HostnameMatchType = "Domain"

	// HostnameMatchExact specifies that the host name provided
	// by the client must exactly match the specified value.
	//
	// This match type MUST be case-insensitive.
	HostnameMatchExact HostnameMatchType = "Exact"

	// HostnameMatchAny specifies that this Listener accepts
	// all client traffic regardless of the presence or value of
	// any host name supplied by the client.
	HostnameMatchAny HostnameMatchType = "Any"
)

// HostnameMatch specifies how a Listener should match the incoming
// host name from a client request. Depending on the incoming protocol,
// the match must apply to names provided by the client at both the
// TLS and the HTTP protocol layers.
type HostnameMatch struct {
	// Match specifies how the host name provided by the client should be
	// matched against the given value.
	//
	// +optional
	// +kubebuilder:validation:Enum=Domain;Exact;Any
	// +kubebuilder:default=Exact
	Match HostnameMatchType `json:"match" protobuf:"bytes,1,name=match"`

	// Name contains the name to match against. This value must
	// be a fully qualified host or domain name conforming to the
	// preferred name syntax defined in
	// [RFC 1034](https://tools.ietf.org/html/rfc1034#section-3.5)
	//
	// In addition to any RFC rules, this field MUST NOT contain
	//
	// 1. IP address literals
	// 2. Colon-delimited port numbers
	// 3. Percent-encoded octets
	//
	// This field is required for the "Domain" and "Exact" match types.
	//
	// +optional
	Name string `json:"name" protobuf:"bytes,2,name=name"`
}

// Listener embodies the concept of a logical endpoint where a
// Gateway can accept network connections.
type Listener struct {
	// Hostname specifies to match the virtual host name for
	// protocol types that define this concept.
	//
	// Incoming requests that include a host name are matched
	// according to the given HostnameMatchType to select
	// the Routes from this Listener.
	//
	// If a match type other than "Any" is supplied, it MUST
	// be compatible with the specified Protocol field.
	//
	// Support: Core
	//
	// +required
	// +kubebuilder:default={match: "Any"}
	Hostname HostnameMatch `json:"hostname,omitempty" protobuf:"bytes,1,opt,name=hostname"`

	// Port is the network port. Multiple listeners may use the
	// same port, subject to the Listener compatibility rules.
	//
	// Support: Core
	//
	// +required
	Port int32 `json:"port,omitempty" protobuf:"varint,2,opt,name=port"`

	// Protocol specifies the network protocol this listener
	// expects to receive. The GatewayClass MUST validate that
	// match type specified in the Hostname field is appropriate
	// for the protocol.
	//
	// * For the "TLS" protocol, the Hostname match MUST be
	//   applied to the [SNI](https://tools.ietf.org/html/rfc6066#section-3)
	//   server name offered by the client.
	// * For the "HTTP" protocol, the Hostname match MUST be
	//   applied to the host portion of the
	//   [effective request URI](https://tools.ietf.org/html/rfc7230#section-5.5)
	//   or the [:authority pseudo-header](https://tools.ietf.org/html/rfc7540#section-8.1.2.3)
	// * For the "HTTPS" protocol, the Hostname match MUST be
	//   applied at both the TLS and HTTP protocol layers.
	//
	// Support: Core
	//
	// +required
	// +kubebuilder:validation:Enum=HTTP;HTTPS;TLS;TCP
	Protocol ProtocolType `json:"protocol,omitempty" protobuf:"bytes,3,opt,name=protocol"`

	// TLS is the TLS configuration for the Listener. This field
	// is required if the Protocol field is "HTTPS" or "TLS".
	//
	// Support: Core
	//
	// +optional
	TLS *TLSConfig `json:"tls,omitempty" protobuf:"bytes,4,opt,name=tls"`

	// Routes specifies a schema for associating routes with the
	// Listener using selectors. A Route is a resource capable of
	// servicing a request and allows a cluster operator to expose
	// a cluster resource (i.e. Service) by externally-reachable
	// URL, load-balance traffic and terminate SSL/TLS.  Typically,
	// a route is a "HTTPRoute" or "TCPRoute" in group
	// "networking.x-k8s.io", however, an implementation may support
	// other types of resources.
	//
	// The Routes selector MUST select a set of objects that
	// are compatible with the application protocol specified in
	// the Protocol field.
	//
	// Support: Core
	//
	// +required
	Routes RouteBindingSelector `json:"routes" protobuf:"bytes,5,opt,name=routes"`
}

// AddressType defines how a network address is represented as a text string.
type AddressType string

const (
	// IPAddressType a textual representation of a numeric IP
	// address. IPv4 addresses, must be in dotted-decimal
	// form. IPv6 addresses must be in a standard IPv6 text
	// representation (see RFC 5952).
	//
	// Implementations should accept any address representation
	// accepted by the inet_pton(3) API.
	//
	// Support: Extended.
	IPAddressType AddressType = "IPAddress"
	// NamedAddressType is an address selected by name. The interpretation of
	// the name is dependent on the controller.
	//
	// Support: Implementation-specific.
	NamedAddressType AddressType = "NamedAddress"
)

// ListenerAddress describes an address for the Listener.
type ListenerAddress struct {
	// Type of the Address. This is one of the *AddressType constants.
	//
	// Support: Extended
	Type AddressType `json:"type" protobuf:"bytes,1,opt,name=type,casttype=AddressType"`
	// Value. Examples: "1.2.3.4", "128::1", "my-ip-address". Validity of the
	// values will depend on `Type` and support by the controller.
	Value string `json:"value" protobuf:"bytes,2,opt,name=value"`
}

// RouteBindingSelector defines a schema for associating routes with the Gateway.
// If NamespaceSelector and RouteSelector are defined, only routes matching both
// selectors are associated with the Gateway.
type RouteBindingSelector struct {
	// NamespaceSelector specifies a set of namespace labels used for selecting
	// routes to associate with the Gateway. If NamespaceSelector is defined,
	// all routes in namespaces matching the NamespaceSelector are associated
	// to the Gateway.
	//
	// An empty NamespaceSelector (default) indicates that routes from any
	// namespace will be associated to this Gateway. This field is intentionally
	// not a pointer because the nil behavior (no namespaces) is undesirable here.
	//
	// Support: Core
	//
	NamespaceSelector metav1.LabelSelector `json:"namespaceSelector" protobuf:"bytes,1,opt,name=namespaceSelector"`
	// RouteSelector specifies a set of route labels used for selecting
	// routes to associate with the Gateway. If RouteSelector is defined,
	// only routes matching the RouteSelector are associated with the Gateway.
	// An empty RouteSelector matches all routes.
	//
	//
	// If undefined, route labels are not used for associating routes to
	// the gateway.
	//
	// Support: Core
	//
	// +optional
	RouteSelector *metav1.LabelSelector `json:"routeSelector,omitempty" protobuf:"bytes,2,opt,name=routeSelector"`
}

// ListenerExtensionObjectReference identifies a listener extension object
// within a known namespace.
//
// +k8s:deepcopy-gen=false
type ListenerExtensionObjectReference = ConfigMapsDefaultLocalObjectReference

// GatewayStatus defines the observed state of Gateway.
type GatewayStatus struct {
	// Conditions describe the current conditions of the Gateway.
	// +optional
	Conditions []GatewayCondition `json:"conditions,omitempty" protobuf:"bytes,1,rep,name=conditions"`
	// Listeners provide status for each listener defined in the Spec. The name
	// in ListenerStatus refers to the corresponding Listener of the same name.
	// +optional
	Listeners []ListenerStatus `json:"listeners,omitempty" protobuf:"bytes,2,rep,name=listeners"`
}

// GatewayConditionType is a type of condition associated with a Gateway.
type GatewayConditionType string

const (
	// ConditionNoSuchGatewayClass indicates that the specified GatewayClass
	// does not exist.
	ConditionNoSuchGatewayClass GatewayConditionType = "NoSuchGatewayClass"
	// ConditionForbiddenNamespaceForClass indicates that this Gateway is in
	// a namespace forbidden by the GatewayClass.
	ConditionForbiddenNamespaceForClass GatewayConditionType = "ForbiddenNamespaceForClass"
	// ConditionGatewayNotScheduled indicates that the Gateway has not been
	// scheduled.
	ConditionGatewayNotScheduled GatewayConditionType = "GatewayNotScheduled"
	// ConditionListenersNotReady indicates that at least one of the specified
	// listeners is not ready. If this condition has a status of True, a more
	// detailed ListenerCondition should be present in the corresponding
	// ListenerStatus.
	ConditionListenersNotReady GatewayConditionType = "ListenersNotReady"
	// ConditionInvalidListeners indicates that at least one of the specified
	// listeners is invalid. If this condition has a status of True, a more
	// detailed ListenerCondition should be present in the corresponding
	// ListenerStatus.
	ConditionInvalidListeners GatewayConditionType = "InvalidListeners"
	// ConditionRoutesNotReady indicates that at least one of the specified
	// routes is not ready.
	ConditionRoutesNotReady GatewayConditionType = "RoutesNotReady"
	// ConditionInvalidRoutes indicates that at least one of the specified
	// routes is invalid.
	ConditionInvalidRoutes GatewayConditionType = "InvalidRoutes"
	// ConditionForbiddenRoutesForClass indicates that at least one of the
	// routes is in a namespace forbidden by the GatewayClass.
	ConditionForbiddenRoutesForClass GatewayConditionType = "ForbiddenRoutesForClass"
)

// GatewayCondition is an error status for a given route.
type GatewayCondition struct {
	// Type indicates the type of condition.
	//
	// +required
	Type GatewayConditionType `json:"type" protobuf:"bytes,1,opt,name=type,casttype=GatewayConditionType"`
	// Status describes the current state of this condition. Can be "True",
	// "False", or "Unknown".
	//
	// +required
	Status core.ConditionStatus `json:"status" protobuf:"bytes,2,opt,name=status,casttype=k8s.io/api/core/v1.ConditionStatus"`
	// Message is a human-understandable message describing the condition.
	// This field may be empty.
	//
	// +required
	Message string `json:"message,omitempty" protobuf:"bytes,3,opt,name=message"`
	// Reason indicates why the condition is in this state.
	// This field must not be empty.
	//
	// +required
	Reason string `json:"reason,omitempty" protobuf:"bytes,4,opt,name=reason"`
	// LastTransitionTime indicates the last time this condition changed.
	// This should be when the underlying condition changed.
	// If that is not known, then using the time when the API field changed is acceptable.
	//
	// +required
	LastTransitionTime metav1.Time `json:"lastTransitionTime,omitempty" protobuf:"bytes,5,opt,name=lastTransitionTime"`
	// If set, this represents the .metadata.generation that the condition was set based upon.
	// For instance, if .metadata.generation is currently 12, but the .status.condition[x].observedGeneration is 9, the condition is out of date
	// with respect to the current state of the instance.
	//
	// +optional
	ObservedGeneration int64 `json:"observedGeneration,omitempty" protobuf:"varint,6,opt,name=observedGeneration"`
}

// ListenerStatus is the status associated with each listener block.
type ListenerStatus struct {
	// Name is the name of the listener this status refers to.
	// TODO(jpeach) Listeners are not indexed by a unique name any more,
	// so this field probably doesn't make sense.
	Name string `json:"name" protobuf:"bytes,1,opt,name=name"`
	// Address bound on this listener.
	// TODO(jpeach) Listeners don't have addresses anymore so this field
	// should move to the GatewayStatus.
	Address *ListenerAddress `json:"address" protobuf:"bytes,2,opt,name=address"`
	// Conditions describe the current condition of this listener.
	Conditions []ListenerCondition `json:"conditions" protobuf:"bytes,3,rep,name=conditions"`
}

// ListenerConditionType is a type of condition associated with the listener.
type ListenerConditionType string

const (
	// ConditionInvalidListener is a generic condition that is a catch all for
	// unsupported configurations that do not match a more specific condition.
	// Implementors should try to use a more specific condition instead of this
	// one to give users and automation more information.
	ConditionInvalidListener ListenerConditionType = "InvalidListener"
	// ConditionListenerNotReady indicates the listener is not ready.
	ConditionListenerNotReady ListenerConditionType = "ListenerNotReady"
	// ConditionInvalidAddress indicates the Address is invalid.
	ConditionInvalidAddress ListenerConditionType = "InvalidAddress"
	// ConditionInvalidName indicates that the name given to the Listener
	// did not meet the requirements of the gateway controller.
	ConditionInvalidName ListenerConditionType = "InvalidName"
	// ConditionNameConflict indicates that two or more Listeners with
	// the same name were bound to this gateway.
	ConditionNameConflict ListenerConditionType = "NameConflict"
	// ConditionPortConflict indicates that two or more Listeners with
	// the same port were bound to this gateway and they could not be
	// collapsed into a single configuration.
	ConditionPortConflict ListenerConditionType = "PortConflict"
	// ConditionInvalidCertificateRef indicates the certificate reference of the
	// listener's TLS configuration is invalid.
	ConditionInvalidCertificateRef ListenerConditionType = "InvalidCertificateRef"
)

// ListenerCondition is an error status for a given listener.
type ListenerCondition struct {
	// Type indicates the type of condition.
	//
	// +required
	Type ListenerConditionType `json:"type" protobuf:"bytes,1,opt,name=type,casttype=ListenerConditionType"`
	// Status describes the current state of this condition. Can be "True",
	// "False", or "Unknown".
	//
	// +required
	Status core.ConditionStatus `json:"status" protobuf:"bytes,2,opt,name=status,casttype=k8s.io/api/core/v1.ConditionStatus"`
	// Message is a human-understandable message describing the condition.
	// This field may be empty.
	//
	// +required
	Message string `json:"message,omitempty" protobuf:"bytes,3,opt,name=message"`
	// Reason indicates why the condition is in this state.
	// This field must not be empty.
	//
	// +required
	Reason string `json:"reason,omitempty" protobuf:"bytes,4,opt,name=reason"`
	// LastTransitionTime indicates the last time this condition changed.
	// This should be when the underlying condition changed.
	// If that is not known, then using the time when the API field changed is acceptable.
	//
	// +required
	LastTransitionTime metav1.Time `json:"lastTransitionTime,omitempty" protobuf:"bytes,5,opt,name=lastTransitionTime"`
	// If set, this represents the .metadata.generation that the condition was set based upon.
	// For instance, if .metadata.generation is currently 12, but the .status.condition[x].observedGeneration is 9, the condition is out of date
	// with respect to the current state of the instance.
	//
	// +optional
	ObservedGeneration int64 `json:"observedGeneration,omitempty" protobuf:"varint,6,opt,name=observedGeneration"`
}

func init() {
	SchemeBuilder.Register(&Gateway{}, &GatewayList{})
}
