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

	"sigs.k8s.io/gateway-api/gwctl/pkg/policymanager"
	"sigs.k8s.io/gateway-api/gwctl/pkg/printer"
	"sigs.k8s.io/gateway-api/gwctl/pkg/resourcediscovery"
	"sigs.k8s.io/gateway-api/gwctl/pkg/utils"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/utils/clock"
)

func NewDescribeCommand() *cobra.Command {
	var namespaceFlag string
	var allNamespacesFlag bool
	var labelSelector string

	cmd := &cobra.Command{
		Use:   "describe {policies|httproutes|gateways|gatewayclasses|backends|namespace|policycrd} RESOURCE_NAME",
		Short: "Show details of a specific resource or group of resources",
		Args:  cobra.RangeArgs(1, 2),
		Run: func(cmd *cobra.Command, args []string) {
			params := getParams(kubeConfigPath)
			runDescribe(cmd, args, params)
		},
	}
	cmd.Flags().StringVarP(&namespaceFlag, "namespace", "n", "default", "")
	cmd.Flags().BoolVarP(&allNamespacesFlag, "all-namespaces", "A", false, "If present, list requested resources from all namespaces.")
	cmd.Flags().StringVarP(&labelSelector, "selector", "l", "", "Selector (label query) to filter on, supports '=', '==', and '!='.(e.g. -l key1=value1,key2=value2). Matching objects must satisfy all of the specified label constraints.")

	return cmd
}

func runDescribe(cmd *cobra.Command, args []string, params *utils.CmdParams) {
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
		ns = metav1.NamespaceAll
	}

	discoverer := resourcediscovery.NewDiscoverer(params.K8sClients, params.PolicyManager)

	policiesPrinter := &printer.PoliciesPrinter{Writer: params.Out, Clock: clock.RealClock{}}
	httpRoutesPrinter := &printer.HTTPRoutesPrinter{Writer: params.Out, Clock: clock.RealClock{}}
	gwPrinter := &printer.GatewaysPrinter{Writer: params.Out, Clock: clock.RealClock{}}
	gwcPrinter := &printer.GatewayClassesPrinter{Writer: params.Out, Clock: clock.RealClock{}}
	backendsPrinter := &printer.BackendsPrinter{Writer: params.Out}
	namespacesPrinter := &printer.NamespacesPrinter{Writer: params.Out, Clock: clock.RealClock{}}

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
		policiesPrinter.PrintPoliciesDescribeView(policyList)

	case "policycrd", "policycrds":
		var policyCrdList []policymanager.PolicyCRD
		if len(args) == 1 {
			policyCrdList = params.PolicyManager.GetCRDs()
		} else {
			var found bool
			policyCrd, found := params.PolicyManager.GetCRD(args[1])
			if !found {
				fmt.Fprintf(os.Stderr, "failed to find PolicyCrd: %v\n", err)
				os.Exit(1)
			}
			policyCrdList = []policymanager.PolicyCRD{policyCrd}
		}
		policiesPrinter.PrintPolicyCRDsDescribeView(policyCrdList)

	case "httproute", "httproutes":
		selector, err := labels.Parse(labelSelector)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Unable to find resources that match the label selector \"%s\": %v\n", labelSelector, err)
			os.Exit(1)
		}
		filter := resourcediscovery.Filter{
			Namespace: ns,
			Labels:    selector,
		}
		if len(args) > 1 {
			filter.Name = args[1]
		}
		resourceModel, err := discoverer.DiscoverResourcesForHTTPRoute(filter)
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to discover HTTPRoute resources: %v\n", err)
			os.Exit(1)
		}
		httpRoutesPrinter.PrintDescribeView(resourceModel)

	case "gateway", "gateways":
		selector, err := labels.Parse(labelSelector)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Unable to find resources that match the label selector \"%s\": %v\n", labelSelector, err)
			os.Exit(1)
		}
		filter := resourcediscovery.Filter{
			Namespace: ns,
			Labels:    selector,
		}
		if len(args) > 1 {
			filter.Name = args[1]
		}
		resourceModel, err := discoverer.DiscoverResourcesForGateway(filter)
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to discover Gateway resources: %v\n", err)
			os.Exit(1)
		}
		gwPrinter.PrintDescribeView(resourceModel)

	case "gatewayclass", "gatewayclasses":
		selector, err := labels.Parse(labelSelector)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Unable to find resources that match the label selector \"%s\": %v\n", labelSelector, err)
			os.Exit(1)
		}
		filter := resourcediscovery.Filter{
			Labels: selector,
		}
		if len(args) > 1 {
			filter.Name = args[1]
		}
		resourceModel, err := discoverer.DiscoverResourcesForGatewayClass(filter)
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to discover GatewayClass resources: %v\n", err)
			os.Exit(1)
		}
		gwcPrinter.PrintDescribeView(resourceModel)

	case "backend", "backends":
		selector, err := labels.Parse(labelSelector)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Unable to find resources that match the label selector \"%s\": %v\n", labelSelector, err)
			os.Exit(1)
		}
		filter := resourcediscovery.Filter{
			Namespace: ns,
			Labels:    selector,
		}
		if len(args) > 1 {
			filter.Name = args[1]
		}
		resourceModel, err := discoverer.DiscoverResourcesForBackend(filter)
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to discover resources related to Backend: %v\n", err)
			os.Exit(1)
		}
		backendsPrinter.PrintDescribeView(resourceModel)

	case "namespace", "namespaces", "ns":
		selector, err := labels.Parse(labelSelector)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Unable to find resources that match the label selector \"%s\": %v\n", labelSelector, err)
			os.Exit(1)
		}
		filter := resourcediscovery.Filter{
			Labels: selector,
		}
		if len(args) > 1 {
			filter.Name = args[1]
		}

		resourceModel, err := discoverer.DiscoverResourcesForNamespace(filter)
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to discover Namespace resources: %v\n", err)
			os.Exit(1)
		}
		namespacesPrinter.PrintDescribeView(resourceModel)

	default:
		fmt.Fprintf(os.Stderr, "Unrecognized RESOURCE_TYPE\n")
	}
}
