/*
Copyright 2020 The Kubernetes Authors.

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
package examples

import (
	"io"
	"os"
	"path"
	"path/filepath"
	"testing"

	"sigs.k8s.io/service-apis/apis/v1alpha1"

	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/apimachinery/pkg/util/yaml"
	"k8s.io/client-go/kubernetes/scheme"
)

func findExamples(t *testing.T, dirPath string) []string {
	var filenames []string

	t.Helper()

	for _, pattern := range []string{
		path.Join(dirPath, "*.yaml"),
		path.Join(dirPath, "*.yml"),
	} {
		matches, err := filepath.Glob(pattern)
		if err != nil {
			t.Fatalf("glob %q failed: %s", pattern, err)
		}

		filenames = append(filenames, matches...)
	}

	return filenames
}

func decodeExample(t *testing.T, filename string) [][]byte {
	parts := [][]byte{}

	t.Helper()

	f, err := os.Open(filename)
	if err != nil {
		t.Fatal(err)
	}

	defer f.Close()

	splitter := yaml.NewDocumentDecoder(f)

	for {
		buf := make([]byte, 4096)
		nread, err := splitter.Read(buf)
		switch err {
		case nil:
			parts = append(parts, buf[:nread])
		case io.EOF:
			return parts
		default:
			t.Fatalf("failed to read YAML from %q: %s", filename, err)
		}
	}
}

func objectName(o metav1.Object) string {
	n := o.GetName()
	ns := o.GetNamespace()

	if ns == "" {
		ns = metav1.NamespaceDefault
	}

	return ns + "/" + n
}

// TestParseExamples finds all the YAML documents in the examples
// directory and ensures that they contain Kubernetes objects that
// can be parsed by the service-apis scheme.
func TestParseExamples(t *testing.T) {
	s := runtime.NewScheme()

	if err := scheme.AddToScheme(s); err != nil {
		t.Fatalf("failed to add builtin scheme: %s", err)
	}

	if err := v1alpha1.AddToScheme(s); err != nil {
		t.Fatalf("failed to add service-api scheme: %s", err)
	}

	decoder := serializer.NewCodecFactory(s).UniversalDeserializer()

	for _, filename := range findExamples(t, ".") {
		t.Run(filename, func(t *testing.T) {
			for n, buf := range decodeExample(t, filename) {
				obj, vers, err := decoder.Decode(buf, nil, nil)
				if err != nil {
					t.Errorf("failed to decode YAML object #%d from %q: %s", n, filename, err)
					continue
				}

				metaObj, err := meta.Accessor(obj)
				if err != nil {
					t.Errorf("invalid type for decoded object: %s", err)
					continue
				}

				t.Logf("decoded YAML object #%d as name=%q group=%q version=%q kind=%q\n",
					n, objectName(metaObj), vers.Group, vers.Version, vers.Kind)
			}
		})
	}
}
