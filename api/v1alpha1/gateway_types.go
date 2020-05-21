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
// +kubebuilder:subresource:status

// Gateway represents an instantiation of a service-traffic handling infrastructure.
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
	// Listeners associated with this Gateway. Listeners define what addresses,
	// ports, protocols are bound on this Gateway.
	Listeners []Listener `json:"listeners" protobuf:"bytes,2,rep,name=listeners"`
	// Routes specifies a schema for associating routes with the Gateway using
	// selectors. A route is a resource capable of servicing a request and
	// allows a cluster operator to expose a cluster resource (i.e. Service) by
	// externally-reachable URL, load-balance traffic and terminate SSL/TLS.
	// Typically, a route is a "httproute" or "tcproute" in group
	// "networking.x-k8s.io". However, an implementation may support other
	// resources.
	//
	// A nil Routes field (default) indicates that no routes should be
	// associated with this Gateway.
	//
	// Support: Core
	//
	// +optional
	Routes *RouteBindingSelector `json:"routes,omitempty" protobuf:"bytes,3,opt,name=routes"`
}

const (
	// HTTPProtocol constant.
	HTTPProtocol = "HTTP"
	// HTTPSProtocol constant.
	HTTPSProtocol = "HTTPS"
)

// Listener defines a
type Listener struct {
	// Name is the listener's name and should be specified as an
	// RFC 1035 DNS_LABEL [1]:
	//
	// [1] https://tools.ietf.org/html/rfc1035
	//
	// Each listener of a Gateway must have a unique name. Name is used
	// for associating a listener in Gateway status.
	//
	// Support: Core
	//
	// +required
	Name string `json:"name" protobuf:"bytes,1,opt,name=name"`
	// Address requested for this listener. This is optional and behavior
	// can depend on GatewayClass. If a value is set in the spec and
	// the request address is invalid, the GatewayClass MUST indicate
	// this in the associated entry in GatewayStatus.Listeners.
	//
	// Support:
	//
	// +optional
	Address *ListenerAddress `json:"address,omitempty" protobuf:"bytes,2,opt,name=address"`
	// Port is a list of ports associated with the Address.
	//
	// Support:
	// +optional
	Port *int32 `json:"port,omitempty" protobuf:"varint,3,opt,name=port"`
	// Protocol to use.
	//
	// Support:
	// +optional
	Protocol *string `json:"protocol,omitempty" protobuf:"bytes,4,opt,name=protocol"`
	// TLS is the TLS configuration for the Listener. If unspecified,
	// the listener will not support TLS connections.
	//
	// Support: Core
	//
	// +optional
	TLS *TLSConfig `json:"tls,omitempty" protobuf:"bytes,5,opt,name=tls"`
	// ExtensionRef for this Listener.  The resource may be "configmaps" or
	// an implementation-defined resource (for example, resource
	// "mylisteners" in group "networking.acme.io").  Omitting or specifying
	// the empty string for both the resource and group indicates that the
	// resource is "configmaps".  If the referent cannot be found, the
	// listener's "InvalidListener" status condition will be true.
	//
	// Support: custom.
	// +optional
	ExtensionRef *ListenerExtensionObjectReference `json:"extensionRef,omitempty" protobuf:"bytes,6,opt,name=extensionRef"`
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
	// routes to associate with the Gateway.
	//
	// An empty NamespaceSelector (default) indicates that routes from any
	// Namespace can be selected by this Gateway.
	//
	// The OnlyLocalRoutes field takes precedence over this field. This selector
	// will only take effect when OnlyLocalRoutes is false.
	//
	// Support: Core
	//
	// +optional
	NamespaceSelector metav1.LabelSelector `json:"namespaceSelector" protobuf:"bytes,1,opt,name=namespaceSelector"`
	// OnlyLocalRoutes is a boolean used to indicate if routes should only be
	// selected from within the same namespace.
	//
	// When true (default) only routes in the same namespace as the Gateway
	// will be selected.
	//
	// This field takes precedence over the NamespaceSelector field. That
	// selector should only take effect when this field is set to false.
	//
	// Support: Core
	//
	// +kubebuilder:default=true
	// +optional
	OnlyLocalRoutes bool `json:"onlyLocalRoutes" protobuf:"bytes,2,opt,name=onlyLocalRoutes"`
	// RouteSelector specifies a set of route labels used for selecting
	// routes to associate with the Gateway. If RouteSelector is defined,
	// only routes matching the RouteSelector are associated with the Gateway.
	// This selector only applies to namespaces allowed by OnlyLocalRoutes or
	// NamespaceSelector.
	//
	// An empty RouteSelector (default) selects all routes.
	//
	// Support: Core
	//
	// +optional
	RouteSelector metav1.LabelSelector `json:"routeSelector" protobuf:"bytes,3,opt,name=routeSelector"`
}

