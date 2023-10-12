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
	"context"
	"fmt"
	"os"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	gatewayv1beta1 "sigs.k8s.io/gateway-api/apis/v1beta1"

	"github.com/spf13/cobra"
	"sigs.k8s.io/gateway-api/gwctl/pkg/cmd/utils"
	"sigs.k8s.io/gateway-api/gwctl/pkg/cmd/utils/printer"
	"sigs.k8s.io/gateway-api/gwctl/pkg/common/resourcehelpers"
	"sigs.k8s.io/gateway-api/gwctl/pkg/effectivepolicy"
	"sigs.k8s.io/gateway-api/gwctl/pkg/policymanager"
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
		ns = ""
	}

	epc := effectivepolicy.NewCalculator(params.K8sClients, params.PolicyManager)
	policiesPrinter := &printer.PoliciesPrinter{Out: params.Out}
	httpRoutesPrinter := &printer.HTTPRoutesPrinter{Out: params.Out, EPC: epc}
	gwPrinter := &printer.GatewaysPrinter{Out: params.Out, EPC: epc}
	gwcPrinter := &printer.GatewayClassesPrinter{Out: params.Out, EPC: epc}
	backendsPrinter := &printer.BackendsPrinter{Out: params.Out, EPC: epc}

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
		var httpRoutes []gatewayv1beta1.HTTPRoute
		if len(args) == 1 {
			var err error
			httpRoutes, err = resourcehelpers.ListHTTPRoutes(context.TODO(), params.K8sClients, ns)
			if err != nil {
				panic(err)
			}
		} else {
			httpRoute, err := resourcehelpers.GetHTTPRoute(context.TODO(), params.K8sClients, ns, args[1])
			if err != nil {
				panic(err)
			}
			httpRoutes = []gatewayv1beta1.HTTPRoute{httpRoute}
		}
		httpRoutesPrinter.PrintDescribeView(context.TODO(), httpRoutes)

	case "gateway", "gateways":
		var gws []gatewayv1beta1.Gateway
		if len(args) == 1 {
			var err error
			gws, err = resourcehelpers.ListGateways(context.TODO(), params.K8sClients, ns)
			if err != nil {
				panic(err)
			}
		} else {
			gw, err := resourcehelpers.GetGateways(context.TODO(), params.K8sClients, ns, args[1])
			if err != nil {
				panic(err)
			}
			gws = []gatewayv1beta1.Gateway{gw}
		}
		gwPrinter.PrintDescribeView(context.TODO(), gws)

	case "gatewayclass", "gatewayclasses":
		var gwClasses []gatewayv1beta1.GatewayClass
		if len(args) == 1 {
			var err error
			gwClasses, err = resourcehelpers.ListGatewayClasses(context.TODO(), params.K8sClients)
			if err != nil {
				panic(err)
			}
		} else {
			gwc, err := resourcehelpers.GetGatewayClass(context.TODO(), params.K8sClients, args[1])
			if err != nil {
				panic(err)
			}
			gwClasses = []gatewayv1beta1.GatewayClass{gwc}
		}
		gwcPrinter.PrintDescribeView(context.TODO(), gwClasses)

	case "backend", "backends":
		var backendsList []unstructured.Unstructured

		// We default the backends to just "Service" types initially.
		resourceType := "service"

		if len(args) == 1 {
			var err error
			backendsList, err = resourcehelpers.ListBackends(context.TODO(), params.K8sClients, resourceType, ns)
			if err != nil {
				panic(err)
			}
		} else {
			backend, err := resourcehelpers.GetBackend(context.TODO(), params.K8sClients, resourceType, ns, args[1])
			if err != nil {
				panic(err)
			}
			backendsList = []unstructured.Unstructured{backend}
		}
		backendsPrinter.PrintDescribeView(context.TODO(), backendsList)

	default:
		fmt.Fprintf(os.Stderr, "Unrecognized RESOURCE_TYPE\n")
	}
}
