# GEP-1713: ListenerSets - Standard Mechanism to Merge Multiple Gateways

* Issue: [#1713](https://github.com/kubernetes-sigs/gateway-api/issues/1713)
* Status: Experimental

(See [status definitions](../overview.md#gep-states).)

## Introduction

The `Gateway` Resource is a point of contention since it is the only place to attach listeners with certificates. We propose a new resource called `ListenerSet` to allow a shared list of listeners to be attached to a single `Gateway`.

## Goals
- Define a mechanism to merge listeners into a single `Gateway`
- Attaching listeners to `Gateways` in different namespaces
- Standardize merging multiple lists of Listeners together ([\#1863](https://github.com/kubernetes-sigs/gateway-api/pull/1863))
- Increase the number of Gateway Listeners that are supported ([\#2869](https://github.com/kubernetes-sigs/gateway-api/issues/2869))

## Future Potential Goals (Beyond the GEP)

From [Gateway Hierarchy Brainstorming](https://docs.google.com/document/d/1qj7Xog2t2fWRuzOeTsWkabUaVeOF7_2t_7appe8EXwA/edit#heading=h.w311n4l5qmwk):

- Provide a mechanism for third party components to generate listeners and attach them to a Gateway ([\#1863](https://github.com/kubernetes-sigs/gateway-api/pull/1863))
- Delegate TLS certificate management to App Owners and/or different namespaces ([\#102](https://github.com/kubernetes-sigs/gateway-api/issues/102), [\#103](https://github.com/kubernetes-sigs/gateway-api/issues/103))
- Delegate domains to different namespaces, but allow those namespace to define TLS and routing configuration within those namespaces with Gateway-like resources ([\#102](https://github.com/kubernetes-sigs/gateway-api/issues/102), [\#103](https://github.com/kubernetes-sigs/gateway-api/issues/103))
- Enable admins to delegate SNI-based routing for TLS passthrough to other teams and/or namespaces ([\#3177](https://github.com/kubernetes-sigs/gateway-api/discussions/3177)) (Remove TLSRoute)
- Simplify L4 routing by removing at least one of the required layers (Gateway \-\> Route \-\> Service)
- Delegate routing to namespaces based on path prefix (previously known as [Route delegation](https://github.com/kubernetes-sigs/gateway-api/issues/1058))
- Static infrastructure attachment ([\#3103](https://github.com/kubernetes-sigs/gateway-api/discussions/3103\#discussioncomment-9678523))

## Use Cases & Motivation

Knative generates on demand per-service certificates using HTTP-01 challenges.
There can be O(1000) Knative `Services` in the cluster which means we have O(1000) distinct certificates.
Thus updating a single `Gateway` resource with this many certificates is a contention point and inhibits horizontal scaling of our controllers.
[Istio Ambient](https://istio.io/v1.15/blog/2022/introducing-ambient-mesh/), similarly, creates a listener per Kubernetes service.

More broadly, large scale gateway users often expose `O(1000)` domains, but are currently limited by the maximum of 64 `listeners`.

The [spec currently has language](https://github.com/kubernetes-sigs/gateway-api/blob/541e9fc2b3c2f62915cb58dc0ee5e43e4096b3e2/apis/v1beta1/gateway_types.go#L76-L78) to indicate implementations `MAY` merge `Gateways` resources but does not define any specific requirements for how that should work.


## Feature Details

We define `ListenerSet` as the name of the feature outlined in this GEP.
The feature will be part of the experimental channel, which implementations can choose to support. All the `MUST` requirements in this document apply to implementations that choose to support this feature.


## API

This proposal introduces a new `ListenerSet` resource that has the ability to attach a set of listeners to multiple parent `Gateways`.

**Note**: While this API is experimental, its `Kind` will be `XListenerSet` and 
once the API is graduated to stable it will be renamed to `ListenerSet`.

### Go

```go
type GatewaySpec struct {
	...
	AllowedListeners *AllowedListeners `json:"allowedListeners"`
	...
}

type AllowedListeners struct {
	// +kubebuilder:default={from: None}
	Namespaces *ListenerNamespaces `json:"namespaces,omitempty"`
}

// ListenerNamespaces indicate which namespaces ListenerSets should be selected from.
type ListenerNamespaces struct {
	// From indicates where ListenerSets can attach to this Gateway. Possible
	// values are:
	//
	// * Same: Only ListenerSets in the same namespace may be attached to this Gateway.
	// * Selector: ListenerSets in namespaces selected by the selector may be attached to this Gateway.
	// * All: ListenerSets in all namespaces may be attached to this Gateway.
	// * None: Only listeners defined in the Gateway's spec are allowed
	//
	// +optional
	// +kubebuilder:default=None
	// +kubebuilder:validation:Enum=Same;None;Selector;All
	From *FromNamespaces `json:"from,omitempty"`

	// Selector must be specified when From is set to "Selector". In that case,
	// only ListenerSets in Namespaces matching this Selector will be selected by this
	// Gateway. This field is ignored for other values of "From".
	//
	// +optional
	Selector *metav1.LabelSelector `json:"selector,omitempty"`
}

// ListenerSet defines a set of additional listeners to attach to an existing Gateway.
type ListenerSet struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// Spec defines the desired state of ListenerSet.
	Spec ListenerSetSpec `json:"spec"`

	// Status defines the current state of ListenerSet.
	Status ListenerSetStatus `json:"status,omitempty"`
}

// ListenerSetSpec defines the desired state of a ListenerSet.
type ListenerSetSpec struct {
	// ParentRef references the Gateway that the listeners are attached to.
	ParentRef ParentGatewayReference `json:"parentRef"`

	// Listeners associated with this ListenerSet. Listeners define
	// logical endpoints that are bound on this referenced parent Gateway's addresses.
	//
	// Listeners in a `Gateway` and their attached `ListenerSets` are concatenated
	// as a list when programming the underlying infrastructure.
	//
	// <gateway:util:excludeFromCRD>
	// Listeners should be merged using the following precedence:
	//
	// 1. "parent" Gateway
	// 2. ListenerSet ordered by creation time (oldest first)
	// 3. ListenerSet ordered alphabetically by “{namespace}/{name}”.
	//
	// Regarding Conflict Management, Listeners in a ListenerSet follow the same
	// rules of Listeners on a Gateway resource.
	// 
	// Listener validation should happen within all of the ListenerSets attached to a 
	// Gateway, and the precedence of "parent Gateway" -> "oldest first" -> 
	// "alphabetically ordered" should be respected.
	// 
	// ListenerSets containing conflicting Listeners MUST set the Conflicted 
	// Condition to true and clearly indicate which Listeners are conflicted.
	// </gateway:util:excludeFromCRD>
	// 
	// +listType=map
	// +listMapKey=name
	// +kubebuilder:validation:MinItems=1
	// +kubebuilder:validation:MaxItems=64
	Listeners []ListenerEntry `json:"listeners"`

}
// ListenerEntry embodies the concept of a logical endpoint where a Gateway accepts
// network connections.
type ListenerEntry struct {
	// Name is the name of the Listener. This name MUST be unique within a
	// ListenerSet.
	//
	// Name is not required to be unique across a Gateway and ListenerSets.
	// Routes can attach to a Listener by having a ListenerSet as a parentRef
	// and setting the SectionName
	Name SectionName `json:"name"`

	// Hostname specifies the virtual hostname to match for protocol types that
	// define this concept. When unspecified, all hostnames are matched. This
	// field is ignored for protocols that don't require hostname based
	// matching.
	//
	// Implementations MUST apply Hostname matching appropriately for each of
	// the following protocols:
	//
	// * TLS: The Listener Hostname MUST match the SNI.
	// * HTTP: The Listener Hostname MUST match the Host header of the request.
	// * HTTPS: The Listener Hostname SHOULD match at both the TLS and HTTP
	//   protocol layers as described above. If an implementation does not
	//   ensure that both the SNI and Host header match the Listener hostname,
	//   it MUST clearly document that.
	//
	// For HTTPRoute and TLSRoute resources, there is an interaction with the
	// `spec.hostnames` array. When both listener and route specify hostnames,
	// there MUST be an intersection between the values for a Route to be
	// accepted. For more information, refer to the Route specific Hostnames
	// documentation.
	//
	// Hostnames that are prefixed with a wildcard label (`*.`) are interpreted
	// as a suffix match. That means that a match for `*.example.com` would match
	// both `test.example.com`, and `foo.test.example.com`, but not `example.com`.
	//
	// +optional
	Hostname *Hostname `json:"hostname,omitempty"`

	// Port is the network port. Multiple listeners may use the
	// same port, subject to the Listener compatibility rules.
    //
	// If the port is not set or specified as zero, the implementation will assign
	// a unique port. If the implementation does not support dynamic port
	// assignment, it MUST set `Accepted` condition to `False` with the
	// `UnsupportedPort` reason.
    //
	// Support: Core
	//
	// +optional
	//
	// +kubebuilder:default=0
	// +kubebuilder:validation:Minimum=0
	// +kubebuilder:validation:Maximum=65535
	Port PortNumber `json:"port,omitempty"`

	// Protocol specifies the network protocol this listener expects to receive.
	//
	// Support: Core
	Protocol ProtocolType `json:"protocol"`

	// TLS is the TLS configuration for the Listener. This field is required if
	// the Protocol field is "HTTPS" or "TLS". It is invalid to set this field
	// if the Protocol field is "HTTP", "TCP", or "UDP".
	//
	// The association of SNIs to Certificate defined in GatewayTLSConfig is
	// defined based on the Hostname field for this listener.
	//
	// The GatewayClass MUST use the longest matching SNI out of all
	// available certificates for any TLS handshake.
	//
	// Support: Core
	//
	// +optional
	TLS *GatewayTLSConfig `json:"tls,omitempty"`

	// AllowedRoutes defines the types of routes that MAY be attached to a
	// Listener and the trusted namespaces where those Route resources MAY be
	// present.
	//
	// Although a client request may match multiple route rules, only one rule
	// may ultimately receive the request. Matching precedence MUST be
	// determined in order of the following criteria:
	//
	// * The most specific match as defined by the Route type.
	// * The oldest Route based on creation timestamp. For example, a Route with
	//   a creation timestamp of "2020-09-08 01:02:03" is given precedence over
	//   a Route with a creation timestamp of "2020-09-08 01:02:04".
	// * If everything else is equivalent, the Route appearing first in
	//   alphabetical order (namespace/name) should be given precedence. For
	//   example, foo/bar is given precedence over foo/baz.
	//
	// All valid rules within a Route attached to this Listener should be
	// implemented. Invalid Route rules can be ignored (sometimes that will mean
	// the full Route). If a Route rule transitions from valid to invalid,
	// support for that Route rule should be dropped to ensure consistency. For
	// example, even if a filter specified by a Route rule is invalid, the rest
	// of the rules within that Route should still be supported.
	//
	// Support: Core
	// +kubebuilder:default={namespaces:{from: Same}}
	// +optional
	AllowedRoutes *AllowedRoutes `json:"allowedRoutes,omitempty"`
}

// ListenerSetStatus defines the observed state of a ListenerSet
type ListenerSetStatus struct {
	// Listeners provide status for each unique listener port defined in the Spec.
	//
	// +optional
	// +listType=map
	// +listMapKey=name
	// +kubebuilder:validation:MaxItems=64
	Listeners []ListenerEntryStatus `json:"listeners,omitempty"`

	// Conditions describe the current conditions of the ListenerSet.
	//
	// Implementations should prefer to express ListenerSet conditions
	// using the `GatewayConditionType` and `GatewayConditionReason`
	// constants so that operators and tools can converge on a common
	// vocabulary to describe Gateway state.
	//
	// Known condition types are:
	//
	// * "Accepted"
	// * "Programmed"
	//
	// +optional
	// +listType=map
	// +listMapKey=type
	// +kubebuilder:validation:MaxItems=8
	Conditions []metav1.Condition `json:"conditions,omitempty"`
}

// ListenerEntryStatus is the status associated with a ListenerEntry.
type ListenerEntryStatus struct {
	// Name is the name of the Listener that this status corresponds to.
	Name SectionName `json:"name"`

	// Port is the network port that this listener is listening on.
	Port PortNumber `json:"port"`

	// SupportedKinds is the list indicating the Kinds supported by this
	// listener. This MUST represent the kinds supported by an implementation for
	// that Listener configuration.
	//
	// If kinds are specified in Spec that are not supported, they MUST NOT
	// appear in this list and an implementation MUST set the "ResolvedRefs"
	// condition to "False" with the "InvalidRouteKinds" reason. If both valid
	// and invalid Route kinds are specified, the implementation MUST
	// reference the valid Route kinds that have been specified.
	//
	// +kubebuilder:validation:MaxItems=8
	SupportedKinds []RouteGroupKind `json:"supportedKinds"`

	// AttachedRoutes represents the total number of Routes that have been
	// successfully attached to this Listener.
	//
	// Successful attachment of a Route to a Listener is based solely on the
	// combination of the AllowedRoutes field on the corresponding Listener
	// and the Route's ParentRefs field. A Route is successfully attached to
	// a Listener when it is selected by the Listener's AllowedRoutes field
	// AND the Route has a valid ParentRef selecting the whole Gateway
	// resource or a specific Listener as a parent resource (more detail on
	// attachment semantics can be found in the documentation on the various
	// Route kinds ParentRefs fields). Listener or Route status does not impact
	// successful attachment, i.e. the AttachedRoutes field count MUST be set
	// for Listeners with condition Accepted: false and MUST count successfully
	// attached Routes that may themselves have Accepted: false conditions.
	//
	// Uses for this field include troubleshooting Route attachment and
	// measuring blast radius/impact of changes to a Listener.
	AttachedRoutes int32 `json:"attachedRoutes"`

	// Conditions describe the current condition of this listener.
	//
	// +listType=map
	// +listMapKey=type
	// +kubebuilder:validation:MaxItems=8
	Conditions []metav1.Condition `json:"conditions"`
}

// ParentGatewayReference identifies an API object including its namespace,
// defaulting to Gateway.
type ParentGatewayReference struct {
	// Group is the group of the referent.
	//
	// +optional
	// +kubebuilder:default="gateway.networking.k8s.io"
	Group *Group `json:"group"`

	// Kind is kind of the referent. For example "Gateway".
	//
	// +optional
	// +kubebuilder:default=Gateway
	Kind *Kind `json:"kind"`

	// Name is the name of the referent.
	Name ObjectName `json:"name"`

	// Namespace is the name of the referent.
	// +optional
	Name *ObjectName `json:"namespace"`
}
```

### YAML

The following example shows a `Gateway` with an HTTP listener and two child HTTPS `ListenerSets` with unique hostnames and certificates. 
Only `ListenerSets` from the same namespace of the `Gateway` will be accepted:

```yaml
apiVersion: gateway.networking.k8s.io/v1
kind: Gateway
metadata:
  name: parent-gateway
spec:
  gatewayClassName: example
  allowedListeners:
    namespaces:
      from: Same
  listeners:
  - name: foo
    hostname: foo.com
    protocol: HTTP
    port: 80
---
apiVersion: gateway.networking.x-k8s.io/v1alpha1
kind: XListenerSet
metadata:
  name: first-workload-listeners
spec:
  parentRef:
    name: parent-gateway
    kind: Gateway
    group: gateway.networking.k8s.io
  listeners:
  - name: first
    hostname: first.foo.com
    protocol: HTTPS
    port: 443
    tls:
      mode: Terminate
      certificateRefs:
      - kind: Secret
        group: ""
        name: first-workload-cert # Provisioned via HTTP01 challenge
---
apiVersion: gateway.networking.x-k8s.io/v1alpha1
kind: XListenerSet
metadata:
  name: second-workload-listeners
spec:
  parentRef:
    name: parent-gateway
    kind: Gateway
    group: gateway.networking.k8s.io
  listeners:
  - name: second
    hostname: second.foo.com
    protocol: HTTPS
    port: 443
    tls:
      mode: Terminate
      certificateRefs:
      - kind: Secret
        group: ""
        name: second-workload-cert # Provisioned via HTTP01 challenge
```
### ListenerEntry

`ListenerEntry` is currently a copy of the `Listener` struct with some changes noted in the below sections

#### Port

`Port` is now optional to allow for dynamic port assignment.  If the port is unspecified or set to zero, the implementation will assign a unique port. If the implementation does not support dynamic port assignment, it MUST set `Accepted` condition to `False` with the `UnsupportedPort` reason.

## Semantics

### Summary of attachments

Nick Young did a great summarization of how a handshake and route attachment should behave. This is copied as/is as it is a great summary:

* Routes attach to all **relevant** Listeners in their **direct Parent** object.
* If a Route has a parentRef that is a ListenerSet, then the ListenerSet is the **direct Parent**.
* If a Route has a parentRef that is a Gateway, then the Gateway is the **direct Parent**.
* If the Route is using Gateway defaulting, then the default Gateway is the **direct Parent**.
* If `sectionName` is not set, then all Listeners in the direct Parent object are **relevant**, and the Route must attach to all of them (subject to any `allowedRoutes` criteria.)
* If `sectionName` is set, then only the Listener in the direct Parent object with the `name` field set to the same value as `sectionname` is **relevant**, and the Route must attach only to that Listener (subject to any `allowedRoutes` criteria).
* If there are no relevant Listeners (because, for example, a sectionName is specified that does not exist on the direct Parent), then the Route has nowhere to attach to and MUST be have `Accepted` set to `false` for that `parentRef`.

Route status is per-`parentRef`, so if a new parentRef is added, that attachment is independent of any existing parentRefs. A Route can be Accepted by one parentRef and not another, that's fine.

If you have a Route that has a Gateway `parentRef` with a sectionName, that is Accepted already, and you add another parentRef that points to a `ListenerSet` with a `sectionName`, and the `ListenerSet` does not include a Listener with a name matching the sectionName, then the `ListenerSet parentRef` must be `Accepted status: false`. That does not affect the Gateway parentRef.

### Gateway Changes

An initial experimental release of `ListenerSets` _will have no modifications_ to listener list on the `Gateway` resource. Using `ListenerSets` will  require a dummy listener to be configured.

In a future (potential) release when an implementation supports `ListenerSets`, `Gateways` MUST allow the list of listeners to be empty. Thus the present `minItems=1` constraint on the listener list will be removed. This allows implementations to avoid security, cost etc. concerns with having dummy listeners.
When there are no listeners the `Gateway`'s `status.listeners` should be empty or unset. `status.listeners` is already an optional field.

Implementations, when creating a `Gateway`, may provision underlying infrastructure when there are no listeners present. The status conditions `Accepted` and `Programmed` conditions should reflect state of this provisioning.

### Gateway & ListenerSet Handshake

By default a `Gateway` MUST NOT allow `ListenerSets` to be attached. Users can enable this behaviour by configuring their `Gateway` to allow `ListenerSet` attachment:

```yaml
apiVersion: gateway.networking.k8s.io/v1
kind: Gateway
metadata:
  name: parent-gateway
spec:
  allowedListeners:
    namespaces:
      from: Same
```

### Route Attachment

Routes MUST be able to specify a `ListenerSet` as a `parentRef`. Routes can use `sectionName`/`port` fields in `ParentReference` to help target a specific listener. If no listener is targeted (`sectionName`/`port` are unset) then the Route attaches to all the listeners in the `ListenerSet`.

```yaml
apiVersion: gateway.networking.k8s.io/v1
kind: HTTPRoute
metadata:
  name: httproute-example
spec:
  parentRefs:
  - name: second-workload-listeners
    kind: XListenerSet
    sectionName: second
```

To attach a Route to a `XListenerSet` and its parent `Gateway`, it MUST have multiple `parentRefs` eg:

```yaml
apiVersion: gateway.networking.k8s.io/v1
kind: HTTPRoute
metadata:
  name: httproute-example
spec:
  parentRefs:
  - name: second-workload-listeners
    kind: XListenerSet
    sectionName: second
  - name: parent-gateway
    kind: Gateway
    sectionName: foo
```

For instance, the following `HTTPRoute` attempts to attach to a listener defined in the parent `Gateway` using the sectionName `foo`, which exists on a `ListenerSet` but not on a `Gateway`.
This is not valid and the route's status `Accepted` condition should be set to `False`

```yaml
apiVersion: gateway.networking.k8s.io/v1
kind: Gateway
metadata:
  name: parent-gateway
spec:
  gatewayClassName: example
  allowedListeners:
    namespaces:
      from: Same
  listeners:
  - name: notfoo
    hostname: foo.com
    protocol: HTTP
    port: 80
---
apiVersion: gateway.networking.x-k8s.io/v1alpha1
kind: XListenerSet
metadata:
  name: first-workload-listeners
spec:
  parentRef:
    name: parent-gateway
    kind: Gateway
    group: gateway.networking.k8s.io
  listeners:
  - name: foo
    hostname: first.foo.com
    protocol: HTTP
    port: 80
---
apiVersion: gateway.networking.k8s.io/v1
kind: HTTPRoute
metadata:
  name: httproute-example
spec:
  parentRefs:
  - name: parent-gateway
    kind: Gateway
    sectionName: foo
```

#### Gateway parents and sectionName

If a `sectionName` in a Route's `parentRef` is not set then the Route MUST attach to only the listeners in the referenced parent. As an example given a `Gateway` and its child `ListenerSets` a route attaching to the `Gateway` with an empty `sectionName` shall only attach to the listeners in the `Gateways` immediate `spec.listeners` list. 

In other words, the Route MUST attach just to the Gateway listeners specified on `.spec.listeners` and MUST NOT not attach to any listeners in the child `ListenerSets` 
of the `Gateway`.

This is necessary because, for UX reasons, the `name` field does not have to be unique across all Listeners merged into a Gateway (see the section below for details).

The following manifest exemplifies this situation:

```yaml
apiVersion: gateway.networking.k8s.io/v1
kind: Gateway
metadata:
  name: parent-gateway
spec:
  gatewayClassName: example
  allowedListeners:
    namespaces:
      from: Same
  listeners:
  - name: foo2
    hostname: foo.com
    protocol: HTTP
    port: 80
  - name: foo3
    hostname: foo1.com
    protocol: HTTP
    port: 80
---
apiVersion: gateway.networking.x-k8s.io/v1alpha1
kind: XListenerSet
metadata:
  name: first-workload-listeners
spec:
  parentRef:
    name: parent-gateway
    kind: Gateway
    group: gateway.networking.k8s.io
  listeners:
  - name: foo
    hostname: first.foo.com
    protocol: HTTP
    port: 80
---
apiVersion: gateway.networking.k8s.io/v1
kind: HTTPRoute
metadata:
  name: httproute-example
spec:
  parentRefs:
  # No SectionName is set on this parentRef
  - name: some-workload-listeners 
    kind: Gateway
```

The example above SHOULD be attached only to the following listeners:
```yaml
apiVersion: gateway.networking.k8s.io/v1
kind: Gateway
.....
spec:
  listeners:
  - name: foo2
    hostname: foo.com
    protocol: HTTP
    port: 80
  - name: foo3
    hostname: foo1.com
    protocol: HTTP
    port: 80
```

### Policy Attachment

Policy attachment is [under discussion] in https://github.com/kubernetes-sigs/gateway-api/discussions/2927

Similar to Routes, `ListenerSet` can inherit policy from a Gateway.
Policies that attach to a `ListenerSet` apply to all listeners defined in that resource, but do not impact listeners in the parent `Gateway`. This allows `ListenerSets` attached to the same `Gateway` to have different policies.
If the implementation cannot apply the policy to only specific listeners, it should reject the policy.

### ReferenceGrant Semantics

When a `ReferenceGrant` is applied to a `Gateway` it MUST NOT be inherited by child `ListenerSets`. Thus a `ListenerSet` listener MUST NOT access secrets granted to the `Gateway` listeners.

When a `ReferenceGrant` is applied to a `ListenerSet` it MUST NOT grant permission to the parent `Gateway`'s listeners. Thus a `Gateway` listener MUST NOT access secrets granted to the `ListenerSet` listeners.

A `ListenerSet` must be able to reference a secret/backend in the same namespace as itself without a `ReferenceGrant`.


### Listener Validation

Within a single resource such as a `Gateway` or `ListenerSet` the list of listeners MUST have unique names. Implementations MUST allow listeners from a child `ListenerSet` to be merged into a parent `Gateway` when listeners have the same name. Likewise implementations MUST allow sibling `ListenerSets` listeners with matching names to be merged into a parent `Gateway`. This allows for authors of Routes to simply attach to their desired parentRef and listener without having to worry about naming conflicts across resources.

It is up to the implementations how unique names are generated internally. One example would be to hash the `ListenerSet` name+namespace and prepend it to the listener entry `name`.

Implementations MUST treat the parent `Gateway`s as having the merged list of all listeners from itself and attached `ListenerSets` and validation of this list of listeners MUST behave the same as if the list were part of a single `Gateway` with the relaxed listener name constraints.

Ordering will follow the semantics defined in [Listener Precedence](#listener-precedence).

From the earlier example the above resources would be equivalent to a single `Gateway` where the listeners are collapsed into a single list.

```yaml
apiVersion: gateway.networking.k8s.io/v1
kind: Gateway
metadata:
  name: parent-gateway
spec:
  gatewayClassName: example
  listeners:
  - name: foo
    hostname: foo.com
    protocol: HTTP
    port: 80
  - name: first
    hostname: first.foo.com
    protocol: HTTPS
    port: 443
    tls:
      mode: Terminate
      certificateRefs:
      - kind: Secret
        group: ""
        name: first-workload-cert # Provisioned via HTTP01 challenge
  - name: second
    hostname: second.foo.com
    protocol: HTTPS
    port: 443
    tls:
      mode: Terminate
      certificateRefs:
      - kind: Secret
        group: ""
        name: second-workload-cert # Provisioned via HTTP01 challenge
```

### Listener Precedence

Listeners in a `Gateway` and their attached `ListenerSets` are concatenated as a list when programming the underlying infrastructure

Listeners should be merged using the following precedence:

1. "parent" Gateway
2. ListenerSet ordered by creation time (oldest first)
3. ListenerSet ordered alphabetically by “{namespace}/{name}”.

Conflicts are covered in the section [Listener and ListenerSet conflicts](#listener-and-listenerset-conflicts)

### Listener and ListenerSet conflicts

ListenerSet conflicts should be managed similarly to [Gateway resource conflict](https://github.com/kubernetes-sigs/gateway-api/blob/372a5b06624cff12117f41dcd26c08cb1def22e7/apis/v1/gateway_types.go#L76)
management.

With ListenerSet this validation should happen within the same ListenerSet resource, 
but MUST be validated also within a Gateway scope and all of the attached Listeners/ListenerSets. 
The SectionName field is an exception for this validation, and while it should 
not conflict within the same ListenerSet, it can be duplicated between different ListenerSets.

This means that the validation should happen now between distinct ListenerSets 
attached to the same Gateway, and in case of a conflict, the [Listener Precedence](#listener-precedence) 
should be respected, so the first Listener on the precedence list MUST be accepted, 
and should not have a `Conflicted` condition, while the conflicting listeners 
MUST have a `Conflicted` condition set to True and with an explicit reason on its message.

Following are some examples of a conflict situation:

#### Conflict between ListenerSet and parent Gateway

Given the following resource definitions:

```yaml
apiVersion: gateway.networking.k8s.io/v1
kind: Gateway
metadata:
  name: parent-gateway
  namespace: infra
spec:
  allowedListeners:
    namespaces:
      from: All
  listeners:
  - name: foo
    hostname: www.something.tld
    protocol: HTTPS
    port: 443
    tls:
      mode: Terminate
      certificateRefs:
      - kind: Secret
        group: ""
        name: default-cert
---
apiVersion: gateway.networking.x-k8s.io/v1alpha1
kind: XListenerSet
metadata:
  name: user-listenerset
  namespace: user01
spec:
  parentRef:
    name: parent-gateway
    kind: Gateway
    group: gateway.networking.k8s.io
  listeners:
  - name: myapp
    hostname: www.something.tld
    protocol: HTTPS
    port: 443
    tls:
      mode: Terminate
      certificateRefs:
      - kind: Secret
        group: ""
        name: app-cert
```

The ListenerSet `user-listenerset` should be marked as Conflicted, as the `parent-gateway`
has a listener definition called `foo` that conflicts with the ListenerSet definition
called `myapp`. The conflict happens because hostname is the same on both `ListenerSet` 
but they use different termination TLS certificates:

```yaml
apiVersion: gateway.networking.x-k8s.io/v1alpha1
kind: XListenerSet
metadata:
  name: user-listenerset
  namespace: user01
....
status:
  listeners:
  - name: myapp
    hostname: www.something.tld
    protocol: HTTPS
    port: 443
    conditions:
	- message: ListenerSet has conflicts with Gateway 'infra/parent-gateway'
      reason: ParentNotAccepted
      status: "False"
      type: Accepted
    - message: ListenerSet has conflicts with Gateway 'infra/parent-gateway'
      reason: ListenerConflict
      status: "True"
      type: Conflicted
```

#### Conflict between two ListenerSets

The following example represents a conflict between two ListenerSets on distinct
namespaces. The controller should avoid setting any Condition that exposes information
from other users, but still provide meaningful information of why a ListenerSet 
was not accepted


```yaml
apiVersion: gateway.networking.x-k8s.io/v1alpha1
kind: XListenerSet
metadata:
  creationTimestamp: "2025-08-11T15:44:05Z"
  name: listenerset1
  namespace: user01
spec:
  parentRef:
    name: parent-gateway
    kind: Gateway
    group: gateway.networking.k8s.io
  listeners:
  - name: myapp
    hostname: www.something.tld
    protocol: HTTPS
    port: 443
    tls:
      mode: Terminate
      certificateRefs:
      - kind: Secret
        group: ""
        name: app-cert
---
apiVersion: gateway.networking.x-k8s.io/v1alpha1
kind: XListenerSet
metadata:
  creationTimestamp: "2025-08-11T13:44:05Z"
  name: listenerset2
  namespace: user02
spec:
  parentRef:
    name: parent-gateway
    kind: Gateway
    group: gateway.networking.k8s.io
  listeners:
  - name: myapp
    hostname: www.something.tld
    protocol: HTTPS
    port: 443
    tls:
      mode: Terminate
      certificateRefs:
      - kind: Secret
        group: ""
        name: other-app-cert
```

In this case, there's a conflict as both users are setting the same hostname and 
port on distinct Listeners. In this case, because the ListenerSet `user02/listenerset2` 
is older, it will be accepted while `user01/listenerset1` should not be accepted, 
and receive a `Conflicted=True` condition.

The status of ListenerSets can be defined as the following:

```yaml
apiVersion: gateway.networking.x-k8s.io/v1alpha1
kind: XListenerSet
metadata:
  creationTimestamp: "2025-08-11T15:44:05Z"
  name: listenerset1
  namespace: user01
status:
  listeners:
  - name: myapp
    hostname: www.something.tld
    protocol: HTTPS
    port: 443
    conditions:
    - message: ListenerSet has conflicts with other listeners attached to the same Gateway
      reason: ListenerConflict
      status: "True"
      type: Conflicted
---
apiVersion: gateway.networking.x-k8s.io/v1alpha1
kind: XListenerSet
metadata:
  creationTimestamp: "2025-08-11T13:44:05Z"
  name: listenerset2
  namespace: user02
status:
  listeners:
  - name: myapp
    hostname: www.something.tld
    protocol: HTTPS
    port: 443
    conditions:
    - reason: Accepted
      status: "True"
      type: Accepted
```

### Gateway Conditions

`Gateway`'s `Accepted` and `Programmed` top-level conditions remain unchanged and reflect the status of the local configuration.

Implementations MUST support a new `Gateway` condition type `AttachedListenerSets`.

The condition's `Status` has the following values:

- `True` when `Spec.AllowedListeners` is set and at least one child Listener arrives from a `ListenerSet`
- `False` when `Spec.AllowedListeners` is set but has no valid listeners are attached
- `Unknown` when no `Spec.AllowedListeners` config is present

Parent `Gateways` MUST NOT have `ListenerSet` listeners in their `status.listeners` conditions list.

### ListenerSet Conditions

`ListenerSets` have a top-level `Accepted` and `Programmed` conditions.

The `Accepted` condition MUST be set on every `ListenerSet`, and indicates that the `ListenerSet` is semantically valid and accepted by its `parentRef`.

Valid reasons for `Accepted` being `False` are:

- `NotAllowed` - the `parentRef` doesn't allow attachment
- `ParentNotAccepted` - the `parentRef` isn't accepted (eg. invalid address)
- `ListenersNotValid` - one or more listeners in the set are invalid (or using an unsupported feature)

The `Programmed` condition MUST be set on every `ListenerSet` and have a similar meaning to the Gateway `Programmed` condition but only reflect the listeners in this `ListenerSet`.

`Accepted` and `Programmed` conditions when surfacing details about listeners, MUST only summarize the `status.listeners` conditions that are exclusive to the `ListenerSet`.
An exception to this is when the parent `Gateway`'s `Accepted` or `Programmed` conditions transition to `False`

`ListenerSets` MUST NOT have their parent `Gateway`'s' listeners in the associated `status.listeners` conditions list.

### ListenerConditions within a ListenerSet

An implementation MAY reject listeners by setting the `ListenerEntryStatus` `Accepted` condition to `False` with the `Reason` `TooManyListeners`

If a listener has a conflict, this should be reported in the `ListenerEntryStatus` of the conflicted `ListenerSet` by setting the `Conflicted` condition to `True`.

Implementations SHOULD be cautious about what information from the parent or siblings are reported to avoid accidentally leaking sensitive information that the child would not otherwise have access to. This can include contents of secrets etc.

Conflicts are covered in the section [Listener and ListenerSet conflicts](#listener-and-listenerset-conflicts)

## Alternatives

### Re-using Gateway Resource

The [first iteration of this GEP](https://github.com/kubernetes-sigs/gateway-api/pull/1863) proposed re-using the `Gateway` resource and introducing an `attachTo` property in the `infrastructure` stanza.

The main downside of this approach is that users still require `Gateway` write access to create listeners. Secondly, it introduces complexity to future `Gateway` features as GEP authors would have now have to account for merging semantics.

### New 'GatewayGroup' Resource

This was proposed in the Gateway Hierarchy Brainstorming document (see references below). The idea is to introduce a central resource that will coalesce Gateways together and offer forms of delegation.

Issues with this is complexity with status propagation, cluster vs. namespace scoping etc. It also lacks a migration path for existing Gateways to help shard listeners.

### Use of Multiple Disjointed Gateways

An alternative would be to encourage users to not use overly large Gateways to minimize the blast radius of any issues. Use of disjoint Gateways could accomplish this but it has the disadvantage of consuming more resources and introducing complexity when it comes to operations work (eg. setting up DNS records etc.)

### Increase the Listener Limit

Increasing the limit may help in situations where you are creating many listeners such as adding certificates created using an ACME HTTP01 challenge. Unfortunately this still makes the Gateway a single point of contention. Unfortunately, there will always be an upper bound because of etcd limitations.
For workloads like Knative we can have O(1000) Services on the cluster with unique subdomains.

### Expand Route Functionality

For workloads with many certificates one option would be to introduce a `tls` stanza somewhere in the Route types. These Routes would then attach to a single Gateway. Then application operators can provide their own certificates. This probably would require some ability to have a handshake agreement with the Gateway.

Somewhat related, there was a Route Delegation GEP (https://github.com/kubernetes-sigs/gateway-api/issues/1058) that was abandoned

## References

First Revision of the GEP

- https://github.com/kubernetes-sigs/gateway-api/pull/1863

Mentioned in Prior GEPs:

- https://github.com/kubernetes-sigs/gateway-api/pull/1757

Prior Discussions:

- https://github.com/kubernetes-sigs/gateway-api/discussions/1248
- https://github.com/kubernetes-sigs/gateway-api/discussions/1246

Gateway Hierarchy Brainstorming:

- https://docs.google.com/document/d/1qj7Xog2t2fWRuzOeTsWkabUaVeOF7_2t_7appe8EXwA/edit
