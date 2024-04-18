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

	gatewayv1 "sigs.k8s.io/gateway-api/apis/v1"
	"sigs.k8s.io/gateway-api/gwctl/pkg/policymanager"
	"sigs.k8s.io/gateway-api/gwctl/pkg/resourcediscovery"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/duration"
	"k8s.io/utils/clock"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/yaml"
)

type GatewayClassesPrinter struct {
	Out   io.Writer
	Clock clock.Clock
}

type gatewayClassDescribeView struct {
	APIVersion  string             `json:",omitempty"`
	Kind        string             `json:",omitempty"`
	Metadata    *metav1.ObjectMeta `json:",omitempty"`
	Labels      *map[string]string `json:",omitempty"`
	Annotations *map[string]string `json:",omitempty"`

	// GatewayClass name
	Name           string `json:",omitempty"`
	ControllerName string `json:",omitempty"`
	// GatewayClass description
	Description *string `json:",omitempty"`

	Status                   *gatewayv1.GatewayClassStatus `json:",omitempty"`
	DirectlyAttachedPolicies []policymanager.ObjRef        `json:",omitempty"`
}

func (gcp *GatewayClassesPrinter) Print(model *resourcediscovery.ResourceModel) {
	tw := tabwriter.NewWriter(gcp.Out, 0, 0, 2, ' ', 0)
	row := []string{"NAME", "CONTROLLER", "ACCEPTED", "AGE"}
	_, err := tw.Write([]byte(strings.Join(row, "\t") + "\n"))
	if err != nil {
		fmt.Fprint(os.Stderr, err)
		os.Exit(1)
	}

	gatewayClassNodes := make([]*resourcediscovery.GatewayClassNode, 0, len(model.GatewayClasses))
	for _, gatewayClassNode := range model.GatewayClasses {
		gatewayClassNodes = append(gatewayClassNodes, gatewayClassNode)
	}

	sort.Slice(gatewayClassNodes, func(i, j int) bool {
		if gatewayClassNodes[i].GatewayClass.GetName() != gatewayClassNodes[j].GatewayClass.GetName() {
			return gatewayClassNodes[i].GatewayClass.GetName() < gatewayClassNodes[j].GatewayClass.GetName()
		}
		return string(gatewayClassNodes[i].GatewayClass.Spec.ControllerName) < string(gatewayClassNodes[j].GatewayClass.Spec.ControllerName)
	})

	for _, gatewayClassNode := range gatewayClassNodes {
		accepted := "Unknown"
		for _, condition := range gatewayClassNode.GatewayClass.Status.Conditions {
			if condition.Type == "Accepted" {
				accepted = string(condition.Status)
			}
		}

		age := duration.HumanDuration(gcp.Clock.Since(gatewayClassNode.GatewayClass.GetCreationTimestamp().Time))

		row := []string{
			gatewayClassNode.GatewayClass.GetName(),
			string(gatewayClassNode.GatewayClass.Spec.ControllerName),
			accepted,
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

func (gcp *GatewayClassesPrinter) PrintDescribeView(resourceModel *resourcediscovery.ResourceModel) {
	index := 0
	for _, gatewayClassNode := range resourceModel.GatewayClasses {
		index++
		apiVersion, kind := gatewayClassNode.GatewayClass.GetObjectKind().GroupVersionKind().ToAPIVersionAndKind()
		metadata := gatewayClassNode.GatewayClass.ObjectMeta.DeepCopy()
		metadata.Labels = nil
		metadata.Annotations = nil
		metadata.Name = ""
		metadata.Namespace = ""

		// views ordered with respect to https://gateway-api.sigs.k8s.io/geps/gep-2722/
		views := []gatewayClassDescribeView{
			{
				Name: gatewayClassNode.GatewayClass.GetName(),
			},
			{
				Labels: ptr.To(gatewayClassNode.GatewayClass.GetLabels()),
			},
			{
				Annotations: ptr.To(gatewayClassNode.GatewayClass.GetAnnotations()),
			},
			{
				APIVersion: apiVersion,
			},
			{
				Kind: kind,
			},
			{
				Metadata: metadata,
			},
			{
				ControllerName: string(gatewayClassNode.GatewayClass.Spec.ControllerName),
			},
		}
		if gatewayClassNode.GatewayClass.Spec.Description != nil {
			views = append(views, gatewayClassDescribeView{
				Description: gatewayClassNode.GatewayClass.Spec.Description,
			})
		}
		views = append(views, gatewayClassDescribeView{
			Status: &gatewayClassNode.GatewayClass.Status,
		})

		if policyRefs := resourcediscovery.ConvertPoliciesMapToPolicyRefs(gatewayClassNode.Policies); len(policyRefs) != 0 {
			views = append(views, gatewayClassDescribeView{
				DirectlyAttachedPolicies: policyRefs,
			})
		}

		for _, view := range views {
			b, err := yaml.Marshal(view)
			if err != nil {
				fmt.Fprintf(os.Stderr, "failed to marshal to yaml: %v\n", err)
				os.Exit(1)
			}
			output := string(b)

			emptyOutput := strings.TrimSpace(output) == "{}"
			if !emptyOutput {
				fmt.Fprint(gcp.Out, output)
			}
		}

		if index+1 <= len(resourceModel.GatewayClasses) {
			fmt.Fprintf(gcp.Out, "\n\n")
		}
	}
}