// ListenerExtensionObjectReference identifies a listener extension object
// within a known namespace.
//
// +k8s:deepcopy-gen=false
type ListenerExtensionObjectReference = ConfigMapsDefaultLocalObjectReference

// GatewayStatus defines the observed state of Gateway.
type GatewayStatus struct {
	// Conditions describe the current conditions of the Gateway.
	Conditions []GatewayCondition `json:"conditions" protobuf:"bytes,1,rep,name=conditions"`
	// Listeners provide status for each listener defined in the Spec. The name
	// in ListenerStatus refers to the corresponding Listener of the same name.
	Listeners []ListenerStatus `json:"listeners" protobuf:"bytes,2,rep,name=listeners"`
}

// GatewayConditionType is a type of condition associated with a Gateway.
type GatewayConditionType string

const (
	// ConditionNoSuchGatewayClass indicates that the specified GatewayClass
	// does not exist.
	ConditionNoSuchGatewayClass GatewayConditionType = "NoSuchGatewayClass"
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
)

// GatewayCondition is an error status for a given route.
type GatewayCondition struct {
	// Type indicates the type of condition.
	// +required
	Type GatewayConditionType `json:"type" protobuf:"bytes,1,opt,name=type,casttype=GatewayConditionType"`
	// Status describes the current state of this condition. Can be "True",
	// "False", or "Unknown".
	// +required
	Status core.ConditionStatus `json:"status" protobuf:"bytes,2,opt,name=status,casttype=k8s.io/api/core/v1.ConditionStatus"`
	// Message is a human-understandable message describing the condition.
	// This field may be empty.
	// +required
	Message string `json:"message,omitempty" protobuf:"bytes,3,opt,name=message"`
	// Reason indicates why the condition is in this state.
	// This field must not be empty.
	// +required
	Reason string `json:"reason,omitempty" protobuf:"bytes,4,opt,name=reason"`
	// LastTransitionTime indicates the last time this condition changed.
	// +required
	LastTransitionTime metav1.Time `json:"lastTransitionTime,omitempty" protobuf:"bytes,5,opt,name=lastTransitionTime"`
}

// ListenerStatus is the status associated with each listener block.
type ListenerStatus struct {
	// Name is the name of the listener this status refers to.
	Name string `json:"name" protobuf:"bytes,1,opt,name=name"`
	// Address bound on this listener.
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
	// ConditionInvalidCertificateRef indicates the certificate reference of the
	// listener's TLS configuration is invalid.
	ConditionInvalidCertificateRef ListenerConditionType = "InvalidCertificateRef"
)

// ListenerCondition is an error status for a given listener.
type ListenerCondition struct {
	// Type indicates the type of condition.
	// +required
	Type ListenerConditionType `json:"type" protobuf:"bytes,1,opt,name=type,casttype=ListenerConditionType"`
	// Status describes the current state of this condition. Can be "True",
	// "False", or "Unknown".
	// +required
	Status core.ConditionStatus `json:"status" protobuf:"bytes,2,opt,name=status,casttype=k8s.io/api/core/v1.ConditionStatus"`
	// Message is a human-understandable message describing the condition.
	// This field may be empty.
	// +required
	Message string `json:"message,omitempty" protobuf:"bytes,3,opt,name=message"`
	// Reason indicates why the condition is in this state.
	// This field must not be empty.
	// +required
	Reason string `json:"reason,omitempty" protobuf:"bytes,4,opt,name=reason"`
	// LastTransitionTime indicates the last time this condition changed.
	// +required
	LastTransitionTime metav1.Time `json:"lastTransitionTime,omitempty" protobuf:"bytes,5,opt,name=lastTransitionTime"`
}

func init() {
	SchemeBuilder.Register(&Gateway{}, &GatewayList{})
}
