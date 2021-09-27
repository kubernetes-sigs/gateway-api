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
	"k8s.io/apimachinery/pkg/util/validation/field"

	gatewayv1a2 "sigs.k8s.io/gateway-api/apis/v1alpha2"
)

var (
	// repeatableHTTPRouteFilters are filter types that can are allowed to be
	// repeated multiple times in a rule.
	repeatableHTTPRouteFilters = []gatewayv1a2.HTTPRouteFilterType{
		gatewayv1a2.HTTPRouteFilterExtensionRef,
	}
)

// ValidateHTTPRoute validates HTTPRoute according to the Gateway API specification.
// For additional details of the HTTPRoute spec, refer to:
// https://gateway-api.sigs.k8s.io/spec/#gateway.networking.k8s.io/v1alpha2.HTTPRoute
func ValidateHTTPRoute(route *gatewayv1a2.HTTPRoute) field.ErrorList {
	return validateHTTPRouteSpec(&route.Spec, field.NewPath("spec"))
}

// validateHTTPRouteSpec validates that required fields of spec are set according to the
// HTTPRoute specification.
func validateHTTPRouteSpec(spec *gatewayv1a2.HTTPRouteSpec, path *field.Path) field.ErrorList {
	return validateHTTPRouteUniqueFilters(spec.Rules, path.Child("rules"))
}

// validateHTTPRouteUniqueFilters validates whether each core and extended filter
// is used at most once in each rule.
func validateHTTPRouteUniqueFilters(rules []gatewayv1a2.HTTPRouteRule, path *field.Path) field.ErrorList {
	var errs field.ErrorList

	for i, rule := range rules {
		counts := map[gatewayv1a2.HTTPRouteFilterType]int{}
		for _, filter := range rule.Filters {
			counts[filter.Type]++
		}
		// custom filters don't have any validation
		for _, key := range repeatableHTTPRouteFilters {
			delete(counts, key)
		}

		for filterType, count := range counts {
			if count > 1 {
				errs = append(errs, field.Invalid(path.Index(i).Child("filters"), filterType, "cannot be used multiple times in the same rule"))
			}
		}

		if errList := validateHTTPBackendUniqueFilters(rule.BackendRefs, path, i); len(errList) > 0 {
			errs = append(errs, errList...)
		}
	}

	return errs
}

func validateHTTPBackendUniqueFilters(ref []gatewayv1a2.HTTPBackendRef, path *field.Path, i int) field.ErrorList {
	var errs field.ErrorList

	for _, bkr := range ref {
		counts := map[gatewayv1a2.HTTPRouteFilterType]int{}
		for _, filter := range bkr.Filters {
			counts[filter.Type]++
		}

		for _, key := range repeatableHTTPRouteFilters {
			delete(counts, key)
		}

		for filterType, count := range counts {
			if count > 1 {
				errs = append(errs, field.Invalid(path.Index(i).Child("BackendRefs"), filterType, "cannot be used multiple times in the same backend"))
			}
		}
	}
	return errs
}
