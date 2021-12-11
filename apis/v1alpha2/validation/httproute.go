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
		errs = append(errs, validateHTTPRouteUniqueFilters(rule.Filters, path.Child("rules").Index(i))...)
		errs = append(errs, validateHTTPRouteFilterTypeMatchesValue(rule.Filters, path.Child("rules").Index(i))...)

		for j, backendRef := range rule.BackendRefs {
			errs = append(errs, validateHTTPRouteUniqueFilters(backendRef.Filters, path.Child("rules").Index(i).Child("backendsrefs").Index(j))...)
			errs = append(errs, validateHTTPRouteFilterTypeMatchesValue(backendRef.Filters, path.Child("rules").Index(i).Child("backendsrefs").Index(j))...)

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

// validateHTTPRouteUniqueFilters validates whether each core and extended filter
// is used at most once.
func validateHTTPRouteUniqueFilters(filters []gatewayv1a2.HTTPRouteFilter, path *field.Path) field.ErrorList {
	var errs field.ErrorList
	counts := map[gatewayv1a2.HTTPRouteFilterType]int{}

	for _, filter := range filters {
		counts[filter.Type]++
	}
	// custom filters don't have any validation
	for _, key := range repeatableHTTPRouteFilters {
		delete(counts, key)
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
		return append(allErrs, field.Required(fldPath.Child("pathType"), "pathType must be specified"))
	}

	if path.Value == nil {
		return append(allErrs, field.Required(fldPath.Child("pathValue"), "pathValue must not be nil."))
	}

	switch *path.Type {
	case gatewayv1a2.PathMatchExact, gatewayv1a2.PathMatchPathPrefix:
		if !strings.HasPrefix(*path.Value, "/") {
			allErrs = append(allErrs, field.Invalid(fldPath.Child("path"), path, "must be an absolute path"))
		}
		if len(*path.Value) > 0 {
			for _, invalidSeq := range invalidPathSequences {
				if strings.Contains(*path.Value, invalidSeq) {
					allErrs = append(allErrs, field.Invalid(fldPath.Child("path"), path, fmt.Sprintf("must not contain '%s'", invalidSeq)))
				}
			}

			for _, invalidSuff := range invalidPathSuffixes {
				if strings.HasSuffix(*path.Value, invalidSuff) {
					allErrs = append(allErrs, field.Invalid(fldPath.Child("path"), path, fmt.Sprintf("cannot end with '%s'", invalidSuff)))
				}
			}
		}
	case gatewayv1a2.PathMatchRegularExpression:
	default:
		pathTypes := []string{string(gatewayv1a2.PathMatchExact), string(gatewayv1a2.PathMatchPathPrefix), string(gatewayv1a2.PathMatchRegularExpression)}
		allErrs = append(allErrs, field.NotSupported(fldPath.Child("pathType"), *path.Type, pathTypes))
	}
	return allErrs
}

// validateHTTPRouteFilterTypeMatchesValue validates that each filter type matches
//the filter value.
func validateHTTPRouteFilterTypeMatchesValue(filters []gatewayv1a2.HTTPRouteFilter, path *field.Path) field.ErrorList {
	var errs field.ErrorList
	for _, filter := range filters {
		switch filter.Type {
		case gatewayv1a2.HTTPRouteFilterExtensionRef:
			if filter.ExtensionRef == nil {
				errs = append(errs, field.Invalid(path.Child("filters"), path, "extensionRef must be specified for ExtensionRef filter"))
			}
		case gatewayv1a2.HTTPRouteFilterRequestHeaderModifier:
			if filter.RequestHeaderModifier == nil {
				errs = append(errs, field.Invalid(path.Child("filters"), path, "requestHeaderModifier must be specified for RequestHeaderModifier filter"))
			}
		case gatewayv1a2.HTTPRouteFilterRequestMirror:
			if filter.RequestMirror == nil {
				errs = append(errs, field.Invalid(path.Child("filters"), path, "requestMirror must be specified for RequestMirror filter"))
			}
		case gatewayv1a2.HTTPRouteFilterRequestRedirect:
			if filter.RequestRedirect == nil {
				errs = append(errs, field.Invalid(path.Child("filters"), path, "requestRedirect must be specified for RequestRedirect filter"))
			}
		case gatewayv1a2.HTTPRouteFilterURLRewrite:
			if filter.URLRewrite == nil {
				errs = append(errs, field.Invalid(path.Child("filters"), path, "urlRewrite must be specified for URLRewrite filter"))
			}
		default:
			errs = append(errs, field.Invalid(path.Child("filters"), path, "unknown type of filter"))
		}
	}
	return errs
}
