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
	"context"
	"encoding/json"
	"fmt"
	"strings"

	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	gatewayv1alpha2 "sigs.k8s.io/gateway-api/apis/v1alpha2"
)

type PolicyManager struct {
	dc dynamic.Interface

	// policyCRDs maps a CRD name to the CRD object.
	policyCRDs map[PolicyCrdID]PolicyCRD
	// policies maps a policy name to the policy object.
	policies map[string]Policy
}

func New(dc dynamic.Interface) *PolicyManager {
	return &PolicyManager{
		dc:         dc,
		policyCRDs: make(map[PolicyCrdID]PolicyCRD),
		policies:   make(map[string]Policy),
	}
}

// Init will construct a local cache of all Policy CRDs and Policy Resources.
func (p *PolicyManager) Init(ctx context.Context) error {
	allCRDs, err := fetchCRDs(ctx, p.dc)
	if err != nil {
		return err
	}
	for _, crd := range allCRDs {
		policyCRD := PolicyCRD{crd}
		// Check if the CRD is a Gateway Policy CRD
		if policyCRD.IsValid() {
			p.policyCRDs[policyCRD.ID()] = policyCRD
		}
	}

	allPolicies, err := fetchPolicies(ctx, p.dc, p.policyCRDs)
	if err != nil {
		return err
	}
	for _, unstrucutredPolicy := range allPolicies {
		p.AddPolicy(unstrucutredPolicy)
	}

	return nil
}

func (p *PolicyManager) PoliciesAttachedTo(objRef ObjRef) []Policy {
	var result []Policy
	for _, policy := range p.policies {
		if policy.IsAttachedTo(objRef) {
			result = append(result, policy)
		}
	}
	return result
}

func (p *PolicyManager) GetCRDs() []PolicyCRD {
	var result []PolicyCRD
	for _, policyCRD := range p.policyCRDs {
		result = append(result, policyCRD)
	}
	return result
}

func (p *PolicyManager) GetPolicies() []Policy {
	var result []Policy
	for _, policy := range p.policies {
		result = append(result, policy)
	}
	return result
}

func (p *PolicyManager) GetPolicy(namespacedName string) (Policy, bool) {
	policy, ok := p.policies[namespacedName]
	return policy, ok
}

func (p *PolicyManager) AddPolicy(unstrucutredPolicy unstructured.Unstructured) error {
	policy, err := PolicyFromUnstructured(unstrucutredPolicy, p.policyCRDs)
	if err != nil {
		return err
	}
	p.policies[unstrucutredPolicy.GetNamespace()+"/"+unstrucutredPolicy.GetName()] = policy
	return nil
}

// fetchCRDs will fetch all CRDs from the API Server
func fetchCRDs(ctx context.Context, dc dynamic.Interface) ([]apiextensionsv1.CustomResourceDefinition, error) {
	gvr := schema.GroupVersionResource{Group: "apiextensions.k8s.io", Version: "v1", Resource: "customresourcedefinitions"}
	unstructuredCRDs, err := dc.Resource(gvr).List(ctx, metav1.ListOptions{})
	if err != nil {
		return []apiextensionsv1.CustomResourceDefinition{}, fmt.Errorf("failed to list CRDs: %v", err)
	}

	crds := &apiextensionsv1.CustomResourceDefinitionList{}
	if err := runtime.DefaultUnstructuredConverter.FromUnstructured(unstructuredCRDs.UnstructuredContent(), crds); err != nil {
		return []apiextensionsv1.CustomResourceDefinition{}, fmt.Errorf("failed to convert unstructured CRDs to structured: %v", err)
	}

	return crds.Items, nil
}

// fetchPolicies will fetch all policy resources corresponding to the CRDs
// present in policyCRDs.
func fetchPolicies(ctx context.Context, dc dynamic.Interface, policyCRDs map[PolicyCrdID]PolicyCRD) ([]unstructured.Unstructured, error) {
	var result []unstructured.Unstructured

	for _, policyCRD := range policyCRDs {
		gvr := schema.GroupVersionResource{
			Group:    policyCRD.crd.Spec.Group,
			Version:  policyCRD.crd.Spec.Versions[0].Name,
			Resource: policyCRD.crd.Spec.Names.Plural, // CRD Kinds directly map to the Resource.
		}

		var policies *unstructured.UnstructuredList
		var err error
		if policyCRD.IsClusterScoped() {
			policies, err = dc.Resource(gvr).List(ctx, metav1.ListOptions{})
		} else {
			// For a namespace-scoped resource, fetch policies from ALL namespaces by
			// passing an empty "" namespace.
			policies, err = dc.Resource(gvr).Namespace("").List(ctx, metav1.ListOptions{})
		}
		if err != nil {
			return result, err
		}

		result = append(result, policies.Items...)
	}

	return result, nil
}

// PolicyCrdID has the structurued "<CRD Kind>.<CRD Group>"
type PolicyCrdID string

type PolicyCRD struct {
	crd apiextensionsv1.CustomResourceDefinition
}

// ID returns a unique identifier for this PolicyCRD.
func (p PolicyCRD) ID() PolicyCrdID {
	return PolicyCrdID(p.crd.Spec.Names.Kind + "." + p.crd.Spec.Group)
}

