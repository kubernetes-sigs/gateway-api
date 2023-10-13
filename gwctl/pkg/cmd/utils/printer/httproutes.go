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
	"context"
	_ "embed"
	"fmt"
	"io"
	"strings"
	"text/tabwriter"

	gatewayv1beta1 "sigs.k8s.io/gateway-api/apis/v1beta1"
	"sigs.k8s.io/yaml"

	"sigs.k8s.io/gateway-api/gwctl/pkg/effectivepolicy"
	"sigs.k8s.io/gateway-api/gwctl/pkg/policymanager"
)

type HTTPRoutesPrinter struct {
	Out io.Writer
	EPC *effectivepolicy.Calculator
}

func (hp *HTTPRoutesPrinter) Print(httpRoutes []gatewayv1beta1.HTTPRoute) {
	tw := tabwriter.NewWriter(hp.Out, 0, 0, 2, ' ', 0)
	row := []string{"NAME", "HOSTNAMES"}
	tw.Write([]byte(strings.Join(row, "\t") + "\n"))

	for _, httpRoute := range httpRoutes {
		var hostNames []string
		for _, hostName := range httpRoute.Spec.Hostnames {
			hostNames = append(hostNames, string(hostName))
		}
		hostNamesOutput := strings.Join(hostNames, ",")
		if cnt := len(hostNames); cnt > 2 {
			hostNamesOutput = fmt.Sprintf("%v + %v more", strings.Join(hostNames[:2], ","), cnt-2)
		}

		row := []string{httpRoute.Name, hostNamesOutput}
		tw.Write([]byte(strings.Join(row, "\t") + "\n"))
	}
	tw.Flush()
}

type httpRouteDescribeView struct {
	Name                     string                                                        `json:",omitempty"`
	Namespace                string                                                        `json:",omitempty"`
	Hostnames                []gatewayv1beta1.Hostname                                     `json:",omitempty"`
	ParentRefs               []gatewayv1beta1.ParentReference                              `json:",omitempty"`
	DirectlyAttachedPolicies []policymanager.ObjRef                                        `json:",omitempty"`
	EffectivePolicies        map[string]map[policymanager.PolicyCrdID]policymanager.Policy `json:",omitempty"`
}

func (hp *HTTPRoutesPrinter) PrintDescribeView(ctx context.Context, httpRoutes []gatewayv1beta1.HTTPRoute) {
	for i, httpRoute := range httpRoutes {
		directlyAttachedPolicies, err := hp.EPC.HTTPRoutes.GetDirectlyAttachedPolicies(ctx, httpRoute.Namespace, httpRoute.Name)
		if err != nil {
			panic(err)
		}
		effectivePolicies, err := hp.EPC.HTTPRoutes.GetEffectivePolicies(ctx, httpRoute.Namespace, httpRoute.Name)
		if err != nil {
			panic(err)
		}

		views := []httpRouteDescribeView{
			{
				Name:      httpRoute.GetName(),
				Namespace: httpRoute.GetNamespace(),
			},
			{
				Hostnames:  httpRoute.Spec.Hostnames,
				ParentRefs: httpRoute.Spec.ParentRefs,
			},
		}
		if policyRefs := policymanager.ToPolicyRefs(directlyAttachedPolicies); len(policyRefs) != 0 {
			views = append(views, httpRouteDescribeView{
				DirectlyAttachedPolicies: policyRefs,
			})
		}
		if len(effectivePolicies) != 0 {
			views = append(views, httpRouteDescribeView{
				EffectivePolicies: effectivePolicies,
			})
		}

		for _, view := range views {
			b, err := yaml.Marshal(view)
			if err != nil {
				panic(err)
			}
			fmt.Fprint(hp.Out, string(b))
		}

		if i+1 != len(httpRoutes) {
			fmt.Fprintf(hp.Out, "\n\n")
		}
	}
}
