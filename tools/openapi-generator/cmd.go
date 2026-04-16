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
	"github.com/spf13/cobra"
)

var (
	name           string
	version        string
	output         string
	gatewayAPIDefs bool
)

func newCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "openapi-generator [crd file]...",
		Short: "Convert CRD definitions to an OpenAPI v2 specification",
		Long: `Read CRD definitions from the specified input files and generate an
OpenAPI/Swagger v2 spec file in JSON format.`,
		Args: cobra.MinimumNArgs(1),
		RunE: run,
	}

	flags := cmd.Flags()

	flags.StringVarP(&name, "name", "n", "undefined", "Name of the API in the output spec")
	flags.StringVarP(&version, "version", "v", "undefined", "Version of the API in the output spec")
	flags.StringVarP(&output, "output", "o", "-", "Output file")
	flags.BoolVar(&gatewayAPIDefs, "add-gateway-api-object-defs", false, "Add the non-top level Gateway API objects to the spec")

	return cmd
}
