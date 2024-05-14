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
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/yaml"

	"sigs.k8s.io/gateway-api/gwctl/pkg/policymanager"
	"sigs.k8s.io/gateway-api/gwctl/pkg/resourcediscovery"
)

type BackendsPrinter struct {
	io.Writer
	Clock clock.Clock
}

func (bp *BackendsPrinter) Print(resourceModel *resourcediscovery.ResourceModel) {
	tw := tabwriter.NewWriter(bp, 0, 0, 2, ' ', 0)
	row := []string{"NAMESPACE", "NAME", "TYPE", "REFERRED BY ROUTES", "AGE", "POLICIES"}
	_, err := tw.Write([]byte(strings.Join(row, "\t") + "\n"))
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to write to the tab writer: %v\n", err)
		os.Exit(1)
	}

	backends := maps.Values(resourceModel.Backends)
	sortedBackends := SortByString(backends)

	for _, backendNode := range sortedBackends {
		backend := backendNode.Backend

		httpRouteNodes := maps.Values(backendNode.HTTPRoutes)
		sortedHTTPRouteNodes := SortByString(httpRouteNodes)
		totalRoutes := len(sortedHTTPRouteNodes)

		var referredByRoutes string
		if totalRoutes == 0 {
			referredByRoutes = "None"
		} else {
			var routes []string
			for i, httpRouteNode := range sortedHTTPRouteNodes {
				if i < 2 {
					namespacedName := client.ObjectKeyFromObject(httpRouteNode.HTTPRoute).String()
					routes = append(routes, namespacedName)
				} else {
					break
				}
			}
			referredByRoutes = strings.Join(routes, ", ")
			if totalRoutes > 2 {
				referredByRoutes += fmt.Sprintf(" + %d more", totalRoutes-2)
			}
		}

		namespace := backend.GetNamespace()
		name := backend.GetName()
		backendType := backend.GetKind()
		age := duration.HumanDuration(bp.Clock.Since(backend.GetCreationTimestamp().Time))
		policiesCount := fmt.Sprintf("%d", len(backendNode.Policies))

		row := []string{
			namespace,
			name,
			backendType,
			referredByRoutes,
			age,
			policiesCount,
		}
		_, err := tw.Write([]byte(strings.Join(row, "\t") + "\n"))
		if err != nil {
			fmt.Fprint(os.Stderr, err)
			os.Exit(1)
		}
	}

	tw.Flush()
}

type backendDescribeView struct {
	Group                    string                 `json:",omitempty"`
	Kind                     string                 `json:",omitempty"`
	Name                     string                 `json:",omitempty"`
	Namespace                string                 `json:",omitempty"`
	DirectlyAttachedPolicies []policymanager.ObjRef `json:",omitempty"`
	EffectivePolicies        any                    `json:",omitempty"`
}

func (bp *BackendsPrinter) PrintDescribeView(resourceModel *resourcediscovery.ResourceModel) {
	index := 0
	for _, backendNode := range resourceModel.Backends {
		index++

		views := []backendDescribeView{
			{
				Group:     backendNode.Backend.GroupVersionKind().Group,
				Kind:      backendNode.Backend.GroupVersionKind().Kind,
				Name:      backendNode.Backend.GetName(),
				Namespace: backendNode.Backend.GetNamespace(),
			},
		}
		if policyRefs := resourcediscovery.ConvertPoliciesMapToPolicyRefs(backendNode.Policies); len(policyRefs) != 0 {
			views = append(views, backendDescribeView{
				DirectlyAttachedPolicies: policyRefs,
			})
		}
		if len(backendNode.EffectivePolicies) != 0 {
			views = append(views, backendDescribeView{
				EffectivePolicies: backendNode.EffectivePolicies,
			})
		}

		for _, view := range views {
			b, err := yaml.Marshal(view)
			if err != nil {
				fmt.Fprintf(os.Stderr, "failed to marshal to yaml: %v\n", err)
				os.Exit(1)
			}
			fmt.Fprint(bp, string(b))
		}

		if index+1 <= len(resourceModel.Backends) {
			fmt.Fprintf(bp, "\n\n")
		}
	}
}
