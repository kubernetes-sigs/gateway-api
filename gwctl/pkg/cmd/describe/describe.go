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

package describe

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"sigs.k8s.io/gateway-api/gwctl/pkg/cmd/utils"
	"sigs.k8s.io/gateway-api/gwctl/pkg/policymanager"
	"sigs.k8s.io/gateway-api/gwctl/pkg/printer"
	"sigs.k8s.io/gateway-api/gwctl/pkg/resourcediscovery"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type describeFlags struct {
	namespace     string
	allNamespaces bool
}

func NewDescribeCommand(params *utils.CmdParams) *cobra.Command {
	flags := &describeFlags{}

	cmd := &cobra.Command{
		Use:   "describe {policies|httproutes|gateways|gatewayclasses|backends} RESOURCE_NAME",
		Short: "Show details of a specific resource or group of resources",
		Args:  cobra.RangeArgs(1, 2),
		Run: func(cmd *cobra.Command, args []string) {
			runDescribe(args, params, flags)
		},
	}
	cmd.Flags().StringVarP(&flags.namespace, "namespace", "n", "default", "")
	cmd.Flags().BoolVarP(&flags.allNamespaces, "all-namespaces", "A", false, "If present, list requested resources from all namespaces.")

	return cmd
}

func runDescribe(args []string, params *utils.CmdParams, flags *describeFlags) {
	kind := args[0]
	ns := flags.namespace
	if flags.allNamespaces {
		ns = metav1.NamespaceAll
	}

	discoverer := resourcediscovery.Discoverer{
		K8sClients:    params.K8sClients,
		PolicyManager: params.PolicyManager,
	}
	policiesPrinter := &printer.PoliciesPrinter{Out: params.Out}
	httpRoutesPrinter := &printer.HTTPRoutesPrinter{Out: params.Out}
	gwPrinter := &printer.GatewaysPrinter{Out: params.Out}
	gwcPrinter := &printer.GatewayClassesPrinter{Out: params.Out}
	backendsPrinter := &printer.BackendsPrinter{Out: params.Out}

	switch kind {
	case "policy", "policies":
		var policyList []policymanager.Policy
		if len(args) == 1 {
			policyList = params.PolicyManager.GetPolicies()
		} else {
			var found bool
			policy, found := params.PolicyManager.GetPolicy(ns + "/" + args[1])
			if !found && ns == "default" {
				policy, found = params.PolicyManager.GetPolicy("/" + args[1])
			}
			if found {
				policyList = []policymanager.Policy{policy}
			}
		}
		policiesPrinter.PrintDescribeView(policyList)

	case "httproute", "httproutes":
		filter := resourcediscovery.Filter{Namespace: ns}
		if len(args) > 1 {
			filter.Name = args[1]
		}
		resourceModel, err := discoverer.DiscoverResourcesForHTTPRoute(filter)
		if err != nil {
			panic(err)
		}
		httpRoutesPrinter.PrintDescribeView(resourceModel)

	case "gateway", "gateways":
		filter := resourcediscovery.Filter{Namespace: ns}
		if len(args) > 1 {
			filter.Name = args[1]
		}
		resourceModel, err := discoverer.DiscoverResourcesForGateway(filter)
		if err != nil {
			panic(err)
		}
		gwPrinter.PrintDescribeView(resourceModel)

	case "gatewayclass", "gatewayclasses":
		filter := resourcediscovery.Filter{}
		if len(args) > 1 {
			filter.Name = args[1]
		}
		resourceModel, err := discoverer.DiscoverResourcesForGatewayClass(filter)
		if err != nil {
			panic(err)
		}
		gwcPrinter.PrintDescribeView(resourceModel)

	case "backend", "backends":
		filter := resourcediscovery.Filter{
			Namespace: ns,
		}
		if len(args) > 1 {
			filter.Name = args[1]
		}
		resourceModel, err := discoverer.DiscoverResourcesForBackend(filter)
		if err != nil {
			panic(err)
		}
		backendsPrinter.PrintDescribeView(resourceModel)

	default:
		fmt.Fprintf(os.Stderr, "Unrecognized RESOURCE_TYPE\n")
	}
}
