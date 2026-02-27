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

/*
Take input CRD definitions and output OpenAPI v2 spec. This will work
with any CRDs, not just Gateway API ones.
*/

package main

import (
	"fmt"
	"os"

	"k8s.io/apiextensions-apiserver/pkg/controller/openapi/builder"

	"github.com/spf13/cobra"
)

func main() {
	err := newCommand().Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to generate spec: %v\n", err)
		os.Exit(1)
	}
}

func run(_ *cobra.Command, args []string) error {
	crds, err := loadCrdFiles(args)
	if err != nil {
		return err
	}
	if len(crds) == 0 {
		return fmt.Errorf("no CRDs found")
	}

	staticSpec := createStaticSpec(name, version)

	specs, err := convertFromCrds(crds)
	if err != nil {
		return err
	}

	mergeSpecs, err := builder.MergeSpecs(staticSpec, specs...)
	if err != nil {
		return err
	}

	json, err := mergeSpecs.MarshalJSON()
	if err != nil {
		return err
	}

	var writer *os.File
	if output == "-" {
		writer = os.Stdout
	} else {
		writer, err = os.Create(output)
		if err != nil {
			return err
		}
	}

	_, err = writer.Write(json)
	if err != nil {
		return err
	}
	writer.Close()
	return nil
}
