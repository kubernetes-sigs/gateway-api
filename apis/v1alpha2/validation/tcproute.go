/*
Copyright 2022 The Kubernetes Authors.

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

package validation

import (
	"k8s.io/apimachinery/pkg/util/validation/field"

	gatewayv1a2 "sigs.k8s.io/gateway-api/apis/v1alpha2"
)

// ValidateTCPRoute validates TCPRoute according to the Gateway API specification.
// For additional details of the TCPRoute spec, refer to:
// https://gateway-api.sigs.k8s.io/references/spec/#gateway.networking.k8s.io/v1alpha2.TCPRoute
func ValidateTCPRoute(route *gatewayv1a2.TCPRoute) field.ErrorList {
	return validateTCPRouteSpec(&route.Spec, field.NewPath("spec"))
}

// validateTCPRouteSpec validates that required fields of spec are set according to the
// TCPRoute specification.
func validateTCPRouteSpec(spec *gatewayv1a2.TCPRouteSpec, path *field.Path) field.ErrorList {
	path = path.Child("rules")
	var errs field.ErrorList
	for i, rule := range spec.Rules {
		for j, ref := range rule.BackendRefs {
			// Avoid referencing to the loop variable.
			ref := ref
			errs = append(errs, validateBackendRefServicePort(&ref, path.Index(i).Child("backendRefs").Index(j))...)
			errs = append(errs, validateParentRefs(spec.ParentRefs, path.Child("spec"))...)
		}
	}
	return errs
}
