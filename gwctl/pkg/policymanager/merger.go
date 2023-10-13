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

	jsonpatch "github.com/evanphx/json-patch"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

// MergePoliciesOfSimilarKind will convert a slice a policies to a map of
// policies by merging policies of similar kind. The returned map will have the
// policy kind as the key.
func MergePoliciesOfSimilarKind(policies []Policy) (map[PolicyCrdID]Policy, error) {
	result := make(map[PolicyCrdID]Policy)
	for _, policy := range policies {
		policyCrdID := policy.PolicyCrdID()

		if _, ok := result[policyCrdID]; !ok {
			// Policy of kind policyCrdID doesn't already exist so simply insert it
			// into the resulting map.
			result[policyCrdID] = policy
			continue
		}

		// At this point, we know that a policy of kind policyCrdID already exists
		// so we need to merge the new policy with the existing one.

		// Merge existing policy with new policy. Reuse existing function to merge
		// policies of similar hierarchy.
		mergedPolicies, err := MergePoliciesOfSameHierarchy(
			map[PolicyCrdID]Policy{
				policyCrdID: result[policyCrdID], // Existing policy.
			},
			map[PolicyCrdID]Policy{
				policyCrdID: policy, // New policy.
			},
		)
		if err != nil {
			return nil, err
		}

		result[policyCrdID] = mergedPolicies[policyCrdID]
	}
	return result, nil
}

func MergePoliciesOfSameHierarchy(policies1, policies2 map[PolicyCrdID]Policy) (map[PolicyCrdID]Policy, error) {
	return mergePolicies(policies1, policies2, orderPolicyByPrecedence)
}

func MergePoliciesOfDifferentHierarchy(parentPolicies, childPolicies map[PolicyCrdID]Policy) (map[PolicyCrdID]Policy, error) {
	return mergePolicies(parentPolicies, childPolicies, func(a, b Policy) (Policy, Policy) { return a, b })
}

// mergePolicies will merge policies which are partitioned by their Kind.
//
// precedence function will order two policies such that the second policy
// returned will have a higher precedence.
func mergePolicies(policies1, policies2 map[PolicyCrdID]Policy, precedence func(a, b Policy) (Policy, Policy)) (map[PolicyCrdID]Policy, error) {
	result := make(map[PolicyCrdID]Policy)

	// Copy policies1 into result.
	for policyCrdID, policy := range policies1 {
		result[policyCrdID] = policy
	}

	// Merge policies2 with result.
	for policyCrdID, policy := range policies2 {
		existingPolicy, ok := result[policyCrdID]
		if !ok {
			// Policy of kind policyCrdID doesn't already exist so simply insert it
			// into the resulting map.
			result[policyCrdID] = policy
			continue
		}

		// Policy of kind policyCrdID already exists so merge them.

		lowerPolicy, higherPolicy := precedence(existingPolicy, policy)

		res, err := mergePolicy(lowerPolicy, higherPolicy)
		if err != nil {
			return nil, err
		}
		result[policyCrdID] = res
	}
	return result, nil
}

// mergePolicy will merge two policies of similar kind.
//   - overrides from parent will take precedence over the overrides from the
//     child.
//   - defaults from child will take precedence over the defaults from the
//     parent.
func mergePolicy(parent, child Policy) (Policy, error) {
	// Only policies of similar kind can be merged.
	if parent.PolicyCrdID() != child.PolicyCrdID() {
		return Policy{}, fmt.Errorf("cannot merge policies of different kind; kind1=%v, kind2=%v", parent.PolicyCrdID(), child.PolicyCrdID())
	}

	resultUnstructured, err := mergeUnstructured(parent.u.UnstructuredContent(), child.u.UnstructuredContent())
	if err != nil {
		return Policy{}, err
	}

	if parent.IsInherited() {
		// In case of an Inherited policy, the "spec.override" field of the parent
		// should take precedence over the child. So we patch the override field
		// from the parent into the result.
		override, ok, err := unstructured.NestedFieldCopy(parent.u.UnstructuredContent(), "spec", "override")
		if err != nil {
			return Policy{}, err
		}
		// If ok=false, it means "spec.override" field was missing, so we have
		// nothing to do in that case. On the other hand, ok=true means
		// "spec.override" field exists so we override the value of the parent.
		if ok {
			resultUnstructured, err = mergeUnstructured(resultUnstructured, map[string]interface{}{
				"spec": map[string]interface{}{
					"override": override,
				},
			})
			if err != nil {
				return Policy{}, err
			}
		}
	}

	result := child.DeepCopy()
	result.u.SetUnstructuredContent(resultUnstructured)
	// Merging two policies means the targetRef no longer makes any sense since
	// since they can be conflicting. So we unset the targetRef.
	result.targetRef = ObjRef{}
	return result, nil
}

func mergeUnstructured(parent, patch map[string]interface{}) (map[string]interface{}, error) {
	currentJSON, err := json.Marshal(parent)
	if err != nil {
		return nil, err
	}

	modifiedJSON, err := json.Marshal(patch)
	if err != nil {
		return nil, err
	}

	resultJSON, err := jsonpatch.MergePatch(currentJSON, modifiedJSON)
	if err != nil {
		return nil, err
	}

	result := make(map[string]interface{})
	if err := json.Unmarshal(resultJSON, &result); err != nil {
		return nil, err
	}

	return result, nil
}

// orderPolicyByPrecedence will decide the precedence of two policies as per the
// [Gateway Specification]. The second policy returned will have a higher
// precedence.
//
// [Gateway Specification]: https://gateway-api.sigs.k8s.io/geps/gep-713/#conflict-resolution
func orderPolicyByPrecedence(a, b Policy) (Policy, Policy) {
	lowerPolicy := a.DeepCopy()  // lowerPolicy will have lower precedence.
	higherPolicy := b.DeepCopy() // higherPolicy will have higher precedence.

	if lowerPolicy.u.GetCreationTimestamp() == higherPolicy.u.GetCreationTimestamp() {
		// Policies have the same creation time, so precedence is decided based
		// on alphabetical ordering.
		higherNN := fmt.Sprintf("%v/%v", higherPolicy.u.GetNamespace(), higherPolicy.u.GetName())
		lowerNN := fmt.Sprintf("%v/%v", lowerPolicy.u.GetNamespace(), lowerPolicy.u.GetName())
		if higherNN > lowerNN {
			higherPolicy, lowerPolicy = lowerPolicy, higherPolicy
		}

	} else if higherPolicy.u.GetCreationTimestamp().Time.After(lowerPolicy.u.GetCreationTimestamp().Time) {
		// Policies have difference creation time, so this will decide the precedence
		higherPolicy, lowerPolicy = lowerPolicy, higherPolicy
	}

	// At this point, higherPolicy will have precedence over lowerPolicy.
	return lowerPolicy, higherPolicy
}
