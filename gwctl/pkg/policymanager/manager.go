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

package policymanager

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"golang.org/x/exp/maps"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"

	gatewayv1alpha2 "sigs.k8s.io/gateway-api/apis/v1alpha2"
	"sigs.k8s.io/gateway-api/gwctl/pkg/common"
)

type PolicyManager struct {
	Fetcher common.GroupKindFetcher

	// policyCRDs maps a CRD name to the CRD object.
	policyCRDs map[PolicyCrdID]*PolicyCRD
	// policies maps a policy name to the policy object.
	policies map[common.GKNN]*Policy
}

func New(fetcher common.GroupKindFetcher) *PolicyManager {
	return &PolicyManager{
		Fetcher:    fetcher,
		policyCRDs: make(map[PolicyCrdID]*PolicyCRD),
		policies:   make(map[common.GKNN]*Policy),
	}
}

// Init will construct a local cache of all Policy CRDs and Policy Resources.
func (p *PolicyManager) Init() error {
	err := p.initPolicyCRDs()
	if err != nil {
		return err
	}

	return p.initPolicies()
}

func (p *PolicyManager) initPolicyCRDs() error {
	crdGK := schema.GroupKind{Group: apiextensionsv1.GroupName, Kind: "CustomResourceDefinition"}

	allUnstructuredCRDs, err := p.Fetcher.Fetch(crdGK)
	if err != nil {
		return err
	}
	for _, uCRD := range allUnstructuredCRDs {
		crd := &apiextensionsv1.CustomResourceDefinition{}
		if err := runtime.DefaultUnstructuredConverter.FromUnstructured(uCRD.UnstructuredContent(), crd); err != nil {
			panic(fmt.Sprintf("failed to convert unstructured CustomResourceDefinition to structured: %v", err))
		}
		policyCRD := &PolicyCRD{crd}
		// Check if the CRD is a Gateway Policy CRD
		if policyCRD.IsValid() {
			p.policyCRDs[policyCRD.ID()] = policyCRD
		}
	}

	return nil
}

func (p *PolicyManager) initPolicies() error {
	for _, policyCRD := range p.policyCRDs {
		gk := schema.GroupKind{Group: policyCRD.CRD.Spec.Group, Kind: policyCRD.CRD.Spec.Names.Kind}
		policies, err := p.Fetcher.Fetch(gk)
		if err != nil {
			return err
		}

		for _, unstrucutredPolicy := range policies {
			policy, err := ConstructPolicy(unstrucutredPolicy, policyCRD.IsInheritable())
			if err != nil {
				return err
			}
			p.policies[policy.GKNN()] = &policy
		}
	}
	return nil
}

func (p *PolicyManager) PoliciesAttachedTo(objRef common.GKNN) []*Policy {
	var result []*Policy
	for _, policy := range p.policies {
		if policy.IsAttachedTo(objRef) {
			result = append(result, policy)
		}
	}
	return result
}

func (p *PolicyManager) GetCRDs() []*PolicyCRD {
	return maps.Values(p.policyCRDs)
}

func (p *PolicyManager) GetCRD(name string) (*PolicyCRD, bool) {
	for _, policyCrd := range p.policyCRDs {
		if name == policyCrd.CRD.Name {
			return policyCrd, true
		}
	}

	return nil, false
}

func (p *PolicyManager) GetPolicies() []*Policy {
	return maps.Values(p.policies)
}

// PolicyCrdID has the structurued "<CRD Kind>.<CRD Group>"
type PolicyCrdID string

type PolicyCRD struct {
	CRD *apiextensionsv1.CustomResourceDefinition
}

// ID returns a unique identifier for this PolicyCRD.
func (p PolicyCRD) ID() PolicyCrdID {
	return PolicyCrdID(p.CRD.Spec.Names.Kind + "." + p.CRD.Spec.Group)
}

// IsValid return true if the PolicyCRD satisfies requirements for qualifying as
// a Gateway Policy CRD.
func (p PolicyCRD) IsValid() bool {
	return p.IsInheritable() || p.IsDirect() || p.CRD.GetLabels()[gatewayv1alpha2.PolicyLabelKey] == "true"
}

func (p PolicyCRD) IsInheritable() bool {
	return strings.ToLower(p.CRD.GetLabels()[gatewayv1alpha2.PolicyLabelKey]) == "inherited"
}

func (p PolicyCRD) IsDirect() bool {
	return strings.ToLower(p.CRD.GetLabels()[gatewayv1alpha2.PolicyLabelKey]) == "direct"
}

// IsClusterScoped returns true if the CRD is cluster scoped. Such policies can
// be used to target a cluster scoped resource like GatewayClass.
func (p PolicyCRD) IsClusterScoped() bool {
	return p.CRD.Spec.Scope == apiextensionsv1.ClusterScoped
}

type Policy struct {
	Unstructured *unstructured.Unstructured
	// TargetRefs references the target objects this policy is attached to. This
	// only makes sense in case of a directly-attached-policy, or an
	// unmerged-inherited-policy.
	TargetRef common.GKNN
	// Indicates whether the policy is supposed to be "inherited" (as opposed to
	// "direct").
	Inheritable bool
}

