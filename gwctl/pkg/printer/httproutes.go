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
	"strings"
	"text/tabwriter"

	gatewayv1 "sigs.k8s.io/gateway-api/apis/v1"
	"sigs.k8s.io/yaml"

	"sigs.k8s.io/gateway-api/gwctl/pkg/policymanager"
	"sigs.k8s.io/gateway-api/gwctl/pkg/resourcediscovery"
)

type HTTPRoutesPrinter struct {
	Out io.Writer
}

func (hp *HTTPRoutesPrinter) Print(resourceModel *resourcediscovery.ResourceModel) {
	tw := tabwriter.NewWriter(hp.Out, 0, 0, 2, ' ', 0)
	row := []string{"NAME", "HOSTNAMES"}
	tw.Write([]byte(strings.Join(row, "\t") + "\n"))

	for _, httpRouteNode := range resourceModel.HTTPRoutes {
		var hostNames []string
		for _, hostName := range httpRouteNode.HTTPRoute.Spec.Hostnames {
			hostNames = append(hostNames, string(hostName))
		}
		hostNamesOutput := strings.Join(hostNames, ",")
		if cnt := len(hostNames); cnt > 2 {
			hostNamesOutput = fmt.Sprintf("%v + %v more", strings.Join(hostNames[:2], ","), cnt-2)
		}

		row := []string{httpRouteNode.HTTPRoute.Name, hostNamesOutput}
		tw.Write([]byte(strings.Join(row, "\t") + "\n"))
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
				panic(err)
			}
			fmt.Fprint(hp.Out, string(b))
		}

		if index+1 <= len(resourceModel.HTTPRoutes) {
			fmt.Fprintf(hp.Out, "\n\n")
		}
	}
}
