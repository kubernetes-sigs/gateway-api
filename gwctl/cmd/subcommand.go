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
	"io"
	"os"
	"strings"

	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/utils/clock"

	gatewayv1 "sigs.k8s.io/gateway-api/apis/v1"
	"sigs.k8s.io/gateway-api/gwctl/pkg/common"
	"sigs.k8s.io/gateway-api/gwctl/pkg/policymanager"
	"sigs.k8s.io/gateway-api/gwctl/pkg/printer"
	"sigs.k8s.io/gateway-api/gwctl/pkg/resourcediscovery"
	cmdutils "sigs.k8s.io/gateway-api/gwctl/pkg/utils"
)

type commandName string

const (
	commandNameGet      commandName = "get"
	commandNameDescribe commandName = "describe"
)

func NewSubCommand(f cmdutils.Factory, out io.Writer, cmdName commandName) *cobra.Command {
	var shortMsg string
	if cmdName == commandNameGet {
		shortMsg = "Display one or many resources"
	} else {
		shortMsg = "Show details of a specific resource or group of resources"
	}

	cmd := &cobra.Command{
		Use:   string(cmdName),
		Short: shortMsg,
	}
	cmd.AddCommand(newCmdNamespaces(f, out, cmdName))
	cmd.AddCommand(newCmdGatewayClasses(f, out, cmdName))
	cmd.AddCommand(newCmdGateways(f, out, cmdName))
	cmd.AddCommand(newCmdHTTPRoutes(f, out, cmdName))
	cmd.AddCommand(newCmdBackends(f, out, cmdName))
	cmd.AddCommand(newCmdPolicies(f, out, cmdName))
	cmd.AddCommand(newCmdPolicyCRDs(f, out, cmdName))
	return cmd
}

func newCmdNamespaces(f cmdutils.Factory, out io.Writer, cmdName commandName) *cobra.Command {
	o := &getOrDescribeOptions{out: out, cmdName: cmdName}
	cmd := &cobra.Command{
		Use:     "namespaces",
		Aliases: []string{"namespace", "ns"},
		Short:   "Display one or more Namespaces",
		Args:    cobra.RangeArgs(0, 1),
		Run: func(_ *cobra.Command, args []string) {
			o.parse(args)
			runGetOrDescribeNamespaces(f, o)
		},
	}
	addLabelSelectorFlag(&o.labelSelectorFlag, cmd)
	if cmdName == commandNameGet {
		addOutputFormatFlag(&o.outputFlag, cmd)
	}
	return cmd
}

func newCmdGatewayClasses(f cmdutils.Factory, out io.Writer, cmdName commandName) *cobra.Command {
	o := &getOrDescribeOptions{out: out, cmdName: cmdName}
	cmd := &cobra.Command{
		Use:     "gatewayclasses",
		Aliases: []string{"gatewayclass"},
		Short:   "Display one or more GatewayClasses",
		Args:    cobra.RangeArgs(0, 1),
		Run: func(_ *cobra.Command, args []string) {
			o.parse(args)
			runGetOrDescribeGatewayClasses(f, o)
		},
	}
	addLabelSelectorFlag(&o.labelSelectorFlag, cmd)
	if cmdName == commandNameGet {
		addForFlag(&o.forFlag, cmd)
		addOutputFormatFlag(&o.outputFlag, cmd)
	}
	return cmd
}

func newCmdGateways(f cmdutils.Factory, out io.Writer, cmdName commandName) *cobra.Command {
	o := &getOrDescribeOptions{out: out, cmdName: cmdName}
	cmd := &cobra.Command{
		Use:     "gateways",
		Aliases: []string{"gateway", "gw", "Gateways", "Gateway"},
		Short:   "Display one or more Gateways",
		Args:    cobra.RangeArgs(0, 1),
		Run: func(_ *cobra.Command, args []string) {
			o.parse(args)
			runGetOrDescribeGateways(f, o)
		},
	}
	addNamespaceFlag(&o.namespaceFlag, cmd)
	addAllNamespacesFlag(&o.allNamespacesFlag, cmd)
	addLabelSelectorFlag(&o.labelSelectorFlag, cmd)
	if cmdName == commandNameGet {
		addForFlag(&o.forFlag, cmd)
		addOutputFormatFlag(&o.outputFlag, cmd)
	}
	return cmd
}

