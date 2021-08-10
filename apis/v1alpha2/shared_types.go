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

// ParentRef identifies an API object (usually a Gateway) that can be considered
// a parent of this resource (usually a route). The only kind of parent resource
// with "Core" support is Gateway. This API may be extended in the future to
// support additional kinds of parent resources, such as HTTPRoute.
type ParentRef struct {
	// Group is the group of the referent.
	//
	// Support: Core
	//
	// +kubebuilder:validation:MinLength=1
	// +kubebuilder:validation:MaxLength=253
	// +kubebuilder:default=gateway.networking.k8s.io
	// +optional
	Group *string `json:"group,omitempty"`

	// Kind is kind of the referent.
	//
	// Support: Core (Gateway)
	// Support: Extended (Other Resources)
	//
	// +kubebuilder:validation:MinLength=1
	// +kubebuilder:validation:MaxLength=253
	// +kubebuilder:default=Gateway
	// +optional
	Kind *string `json:"kind,omitempty"`

	// Namespace is the namespace of the referent. When unspecified (empty
	// string), this will either be:
	//
	// * local namespace of the target is a namespace scoped resource
	// * no namespace (not applicable) if the target is cluster-scoped.
	//
	// Support: Extended
	//
	// +kubebuilder:validation:MinLength=1
	// +kubebuilder:validation:MaxLength=253
	// +optional
	Namespace *string `json:"namespace,omitempty"`

	// Scope represents if this refers to a cluster or namespace scoped resource.
	// This may be set to "Cluster" or "Namespace".
	//
	// Support: Core (Namespace)
	// Support: Extended (Cluster)
	//
	// +kubebuilder:validation:Enum=Cluster;Namespace
	// +kubebuilder:default=Namespace
	// +optional
	Scope *string `json:"scope,omitempty"`

	// Name is the name of the referent.
	//
	// Support: Core
	//
	// +kubebuilder:validation:MinLength=1
	// +kubebuilder:validation:MaxLength=253
	Name string `json:"name"`

	// SectionName is the name of a section within the target resource. In the
	// following resources, SectionName is interpreted as the following:
	//
	// * Gateway: Listener Name
	//
	// Implementations MAY choose to support attaching Routes to other resources.
	// If that is the case, they MUST clearly document how SectionName is
	// interpreted.
	//
	// When unspecified (empty string), this will reference the entire resource.
	// For the purpose of status, an attachment is considered successful if at
	// least one section in the parent resource accepts it. For example, Gateway
	// listeners can restrict which Routes can bind to them by Route kind,
	// namespace, or hostname. If 1 of 2 Gateway listeners accept attachment from
	// the referencing Route, the Route MUST be considered successfully
	// attached. If no Gateway listeners accept attachment from this Route, the
	// Route MUST be considered detached from the Gateway.
	//
	// Support: Core
	//
	// +kubebuilder:validation:MinLength=1
	// +kubebuilder:validation:MaxLength=253
	// +optional
	SectionName *string `json:"sectionName,omitempty"`
}

// PortNumber defines a network port.
//
// +kubebuilder:validation:Minimum=1
// +kubebuilder:validation:Maximum=65535
type PortNumber int32

// BackendRef defines how a Route should forward a request to a Kubernetes
// resource.
//
// Note that when a namespace is specified, a ReferencePolicy object
// is required in the referent namespace to allow that namespace's
// owner to accept the reference. See the ReferencePolicy documentation
// for details.
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

// RouteParentStatus describes the status of a route with respect to an
// associated Parent.
type RouteParentStatus struct {
	// ParentRef is a reference to the parent resource that the route wants to
	// be attached to.
	ParentRef ParentRef `json:"parentRef"`

	// Controller is a domain/path string that indicates the controller that
	// wrote this status. This corresponds with the controller field on
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
	Parents []RouteParentStatus `json:"parents"`
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
