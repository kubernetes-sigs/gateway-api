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

package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"regexp"
	"strconv"
	"strings"

	"golang.org/x/tools/go/packages"
	apiext "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-tools/pkg/crd"
	crdmarkers "sigs.k8s.io/controller-tools/pkg/crd/markers"
	"sigs.k8s.io/controller-tools/pkg/loader"
	"sigs.k8s.io/controller-tools/pkg/markers"
	"sigs.k8s.io/yaml"

	"sigs.k8s.io/gateway-api/pkg/consts"
)

type customResourceDefinitionManifest struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitzero"`
	Spec              apiext.CustomResourceDefinitionSpec `json:"spec"`
}

var experimentalValidationRegistry = func() *markers.Registry {
	registry := &markers.Registry{}
	if err := crdmarkers.Register(registry); err != nil {
		panic(err)
	}
	return registry
}()

var standardKinds = map[string]bool{
	"GatewayClass":     true,
	"Gateway":          true,
	"GRPCRoute":        true,
	"HTTPRoute":        true,
	"ReferenceGrant":   true,
	"BackendTLSPolicy": true,
	"ListenerSet":      true,
	"TCPRoute":         true,
	"TLSRoute":         true,
	"UDPRoute":         true,
}

// This generation code is largely copied from
// github.com/kubernetes-sigs/controller-tools/blob/ab52f76cc7d167925b2d5942f24bf22e30f49a02/pkg/crd/gen.go
func main() {
	channels := []string{"standard", "experimental"}
	bundleVersion := consts.BundleVersion

	experimentalOnly := flag.Bool("experimental-only", false, "Generate only experimental CRDs")
	versionFlag := flag.String("version", bundleVersion, "The bundle version to set in the CRD annotations")

	flag.Parse()

	if *experimentalOnly {
		channels = []string{"experimental"}
	}

	if *versionFlag != "" {
		bundleVersion = *versionFlag
	}

	roots, err := loader.LoadRoots(
		"k8s.io/apimachinery/pkg/runtime/schema", // Needed to parse generated register functions.
		"sigs.k8s.io/gateway-api/apis/v1alpha3",
		"sigs.k8s.io/gateway-api/apis/v1alpha2",
		"sigs.k8s.io/gateway-api/apis/v1beta1",
		"sigs.k8s.io/gateway-api/apis/v1",
		"sigs.k8s.io/gateway-api/apisx/v1alpha1",
	)
	if err != nil {
		log.Fatalf("failed to load package roots: %s", err)
	}

	generator := &crd.Generator{}
	newParser := func() *crd.Parser {
		parser := &crd.Parser{
			Collector: &markers.Collector{Registry: &markers.Registry{}},
			Checker: &loader.TypeChecker{
				NodeFilters: []loader.NodeFilter{generator.CheckFilter()},
			},
		}
		if err := generator.RegisterMarkers(parser.Collector.Registry); err != nil {
			log.Fatalf("failed to register markers: %s", err)
		}
		crd.AddKnownTypes(parser)
		for _, r := range roots {
			parser.NeedPackage(r)
		}
		return parser
	}

	parser := newParser()

	metav1Pkg := crd.FindMetav1(roots)
	if metav1Pkg == nil {
		log.Fatalf("no objects in the roots, since nothing imported metav1")
	}

	kubeKinds := crd.FindKubeKinds(parser, metav1Pkg)
	if len(kubeKinds) == 0 {
		log.Fatalf("no objects in the roots")
	}

	for _, channel := range channels {
		parser = newParser()
		if err := applyGatewayTypeValidations(parser, channel); err != nil {
			log.Fatalf("failed to apply Gateway validation markers: %s", err)
		}
		for _, groupKind := range kubeKinds {
			if channel == "standard" && !standardKinds[groupKind.Kind] {
				continue
			}

			log.Printf("generating %s %s CRD for %v\n", channel, bundleVersion, groupKind)

			parser.NeedCRDFor(groupKind, nil)
			crdRaw := parser.CustomResourceDefinitions[groupKind]

			// Inline version of "addAttribution(&crdRaw)" ...
			if crdRaw.Annotations == nil {
				crdRaw.Annotations = map[string]string{}
			}
			crdRaw.Annotations[consts.BundleVersionAnnotation] = bundleVersion
			crdRaw.Annotations[consts.ChannelAnnotation] = channel
			crdRaw.Annotations[apiext.KubeAPIApprovedAnnotation] = consts.ApprovalLink

			// Prevent the top level metadata for the CRD to be generated regardless of the intention in the arguments
			crd.FixTopLevelMetadata(crdRaw)

			channelCrd := crdRaw.DeepCopy()
			for i, version := range channelCrd.Spec.Versions {
				if channel == "standard" && strings.Contains(version.Name, "alpha") {
					channelCrd.Spec.Versions[i].Served = false
				}
				version.Schema.OpenAPIV3Schema.Properties = gatewayTweaksMap(channel, version.Schema.OpenAPIV3Schema.Properties)
			}

			convObj, err := crd.AsVersion(*channelCrd, apiext.SchemeGroupVersion)
			if err != nil {
				log.Fatalf("failed to convert CRD: %s", err)
			}

			conv, ok := convObj.(*apiext.CustomResourceDefinition)
			if !ok {
				log.Fatalf("failed to convert CRD: unexpected type %T", convObj)
			}

			out, err := marshalCRDManifest(*conv)
			if err != nil {
				log.Fatalf("failed to marshal CRD: %s", err)
			}

			fileName := fmt.Sprintf("config/crd/%s/%s_%s.yaml", channel, crdRaw.Spec.Group, crdRaw.Spec.Names.Plural)
			err = os.WriteFile(fileName, out, 0o600)
			if err != nil {
				log.Fatalf("failed to write CRD: %s", err)
			}
		}
	}

	for _, channel := range channels {
		err := updateVAP(channel, bundleVersion)
		if err != nil {
			log.Fatalf("failed to update vap: %s", err)
		}
	}

	if loader.PrintErrors(roots, packages.TypeError) {
		log.Fatalf("not all generators ran successfully")
	}
}

