# GEP-2648: Direct Policy Attachment

* Issue: [#2648](https://github.com/kubernetes-sigs/gateway-api/issues/2648)
* Status: Experimental

(See status definitions [here](overview.md#status).)

## TLDR

Describe and specify a design pattern for a class of metaresource that can
affect specific settings across a single target object.

This is a design for a _pattern_, not an API field or new object.

## WIP TODO

- Add changelog including the original PRs plus Flynn's one
- Add details about how Direct is more specific and so overrides Inherited, NOT
  MERGED
- Update example to be BackendTLSPolicy
- Specify some requirements for status (basically, do what we did for TLSBackendPolicy)


## Goals

* Specify what common properties all Direct Attached Policies MUST have
* Recommend design patterns for areas that cannot be mandated but that could
  cause problems for API designers.

## Non-Goals

* Fully specify the entire design space for Direct Policy Attachment

## Introduction

GEP-713 defines two classes of Policy Attachment: Direct and Inherited.

Direct Attached Policies (or Direct Policies) _only)_ affect the object they are
attached to; that is, the object specified in their `targetRef`.

Note that as soon as the Policy affects more objects than the referenced object,
it is an Inherited Policy.

## Direct Policy Attachment requirements in brief

The following parts of GEP-713 also apply here. Direct Policy Attachments:
- MUST be their own CRDs (e.g. `TimeoutPolicy`, `RetryPolicy` etc),
- MUST include both `spec` and `status` stanzas
- MUST have the `status` stanza include a `conditions` section using the standard
  upstream Condition type
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

As an example, the BackendTLSPolicy is specified in GEP-1897, BackendTLSPolicy -
Explicit Backend TLS Connection Configuration. This Policy attaches to the
Service object and tells Gateway API implementations what TLS settings should be
used to connect to that Service when it is used as a backend by a Route.

See GEP-1897 for all the details of this Policy object.

## Direct Policy design guidelines

With this example in mind, here are some guidelines for when to consider
using Direct Policy Attachment:

* The number or scope of objects is exactly _one_ object.
* The modifications to be made to the objects don’t have any transitive information -
  that is, the modifications MUST only affect the single object that the targeted
  metaresource is bound to, and MUST NOT have ramifications that flow beyond that
  object.
* In terms of status, it SHOULD be reasonably easy for a user to understand that
  everything is working - basically, as long as the targeted object exists, and
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

## Apply Policies to Sections of a Resource

The `sectionName` field of `targetRef` can be used to target a specific section of other resources:

* Service.Ports.Name
* xRoute.Rules.Name

For example, the RetryPolicy below applies to a RouteRule inside an HTTPRoute.

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

If a `sectionName` is specified, but does not exist on the targeted object, the Policy must fail to attach,
and the policy implementation should record a `resolvedRefs` or similar Condition in the Policy's status.

When multiple Policies of the same type target the same object, one with a `sectionName` specified, and one without,
the one with a `sectionName` is more specific, and so will have all its settings apply. The less-specific Policy will
not attach to the target.

## Interactions with other Policy objects and settings

TODO: See the discussion on https://github.com/kubernetes-sigs/gateway-api/pull/2442

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
as defined in the Go spec, a snapshot of which is included below. The canonical
representation is in the actual Go files. (At the time of writing, this is in
`apis/v1alpha2/policy_types.go`)

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
is present.

If they do, that Condition MUST have a `type` ending in `PolicyAffected` (like
`gateway.networking.k8s.io/PolicyAffected`),
and have the optional `observedGeneration` field kept up to date when the `spec`
of the Policy-attached object changes.

Implementations SHOULD use their own unique domain prefix for this Condition
`type` - it is recommended that implementations use the same domain as in the
`controllerName` field on GatewayClass (or some other implementation-unique
domain for implementations that do not use GatewayClass).

For objects that do _not_ have a `status.Conditions` field available (`Secret`
is a good example), that object SHOULD instead have an annotation of
`gateway.networking.k8s.io/PolicyAffected: true` (or with an
implementation-specific domain prefix) added instead.

Because these Conditions or annotations are namespaced per-implementation,
implementations SHOULD:
- Add the Condition or annotation if an object is policy affected when it is not
  already present
- Remove the Condition or annotation when the last policy object stops referencing
  the targeted object.

### Other Status designs

This section contains other recommendations for status designs. Note that this
is a SHOULD rather than a MUST, as this design is still not final.


#### Standard status struct

Status: Experimental

This design is not final and we invite feedback on any use of it in implementations.

Policy objects SHOULD use the upstream `PolicyAncestorStatus` struct in their
respective Status structs. Please see the included `PolicyAncestorStatus` struct,
and its use in the `BackendTLSPolicy` object for detailed examples. Included here
is a representative version.

This pattern enables different conditions to be set for different "Ancestors"
of the target resource. This is particularly helpful for policies that may be
implemented by multiple controllers or attached to resources with different
capabilities. This pattern also provides a clear view of what resources a
policy is affecting.

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

### Direct Policy Attachment

The following Policy sets the minimum TLS version required on a Gateway Listener:

```yaml
apiVersion: networking.example.io/v1alpha1
kind: TLSMinimumVersionPolicy
metadata:
  name: minimum12
  namespace: appns
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
