/*
Copyright 2020 The Kubernetes Authors.

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

package v1alpha2

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// GatewayAllowType specifies which Gateways should be allowed to use a Route.
type GatewayAllowType string

const (
	// Any Gateway will be able to use this route.
	GatewayAllowAll GatewayAllowType = "All"
	// Only Gateways that have been  specified in GatewayRefs will be able to use this route.
	GatewayAllowFromList GatewayAllowType = "FromList"
	// Only Gateways within the same namespace as the route will be able to use this route.
	GatewayAllowSameNamespace GatewayAllowType = "SameNamespace"
)

// RouteGateways defines which Gateways will be able to use a route. If this
// field results in preventing the selection of a Route by a Gateway, an
// "Admitted" condition with a status of false must be set for the Gateway on
// that Route.
type RouteGateways struct {
	// Allow indicates which Gateways will be allowed to use this route.
	// Possible values are:
	// * All: Gateways in any namespace can use this route.
	// * FromList: Only Gateways specified in GatewayRefs may use this route.
	// * SameNamespace: Only Gateways in the same namespace may use this route.
	//
	// +optional
	// +kubebuilder:validation:Enum=All;FromList;SameNamespace
	// +kubebuilder:default=SameNamespace
	Allow *GatewayAllowType `json:"allow,omitempty"`

	// GatewayRefs must be specified when Allow is set to "FromList". In that
	// case, only Gateways referenced in this list will be allowed to use this
	// route. This field is ignored for other values of "Allow".
	//
	// +optional
	GatewayRefs []GatewayReference `json:"gatewayRefs,omitempty"`
}

// PortNumber defines a network port.
//
// +kubebuilder:validation:Minimum=1
// +kubebuilder:validation:Maximum=65535
type PortNumber int32

// GatewayReference identifies a Gateway in a specified namespace.
type GatewayReference struct {
	// Name is the name of the referent.
	//
	// +kubebuilder:validation:MinLength=1
	// +kubebuilder:validation:MaxLength=253
	Name string `json:"name"`

	// Namespace is the namespace of the referent.
	//
	// +kubebuilder:validation:MinLength=1
	// +kubebuilder:validation:MaxLength=253
	Namespace string `json:"namespace"`
}

// BackendRef defines how a Route should forward a request to a Kubernetes
// resource.
//
// Note that when a namespace is specified, a ReferencePolicy object
// is required in the referent namespace to allow that namespace's
// owner to accept the reference. See the ReferencePolicy object for details.
type BackendRef struct {
	// BackendObjectReference references a Kubernetes object.
	BackendObjectReference `json:",inline"`

	// Weight specifies the proportion of HTTP requests forwarded to the
	// referenced backend. This is computed as
	// weight/(sum of all weights in this ForwardTo list). For non-zero values,
	// there may be some epsilon from the exact proportion defined here
	// depending on the precision an implementation supports. Weight is not a
	// percentage and the sum of weights does not need to equal 100.
	//
	// If only one backend is specified and it has a weight greater than 0, 100%
	// of the traffic is forwarded to that backend. If weight is set to 0, no
	// traffic should be forwarded for this entry. If unspecified, weight
	// defaults to 1.
	//
	// Support for this field varies based on the context where used.
	//
	// +optional
	// +kubebuilder:default=1
	// +kubebuilder:validation:Minimum=0
	// +kubebuilder:validation:Maximum=1000000
	Weight *int32 `json:"weight,omitempty"`
}

// RouteConditionType is a type of condition for a route.
type RouteConditionType string

const (
	// This condition indicates whether the route has been admitted
	// or rejected by a Gateway, and why.
	ConditionRouteAdmitted RouteConditionType = "Admitted"
)

// RouteGatewayStatus describes the status of a route with respect to an
// associated Gateway.
type RouteGatewayStatus struct {
	// GatewayRef is a reference to a Gateway object that is associated with
	// the route.
	GatewayRef RouteStatusGatewayReference `json:"gatewayRef"`

	// Conditions describes the status of the route with respect to the
	// Gateway. The "Admitted" condition must always be specified by controllers
	// to indicate whether the route has been admitted or rejected by the Gateway,
	// and why. Note that the route's availability is also subject to the Gateway's
	// own status conditions and listener status.
	//
	// +listType=map
	// +listMapKey=type
	// +kubebuilder:validation:MinItems=1
	// +kubebuilder:validation:MaxItems=8
	Conditions []metav1.Condition `json:"conditions,omitempty"`
}

// RouteStatusGatewayReference identifies a Gateway in a specified namespace.
// This reference also includes a controller name to simplify cleaning up status
// entries.
type RouteStatusGatewayReference struct {
	// Name is the name of the referent.
	//
	// +kubebuilder:validation:MinLength=1
	// +kubebuilder:validation:MaxLength=253
	Name string `json:"name"`

	// Namespace is the namespace of the referent.
	//
	// +kubebuilder:validation:MinLength=1
	// +kubebuilder:validation:MaxLength=253
	Namespace string `json:"namespace"`

	// Controller is a domain/path string that indicates the controller
	// implementing the Gateway. This corresponds with the controller field on
	// GatewayClass.
	//
	// Example: "acme.io/gateway-controller".
	//
	// The format of this field is DOMAIN "/" PATH, where DOMAIN and PATH are
	// valid Kubernetes names
	// (https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#names).
	//
	// +kubebuilder:validation:MinLength=1
	// +kubebuilder:validation:MaxLength=253
	Controller string `json:"controller"`
}

// RouteStatus defines the observed state that is required across
// all route types.
type RouteStatus struct {
	// Gateways is a list of Gateways that are associated with the route,
	// and the status of the route with respect to each Gateway. When a
	// Gateway selects this route, the controller that manages the Gateway
	// must add an entry to this list when the controller first sees the
	// route and should update the entry as appropriate when the route is
	// modified.
	//
	// A maximum of 100 Gateways will be represented in this list. If this list
	// is full, there may be additional Gateways using this Route that are not
	// included in the list. An empty list means the route has not been admitted
	// by any Gateway.
	//
	// +kubebuilder:validation:MaxItems=100
	Gateways []RouteGatewayStatus `json:"gateways"`
}

// Hostname is the fully qualified domain name of a network host, as defined
// by RFC 3986. Note the following deviations from the "host" part of the
// URI as defined in the RFC:
//
// 1. IP literals are not allowed.
// 2. The `:` delimiter is not respected because ports are not allowed.
//
// Hostname can be "precise" which is a domain name without the terminating
// dot of a network host (e.g. "foo.example.com") or "wildcard", which is a
// domain name prefixed with a single wildcard label (e.g. `*.example.com`).
// The wildcard character `*` must appear by itself as the first DNS label
// and matches only a single label.
//
// Note that as per RFC1035 and RFC1123, a *label* must consist of lower case
// alphanumeric characters or '-', and must start and end with an alphanumeric
// character. No other punctuation is allowed.
//
// +kubebuilder:validation:MinLength=1
// +kubebuilder:validation:MaxLength=253
type Hostname string
