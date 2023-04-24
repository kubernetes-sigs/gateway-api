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
	"fmt"
	"net/http"
	"regexp"
	"strings"

	"k8s.io/apimachinery/pkg/util/validation/field"

	gatewayv1a2 "sigs.k8s.io/gateway-api/apis/v1alpha2"
)

var (
	// repeatableGRPCRouteFilters are filter types that are allowed to be
	// repeated multiple times in a rule.
	repeatableGRPCRouteFilters = []gatewayv1a2.GRPCRouteFilterType{
		gatewayv1a2.GRPCRouteFilterExtensionRef,
	}
	validServiceName      = `^(?i)\.?[a-z_][a-z_0-9]*(\.[a-z_][a-z_0-9]*)*$`
	validServiceNameRegex = regexp.MustCompile(validServiceName)
	validMethodName       = `^[A-Za-z_][A-Za-z_0-9]*$`
	validMethodNameRegex  = regexp.MustCompile(validMethodName)
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
		errs = append(errs, validateGRPCRouteFilters(rule.Filters, path.Index(i).Child(("filters")))...)
		for j, backendRef := range rule.BackendRefs {
			errs = append(errs, validateGRPCRouteFilters(backendRef.Filters, path.Child("rules").Index(i).Child("backendRefs").Index(j))...)
		}
	}
	return errs
}

// validateRuleMatches validates GRPCMethodMatch
func validateRuleMatches(matches []gatewayv1a2.GRPCRouteMatch, path *field.Path) field.ErrorList {
	var errs field.ErrorList
	for i, m := range matches {
		if m.Method != nil {
			if m.Method.Service == nil && m.Method.Method == nil {
				errs = append(errs, field.Required(path.Index(i).Child("method"), "one or both of `service` or `method` must be specified"))
			}
			// GRPCRoute method matcher admits two types: Exact and RegularExpression.
			// If not specified, the match will be treated as type Exact (also the default value for this field).
			if m.Method.Type == nil || *m.Method.Type == gatewayv1a2.GRPCMethodMatchExact {
				if m.Method.Service != nil && !validServiceNameRegex.MatchString(*m.Method.Service) {
					errs = append(errs, field.Invalid(path.Index(i).Child("method"), *m.Method.Service,
						fmt.Sprintf("must only contain valid characters (matching %s)", validServiceName)))
				}
				if m.Method.Method != nil && !validMethodNameRegex.MatchString(*m.Method.Method) {
					errs = append(errs, field.Invalid(path.Index(i).Child("method"), *m.Method.Method,
						fmt.Sprintf("must only contain valid characters (matching %s)", validMethodName)))
				}
			}
		}
		if m.Headers != nil {
			errs = append(errs, validateGRPCHeaderMatches(m.Headers, path.Index(i).Child("headers"))...)
		}
	}
	return errs
}

// validateGRPCHeaderMatches validates that no header name is matched more than
// once (case-insensitive), and that at least one of service or method was
// provided.
func validateGRPCHeaderMatches(matches []gatewayv1a2.GRPCHeaderMatch, path *field.Path) field.ErrorList {
	var errs field.ErrorList
	counts := map[string]int{}

	for _, match := range matches {
		// Header names are case-insensitive.
		counts[strings.ToLower(string(match.Name))]++
	}

	for name, count := range counts {
		if count > 1 {
			errs = append(errs, field.Invalid(path, http.CanonicalHeaderKey(name), "cannot match the same header multiple times in the same rule"))
		}
	}

	return errs
}

