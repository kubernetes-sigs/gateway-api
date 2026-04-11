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
	"encoding/json"
	"errors"
	"io"
	"os"

	extensionv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/yaml"
)

func loadCrdFile(filename string) ([]*extensionv1.CustomResourceDefinition, error) {
	var crds []*extensionv1.CustomResourceDefinition
	var file *os.File
	if filename == "-" {
		file = os.Stdin
	} else {
		var err error
		file, err = os.Open(filename)
		if err != nil {
			return nil, err
		}
		defer func() { _ = file.Close() }()
	}

	decode := yaml.NewYAMLOrJSONDecoder(file, 4096)
	for {
		raw := &runtime.RawExtension{}
		crd := &extensionv1.CustomResourceDefinition{}
		err := decode.Decode(raw)
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return nil, err
		}
		err = json.Unmarshal(raw.Raw, crd)
		if err != nil {
			return nil, err
		}

		crds = append(crds, crd)
	}
	return crds, nil
}

func loadCrdFiles(filenames []string) ([]*extensionv1.CustomResourceDefinition, error) {
	var crds []*extensionv1.CustomResourceDefinition
	for _, filename := range filenames {
		newCrds, err := loadCrdFile(filename)
		if err != nil {
			return nil, err
		}
		crds = append(crds, newCrds...)
	}
	return crds, nil
}
