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
	// +kubebuilder:default=gateway.networking.k8s.io
	// +optional
	Group *Group `json:"group,omitempty"`

	// Kind is kind of the referent.
	//
	// Support: Core (Gateway)
	// Support: Custom (Other Resources)
	//
	// +kubebuilder:default=Gateway
	// +optional
	Kind *Kind `json:"kind,omitempty"`

	// Namespace is the namespace of the referent. When unspecified (or empty
	// string), this will either be:
	//
	// * local namespace of the route when scope is set to Namespace.
	// * no namespace when scope is set to Cluster.
	//
	// Support: Core
	//
	// +optional
	Namespace *Namespace `json:"namespace,omitempty"`

	// Scope represents if this refers to a cluster or namespace scoped
	// resource. This may be set to "Cluster" or "Namespace".
	//
	// Support: Core (Namespace)
	// Support: Custom (Cluster)
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
	// listeners can restrict which Routes can attach to them by Route kind,
	// namespace, or hostname. If 1 of 2 Gateway listeners accept attachment from
	// the referencing Route, the Route MUST be considered successfully
	// attached. If no Gateway listeners accept attachment from this Route, the
	// Route MUST be considered detached from the Gateway.
	//
	// Support: Core
	//
	// +optional
	SectionName SectionName `json:"sectionName,omitempty"`
}

// CommonRouteSpec defines the common attributes that all Routes should include
// within their spec.
type CommonRouteSpec struct {
	// ParentRefs references the resources (usually Gateways) that a Route wants
	// to be attached to. Note that the referenced parent resource needs to
	// allow this for the attachment to be complete. For Gateways, that means
	// the Gateway needs to allow attachment from Routes of this kind and
	// namespace.
	//
	// The only kind of parent resource with "Core" support is Gateway. This API
	// may be extended in the future to support additional kinds of parent
	// resources such as one of the route kinds.
	//
	// It is invalid to reference an identical parent more than once. It is
	// valid to reference multiple distinct sections within the same parent
	// resource, such as 2 Listeners within a Gateway.
	//
	// It is possible to separately reference multiple distinct objects that may
	// be collapsed by an implementation. For example, some implementations may
	// choose to merge compatible Gateway Listeners together. If that is the
	// case, the list of routes attached to those resources should also be
	// merged.
	//
	// +optional
	// +kubebuilder:validation:MaxItems=32
	ParentRefs []ParentRef `json:"parentRefs,omitempty"`
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
	// ParentRef corresponds with a ParentRef in the spec that this
	// RouteParentStatus struct describes the status of.
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
	Controller GatewayController `json:"controller"`

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
	// Parents is a list of parent resources (usually Gateways) that are
	// associated with the route, and the status of the route with respect to
	// each parent. When this route attaches to a parent, the controller that
	// manages the parent must add an entry to this list when the controller
	// first sees the route and should update the entry as appropriate when the
	// route or gateway is modified.
	//
	// A maximum of 32 Gateways will be represented in this list. An empty list
	// means the route has not been attached to any Gateway.
	//
	// +kubebuilder:validation:MaxItems=32
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
// +kubebuilder:validation:Pattern=`^(\*\.)?[a-z0-9]([-a-z0-9]*[a-z0-9])?(\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*$`
type Hostname string

// Group refers to a Kubernetes Group. It must either be an empty string or a
// RFC 1123 subdomain.
//
// This validation is based off of the corresponding Kubernetes validation:
// https://github.com/kubernetes/apimachinery/blob/02cfb53916346d085a6c6c7c66f882e3c6b0eca6/pkg/util/validation/validation.go#L208
//
// Valid values include:
//
// * "" - empty string implies core Kubernetes API group
// * "networking.k8s.io"
// * "foo.example.com"
//
// Invalid values include:
//
// * "example.com/bar" - "/" is an invalid character
//
// +kubebuilder:validation:MaxLength=253
// +kubebuilder:validation:Pattern=`^$|^[a-z0-9]([-a-z0-9]*[a-z0-9])?(\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*$`
type Group string

// Kind refers to a Kubernetes Kind.
//
// Valid values include:
//
// * "Service"
// * "HTTPRoute"
//
// Invalid values include:
//
// * "invalid/kind" - "/" is an invalid character
//
// +kubebuilder:validation:MinLength=1
// +kubebuilder:validation:MaxLength=63
// +kubebuilder:validation:Pattern=`^[a-zA-Z]([-a-zA-Z0-9]*[a-zA-Z0-9])?$`
type Kind string

// Namespace refers to a Kubernetes namespace. It must be a RFC 1123 label.
//
// This validation is based off of the corresponding Kubernetes validation:
// https://github.com/kubernetes/apimachinery/blob/02cfb53916346d085a6c6c7c66f882e3c6b0eca6/pkg/util/validation/validation.go#L187
//
// This is used for Namespace name validation here:
// https://github.com/kubernetes/apimachinery/blob/02cfb53916346d085a6c6c7c66f882e3c6b0eca6/pkg/api/validation/generic.go#L63
//
// Valid values include:
//
// * "example"
//
// Invalid values include:
//
// * "example.com" - "." is an invalid character
//
// +kubebuilder:validation:Pattern=`^[a-z0-9]([-a-z0-9]*[a-z0-9])?$`
// +kubebuilder:validation:MinLength=1
// +kubebuilder:validation:MaxLength=63
type Namespace string

// SectionName is the name of a section in a Kubernetes resource. It must be a
// RFC 1123 subdomain.
//
// This validation is based off of the corresponding Kubernetes validation:
// https://github.com/kubernetes/apimachinery/blob/02cfb53916346d085a6c6c7c66f882e3c6b0eca6/pkg/util/validation/validation.go#L208
//
// Valid values include:
//
// * "example.com"
// * "foo.example.com"
//
// Invalid values include:
//
// * "example.com/bar" - "/" is an invalid character
//
// +kubebuilder:validation:Pattern=`^[a-z0-9]([-a-z0-9]*[a-z0-9])?(\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*$`
// +kubebuilder:validation:MinLength=1
// +kubebuilder:validation:MaxLength=253
type SectionName string

// GatewayController is the name of a Gateway API controller. It must be a
// domain prefixed path.
//
// Valid values include:
//
// * "example.com/bar"
//
// Invalid values include:
//
// * "example.com" - must include path
// * "foo.example.com" - must include path
//
// +kubebuilder:validation:MinLength=1
// +kubebuilder:validation:MaxLength=253
// +kubebuilder:validation:Pattern=`^[a-z0-9]([-a-z0-9]*[a-z0-9])?(\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*\/[A-Za-z0-9\/\-._~%!$&'()*+,;=:]+$`
type GatewayController string
