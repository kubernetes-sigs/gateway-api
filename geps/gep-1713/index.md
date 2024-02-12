# GEP-1713: Standard Mechanism to Merge Multiple Gateways

* Issue: [#1713](/kubernetes-sigs/gateway-api/issues/1713)
* Status: Provisional

(See status definitions [here](overview.md#status).)

## tl;dr

The Gateway Resource is a contention point since it is the only place to attach listeners with certificates. We propose a mechanism to allow distinct Gateway resources to be logically merged.

## Goals

- Define a mechanic to merge multiple Gateways (logically)
- Define a set of acceptable properties that can be merged and their semantics

## Non-Goals

- Apply a Gateway resource onto N distinct gateways (one to many)

## Introduction

Knative generates on demand per-service certificates using HTTP-01 challenges. 
There can be O(1000) Knative Services in the cluster which means we have O(1000) distinct certificates. 
Thus updating a single Gateway resource with this many certificates is a contention point and inhibits horizontal scaling of our controllers.
[Istio Ambient](https://istio.io/v1.15/blog/2022/introducing-ambient-mesh/), similarly, creates a listener per Kubernetes service.

More broadly, large scale gateway users often expose O(1000) domains, but are currently limited by the maximum of 16 `listeners`.

The spec currently has language to indicate implementations `MAY` merge Gateways resources but the mechanic isn't defined.
https://github.com/kubernetes-sigs/gateway-api/blob/541e9fc2b3c2f62915cb58dc0ee5e43e4096b3e2/apis/v1beta1/gateway_types.go#L76-L78

## API

This proposal has two aspects: configuration for the parent Gateway, and configuration for the child Gateway.

A "parent" Gateway _does not_ reference another Gateway. A parent Gateway MUST explicitly opt into merging.
This is done with a new field `allowedChildren` in the `spec.infrastructure` of a Gateway.
If, and only if, they have done so, the same Gateway is also permitted to leave `listeners` empty (currently, there is a `MinItems=1` restriction).

A "child" Gateway references a parent Gateway. 
This is done with a new field `attachTo` in the `spec.infrastructure` stanza of a Gateway.
The `attachTo` field is a new type `GatewayObjectReference`.
Although the use of `GatewayObjectReference` allows users to attach to any `kind`, this GEP only defines the behavior of attaching a Gateway to another Gateway.

A "sibling" is a Gateway that shares a parent with another child Gateway.

Status requirements are specified [below](#status-fields).

See [GEP-1867](https://github.com/kubernetes-sigs/gateway-api/pull/1868) for more use cases of `infrastructure`.


#### Go

```go
type GatewayInfrastructure struct {
	// ...

	// AllowedChildren allows child objects to attach to this Gateway.
	// A common scenario is to allow other objects to add listeners to this Gateway.
	AllowedChildren *AllowedChildren `json:"allowedChildren,omitempty"`

	// AttachTo allows the Gateway to associate itself with another resource.
	// A common scenario is to reference another Gateway which marks
	// this Gateway a child of another.
	AttachTo GatewayObjectReference `json:"attachTo"`
}

// AllowedChildren defines which objects may be attached as children
type AllowedChildren struct {
	// Namespaces indicates namespaces from which children may be attached to this
	// Gateway. This is restricted to the namespace of this Gateway by default.
	//
	// Support: Core
	//
	// +optional
	// +kubebuilder:default={from: Same}
	Namespaces *ChildrenNamespaces `json:"namespaces,omitempty"`
}

// ChildrenNamespaces indicate which namespaces Children should be selected from.
type ChildrenNamespaces struct {
	// From indicates where Children will be selected for this Gateway. Possible
	// values are:
	//
	// * All: Children in all namespaces may be used by this Gateway.
	// * Selector: Children in namespaces selected by the selector may be used by
	//   this Gateway.
	// * Same: Only Children in the same namespace may be used by this Gateway.
	//
	// Support: Core
	//
	// +optional
	// +kubebuilder:default=Same
	From *FromNamespaces `json:"from,omitempty"`

	// Selector must be specified when From is set to "Selector". In that case,
	// only Children in Namespaces matching this Selector will be selected by this
	// Gateway. This field is ignored for other values of "From".
	//
	// Support: Core
	//
	// +optional
	Selector *metav1.LabelSelector `json:"selector,omitempty"`
}

// GatewayObjectReference identifies an API object including its namespace,
// defaulting to Gateway.
type GatewayObjectReference struct {
	// Group is the group of the referent. For example, "gateway.networking.k8s.io".
	// When unspecified or empty string, core API group is inferred.
	//
	// +optional
	// +kubebuilder:default=""
	Group *Group `json:"group"`

	// Kind is kind of the referent. For example "Gateway".
	//
	// +optional
	// +kubebuilder:default=Gateway
	Kind *Kind `json:"kind"`

	// Name is the name of the referent.
	Name ObjectName `json:"name"`

	// Namespace is the namespace of the referenced object. When unspecified, the local
	// namespace is inferred.
	//
	// Support: Core
	//
	// +optional
	Namespace *Namespace `json:"namespace,omitempty"`
}
```

#### YAML

Below shows an example of an end to end configuration with Gateway merging.
Here we define a parent resource, which allows children from the same namespace.
A single listener is defined in the parent.

The child Gateway attaches to this Gateway and specifies an additional listener.

```yaml
apiVersion: gateway.networking.k8s.io/v1beta1
kind: Gateway
metadata:
  name: parent-gateway
spec:
  gatewayClassName: example
  infrastructure:
    allowedChildren:
      namespaces:
        from: Same
  listeners:
  - name: common-monitoring
    port: 8080
    protocol: HTTP
---
apiVersion: gateway.networking.k8s.io/v1beta1
kind: Gateway
metadata:
  name: child-gateway
spec:
  gatewayClassName: example
  infrastructure:
    attachTo:
      name: parent-gateway
      kind: Gateway
      group: gateway.networking.k8s.io
  listeners:
  - name: domain-a
    hostname: a.example.com
    protocol: HTTP
    port: 80
```

Logically, this is equivalent to a single Gateway as below:

```yaml
apiVersion: gateway.networking.k8s.io/v1beta1
kind: Gateway
metadata:
  name: gateway
spec:
  gatewayClassName: example
  listeners:
  - name: common-monitoring
    port: 8080
    protocol: HTTP
  - name: domain-a
    hostname: a.example.com
    protocol: HTTP
    port: 80
```

### Semantics

#### Route Attaching

Routes MUST be able to specify a child Gateway as a `parentRef` and make use of the fields in `ParentReference` to help target a specific listener.
If no listener is targeted (`sectionName`/`port` are unset) then the Route references all the listeners on the child Gateway. It MUST NOT attach
to a listener on a parent or sibling Gateway.

```yaml
apiVersion: gateway.networking.k8s.io/v1beta1
kind: HTTPRoute
metadata:
  name: httproute-example
spec:
  parentRefs:
  - name: child-gateway
    sectionName: metrics
```

Routes can only bind to listeners *directly* defined in the `Gateway` referenced.
For instance, the following route referencing a listener defined in the parent `Gateway`, but attaching to the child `Gateway` is not valid.
This will be reported in [status](#status-fields).

```yaml
apiVersion: gateway.networking.k8s.io/v1beta1
kind: HTTPRoute
metadata:
  name: httproute-example
spec:
  parentRefs:
  - name: child-gateway
    sectionName: section-in-parent
```

#### Merging Spec Fields

##### Validation

A parent and child MUST have the same `gatewayClassName`.
This will be detected by the implementation and reported in [status](#status-fields).

A child resource MUST not set any `spec.infrastructure` fields beyond `attachTo`, and cannot set `spec.address`.
This can be validated in the CRD schema.

A parent resource MUST not set `spec.infrastructure.attachTo`.
That is, we do not allow multiple tiers of Gateways chaining to each other; there is only a single parent with children.
This can be validated in the CRD schema.

A child resource cannot `attachTo` any Gateway resource that doesn't allow attachment (eg. it does not specify `spec.infrastructure.allowedChildren` for `Gateway`s).
This will be detected by the implementation and reported in [status](#status-fields).

##### Listeners

Implementations MUST treat the "parent" Gateway as having the concatenated list of all listeners from itself and "child" Gateways.

Validation of this list of listeners MUST behave the same as if the list were part of a single "parent" Gateway.

eg.
```yaml
apiVersion: gateway.networking.k8s.io/v1beta1
kind: Gateway
metadata:
  name: parent-gateway
spec:
  infrastructure:
    allowedChildren:
      namespaces:
        from: Same
  gatewayClassName: example
  listeners:
  - name: HTTP
    port: 80
    protocol: HTTP
---
apiVersion: gateway.networking.k8s.io/v1beta1
kind: Gateway
metadata:
  name: child-gateway
spec:
  gatewayClassName: example
  infrastructure:
    attachTo:
      name: parent-gateway
      kind: Gateway
      group: gateway.networking.k8s.io
  listeners:
  - name: metrics
    port: 8080
    protocol: HTTP
```

With this configuration, the realized "parent" Gateway should listen on port `80` and `8080`.

###### Listener Precedence

Gateway Listeners should be merged using the following precedence:
- "parent" Gateway
- "child" Gateway ordered by creation time (oldest first)
- "child" Gateway ordered alphabetical by “{namespace}/{name}”.

If there are conflicts between these, this should be reported as `Conflicted=True` in the listener as usual.

### Status Fields

#### Addresses

The list of `Addresses` that appear in the status of the "child" Gateway MUST be the same as the "parent" Gateway.

#### Gateway Conditions

Gateway conditions currently supports the following condition types: `Accepted` and `Programmed`

For parent gateways, `Accepted` should be set based on the entire set of merged listeners.
For instance, if a child listener is invalid, `ListenersNotValid` would be reported.
`Programmed` is not expected, generally, to depend on the children resources, but if an implementation does depend on these
they should consider child resources when reporting this status.

For child gateways, `Accepted` and `Programmed` should consider the overall merged Gateway status, but only the child's own listeners.

For example, if I have a `parent`, `child-1`, and `child-2`:
* If parent is entirely invalid (for example, an invalid `address`), all three Gateways will reported `Accepted=False`.
* If `child-1` has an invalid listener, `parent` and `child-1` will report `ListenersNotValid`, while `child-2` will not.
* If `child-1` references a parent that doesn't exist then `child-1` will report `Accepted=False`
* If `child-1` references a parent that doesn't allow merging then `child-1` will report `Accepted=False`
* If `child-1` references another child (eg. `child-2`) then `child-1` will report `Accepted=False` 
* If `child-1` references itself then `child-1` will report `Accepted=False`
* If `child-1` and `parent` have different gatewayClassNames then `child-1` will report `Accepted=False`

When reporting status of a child, an implementation SHOULD be cautious about what information from the parent or siblings are reported
to avoid accidentally leaking sensitive information that the child would not otherwise have access to.

#### Listener Conditions

Listener conditions should only be set for listeners directly defined in a given Gateway.
Parent gateways MUST NOT have children's resources in their listener conditions list.
Children gateways MUST NOT have parent's or sibling's resources in their listener conditions list.

#### Policy attachment

Policies attached to a parent Gateway apply to both the parent and all children listeners.

Policies that attach to a child Gateway apply to all listeners defined in that Gateway, but do not impact
parent or sibling listeners.
If the implementation cannot apply the policy to only specific listeners, it should reject the policy.

## Future Goals

### Requirement Level

We want to keep this API very simple so that the merging requirement level could increase from `MAY` to `MUST`

## Alternatives

#### New Resource
A `GatewayListener` resource could be a simpler solution as we would not have to set required fields (ie. gatewayClassName)

```
apiVersion: gateway.networking.k8s.io/v1beta1
kind: GatewayListener
metadata:
  name: listener
spec:
  gateway: parent-gateway
  listeners:
  - name: metrics
    port: 8080
    protocol: HTTP
status: ...
```

#### Use of the `gateway.networking.k8s.io/parent-gateway` label

Use of a label (ie. `gateway.networking.k8s.io/parent-gateway: name`) could be used to select child gateways vs using `spec.infrastructure.attachTo`

## References

Mentioned in Prior GEPs:
- https://github.com/kubernetes-sigs/gateway-api/pull/1757

Prior Discussions:
- https://github.com/kubernetes-sigs/gateway-api/discussions/1248
- https://github.com/kubernetes-sigs/gateway-api/discussions/1246