func applyGatewayTypeValidations(parser *crd.Parser, channel string) error {
	prefix := fmt.Sprintf("<gateway:%s:validation:", channel)
	for _, info := range parser.Types {
		if info.Markers == nil {
			info.Markers = markers.MarkerValues{}
		}
		values, err := experimentalValidationMarkerValues(info.Doc, prefix, markers.DescribesType)
		if err != nil {
			return fmt.Errorf("type %s: %w", info.Name, err)
		}
		for name, values := range values {
			info.Markers[name] = append(info.Markers[name], values...)
		}
		for i := range info.Fields {
			if info.Fields[i].Markers == nil {
				info.Fields[i].Markers = markers.MarkerValues{}
			}
			values, err := experimentalValidationMarkerValues(info.Fields[i].Doc, prefix, markers.DescribesField)
			if err != nil {
				return fmt.Errorf("type %s field %s: %w", info.Name, info.Fields[i].Name, err)
			}
			for name, values := range values {
				info.Fields[i].Markers[name] = append(info.Fields[i].Markers[name], values...)
			}
		}
	}

	for ident, info := range parser.Types {
		if strings.Contains(info.Doc, prefix) {
			parser.NeedSchemaFor(ident)
			continue
		}
		for _, field := range info.Fields {
			if strings.Contains(field.Doc, prefix) {
				parser.NeedSchemaFor(ident)
				break
			}
		}
	}
	return nil
}

// updateVAP updates the hand-maintained ValidatingAdmissionPolicy manifest
// for the given channel.
func updateVAP(channel, bundleVersion string) error {
	path := fmt.Sprintf("config/crd/%s/gateway.networking.k8s.io_vap_safeupgrades.yaml", channel)

	manifest, err := readVAPManifest(path)
	if err != nil {
		return err
	}

	updated, err := updateVAPManifest(manifest, channel, bundleVersion)
	if err != nil {
		return fmt.Errorf("failed to update VAP manifest %s: %w", path, err)
	}
	if updated == manifest {
		// Nothing changed, e.g. when running on main with the `v0.0.0-dev`
		// bundle version.
		return nil
	}

	if err := writeVAPManifest(path, updated); err != nil {
		return err
	}

	log.Printf("updated %s %s to %s\n", path, consts.BundleVersionAnnotation, bundleVersion)
	return nil
}

func readVAPManifest(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("failed to read VAP manifest %s: %w", path, err)
	}
	return string(data), nil
}

