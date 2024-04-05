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

package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/utils/clock"

	"sigs.k8s.io/gateway-api/gwctl/pkg/printer"
	"sigs.k8s.io/gateway-api/gwctl/pkg/resourcediscovery"
	"sigs.k8s.io/gateway-api/gwctl/pkg/utils"
)

func NewGetCommand() *cobra.Command {
	var namespaceFlag string
	var allNamespacesFlag bool
	var labelSelector string

	cmd := &cobra.Command{
		Use:   "get {namespaces|gateways|gatewayclasses|policies|policycrds|httproutes}",
		Short: "Display one or many resources",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			params := getParams(kubeConfigPath)
			runGet(cmd, args, params)
		},
	}
	cmd.Flags().StringVarP(&namespaceFlag, "namespace", "n", "default", "")
	cmd.Flags().BoolVarP(&allNamespacesFlag, "all-namespaces", "A", false, "If present, list requested resources from all namespaces.")
	cmd.Flags().StringVarP(&labelSelector, "selector", "l", "", "Label selector.")

	return cmd
}

func runGet(cmd *cobra.Command, args []string, params *utils.CmdParams) {
	kind := args[0]
	ns, err := cmd.Flags().GetString("namespace")
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to read flag \"namespace\": %v\n", err)
		os.Exit(1)
	}

	allNs, err := cmd.Flags().GetBool("all-namespaces")
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to read flag \"all-namespaces\": %v\n", err)
		os.Exit(1)
	}

	labelSelector, err := cmd.Flags().GetString("selector")
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to read flag \"selector\": %v\n", err)
		os.Exit(1)
	}

	if allNs {
		ns = ""
	}

	discoverer := resourcediscovery.Discoverer{
		K8sClients:    params.K8sClients,
		PolicyManager: params.PolicyManager,
	}
	realClock := clock.RealClock{}
	nsPrinter := &printer.NamespacesPrinter{Out: params.Out, Clock: realClock}
	gwPrinter := &printer.GatewaysPrinter{Out: params.Out, Clock: realClock}
	gwcPrinter := &printer.GatewayClassesPrinter{Out: params.Out, Clock: realClock}
	policiesPrinter := &printer.PoliciesPrinter{Out: params.Out, Clock: realClock}
	httpRoutesPrinter := &printer.HTTPRoutesPrinter{Out: params.Out, Clock: realClock}

	switch kind {
	case "namespace", "namespaces", "ns":
		selector, err := labels.Parse(labelSelector)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Unable to find resources that match the label selector \"%s\": %v\n", labelSelector, err)
			os.Exit(1)
		}
		resourceModel, err := discoverer.DiscoverResourcesForNamespace(resourcediscovery.Filter{Labels: selector})
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to discover Namespace resources: %v\n", err)
			os.Exit(1)
		}
		nsPrinter.Print(resourceModel)

	case "gateway", "gateways":
		selector, err := labels.Parse(labelSelector)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Unable to find resources that match the label selector \"%s\": %v\n", labelSelector, err)
			os.Exit(1)
		}
		filter := resourcediscovery.Filter{Namespace: ns, Labels: selector}
		if len(args) > 1 {
			filter.Name = args[1]
		}
		resourceModel, err := discoverer.DiscoverResourcesForGateway(filter)
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to discover Gateway resources: %v\n", err)
			os.Exit(1)
		}
		gwPrinter.Print(resourceModel)

	case "gatewayclass", "gatewayclasses":
		selector, err := labels.Parse(labelSelector)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Unable to find resources that match the label selector \"%s\": %v\n", labelSelector, err)
			os.Exit(1)
		}
		filter := resourcediscovery.Filter{Namespace: ns, Labels: selector}
		if len(args) > 1 {
			filter.Name = args[1]
		}
		resourceModel, err := discoverer.DiscoverResourcesForGatewayClass(filter)
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to discover GatewayClass resources: %v\n", err)
			os.Exit(1)
		}
		gwcPrinter.Print(resourceModel)

	case "policy", "policies":
		list := params.PolicyManager.GetPolicies()
		policiesPrinter.PrintPoliciesGetView(list)

	case "policycrd", "policycrds":
		list := params.PolicyManager.GetCRDs()
		policiesPrinter.PrintPolicyCRDsGetView(list)

	case "httproute", "httproutes":
		selector, err := labels.Parse(labelSelector)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Unable to find resources that match the label selector \"%s\": %v\n", labelSelector, err)
			os.Exit(1)
		}
		filter := resourcediscovery.Filter{Namespace: ns, Labels: selector}
		if len(args) > 1 {
			filter.Name = args[1]
		}
		resourceModel, err := discoverer.DiscoverResourcesForHTTPRoute(filter)
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to discover HTTPRoute resources: %v\n", err)
			os.Exit(1)
		}
		httpRoutesPrinter.Print(resourceModel)

	default:
		fmt.Fprintf(os.Stderr, "Unrecognized RESOURCE_TYPE\n")
	}
}
