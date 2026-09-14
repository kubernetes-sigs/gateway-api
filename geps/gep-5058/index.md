---
title: "GEP-5058: Support multiple targets in PolicyStatus"
---

* Issue: [#5058](https://github.com/kubernetes-sigs/gateway-api/issues/5058)
* Status: Provisional

## Summary

This GEP proposes a refinement to the Gateway API Policy Attachment status model to support reporting distinct status conditions for each individual target resource of a Policy. Currently, if a policy targets multiple resources (e.g., via `targetRefs`) that share a common ancestor (e.g., a `Gateway`), their statuses are conflated because the status is keyed only by the ancestor and the controller. This proposal introduces a nested `targets` list within `PolicyAncestorStatus` to provide granular, per-target status reporting while maintaining backward compatibility.

## Motivation

The Policy Attachment model (established in [GEP-713]) allows policies to target multiple resources to reduce object duplication. For example, a single `BackendTLSPolicy` can target multiple `Services` using the `targetRefs` field.

However, the current `PolicyStatus` schema only allows reporting status per `Ancestor` (typically a `Gateway`) and `ControllerName`:

```go
type PolicyStatus struct {
    Ancestors []PolicyAncestorStatus `json:"ancestors"`
}

type PolicyAncestorStatus struct {
    AncestorRef ParentReference     `json:"ancestorRef"`
    ControllerName GatewayController `json:"controllerName"`
    Conditions []metav1.Condition   `json:"conditions,omitempty"`
}
```

### The Problem

If a `BackendTLSPolicy` targets `Service A` (valid) and `Service B` (non-existent), and both services are share the same `Gateway` (Ancestor) programmed by the same controller:

1. The controller must write status for both targets.
2. Since both targets share the same `AncestorRef` (the Gateway) and `ControllerName`, they map to the **same entry** in the `Ancestors` slice and share `Conditions` list.
3. `Conditions` need to be unique for an `Ancestor` and the controller cannot report disting status for each TargetRef. One status will overwrite the other, or they must be merged into a confusing hybrid status.

This limitation is so severe that the `BackendTLSPolicy` spec currently warns implementations **not** to support more than one `targetRef` until this is resolved.

### Goals

* Allow controllers to report independent status conditions for each target of a policy.
* Provide a clear UX for users to identify which specific target failed or succeeded.
* Maintain backward compatibility with existing `v1` Policy APIs.
* Ensure the design is generic enough to apply to all Direct and Inherited policies.

### Non-Goals

* Changing the policy attachment mechanism itself (keeping `targetRefs` as the way to bind policies).
* Redesigning the entire Policy Attachment model.
* Droping the ability to attach a Policy to multiple targets.

---

## Proposal

We propose restructuring `PolicyAncestorStatus` to introduce a nested `targets` list. Each entry in this list will represent the status of the policy with respect to a specific target, under the context of the given ancestor. Introduce `NamespacedTargetReferenceWithSectionName` to report status per target.

### API Changes

We will introduce a new `PolicyTargetStatus` struct and update `PolicyAncestorStatus` in `apis/v1/policy_types.go`:

```go
// PolicyAncestorStatus describes the status of a route with respect to an
// associated Ancestor.
type PolicyAncestorStatus struct {
	// AncestorRef corresponds with a ParentRef in the spec that this
	// PolicyAncestorStatus struct describes the status of.
	// +required
	AncestorRef ParentReference `json:"ancestorRef"`

	// ControllerName is a domain/path string that indicates the name of the
	// controller that wrote this status.
	// +required
	ControllerName GatewayController `json:"controllerName"`

	// Conditions describes the status of the Policy with respect to the given Ancestor.
	//
	// Deprecated: Use Targets instead to support granular status per target.
	// If Targets is populated, this field SHOULD be ignored by consumers, or
	// set to a summary condition by the controller.
	//
	// +optional
	// +listType=map
	// +listMapKey=type
	// +kubebuilder:validation:MaxItems=8
	Conditions []metav1.Condition `json:"conditions,omitempty"`

	// Targets describes the status of the Policy with respect to each target
	// selected by the policy. This MUST be used when the policy targets
	// multiple resources.
	//
	// +optional
	// +listType=atomic
	// +kubebuilder:validation:MaxItems=16
	Targets []PolicyTargetStatus `json:"targets,omitempty"`
}

// PolicyTargetStatus describes the status of a policy with respect to a specific target.
type PolicyTargetStatus struct {
	// TargetRef identifies the target resource this status applies to.
	// +required
	TargetRef NamespacedTargetReferenceWithSectionName `json:"targetRef"`

	// Conditions describes the status of the Policy with respect to this target.
// +optional
// +listType=map
	// +listMapKey=type
	// +kubebuilder:validation:MinItems=1
	// +kubebuilder:validation:MaxItems=8
	Conditions []metav1.Condition `json:"conditions"`
}

// NamespacedPolicyTargetReferenceWithSectionName identifies an API object which can be targetted by a SectionName.
//
// Note: This should only be used for Policies Status reporting.
type NamespacedPolicyTargetReferenceWithSectionName struct {
	NamespacedPolicyTargetReference `json:",inline"`

	// SectionName is the name of a section within the target resource. When
	// unspecified, this targetRef targets the entire resource.
	//
	// +optional
	SectionName *SectionName `json:"sectionName,omitempty"`
}
```

### Explanation of Design Choices

1. **`NamespaceTargetReferenceWithSectionName` for `TargetRef`:**To ensure all policies can utilize this status, a new type will be introduced. Currently, `BackendTLSPolicy` supports `LocalPolicyTargetReferenceWithSectionName` and `XBackendTrafficPolicy` supports `LocalPolicyTargetReference`, but `NamespacedPolicyTargetReference` is also defined. Therefore, we propose using a merged type that combines all supported target types.
2. **`Targets` is `listType=atomic`:** Since `TargetRef` is a struct, it cannot be used as a `listMapKey` in Kubernetes. Therefore, the `Targets` list must be `atomic`. This is acceptable because the entire `PolicyAncestorStatus` block is owned by a single controller (identified by `ControllerName`), preventing multi-controller write conflicts on the `Targets` list.
3. **`Conditions` inside `Targets` is `listType=map`:** This allows standard Kubernetes condition merging behavior (keyed by `type`) within the scope of a single target.

---

## YAML Example

Consider a `BackendTLSPolicy` targeting two services: `backend-ok` (valid) and `backend-fail` (non-existent).

```
apiVersion: gateway.networking.k8s.io/v1
kind: BackendTLSPolicy
metadata:
  name: multi-target-policy
  namespace: prod
spec:
  targetRefs:
  - group: ""
    kind: Service
    name: backend-ok
    sectionName: https
  - group: ""
    kind: Service
    name: backend-fail
  validation:
    hostname: "backend.example.com"
    # ... CA cert refs ...
status:
  ancestors:
  - ancestorRef:
      group: gateway.networking.k8s.io
      kind: Gateway
      name: prod-gateway
      namespace: prod
    controllerName: example.com/gateway-controller
    # Top-level conditions are deprecated/summary
    conditions:
    - type: Accepted
      status: "False"
      reason: "PartiallyAccepted"
      message: "Policy applied to some but not all targets."
    # Granular status per target
    targets:
    - targetRef:
        group: ""
        kind: Service
        name: backend-ok
	 namespace: prod
        sectionName: https
      conditions:
      - type: Accepted
        status: "True"
        reason: Accepted
      - type: Programmed
        status: "True"
        reason: Programmed
    - targetRef:
        group: ""
        kind: Service
        name: backend-fail
      	 namespace: prod
      conditions:
      - type: Accepted
        status: "False"
        reason: TargetNotFound
        message: "Service \"backend-fail\" not found."
```

---

## Backward Compatibility & Upgrade Path

### Controller (Writer) Behavior

* **Old Controllers:** Will continue to write only to the top-level `Conditions` field in `PolicyAncestorStatus`. They will likely only support a single target.
* **New Controllers:**
  * MUST populate the `Targets` list if the policy targets multiple resources.
  * SHOULD populate both `Targets` and the deprecated top-level `Conditions` (as a summary) to remain compatible with old clients.
  * If the policy targets a single resource, they may choose to populate both for transition.

### Client (Reader) Behavior

* **Old Clients (e.g., old dashboards, CLI tools):** Will only read the top-level `Conditions` field. They will see the summary status (e.g., `PartiallyAccepted` or the status of the first target).
* **New Clients:**
  * SHOULD check if `Targets` is populated.
  * If `Targets` is present, they MUST use it as the source of truth for per-target status and ignore the top-level `Conditions` (or use it only as a high-level summary).
  * If `Targets` is empty, they MUST fall back to reading the top-level `Conditions` (assuming an older controller).

---

## Alternatives Considered

### Alternative 1: Add `TargetRef` to `PolicyAncestorStatus` Key

Instead of nesting, we could add `TargetRef` directly to `PolicyAncestorStatus`, making the list of ancestors flat:

```go
type PolicyAncestorStatus struct {
    AncestorRef ParentReference
    ControllerName GatewayController
    TargetRef LocalPolicyTargetReference // Added here
    Conditions []metav1.Condition
}
```

* **Why Rejected:**
  * **Scale Issues:** The `Ancestors` list has a `MaxItems` limit of 16\. If a policy targets 8 services, and 3 Gateways implement it, we would need $8 \\times 3 \= 24$ entries, exceeding the limit.
  * **Duplication:** It duplicates the `AncestorRef` and `ControllerName` metadata for every single target, leading to unnecessary API bloat.
  * **Semantic Misalignment:** The struct is named `PolicyAncestorStatus`; it should represent the status *with respect to the ancestor*. Nesting the targets within the ancestor status is semantically more correct.

### Alternative 2: Use a Map in Status (`map[string][]Condition`)

We could use a map where the key is a serialized representation of the target.

* **Why Rejected:**
  * Kubernetes CRD schemas do not support complex keys.
  * Serializing `TargetRef` to a string (e.g., `\"Service/backend-ok\"`) is brittle, prone to formatting inconsistencies across implementations, and makes it impossible to use CEL for validation.

### Alternative 3: Status on the Target Resource

Instead of putting status on the Policy, we could write the status back to the target resource (e.g., adding a status field to `Service` or using annotations).

* **Why Rejected:**
  * We cannot modify core Kubernetes resources like `Service`.
  * Writing to the target resource violates the "metaresource" design principle, where the policy is a one-way attachment and should not modify the target.
  * It would require the Gateway controller to have write permissions on all target resources, which is a security concern.

[GEP-713]: gep-713/index.md