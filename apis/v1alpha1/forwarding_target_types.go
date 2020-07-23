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

// ForwardToTarget identifies a target object within a known namespace.
type ForwardToTarget struct {
	// TargetRef is an object reference to forward matched requests to.
	//
	// Support: Core (Kubernetes Services)
	// Support: Implementation-specific (Other resource types)
	//
	TargetRef ForwardToTargetObjectReference `json:"targetRef" protobuf:"bytes,1,opt,name=targetRef"`

	// TargetPort specifies the destination port number to use for the TargetRef.
	// If unspecified and TargetRef is a Service object consisting of a single
	// port definition, that port will be used. If unspecified and TargetRef is
	// a Service object consisting of multiple port definitions, an error is
	// surfaced in status.
	//
	// Support: Core
	//
	// +optional
	TargetPort *TargetPort `json:"targetPort" protobuf:"bytes,2,opt,name=targetPort"`
}

// TargetPort specifies the destination port number to use for a TargetRef.
type TargetPort int32

// ForwardToTargetObjectReference identifies a target object of a ForwardTo
// route action within a known namespace.
//
// +k8s:deepcopy-gen=false
type ForwardToTargetObjectReference = ServicesDefaultLocalObjectReference
