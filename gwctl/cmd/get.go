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
	var outputFormat string

	cmd := &cobra.Command{
		Use:   "get {namespaces|gateways|gatewayclasses|policies|policycrds|httproutes} RESOURCE_NAME",
		Short: "Display one or many resources",
		Args:  cobra.RangeArgs(1, 2),
		Run: func(cmd *cobra.Command, args []string) {
			params := getParams(kubeConfigPath)
			runGet(cmd, args, params)
		},
	}
	cmd.Flags().StringVarP(&namespaceFlag, "namespace", "n", "default", "")
	cmd.Flags().BoolVarP(&allNamespacesFlag, "all-namespaces", "A", false, "If present, list requested resources from all namespaces.")
	cmd.Flags().StringVarP(&labelSelector, "selector", "l", "", "Selector (label query) to filter on, supports '=', '==', and '!='.(e.g. -l key1=value1,key2=value2). Matching objects must satisfy all of the specified label constraints.")
	cmd.Flags().StringVarP(&outputFormat, "output", "o", "", `Output format. Must be one of (yaml, json)`)

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
	output, err := cmd.Flags().GetString("output")
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to read flag \"output\": %v\n", err)
		os.Exit(1)
	}
	outputFormat, err := utils.ValidateAndReturnOutputFormat(output)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}

	if allNs {
		ns = ""
	}

	discoverer := resourcediscovery.NewDiscoverer(params.K8sClients, params.PolicyManager)
	realClock := clock.RealClock{}

	nsPrinter := &printer.NamespacesPrinter{Writer: params.Out, Clock: realClock}
	gwPrinter := &printer.GatewaysPrinter{Writer: params.Out, Clock: realClock}
	gwcPrinter := &printer.GatewayClassesPrinter{Writer: params.Out, Clock: realClock}
	policiesPrinter := &printer.PoliciesPrinter{Writer: params.Out, Clock: realClock}
	httpRoutesPrinter := &printer.HTTPRoutesPrinter{Writer: params.Out, Clock: realClock}
	backendsPrinter := &printer.BackendsPrinter{Writer: params.Out, Clock: realClock}

	var resourceModel *resourcediscovery.ResourceModel
	var printerImpl printer.Printer

	switch kind {
	case "namespace", "namespaces", "ns":
		selector, err := labels.Parse(labelSelector)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to parse label selector %q: %v\n", labelSelector, err)
			os.Exit(1)
		}
		resourceModel, err = discoverer.DiscoverResourcesForNamespace(resourcediscovery.Filter{Labels: selector})
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to discover Namespace resources: %v\n", err)
			os.Exit(1)
		}
		printerImpl = nsPrinter

	case "gateway", "gateways":
		selector, err := labels.Parse(labelSelector)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to parse label selector %q: %v\n", labelSelector, err)
			os.Exit(1)
		}
		filter := resourcediscovery.Filter{Namespace: ns, Labels: selector}
		if len(args) > 1 {
			filter.Name = args[1]
		}
		resourceModel, err = discoverer.DiscoverResourcesForGateway(filter)
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to discover Gateway resources: %v\n", err)
			os.Exit(1)
		}
		printerImpl = gwPrinter

	case "gatewayclass", "gatewayclasses":
		selector, err := labels.Parse(labelSelector)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to parse label selector %q: %v\n", labelSelector, err)
			os.Exit(1)
		}
		filter := resourcediscovery.Filter{Namespace: ns, Labels: selector}
		if len(args) > 1 {
			filter.Name = args[1]
		}
		resourceModel, err = discoverer.DiscoverResourcesForGatewayClass(filter)
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to discover GatewayClass resources: %v\n", err)
			os.Exit(1)
		}
		printerImpl = gwcPrinter

	case "policy", "policies":
		list := params.PolicyManager.GetPolicies()
		policiesPrinter.PrintPolicies(list, outputFormat)
		return

	case "policycrd", "policycrds":
		list := params.PolicyManager.GetCRDs()
		policiesPrinter.PrintCRDs(list, outputFormat)
		return

	case "httproute", "httproutes":
		selector, err := labels.Parse(labelSelector)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to parse label selector %q: %v\n", labelSelector, err)
			os.Exit(1)
		}
		filter := resourcediscovery.Filter{Namespace: ns, Labels: selector}
		if len(args) > 1 {
			filter.Name = args[1]
		}
		resourceModel, err = discoverer.DiscoverResourcesForHTTPRoute(filter)
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to discover HTTPRoute resources: %v\n", err)
			os.Exit(1)
		}
		printerImpl = httpRoutesPrinter

	case "backend", "backends":
		selector, err := labels.Parse(labelSelector)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to parse label selector %q: %v\n", labelSelector, err)
			os.Exit(1)
		}
		filter := resourcediscovery.Filter{Namespace: ns, Labels: selector}
		if len(args) > 1 {
			filter.Name = args[1]
		}
		resourceModel, err = discoverer.DiscoverResourcesForBackend(filter)
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to discover backend resources: %v\n", err)
			os.Exit(1)
		}
		backendsPrinter.Print(resourceModel)
		return

	default:
		fmt.Fprintf(os.Stderr, "Unrecognized RESOURCE_TYPE\n")
		os.Exit(1)
	}
	printer.Print(printerImpl, resourceModel, outputFormat)
}