func newCmdHTTPRoutes(f cmdutils.Factory, out io.Writer, cmdName commandName) *cobra.Command {
	o := &getOrDescribeOptions{out: out, cmdName: cmdName}
	cmd := &cobra.Command{
		Use:     "httproutes",
		Aliases: []string{"httproute"},
		Short:   "Display one or more HTTPRoutes",
		Args:    cobra.RangeArgs(0, 1),
		Run: func(_ *cobra.Command, args []string) {
			o.parse(args)
			runGetOrDescribeHTTPRoutes(f, o)
		},
	}
	addNamespaceFlag(&o.namespaceFlag, cmd)
	addAllNamespacesFlag(&o.allNamespacesFlag, cmd)
	addLabelSelectorFlag(&o.labelSelectorFlag, cmd)
	if cmdName == commandNameGet {
		addForFlag(&o.forFlag, cmd)
		addOutputFormatFlag(&o.outputFlag, cmd)
	}
	return cmd
}

func newCmdBackends(f cmdutils.Factory, out io.Writer, cmdName commandName) *cobra.Command {
	o := &getOrDescribeOptions{out: out, cmdName: cmdName}
	cmd := &cobra.Command{
		Use:     "backends",
		Aliases: []string{"backend"},
		Short:   "Display one or more Backends",
		Args:    cobra.RangeArgs(0, 1),
		Run: func(_ *cobra.Command, args []string) {
			o.parse(args)
			runGetOrDescribeBackends(f, o)
		},
	}
	addNamespaceFlag(&o.namespaceFlag, cmd)
	addAllNamespacesFlag(&o.allNamespacesFlag, cmd)
	addLabelSelectorFlag(&o.labelSelectorFlag, cmd)
	if cmdName == commandNameGet {
		addForFlag(&o.forFlag, cmd)
		addOutputFormatFlag(&o.outputFlag, cmd)
	}
	return cmd
}

func newCmdPolicies(f cmdutils.Factory, out io.Writer, cmdName commandName) *cobra.Command {
	o := &getOrDescribeOptions{out: out, cmdName: cmdName}
	cmd := &cobra.Command{
		Use:     "policies",
		Aliases: []string{"policy"},
		Short:   "Display one or more Policies",
		Args:    cobra.RangeArgs(0, 1),
		Run: func(_ *cobra.Command, args []string) {
			o.parse(args)
			runGetOrDescribePolicies(f, o)
		},
	}
	if cmdName == commandNameGet {
		addForFlag(&o.forFlag, cmd)
		addOutputFormatFlag(&o.outputFlag, cmd)
	}
	return cmd
}

func newCmdPolicyCRDs(f cmdutils.Factory, out io.Writer, cmdName commandName) *cobra.Command {
	o := &getOrDescribeOptions{out: out, cmdName: cmdName}
	cmd := &cobra.Command{
		Use:     "policycrds",
		Aliases: []string{"policycrd"},
		Short:   "Display one or more Policy CRDs",
		Args:    cobra.RangeArgs(0, 1),
		Run: func(_ *cobra.Command, args []string) {
			o.parse(args)
			runGetOrDescribePolicyCRDs(f, o)
		},
	}
	if cmdName == commandNameGet {
		addOutputFormatFlag(&o.outputFlag, cmd)
	}
	return cmd
}

func runGetOrDescribeNamespaces(f cmdutils.Factory, o *getOrDescribeOptions) {
	k8sClients, err := f.K8sClients()
	handleErrOrExitWithMsg(err, "")
	policyManager, err := f.PolicyManager()
	handleErrOrExitWithMsg(err, "")

	discoverer := resourcediscovery.NewDiscoverer(k8sClients, policyManager)
	resourceModel, err := discoverer.DiscoverResourcesForNamespace(o.toResourceDiscoveryFilter())
	handleErrOrExitWithMsg(err, "failed to discover Namespace resources")

	realClock := clock.RealClock{}
	nsPrinter := &printer.NamespacesPrinter{Writer: o.out, Clock: realClock, EventFetcher: discoverer}
	if o.cmdName == commandNameGet {
		printer.Print(nsPrinter, resourceModel, o.outputFormat)
	} else {
		nsPrinter.PrintDescribeView(resourceModel)
	}
}

