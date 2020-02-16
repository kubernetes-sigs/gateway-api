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

// TLSConfiguration describes the TLS configuration for a party in a TLS
// session.
//
// References
// - nginx: https://nginx.org/en/docs/http/configuring_https_servers.html
// - envoy: https://www.envoyproxy.io/docs/envoy/latest/api-v2/api/v2/auth/cert.proto
// - haproxy: https://www.haproxy.com/documentation/aloha/9-5/traffic-management/lb-layer7/tls/
// - gcp: https://cloud.google.com/load-balancing/docs/use-ssl-policies#creating_an_ssl_policy_with_a_custom_profile
// - aws: https://docs.aws.amazon.com/elasticloadbalancing/latest/application/create-https-listener.html#describe-ssl-policies
// - azure: https://docs.microsoft.com/en-us/azure/app-service/configure-ssl-bindings#enforce-tls-1112
type TLSConfiguration struct {
	// ALPNProtocols is the list of IANA-registered ALPN protocol names
	// that this TLS party is willing to accept.
	//
	// https://www.iana.org/assignments/tls-extensiontype-values/tls-extensiontype-values.xhtml#alpn-protocol-ids
	//
	// +optional
	ALPNProtocols []string `json:"alpnProtocols,omitempty"`
	// Certificates is a list of references to Kubernetes objects that each
	// contain an identity certificate that is bound to the listener.  The
	// host name in a TLS SNI client hello message is used for certificate
	// matching and route host name selection.  The SNI server_name must
	// match a route host name for the Gateway to route the TLS request.  If
	// an entry in this list specifies the empty string for both the group
	// and the resource, the resource defaults to "secret".  An
	// implementation may support other resources (for example, resource
	// "mycertificate" in group "networking.acme.io").
	//
	// Support: Core (Kubernetes Secrets)
	// Support: Implementation-specific (Other resource types)
	//
	// +required
	Certificates []CertificateObjectReference `json:"certificates,omitempty"`
	// MinimumVersion of TLS allowed. It is recommended to use one of
	// the TLS_* constants above. Note: this is not strongly
	// typed to allow implementation-specific versions to be used without
	// requiring updates to the API types. String must be of the form
	// "<protocol><major>_<minor>".
	//
	// Support: Core for TLS1_{1,2,3}. Implementation-specific for all other
	// values.
	//
	// +optional
	MinimumVersion *string `json:"minimumVersion"`
	// Options are a list of key/value pairs to give extended options
	// to the provider.
	//
	// There variation among providers as to how ciphersuites are
	// expressed. If there is a common subset for expressing ciphers
	// then it will make sense to loft that as a core API
	// construct.
	//
	// Support: Implementation-specific.
	Options map[string]string `json:"options"`
}

// TLSPeerValidation policy specifies the requirements that a party in a TLS
// session has on communicating with its peer.
type TLSPeerValidationPolicy struct {
	// TODO(jpeach): obviously!
}

// TLSAcceptor describes the TLS configuration for accepting TLS sessions.
type TLSAcceptor struct {
	TLSConfiguration `json:",inline"`
	Peer             TLSPeerValidationPolicy `json:"peer"`
}

// TLSInitiator describes the TLS configuration for a party that initiates TLS
// sessions.
//
// TODO(jpeach): consider merging TLSAcceptor and TLSInitiator if it turns out
// that there are no differences between the information we need to capture for
// each usage.
type TLSInitiator struct {
	TLSConfiguration `json:",inline"`
	Peer             TLSPeerValidationPolicy `json:"peer"`
}
