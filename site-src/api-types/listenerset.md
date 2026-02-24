# ListenerSet

??? success "Standard Channel since v1.5.0"

    The `ListenerSet` resource is GA and has been part of the Standard Channel since
    `v1.5.0`. For more information on release channels, refer to our [versioning
    guide](../concepts/versioning.md).

A `ListenerSet` is a Gateway API type for specifying additional listeners for a Gateway.
It decouples network listener configurations—such as ports, hostnames, and TLS
termination—from the central Gateway resource.

## Background

ListenerSets allow teams to independently define and attach groups of listeners to a central,
shared Gateway. This enables self-service TLS configuration, improves multi-tenancy by allowing
decentralized listener management, and allows scaling beyond the 64-listener limit of a single
Gateway resource.

ListenerSets offer the following advantages:

- *Multitenancy*: You can let different teams create their own ListenerSets while sharing the same
Gateway and backing load-balancing infrastructure.

- *Large scale deployments*: By using ListenerSets, Gateways can have more than 64 listeners attached.
Teams can also share the same ListenerSet configuration to avoid duplication.

- *Certificates for more listeners per gateway*: Because you can now have more than 64 listeners per Gateway,
a single Gateway can forward secured traffic to more backends that might have their own certificates.
This approach aligns with projects that require service-level certificates, such as Istio Ambient Mesh or Knative.

## Spec

The `ListenerSet` spec defines the following:

*   `ParentRef`- Define which Gateway this ListenerSet wants to be attached
  to.
*   `Listeners`-  Define the hostnames, ports, protocol, termination, TLS
    settings and which routes can be attached to a listener.

## Attaching ListenerSets
### Gateway Configuration

By default a `Gateway` does not allow `ListenerSets` to be attached. Users can enable this behaviour
by configuring their `Gateway` to allow `ListenerSet` attachment via `spec.allowedListeners` :

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

The `namespaces.from` field within `AllowedListeners` can take the following four values:

- `None` (Default): No external ListenerSets are allowed to attach. Only the listeners defined
directly inside the Gateway resource will be used.

- `Same`: Only ListenerSets located in the same namespace as the Gateway can be attached.

- `All`: ListenerSets from any namespace in the cluster are allowed to attach, provided they
have a valid parentRef pointing to the Gateway.

- `Selector`: Only ListenerSets in namespaces that match a specific label selector are allowed.
When using this value, you must also provide the selector field

### ListenerSet Configuration
A ListenerSet uses a parentRef to point to a specific Gateway

```yaml
apiVersion: gateway.networking.k8s.io/v1
kind: ListenerSet
metadata:
  name: workload-listeners
spec:
  parentRef:
    name: parent-gateway
    kind: Gateway
    group: gateway.networking.k8s.io
```

### Route Attachment

Routes can specify a `ListenerSet` as a `parentRef`. Routes can use `sectionName` fields
in `ParentReference` to help target a specific listener. If no listener is targeted (`sectionName`
is unset) then the Route attaches to all the listeners in the `ListenerSet`.

```yaml
apiVersion: gateway.networking.k8s.io/v1
kind: HTTPRoute
metadata:
  name: httproute-example
spec:
  parentRefs:
  - name: workload-listeners
    kind: ListenerSet
    group: gateway.networking.k8s.io
    sectionName: second
```

In some rare cases, you may want to attach a Route to separate Listeners in both a ListenerSet and its parent Gateway. In that case, your configuration would look like this:

```yaml
apiVersion: gateway.networking.k8s.io/v1
kind: HTTPRoute
metadata:
  name: httproute-example
spec:
  parentRefs:
  - name: workload-listeners
    kind: ListenerSet
    group: gateway.networking.k8s.io
    sectionName: second
  - name: parent-gateway
    kind: Gateway
    sectionName: foo
```

## Listener Conflicts

Conflicts occur when multiple listeners claim the same Port, Protocol, and/or Hostname. The controller resolves these using the following priority:

1. Parent Gateway Listeners: Listeners defined directly in the Gateway spec always have the highest priority.

2. ListenerSet Creation Time: If two ListenerSets conflict, the one with the older creationTimestamp wins.

3. ListenerSet Alphabetical Order: If creation timestamps are identical, priority is given based on the alphabetical order of the resource's `{namespace}`/`{name}`.

The listener with the highest priority is Accepted and Programmed
The lower-priority listener is marked with a `Conflicted: True` condition in its status.

!!! note  Partial ListenerSet Acceptance
    A ListenerSet may be partially accepted if only some of its listeners are in conflict. Valid listeners will continue to route traffic, while conflicted ones will not route traffic.

## Status Updates

A ListenerSet successfully attaches to a Gateway when all three of the following conditions are met:

1. Gateway AllowedListeners Configuration : By default, Gateways do not allow any external ListenerSets to attach. The Gateway must have an `allowedListeners` field in its spec that selects the namespace of the ListenerSet.

2. Valid Parent Reference : The ListenerSet must explicitly point back to the target Gateway.

3. Resource-Level Acceptance : The Gateway Controller must validate and "accept" the resource (all listeners are valid, etc.).

!!! note Partial ListenerSet Acceptance
    A ListenerSet can be Accepted overall even if one of its individual listeners is in conflict with another set. In this case, only the non-conflicting listeners are "Programmed" into the data plane.

### Gateway Status
The parent `Gateway` status reports the number of successful attached listeners to `.status.attachedListenerSets`.

```yaml
apiVersion: gateway.networking.k8s.io/v1
kind: Gateway
metadata:
  name: parent-gateway
...
status:
  attachedListenerSets: 2
```

### ListenerSet Status
`ListenerSets` have a top-level `Accepted` and `Programmed` conditions. The details are as follows :

The `Accepted: True` condition is set if the ListenerSet is accepted.

The `Accepted: False` condition can be set for multiple reasons such as the parent Gateway does not allow ListenerSets, all the listeners are invalid, etc.

Because a ListenerSet can contain multiple listeners, each one gets its own status entry and follows the same logic as Gateway listeners.