func runGetOrDescribeGatewayClasses(f cmdutils.Factory, o *getOrDescribeOptions) {
	k8sClients, err := f.K8sClients()
	handleErrOrExitWithMsg(err, "")
	policyManager, err := f.PolicyManager()
	handleErrOrExitWithMsg(err, "")

	discoverer := resourcediscovery.NewDiscoverer(k8sClients, policyManager)
	emptyObjRef := common.ObjRef{}
	var resourceModel *resourcediscovery.ResourceModel
	if o.cmdName == commandNameGet && o.forObjRef != emptyObjRef {
		switch o.forObjRef.Kind {
		case "Gateway":
			resourceModel, err = discoverer.DiscoverResourcesForGateway(o.forObjRefToResourceDiscoveryFilter())
		default:
			fmt.Fprintf(os.Stderr, "Filtering by type %q is not supported for GatewayClasses", o.forObjRef.Kind)
			os.Exit(1)
		}
	} else {
		resourceModel, err = discoverer.DiscoverResourcesForGatewayClass(o.toResourceDiscoveryFilter())
	}
	handleErrOrExitWithMsg(err, "failed to discover GatewayClass resources")

	realClock := clock.RealClock{}
	gwcPrinter := &printer.GatewayClassesPrinter{Writer: o.out, Clock: realClock, EventFetcher: discoverer}
	if o.cmdName == commandNameGet {
		printer.Print(gwcPrinter, resourceModel, o.outputFormat)
	} else {
		gwcPrinter.PrintDescribeView(resourceModel)
	}
}

func runGetOrDescribeGateways(f cmdutils.Factory, o *getOrDescribeOptions) {
	k8sClients, err := f.K8sClients()
	handleErrOrExitWithMsg(err, "")
	policyManager, err := f.PolicyManager()
	handleErrOrExitWithMsg(err, "")

	discoverer := resourcediscovery.NewDiscoverer(k8sClients, policyManager)
	emptyObjRef := common.ObjRef{}
	var resourceModel *resourcediscovery.ResourceModel
	if o.cmdName == commandNameGet && o.forObjRef != emptyObjRef {
		switch o.forObjRef.Kind {
		case "GatewayClass":
			resourceModel, err = discoverer.DiscoverResourcesForGatewayClass(o.forObjRefToResourceDiscoveryFilter())
		case "HTTPRoute":
			resourceModel, err = discoverer.DiscoverResourcesForHTTPRoute(o.forObjRefToResourceDiscoveryFilter())
		default:
			fmt.Fprintf(os.Stderr, "Filtering by type %q is not supported for Gateways", o.forObjRef.Kind)
			os.Exit(1)
		}
	} else {
		resourceModel, err = discoverer.DiscoverResourcesForGateway(o.toResourceDiscoveryFilter())
	}
	handleErrOrExitWithMsg(err, "failed to discover Gateway resources")

	realClock := clock.RealClock{}
	gwPrinter := &printer.GatewaysPrinter{Writer: o.out, Clock: realClock, EventFetcher: discoverer}
	if o.cmdName == commandNameGet {
		printer.Print(gwPrinter, resourceModel, o.outputFormat)
	} else {
		gwPrinter.PrintDescribeView(resourceModel)
	}
}