// updateVAPManifest stamps the bundle version into the manifest's annotations
// and, for the standard channel, bumps the prohibited-version range.
// The manifest is edited textually rather than round-tripped through a YAML
// marshal to preserve its formatting, comments, and multi-document structure.
func updateVAPManifest(manifest, channel, bundleVersion string) (string, error) {
	versionMatch := regexp.MustCompile(`^v(\d+)\.(\d+)`).FindStringSubmatch(bundleVersion)
	if versionMatch == nil {
		return "", fmt.Errorf("bundle version %q is not of the form vMAJOR.MINOR", bundleVersion)
	}
	minor, err := strconv.Atoi(versionMatch[2])
	if err != nil {
		return "", fmt.Errorf("invalid minor version in bundle version %q: %w", bundleVersion, err)
	}
	if minor < 1 {
		// Skip if this runs on main or with the `v0.0.0-dev` bundle version
		return manifest, nil
	}

	// Only match annotation lines; the annotation key also appears inside CEL
	// expressions, where it is never at the start of a line.
	re := regexp.MustCompile(`(?m)^(\s*` + regexp.QuoteMeta(consts.BundleVersionAnnotation) + `:\s*).*$`)
	manifest = re.ReplaceAllString(manifest, "${1}"+bundleVersion)

	if channel == "standard" {
		previousMinor := fmt.Sprintf("v1.[0-%d].", minor-2)
		latestMinor := fmt.Sprintf("v1.[0-%d].", minor-1)
		log.Printf("updating prohibitions from version %s to %s\n", previousMinor, latestMinor)

		// Prohibit installing bundle versions older than the previous minor
		// version, e.g. anything matching v1.[0-5]. when generating v1.6.x.
		manifest = strings.ReplaceAll(manifest, previousMinor, latestMinor)
	}

	return manifest, nil
}

func writeVAPManifest(path, manifest string) error {
	if err := os.WriteFile(path, []byte(manifest), 0o600); err != nil { // #nosec G703 -- path is derived from a hardcoded channel list
		return fmt.Errorf("failed to write VAP manifest %s: %w", path, err)
	}
	return nil
}

func marshalCRDManifest(customResourceDefinition apiext.CustomResourceDefinition) ([]byte, error) {
	manifest := customResourceDefinitionManifest{
		TypeMeta:   customResourceDefinition.TypeMeta,
		ObjectMeta: customResourceDefinition.ObjectMeta,
		Spec:       customResourceDefinition.Spec,
	}
	return yaml.Marshal(manifest)
}

func gatewayTweaksMap(channel string, props map[string]apiext.JSONSchemaProps) map[string]apiext.JSONSchemaProps {
	for name := range props {
		jsonProps := props[name]
		p := gatewayTweaks(channel, name, jsonProps)
		if p == nil {
			delete(props, name)
		} else {
			props[name] = *p
		}
	}
	return props
}

// Custom Gateway API Tweaks for tags prefixed with `<gateway:` that get past
// the limitations of Kubebuilder annotations.
func gatewayTweaks(channel string, name string, jsonProps apiext.JSONSchemaProps) *apiext.JSONSchemaProps {
	if strings.Contains(jsonProps.Description, "<gateway:validateIPAddress>") {
		jsonProps.Items.Schema.OneOf = []apiext.JSONSchemaProps{{
			Properties: map[string]apiext.JSONSchemaProps{
				"type": {
					Enum: []apiext.JSON{{Raw: []byte("\"IPAddress\"")}},
				},
				"value": {
					AnyOf: []apiext.JSONSchemaProps{{
						Format: "ipv4",
					}, {
						Format: "ipv6",
					}},
				},
			},
		}, {
			Properties: map[string]apiext.JSONSchemaProps{
				"type": {
					Not: &apiext.JSONSchemaProps{
						Enum: []apiext.JSON{{Raw: []byte("\"IPAddress\"")}},
					},
				},
			},
		}}
	}

	if channel == "standard" {
		if strings.Contains(jsonProps.Description, "<gateway:experimental>") {
			return nil
		}
	}

	// TODO(robscott): Figure out why crdgen switched this to "object"
	if jsonProps.Format == "date-time" {
		jsonProps.Type = "string"
	}

	jsonProps.Description = formatDescription(jsonProps.Description, channel, name)

	if len(jsonProps.Properties) > 0 {
		jsonProps.Properties = gatewayTweaksMap(channel, jsonProps.Properties)
	} else if jsonProps.Items != nil && jsonProps.Items.Schema != nil {
		jsonProps.Items.Schema = gatewayTweaks(channel, name, *jsonProps.Items.Schema)
	}

	return &jsonProps
}

