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

// GatewayObjectReference identifies a Gateway object.
type GatewayObjectReference struct {
	// Namespace is the namespace of the referent.
	// +optional
	Namespace string `json:"namespace,omitempty"`
	// Name is the name of the referent.
	//
	// +kubebuilder:validation:Required
	// +required
	Name string `json:"name"`
}

// RouteConditionType is a type of condition for a route.
type RouteConditionType string

const (
	// ConditionRouteAdmitted indicates whether the route has been admitted
	// or rejected by a Gateway, and why.
	ConditionRouteAdmitted RouteConditionType = "Admitted"
)

// RouteCondition is a status condition for a given route.
type RouteCondition struct {
	// Type indicates the type of condition.
	Type RouteConditionType `json:"type"`
	// Status describes the current state of this condition.  Can be "True",
	// "False", or "Unknown".
	Status core.ConditionStatus `json:"status"`
	// Message is a human-understandable message describing the condition.
	// +optional
	Message string `json:"message,omitempty"`
	// Reason indicates why the condition is in this state.
	// +optional
	Reason string `json:"reason,omitempty"`
	// LastTransitionTime indicates the last time this condition changed.
	// +optional
	LastTransitionTime metav1.Time `json:"lastTransitionTime,omitempty"`
}

// RouteGatewayStatus describes the status of a route with respect to an
// associated Gateway.
type RouteGatewayStatus struct {
	// GatewayRef is a reference to a Gateway object that is associated with
	// the route.
	GatewayRef GatewayObjectReference `json:"gatewayRef"`
	// Conditions describes the status of the route with respect to the
	// Gateway.  For example, the "Admitted" condition indicates whether the
	// route has been admitted or rejected by the Gateway, and why.  Note
	// that the route's availability is also subject to the Gateway's own
	// status conditions and listener status.
	Conditions []RouteCondition `json:"conditions,omitempty"`
}

// RouteStatus defines the observed state that is required across
// all route types.
type RouteStatus struct {
	// Gateways is a list of the Gateways that are associated with the
	// route, and the status of the route with respect to each of these
	// Gateways. When a Gateway selects this route, the controller that
	// manages the Gateway should add an entry to this list when the
	// controller first sees the route and should update the entry as
	// appropriate when the route is modified.
	Gateways []RouteGatewayStatus `json:"gateways"`
}
