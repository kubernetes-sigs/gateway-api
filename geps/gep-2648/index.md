# GEP-2648: Direct Policy Attachment

* Issue: [#2648](https://github.com/kubernetes-sigs/gateway-api/issues/2648)
* Status: Provisional

(See status definitions [here](/geps/overview/#gep-states)

## TLDR

Describe and specify a design pattern for a class of metaresource that can
affect specific settings across a single target object.

This is a strict subset of all Policy objects that meet a set of criteria
designed to be easier to understand for users than Inherited Policy, and so to
not require solving the much harder problem of communicating Inherited Policy
to users.

This will allow the graduation of this _limited_ subcategory of Policy objects
_separately_ to solving the larger problem of communicating status for Inherited
Policy.

This is a design for a _pattern_, not an API field or new object.

!!! danger
    This GEP is in the process of being updated.
    Please see the discussion at https://github.com/kubernetes-sigs/gateway-api/discussions/2927
    and expect further changes.
    Some options under discussion there may make the distinction between Direct
    and Inherited Policies moot, which would require a rework.

## Goals

* Specify what common properties all Direct Attached Policies MUST have
* Recommend design patterns for areas that cannot be mandated but that could
  cause problems for API designers.

## Non-Goals

* Fully specify the entire design space for Direct Policy Attachment

## Introduction

GEP-713 defines two classes of Policy Attachment: Direct and Inherited.

Direct Attached Policies (or Direct Policies) _only_ affect the object they are
attached to; that is, the object specified in their `targetRef`.

Note that as soon as the Policy affects more objects than the referenced object,
it is an Inherited Policy.

## Direct Policy Attachment requirements in brief

The following parts of GEP-713 also apply here. Direct Policy Attachments:

- MUST be their own CRDs (e.g. `TimeoutPolicy`, `RetryPolicy` etc),
- MUST include both `spec` and `status` stanzas
- MUST have the `status` stanza include a `conditions` section using the standard
  upstream Condition type. This includes using reasons such as `Conflicted` (for
  when a Policy cannot be accepted because another Policy is targeting the same
  object) or `TargetNotFound` when the Policy targets a nonexistent object.
- MUST use the `targetRef` struct to specify their target resource.
- MUST follow the naming requirements (MUST be named to clearly indicate that the
kind is a Policy, and SHOULD use the `Policy` suffix at the end of the Kind and
`policies` at the end of the Resource names).
- MAY target a subsection of a resource using the `sectionName` field of the
  `targetRef` struct. If it does, it MUST abide by the rules in this document,
  and MUST clearly indicate the objects that `sectionName` may be used on in its
  documentation.

## Direct Policy Attachment

A Direct Policy Attachment is tightly bound to one instance of a particular
Kind within a single namespace (or to an instance of a single Kind at cluster scope),
and only modifies the behavior of the object that matches its binding.

As an example, the BackendTLSPolicy is specified in [GEP-1897](https://gateway-api.sigs.k8s.io/geps/gep-1897/),
BackendTLSPolicy - Explicit Backend TLS Connection Configuration. This Policy
attaches to the Service object and tells Gateway API implementations what TLS
settings should be used to connect to that Service when it is used as a backend
by a Route.

See GEP-1897 for all the details of this Policy object.

## Direct Policy design rules

In these rules, "affects" means to change properties of an object that are
relevant in objects that are stored in the storage medium of the objects in the
hierarchy. If the combination of objects in the object hierarchy cause the creation
of some other object in the object store (usually, the Kubernetes API),
differences in that object do not count as "affecting" for the purposes of these
rules. For example, if you take a GatewayClass -> Gateway -> Route
-> Service hierarchy, and attach a Policy somewhere, which leads to the creation
of a DataplaneConfig object that will be different because of the inclusion of
the Policy object, the DataplaneConfig object does not affect if the Policy is
a Direct one or not. This is because _a user can understand the state of the
hierarchy by looking at all the objects in the hierarchy_. DataplaneConfig is
_outside_ the hierarchy in terms of understanding the state of the Policy.
Direct Attacthed Policy is intended as a way to _manage the complexity_ of
Policy objects and allow a _limited_ set of Policies to follow vastly more
simple design patterns _if they meet a set of criteria_.

With that background and the previous example in mind, here are some rules for
when a Policy is a Direct Attached Policy:

* The Policy can only be attached at exactly _one_ layer in the hierarchy. Any Policy
  that can be attached at multiple levels must necessarily have some defaulting
  behavior in the case that two of the same kind are attached at different points
  in the same hierarchy, so it cannot be Direct.
* The Policy can have effects only at the layer it attaches to. That is, the
  behavior modifications MUST only affect the single object that the targeted
  metaresource is bound to, and MUST NOT have ramifications that flow beyond that
  object. No attaching a Policy to a Gateway and affecting settings in Routes or
  backends. If a Direct Attached Policy attaches to an object, it can only affect
  properties _of that object_ and _at that layer_ of the hierarchy.
* The Policy can have effects only on the object it attaches to within the layer
  of the hierarchy it attaches to. A Direct Attached Policy cannot affect sibling
  objects in the same hierarchy directly.
* In terms of status, it SHOULD be reasonably easy for a user to understand that
  everything is working - basically, as long as the targeted object exists,
  the modifications are valid, the metaresource is valid, and this should be
  straightforward to communicate in one or two Conditions. The `status` stanza
  in BackendTLSPolicy is an example of one of the recommended ways to achieve this.
* Direct Policy Attachment SHOULD only be used to target objects in the same
  namespace as the Policy object. Allowing cross-namespace references brings in
  significant security concerns, and/or difficulties about merging cross-namespace
  policy objects. Notably, Mesh use cases may need to do something like this for
  consumer policies, but in general, Policy objects that modify the behavior of
  things outside their own namespace should be avoided unless it uses a handshake
  of some sort, where the things outside the namespace can opt–out of the behavior.
  (Notably, this is the design that we used for ReferenceGrant).

## Target References
### Cross Namespace

In all cases, Gateway API policies should only have an effect on the namespace
they exist within. In the case of policies that could apply to mesh
implementations, it may be desirable to have a policy that affects traffic
originating from the local namespace but going to a separate namespace. Unless
that specific case is desired, all policy target refs should be local and
exclude the "namespace" field.

### Multiple

In some cases, it may be desirable for a policy to target more than one resource
at a time. For example, a policy may apply to different variations of what is
effectively the same Service (store, store-blue, and store-green). If this is
desired, a policy can choose to support a `targetRefs` list instead of a
singular `targetRef` field. This list can have a maximum of 16 entries, though
it may be desirable to start with a lower limit depending on the policy.

#### Migration from Single to Multiple Targets

Existing policies with a single `targetRef` may want to transition to supporting
multiple `targetRefs`. To accomplish this, we recommend adding CEL validation
to your CRD to allow only one of the fields to be set. Users will be able to
set `targetRefs` in the same update that they unset `targetRef`.

### Section Names

The `sectionName` field of `targetRef` can be used to target a specific section
of other resources, for example:

* Service.Ports.Name
* Gateway.Listeners.Name
* HTTPRoute.Rules.Name (once they are added in [GEP-995](https://gateway-api.sigs.k8s.io/geps/gep-995), implementation tracked by [#2985](https://github.com/kubernetes-sigs/gateway-api/pull/2985))

Implementations SHOULD NOT use the name of a `backendRef` for applying Policy,
since the `backendRef` both is not guaranteed to be unique across a Route's rules,
and also the `backendRef` is also a link to another object. Target the policy
at the thing the `backendRef` points to instead.

For example, the RetryPolicy below applies to a RouteRule inside an HTTPRoute.
(or rather, it will when GEP-995 merges).

```yaml
apiVersion: gateway.networking.k8s.io/v1alpha2
kind: HTTPRoute
metadata:
  name: http-app-1
  labels:
    app: foo
spec:
  hostnames:
  - "foo.com"
  rules:
  - name: bar
    matches:
    - path:
        type: Prefix
        value: /bar
    backendRefs:
    - name: my-service1
      port: 8080
---
apiVersion: networking.acme.io/v1alpha2
kind: RetryPolicy
metadata:
  name: foo
spec:
  maxRetries: 5
  targetRef:
    name: http-app-1
    group: gateway.networking.k8s.io
    kind: HTTPRoute
    sectionName: bar
```

If a `sectionName` is specified, but does not exist on the targeted object, the
Policy must fail to attach, and the policy implementation should record a
`resolvedRefs` failure or similar Condition in the Policy's status.

When multiple Policies of the same type target the same object, one with a
`sectionName` and one without, the more specific policy (i.e., the one with a
`sectionName`) will have its entire `spec` applied to the named section.
The less specific policy will also have its `spec` applied to the target but
MUST not affect the named section. The less specific policy will have its `spec`
applied to all other sections of the target that are not targeted by any other
more specific policies.

When more than one object matches the same object _and_ `sectionName`, the usual
conflict-resolution rules (as defined in GEP-713 should be used). These boil down
to "oldest by creation date wins".

## User discoverability and status

### Standard label on CRD objects

Each CRD that defines a Direct Policy object MUST include a label that specifies that
it is a Policy object, and that label MUST specify that the object is a `direct`
one.

The label is `gateway.networking.k8s.io/policy: direct`.

This solution is intended to allow both users and tooling to identify which CRDs
in the cluster should be treated as Policy objects, and so is intended to help
with discoverability generally. It will also be used by the forthcoming `kubectl`
plugin.

### Conditions

Implementations using Policy objects MUST include a `spec` and `status` stanza,
and the `status` stanza MUST contain a `conditions` stanza, using the standard
Condition format.

Policy authors should consider namespacing the `conditions` stanza with a
`controllerName`, as in Route status, if more than one implementation will be
reconciling the Policy type.

#### On `Policy` objects

Each Direct Attached Policy MUST populate the `Accepted` condition and reasons
as defined in the PolicyCondition API, a snapshot of which is included below.
The canonical representation is in the actual Go files. (At the time of writing,
this is in `apis/v1alpha2/policy_types.go`)

```go
// PolicyConditionType is a type of condition for a policy.
type PolicyConditionType string

// PolicyConditionReason is a reason for a policy condition.
type PolicyConditionReason string

const (
  // PolicyConditionAccepted indicates whether the policy has been accepted or rejected
  // by a targeted resource, and why.
  //
  // Possible reasons for this condition to be True are:
  //
  // * "Accepted"
  //
  // Possible reasons for this condition to be False are:
  //
  // * "Conflicted"
  // * "Invalid"
  // * "TargetNotFound"
  //
  PolicyConditionAccepted PolicyConditionType = "Accepted"

  // PolicyReasonAccepted is used with the "Accepted" condition when the policy has been
  // accepted by the targeted resource.
  PolicyReasonAccepted PolicyConditionReason = "Accepted"

  // PolicyReasonConflicted is used with the "Accepted" condition when the policy has not
  // been accepted by a targeted resource because there is another policy that targets the same
  // resource and a merge is not possible.
  PolicyReasonConflicted PolicyConditionReason = "Conflicted"

  // PolicyReasonInvalid is used with the "Accepted" condition when the policy is syntactically
  // or semantically invalid.
  PolicyReasonInvalid PolicyConditionReason = "Invalid"

  // PolicyReasonTargetNotFound is used with the "Accepted" condition when the policy is attached to
  // an invalid target resource
  PolicyReasonTargetNotFound PolicyConditionReason = "TargetNotFound"
)
```

#### On targeted resources

Implementations that use Direct Policy objects SHOULD put a Condition into
`status.Conditions` of any objects affected by a Direct Policy, if that field
is present. Ideally, there should be a set of Conditions that can be namespaced
by the implementing controller, but if that is not posisble, use the guidance below.

If they do, that Condition MUST have a `type` ending in `PolicyAffected` (like
`gateway.networking.k8s.io/PolicyAffected`),
and have the optional `observedGeneration` field kept up to date when the `spec`
of the Policy-attached object changes.

Implementations SHOULD use their own unique domain prefix for this Condition
`type` - it is recommended that implementations use the same domain as in the
`controllerName` field on GatewayClass (or some other implementation-unique
domain for implementations that do not use GatewayClass).

For objects that do _not_ have a `status.Conditions` field available (`Secret`
is a good example), that object SHOULD instead have a label of
`gateway.networking.k8s.io/PolicyAffected: true` (or with an
implementation-specific domain prefix) added instead.

Because these Conditions or labels are namespaced per-implementation,
implementations SHOULD:

- Add the Condition or label if an object is policy affected when it is not
  already present
- Remove the Condition or label when the last policy object stops referencing
  the targeted object.

### Other Status designs

This section contains other recommendations for status designs. Note that this
is a SHOULD rather than a MUST, as this design is still not final.


#### Standard status struct

This design is not final and we invite feedback on any use of it in implementations.

Policy objects SHOULD use the upstream `PolicyAncestorStatus` struct in their
respective Status structs. Please see the included `PolicyAncestorStatus` struct,
and its use in the `BackendTLSPolicy` object for detailed examples. Included here
is a representative version.

This pattern enables different conditions to be set for different "Ancestors"
of the target resource. This is particularly helpful for policies that may be
implemented by multiple controllers or attached to resources with different
effects or capabilities. For example a Policy that could attach to Route or Service
to set load balancing properties may be reconciled by multiple controllers, and
so needs further namespacing of its status. This pattern also provides a clear
view of what resources a policy is affecting.

For the best integration with community tooling and consistency across
the broader community, we recommend that all implementations transition
to Policy status with this kind of nested structure.

This is an `Ancestor` status rather than a `Parent` status, as in the Route status
because for Policy attachment, the relevant object may or may not be the direct
parent.

For example, `BackendTLSPolicy` directly attaches to a Service, which may be included
in multiple Routes, in multiple Gateways. However, for many implementations,
the status of the `BackendTLSPolicy` will be different only at the Gateway level,
so Gateway is the relevant Ancestor for the status.

Each Gateway that has a Route that includes a backend with an attached `BackendTLSPolicy`
MUST have a separate `PolicyAncestorStatus` section in the `BackendTLSPolicy`'s
`status.ancestors` stanza, which mandates that entries must be distinct using the
combination of the `AncestorRef` and the `ControllerName` fields as a key.

See [GEP-1897][gep-1897] for the exact details. A snapshot of the Go code is
included here for reference, but the canonical representation is in the code
itself (at the time of writing, this is in `apis/v1alpha2/policy_types.go`).

[gep-1897]: /geps/gep-1897

```go
// PolicyAncestorStatus describes the status of a route with respect to an
// associated Ancestor.
//
// Ancestors refer to objects that are either the Target of a policy or above it in terms
// of object hierarchy. For example, if a policy targets a Service, an Ancestor could be
// a Route or a Gateway.

// In the context of policy attachment, the Ancestor is used to distinguish which
// resource results in a distinct application of this policy. For example, if a policy
// targets a Service, it may have a distinct result per attached Gateway.
//
// Policies targeting the same resource may have different effects depending on the
// ancestors of those resources. For example, different Gateways targeting the same
// Service may have different capabilities, especially if they have different underlying
// implementations.
//
// For example, in BackendTLSPolicy, the Policy attaches to a Service that is
// used as a backend in a HTTPRoute that is itself attached to a Gateway.
// In this case, the relevant object for status is the Gateway, and that is the
// ancestor object referred to in this status.
//
// Note that a Target of a Policy is also a valid Ancestor, so for objects where
// the Target is the relevant object for status, this struct SHOULD still be used.
type PolicyAncestorStatus struct {
	// AncestorRef corresponds with a ParentRef in the spec that this
	// RouteParentStatus struct describes the status of.
	AncestorRef ParentReference `json:"ancestorRef"`

	// ControllerName is a domain/path string that indicates the name of the
	// controller that wrote this status. This corresponds with the
	// controllerName field on GatewayClass.
	//
	// Example: "example.net/gateway-controller".
	//
	// The format of this field is DOMAIN "/" PATH, where DOMAIN and PATH are
	// valid Kubernetes names
	// (https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#names).
	//
	// Controllers MUST populate this field when writing status. Controllers should ensure that
	// entries to status populated with their ControllerName are cleaned up when they are no
	// longer necessary.
	ControllerName GatewayController `json:"controllerName"`

	// Conditions describes the status of the Policy with respect to the given Ancestor.
	//
	// +listType=map
	// +listMapKey=type
	// +kubebuilder:validation:MinItems=1
	// +kubebuilder:validation:MaxItems=8
	Conditions []metav1.Condition `json:"conditions,omitempty"`
}


// PolicyStatus defines the common attributes that all Policies SHOULD include
// within their status.
type PolicyStatus struct {
	// Ancestors is a list of ancestor resources (usually Gateways) that are
	// associated with the route, and the status of the route with respect to
	// each ancestor. When this route attaches to a parent, the controller that
	// manages the parent and the ancestors MUST add an entry to this list when
	// the controller first sees the route and SHOULD update the entry as
	// appropriate when the relevant ancestor is modified.
	//
	// Note that choosing the relevant ancestor is left to the Policy designers;
	// an important part of Policy design is designing the right object level at
	// which to namespace this status.
	//
	// Note also that implementations MUST ONLY populate ancestor status for
	// the Ancestor resources they are responsible for. Implementations MUST
	// use the ControllerName field to uniquely identify the entries in this list
	// that they are responsible for.
	//
	// A maximum of 32 ancestors will be represented in this list. An empty list
	// means the Policy is not relevant for any ancestors.
	//
	// +kubebuilder:validation:MaxItems=32
	Ancestors []PolicyAncestorStatus `json:"ancestors"`
}
```

## Examples

### Hypothetical TLSMinimumVersionPolicy

The following hypothetical Policy sets the minimum TLS version required on a
Gateway Listener:

```yaml
apiVersion: networking.example.io/v1alpha1
kind: TLSMinimumVersionPolicy
metadata:
  name: minimum12
  namespace: appns
  labels:
    "gateway.networking.k8s.io/policy": "direct"
spec:
  minimumTLSVersion: 1.2
  targetRef:
    name: internet
    group: gateway.networking.k8s.io
    kind: Gateway
```

Note that because there is no version controlling the minimum TLS version in the
Gateway `spec`, this is an example of a non-field Policy.

This is an example of a Direct Attached Policy because it affects a field on the
Gateway itself, rather than fields or behavior associated with Routes attached
to that Gateway.

### BackendTLSPolicy

BackendTLSPolicy, introduced in [GEP-1897](https://gateway-api.sigs.k8s.io/geps/gep-1897/)
allows backends to set the TLS details that a Gateway implementation must use
to connect to that backend.

It does this using a Direct Attached Policy that attaches to a Service.

Work on this Policy is still ongoing, please see GEP-1897 for details.
