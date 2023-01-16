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
	"net/http"
	"regexp"
	"strings"

	"k8s.io/apimachinery/pkg/util/validation/field"

	gatewayv1b1 "sigs.k8s.io/gateway-api/apis/v1beta1"
)

var (
	// repeatableHTTPRouteFilters are filter types that are allowed to be
	// repeated multiple times in a rule.
	repeatableHTTPRouteFilters = []gatewayv1b1.HTTPRouteFilterType{
		gatewayv1b1.HTTPRouteFilterExtensionRef,
	}

	// Invalid path sequences and suffixes, primarily related to directory traversal
	invalidPathSequences = []string{"//", "/./", "/../", "%2f", "%2F", "#"}
	invalidPathSuffixes  = []string{"/..", "/."}

	// All valid path characters per RFC-3986
	validPathCharacters = "^(?:[A-Za-z0-9\\/\\-._~!$&'()*+,;=:@]|[%][0-9a-fA-F]{2})+$"
)

// ValidateHTTPRoute validates HTTPRoute according to the Gateway API specification.
// For additional details of the HTTPRoute spec, refer to:
// https://gateway-api.sigs.k8s.io/v1beta1/references/spec/#gateway.networking.k8s.io/v1beta1.HTTPRoute
func ValidateHTTPRoute(route *gatewayv1b1.HTTPRoute) field.ErrorList {
	return ValidateHTTPRouteSpec(&route.Spec, field.NewPath("spec"))
}

// ValidateHTTPRouteSpec validates that required fields of spec are set according to the
// HTTPRoute specification.
func ValidateHTTPRouteSpec(spec *gatewayv1b1.HTTPRouteSpec, path *field.Path) field.ErrorList {
	var errs field.ErrorList
	for i, rule := range spec.Rules {
		errs = append(errs, validateHTTPRouteFilters(rule.Filters, rule.Matches, path.Child("rules").Index(i))...)
		for j, backendRef := range rule.BackendRefs {
			errs = append(errs, validateHTTPRouteFilters(backendRef.Filters, rule.Matches, path.Child("rules").Index(i).Child("backendRefs").Index(j))...)
		}
		for j, m := range rule.Matches {
			matchPath := path.Child("rules").Index(i).Child("matches").Index(j)

			if m.Path != nil {
				errs = append(errs, validateHTTPPathMatch(m.Path, matchPath.Child("path"))...)
			}
			if len(m.Headers) > 0 {
				errs = append(errs, validateHTTPHeaderMatches(m.Headers, matchPath.Child("headers"))...)
			}
			if len(m.QueryParams) > 0 {
				errs = append(errs, validateHTTPQueryParamMatches(m.QueryParams, matchPath.Child("queryParams"))...)
			}
		}
	}
	errs = append(errs, validateHTTPRouteBackendServicePorts(spec.Rules, path.Child("rules"))...)
	errs = append(errs, ValidateParentRefs(spec.ParentRefs, path.Child("spec"))...)
	return errs
}

