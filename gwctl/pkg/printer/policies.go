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

	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/duration"
	"k8s.io/klog/v2"

	"sigs.k8s.io/gateway-api/gwctl/pkg/common"
	"sigs.k8s.io/gateway-api/gwctl/pkg/policymanager"
	"sigs.k8s.io/gateway-api/gwctl/pkg/topology"
)

func (p *TablePrinter) printPolicy(policyNode *topology.Node, w io.Writer) error {
	if err := p.checkTypeChange("Policy", w); err != nil {
		return err
	}

	if p.table == nil {
		p.table = &Table{
			ColumnNames:  []string{"NAME", "KIND", "TARGET NAME", "TARGET KIND", "POLICY TYPE", "AGE"},
			UseSeparator: false,
		}
	}

	var err error
	var policy *policymanager.Policy
	if policy, err = accessPolicyOrCRD[policymanager.Policy](policyNode, common.PolicyGK); err != nil {
		return err
	}

	policyType := "Direct"
	if policy.IsInheritable() {
		policyType = "Inherited"
	}

	kind := fmt.Sprintf("%v.%v", policy.Unstructured.GroupVersionKind().Kind, policy.Unstructured.GroupVersionKind().Group)

	age := "<unknown>"
	creationTimestamp := policy.Unstructured.GetCreationTimestamp()
	if !creationTimestamp.IsZero() {
		age = duration.HumanDuration(p.Clock.Since(creationTimestamp.Time))
	}

	row := []string{
		policy.Unstructured.GetName(),
		kind,
		policy.TargetRef.Name,
		policy.TargetRef.Kind,
		policyType,
		age,
	}
	p.table.Rows = append(p.table.Rows, row)

	return nil
}

func (p *TablePrinter) printPolicyCRD(policyCRDNode *topology.Node, w io.Writer) error {
	if err := p.checkTypeChange("Policy", w); err != nil {
		return err
	}

	if p.table == nil {
		p.table = &Table{
			ColumnNames:  []string{"NAME", "POLICY TYPE", "SCOPE", "AGE"},
			UseSeparator: false,
		}
	}

	var err error
	var policyCRD *policymanager.PolicyCRD
	if policyCRD, err = accessPolicyOrCRD[policymanager.PolicyCRD](policyCRDNode, common.PolicyCRDGK); err != nil {
		return err
	}

	policyType := "Direct"
	if policyCRD.IsInheritable() {
		policyType = "Inherited"
	}

	age := "<unknown>"
	creationTimestamp := policyCRD.CRD.GetCreationTimestamp()
	if !creationTimestamp.IsZero() {
		age = duration.HumanDuration(p.Clock.Since(creationTimestamp.Time))
	}

	row := []string{
		policyCRD.CRD.Name,
		policyType,
		string(policyCRD.CRD.Spec.Scope),
		age,
	}
	p.table.Rows = append(p.table.Rows, row)
	return nil
}

func (p *DescriptionPrinter) printPolicy(policyNode *topology.Node, w io.Writer) error {
	if p.printSeparator {
		fmt.Fprintf(w, "\n\n")
	}
	p.printSeparator = true

	var err error
	var policy *policymanager.Policy
	if policy, err = accessPolicyOrCRD[policymanager.Policy](policyNode, common.PolicyGK); err != nil {
		return err
	}

	pairs := []*DescriberKV{
		{Key: "Name", Value: policy.Unstructured.GetName()},
		{Key: "Namespace", Value: policy.Unstructured.GetNamespace()},
		{Key: "Group", Value: policy.Unstructured.GroupVersionKind().Group},
		{Key: "Kind", Value: policy.Unstructured.GroupVersionKind().Kind},
		{Key: "Inherited", Value: fmt.Sprintf("%v", policy.IsInheritable())},
		{Key: "Spec", Value: policy.Spec()},
	}

	Describe(w, pairs)
	return nil
}

func (p *DescriptionPrinter) printPolicyCRD(policyCRDNode *topology.Node, w io.Writer) error {
	if p.printSeparator {
		fmt.Fprintf(w, "\n\n")
	}
	p.printSeparator = true

	var err error
	var policyCRD *policymanager.PolicyCRD
	if policyCRD, err = accessPolicyOrCRD[policymanager.PolicyCRD](policyCRDNode, common.PolicyCRDGK); err != nil {
		return err
	}

	crd := policyCRD.CRD

	metadata := crd.ObjectMeta.DeepCopy()
	metadata.Labels = nil
	metadata.Annotations = nil
	metadata.Name = ""
	metadata.Namespace = ""

	pairs := []*DescriberKV{
		{Key: "Name", Value: crd.Name},
		{Key: "Namespace", Value: crd.Namespace},
		{Key: "APIVersion", Value: crd.APIVersion},
		{Key: "Kind", Value: crd.Kind},
		{Key: "Labels", Value: crd.Labels},
		{Key: "Annotations", Value: crd.Annotations},
		{Key: "Metadata", Value: metadata},
		{Key: "Spec", Value: crd.Spec},
		{Key: "Status", Value: crd.Status},
	}
	Describe(w, pairs)
	return nil
}

func accessPolicyOrCRD[T any](node *topology.Node, gk schema.GroupKind) (*T, error) {
	rawData, ok := node.Metadata[gk.String()]
	if !ok || rawData == nil {
		klog.V(3).InfoS(fmt.Sprintf("no %v found in node", gk.String()), "node", node.GKNN())
		return nil, nil
	}
	data, ok := rawData.(*T)
	if !ok {
		return nil, fmt.Errorf("unable to perform type assertion to %v in node %v", gk.String(), node.GKNN())
	}
	return data, nil
}
