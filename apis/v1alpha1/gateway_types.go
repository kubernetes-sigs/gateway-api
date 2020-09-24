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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// Gateway represents an instantiation of a service-traffic handling
// infrastructure by binding Listeners to a set of IP addresses.
type Gateway struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   GatewaySpec   `json:"spec,omitempty"`
	Status GatewayStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// GatewayList contains a list of Gateway
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
type GatewaySpec struct {
	// GatewayClassName used for this Gateway. This is the name of a
	// GatewayClass resource.
	GatewayClassName string `json:"gatewayClassName"`

	// Listeners associated with this Gateway. Listeners define
	// logical endpoints that are bound on this Gateway's addresses.
	// At least one Listener MUST be specified.
	//
	// Each Listener in this array must have a unique Port field,
	// however a GatewayClass may collapse compatible Listener
	// definitions into a single implementation-defined acceptor
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
	// hostname provided in the incoming client request MUST be
	// matched to a Listener to find the correct set of Routes.
	// The incoming hostname MUST be matched using the Hostname
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
	Listeners []Listener `json:"listeners"`

	// Addresses requested for this gateway. This is optional and
	// behavior can depend on the GatewayClass. If a value is set
	// in the spec and the requested address is invalid, the
	// GatewayClass MUST indicate this in the associated entry in
	// GatewayStatus.Addresses.
	//
	// If no Addresses are specified, the GatewayClass may
	// schedule the Gateway in an implementation-defined manner,
	// assigning an appropriate set of Addresses.
	//
	// The GatewayClass MUST bind all Listeners to every
	// GatewayAddress that it assigns to the Gateway.
	//
	// Support: Core
	//
	// +optional
	Addresses []GatewayAddress `json:"addresses"`
}

// Listener embodies the concept of a logical endpoint where a Gateway can
// accept network connections. Each listener in a Gateway must have a unique
// combination of Hostname, Port, and Protocol. This will be enforced by a
// validating webhook.
type Listener struct {
	// Hostname specifies to match the virtual hostname for
	// protocol types that define this concept.
	//
	// Incoming requests that include a hostname are matched
	// according to the given HostnameMatchType to select
	// the Routes from this Listener.
	//
	// If a match type other than "Any" is supplied, it MUST
	// be compatible with the specified Protocol field.
	//
	// Support: Core
	//
	// +optional
	// +kubebuilder:default={match: "Any"}
	Hostname HostnameMatch `json:"hostname,omitempty"`

	// Port is the network port. Multiple listeners may use the
	// same port, subject to the Listener compatibility rules.
	//
	// Support: Core
	//
	// +required
	// +kubebuilder:validation:Minimum=0
	// +kubebuilder:validation:Maximum=65536
	// +kubebuilder:validation:ExclusiveMinimum=true
	// +kubebuilder:validation:ExclusiveMaximum=true
	Port int32 `json:"port,omitempty"`

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
	Protocol ProtocolType `json:"protocol,omitempty"`

	// TLS is the TLS configuration for the Listener. This field
	// is required if the Protocol field is "HTTPS" or "TLS" and
	// ignored otherwise.
	//
	// The association of SNIs to Certificate defined in GatewayTLSConfig is
	// defined based on the Hostname field for this listener:
	// - "Domain": Certificate should be used for the domain and its
	//   first-level subdomains.
	// - "Exact": Certificate should be used for the domain only.
	// - "Any": Certificate in GatewayTLSConfig is the default certificate to use.
	//
	// The GatewayClass MUST use the longest matching SNI out of all
	// available certificates for any TLS handshake.
	//
	// Support: Core
	//
	// +optional
	TLS *GatewayTLSConfig `json:"tls,omitempty"`

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
	// Although a client request may technically match multiple route rules,
	// only one rule may ultimately receive the request. Matching precedence
	// MUST be determined in order of the following criteria:
	//
	// * The most specific match. For example, the most specific HTTPRoute match
	//   is determined by the longest matching combination of hostname and path.
	// * The oldest Route based on creation timestamp. For example, a Route with
	//   a creation timestamp of "2020-09-08 01:02:03" is given precedence over
	//   a Route with a creation timestamp of "2020-09-08 01:02:04".
	// * If everything else is equivalent, the Route appearing first in
	//   alphabetical order (namespace/name) should be given precedence. For
	//   example, foo/bar is given precedence over foo/baz.
	//
	// All valid portions of a Route selected by this field should be supported.
	// Invalid portions of a Route can be ignored (sometimes that will mean the
	// full Route). If a portion of a Route transitions from valid to invalid,
	// support for that portion of the Route should be dropped to ensure
	// consistency. For example, even if a filter specified by a Route is
	// invalid, the rest of the Route should still be supported.
	//
	// Support: Core
	//
	// +required
	Routes RouteBindingSelector `json:"routes"`
}

