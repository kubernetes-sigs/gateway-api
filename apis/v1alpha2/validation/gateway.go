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

// ValidateGateway validates gw according to the Gateway API specification.
// For additional details of the Gateway spec, refer to:
//  https://gateway-api.sigs.k8s.io/spec/#gateway.networking.k8s.io/v1alpha2.Gateway
//
// This function is currently empty as the validations have been moved to CRD field validations.
// It's expected that further validation will be added back here in the future.
// Here is the related PR for the reference:
//  https://github.com/kubernetes-sigs/gateway-api/pull/872
func ValidateGateway(gw *gatewayv1a2.Gateway) field.ErrorList {
	return nil
}