// IsValid return true if the PolicyCRD satisfies requirements for qualifying as
// a Gateway Policy CRD.
func (p PolicyCRD) IsValid() bool {
	return p.IsInherited() || p.IsDirect() || p.crd.GetLabels()[gatewayv1alpha2.PolicyLabelKey] == "true"
}

func (p PolicyCRD) IsInherited() bool {
	return strings.ToLower(p.crd.GetLabels()[gatewayv1alpha2.PolicyLabelKey]) == "inherited"
}

func (p PolicyCRD) IsDirect() bool {
	return strings.ToLower(p.crd.GetLabels()[gatewayv1alpha2.PolicyLabelKey]) == "direct"
}

func (p PolicyCRD) CRD() *apiextensionsv1.CustomResourceDefinition {
	return p.crd.DeepCopy()
}

// IsClusterScoped returns true if the CRD is cluster scoped. Such policies can
// be used to target a cluster scoped resource like GatewayClass.
func (p PolicyCRD) IsClusterScoped() bool {
	return p.crd.Spec.Scope == apiextensionsv1.ClusterScoped
}

type Policy struct {
	u unstructured.Unstructured
	// targetRef references the target object this policy is attached to. This
	// only makes sense in case of a directly-attached-policy, or an
	// unmerged-inherited-policy.
	targetRef ObjRef
	// Indicates whether the policy is supposed to be "inherited" (as opposed to
	// "direct").
	inherited bool
}

type ObjRef struct {
	Group     string `json:",omitempty"`
	Kind      string `json:",omitempty"`
	Name      string `json:",omitempty"`
	Namespace string `json:",omitempty"`
}

func PolicyFromUnstructured(u unstructured.Unstructured, policyCRDs map[PolicyCrdID]PolicyCRD) (Policy, error) {
	result := Policy{u: u}

	// Identify targetRef of Policy.
	type genericPolicy struct {
		metav1.TypeMeta   `json:",inline"`
		metav1.ObjectMeta `json:"metadata,omitempty"`
		Spec              struct {
			TargetRef gatewayv1alpha2.PolicyTargetReference
		}
	}
	structuredPolicy := &genericPolicy{}
	if err := runtime.DefaultUnstructuredConverter.FromUnstructured(u.UnstructuredContent(), structuredPolicy); err != nil {
		return Policy{}, fmt.Errorf("failed to convert unstructured policy resource to structured: %v", err)
	}
	result.targetRef = ObjRef{
		Group:     string(structuredPolicy.Spec.TargetRef.Group),
		Kind:      string(structuredPolicy.Spec.TargetRef.Kind),
		Name:      string(structuredPolicy.Spec.TargetRef.Name),
		Namespace: structuredPolicy.GetNamespace(),
	}
	if result.targetRef.Namespace == "" {
		result.targetRef.Namespace = "default"
	}
	if structuredPolicy.Spec.TargetRef.Namespace != nil {
		result.targetRef.Namespace = string(*structuredPolicy.Spec.TargetRef.Namespace)
	}

	// Get the CRD corresponding to this policy object.
	policyCRD, ok := policyCRDs[result.PolicyCrdID()]
	if !ok {
		return Policy{}, fmt.Errorf("unable to find CRD corresponding to policy object")
	}
	result.inherited = policyCRD.IsInherited()

	return result, nil
}

func (p Policy) Name() string {
	return fmt.Sprintf("%v/%v/%v", p.PolicyCrdID(), p.u.GetNamespace(), p.u.GetName())
}

// PolicyCrdID returns a unique identifier for the CRD of this policy.
func (p Policy) PolicyCrdID() PolicyCrdID {
	return PolicyCrdID(p.u.GetObjectKind().GroupVersionKind().Kind + "." + p.u.GetObjectKind().GroupVersionKind().Group)
}

func (p Policy) TargetRef() ObjRef {
	return p.targetRef
}

func (p Policy) IsInherited() bool {
	return p.inherited
}

func (p Policy) IsDirect() bool {
	return !p.inherited
}

func (p Policy) IsAttachedTo(objRef ObjRef) bool {
	if p.targetRef.Kind == "Namespace" && p.targetRef.Name == "" {
		p.targetRef.Name = "default"
	}
	if objRef.Kind == "Namespace" && objRef.Name == "" {
		objRef.Name = "default"
	}
	if p.targetRef.Kind != "Namespace" && p.targetRef.Namespace == "" {
		p.targetRef.Namespace = "default"
	}
	if objRef.Kind != "Namespace" && objRef.Namespace == "" {
		objRef.Namespace = "default"
	}
	return p.targetRef == objRef
}

func (p Policy) Unstructured() *unstructured.Unstructured {
	return &p.u
}

func (p Policy) DeepCopy() Policy {
	clone := Policy{
		u:         *p.u.DeepCopy(),
		targetRef: p.targetRef,
		inherited: p.inherited,
	}
	return clone
}

func (p Policy) Spec() map[string]interface{} {
	spec, ok, err := unstructured.NestedFieldCopy(p.u.UnstructuredContent(), "spec")
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
	if !p.IsInherited() {
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

func (p Policy) MarshalJSON() ([]byte, error) {
	effectiveSpec, err := p.EffectiveSpec()
	if err != nil {
		return nil, err
	}
	return json.Marshal(effectiveSpec)
}
