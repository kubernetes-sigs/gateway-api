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
	"fmt"
	"strings"
	"testing"

	apiext "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"sigs.k8s.io/controller-tools/pkg/crd"
	"sigs.k8s.io/controller-tools/pkg/loader"
	"sigs.k8s.io/controller-tools/pkg/markers"

	"sigs.k8s.io/gateway-api/pkg/consts"
)

func TestMarshalCRDManifestOmitsTopLevelStatus(t *testing.T) {
	t.Parallel()

	crd := apiext.CustomResourceDefinition{
		APIVersion: apiext.SchemeGroupVersion.String(),
		Kind:       "CustomResourceDefinition",
		Name:       "httproutes.gateway.networking.k8s.io",
		Spec: apiext.CustomResourceDefinitionSpec{
			Group: "gateway.networking.k8s.io",
			Names: apiext.CustomResourceDefinitionNames{
				Plural: "httproutes",
				Kind:   "HTTPRoute",
			},
			Scope: "Namespaced",
		},
		Status: apiext.CustomResourceDefinitionStatus{
			Conditions:     nil,
			StoredVersions: nil,
		},
	}

	out, err := marshalCRDManifest(crd)
	if err != nil {
		t.Fatalf("marshalCRDManifest returned error: %v", err)
	}

	yaml := string(out)
	if strings.Contains(yaml, "\nstatus:\n") || strings.HasPrefix(yaml, "status:\n") {
		t.Fatalf("expected top-level status to be omitted, got:\n%s", yaml)
	}
	if strings.Contains(yaml, "\nTypeMeta:\n") || strings.HasPrefix(yaml, "TypeMeta:\n") {
		t.Fatalf("expected type metadata to stay inlined, got:\n%s", yaml)
	}
	if !strings.Contains(yaml, "apiVersion: apiextensions.k8s.io/v1\nkind: CustomResourceDefinition\nmetadata:\n") {
		t.Fatalf("expected top-level apiVersion/kind metadata, got:\n%s", yaml)
	}
	if !strings.Contains(yaml, "\nspec:\n") {
		t.Fatalf("expected spec to be present, got:\n%s", yaml)
	}
}

func TestGatewayValidationItemsMarkerGeneratesSchema(t *testing.T) {
	parser, pkg := newValidationTestParser(t)
	info := parser.LookupType(pkg, "HTTPRouteSpec")
	if info == nil {
		t.Fatal("HTTPRouteSpec was not loaded")
	}
	for i := range info.Fields {
		if info.Fields[i].Name == "Hostnames" {
			info.Fields[i].Doc += "\n<gateway:experimental:validation:items:MaxLength=88>"
		}
	}

	if err := applyGatewayTypeValidations(parser, "experimental"); err != nil {
		t.Fatalf("apply Gateway validations: %v", err)
	}
	ident := crd.TypeIdent{Package: pkg, Name: "HTTPRouteSpec"}
	schema := parser.Schemata[ident]
	hostnames := schema.Properties["hostnames"]
	if hostnames.Items == nil || hostnames.Items.Schema == nil {
		t.Fatal("expected hostnames item schema")
	}
	if got := hostnames.Items.Schema.MaxLength; got == nil || *got != 88 {
		t.Fatalf("expected hostnames item maxLength 88, got %v", got)
	}
}

func TestGatewayValidationTypeMarkerGeneratesSchemaForItsChannel(t *testing.T) {
	for _, tc := range []struct {
		channel string
		want    int64
	}{
		{channel: "standard", want: 77},
		{channel: "experimental", want: 88},
	} {
		t.Run(tc.channel, func(t *testing.T) {
			parser, pkg := newValidationTestParser(t)
			info := parser.LookupType(pkg, "Hostname")
			if info == nil {
				t.Fatal("Hostname was not loaded")
			}
			info.Doc += "\n<gateway:standard:validation:MaxLength=77>" +
				"\n<gateway:experimental:validation:MaxLength=88>"

			if err := applyGatewayTypeValidations(parser, tc.channel); err != nil {
				t.Fatalf("apply Gateway validations: %v", err)
			}
			schema := parser.Schemata[crd.TypeIdent{Package: pkg, Name: "Hostname"}]
			if got := schema.MaxLength; got == nil || *got != tc.want {
				t.Fatalf("expected %s channel maxLength %d, got %v", tc.channel, tc.want, got)
			}
		})
	}
}

