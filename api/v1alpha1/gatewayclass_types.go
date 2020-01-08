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
// +kubebuilder:resource:scope=Cluster
// +genclient
// +genclient:nonNamespaced
// +k8s:openapi-gen=true
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// GatewayClass describes a class of gateways av
// for defining access to their routed services. GatewayClass allow a
//
// GatewayClass is a non-namespaced resource.
type GatewayClass struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// Controller is a domain/path that denotes which controller is responsible
	// for this class. Example: "acme.io/gateway-controller".
	Controller string `json:"controller" protobuf:"bytes,2,opt,name=controller"`
	// Parameters is a controller specific resource containing the
	// configuration parameters corresponding to this class. This is optional
	// if the controllers does not require any additional configuration.
	// +optional
	Parameters *core.TypedLocalObjectReference `json:"parameters,omitempty" protobuf:"bytes,3,opt,name=parameters"`
}

// +kubebuilder:object:root=true

// GatewayClassList contains a list of GatewayClass
type GatewayClassList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []GatewayClass `json:"items"`
}

func init() {
	SchemeBuilder.Register(&GatewayClass{}, &GatewayClassList{})
}
