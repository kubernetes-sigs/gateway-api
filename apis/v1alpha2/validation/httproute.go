/*
Copyright 2021 The Kubernetes Authors.

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
	"strings"

	"k8s.io/apimachinery/pkg/util/validation/field"

	gatewayv1a2 "sigs.k8s.io/gateway-api/apis/v1alpha2"
)

var (
	// repeatableHTTPRouteFilters are filter types that can are allowed to be
	// repeated multiple times in a rule.
	repeatableHTTPRouteFilters = []gatewayv1a2.HTTPRouteFilterType{
		gatewayv1a2.HTTPRouteFilterExtensionRef,
	}

	invalidPathSequences = []string{"//", "/./", "/../", "%2f", "%2F", "#"}
	invalidPathSuffixes  = []string{"/..", "/."}
)

// ValidateHTTPRoute validates HTTPRoute according to the Gateway API specification.
// For additional details of the HTTPRoute spec, refer to:
// https://gateway-api.sigs.k8s.io/v1alpha2/references/spec/#gateway.networking.k8s.io/v1alpha2.HTTPRoute
func ValidateHTTPRoute(route *gatewayv1a2.HTTPRoute) field.ErrorList {
	return validateHTTPRouteSpec(&route.Spec, field.NewPath("spec"))
}

// validateHTTPRouteSpec validates that required fields of spec are set according to the
// HTTPRoute specification.
func validateHTTPRouteSpec(spec *gatewayv1a2.HTTPRouteSpec, path *field.Path) field.ErrorList {
	var errs field.ErrorList
	for i, rule := range spec.Rules {
		errs = append(errs, validateHTTPRouteFilters(rule.Filters, path.Child("rules").Index(i))...)
		for j, backendRef := range rule.BackendRefs {
			errs = append(errs, validateHTTPRouteFilters(backendRef.Filters, path.Child("rules").Index(i).Child("backendsrefs").Index(j))...)
		}
		for j, m := range rule.Matches {
			if m.Path != nil {
				errs = append(errs, validateHTTPPathMatch(m.Path, path.Child("matches").Index(j).Child("path"))...)
			}
		}
	}
	errs = append(errs, validateHTTPRouteBackendServicePorts(spec.Rules, path.Child("rules"))...)
	return errs
}

// validateHTTPRouteBackendServicePorts validates that v1.Service backends always have a port.
func validateHTTPRouteBackendServicePorts(rules []gatewayv1a2.HTTPRouteRule, path *field.Path) field.ErrorList {
	var errs field.ErrorList

	for i, rule := range rules {
		path = path.Index(i).Child("backendRefs")
		for i, ref := range rule.BackendRefs {
			if ref.BackendObjectReference.Group != nil &&
				*ref.BackendObjectReference.Group != "" {
				continue
			}

			if ref.BackendObjectReference.Kind != nil &&
				*ref.BackendObjectReference.Kind != "Service" {
				continue
			}

			if ref.BackendObjectReference.Port == nil {
				errs = append(errs, field.Required(path.Index(i).Child("port"), "missing port for Service reference"))
			}
		}
	}

	return errs
}

// validateHTTPRouteFilters validates that a list of core and extended filters
// is used at most once and that the filter type matches its value
func validateHTTPRouteFilters(filters []gatewayv1a2.HTTPRouteFilter, path *field.Path) field.ErrorList {
	var errs field.ErrorList
	counts := map[gatewayv1a2.HTTPRouteFilterType]int{}

	for i, filter := range filters {
		counts[filter.Type]++
		if filter.RequestRedirect != nil && filter.RequestRedirect.Path != nil {
			errs = append(errs, validateHTTPPathModifier(*filter.RequestRedirect.Path, path.Index(i).Child("requestRedirect", "path"))...)
		}
		if filter.URLRewrite != nil && filter.URLRewrite.Path != nil {
			errs = append(errs, validateHTTPPathModifier(*filter.URLRewrite.Path, path.Index(i).Child("urlRewrite", "path"))...)
		}
		errs = append(errs, validateHTTPRouteFilterTypeMatchesValue(filter, path.Index(i))...)
	}
	// custom filters don't have any validation
	for _, key := range repeatableHTTPRouteFilters {
		delete(counts, key)
	}

	if counts[gatewayv1a2.HTTPRouteFilterRequestRedirect] > 0 && counts[gatewayv1a2.HTTPRouteFilterURLRewrite] > 0 {
		errs = append(errs, field.Invalid(path.Child("filters"), gatewayv1a2.HTTPRouteFilterRequestRedirect, "Redirect and Rewrite filters cannot be defined in the same list of filters"))
	}

	for filterType, count := range counts {
		if count > 1 {
			errs = append(errs, field.Invalid(path.Child("filters"), filterType, "cannot be used multiple times in the same rule"))
		}
	}
	return errs
}

// webhook validation of HTTPPathMatch
func validateHTTPPathMatch(path *gatewayv1a2.HTTPPathMatch, fldPath *field.Path) field.ErrorList {
	allErrs := field.ErrorList{}

	if path.Type == nil {
		return append(allErrs, field.Required(fldPath.Child("type"), "must be specified"))
	}

	if path.Value == nil {
		return append(allErrs, field.Required(fldPath.Child("value"), "must be specified"))
	}

	switch *path.Type {
	case gatewayv1a2.PathMatchExact, gatewayv1a2.PathMatchPathPrefix:
		if !strings.HasPrefix(*path.Value, "/") {
			allErrs = append(allErrs, field.Invalid(fldPath.Child("value"), *path.Value, "must be an absolute path"))
		}
		if len(*path.Value) > 0 {
			for _, invalidSeq := range invalidPathSequences {
				if strings.Contains(*path.Value, invalidSeq) {
					allErrs = append(allErrs, field.Invalid(fldPath.Child("value"), *path.Value, fmt.Sprintf("must not contain %q", invalidSeq)))
				}
			}

			for _, invalidSuff := range invalidPathSuffixes {
				if strings.HasSuffix(*path.Value, invalidSuff) {
					allErrs = append(allErrs, field.Invalid(fldPath.Child("value"), *path.Value, fmt.Sprintf("cannot end with '%s'", invalidSuff)))
				}
			}
		}
	case gatewayv1a2.PathMatchRegularExpression:
	default:
		pathTypes := []string{string(gatewayv1a2.PathMatchExact), string(gatewayv1a2.PathMatchPathPrefix), string(gatewayv1a2.PathMatchRegularExpression)}
		allErrs = append(allErrs, field.NotSupported(fldPath.Child("type"), *path.Type, pathTypes))
	}
	return allErrs
}

// validateHTTPRouteFilterTypeMatchesValue validates that only the expected fields are
//// set for the specified filter type.
func validateHTTPRouteFilterTypeMatchesValue(filter gatewayv1a2.HTTPRouteFilter, path *field.Path) field.ErrorList {
	var errs field.ErrorList
	if filter.ExtensionRef != nil && filter.Type != gatewayv1a2.HTTPRouteFilterExtensionRef {
		errs = append(errs, field.Invalid(path, filter.ExtensionRef, "must be nil if the HTTPRouteFilter.Type is not ExtensionRef"))
	}
	if filter.ExtensionRef == nil && filter.Type == gatewayv1a2.HTTPRouteFilterExtensionRef {
		errs = append(errs, field.Required(path, "filter.ExtensionRef must be specified for ExtensionRef HTTPRouteFilter.Type"))
	}
	if filter.RequestHeaderModifier != nil && filter.Type != gatewayv1a2.HTTPRouteFilterRequestHeaderModifier {
		errs = append(errs, field.Invalid(path, filter.RequestHeaderModifier, "must be nil if the HTTPRouteFilter.Type is not RequestHeaderModifier"))
	}
	if filter.RequestHeaderModifier == nil && filter.Type == gatewayv1a2.HTTPRouteFilterRequestHeaderModifier {
		errs = append(errs, field.Required(path, "filter.RequestHeaderModifier must be specified for RequestHeaderModifier HTTPRouteFilter.Type"))
	}
	if filter.RequestMirror != nil && filter.Type != gatewayv1a2.HTTPRouteFilterRequestMirror {
		errs = append(errs, field.Invalid(path, filter.RequestMirror, "must be nil if the HTTPRouteFilter.Type is not RequestMirror"))
	}
	if filter.RequestMirror == nil && filter.Type == gatewayv1a2.HTTPRouteFilterRequestMirror {
		errs = append(errs, field.Required(path, "filter.RequestMirror must be specified for RequestMirror HTTPRouteFilter.Type"))
	}
	if filter.RequestRedirect != nil && filter.Type != gatewayv1a2.HTTPRouteFilterRequestRedirect {
		errs = append(errs, field.Invalid(path, filter.RequestRedirect, "must be nil if the HTTPRouteFilter.Type is not RequestRedirect"))
	}
	if filter.RequestRedirect == nil && filter.Type == gatewayv1a2.HTTPRouteFilterRequestRedirect {
		errs = append(errs, field.Required(path, "filter.RequestRedirect must be specified for RequestRedirect HTTPRouteFilter.Type"))
	}
	if filter.URLRewrite != nil && filter.Type != gatewayv1a2.HTTPRouteFilterURLRewrite {
		errs = append(errs, field.Invalid(path, filter.URLRewrite, "must be nil if the HTTPRouteFilter.Type is not URLRewrite"))
	}
	if filter.URLRewrite == nil && filter.Type == gatewayv1a2.HTTPRouteFilterURLRewrite {
		errs = append(errs, field.Required(path, "filter.URLRewrite must be specified for URLRewrite HTTPRouteFilter.Type"))
	}
	return errs
}

// validateHTTPPathModifier validates that only the expected fields are set in a
// path modifier.
func validateHTTPPathModifier(modifier gatewayv1a2.HTTPPathModifier, path *field.Path) field.ErrorList {
	var errs field.ErrorList
	if modifier.ReplaceFullPath != nil && modifier.Type != gatewayv1a2.FullPathHTTPPathModifier {
		errs = append(errs, field.Invalid(path, modifier.ReplaceFullPath, "must be nil if the HTTPRouteFilter.Type is not ReplaceFullPath"))
	}
	if modifier.ReplaceFullPath == nil && modifier.Type == gatewayv1a2.FullPathHTTPPathModifier {
		errs = append(errs, field.Invalid(path, modifier.ReplaceFullPath, "must not be nil if the HTTPRouteFilter.Type is ReplaceFullPath"))
	}
	if modifier.ReplacePrefixMatch != nil && modifier.Type != gatewayv1a2.PrefixMatchHTTPPathModifier {
		errs = append(errs, field.Invalid(path, modifier.ReplacePrefixMatch, "must be nil if the HTTPRouteFilter.Type is not ReplacePrefixMatch"))
	}
	if modifier.ReplacePrefixMatch == nil && modifier.Type == gatewayv1a2.PrefixMatchHTTPPathModifier {
		errs = append(errs, field.Invalid(path, modifier.ReplacePrefixMatch, "must not be nil if the HTTPRouteFilter.Type is ReplacePrefixMatch"))
	}
	return errs
}
