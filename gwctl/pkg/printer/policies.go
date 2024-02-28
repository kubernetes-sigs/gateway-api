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

	"sigs.k8s.io/yaml"

	"sigs.k8s.io/gateway-api/gwctl/pkg/policymanager"
)

type PoliciesPrinter struct {
	Out io.Writer
}

func (pp *PoliciesPrinter) Print(policies []policymanager.Policy) {
	sort.Slice(policies, func(i, j int) bool {
		a := fmt.Sprintf("%v/%v", policies[i].Unstructured().GetNamespace(), policies[i].Unstructured().GetName())
		b := fmt.Sprintf("%v/%v", policies[j].Unstructured().GetNamespace(), policies[j].Unstructured().GetName())
		return a < b
	})

	tw := tabwriter.NewWriter(pp.Out, 0, 0, 2, ' ', 0)
	row := []string{"POLICY NAME", "POLICY KIND", "TARGET NAME", "TARGET KIND", "POLICY TYPE"}
	tw.Write([]byte(strings.Join(row, "\t") + "\n"))

	for _, policy := range policies {
		policyType := "Direct"
		if policy.IsInherited() {
			policyType = "Inherited"
		}
		row := []string{
			policy.Unstructured().GetName(),
			policy.Unstructured().GroupVersionKind().Kind,
			policy.TargetRef().Name,
			policy.TargetRef().Kind,
			policyType,
		}
		tw.Write([]byte(strings.Join(row, "\t") + "\n"))
	}
	tw.Flush()
}

func (pp *PoliciesPrinter) PrintCRDs(policyCRDs []policymanager.PolicyCRD) {
	sort.Slice(policyCRDs, func(i, j int) bool {
		a := fmt.Sprintf("%v/%v", policyCRDs[i].CRD().GetNamespace(), policyCRDs[i].CRD().GetName())
		b := fmt.Sprintf("%v/%v", policyCRDs[j].CRD().GetNamespace(), policyCRDs[j].CRD().GetName())
		return a < b
	})

	tw := tabwriter.NewWriter(pp.Out, 0, 0, 2, ' ', 0)
	row := []string{"NAME", "GROUP", "KIND", "POLICY TYPE", "SCOPE"}
	tw.Write([]byte(strings.Join(row, "\t") + "\n"))

	for _, policyCRD := range policyCRDs {
		policyType := "Direct"
		if policyCRD.IsInherited() {
			policyType = "Inherited"
		}
		row := []string{
			policyCRD.CRD().Name,
			policyCRD.CRD().Spec.Group,
			policyCRD.CRD().Spec.Names.Kind,
			policyType,
			string(policyCRD.CRD().Spec.Scope),
		}
		tw.Write([]byte(strings.Join(row, "\t") + "\n"))
	}
	tw.Flush()
}

type policyDescribeView struct {
	Name      string                 `json:",omitempty"`
	Namespace string                 `json:",omitempty"`
	Group     string                 `json:",omitempty"`
	Kind      string                 `json:",omitempty"`
	Inherited string                 `json:",omitempty"`
	Spec      map[string]interface{} `json:",omitempty"`
}

func (pp *PoliciesPrinter) PrintDescribeView(policies []policymanager.Policy) {
	sort.Slice(policies, func(i, j int) bool {
		a := fmt.Sprintf("%v/%v", policies[i].Unstructured().GetNamespace(), policies[i].Unstructured().GetName())
		b := fmt.Sprintf("%v/%v", policies[j].Unstructured().GetNamespace(), policies[j].Unstructured().GetName())
		return a < b
	})

	for i, policy := range policies {
		views := []policyDescribeView{
			{
				Name:      policy.Unstructured().GetName(),
				Namespace: policy.Unstructured().GetNamespace(),
			},
			{
				Group: policy.Unstructured().GroupVersionKind().Group,
				Kind:  policy.Unstructured().GroupVersionKind().Kind,
			},
			{
				Inherited: fmt.Sprintf("%v", policy.IsInherited()),
			},
			{
				Spec: policy.Spec(),
			},
		}

		for _, view := range views {
			b, err := yaml.Marshal(view)
			if err != nil {
				fmt.Fprintf(os.Stderr, "failed to marshal to yaml: %v\n", err)
				os.Exit(1)
			}
			fmt.Fprint(pp.Out, string(b))
		}

		if i+1 != len(policies) {
			fmt.Fprintf(pp.Out, "\n\n")
		}
	}
}
