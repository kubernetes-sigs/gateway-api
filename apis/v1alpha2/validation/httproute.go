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
	gatewayv1b1validation "sigs.k8s.io/gateway-api/apis/v1beta1/validation"
)

// ValidateHTTPRoute validates HTTPRoute according to the Gateway API specification.
// For additional details of the HTTPRoute spec, refer to:
// https://gateway-api.sigs.k8s.io/v1beta1/references/spec/#gateway.networking.k8s.io/v1beta1.HTTPRoute
func ValidateHTTPRoute(route *gatewayv1a2.HTTPRoute) field.ErrorList {
	return gatewayv1b1validation.ValidateHTTPRouteSpec(&route.Spec, field.NewPath("spec"))
}