// HostnameMatch specifies how a Listener should match the incoming
// hostname from a client request. Depending on the incoming protocol,
// the match must apply to names provided by the client at both the
// TLS and the HTTP protocol layers.
type HostnameMatch struct {
	// Match specifies how the hostname provided by the client should be
	// matched against the given value.
	//
	// +optional
	// +kubebuilder:default=Exact
	Match HostnameMatchType `json:"match"`

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
	Name string `json:"name"`
}

// HostnameMatchType specifies the types of matches that are valid
// for hostnames.
// Valid match types are:
//
// * "Domain"
// * "Exact"
// * "Any"
//
// +kubebuilder:validation:Enum=Domain;Exact;Any
type HostnameMatchType string

const (
	// HostnameMatchExact specifies that the hostname provided
	// by the client must exactly match the specified value.
	//
	// This match type MUST be case-insensitive.
	HostnameMatchExact HostnameMatchType = "Exact"

	// HostnameMatchDomain specifies that the hostname provided
	// by the client should be matched against a DNS domain value.
	// The domain match removes the leftmost DNS label from the
	// hostname provided by the client and compares the resulting
	// value.
	//
	// For example, "example.com" is a "Domain" match for the host
	// name "foo.example.com", but not for "foo.bar.example.com"
	// or for "example.foo.com".
	//
	// This match type MUST be case-insensitive.
	HostnameMatchDomain HostnameMatchType = "Domain"

	// HostnameMatchAny specifies that this Listener accepts
	// all client traffic regardless of the presence or value of
	// any hostname supplied by the client.
	HostnameMatchAny HostnameMatchType = "Any"
)

// ProtocolType defines the application protocol accepted by a
// Listener. Implementations are not required to accept all the
// defined protocols. If an implementation does not support a
// specified protocol, it should raise a "ConditionUnsupportedProtocol"
// condition for the affected Listener.
//
// Valid ProtocolType values are:
//
// * "HTTP"
// * "HTTPS"
// * "TLS"
// * "TCP"
// * "UDP"
//
// +kubebuilder:validation:Enum=HTTP;HTTPS;TLS;TCP;UDP
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

	// UDPProtocolType accepts UDP packets.
	UDPProtocolType ProtocolType = "UDP"
)

// TLSRouteOverrideType type defines the level of allowance for Routes
// to override a specific TLS setting.
// +kubebuilder:validation:Enum=Allow;Deny
// +kubebuilder:default=Deny
type TLSRouteOverrideType string

const (
	// TLSROuteOVerrideAllow allows the parameter to be configured from all routes.
	TLSROuteOVerrideAllow TLSRouteOverrideType = "Allow"

	// TLSRouteOverrideDeny prohibits the parameter to be configured from any route.
	TLSRouteOverrideDeny TLSRouteOverrideType = "Deny"
)

// TLSOverridePolicy defines a schema for overriding TLS settings at the Route
// level.
type TLSOverridePolicy struct {
	// Certificate dictates if TLS certificates can be configured
	// via Routes. If set to 'Allow', a TLS certificate for a hostname
	// defined in a Route takes precedence over the certificate defined in
	// Gateway.
	//
	// Support: Core
	//
	// +kubebuilder:default=Deny
	Certificate TLSRouteOverrideType `json:"certificate"`
}

