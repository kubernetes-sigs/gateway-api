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

package printer

import (
	"fmt"
	"io"
	"os"
	"strings"
	"text/tabwriter"

	"golang.org/x/exp/maps"
	"k8s.io/apimachinery/pkg/util/duration"
	"k8s.io/utils/clock"
	"sigs.k8s.io/yaml"

	gatewayv1 "sigs.k8s.io/gateway-api/apis/v1"
	"sigs.k8s.io/gateway-api/gwctl/pkg/policymanager"
	"sigs.k8s.io/gateway-api/gwctl/pkg/resourcediscovery"
)

var _ Printer = (*HTTPRoutesPrinter)(nil)

type HTTPRoutesPrinter struct {
	io.Writer
	Clock clock.Clock
}

func (hp *HTTPRoutesPrinter) GetPrintableNodes(resourceModel *resourcediscovery.ResourceModel) []NodeResource {
	return NodeResources(maps.Values(resourceModel.HTTPRoutes))
}

func (hp *HTTPRoutesPrinter) PrintTable(resourceModel *resourcediscovery.ResourceModel) {
	tw := tabwriter.NewWriter(hp, 0, 0, 2, ' ', 0)
	row := []string{"NAMESPACE", "NAME", "HOSTNAMES", "PARENT REFS", "AGE"}
	_, err := tw.Write([]byte(strings.Join(row, "\t") + "\n"))
	if err != nil {
		fmt.Fprint(os.Stderr, err)
		os.Exit(1)
	}

	httpRouteNodes := maps.Values(resourceModel.HTTPRoutes)

	for _, httpRouteNode := range SortByString(httpRouteNodes) {
		var hostNames []string
		for _, hostName := range httpRouteNode.HTTPRoute.Spec.Hostnames {
			hostNames = append(hostNames, string(hostName))
		}
		hostNamesOutput := "None"
		if hostNamesCount := len(hostNames); hostNamesCount > 0 {
			if hostNamesCount > 2 {
				hostNamesOutput = fmt.Sprintf("%v + %v more", strings.Join(hostNames[:2], ","), hostNamesCount-2)
			} else {
				hostNamesOutput = strings.Join(hostNames, ",")
			}
		}

		parentRefsCount := fmt.Sprintf("%d", len(httpRouteNode.HTTPRoute.Spec.ParentRefs))

		age := duration.HumanDuration(hp.Clock.Since(httpRouteNode.HTTPRoute.GetCreationTimestamp().Time))

		row := []string{
			httpRouteNode.HTTPRoute.GetNamespace(),
			httpRouteNode.HTTPRoute.GetName(),
			hostNamesOutput,
			parentRefsCount,
			age,
		}
		_, err := tw.Write([]byte(strings.Join(row, "\t") + "\n"))
		if err != nil {
			fmt.Fprint(os.Stderr, err)
			os.Exit(1)
		}
	}
	tw.Flush()
}

type httpRouteDescribeView struct {
	Name                     string                      `json:",omitempty"`
	Namespace                string                      `json:",omitempty"`
	Hostnames                []gatewayv1.Hostname        `json:",omitempty"`
	ParentRefs               []gatewayv1.ParentReference `json:",omitempty"`
	DirectlyAttachedPolicies []policymanager.ObjRef      `json:",omitempty"`
	EffectivePolicies        any                         `json:",omitempty"`
}

func (hp *HTTPRoutesPrinter) PrintDescribeView(resourceModel *resourcediscovery.ResourceModel) {
	index := 0
	for _, httpRouteNode := range resourceModel.HTTPRoutes {
		index++

		views := []httpRouteDescribeView{
			{
				Name:      httpRouteNode.HTTPRoute.GetName(),
				Namespace: httpRouteNode.HTTPRoute.GetNamespace(),
			},
			{
				Hostnames:  httpRouteNode.HTTPRoute.Spec.Hostnames,
				ParentRefs: httpRouteNode.HTTPRoute.Spec.ParentRefs,
			},
		}
		if policyRefs := resourcediscovery.ConvertPoliciesMapToPolicyRefs(httpRouteNode.Policies); len(policyRefs) != 0 {
			views = append(views, httpRouteDescribeView{
				DirectlyAttachedPolicies: policyRefs,
			})
		}
		if len(httpRouteNode.EffectivePolicies) != 0 {
			views = append(views, httpRouteDescribeView{
				EffectivePolicies: httpRouteNode.EffectivePolicies,
			})
		}

		for _, view := range views {
			b, err := yaml.Marshal(view)
			if err != nil {
				fmt.Fprintf(os.Stderr, "failed to marshal to yaml: %v\n", err)
				os.Exit(1)
			}
			fmt.Fprint(hp, string(b))
		}

		if index+1 <= len(resourceModel.HTTPRoutes) {
			fmt.Fprintf(hp, "\n\n")
		}
	}
}
