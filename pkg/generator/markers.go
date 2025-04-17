package main

import (
	apiext "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"sigs.k8s.io/controller-tools/pkg/markers"
)

type Minimum float64

func (m Minimum) Value() float64 {
	return float64(m)
}
func (m Minimum) ApplyToSchema(schema *apiext.JSONSchemaProps) error {
	val := m.Value()
	schema.Minimum = &val
	return nil
}

type Maximum float64

func (m Maximum) Value() float64 {
	return float64(m)
}
func (m Maximum) ApplyToSchema(schema *apiext.JSONSchemaProps) error {
	val := m.Value()
	schema.Maximum = &val
	return nil
}

// kubebuilder Min Max markers are broken with type aliases
func registerMarkerOverrides(into *markers.Registry) {
	min, _ := markers.MakeDefinition(
		"kubebuilder:validation:Minimum",
		markers.DescribesField,
		Minimum(0),
	)

	max, _ := markers.MakeDefinition(
		"kubebuilder:validation:Maximum",
		markers.DescribesField,
		Maximum(0),
	)
	into.Register(min)
	into.Register(max)
}
