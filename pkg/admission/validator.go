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

package admission

import "sigs.k8s.io/gateway-api/apis/v1alpha1"

// ValidateHTTPRoute checks if a httpRoute contains only 1 of
// each core or extended HTTPRouteFilterType. ImplementationSpecific
// types such as ExtensionRef filter counts are not restricted.
func ValidateHTTPRoute(httpRoute v1alpha1.HTTPRoute) (bool, string, error) {
	if len(httpRoute.Spec.Rules) == 0 {
		return true, "", nil
	}
	numReqHeaderMod := 0
	numReqMirror := 0

	for _, rule := range httpRoute.Spec.Rules {
		if len(rule.Filters) != 0 {
			for _, filter := range rule.Filters {
				switch filter.Type {
				case v1alpha1.HTTPRouteFilterRequestHeaderModifier:
					numReqHeaderMod++
				case v1alpha1.HTTPRouteFilterRequestMirror:
					numReqMirror++
				case v1alpha1.HTTPRouteFilterExtensionRef:
					continue
				}
			}
		}
		// Multiple filters are allowed but only 1 of each core or extended HTTPRouteFilterType
		if numReqHeaderMod > 1 || numReqMirror > 1 {
			return false, "HTTPRules cannot contain more than one instance of each core or extended HTTPRouteFilterType", nil
		}

	}

	return true, "", nil
}
