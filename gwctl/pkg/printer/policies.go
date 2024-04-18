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

	"sigs.k8s.io/gateway-api/gwctl/pkg/policymanager"

	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/duration"
	"k8s.io/utils/clock"
	"sigs.k8s.io/yaml"
)

type PoliciesPrinter struct {
	Out   io.Writer
	Clock clock.Clock
}

func (pp *PoliciesPrinter) PrintPoliciesGetView(policies []policymanager.Policy) {
	sort.Slice(policies, func(i, j int) bool {
		a := fmt.Sprintf("%v/%v", policies[i].Unstructured().GetNamespace(), policies[i].Unstructured().GetName())
		b := fmt.Sprintf("%v/%v", policies[j].Unstructured().GetNamespace(), policies[j].Unstructured().GetName())
		return a < b
	})

	tw := tabwriter.NewWriter(pp.Out, 0, 0, 2, ' ', 0)
	row := []string{"NAME", "KIND", "TARGET NAME", "TARGET KIND", "POLICY TYPE", "AGE"}
	_, err := tw.Write([]byte(strings.Join(row, "\t") + "\n"))
	if err != nil {
		fmt.Fprint(os.Stderr, err)
		os.Exit(1)
	}

	for _, policy := range policies {
		policyType := "Direct"
		if policy.IsInherited() {
			policyType = "Inherited"
		}

		kind := fmt.Sprintf("%v.%v", policy.Unstructured().GroupVersionKind().Kind, policy.Unstructured().GroupVersionKind().Group)

		age := duration.HumanDuration(pp.Clock.Since(policy.Unstructured().GetCreationTimestamp().Time))

		row := []string{
			policy.Unstructured().GetName(),
			kind,
			policy.TargetRef().Name,
			policy.TargetRef().Kind,
			policyType,
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

func (pp *PoliciesPrinter) PrintPolicyCRDsGetView(policyCRDs []policymanager.PolicyCRD) {
	sort.Slice(policyCRDs, func(i, j int) bool {
		a := fmt.Sprintf("%v/%v", policyCRDs[i].CRD().GetNamespace(), policyCRDs[i].CRD().GetName())
		b := fmt.Sprintf("%v/%v", policyCRDs[j].CRD().GetNamespace(), policyCRDs[j].CRD().GetName())
		return a < b
	})

	tw := tabwriter.NewWriter(pp.Out, 0, 0, 2, ' ', 0)
	row := []string{"NAME", "POLICY TYPE", "SCOPE", "AGE"}
	_, err := tw.Write([]byte(strings.Join(row, "\t") + "\n"))
	if err != nil {
		fmt.Fprint(os.Stderr, err)
		os.Exit(1)
	}

	for _, policyCRD := range policyCRDs {
		policyType := "Direct"
		if policyCRD.IsInherited() {
			policyType = "Inherited"
		}

		age := duration.HumanDuration(pp.Clock.Since(policyCRD.CRD().GetCreationTimestamp().Time))

		row := []string{
			policyCRD.CRD().Name,
			policyType,
			string(policyCRD.CRD().Spec.Scope),
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

type policyDescribeView struct {
	Name      string                 `json:",omitempty"`
	Namespace string                 `json:",omitempty"`
	Group     string                 `json:",omitempty"`
	Kind      string                 `json:",omitempty"`
	Inherited string                 `json:",omitempty"`
	Spec      map[string]interface{} `json:",omitempty"`
}

func (pp *PoliciesPrinter) PrintPoliciesDescribeView(policies []policymanager.Policy) {
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

type policyCrdDescribeView struct {
	Name        string                                          `json:",omitempty"`
	Namespace   string                                          `json:",omitempty"`
	APIVersion  string                                          `json:",omitempty"`
	Kind        string                                          `json:",omitempty"`
	Labels      map[string]string                               `json:",omitempty"`
	Annotations map[string]string                               `json:",omitempty"`
	Metadata    *metav1.ObjectMeta                              `json:",omitempty"`
	Spec        *apiextensionsv1.CustomResourceDefinitionSpec   `json:",omitempty"`
	Status      *apiextensionsv1.CustomResourceDefinitionStatus `json:",omitempty"`
}

func (pp *PoliciesPrinter) PrintPolicyCRDsDescribeView(policyCrds []policymanager.PolicyCRD) {
	sort.Slice(policyCrds, func(i, j int) bool {
		a := fmt.Sprintf("%v/%v", policyCrds[i].CRD().GetNamespace(), policyCrds[i].CRD().GetName())
		b := fmt.Sprintf("%v/%v", policyCrds[j].CRD().GetNamespace(), policyCrds[j].CRD().GetName())
		return a < b
	})

	for i, policyCrd := range policyCrds {
		crd := policyCrd.CRD()

		metadata := crd.ObjectMeta.DeepCopy()
		metadata.Labels = nil
		metadata.Annotations = nil
		metadata.Name = ""
		metadata.Namespace = ""

		views := []policyCrdDescribeView{
			{
				Name:      crd.Name,
				Namespace: crd.Namespace,
			},
			{
				APIVersion: crd.APIVersion,
				Kind:       crd.Kind,
			},
			{
				Labels:      crd.Labels,
				Annotations: crd.Annotations,
			},
			{
				Metadata: metadata,
			},
			{
				Spec: &crd.Spec,
			},
			{
				Status: &crd.Status,
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

		if i+1 != len(policyCrds) {
			fmt.Fprintf(pp.Out, "\n\n")
		}
	}
}
