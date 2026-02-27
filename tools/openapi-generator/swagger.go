/*
Copyright The Kubernetes Authors.

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

package main

import (
	extensionv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apiextensions-apiserver/pkg/controller/openapi/builder"
	"k8s.io/kube-openapi/pkg/validation/spec"
)

func convertFromCrd(crd *extensionv1.CustomResourceDefinition) ([]*spec.Swagger, error) {
	var crdSpecs []*spec.Swagger
	for _, v := range crd.Spec.Versions {
		if !v.Served {
			continue
		}
		sw, err := builder.BuildOpenAPIV2(crd, v.Name, builder.Options{V2: true, StripValueValidation: false, StripNullable: false, AllowNonStructural: false, IncludeSelectableFields: true})
		if err != nil {
			return nil, err
		}
		crdSpecs = append(crdSpecs, sw)
	}
	return crdSpecs, nil
}

func convertFromCrds(crds []*extensionv1.CustomResourceDefinition) ([]*spec.Swagger, error) {
	var allSpecs []*spec.Swagger
	for _, resourceDefinition := range crds {
		specs, err := convertFromCrd(resourceDefinition)
		if err != nil {
			return nil, err
		}
		allSpecs = append(allSpecs, specs...)
	}
	return allSpecs, nil
}

func createStaticSpec(title string, version string) *spec.Swagger {
	return &spec.Swagger{
		SwaggerProps: spec.SwaggerProps{
			Info: &spec.Info{
				InfoProps: spec.InfoProps{
					Title:   title,
					Version: version,
				},
			},
			Swagger: "2.0",
		},
	}
}
