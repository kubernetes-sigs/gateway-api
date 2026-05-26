/*
Copyright The Kubernetes Authors.

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

	v1 "sigs.k8s.io/gateway-api/apis/v1"
)

// +genclient
// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:storageversion
// +kubebuilder:resource:categories=gateway-api,shortName=xbackend
// +kubebuilder:printcolumn:name="Age",type=date,JSONPath=`.metadata.creationTimestamp`

// XBackend is a Gateway API resource that represents a backend destination for
// routing traffic. It serves as a Gateway-native way to configure external
// hostname destinations and (in the future) internal service backends.
//
// A Backend of type ExternalHostname provides first-class support for external
// FQDNs, replacing the need for synthetic ExternalName Services.
//
// Support: Extended
type XBackend struct {
	metav1.TypeMeta `json:",inline"`
	// +optional
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// Spec defines the desired state of XBackend.
	// +required
	Spec BackendSpec `json:"spec"`

	// Status defines the current state of XBackend.
	// +optional
	Status BackendStatus `json:"status,omitempty"`
}

// XBackendList contains a list of XBackends.
// +kubebuilder:object:root=true
type XBackendList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []XBackend `json:"items"`
}

// BackendType defines the type of backend destination.
//
// +kubebuilder:validation:Enum=ExternalHostname
type BackendType string

const (
	// BackendTypeExternalHostname indicates that the backend is an external
	// hostname destination. This type provides first-class support for external
	// FQDNs, replacing the need for synthetic ExternalName Services.
	//
	// Support: Extended
	BackendTypeExternalHostname BackendType = "ExternalHostname"
)

// BackendSpec defines the desired state of a Backend.
//
// +kubebuilder:validation:XValidation:rule="self.type == 'ExternalHostname' ? has(self.externalHostname) : !has(self.externalHostname)",message="externalHostname must be set when type is ExternalHostname and must be unset otherwise"
type BackendSpec struct {
	// Type defines the backend type.
	//
	// +unionDiscriminator
	// +required
	Type BackendType `json:"type"`

	// Port defines the port that the implementation should use when connecting
	// to this backend.
	//
	// +required
	Port BackendPort `json:"port"`

	// ExternalHostname specifies the configuration for an ExternalHostname
	// backend. This field must be set when type is ExternalHostname and must
	// be unset otherwise.
	//
	// Support: Extended
	//
	// +optional
	ExternalHostname *ExternalHostnameBackend `json:"externalHostname,omitempty"`

	// Protocol defines the protocol for backend communication.
	//
	// In the common case, the underlying transport protocol for the
	// proxied traffic will already have been determined and processed
	// by the dataplane at the routing step. Where this field is useful
	// is either for higher level protocols or asymmetrical protocol
	// configurations (e.g. version upgrades or h2c). In cases where the
	// protocol is negotiated on the wire (e.g. HTTP/1.1 Upgrade or ALPN),
	// implementations MUST include the protocol set here in the negotiation
	// options presented to the backend.
	//
	// Support: Extended for MCP; Core for TCP, HTTP, HTTP2, H2C, and HTTP11
	//
	// +optional
	Protocol *BackendProtocol `json:"protocol,omitempty"`

	// TLS defines the TLS configuration that the implementation should use
	// when connecting to the backend.
	//
	// ExternalHostname backends SHOULD have TLS configured; the lack of TLS
	// for external hostnames should be considered insecure and a security risk.
	//
	// Support: Extended
	//
	// +optional
	TLS *BackendTLS `json:"tls,omitempty"`
}

// BackendPort describes the port the implementation should use when connecting
// to a Backend. Inspired by discoveryv1.EndpointPort.
type BackendPort struct {
	// Name represents the name of this port. All ports in a Backend must have
	// a unique name. Name must either be an empty string or pass DNS_LABEL
	// validation:
	// * must be no more than 63 characters long.
	// * must consist of lower case alphanumeric characters or '-'.
	// * must start and end with an alphanumeric character.
	//
	// +optional
	// +kubebuilder:validation:MaxLength=63
	// +kubebuilder:validation:Pattern=`^[a-z0-9]([a-z0-9-]*[a-z0-9])?$`
	Name *string `json:"name,omitempty"`

	// Port represents the port number of the endpoint.
	//
	// +kubebuilder:validation:Minimum=1
	// +kubebuilder:validation:Maximum=65535
	//
	// +required
	Port PortNumber `json:"port,omitempty"`
}

// ExternalHostnameBackend specifies the configuration for a backend that
// represents an external hostname destination.
type ExternalHostnameBackend struct {
	// Hostname specifies the FQDN used to reach this backend.
	// IP addresses are not allowed in this field. Implementations that are
	// aware of custom trust domains being used for Service FQDNs MUST also
	// enforce that hostnames ending with those trust domains
	// (e.g. .cluster.local) are not allowed.
	//
	// +kubebuiler:validation:XValidation:rule="!endsWith(self.hostname, '.cluster.local')))",message="hostname must not be an IP address or end with .cluster.local"
	// +required
	Hostname v1.PreciseHostname `json:"hostname,omitempty"`
}

// BackendProtocol defines the protocol used when connecting to a backend.
//
// +kubebuilder:validation:Enum=TCP;HTTP;HTTP2;HTTP11;H2C;MCP
type BackendProtocol string

const (
	// BackendProtocolMCP indicates the Model Context Protocol.
	//
	// Support: Extended
	BackendProtocolMCP BackendProtocol = "MCP"

	// BackendProtocolTCP indicates plain TCP.
	//
	// Support: Extended
	BackendProtocolTCP BackendProtocol = "TCP"

	// BackendProtocolHTTP indicates HTTP (version negotiated via ALPN or
	// implementation default).
	//
	// Support: Core
	BackendProtocolHTTP BackendProtocol = "HTTP"

	// BackendProtocolHTTP2 indicates HTTP/2.
	//
	// Support: Core
	BackendProtocolHTTP2 BackendProtocol = "HTTP2"

	// BackendProtocolHTTP11 indicates HTTP/1.1.
	//
	// Support: Core
	BackendProtocolHTTP11 BackendProtocol = "HTTP11"

	// BackendProtocolH2C indicates HTTP/2 over cleartext (h2c).
	//
	// Support: Core
	BackendProtocolH2C BackendProtocol = "H2C"
)

// BackendTLSMode defines the TLS mode for backend connections.
//
// +kubebuilder:validation:Enum=None;ServerOnly;ClientAndServer
type BackendTLSMode string

const (
	// BackendTLSModeNone disables TLS when connecting to the backend.
	BackendTLSModeNone BackendTLSMode = "None"

	// BackendTLSModeServerOnly enables TLS with server certificate verification.
	BackendTLSModeServerOnly BackendTLSMode = "ServerOnly"

	// BackendTLSModeClientAndServer enables mutual TLS (mTLS).
	BackendTLSModeClientAndServer BackendTLSMode = "ClientAndServer"
)

// BackendTLS defines TLS configuration for connecting to a backend.
//
// +kubebuilder:validation:XValidation:rule="self.mode == 'ClientAndServer' ? has(self.clientCertificateRef) : !has(self.clientCertificateRef)",message="clientCertificateRef must be set if and only if mode is ClientAndServer"
type BackendTLS struct {
	// Mode defines the TLS mode for the backend connection.
	//
	// +required
	Mode BackendTLSMode `json:"mode"`

	// ClientCertificateRef is a reference to a Secret containing the client
	// TLS certificate and private key for mutual TLS. This field is required
	// when mode is ClientAndServer and must be unset otherwise.
	//
	// +optional
	ClientCertificateRef *v1.SecretObjectReference `json:"clientCertificateRef,omitempty"`

	// Validation contains TLS validation configuration for the backend connection.
	// This re-uses the BackendTLSPolicy validation fields for consistency.
	//
	// +optional
	Validation v1.BackendTLSPolicyValidation `json:"validation,omitempty"`
}

// BackendStatus defines the observed state of a Backend.
type BackendStatus struct {
	// Parents is a list of parent resources associated with this Backend,
	// and the status of the Backend with respect to each parent.
	//
	// A controller that manages the Backend must add an entry for each parent
	// it manages and remove the entry when the controller no longer considers
	// the Backend to be associated with that parent.
	//
	// A maximum of 32 parents will be represented in this list. An empty list
	// indicates that the Backend is not associated with any parents.
	//
	// +kubebuilder:validation:MaxItems=32
	// +optional
	// +listType=atomic
	Parents []BackendParentStatus `json:"parents,omitempty"`
}

// BackendParentStatus describes the status of a Backend with respect to a
// specific parent resource (typically a Gateway).
type BackendParentStatus struct {
	// ControllerName is a domain/path string that indicates the name of the
	// controller that manages the Backend.
	//
	// Example: "example.net/gateway-controller".
	//
	// The format of this field is DOMAIN "/" PATH, where DOMAIN and PATH are
	// valid Kubernetes names
	// (https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#names).
	//
	// A controller MUST populate this field when writing status and ensure that
	// entries to status populated with their controller name are removed when
	// they are no longer necessary.
	//
	// +required
	ControllerName v1.GatewayController `json:"controllerName"`

	// ParentRef identifies the parent resource that this status is associated with.
	//
	// +required
	ParentRef v1.ParentReference `json:"parentRef"`

	// Conditions describe the current state of the Backend with respect to
	// this parent.
	//
	// +listType=map
	// +listMapKey=type
	// +optional
	Conditions []metav1.Condition `json:"conditions,omitempty"`
}
