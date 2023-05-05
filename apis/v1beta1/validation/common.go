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
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/apimachinery/pkg/util/validation/field"

	gatewayv1b1 "sigs.k8s.io/gateway-api/apis/v1beta1"
)

// ValidateParentRefs validates ParentRefs SectionName must be set and unique
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
	type parentQualifier struct {
		section gatewayv1b1.SectionName
		port    gatewayv1b1.PortNumber
	}
	parentRefsSectionMap := make(map[sameKindParentRefs]sets.Set[parentQualifier])
	for i, p := range parentRefs {
		targetParentRefs := sameKindParentRefs{name: p.Name, namespace: *new(gatewayv1b1.Namespace), kind: *new(gatewayv1b1.Kind)}
		pq := parentQualifier{}
		if p.Namespace != nil {
			targetParentRefs.namespace = *p.Namespace
		}
		if p.Kind != nil {
			targetParentRefs.kind = *p.Kind
		}
		if p.SectionName != nil {
			pq.section = *p.SectionName
		}
		if p.Port != nil {
			pq.port = *p.Port
		}
		if s, ok := parentRefsSectionMap[targetParentRefs]; ok {
			if s.UnsortedList()[0] == (parentQualifier{}) || pq == (parentQualifier{}) {
				errs = append(errs, field.Required(path.Child("parentRefs"), "sectionNames or ports must be specified when more than one parentRef refers to the same parent"))
				return errs
			}
			if s.Has(pq) {
				fieldPath := path.Index(i).Child("parentRefs")
				var val any
				if len(pq.section) > 0 {
					fieldPath = fieldPath.Child("sectionName")
					val = pq.section
				} else {
					fieldPath = fieldPath.Child("port")
					val = pq.port
				}
				errs = append(errs, field.Invalid(fieldPath, val, "must be unique when ParentRefs includes 2 or more references to the same parent"))
				return errs
			}
			parentRefsSectionMap[targetParentRefs].Insert(pq)
		} else {
			parentRefsSectionMap[targetParentRefs] = sets.New(pq)
		}
	}
	return errs
}

func ptrTo[T any](a T) *T {
	return &a
}