func ConstructPolicy(u *unstructured.Unstructured, inherited bool) (Policy, error) {
	result := Policy{Unstructured: u}

	// Identify targetRef of Policy.
	type genericPolicy struct {
		metav1.TypeMeta   `json:",inline"`
		metav1.ObjectMeta `json:"metadata,omitempty"`
		Spec              struct {
			TargetRef gatewayv1alpha2.NamespacedPolicyTargetReference
		}
	}
	structuredPolicy := &genericPolicy{}
	if err := runtime.DefaultUnstructuredConverter.FromUnstructured(u.UnstructuredContent(), structuredPolicy); err != nil {
		return Policy{}, fmt.Errorf("failed to convert unstructured policy resource to structured: %v", err)
	}
	result.TargetRef = common.GKNN{
		Group:     string(structuredPolicy.Spec.TargetRef.Group),
		Kind:      string(structuredPolicy.Spec.TargetRef.Kind),
		Namespace: structuredPolicy.GetNamespace(),
		Name:      string(structuredPolicy.Spec.TargetRef.Name),
	}
	if result.TargetRef.Namespace == "" {
		result.TargetRef.Namespace = result.Unstructured.GetNamespace()
	}
	if structuredPolicy.Spec.TargetRef.Namespace != nil {
		result.TargetRef.Namespace = string(*structuredPolicy.Spec.TargetRef.Namespace)
	}

	result.Inheritable = inherited

	return result, nil
}

func (p Policy) GKNN() common.GKNN {
	return common.GKNN{
		Group:     p.Unstructured.GroupVersionKind().Group,
		Kind:      p.Unstructured.GroupVersionKind().Kind,
		Namespace: p.Unstructured.GetNamespace(),
		Name:      p.Unstructured.GetName(),
	}
}

// PolicyCrdID returns a unique identifier for the CRD of this policy.
func (p Policy) PolicyCrdID() PolicyCrdID {
	return PolicyCrdID(p.Unstructured.GetObjectKind().GroupVersionKind().Kind + "." + p.Unstructured.GetObjectKind().GroupVersionKind().Group)
}

func (p Policy) IsInheritable() bool {
	return p.Inheritable
}

func (p Policy) IsDirect() bool {
	return !p.Inheritable
}

func (p Policy) IsAttachedTo(objRef common.GKNN) bool {
	if p.TargetRef.Kind == "Namespace" && p.TargetRef.Name == "" {
		p.TargetRef.Name = "default"
	}
	if objRef.Kind == "Namespace" && objRef.Name == "" {
		objRef.Name = "default"
	}
	if p.TargetRef.Kind != "Namespace" && p.TargetRef.Namespace == "" {
		p.TargetRef.Namespace = "default"
	}
	if objRef.Kind != "Namespace" && objRef.Namespace == "" {
		objRef.Namespace = "default"
	}
	return p.TargetRef == objRef
}

func (p Policy) DeepCopy() *Policy {
	clone := &Policy{
		Unstructured: p.Unstructured.DeepCopy(),
		TargetRef:    p.TargetRef,
		Inheritable:  p.Inheritable,
	}
	return clone
}

func (p Policy) Spec() map[string]interface{} {
	spec, ok, err := unstructured.NestedFieldCopy(p.Unstructured.UnstructuredContent(), "spec")
	if err != nil || !ok {
		return nil
	}

	result, ok := spec.(map[string]interface{})
	if !ok {
		return nil
	}
	return result
}

func (p Policy) EffectiveSpec() (map[string]interface{}, error) {
	if !p.IsInheritable() {
		// No merging is required in case of Direct policies.
		result := p.Spec()
		delete(result, "targetRef")
		return result, nil
	}

	spec := p.Spec()
	if spec == nil {
		return nil, nil
	}

	defaultSpec, ok := p.Spec()["default"]
	if !ok {
		defaultSpec = make(map[string]interface{})
	}
	overrideSpec, ok := p.Spec()["override"]
	if !ok {
		overrideSpec = make(map[string]interface{})
	}

	// Check if both are non-scalar and merge them.
	defaultSpecNonScalar, isDefaultSpecNonScalar := defaultSpec.(map[string]interface{})
	overrideSpecNonScalar, isOverrideSpecNonScalar := overrideSpec.(map[string]interface{})
	if !isDefaultSpecNonScalar || !isOverrideSpecNonScalar {
		return nil, fmt.Errorf("spec.default and spec.override must be non-scalar")
	}

	result, err := mergeUnstructured(defaultSpecNonScalar, overrideSpecNonScalar)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (p *Policy) MarshalJSON() ([]byte, error) {
	effectiveSpec, err := p.EffectiveSpec()
	if err != nil {
		return nil, err
	}
	return json.Marshal(effectiveSpec)
}

func ConvertPoliciesMapToSlice(policies map[common.GKNN]*Policy) []*Policy {
	result := maps.Values(policies)
	sort.Slice(result, func(i, j int) bool {
		a := fmt.Sprintf("%v/%v/%v", result[i].PolicyCrdID(), result[i].Unstructured.GetNamespace(), result[i].Unstructured.GetName())
		b := fmt.Sprintf("%v/%v/%v", result[j].PolicyCrdID(), result[j].Unstructured.GetNamespace(), result[j].Unstructured.GetName())
		return a < b
	})
	return result
}
