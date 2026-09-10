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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"sigs.k8s.io/gateway-api/pkg/consts"
)

func TestMarshalCRDManifestOmitsTopLevelStatus(t *testing.T) {
	t.Parallel()

	crd := apiext.CustomResourceDefinition{
		TypeMeta: metav1.TypeMeta{
			APIVersion: apiext.SchemeGroupVersion.String(),
			Kind:       "CustomResourceDefinition",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: "httproutes.gateway.networking.k8s.io",
		},
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
