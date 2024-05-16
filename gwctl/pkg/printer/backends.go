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

	"golang.org/x/exp/maps"
	"k8s.io/apimachinery/pkg/util/duration"
	"k8s.io/utils/clock"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"sigs.k8s.io/gateway-api/gwctl/pkg/resourcediscovery"
)

type BackendsPrinter struct {
	io.Writer
	Clock clock.Clock
}

func (bp *BackendsPrinter) Print(resourceModel *resourcediscovery.ResourceModel) {
	table := &Table{
		ColumnNames:  []string{"NAMESPACE", "NAME", "TYPE", "REFERRED BY ROUTES", "AGE", "POLICIES"},
		UseSeparator: false,
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
		table.Rows = append(table.Rows, row)
	}

	table.Write(bp, 0)
}

func (bp *BackendsPrinter) PrintDescribeView(resourceModel *resourcediscovery.ResourceModel) {
	index := 0
	for _, backendNode := range resourceModel.Backends {
		index++

		backend := backendNode.Backend.DeepCopy()
		backend.SetLabels(nil)
		backend.SetAnnotations(nil)

		pairs := []*DescriberKV{
			{Key: "Name", Value: backendNode.Backend.GetName()},
			{Key: "Namespace", Value: backendNode.Backend.GetNamespace()},
			{Key: "Labels", Value: backendNode.Backend.GetLabels()},
			{Key: "Annotations", Value: backendNode.Backend.GetAnnotations()},
			{Key: "Backend", Value: backend},
		}

		// ReferencedByRoutes
		routes := &Table{
			ColumnNames:  []string{"Kind", "Name"},
			UseSeparator: true,
		}
		for _, httpRouteNode := range backendNode.HTTPRoutes {
			row := []string{
				httpRouteNode.HTTPRoute.Kind, // Kind
				fmt.Sprintf("%v/%v", httpRouteNode.HTTPRoute.Namespace, httpRouteNode.HTTPRoute.Name), // Name
			}
			routes.Rows = append(routes.Rows, row)
		}
		pairs = append(pairs, &DescriberKV{Key: "ReferencedByRoutes", Value: routes})

		// DirectlyAttachedPolicies
		policyRefs := resourcediscovery.ConvertPoliciesMapToPolicyRefs(backendNode.Policies)
		pairs = append(pairs, &DescriberKV{Key: "DirectlyAttachedPolicies", Value: convertPolicyRefsToTable(policyRefs)})

		// EffectivePolicies
		if len(backendNode.EffectivePolicies) != 0 {
			pairs = append(pairs, &DescriberKV{Key: "EffectivePolicies", Value: backendNode.EffectivePolicies})
		}

		// ReferenceGrants
		if len(backendNode.ReferenceGrants) != 0 {
			var names []string
			for _, refGrantNode := range backendNode.ReferenceGrants {
				names = append(names, refGrantNode.ReferenceGrant.Name)
			}
			pairs = append(pairs, &DescriberKV{Key: "ReferenceGrants", Value: names})
		}

		// Analysis
		if len(backendNode.Errors) != 0 {
			pairs = append(pairs, &DescriberKV{Key: "Analysis", Value: convertErrorsToString(backendNode.Errors)})
		}

		// Events
		pairs = append(pairs, &DescriberKV{Key: "Events", Value: convertEventsSliceToTable(backendNode.Events, bp.Clock)})

		Describe(bp, pairs)

		if index+1 <= len(resourceModel.Backends) {
			fmt.Fprintf(bp, "\n\n")
		}
	}
}