// The level-up tag shares the channel marker prefix but is not a kubebuilder marker,
// so the marker pass must leave it to the schema pass instead of rejecting it.
func TestGatewayValidationIgnoresLevelUpTag(t *testing.T) {
	parser, pkg := newValidationTestParser(t)
	info := parser.LookupType(pkg, "HTTPRouteSpec")
	if info == nil {
		t.Fatal("HTTPRouteSpec was not loaded")
	}
	for i := range info.Fields {
		if info.Fields[i].Name == "Rules" {
			info.Fields[i].Doc += "\n<gateway:experimental:validation:levelUpValidation=\"path.value\">" +
				"\n<gateway:experimental:validation:MaxItems=99>"
		}
	}

	if err := applyGatewayTypeValidations(parser, "experimental"); err != nil {
		t.Fatalf("apply Gateway validations: %v", err)
	}
	schema := parser.Schemata[crd.TypeIdent{Package: pkg, Name: "HTTPRouteSpec"}]
	if got := schema.Properties["rules"].MaxItems; got == nil || *got != 99 {
		t.Fatalf("expected rules maxItems 99, got %v", got)
	}
}

func TestGatewayExactlyOneOfGeneratesTypeValidation(t *testing.T) {
	parser, pkg := newValidationTestParser(t)
	info := parser.LookupType(pkg, "GatewaySpecAddress")
	if info == nil {
		t.Fatal("GatewaySpecAddress was not loaded")
	}
	info.Doc += "\n<gateway:experimental:validation:ExactlyOneOf=type;value>"

	if err := applyGatewayTypeValidations(parser, "experimental"); err != nil {
		t.Fatalf("apply Gateway validations: %v", err)
	}
	schema := parser.Schemata[crd.TypeIdent{Package: pkg, Name: "GatewaySpecAddress"}]
	for _, validation := range schema.XValidations {
		if validation.Message == "exactly one of the fields in [type value] must be set" &&
			strings.Contains(validation.Rule, "has(self.type)") && strings.Contains(validation.Rule, "has(self.value)") {
			return
		}
	}
	t.Fatalf("expected ExactlyOneOf validation in schema, got %+v", schema.XValidations)
}

func TestGatewayValidationRejectsUnterminatedTag(t *testing.T) {
	parser, pkg := newValidationTestParser(t)
	info := parser.LookupType(pkg, "Hostname")
	if info == nil {
		t.Fatal("Hostname was not loaded")
	}
	info.Doc += "\n<gateway:experimental:validation:MaxLength=88"

	if err := applyGatewayTypeValidations(parser, "experimental"); err == nil || !strings.Contains(err.Error(), "unterminated validation tag") {
		t.Fatalf("expected unterminated validation tag error, got %v", err)
	}
}

func newValidationTestParser(t *testing.T) (*crd.Parser, *loader.Package) {
	t.Helper()
	roots, err := loader.LoadRoots("sigs.k8s.io/gateway-api/apis/v1")
	if err != nil {
		t.Fatalf("load API package: %v", err)
	}
	generator := &crd.Generator{}
	parser := &crd.Parser{
		Collector: &markers.Collector{Registry: &markers.Registry{}},
		Checker: &loader.TypeChecker{
			NodeFilters: []loader.NodeFilter{generator.CheckFilter()},
		},
	}
	if err := generator.RegisterMarkers(parser.Collector.Registry); err != nil {
		t.Fatalf("register CRD markers: %v", err)
	}
	crd.AddKnownTypes(parser)
	parser.NeedPackage(roots[0])
	return parser, roots[0]
}

// vapFixture mimics the structure of the hand-maintained VAP manifests: the
// bundle-version annotation appears on annotation lines in two documents and
// inside a CEL expression, and the standard-channel prohibition pattern
// matches versions up to v1.4.x.
const vapFixture = `apiVersion: admissionregistration.k8s.io/v1
kind: ValidatingAdmissionPolicy
metadata:
  annotations:
    gateway.networking.k8s.io/bundle-version: v0.0.0-dev
  name: safe-upgrades.gateway.networking.k8s.io
spec:
  validations:
    - expression: "!matches(object.metadata.annotations['gateway.networking.k8s.io/bundle-version'], 'v1.[0-4].\\\\d+')"
      message: "Installing CRDs matching v1.[0-4]. is prohibited."
---
apiVersion: admissionregistration.k8s.io/v1
kind: ValidatingAdmissionPolicyBinding
metadata:
  annotations:
    gateway.networking.k8s.io/bundle-version: v0.0.0-dev
  name: safe-upgrades.gateway.networking.k8s.io
`

