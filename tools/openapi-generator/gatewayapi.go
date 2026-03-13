package main

import (
	"strings"

	"k8s.io/kube-openapi/pkg/validation/spec"

	"sigs.k8s.io/gateway-api/pkg/generated/openapi"
)

func generateGatewayAPIModels() map[string]spec.Schema {
	models := make(map[string]spec.Schema)
	definitions := openapi.GetOpenAPIDefinitions(func(path string) spec.Ref {
		return spec.MustCreateRef(standarizePath(path))
	})
	for name, definition := range definitions {
		models[standarizePath(name)] = definition.Schema
	}
	return models
}

// Turn sigs.k8s.io/gateway-api/apis/v1.Gateway
// into io.k8s.networking.gateway.v1.Gateway
// Turn sigs.k8s.io/gateway-api/apisx/v1alpha1.XBackendTrafficPolicy
// into io.x-k8s.networking.gateway.v1alpha1.XBackendTrafficPolicy
func standarizePath(path string) string {
	if remainder, has := strings.CutPrefix(path, "sigs.k8s.io/gateway-api/apis/"); has {
		return "io.k8s.networking.gateway." + remainder
	} else if remainder, has := strings.CutPrefix(path, "sigs.k8s.io/gateway-api/apisx/"); has {
		return "io.x-k8s.networking.gateway." + remainder
	// The following seem to be correct for v1.5.0+, but the fix is required for older versions.
	} else if remainder, has := strings.CutPrefix(path, "k8s.io/apimachinery/pkg/apis/meta/"); has {
		return "io.k8s.apimachinery.pkg.apis.meta." + remainder
	} else if remainder, has := strings.CutPrefix(path, "k8s.io/apimachinery/pkg/"); has {
		return "io.k8s.apimachinery.pkg." + remainder
	}
	return path
}