// validateHTTPRouteBackendServicePorts validates that v1.Service backends always have a port.
func validateHTTPRouteBackendServicePorts(rules []gatewayv1b1.HTTPRouteRule, path *field.Path) field.ErrorList {
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
func validateHTTPRouteFilters(filters []gatewayv1b1.HTTPRouteFilter, matches []gatewayv1b1.HTTPRouteMatch, path *field.Path) field.ErrorList {
	var errs field.ErrorList
	counts := map[gatewayv1b1.HTTPRouteFilterType]int{}

	for i, filter := range filters {
		counts[filter.Type]++
		if filter.RequestRedirect != nil && filter.RequestRedirect.Path != nil {
			errs = append(errs, validateHTTPPathModifier(*filter.RequestRedirect.Path, matches, path.Index(i).Child("requestRedirect", "path"))...)
		}
		if filter.URLRewrite != nil && filter.URLRewrite.Path != nil {
			errs = append(errs, validateHTTPPathModifier(*filter.URLRewrite.Path, matches, path.Index(i).Child("urlRewrite", "path"))...)
		}
		if filter.RequestHeaderModifier != nil {
			errs = append(errs, validateHTTPHeaderModifier(*filter.RequestHeaderModifier, path.Index(i).Child("requestHeaderModifier"))...)
		}
		if filter.ResponseHeaderModifier != nil {
			errs = append(errs, validateHTTPHeaderModifier(*filter.ResponseHeaderModifier, path.Index(i).Child("responseHeaderModifier"))...)
		}
		errs = append(errs, validateHTTPRouteFilterTypeMatchesValue(filter, path.Index(i))...)
	}
	// custom filters don't have any validation
	for _, key := range repeatableHTTPRouteFilters {
		delete(counts, key)
	}

	if counts[gatewayv1b1.HTTPRouteFilterRequestRedirect] > 0 && counts[gatewayv1b1.HTTPRouteFilterURLRewrite] > 0 {
		errs = append(errs, field.Invalid(path.Child("filters"), gatewayv1b1.HTTPRouteFilterRequestRedirect, "may specify either httpRouteFilterRequestRedirect or httpRouteFilterRequestRewrite, but not both"))
	}

	for filterType, count := range counts {
		if count > 1 {
			errs = append(errs, field.Invalid(path.Child("filters"), filterType, "cannot be used multiple times in the same rule"))
		}
	}
	return errs
}

// webhook validation of HTTPPathMatch
func validateHTTPPathMatch(path *gatewayv1b1.HTTPPathMatch, fldPath *field.Path) field.ErrorList {
	allErrs := field.ErrorList{}

	if path.Type == nil {
		return append(allErrs, field.Required(fldPath.Child("type"), "must be specified"))
	}

	if path.Value == nil {
		return append(allErrs, field.Required(fldPath.Child("value"), "must be specified"))
	}

	switch *path.Type {
	case gatewayv1b1.PathMatchExact, gatewayv1b1.PathMatchPathPrefix:
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

		r, err := regexp.Compile(validPathCharacters)
		if err != nil {
			allErrs = append(allErrs, field.InternalError(fldPath.Child("value"),
				fmt.Errorf("could not compile path matching regex: %w", err)))
		} else if !r.MatchString(*path.Value) {
			allErrs = append(allErrs, field.Invalid(fldPath.Child("value"), *path.Value,
				fmt.Sprintf("must only contain valid characters (matching %s)", validPathCharacters)))
		}

	case gatewayv1b1.PathMatchRegularExpression:
	default:
		pathTypes := []string{string(gatewayv1b1.PathMatchExact), string(gatewayv1b1.PathMatchPathPrefix), string(gatewayv1b1.PathMatchRegularExpression)}
		allErrs = append(allErrs, field.NotSupported(fldPath.Child("type"), *path.Type, pathTypes))
	}
	return allErrs
}

// validateHTTPHeaderMatches validates that no header name
// is matched more than once (case-insensitive).
func validateHTTPHeaderMatches(matches []gatewayv1b1.HTTPHeaderMatch, path *field.Path) field.ErrorList {
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

// validateHTTPQueryParamMatches validates that no query param name
// is matched more than once (case-sensitive).
func validateHTTPQueryParamMatches(matches []gatewayv1b1.HTTPQueryParamMatch, path *field.Path) field.ErrorList {
	var errs field.ErrorList
	counts := map[string]int{}

	for _, match := range matches {
		// Query param names are case-sensitive.
		counts[string(match.Name)]++
	}

	for name, count := range counts {
		if count > 1 {
			errs = append(errs, field.Invalid(path, name, "cannot match the same query parameter multiple times in the same rule"))
		}
	}

	return errs
}

// validateHTTPRouteFilterTypeMatchesValue validates that only the expected fields are
// set for the specified filter type.
func validateHTTPRouteFilterTypeMatchesValue(filter gatewayv1b1.HTTPRouteFilter, path *field.Path) field.ErrorList {
	var errs field.ErrorList
	if filter.ExtensionRef != nil && filter.Type != gatewayv1b1.HTTPRouteFilterExtensionRef {
		errs = append(errs, field.Invalid(path, filter.ExtensionRef, "must be nil if the HTTPRouteFilter.Type is not ExtensionRef"))
	}
	if filter.ExtensionRef == nil && filter.Type == gatewayv1b1.HTTPRouteFilterExtensionRef {
		errs = append(errs, field.Required(path, "filter.ExtensionRef must be specified for ExtensionRef HTTPRouteFilter.Type"))
	}
	if filter.RequestHeaderModifier != nil && filter.Type != gatewayv1b1.HTTPRouteFilterRequestHeaderModifier {
		errs = append(errs, field.Invalid(path, filter.RequestHeaderModifier, "must be nil if the HTTPRouteFilter.Type is not RequestHeaderModifier"))
	}
	if filter.RequestHeaderModifier == nil && filter.Type == gatewayv1b1.HTTPRouteFilterRequestHeaderModifier {
		errs = append(errs, field.Required(path, "filter.RequestHeaderModifier must be specified for RequestHeaderModifier HTTPRouteFilter.Type"))
	}
	if filter.ResponseHeaderModifier != nil && filter.Type != gatewayv1b1.HTTPRouteFilterResponseHeaderModifier {
		errs = append(errs, field.Invalid(path, filter.ResponseHeaderModifier, "must be nil if the HTTPRouteFilter.Type is not ResponseHeaderModifier"))
	}
	if filter.ResponseHeaderModifier == nil && filter.Type == gatewayv1b1.HTTPRouteFilterResponseHeaderModifier {
		errs = append(errs, field.Required(path, "filter.ResponseHeaderModifier must be specified for ResponseHeaderModifier HTTPRouteFilter.Type"))
	}
	if filter.RequestMirror != nil && filter.Type != gatewayv1b1.HTTPRouteFilterRequestMirror {
		errs = append(errs, field.Invalid(path, filter.RequestMirror, "must be nil if the HTTPRouteFilter.Type is not RequestMirror"))
	}
	if filter.RequestMirror == nil && filter.Type == gatewayv1b1.HTTPRouteFilterRequestMirror {
		errs = append(errs, field.Required(path, "filter.RequestMirror must be specified for RequestMirror HTTPRouteFilter.Type"))
	}
	if filter.RequestRedirect != nil && filter.Type != gatewayv1b1.HTTPRouteFilterRequestRedirect {
		errs = append(errs, field.Invalid(path, filter.RequestRedirect, "must be nil if the HTTPRouteFilter.Type is not RequestRedirect"))
	}
	if filter.RequestRedirect == nil && filter.Type == gatewayv1b1.HTTPRouteFilterRequestRedirect {
		errs = append(errs, field.Required(path, "filter.RequestRedirect must be specified for RequestRedirect HTTPRouteFilter.Type"))
	}
	if filter.URLRewrite != nil && filter.Type != gatewayv1b1.HTTPRouteFilterURLRewrite {
		errs = append(errs, field.Invalid(path, filter.URLRewrite, "must be nil if the HTTPRouteFilter.Type is not URLRewrite"))
	}
	if filter.URLRewrite == nil && filter.Type == gatewayv1b1.HTTPRouteFilterURLRewrite {
		errs = append(errs, field.Required(path, "filter.URLRewrite must be specified for URLRewrite HTTPRouteFilter.Type"))
	}
	return errs
}

// validateHTTPPathModifier validates that only the expected fields are set in a
// path modifier.
func validateHTTPPathModifier(modifier gatewayv1b1.HTTPPathModifier, matches []gatewayv1b1.HTTPRouteMatch, path *field.Path) field.ErrorList {
	var errs field.ErrorList
	if modifier.ReplaceFullPath != nil && modifier.Type != gatewayv1b1.FullPathHTTPPathModifier {
		errs = append(errs, field.Invalid(path, modifier.ReplaceFullPath, "must be nil if the HTTPRouteFilter.Type is not ReplaceFullPath"))
	}
	if modifier.ReplaceFullPath == nil && modifier.Type == gatewayv1b1.FullPathHTTPPathModifier {
		errs = append(errs, field.Invalid(path, modifier.ReplaceFullPath, "must not be nil if the HTTPRouteFilter.Type is ReplaceFullPath"))
	}
	if modifier.ReplacePrefixMatch != nil && modifier.Type != gatewayv1b1.PrefixMatchHTTPPathModifier {
		errs = append(errs, field.Invalid(path, modifier.ReplacePrefixMatch, "must be nil if the HTTPRouteFilter.Type is not ReplacePrefixMatch"))
	}
	if modifier.ReplacePrefixMatch == nil && modifier.Type == gatewayv1b1.PrefixMatchHTTPPathModifier {
		errs = append(errs, field.Invalid(path, modifier.ReplacePrefixMatch, "must not be nil if the HTTPRouteFilter.Type is ReplacePrefixMatch"))
	}

	if modifier.Type == gatewayv1b1.PrefixMatchHTTPPathModifier && modifier.ReplacePrefixMatch != nil {
		if !hasExactlyOnePrefixMatch(matches) {
			errs = append(errs, field.Invalid(path, modifier.ReplacePrefixMatch, "exactly one PathPrefix match must be specified to use this path modifier"))
		}
	}
	return errs
}

func validateHTTPHeaderModifier(filter gatewayv1b1.HTTPHeaderFilter, path *field.Path) field.ErrorList {
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

func hasExactlyOnePrefixMatch(matches []gatewayv1b1.HTTPRouteMatch) bool {
	if len(matches) != 1 || matches[0].Path == nil {
		return false
	}
	pathMatchType := matches[0].Path.Type
	if *pathMatchType != gatewayv1b1.PathMatchPathPrefix {
		return false
	}

	return true
}