// The tweaks strip every <gateway:...> tag from the descriptions and replace the
// schemas they walk, so tags must be collected before gatewayTweaksMap and applied to
// the schema it returns.
func TestLevelUpTagsSurviveDescriptionStripping(t *testing.T) {
	t.Parallel()

	maxItems := int64(2)
	root := &apiext.JSONSchemaProps{Properties: map[string]apiext.JSONSchemaProps{
		"spec": {Properties: map[string]apiext.JSONSchemaProps{
			"rules": {Type: "array", Items: &apiext.JSONSchemaPropsOrArray{Schema: &apiext.JSONSchemaProps{
				Type: "object",
				Properties: map[string]apiext.JSONSchemaProps{
					"matches": {
						Type:        "array",
						MaxItems:    &maxItems,
						Description: "Matches are evaluated in order.\n\n\n<gateway:validation:levelUpValidation=\"path.value\">",
						Items: &apiext.JSONSchemaPropsOrArray{Schema: &apiext.JSONSchemaProps{
							Type: "object",
							Properties: map[string]apiext.JSONSchemaProps{
								"path": {Type: "object", XValidations: apiext.ValidationRules{{
									Rule:    "self.type == 'Exact' ? self.value.matches('^/[a-z]+$') : true",
									Message: "invalid path",
								}}},
							},
						}},
					},
				},
			}}},
		}},
	}}

	targets, err := collectLevelUpTargets(root, "standard")
	if err != nil {
		t.Fatalf("collecting tags: %v", err)
	}
	if len(targets) != 1 {
		t.Fatalf("expected 1 target, got %d", len(targets))
	}

	root.Properties = gatewayTweaksMap("standard", root.Properties)
	if err := applyLevelUpTargets(root, targets); err != nil {
		t.Fatalf("applying tags: %v", err)
	}

	rule := root.Properties["spec"].Properties["rules"].Items.Schema
	if description := rule.Properties["matches"].Description; strings.Contains(description, "<gateway:") {
		t.Fatalf("tweaks left a tag in the description: %q", description)
	}
	if len(rule.XValidations) != 1 {
		t.Fatalf("expected 1 generated validation, got %d", len(rule.XValidations))
	}
	if path := rule.Properties["matches"].Items.Schema.Properties["path"]; len(path.XValidations) != 0 {
		t.Fatalf("expected the leaf validation to move, got %d", len(path.XValidations))
	}
}

func TestUpdateVAPManifestStandardChannel(t *testing.T) {
	t.Parallel()

	manifest, err := updateVAPManifest(vapFixture, "standard", "v1.6.0")
	if err != nil {
		t.Fatalf("updateVAPManifest returned error: %v", err)
	}

	annotationLine := fmt.Sprintf("%s: v1.6.0", consts.BundleVersionAnnotation)
	if got := strings.Count(manifest, annotationLine); got != 2 {
		t.Errorf("expected annotation to be updated in both documents, found %d occurrences of %q in:\n%s", got, annotationLine, manifest)
	}
	if strings.Contains(manifest, "v0.0.0-dev") {
		t.Errorf("expected all placeholder versions to be replaced, got:\n%s", manifest)
	}
	if strings.Contains(manifest, "v1.[0-4].") {
		t.Errorf("expected prohibition range v1.[0-4]. to be bumped, got:\n%s", manifest)
	}
	if got := strings.Count(manifest, "v1.[0-5]."); got != 2 {
		t.Errorf("expected prohibition range v1.[0-5]. in expression and message, found %d occurrences in:\n%s", got, manifest)
	}
}

func TestUpdateVAPManifestExperimentalChannelLeavesProhibitionUntouched(t *testing.T) {
	t.Parallel()

	manifest, err := updateVAPManifest(vapFixture, "experimental", "v1.6.0")
	if err != nil {
		t.Fatalf("updateVAPManifest returned error: %v", err)
	}

	annotationLine := fmt.Sprintf("%s: v1.6.0", consts.BundleVersionAnnotation)
	if got := strings.Count(manifest, annotationLine); got != 2 {
		t.Errorf("expected annotation to be updated in both documents, found %d occurrences of %q in:\n%s", got, annotationLine, manifest)
	}
	if got := strings.Count(manifest, "v1.[0-4]."); got != 2 {
		t.Errorf("expected prohibition range to be left untouched on experimental channel, found %d occurrences of v1.[0-4]. in:\n%s", got, manifest)
	}
}

func TestUpdateVAPManifestDoesNotTouchCELAnnotationKey(t *testing.T) {
	t.Parallel()

	manifest, err := updateVAPManifest(vapFixture, "standard", "v1.6.0")
	if err != nil {
		t.Fatalf("updateVAPManifest returned error: %v", err)
	}

	// The annotation key inside the CEL expression must survive as a map
	// lookup, not be rewritten like an annotation line.
	celLookup := fmt.Sprintf("object.metadata.annotations['%s']", consts.BundleVersionAnnotation)
	if !strings.Contains(manifest, celLookup) {
		t.Errorf("expected CEL annotation lookup %q to be preserved, got:\n%s", celLookup, manifest)
	}
}

func TestUpdateVAPManifestSkipsDevBundleVersion(t *testing.T) {
	t.Parallel()

	manifest, err := updateVAPManifest(vapFixture, "standard", "v0.0.0-dev")
	if err != nil {
		t.Fatalf("expected dev bundle version to be skipped, got error: %v", err)
	}
	if manifest != vapFixture {
		t.Errorf("expected manifest to be unchanged for dev bundle version, got:\n%s", manifest)
	}
}