func runGetOrDescribeHTTPRoutes(f cmdutils.Factory, o *getOrDescribeOptions) {
	k8sClients, err := f.K8sClients()
	handleErrOrExitWithMsg(err, "")
	policyManager, err := f.PolicyManager()
	handleErrOrExitWithMsg(err, "")

	discoverer := resourcediscovery.NewDiscoverer(k8sClients, policyManager)
	emptyObjRef := common.ObjRef{}
	var resourceModel *resourcediscovery.ResourceModel
	if o.cmdName == commandNameGet && o.forObjRef != emptyObjRef {
		switch o.forObjRef.Kind {
		case "Gateway":
			resourceModel, err = discoverer.DiscoverResourcesForGateway(o.forObjRefToResourceDiscoveryFilter())
		case "Service":
			resourceModel, err = discoverer.DiscoverResourcesForBackend(o.forObjRefToResourceDiscoveryFilter())
		default:
			fmt.Fprintf(os.Stderr, "Filtering by type %q is not supported for HTTPRoutes", o.forObjRef.Kind)
			os.Exit(1)
		}
	} else {
		resourceModel, err = discoverer.DiscoverResourcesForHTTPRoute(o.toResourceDiscoveryFilter())
	}
	handleErrOrExitWithMsg(err, "failed to discover HTTPRoute resources")

	realClock := clock.RealClock{}
	httpRoutesPrinter := &printer.HTTPRoutesPrinter{Writer: o.out, Clock: realClock}
	if o.cmdName == commandNameGet {
		printer.Print(httpRoutesPrinter, resourceModel, o.outputFormat)
	} else {
		httpRoutesPrinter.PrintDescribeView(resourceModel)
	}
}

func runGetOrDescribeBackends(f cmdutils.Factory, o *getOrDescribeOptions) {
	k8sClients, err := f.K8sClients()
	handleErrOrExitWithMsg(err, "")
	policyManager, err := f.PolicyManager()
	handleErrOrExitWithMsg(err, "")

	discoverer := resourcediscovery.NewDiscoverer(k8sClients, policyManager)
	emptyObjRef := common.ObjRef{}
	var resourceModel *resourcediscovery.ResourceModel
	if o.cmdName == commandNameGet && o.forObjRef != emptyObjRef {
		switch o.forObjRef.Kind {
		case "Gateway":
			resourceModel, err = discoverer.DiscoverResourcesForGateway(o.forObjRefToResourceDiscoveryFilter())
		case "HTTPRoute":
			resourceModel, err = discoverer.DiscoverResourcesForHTTPRoute(o.forObjRefToResourceDiscoveryFilter())
		default:
			fmt.Fprintf(os.Stderr, "Filtering by type %q is not supported for Backends", o.forObjRef.Kind)
			os.Exit(1)
		}
	} else {
		resourceModel, err = discoverer.DiscoverResourcesForBackend(o.toResourceDiscoveryFilter())
	}
	handleErrOrExitWithMsg(err, "failed to discover Backend resources")

	realClock := clock.RealClock{}
	backendsPrinter := &printer.BackendsPrinter{Writer: o.out, Clock: realClock, EventFetcher: discoverer}
	if o.cmdName == commandNameGet {
		printer.Print(backendsPrinter, resourceModel, o.outputFormat)
	} else {
		backendsPrinter.PrintDescribeView(resourceModel)
	}
}

func runGetOrDescribePolicies(f cmdutils.Factory, o *getOrDescribeOptions) {
	policyManager, err := f.PolicyManager()
	handleErrOrExitWithMsg(err, "")

	realClock := clock.RealClock{}
	policiesPrinter := &printer.PoliciesPrinter{Writer: o.out, Clock: realClock}

	var policyList []policymanager.Policy
	emptyObjRef := common.ObjRef{}
	switch {
	case o.cmdName == commandNameGet && o.forObjRef != emptyObjRef: // Fetch policies attached to some resource.
		policyList = policyManager.PoliciesAttachedTo(o.forObjRef)
	case o.resourceName == "": // Fetch all policies.
		policyList = policyManager.GetPolicies()
	default: // Fetch a single policy by its name.
		var found bool
		policy, found := policyManager.GetPolicy(o.namespace + "/" + o.resourceName)
		if !found && o.resourceName == "default" {
			policy, found = policyManager.GetPolicy("/" + o.resourceName)
		}
		if found {
			policyList = []policymanager.Policy{policy}
		}
	}

	if o.cmdName == commandNameGet {
		policiesPrinter.PrintPolicies(policyList, o.outputFormat)
	} else {
		policiesPrinter.PrintPoliciesDescribeView(policyList)
	}
}

