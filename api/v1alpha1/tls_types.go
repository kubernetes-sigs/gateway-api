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
	// Certificates is a list of references to Kubernetes objects that each
	// contain an identity certificate that is bound to the listener.  The
	// host name in a TLS SNI client hello message is used for certificate
	// matching and VirtualHost name selection.  The SNI server_name must
	// match a VirtualHost hostname for the Gateway to route the TLS request.
	// If an entry in this list specifies the empty string for both the group
	// and the resource, the resource defaults to "secret".  An
	// implementation may support other resources (for example, resource
	// "mycertificate" in group "networking.acme.io").
	//
	// Support: Core (Kubernetes Secrets)
	// Support: Implementation-specific (Other resource types)
	//
	// +required
	Certificates []CertificateObjectReference `json:"certificates" protobuf:"bytes,1,rep,name=certificates"`
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
	MinimumVersion *string `json:"minimumVersion,omitempty" protobuf:"bytes,2,opt,name=minimumVersion"`
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
	Options map[string]string `json:"options,omitempty" protobuf:"bytes,3,rep,name=options"`
}

// TLSVersionType identifies supported TLS versions.
//
// +k8s:deepcopy-gen=false
// +protobuf=false
type TLSVersionType string

const (
	// TLSVersion10 denotes the TLS v1.0.
	TLSVersion10 TLSVersionType = "TLS1_0"
	// TLSVersion11 denotes the TLS v1.1.
	TLSVersion11 TLSVersionType = "TLS1_1"
	// TLSVersion12 denotes the TLS v1.2.
	TLSVersion12 TLSVersionType = "TLS1_2"
	// TLSVersion13 denotes the TLS v1.3.
	TLSVersion13 TLSVersionType = "TLS1_3"
)