// GatewayTLSConfig describes a TLS configuration.
//
// References
// - nginx: https://nginx.org/en/docs/http/configuring_https_servers.html
// - envoy: https://www.envoyproxy.io/docs/envoy/latest/api-v2/api/v2/auth/cert.proto
// - haproxy: https://www.haproxy.com/documentation/aloha/9-5/traffic-management/lb-layer7/tls/
// - gcp: https://cloud.google.com/load-balancing/docs/use-ssl-policies#creating_an_ssl_policy_with_a_custom_profile
// - aws: https://docs.aws.amazon.com/elasticloadbalancing/latest/application/create-https-listener.html#describe-ssl-policies
// - azure: https://docs.microsoft.com/en-us/azure/app-service/configure-ssl-bindings#enforce-tls-1112
type GatewayTLSConfig struct {
	// Mode defines the TLS behavior for the TLS session initiated by the client.
	// There are two possible modes:
	// - Terminate: The TLS session between the downstream client
	//   and the Gateway is terminated at the Gateway.
	// - Passthrough: The TLS session is NOT terminated by the Gateway. This
	//   implies that the Gateway can't decipher the TLS stream except for
	//   the ClientHello message of the TLS protocol.
	//   CertificateRef field is ignored in this mode.
	Mode TLSModeType `json:"mode,omitempty"`

	// CertificateRef is the reference to Kubernetes object that
	// contain a TLS certificate and private key.
	// This certificate MUST be used for TLS handshakes for the domain
	// this GatewayTLSConfig is associated with.
	// If an entry in this list omits or specifies the empty
	// string for both the group and the resource, the resource defaults to "secrets".
	// An implementation may support other resources (for example, resource
	// "mycertificates" in group "networking.acme.io").
	// Support: Core (Kubernetes Secrets)
	// Support: Implementation-specific (Other resource types)
	//
	// +optional
	CertificateRef CertificateObjectReference `json:"certificateRef,omitempty"`

	// RouteOverride dictates if TLS settings can be configured
	// via Routes or not.
	//
	// CertificateRef must be defined even if `routeOverride.certificate` is
	// set to 'Allow' as it will be used as the default certificate for the
	// listener.
	//
	// +kubebuilder:default={certificate:Deny}
	RouteOverride TLSOverridePolicy `json:"routeOverride,omitempty"`

	// Options are a list of key/value pairs to give extended options
	// to the provider.
	//
	// There variation among providers as to how ciphersuites are
	// expressed. If there is a common subset for expressing ciphers
	// then it will make sense to loft that as a core API
	// construct.
	//
	// Support: Implementation-specific.
	//
	// +optional
	Options map[string]string `json:"options"`
}

// TLSModeType type defines behavior of gateway with TLS protocol.
// +kubebuilder:validation:Enum=Terminate;Passthrough
// +kubebuilder:default=Terminate
type TLSModeType string

const (
	// TLSModeTerminate represents the Terminate mode.
	// In this mode, TLS session between the downstream client
	// and the Gateway is terminated at the Gateway.
	TLSModeTerminate TLSModeType = "Terminate"
	// TLSModePassthrough represents the Passthrough mode.
	// In this mode, the TLS session NOT terminated by the Gateway. This
	// implies that the Gateway can't decipher the TLS stream except for
	// the ClientHello message of the TLS protocol.
	TLSModePassthrough TLSModeType = "Passthrough"
)

// CertificateObjectReference identifies a certificate object within a known
// namespace.
//
// +k8s:deepcopy-gen=false
type CertificateObjectReference = SecretsDefaultLocalObjectReference

// RouteBindingSelector defines a schema for associating routes with the Gateway.
// If NamespaceSelector and RouteSelector are defined, only routes matching both
// selectors are associated with the Gateway.
type RouteBindingSelector struct {
	// RouteNamespaces indicates in which namespaces Routes should be selected
	// for this Gateway. This is restricted to the namespace of this Gateway by
	// default.
	//
	// Support: Core
	//
	// +optional
	// +kubebuilder:default={onlySameNamespace:true}
	RouteNamespaces RouteNamespaces `json:"routeNamespaces,omitempty"`
	// RouteSelector specifies a set of route labels used for selecting
	// routes to associate with the Gateway. If RouteSelector is defined,
	// only routes matching the RouteSelector are associated with the Gateway.
	// An empty RouteSelector matches all routes.
	//
	// Support: Core
	//
	// +optional
	RouteSelector metav1.LabelSelector `json:"routeSelector,omitempty"`
	// Group is the group of the route resource to select. Omitting the value or specifying
	// the empty string indicates the networking.x-k8s.io API group.
	// For example, use the following to select an HTTPRoute:
	//
	// routes:
	//   resource: httproutes
	//
	// Otherwise, if an alternative API group is desired, specify the desired
	// group:
	//
	// routes:
	//   group: acme.io
	//   resource: fooroutes
	//
	// Support: Core
	//
	// +optional
	// +kubebuilder:default=networking.x-k8s.io
	Group string `json:"group"`
	// Resource is the API resource name of the route resource to select.
	//
	// Resource MUST correspond to route resources that are compatible with the
	// application protocol specified in the Listener's Protocol field.
	//
	// If an implementation does not support or recognize this
	// resource type, it SHOULD raise a "ConditionInvalidRoutes"
	// condition for the affected Listener.
	//
	// Support: Core
	//
	// +required
	Resource string `json:"resource"`
}

