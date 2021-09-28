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

// ValidateGatewayClassUpdate validates an update to oldClass according to the
// Gateway API specification. For additional details of the GatewayClass spec, refer to:
// https://gateway-api.sigs.k8s.io/v1alpha2/references/spec/#gateway.networking.k8s.io/v1alpha2.GatewayClass
func ValidateGatewayClassUpdate(oldClass, newClass *gatewayv1a2.GatewayClass) field.ErrorList {
	if oldClass == nil || newClass == nil {
		return nil
	}
	var errs field.ErrorList
	if oldClass.Spec.ControllerName != newClass.Spec.ControllerName {
		errs = append(errs, field.Invalid(field.NewPath("spec.controllerName"), newClass.Spec.ControllerName,
			"cannot update an immutable field"))
	}
	return errs
}
