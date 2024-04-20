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
	"fmt"
	"log"
	"os"
	"regexp"
	"strings"

	apiext "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"sigs.k8s.io/controller-tools/pkg/crd"
	"sigs.k8s.io/controller-tools/pkg/loader"
	"sigs.k8s.io/controller-tools/pkg/markers"
	"sigs.k8s.io/yaml"

	"sigs.k8s.io/gateway-api/pkg/consts"
)

var standardKinds = map[string]bool{
	"GatewayClass":   true,
	"Gateway":        true,
	"GRPCRoute":      true,
	"HTTPRoute":      true,
	"ReferenceGrant": true,
}

// This generation code is largely copied from
// github.com/kubernetes-sigs/controller-tools/blob/ab52f76cc7d167925b2d5942f24bf22e30f49a02/pkg/crd/gen.go
func main() {
	roots, err := loader.LoadRoots(
		"k8s.io/apimachinery/pkg/runtime/schema", // Needed to parse generated register functions.
		"sigs.k8s.io/gateway-api/apis/v1alpha3",
		"sigs.k8s.io/gateway-api/apis/v1alpha2",
		"sigs.k8s.io/gateway-api/apis/v1beta1",
		"sigs.k8s.io/gateway-api/apis/v1",
	)
	if err != nil {
		log.Fatalf("failed to load package roots: %s", err)
	}

	generator := &crd.Generator{}

	parser := &crd.Parser{
		Collector: &markers.Collector{Registry: &markers.Registry{}},
		Checker: &loader.TypeChecker{
			NodeFilters: []loader.NodeFilter{generator.CheckFilter()},
		},
	}

	err = generator.RegisterMarkers(parser.Collector.Registry)
	if err != nil {
		log.Fatalf("failed to register markers: %s", err)
	}

	crd.AddKnownTypes(parser)
	for _, r := range roots {
		parser.NeedPackage(r)
	}

	metav1Pkg := crd.FindMetav1(roots)
	if metav1Pkg == nil {
		log.Fatalf("no objects in the roots, since nothing imported metav1")
	}

	kubeKinds := crd.FindKubeKinds(parser, metav1Pkg)
	if len(kubeKinds) == 0 {
		log.Fatalf("no objects in the roots")
	}

	channels := []string{"standard", "experimental"}
	for _, channel := range channels {
		for _, groupKind := range kubeKinds {
			if channel == "standard" && !standardKinds[groupKind.Kind] {
				continue
			}

			log.Printf("generating %s CRD for %v\n", channel, groupKind)

			parser.NeedCRDFor(groupKind, nil)
			crdRaw := parser.CustomResourceDefinitions[groupKind]

			// Inline version of "addAttribution(&crdRaw)" ...
			if crdRaw.ObjectMeta.Annotations == nil {
				crdRaw.ObjectMeta.Annotations = map[string]string{}
			}
			crdRaw.ObjectMeta.Annotations[consts.BundleVersionAnnotation] = consts.BundleVersion
			crdRaw.ObjectMeta.Annotations[consts.ChannelAnnotation] = channel
			crdRaw.ObjectMeta.Annotations[apiext.KubeAPIApprovedAnnotation] = consts.ApprovalLink

			// Prevent the top level metadata for the CRD to be generated regardless of the intention in the arguments
			crd.FixTopLevelMetadata(crdRaw)

			channelCrd := crdRaw.DeepCopy()
			for i, version := range channelCrd.Spec.Versions {
				if channel == "standard" && strings.Contains(version.Name, "alpha") {
					channelCrd.Spec.Versions[i].Served = false
				}
				version.Schema.OpenAPIV3Schema.Properties = gatewayTweaks(channel, version.Schema.OpenAPIV3Schema.Properties)
			}

			conv, err := crd.AsVersion(*channelCrd, apiext.SchemeGroupVersion)
			if err != nil {
				log.Fatalf("failed to convert CRD: %s", err)
			}

			out, err := yaml.Marshal(conv)
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
}

// Custom Gateway API Tweaks for tags prefixed with `<gateway:` that get past
// the limitations of Kubebuilder annotations.
func gatewayTweaks(channel string, props map[string]apiext.JSONSchemaProps) map[string]apiext.JSONSchemaProps {
	for name := range props {
		jsonProps, _ := props[name]

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

		if channel == "standard" && strings.Contains(jsonProps.Description, "<gateway:experimental>") {
			delete(props, name)
			continue
		}

		// TODO(robscott): Figure out why crdgen switched this to "object"
		if jsonProps.Format == "date-time" {
			jsonProps.Type = "string"
		}

		validationPrefix := fmt.Sprintf("<gateway:%s:validation:", channel)
		numExpressions := strings.Count(jsonProps.Description, validationPrefix)
		numValid := 0
		if numExpressions > 0 {
			enumRe := regexp.MustCompile(validationPrefix + "Enum=([A-Za-z;]*)>")
			enumMatches := enumRe.FindAllStringSubmatch(jsonProps.Description, 64)
			for _, enumMatch := range enumMatches {
				if len(enumMatch) != 2 {
					log.Fatalf("Invalid %s Enum tag for %s", validationPrefix, name)
				}

				numValid++
				jsonProps.Enum = []apiext.JSON{}
				for _, val := range strings.Split(enumMatch[1], ";") {
					jsonProps.Enum = append(jsonProps.Enum, apiext.JSON{Raw: []byte("\"" + val + "\"")})
				}
			}

			celRe := regexp.MustCompile(validationPrefix + "XValidation:message=\"([^\"]*)\",rule=\"([^\"]*)\">")
			celMatches := celRe.FindAllStringSubmatch(jsonProps.Description, 64)
			for _, celMatch := range celMatches {
				if len(celMatch) != 3 {
					log.Fatalf("Invalid %s CEL tag for %s", validationPrefix, name)
				}

				numValid++
				jsonProps.XValidations = append(jsonProps.XValidations, apiext.ValidationRule{
					Message: celMatch[1],
					Rule:    celMatch[2],
				})
			}
		}
		startTag := "<gateway:experimental:description>"
		endTag := "</gateway:experimental:description>"
		regexPattern := regexp.QuoteMeta(startTag) + `(?s:(.*?))` + regexp.QuoteMeta(endTag)
		if channel == "standard" && strings.Contains(jsonProps.Description, "<gateway:experimental:description>") {
			re := regexp.MustCompile(regexPattern)
			match := re.FindStringSubmatch(jsonProps.Description)
			if len(match) != 2 {
				log.Fatalf("Invalid <gateway:experimental:description> tag for %s", name)
			}
			modifiedDescription := re.ReplaceAllString(jsonProps.Description, "")
			jsonProps.Description = modifiedDescription
		} else {
			jsonProps.Description = strings.ReplaceAll(jsonProps.Description, startTag, "")
			jsonProps.Description = strings.ReplaceAll(jsonProps.Description, endTag, "")
		}

		if numValid < numExpressions {
			fmt.Printf("Description: %s\n", jsonProps.Description)
			log.Fatalf("Found %d Gateway validation expressions, but only %d were valid", numExpressions, numValid)
		}

		gatewayRe := regexp.MustCompile(`<gateway:.*>`)
		jsonProps.Description = gatewayRe.ReplaceAllLiteralString(jsonProps.Description, "")

		if len(jsonProps.Properties) > 0 {
			jsonProps.Properties = gatewayTweaks(channel, jsonProps.Properties)
		} else if jsonProps.Items != nil && jsonProps.Items.Schema != nil {
			jsonProps.Items.Schema.Properties = gatewayTweaks(channel, jsonProps.Items.Schema.Properties)
		}
		props[name] = jsonProps
	}
	return props
}
