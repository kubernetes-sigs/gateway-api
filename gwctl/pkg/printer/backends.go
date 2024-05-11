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
	"sort"
	"strings"
	"text/tabwriter"

	"k8s.io/apimachinery/pkg/util/duration"
	"k8s.io/utils/clock"

	"sigs.k8s.io/controller-runtime/pkg/client"

	"sigs.k8s.io/gateway-api/gwctl/pkg/policymanager"
	"sigs.k8s.io/gateway-api/gwctl/pkg/resourcediscovery"

	"sigs.k8s.io/yaml"
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

	backendNodes := make([]*resourcediscovery.BackendNode, 0, len(resourceModel.Backends))
	for _, backendNode := range resourceModel.Backends {
		backendNodes = append(backendNodes, backendNode)
	}

	sort.Slice(backendNodes, func(i, j int) bool {
		if backendNodes[i].Backend.GetNamespace() != backendNodes[j].Backend.GetNamespace() {
			return backendNodes[i].Backend.GetNamespace() < backendNodes[j].Backend.GetNamespace()
		}
		return backendNodes[i].Backend.GetName() < backendNodes[j].Backend.GetName()
	})

	for _, backendNode := range backendNodes {
		backend := backendNode.Backend

		parentHTTPRoutes := []string{}
		remainderHTTPRoutes := 0

		httpRouteNodes := make([]*resourcediscovery.HTTPRouteNode, len(backendNode.HTTPRoutes))
		i := 0
		for _, node := range backendNode.HTTPRoutes {
			httpRouteNodes[i] = node
			i++
		}
		sort.Slice(httpRouteNodes, func(i, j int) bool {
			if httpRouteNodes[i].HTTPRoute.GetNamespace() != httpRouteNodes[j].HTTPRoute.GetNamespace() {
				return httpRouteNodes[i].HTTPRoute.GetNamespace() < httpRouteNodes[j].HTTPRoute.GetNamespace()
			}
			return httpRouteNodes[i].HTTPRoute.GetName() < httpRouteNodes[j].HTTPRoute.GetName()
		})

		for _, httpRouteNode := range httpRouteNodes {
			httpRoute := httpRouteNode.HTTPRoute

			if len(parentHTTPRoutes) < 2 {
				namespacedName := client.ObjectKeyFromObject(httpRoute).String()
				parentHTTPRoutes = append(parentHTTPRoutes, namespacedName)
			} else {
				remainderHTTPRoutes++
			}
		}

		referredByRoutes := "None"
		if len(parentHTTPRoutes) != 0 {
			referredByRoutes = strings.Join(parentHTTPRoutes, ",")
			if remainderHTTPRoutes != 0 {
				referredByRoutes += fmt.Sprintf(" + %d more", remainderHTTPRoutes)
			}
		}

		namespace := backendNode.Backend.GetNamespace()
		name := backendNode.Backend.GetName()
		backendType := backendNode.Backend.GetKind()
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
		if _, err = tw.Write([]byte(strings.Join(row, "\t") + "\n")); err != nil {
			fmt.Fprintf(os.Stderr, "failed to write to the tab writer: %v\n", err)
			os.Exit(1)
		}
	}
	if err = tw.Flush(); err != nil {
		fmt.Fprintf(os.Stderr, "failed to flush to the tab writer: %v\n", err)
		os.Exit(1)
	}
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
