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

// TcpRouteSpec defines the desired state of TcpRoute
type StreamServerSpec struct {
	// Hostnames are the set of domain name that refers to this
	// StreamServer. These names must be unique across the Listener.
	Hostnames []string `json:"hostnames,omitempty"`

	// If this host has multiple names, each name should be present in the
	// server certificate as a DNS SAN.
	//
	// If this server does not have a TLS configuration, or the TLS
	// configuration does not specify any ALPN protocol names, it must
	// be attached to a Dedicated listener.
	TLS *TLSAcceptor

	// Rules are a list of HTTP matchers, filters and actions.
	Rules []StreamRouteRule `json:"rules"`
}

// StreamrouteRule describes how a byte stream is forwarded to its destination.
type StreamRouteRule struct {
}

// StreamServerStatus defines the observed state of TcpRoute
type StreamServerStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

// +kubebuilder:object:root=true

// StreamServer is a virtual server that accepts a stream of bytes and forwards
// it to a subsequent destination.
type StreamServer struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   TcpRouteSpec   `json:"spec,omitempty"`
	Status TcpRouteStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// TcpRouteList contains a list of TcpRoute
type TcpRouteList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []TcpRoute `json:"items"`
}

func init() {
	SchemeBuilder.Register(&TcpRoute{}, &TcpRouteList{})
}