// validateGRPCRouteFilterType validates that only the expected fields are
// set for the specified filter type.
func validateGRPCRouteFilterType(filter gatewayv1a2.GRPCRouteFilter, path *field.Path) field.ErrorList {
	var errs field.ErrorList
	if filter.ExtensionRef != nil && filter.Type != gatewayv1a2.GRPCRouteFilterExtensionRef {
		errs = append(errs, field.Invalid(path, filter.ExtensionRef, "must be nil if the GRPCRouteFilter.Type is not ExtensionRef"))
	}
	if filter.ExtensionRef == nil && filter.Type == gatewayv1a2.GRPCRouteFilterExtensionRef {
		errs = append(errs, field.Required(path, "filter.ExtensionRef must be specified for ExtensionRef GRPCRouteFilter.Type"))
	}
	if filter.RequestHeaderModifier != nil && filter.Type != gatewayv1a2.GRPCRouteFilterRequestHeaderModifier {
		errs = append(errs, field.Invalid(path, filter.RequestHeaderModifier, "must be nil if the GRPCRouteFilter.Type is not RequestHeaderModifier"))
	}
	if filter.RequestHeaderModifier == nil && filter.Type == gatewayv1a2.GRPCRouteFilterRequestHeaderModifier {
		errs = append(errs, field.Required(path, "filter.RequestHeaderModifier must be specified for RequestHeaderModifier GRPCRouteFilter.Type"))
	}
	if filter.ResponseHeaderModifier != nil && filter.Type != gatewayv1a2.GRPCRouteFilterResponseHeaderModifier {
		errs = append(errs, field.Invalid(path, filter.ResponseHeaderModifier, "must be nil if the GRPCRouteFilter.Type is not ResponseHeaderModifier"))
	}
	if filter.ResponseHeaderModifier == nil && filter.Type == gatewayv1a2.GRPCRouteFilterResponseHeaderModifier {
		errs = append(errs, field.Required(path, "filter.ResponseHeaderModifier must be specified for ResponseHeaderModifier GRPCRouteFilter.Type"))
	}
	if filter.RequestMirror != nil && filter.Type != gatewayv1a2.GRPCRouteFilterRequestMirror {
		errs = append(errs, field.Invalid(path, filter.RequestMirror, "must be nil if the GRPCRouteFilter.Type is not RequestMirror"))
	}
	if filter.RequestMirror == nil && filter.Type == gatewayv1a2.GRPCRouteFilterRequestMirror {
		errs = append(errs, field.Required(path, "filter.RequestMirror must be specified for RequestMirror GRPCRouteFilter.Type"))
	}
	return errs
}

// validateGRPCRouteFilters validates that a list of core and extended filters
// is used at most once and that the filter type matches its value
func validateGRPCRouteFilters(filters []gatewayv1a2.GRPCRouteFilter, path *field.Path) field.ErrorList {
	var errs field.ErrorList
	counts := map[gatewayv1a2.GRPCRouteFilterType]int{}

	for i, filter := range filters {
		counts[filter.Type]++
		if filter.RequestHeaderModifier != nil {
			errs = append(errs, validateGRPCHeaderModifier(*filter.RequestHeaderModifier, path.Index(i).Child("requestHeaderModifier"))...)
		}
		if filter.ResponseHeaderModifier != nil {
			errs = append(errs, validateGRPCHeaderModifier(*filter.ResponseHeaderModifier, path.Index(i).Child("responseHeaderModifier"))...)
		}
		errs = append(errs, validateGRPCRouteFilterType(filter, path.Index(i))...)
	}
	// custom filters don't have any validation
	for _, key := range repeatableGRPCRouteFilters {
		delete(counts, key)
	}

	for filterType, count := range counts {
		if count > 1 {
			errs = append(errs, field.Invalid(path, filterType, "cannot be used multiple times in the same rule"))
		}
	}
	return errs
}

// validateGRPCHeaderModifier ensures that multiple actions cannot be set for
// the same header.
func validateGRPCHeaderModifier(filter gatewayv1a2.HTTPHeaderFilter, path *field.Path) field.ErrorList {
	var errs field.ErrorList
	singleAction := make(map[string]bool)
	for i, action := range filter.Add {
		if needsErr, ok := singleAction[strings.ToLower(string(action.Name))]; ok {
			if needsErr {
				errs = append(errs, field.Invalid(path.Child("add"), filter.Add[i], "cannot specify multiple actions for header"))
			}
			singleAction[strings.ToLower(string(action.Name))] = false
		} else {
			singleAction[strings.ToLower(string(action.Name))] = true
		}
	}
	for i, action := range filter.Set {
		if needsErr, ok := singleAction[strings.ToLower(string(action.Name))]; ok {
			if needsErr {
				errs = append(errs, field.Invalid(path.Child("set"), filter.Set[i], "cannot specify multiple actions for header"))
			}
			singleAction[strings.ToLower(string(action.Name))] = false
		} else {
			singleAction[strings.ToLower(string(action.Name))] = true
		}
	}
	for i, name := range filter.Remove {
		if needsErr, ok := singleAction[strings.ToLower(name)]; ok {
			if needsErr {
				errs = append(errs, field.Invalid(path.Child("remove"), filter.Remove[i], "cannot specify multiple actions for header"))
			}
			singleAction[strings.ToLower(name)] = false
		} else {
			singleAction[strings.ToLower(name)] = true
		}
	}
	return errs
}
