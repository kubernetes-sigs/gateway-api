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

// TLSConfig describes a TLS configuration.
//
// References
// - nginx: https://nginx.org/en/docs/http/configuring_https_servers.html
// - envoy: https://www.envoyproxy.io/docs/envoy/latest/api-v2/api/v2/auth/cert.proto
// - haproxy: https://www.haproxy.com/documentation/aloha/9-5/traffic-management/lb-layer7/tls/
// - gcp: https://cloud.google.com/load-balancing/docs/use-ssl-policies#creating_an_ssl_policy_with_a_custom_profile
// - aws: https://docs.aws.amazon.com/elasticloadbalancing/latest/application/create-https-listener.html#describe-ssl-policies
// - azure: https://docs.microsoft.com/en-us/azure/app-service/configure-ssl-bindings#enforce-tls-1112
type TLSConfig struct {
	// Mode defines the TLS behavior for the TLS session initiated by the client.
	// There are two possible modes:
	// - Terminate: The TLS session between the downstream client
	//   and the Gateway is terminated at the Gateway.
	// - Passthrough: The TLS session is NOT terminated by the Gateway. This
	//   implies that the Gateway can't decipher the TLS stream except for
	//   the ClientHello message of the TLS protocol.
	//   CertificateRef field is ignored in this mode.
	Mode TLSModeType `json:"mode,omitempty" protobuf:"bytes,1,opt,name=mode"`

	// CertificateRef is the reference to Kubernetes object that
	// contain a TLS certificate and private key.
	// This certificate MUST be used for TLS handshakes for the domain
	// this TLSConfig is associated with.
	// If an entry in this list omits or specifies the empty
	// string for both the group and the resource, the resource defaults to "secrets".
	// An implementation may support other resources (for example, resource
	// "mycertificates" in group "networking.acme.io").
	// Support: Core (Kubernetes Secrets)
	// Support: Implementation-specific (Other resource types)
	//
	// +optional
	CertificateRef CertificateObjectReference `json:"certificateRef,omitempty"`
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

// CertificateObjectReference identifies a certificate object within a known
// namespace.
//
// +k8s:deepcopy-gen=false
type CertificateObjectReference = SecretsDefaultLocalObjectReference
