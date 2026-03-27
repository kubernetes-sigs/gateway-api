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
	"strings"
	"testing"

	apiext "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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