func experimentalValidationMarkerValues(description, prefix string, target markers.TargetType) (markers.MarkerValues, error) {
	values := markers.MarkerValues{}
	markersInDescription, err := experimentalValidationMarkers(description, prefix)
	if err != nil {
		return nil, err
	}
	for _, marker := range markersInDescription {
		rawMarker := "+kubebuilder:validation:" + marker
		definition := experimentalValidationRegistry.Lookup(rawMarker, target)
		if definition == nil {
			return nil, fmt.Errorf("unsupported %s marker %q", prefix, marker)
		}

		value, err := definition.Parse(rawMarker)
		if err != nil {
			return nil, fmt.Errorf("invalid %s marker %q: %w", prefix, marker, err)
		}
		values[definition.Name] = append(values[definition.Name], value)
	}
	return values, nil
}

func experimentalValidationMarkers(description, prefix string) ([]string, error) {
	var result []string
	for offset := 0; offset < len(description); {
		start := strings.Index(description[offset:], prefix)
		if start < 0 {
			break
		}
		start += offset
		end := start + len(prefix)
		quoted := byte(0)
		escaped := false
		closed := false
		for ; end < len(description); end++ {
			c := description[end]
			if escaped {
				escaped = false
				continue
			}
			if c == '\\' && quoted == '"' {
				escaped = true
				continue
			}
			if c == '`' || c == '"' {
				switch quoted {
				case 0:
					quoted = c
				case c:
					quoted = 0
				}
				continue
			}
			if c == '>' && quoted == 0 {
				marker := description[start+len(prefix) : end]
				result = append(result, marker)
				offset = end + 1
				closed = true
				break
			}
		}
		if !closed {
			return nil, fmt.Errorf("unterminated validation tag %q", description[start:])
		}
	}
	return result, nil
}

func formatDescription(description string, channel string, name string) string {
	startTag := "<gateway:experimental:description>"
	endTag := "</gateway:experimental:description>"
	if channel == "standard" && strings.Contains(description, "<gateway:experimental:description>") {
		regexPattern := `\n*` + regexp.QuoteMeta(startTag) + `(?s:(.*?))` + regexp.QuoteMeta(endTag) + `\n*`
		re := regexp.MustCompile(regexPattern)
		match := re.FindStringSubmatch(description)
		if len(match) != 2 {
			log.Fatalf("Invalid <gateway:experimental:description> tag for %s", name)
		}
		description = re.ReplaceAllString(description, "\n\n")
	} else {
		description = strings.ReplaceAll(description, startTag, "")
		description = strings.ReplaceAll(description, endTag, "")
	}

	// Comments within "gateway:util:excludeFromCRD" tag is not included in the generated CRD and all trailing \n operators before
	// and after the tags are removed and replaced with three \n operators.
	startTag = "<gateway:util:excludeFromCRD>"
	endTag = "</gateway:util:excludeFromCRD>"
	if strings.Contains(description, "<gateway:util:excludeFromCRD>") {
		regexPattern := `\n*` + regexp.QuoteMeta(startTag) + `(?s:(.*?))` + regexp.QuoteMeta(endTag) + `\n*`
		re := regexp.MustCompile(regexPattern)
		match := re.FindStringSubmatch(description)
		if len(match) != 2 {
			log.Fatalf("Invalid <gateway:util:excludeFromCRD> tag for %s", name)
		}
		description = re.ReplaceAllString(description, "\n\n\n")
	}

	gatewayRe := regexp.MustCompile(`<gateway:.*>`)
	description = gatewayRe.ReplaceAllLiteralString(description, "")

	// Remove any extra \n (more than 3 and all trailing at the end)
	regexPattern := `\n\n\n+`
	re := regexp.MustCompile(regexPattern)
	description = re.ReplaceAllString(description, "\n\n\n")
	description = strings.Trim(description, "\n")

	return description
}