func runGetOrDescribePolicyCRDs(f cmdutils.Factory, o *getOrDescribeOptions) {
	policyManager, err := f.PolicyManager()
	handleErrOrExitWithMsg(err, "")

	realClock := clock.RealClock{}
	policiesPrinter := &printer.PoliciesPrinter{Writer: o.out, Clock: realClock}

	var policyCrdList []policymanager.PolicyCRD
	if o.resourceName == "" {
		policyCrdList = policyManager.GetCRDs()
	} else {
		var found bool
		policyCrd, found := policyManager.GetCRD(o.resourceName)
		if !found {
			fmt.Fprintf(os.Stderr, "failed to find PolicyCrd: %v\n", err)
			os.Exit(1)
		}
		policyCrdList = []policymanager.PolicyCRD{policyCrd}
	}

	if o.cmdName == commandNameGet {
		policiesPrinter.PrintCRDs(policyCrdList, o.outputFormat)
	} else {
		policiesPrinter.PrintPolicyCRDsDescribeView(policyCrdList)
	}
}

type getOrDescribeOptions struct {
	cmdName commandName

	namespaceFlag     string
	allNamespacesFlag bool
	labelSelectorFlag string
	outputFlag        string
	forFlag           string

	namespace     string
	resourceName  string
	labelSelector labels.Selector
	outputFormat  cmdutils.OutputFormat
	forObjRef     common.ObjRef

	out io.Writer
}

func (o *getOrDescribeOptions) parse(args []string) {
	o.namespace = o.namespaceFlag
	if o.allNamespacesFlag {
		o.namespace = metav1.NamespaceAll
	}

	if len(args) >= 1 {
		o.resourceName = args[0]
	}

	var err error
	o.labelSelector, err = labels.Parse(o.labelSelectorFlag)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to parse label selector %q: %v\n", o.labelSelectorFlag, err)
		os.Exit(1)
	}

	o.outputFormat, err = cmdutils.ValidateAndReturnOutputFormat(o.outputFlag)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}

	// Parse `--for` flag
	if o.forFlag != "" {
		parts := strings.Split(o.forFlag, "/")
		if len(parts) < 2 || len(parts) > 3 {
			fmt.Fprintf(os.Stderr, "invalid value used in --for flag; value must be in the format TYPE[/NAMESPACE]/NAME\n")
			os.Exit(1)
		}
		if len(parts) == 2 {
			o.forObjRef = common.ObjRef{Kind: parts[0], Namespace: metav1.NamespaceDefault, Name: parts[1]}
		} else {
			o.forObjRef = common.ObjRef{Kind: parts[0], Namespace: parts[1], Name: parts[2]}
		}
		switch strings.ToLower(o.forObjRef.Kind) {
		case "gatewayclass", "gateawyclasses":
			o.forObjRef.Group = gatewayv1.GroupVersion.Group
			o.forObjRef.Kind = "GatewayClass"
			o.forObjRef.Namespace = ""
		case "gateway", "gateways":
			o.forObjRef.Group = gatewayv1.GroupVersion.Group
			o.forObjRef.Kind = "Gateway"
		case "httproute", "httproutes":
			o.forObjRef.Group = gatewayv1.GroupVersion.Group
			o.forObjRef.Kind = "HTTPRoute"
		case "service", "services":
			o.forObjRef.Kind = "Service"
		default:
			fmt.Fprintf(os.Stderr, "invalid type provided in --for flag; type must be one of [gatewayclass, gateway, httproute, service]\n")
			os.Exit(1)
		}
	}
}

func (o *getOrDescribeOptions) toResourceDiscoveryFilter() resourcediscovery.Filter {
	return resourcediscovery.Filter{
		Name:      o.resourceName,
		Namespace: o.namespace,
		Labels:    o.labelSelector,
	}
}

func (o *getOrDescribeOptions) forObjRefToResourceDiscoveryFilter() resourcediscovery.Filter {
	return resourcediscovery.Filter{
		Name:      o.forObjRef.Name,
		Namespace: o.forObjRef.Namespace,
	}
}

func handleErrOrExitWithMsg(err error, msg string) {
	if err == nil {
		return
	}
	var str string
	if msg != "" {
		str = msg + ": "
	}
	str += err.Error()
	fmt.Fprintf(os.Stderr, "Error: %s\n", str)
	os.Exit(1)
}
