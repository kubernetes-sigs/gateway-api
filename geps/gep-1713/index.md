# GEP-1713: Standard Mechanism to Merge Multiple Gateways

* Issue: [#1713](/kubernetes-sigs/gateway-api/issues/1713)
* Status: Provisional

(See status definitions [here](overview.md#status).)

## tl;dr

The `Gateway` Resource is a point of contention since it is the only place to attach listeners with certificates. We propose a new resource called `ListenerSet` to allow a shared list of listeners to be attached to a single `Gateway`.

## Goals
- Define a mechanism to merge listeners into a single `Gateway`

## Future Goals (Beyond the GEP)
- Attaching listeners to `Gateways` in different namespaces

## Introduction

Knative generates on demand per-service certificates using HTTP-01 challenges. 
There can be O(1000) Knative `Services` in the cluster which means we have O(1000) distinct certificates. 
Thus updating a single `Gateway` resource with this many certificates is a contention point and inhibits horizontal scaling of our controllers.

More broadly, large scale gateway users often expose `O(1000)` domains, but are currently limited by the maximum of 64 `listeners`.

The spec currently has language to indicate implementations `MAY` merge `Gateways` resources but the mechanic isn't defined.
https://github.com/kubernetes-sigs/gateway-api/blob/541e9fc2b3c2f62915cb58dc0ee5e43e4096b3e2/apis/v1beta1/gateway_types.go#L76-L78

## API

This proposal introduces a new `ListenerSet` resource that has the ability to attach to a set of listeners to a parent `Gateway`.

#### Go

```go
type GatewaySpec struct {
  ...
  // Note: this is a list to allow future potential features
  AllowedListeners []*AllowedListeners `json:"allowedListeners"`
  ...
}

type AllowedListeners struct {
    // +kubebuilder:default={from: Same}
    Namespaces *ListenerNamespaces `json:"namespaces,omitempty"`
}

// ListenerNamespaces indicate which namespaces ListenerSets should be selected from.
type ListenerNamespaces struct {
	// From indicates where ListenerSets can attach to this Gateway. Possible
	// values are:
	//
	// * Same: Only ListenerSets in the same namespace may be attached to this Gateway.
	// * None: Only listeners defined in the Gateway's spec are allowed
	//
	// +optional
	// +kubebuilder:default=Same
	// +kubebuilder:validation:Enum=Same;None
	From *FromNamespaces `json:"from,omitempty"`
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
    ParentRef ParentGatewayReference `json:"parentRefs,omitempty"`
    
    // Listeners associated with this ListenerSet. Listeners define
    // logical endpoints that are bound on this referenced parent Gateway's addresses.
	//
	// At least one Listener MUST be specified.
	//
	// Note: this is the same Listener type in the GatewaySpec struct
    Listeners []Listener 
}

// ListenerSetStatus defines the observed state of ListenerSet.
type ListenerSetStatus struct {
    // Listeners provide status for each unique listener port defined in the Spec.
	//
	// +optional
	// +listType=map
	// +listMapKey=name
	// +kubebuilder:validation:MaxItems=64
	//
	// Note: this is the same ListenerStatus type in the GatewayStatus struct
	Listeners []ListenerStatus `json:"listeners,omitempty"`
	
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
}
```

#### YAML

The following example shows a `Gateway` with an HTTP listener and two child HTTPS `ListenerSets` with unique hostnames and certificates. 

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
---
apiVersion: gateway.networking.k8s.io/v1alpha1
kind: ListenerSet
metadata:
  name: first-workload-listeners
  labels:
    gateway.networking.k8s.io/gateway.name: parent-gateway
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
apiVersion: gateway.networking.k8s.io/v1alpha1
kind: ListenerSet
metadata:
  name: second-workload-listeners
  labels:
    gateway.networking.k8s.io/gateway.name: parent-gateway
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
### Semantics

#### Gateway <> ListenerSet Handshake

By default a `Gateway` will allow `ListenerSets` in the same namespace to be attached. Users can prevent this behaviour by configuring their `Gateway` to disallow any listener attachment:

```
apiVersion: gateway.networking.k8s.io/v1
kind: Gateway
metadata:
  name: parent-gateway
spec:
  allowedListeners:
  - from: None
```

#### Listeners

####  `ListenerSet.spec.parentRef` is immutable

A `ListenerSet` `MUST` be created with a `parentRef` Updating the `parentRef` of a ListenerSet `MUST` be rejected by the API.

#### `ListenerSet` Default Labels

In order for clients of the API to easily discover `Gateways` and their `ListenerSets` we propose a set of labels to be defaulted when creating `ListenerSets`. This only applies if the `parentRef.kind` is a `Gateway`

- `gateway.networking.k8s.io/gateway.name` is equal to `spec.parentRef.name`
- `gateway.networking.k8s.io/gateway.namespace` is equal to `spec.parentRef.namespace`

This allows clients to perform API [lookups](https://kubernetes.io/docs/concepts/overview/working-with-objects/labels/#list-and-watch-filtering) using label selectors.

TBD: Can conditional defaulting be done with CEL?

#### GatewaySpec.Listeners MinItems

TBD: Do we want to make this change?

Currently when creating a Gateway you must specify at least a single listener. With `ListenerSets` this opens the possibility of wanting to create a Gateway with no listeners.

#### Route Attaching

Routes MUST be able to specify a `ListenerSet` as a `parentRef` and make use of the `sectionName` field in `ParentReference` to help target a specific listener. If no listener is targeted (`sectionName`/`port` are unset) then the Route references all the listeners on the `ListenerSet`. It `MUST NOT` attach to additional listeners on the parent `Gateway`.

```yaml
apiVersion: gateway.networking.k8s.io/v1beta1
kind: HTTPRoute
metadata:
  name: httproute-example
spec:
  parentRefs:
  - name: second-workload-listeners
    kind: ListenerSet
    sectionName: second
```

For instance, the following `HTTPRoute` attemps to attach to a listener defined in the parent `Gateway`. This is not valid and the route's status should reflect that.

```yaml
apiVersion: gateway.networking.k8s.io/v1beta1
kind: HTTPRoute
metadata:
  name: httproute-example
spec:
  parentRefs:
  - name: some-workload-listeners
    kind: ListenerSet
    sectionName: foo
```

#### Listener Validation

Implementations MUST treat the parent `Gateway` as having the concatenated list of all listeners from itself and attached `ListenerSets`
Validation of this list of listeners MUST behave the same as if the list were part of a single `Gateway`.

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

#### Listener Precedence

Listeners should be merged using the following precedence:
1. "parent" Gateway
2. ListenerSet ordered by creation time (oldest first)
3. ListenerSet ordered alphabetically by “{namespace}/{name}”.

If there are listener conflicts, this should be reported as `Conflicted=True` in the ListenerStatus as usual. See 'Conditions' section below for more details on object should report the conflict.

###  Gateway Conditions

`Gateway` currently supports the following top-level condition types: `Accepted` and `Programmed`

For a `Gateway`, `Accepted` should be set based on the entire set of merged listeners.
For instance, if a `ListenerSet` listener is invalid, `ListenersNotValid` would be reported. 
`Programmed` is not expected, generally, to depend on the children resources, but if an implementation does depend on these
they should consider child resources when reporting this status.

Parent gateways MUST NOT have `ListenerSet` listeners in their `status.listeners` conditions list.

It is up to the implementation whether an invalid listener affects other listeners in the Gateway.

### ListenerSet Conditions

`ListenerSets` MUST NOT have their parent `Gateway`'s' listeners in the `status.listeners` conditions list.  An implementation MAY reject listeners with `ListenerConditionAccepted=False` and Reason `TooManyListeners` `ListenerSets`, like a `Gateway`, also have two top-level conditions: `Accepted` and `Programmed`. These conditions, when surfacing details about listeners, MUST only summarize the `status.listener` conditions that are exclusive to the `ListenerSet`. 

These conditions MUST also surface top-level `Gateway` conditions that impact the `ListenerSet`. For example, if a `Gateway` requests an invalid address and it cannot be accepted/programmed then the `ListenerSet`'s' `Accepted` condition MUST be set to `False`.

For example, if I have a `Gateway` named `parent`, and two `ListenerSets` named `child-1`, and `child-2` then:
* If `parent` is entirely invalid (for example, an invalid `address`) and `Accepted=False`, all two `ListenerSets` will reported `Accepted=False`.
* If `child-1` has an invalid listener, `parent` and `child-1` will report `ListenersNotValid`, while `child-2` will not.
* If `child-1` references a parent that doesn't allow merging then `child-1` will report `Accepted=False`
* If `child-1` references another child (eg. `child-2`) then `child-1` will report `Accepted=False` 
* If `child-1` is valid, then when `child-2` is created if it conflicts with `child-1` then `child-2` will report `Accepted=False`. `child-1` status conditions will remain unchanged. `parent` will report `ListenersNotValid`

When reporting status of a child, an implementation SHOULD be cautious about what information from the parent or siblings are reported
to avoid accidentally leaking sensitive information that the child would not otherwise have access to.

#### Policy Attachment

Policies attached to a parent `Gateway` apply to both the parent and all `ListenerSet` listeners.

Policies that attach to a `ListenerSet` apply to all listeners defined in that resource, but do not impact listeners in the parent `Gateway`
If the implementation cannot apply the policy to only specific listeners, it should reject the policy.

## Alternatives

#### Re-using Gateway Resource

The [first iteration of this GEP](https://github.com/kubernetes-sigs/gateway-api/pull/1863) proposed re-using the `Gateway` resource and introducing an `attachTo` property in the `infrastructure` stanza.

The main downside of this approach is that users still require `Gateway` write access to create listeners. Secondly, it introduces complexity to future `Gateway` features as GEP authors would have now have to account for merging semantics.

#### New 'GatewayGroup' Resource

This was proposed in the Gateway Hiearchy Brainstorming document (see references below). The idea is to introduce a central resource that will coalease Gateways together and offer forms of delegation.

Issues with this is complexity with status propagation, cluster vs. namespace scoping etc. It also lacks a migration path for existing Gateways to help shard listeners.

#### Use of Multiple Disjointed Gateways

An alternative would be to encourage users to not use overly large Gateways to minimize the blast radius of any issues. Use of disjoint Gateways could accomplish this but it has the disadvantage of consuming more resources and introducing complexity when it comes to operations work (eg. setting up DNS records etc.)

#### Increase the Listener Limit

Increasing the limit may help in situations where you are creating many listeners such as adding certificates created using an ACME HTTP01 challenge. Unfortunately this still makes the Gateway a single point of contention. Unfortunately, there will always be an upper bound because of etcd limitations.
For workloads like Knative we can have O(1000) Services on the cluster with unique subdomains.

#### Expand Route Functionality

For workloads with many certificates one option would be to introduce a `tls` stanza somewhere in the Route types. These Routes would then attach to a single Gateway. Then application operators can provide their own certificates. This probably would require some ability to have a handshake agreement with the Gateway.

Sorta related there was a Route Delegation GEP (https://github.com/kubernetes-sigs/gateway-api/issues/1058) that was abandoned

## References

First Revision of the GEP
- https://github.com/kubernetes-sigs/gateway-api/pull/1863

Mentioned in Prior GEPs:
- https://github.com/kubernetes-sigs/gateway-api/pull/1757

Prior Discussions:
- https://github.com/kubernetes-sigs/gateway-api/discussions/1248
- https://github.com/kubernetes-sigs/gateway-api/discussions/1246

Gateway Hierarchy Brainstorming
- https://docs.google.com/document/d/1qj7Xog2t2fWRuzOeTsWkabUaVeOF7_2t_7appe8EXwA/edit