// GatewayAddress describes an address that can be bound to a Gateway.
type GatewayAddress struct {
	// Type of the Address. This is either "IPAddress" or "NamedAddress".
	//
	// Support: Extended
	//
	// +optional
	// +kubebuilder:default=IPAddress
	Type AddressType `json:"type"`

	// Value. Examples: "1.2.3.4", "128::1", "my-ip-address". Validity of the
	// values will depend on `Type` and support by the controller.
	//
	// +required
	Value string `json:"value"`
}

// AddressType defines how a network address is represented as a text string.
// Valid AddressType values are:
//
// * "IPAddress"
// * "NamedAddress"
//
// +kubebuilder:validation:Enum=IPAddress;NamedAddress
type AddressType string

const (
	// IPAddressType a textual representation of a numeric IP
	// address. IPv4 addresses must be in dotted-decimal
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

// ListenerExtensionObjectReference identifies a listener extension object
// within a known namespace.
//
// +k8s:deepcopy-gen=false
type ListenerExtensionObjectReference = LocalObjectReference

// GatewayStatus defines the observed state of Gateway.
type GatewayStatus struct {
	// Addresses lists the IP addresses that have actually been
	// bound to the Gateway. These addresses may differ from the
	// addresses in the Spec, e.g. if the Gateway automatically
	// assigns an address from a reserved pool.
	//
	// These addresses should all be of type "IPAddress".
	//
	// +required
	Addresses []GatewayAddress `json:"addresses"`

	// Conditions describe the current conditions of the Gateway.
	// +optional
	Conditions []metav1.Condition `json:"conditions,omitempty"`

	// Listeners provide status for each unique listener port defined in the Spec.
	// +optional
	Listeners []ListenerStatus `json:"listeners,omitempty"`
}

// GatewayConditionType is a type of condition associated with a
// Gateway. This type should be used with the GatewayStatus.Conditions
// field.
type GatewayConditionType string

const (
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

	// ConditionInvalidAddress indicates one or more of the
	// Gateway's Addresses is invalid or could not be assigned.
	ConditionInvalidAddress GatewayConditionType = "InvalidAddress"
)

// ListenerStatus is the status associated with a Listener port.
type ListenerStatus struct {
	// Port is the unique Listener port value for which this message
	// is reporting the status. If more than one Gateway Listener
	// shares the same port value, this message reports the combined
	// status of all such Listeners.
	//
	// +required
	Port string `json:"port"`

	// Conditions describe the current condition of this listener.
	//
	// +required
	Conditions []metav1.Condition `json:"conditions"`
}

// ListenerConditionType is a type of condition associated with the
// listener. This type should be used with the ListenerStatus.Conditions
// field.
type ListenerConditionType string

const (
	// ConditionInvalidListener is a generic condition that is a catch all for
	// unsupported configurations that do not match a more specific condition.
	// Implementers should try to use a more specific condition instead of this
	// one to give users and automation more information.
	ConditionInvalidListener ListenerConditionType = "InvalidListener"

	// ConditionListenerNotReady indicates the listener is not ready.
	ConditionListenerNotReady ListenerConditionType = "ListenerNotReady"

	// ConditionPortConflict indicates that two or more Listeners with
	// the same port were bound to this gateway and they could not be
	// collapsed into a single configuration.
	ConditionPortConflict ListenerConditionType = "PortConflict"

	// ConditionInvalidCertificateRef indicates the certificate reference of the
	// listener's TLS configuration is invalid.
	ConditionInvalidCertificateRef ListenerConditionType = "InvalidCertificateRef"

	// ConditionRoutesNotReady indicates that at least one of the specified
	// routes is not ready.
	ConditionRoutesNotReady ListenerConditionType = "RoutesNotReady"

	// ConditionInvalidRoutes indicates that at least one of the specified
	// routes is invalid.
	ConditionInvalidRoutes ListenerConditionType = "InvalidRoutes"

	// ConditionForbiddenRoutesForClass indicates that at least one of the
	// routes is in a namespace forbidden by the GatewayClass.
	ConditionForbiddenRoutesForClass ListenerConditionType = "ForbiddenRoutesForClass"

	// ConditionUnsupportedProtocol indicates that an invalid
	// or unsupported protocol type was requested.
	ConditionUnsupportedProtocol ListenerConditionType = "UnsupportedProtocol"
)
