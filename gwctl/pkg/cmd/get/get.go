/*
Copyright 2023 The Kubernetes Authors.

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

package get

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"sigs.k8s.io/gateway-api/gwctl/pkg/cmd/utils"
	"sigs.k8s.io/gateway-api/gwctl/pkg/cmd/utils/printer"
	"sigs.k8s.io/gateway-api/gwctl/pkg/common/resourcehelpers"
	"sigs.k8s.io/gateway-api/gwctl/pkg/effectivepolicy"
)

type getFlags struct {
	namespace     string
	allNamespaces bool
}

func NewGetCommand(params *utils.CmdParams) *cobra.Command {
	flags := &getFlags{}

	cmd := &cobra.Command{
		Use:   "get {policies|policycrds|httproutes}",
		Short: "Display one or many resources",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			runGet(args, params, flags)
		},
	}
	cmd.Flags().StringVarP(&flags.namespace, "namespace", "n", "default", "")
	cmd.Flags().BoolVarP(&flags.allNamespaces, "all-namespaces", "A", false, "If present, list requested resources from all namespaces.")

	return cmd
}

func runGet(args []string, params *utils.CmdParams, flags *getFlags) {
	kind := args[0]
	ns := flags.namespace
	if flags.allNamespaces {
		ns = ""
	}

	epc := effectivepolicy.NewCalculator(params.K8sClients, params.PolicyManager)
	policiesPrinter := &printer.PoliciesPrinter{Out: params.Out}
	httpRoutesPrinter := &printer.HTTPRoutesPrinter{Out: params.Out, EPC: epc}

	switch kind {
	case "policy", "policies":
		list := params.PolicyManager.GetPolicies()
		policiesPrinter.Print(list)

	case "policycrds":
		list := params.PolicyManager.GetCRDs()
		policiesPrinter.PrintCRDs(list)

	case "httproute", "httproutes":
		list, err := resourcehelpers.ListHTTPRoutes(context.TODO(), params.K8sClients, ns)
		if err != nil {
			panic(err)
		}
		httpRoutesPrinter.Print(list)

	default:
		fmt.Fprintf(os.Stderr, "Unrecognized RESOURCE_TYPE\n")
	}
}
