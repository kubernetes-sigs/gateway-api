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
	// The resource may be "services" (omit or use the empty string for the
	// group), or an implementation may support other resources (for
	// example, resource "myroutetargets" in group "networking.acme.io").
	// Omitting or specifying the empty string for both the resource and
	// group indicates that the resource is "services".  If the referent
	// cannot be found, the "InvalidRoutes" status condition on any Gateway
	// that includes the HTTPRoute will be true.
	//
	// Support: Core (Kubernetes Services)
	// Support: Implementation-specific (Other resource types)
	//
	TargetRef ForwardToTargetObjectReference `json:"targetRef"`

	// TargetPort specifies the destination port number to use for the TargetRef.
	// If unspecified and TargetRef is a Service object consisting of a single
	// port definition, that port will be used. If unspecified and TargetRef is
	// a Service object consisting of multiple port definitions, an error is
	// surfaced in status.
	//
	// Support: Core
	//
	// +optional
	TargetPort *TargetPort `json:"targetPort,omitempty"`

	// Weight specifies the proportion of traffic forwarded to a targetRef, computed
	// as weight/(sum of all weights in targetRefs). Weight is not a percentage and
	// the sum of weights does not need to equal 100. The following example (in yaml)
	// sends 70% of traffic to service "my-trafficsplit-sv1" and 30% of the traffic
	// to service "my-trafficsplit-sv2":
	//
	//   forwardTo:
	//     - targetRef:
	//         name: my-trafficsplit-sv1
	//         weight: 70
	//     - targetRef:
	//         name: my-trafficsplit-sv2
	//         weight: 30
	//
	// If only one targetRef is specified, 100% of the traffic is forwarded to the
	// targetRef. If unspecified, weight defaults to 1.
	//
	// Support: Core (HTTPRoute)
	// Support: Extended (TCPRoute)
	//
	// +optional
	// +kubebuilder:default=1
	Weight TargetWeight `json:"weight"`

	// Filters defined at this-level should be executed if and only if
	// the request is being forwarded to the target defined here.
	//
	// Conformance: For any implementation, filtering support, including core
	// filters, is NOT guaranteed at this-level.
	// Use Filters in HTTPRouteRule for portable filters across implementations.
	//
	// Support: custom
	//
	// +optional
	Filters []HTTPRouteFilter `json:"filters"`
}

// TargetPort specifies the destination port number to use for a TargetRef.
type TargetPort int32

// TargetWeight specifies weight used for making a forwarding decision
// to a TargetRef.
type TargetWeight int32

// ForwardToTargetObjectReference identifies a target object of a ForwardTo
// route action within a known namespace.
//
// +k8s:deepcopy-gen=false
type ForwardToTargetObjectReference = ServicesDefaultLocalObjectReference
