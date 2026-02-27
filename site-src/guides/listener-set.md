# ListenerSet

ListenerSets allow teams to define ports, hostnames, and TLS certificates in separate resources
rather than cramming everything into one giant Gateway object which has a limit of 64 listeners.
This enables a delegated management model suitable for high-scale, multi-tenant environments.

As such, you might use ListenerSets for the following advantages:

- *Multitenancy*: You can let different teams create their own ListenerSets while sharing the same
Gateway and backing load-balancing infrastructure.

- *Large scale deployments*: By using ListenerSets, Gateways can have more than 64 listeners attached.
Teams can also share the same ListenerSet configuration to avoid duplication.

- *Certificates for more listeners per gateway*: Because you can now have more than 64 listeners per Gateway,
a single Gateway can forward secured traffic to more backends that might have their own certificates.
This approach aligns with projects that require service-level certificates, such as Istio Ambient Mesh or Knative.

The following diagram illustrates how ListenerSets help decentralize route configuration in a multi-tenant environment:

- Team 1 and Team 2 each manage their own Service and HTTPRoute resources within their respective namespaces.

- Each HTTPRoute refers to a namespace-local ListenerSet. This way, each team controls how their routes are exposed,
such as the protocol, port, and TLS certificate settings.

- The ListenerSets from both teams share a common Gateway in a separate namespace. A separate Gateway team can setup
and manage centralized infrastructure or enforce policies as appropriate.

```mermaid
flowchart TD

  subgraph team1 namespace
    SVC1[Services]
    HR1[HTTPRoutes]
    LS1[ListenerSet]
  end

  subgraph team2 namespace
    SVC2[Services]
    HR2[HTTPRoutes]
    LS2[ListenerSet]
  end

  subgraph shared namespace
    GW[Gateway]
  end

  HR1 -- "parentRef" --> LS1
  LS1 -- "parentRef" --> GW
  HR1 -- "backendRef" --> SVC1

  HR2 -- "parentRef" --> LS2
  LS2 -- "parentRef" --> GW
  HR2 -- "backendRef" --> SVC2
```

## Using ListenerSets
### Gateway Configuration

By default, a Gateway does not allow ListenerSets to be attached. Users can enable this behaviour
by configuring their Gateway to allow ListenerSets by adding the `allowedListeners` field to the Gateway spec.
This defines which namespaces are permitted to contribute listeners.

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

### ListenerSet Configuration
The ListenerSet references the parent Gateway via `parentRef`:

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

To attach an `HTTPRoute` (or any other route type) to a ListenerSet,
it must reference the `ListenerSet` as the Parent Resource via `parentRefs`.

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

## Listener Conflicts

Since multiple ListenerSets can attach to one Gateway, there are rules about how to manage conflicts.
A Listener can be **distinct** across any group of Listeners (within a single ListenerSet object, or in the group of Listeners attached to a Gateway) based on the combination of Port, Protocol, and (depending on the Protocol) Hostname.

When Listeners are not distinct, there are Listener precedence rules that are used to choose a is used to resolve the conflict, and choose which Listener takes effect. It's often easiest to think of this as a competition, with the "winner" being the Listener that takes effect. The rules go like this, continuing on ties.

* Listeners on the parent Gateway take priority over all others.
* The ListenerSet with the earliest creation time takes priority.
* The first ListenerSet alphabetically takes priority.
 
The winning ListenerSet is marked as `Accepted: true`, and the losing ListenerSet(s) are marked with `Accepted: false`, and `Conflidted: true`.

As with all the other conflict resolution rules in Gateway API, this is intended to provide traffic stability - so adding a new, conflicting ListenerSet will never take over an existing config.

## Examples

The following example shows a `Gateway` with an HTTP listener and two child HTTPS `ListenerSets`
with unique hostnames and certificates. Only `ListenerSets` in namespaces that have the `belongs: shared-gateway`
label will be accepted :

```yaml
{% include 'standard/listenerset/listenerset.yaml' %}
```
