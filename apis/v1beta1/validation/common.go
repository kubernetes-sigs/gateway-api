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

	gatewayv1b1 "sigs.k8s.io/gateway-api/apis/v1beta1"
	utils "sigs.k8s.io/gateway-api/apis/v1beta1/util/validation"
)

// ValidateParentRefs validates ParentRefs SectionName must be set and uique
// when ParentRefs includes 2 or more references to the same parent
func ValidateParentRefs(parentRefs []gatewayv1b1.ParentReference, path *field.Path) field.ErrorList {
	var errs field.ErrorList
	if len(parentRefs) <= 1 {
		return nil
	}
	type sameKindParentRefs struct {
		name      gatewayv1b1.ObjectName
		namespace gatewayv1b1.Namespace
		kind      gatewayv1b1.Kind
	}
	parentRefsSectionMap := make(map[sameKindParentRefs][]gatewayv1b1.SectionName)
	for i, p := range parentRefs {
		targetParentRefs := sameKindParentRefs{name: p.Name, namespace: *new(gatewayv1b1.Namespace), kind: *new(gatewayv1b1.Kind)}
		targetSection := new(gatewayv1b1.SectionName)
		if p.Namespace != nil {
			targetParentRefs.namespace = *p.Namespace
		}
		if p.Kind != nil {
			targetParentRefs.kind = *p.Kind
		}
		if p.SectionName != nil {
			targetSection = p.SectionName
		}
		if s, ok := parentRefsSectionMap[targetParentRefs]; ok {
			if len(s[0]) == 0 || len(*targetSection) == 0 {
				errs = append(errs, field.Required(path, "ParentRefs section names must all be set when ParentRefs includes 2 or more references to the same parent"))
				return errs
			}
			if utils.ContainsInSectionNameSlice(s, targetSection) {
				errs = append(errs, field.Invalid(path.Index(i).Child("parentRefs").Child("sectionName"), targetSection, "must be unique when ParentRefs includes 2 or more references to the same parent"))
				return errs
			}
		}
		parentRefsSectionMap[targetParentRefs] = append(parentRefsSectionMap[targetParentRefs], *targetSection)
	}
	return errs
}
