/*
Copyright 2025 The Kubernetes Authors.

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
	apiext "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"sigs.k8s.io/controller-tools/pkg/markers"
)

type Minimum float64

func (m Minimum) Value() float64 {
	return float64(m)
}

//nolint:unparam
func (m Minimum) ApplyToSchema(schema *apiext.JSONSchemaProps) error {
	val := m.Value()
	schema.Minimum = &val
	return nil
}

type Maximum float64

func (m Maximum) Value() float64 {
	return float64(m)
}

//nolint:unparam
func (m Maximum) ApplyToSchema(schema *apiext.JSONSchemaProps) error {
	val := m.Value()
	schema.Maximum = &val
	return nil
}

// kubebuilder Min Max markers are broken with type aliases
func registerMarkerOverrides(into *markers.Registry) {
	minMarker, _ := markers.MakeDefinition(
		"kubebuilder:validation:Minimum",
		markers.DescribesField,
		Minimum(0),
	)

	maxMarker, _ := markers.MakeDefinition(
		"kubebuilder:validation:Maximum",
		markers.DescribesField,
		Maximum(0),
	)

	into.Register(minMarker) //nolint:errcheck
	into.Register(maxMarker) //nolint:errcheck
}
