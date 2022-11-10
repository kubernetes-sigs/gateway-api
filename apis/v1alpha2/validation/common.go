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

	v1a2 "sigs.k8s.io/gateway-api/apis/v1alpha2"
	gatewayvalidationv1b1 "sigs.k8s.io/gateway-api/apis/v1beta1/validation"
)

var (
	// validateParentRefs validates ParentRefs SectionName must be set and uique
	// when ParentRefs includes 2 or more references to the same parent
	validateParentRefs = gatewayvalidationv1b1.ValidateParentRefs
)

func validateBackendRefServicePort(ref *v1a2.BackendRef, path *field.Path) field.ErrorList {
	var errs field.ErrorList

	if ref.Group != nil && *ref.Group != "" {
		return nil
	}

	if ref.Kind != nil && *ref.Kind != "Service" {
		return nil
	}

	if ref.Port == nil {
		errs = append(errs, field.Required(path.Child("port"), "missing port for Service reference"))
	}

	return errs
}
