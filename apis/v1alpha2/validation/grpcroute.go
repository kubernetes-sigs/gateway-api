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

// ValidateGRPCRoute validates GRPCRoute according to the Gateway API specification.
// For additional details of the GRPCRoute spec, refer to:
// https://gateway-api.sigs.k8s.io/v1alpha2/references/spec/#gateway.networking.k8s.io/v1alpha2.GRPCRoute
func ValidateGRPCRoute(route *gatewayv1a2.GRPCRoute) field.ErrorList {
	return validateGRPCRouteSpec(&route.Spec, field.NewPath("spec"))
}

// validateRouteSpec validates that required fields of spec are set according to the
// Gateway API specification.
func validateGRPCRouteSpec(spec *gatewayv1a2.GRPCRouteSpec, path *field.Path) field.ErrorList {
	var errs field.ErrorList
	errs = append(errs, validateGRPCRouteRules(spec.Rules, path.Child("rules"))...)
	errs = append(errs, validateParentRefs(spec.ParentRefs, path.Child("spec"))...)
	return errs
}

// validateGRPCRouteRules validates whether required fields of rules are set according
// to the Gateway API specification.
func validateGRPCRouteRules(rules []gatewayv1a2.GRPCRouteRule, path *field.Path) field.ErrorList {
	var errs field.ErrorList
	for i, rule := range rules {
		errs = append(errs, validateRuleMatches(rule.Matches, path.Index(i).Child("matches"))...)
	}
	return errs
}

// validateRuleMatches validates that at least one of the fields Service or Method of
// GRPCMethodMatch to be specified
func validateRuleMatches(matches []gatewayv1a2.GRPCRouteMatch, path *field.Path) field.ErrorList {
	var errs field.ErrorList
	for i, m := range matches {
		if m.Method != nil && m.Method.Service == nil && m.Method.Method == nil {
			errs = append(errs, field.Required(path.Index(i).Child("methods"), "should have at least one of fields Service or Method"))
			return errs
		}
	}
	return errs
}
