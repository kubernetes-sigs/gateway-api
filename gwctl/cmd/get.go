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
	"strings"

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
		Args:  cobra.RangeArgs(1, 2),
		Run: func(cmd *cobra.Command, args []string) {
			params := getParams(kubeConfigPath)
			runGet(cmd, args, params)
		},
	}
	cmd.Flags().StringVarP(&namespaceFlag, "namespace", "n", "default", "")
	cmd.Flags().BoolVarP(&allNamespacesFlag, "all-namespaces", "A", false, "If present, list requested resources from all namespaces.")
	cmd.Flags().StringVarP(&labelSelector, "selector", "l", "", "Selector (label query) to filter on, supports '=', '==', and '!='.(e.g. -l key1=value1,key2=value2). Matching objects must satisfy all of the specified label constraints.")

	return cmd
}

func getKindAndName(args []string) (string, string, error) {
	if len(args) < 1 {
		return "", "", fmt.Errorf("no arguments found to be provided")
	}

	firstArg := args[0]
	splittedFirstArg := strings.Split(firstArg, "/")

	if len(splittedFirstArg) > 2 {
		return "", "", fmt.Errorf("more than two slashes found in the first argument")
	}
	if len(splittedFirstArg) == 2 {
		if len(args) > 1 {
			return "", "", fmt.Errorf("cannot provide name in a separate arg if already provided alongside the first arg as RESOURCE_TYPE/NAME")
		}
		kind, name := splittedFirstArg[0], splittedFirstArg[1]
		return kind, name, nil
	}

	kind, name := firstArg, ""
	if len(args) > 1 {
		name = args[1]
	}
	return kind, name, nil
}

func runGet(cmd *cobra.Command, args []string, params *utils.CmdParams) {
	kind, name, err := getKindAndName(args)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to parse kind and name: %v\n", err)
		os.Exit(1)
	}
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
	backendsPrinter := &printer.BackendsPrinter{Out: params.Out, Clock: realClock}

	switch kind {
	case "namespace", "namespaces", "ns":
		selector, err := labels.Parse(labelSelector)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Unable to find resources that match the label selector \"%s\": %v\n", labelSelector, err)
			os.Exit(1)
		}
		resourceModel, err := discoverer.DiscoverResourcesForNamespace(resourcediscovery.Filter{Labels: selector, Name: name})
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
		filter := resourcediscovery.Filter{Namespace: ns, Labels: selector, Name: name}
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
		filter := resourcediscovery.Filter{Namespace: ns, Labels: selector, Name: name}
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
		filter := resourcediscovery.Filter{Namespace: ns, Labels: selector, Name: name}
		resourceModel, err := discoverer.DiscoverResourcesForHTTPRoute(filter)
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to discover HTTPRoute resources: %v\n", err)
			os.Exit(1)
		}
		httpRoutesPrinter.Print(resourceModel)

	case "backend", "backends":
		selector, err := labels.Parse(labelSelector)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Unable to find resources that match the label selector \"%s\": %v\n", labelSelector, err)
			os.Exit(1)
		}
		filter := resourcediscovery.Filter{Namespace: ns, Labels: selector, Name: name}
		resourceModel, err := discoverer.DiscoverResourcesForBackend(filter)
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to discover backend resources: %v\n", err)
			os.Exit(1)
		}
		backendsPrinter.Print(resourceModel)

	default:
		fmt.Fprintf(os.Stderr, "Unrecognized RESOURCE_TYPE\n")
		os.Exit(1)
	}
}
