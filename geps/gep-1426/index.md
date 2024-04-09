# GEP-1426: xRoutes Mesh Binding

* Issue: [#1294](https://github.com/kubernetes-sigs/gateway-api/issues/1294)
* Status: Standard

## Overview

Similar to how `xRoutes` bind to `Gateways` and manage North/South traffic flows in Gateway API’s ingress use-case, it would be natural to adopt a similar model for traffic routing concerns in service mesh deployments. The purpose of this GEP is to add a mechanism to the Gateway API spec for the purpose of associating the various `xRoute` types to a service mesh and offering a model for service owners to manage traffic splitting configurations.

This GEP is intended to establish an implementable, but experimental, baseline for supporting basic service mesh traffic routing functionality through the Gateway API spec.

## Personas

This GEP uses the [roles and personas](https://gateway-api.sigs.k8s.io/concepts/security-model/#roles-and-personas) defined in the Gateway API security model, and the service "producer" and "consumer" roles defined in [GEP-1324: Service Mesh in Gateway API](https://gateway-api.sigs.k8s.io/geps/gep-1324/#producer-and-consumer).

## Goals

* MUST allow `xRoute` traffic rules to be configurable for a mesh service by the application owner/producer.
* SHOULD allow control by the cluster operator (mesh administrator) to grant permission for whether `xRoute` resources in a given namespace are allowed to configure mesh traffic routing.
* SHOULD NOT require downstream "consumer" services to update configuration or DNS addresses for traffic to follow "producer" mesh routing rules configured by upstream services.
* SHOULD NOT require reconfiguring existing `xRoute` resources for North/South Gateway configuration.

## Non-Goals

* Supporting "egress" use cases, which is currently a deferred goal, including:
    * Defining how "consumer" traffic rules which could override routing for service upstreams only within the local scope of a namespace or service might be configured.
    * Redirecting calls from arbitrary custom domains to an in-cluster service.
* Defining how multiple `Services` or `EndpointSlices` representing instances of a single "logical" service should present an identity for AuthN/AuthZ or be associated with each other beyond routing rules.
* Defining how AuthZ should be implemented to secure East/West traffic between services.
* Defining how [Policy Attachment](https://gateway-api.sigs.k8s.io/reference/policy-attachment/) would bind to `xRoute`, services or a mesh.
* Defining how `Routes` configured for East/West service mesh traffic management might integrate with North/South `Gateways`.
    * This is a bit tricky in that it's effectively a form of delegation as described in [GEP-1058: Route Inclusion and Delegation](https://github.com/kubernetes-sigs/gateway-api/pull/1085), and is planned to be explored in a future GEP.
* Handling East/West traffic outside the cluster (VMs, etc).

## Implementation Details and Constraints

* MUST set a status field on `xRoute` to show if the routing configuration has been applied to the mesh.
* MUST only be allowed to configure "producer" traffic rules for a `Service` in the same namespace as the `xRoute`.
    * Traffic routing configuration defined in this way SHOULD be respected by ALL consumer services in all namespaces in the mesh.
* MAY assume that a mesh implements "transparent proxy" functionality to redirect calls to the Kubernetes DNS address for a `Service` through mesh routing rules.

## Introduction

It is proposed that an application owner should configure traffic rules for a mesh service by configuring an `xRoute` with a Kubernetes `Service` resource as a `parentRef`.

This approach is dependent on both the "frontend" role of the Kubernetes `Service` resource as defined in [GEP-1324: Service Mesh in Gateway API](https://gateway-api.sigs.k8s.io/geps/gep-1324/#service) when used as a `parentRef` and the "backend" role of `Service` when used as a `backendRef`. The conformant implementation would use the Kubernetes `Service` name to match traffic for meshes, but the `backendRef` endpoints would ultimately be used for the canonical IP address(es) to which traffic should be redirected by rules defined in this `xRoute`. This approach leverages the existing points of extensibility within the Gateway API spec, and would not require introducing any API changes or new resources, only defining expected behavior.

### Why Service?

The GAMMA initiative has been working to bring service mesh use-cases to the Gateway API spec, taking the best practices and learnings from mesh implementations and codifying them in a spec. Most mesh users are familiar with using the Kubernetes `Service` resource as the foundation for traffic routing. Generally, this architecture makes perfect sense; unfortunately, `Service` is far too coupled of a resource. It orchestrates IP address allocation, DNS, endpoint collection and propagation, load balancing, etc. For this reason, it **cannot** be the right long-term answer for `parentRef` binding; however, it is the only feasible option that Kubernetes has for mesh implementations today. We expect this to change (indeed, we hope to be a part of that change), but in the interest of developing a spec now, we must once again lean on the `Service` resource. However, we will provide provisions to support additional resources as a `parentRef`.

## API

```yaml
metadata:
  name: foo-route
  namespace: store
spec:
  parentRefs:
  - kind: Service
    name: foo
  rules:
    backendRefs:
    - kind: Service
      name: foo
      weight: 90
    - kind: Service
      name: foo-v2
      weight: 10
```

In the example above, routing rules have been configured to direct 90% of traffic for the `foo` `Service` to the default "backend" endpoints specified by the `foo` `Service` [`selector`](https://kubernetes.io/docs/concepts/services-networking/service/#defining-a-service) field, and 10% to the `foo-v2` `Service`. This is determined based on the `ClusterIP` (for `Service`) and `ClusterSetIP` (for `ServiceImport`) matching, and for "transparent proxy" mesh implementations would match all requests to `foo.svc.cluster.local` (or arbitrary custom suffix, as the hostname is not specified manually) from within the same namespace, all requests to `foo.store.svc.cluster.local` from other namespaces, and all requests to `foo.store.svc.clusterset.local` for multicluster services, within the scope of the service mesh.

### Route presence

When no `xRoute` resources are defined, all traffic should implicitly work - this is just how Kubernetes functions. When you create an `xRoute` targeting a service as a `parentRef` you are replacing that implicit logic - not adding to it. Therefore, you may be reshaping or restricting traffic via an `xRoute` configuration (which should be noted, is distinct from *disallowing* traffic by AuthZ).

### Allowed service types

Services valid to be selected as a `parentRef` SHOULD have a way to identify traffic to them - typically by one or more virtual IP(s), DNS hostname(s), or name(s).

Implementations SHOULD support the default [`ClusterIP`](https://kubernetes.io/docs/concepts/services-networking/service/#publishing-services-service-types) `Service` type as a `parentRef`, with or without selectors.

["Headless" `Services`](https://kubernetes.io/docs/concepts/services-networking/service/#headless-services) SHOULD NOT be supported as a `parentRef`, because they do not implement the "frontend" functionality of a service.

`Service` resource with [`type: NodePort`](https://kubernetes.io/docs/concepts/services-networking/service/#type-nodeport) or [`type: LoadBalancer`](https://kubernetes.io/docs/concepts/services-networking/service/#loadbalancer) MAY be allowed as `parentRefs` or `backendRefs`, as these do provision virtual IPs and are effectively `ClusterIP` services with additional functionality, but it should typically be preferred to expose services publicly through the North/South Gateway API interfaces instead.

`Service` resources with [`type: ExternalName`](https://kubernetes.io/docs/concepts/services-networking/service/#externalname) SHOULD NOT be allowed as `parentRefs` or `backendRefs` due to [security concerns](https://github.com/kubernetes/kubernetes/issues/103675), although might eventually play some role in [configuring egress functionality](https://github.com/kubernetes-sigs/gateway-api/issues/1070).

Services supported as `backendRefs` SHOULD be consistent with expectations for North/South Gateway API implementations, and MUST have associated endpoints. `ClusterIP` `Services` with selectors SHOULD be supported as a `backendRef`.

#### `Service` without selectors

An alternate pattern additionally supported by this approach would be to target a `Service` without selectors as the `parentRef`. This could be a clean way to create a pure routing construct and abstract a logical frontend, as traffic would resolve to a `backendRef` `Service` with selectors defined on the `HTTPRoute`, or receive a 4xx/5xx error response if no matching path or valid backend was found.

### `parentRef` Conformance Levels

Currently (v0.7.0), this spec only considers the `Service` resource to be under Core conformance as a `parentRef`. However, Service is not the only resource that can fulfill the frontend role. While the Gateway API spec couldn’t possibly enumerate every existing (and future) frontend-like resource, it can specify a subset of resources that implementations MUST support as parentRefs under as a part of core conformance. Meshes MAY support other implementation-specific resources as parentRefs. The spec maintainers also reserve the right to add additional resources to core conformance as the spec evolves.

#### Extended Conformance

In addition to Service, there are other optional parentRef resources that, if used by implementations, MUST adhere to the spec’s prescriptions. At the time of writing (v0.7.0), there is one resource in extended conformance: `ServiceImport` (part of the [MCS API](https://github.com/kubernetes-sigs/mcs-api), currently in alpha). The semantics of `ServiceImport` `parentRef` binding can be found in [GEP-1748](https://gateway-api.sigs.k8s.io/geps/gep-1748/) (Note: Headless `ServiceImport` is out of scope and not currently a part of the spec).

##### Why not `IPAddress`

In Kubernetes 1.27, there will be a new IPAddress resource added to networking.k8s.io/v1alpha1 as part of [KEP 1880](https://github.com/kubernetes/enhancements/tree/master/keps/sig-network/1880-multiple-service-cidrs#proposal). Naturally, it makes sense to examine whether or not this new resource makes sense as a GAMMA aware parentRef. At first glance, IPAddress seems to be an appropriate abstraction for the “frontend” role we’ve been discussing; every Kubernetes Service is accessed over the network via one of its ip addresses. Furthermore, the fact that the Service resource auto-creates an IPAddress is encouraging. However, the fact that the name of the IPAddress is simply the decimal/hex ip address and not a human-readable Service name makes the UX untenable as a spec-supported parentRef. However, `IPAddress` is NOT disallowed; implementations may use it if they wish.

#### Implementation-specific `parentRef`s

If mesh implementations wish to enable an implementation-specific resource as a parentRef, they may do so as long as that resource meets certain conditions. Recall that the frontend role of a (generic) service is how one calls the service. In the service mesh transparent proxy context, the frontend role (and parentRef by extension) is effectively the matching mechanism for the specified route. For the Service parentRef, this means that the mesh should apply a particular xRoute’s configuration if the destination ip address for a given connection is the ClusterIP of that parentRef Service. If a mesh wishes to use an implementation-specific resource for parentRef, that resource MUST contain layer-appropriate information suitable for traffic matching (e.g. no Host header capture in TCPRoute). For example, the following HTTPRoute with an Istio `ServiceEntry` as a parentRef would be a valid implementation-specific reference:

```yaml
apiVersion: networking.istio.io/v1beta1
kind: ServiceEntry
metadata:
  name: internal-svc-httpbin
  namespace : egress
spec:
  hosts:
  - example.com
  exportTo:
  - "."
  location: MESH_INTERNAL
  ports:
  - number: 80
    name: http
    protocol: HTTP
  resolution: DNS
---
apiVersion: gateway.networking.k8s.io/v1beta1
kind: HTTPRoute
metadata:
  name: mongo-internal
spec:
  parentRefs:
  - kind: ServiceEntry
    group: networking.istio.io/v1beta1
    name: internal-svc-httpbin
    namespace: egress
    sectionName: http # referencing the port name
  rules:
  - backendRefs:
    - name: internal-example
      port: 80
```

##### `Gateways`

There has been much discussion around cluster local Gateways (i.e. Gateways not associated with a traditional load balancer). While there are various potential UX impairments (e.g. what’s the difference between a GAMMA HTTPRoute with a Gateway parentRef and an ingress implementation’s HTTPRoute?), there is no technical reason why a Gateway cannot be a valid GAMMA parentRef if an implementation wishes to do so.

### Route types

All types currently defined in the gateway-api core (`HTTP`, `GRPC`, `TCP`, `TLS`, and `UDP`) are available for use in a Mesh implementation.

If multiple routes with different types both bind to the same Service and Port pair, only a single route type should be applied. The rejected routes should be ignored and have the `RouteConditionAccepted` status set to the (new) reason `RouteReasonConflicted`.

Route type specificity is defined in the following order (first one wins):

1. GRPCRoute
2. HTTPRoute
3. TLSRoute
4. TCPRoute

Because UDP is its own protocol, it is orthogonal to these precedence order. Since there is only one UDP-based route, there is currently no conflicts possible; if other UDP-based routes are added a similar ordering will be defined.

Note: these conflicts only occur when multiple *different* route types apply to the same Service+Port pair. Multiple routes of the same type are valid, and merge according to the route-specific merging semantics.

### Ports

By default, a `Service` attachment applies to all ports in the service. Users may want to attach routes to only a *specific* port in a Service. To do so, the `parentRef.port` field should be used.

If `port` is set, the implementation MUST associate the route only with that port.
If `port` is not set, the implementation MUST associate the route with all ports defined in the Service.

### `hostnames` field

GAMMA implementations SHOULD NOT infer any functionality from the `hostnames` field on `xRoute`s (currently, `TLSRoute`, `HTTPRoute`, and `GRPCRoute` have this field) due to current under-specification and reserved potential for future usage or API changes.

For the use case of filtering incoming traffic from selected HTTP hostnames, it is recommended to guide users toward configuring [`HTTPHeaderMatch`](https://gateway-api.sigs.k8s.io/v1alpha2/reference/spec/#gateway.networking.k8s.io%2fv1beta1.HTTPHeaderMatch) rules for the [`Host`](https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Host) header. Functionality to be explored in future GEPs may include supporting concurrent usage of an `xRoute` traffic configuration for multiple North/South `Gateways` and East/West mesh use cases or redirection of egress traffic to an in-cluster `Service`.

### Namespace boundaries

In a mesh, routes can be configured by two personas:

* Service producers, who want to modify behavior of *inbound* requests to their services
* Service consumers, who want to modify behavior of *outbound* requests to other services.

While these concepts are not directly exposed in the API, a route is implicitly fulfilling one of these roles and behaves differently depending on the role.

A route is a producer route when the `parentRef` refers to a service in the *same namespace*. This route SHOULD apply to all incoming requests to the service, including from clients in other namespaces.

Note: Some implementations may only be able to apply routes on client-side proxies. As a result, these will likely only apply to requests from clients who are also in the mesh.

A route is a consumer route when the `parentRef` refers to a service in *another namespace*. Unlike producer routes, consumer routes are scoped only the same namespace. This ensures that for traffic between two namespaces, another unrelated namespace cannot modify their traffic.

Routes of either type can send traffic to `backendRefs` in any namespace. Unlike `Gateway` bound routes, this is allowed without a `ReferenceGrant`. In `Gateway`-bound routes (North-South), routes are opt-in; by default, no Services are exposed (often to the public internet), and a service producer must explicitly opt-in by creating a route themselves, or allowing another namespace to via `ReferenceGrant`. For mesh, routes *augment* existing Services, rather than exposing them to a broader scope. As a result, a `ReferenceGrant` is not required in most mesh implementations. Access control, if desired, is handled by other mechanism such as `NetworkPolicy`. While uncommon, if a mesh implementation *does* expose the ability to access a broader scope than would otherwise be reachable, then `ReferenceGrant` must be used for cross namespaces references.

### Multiple routes for a Service

A service may be used as a `parentRef` (where we attach to the "Service Frontend") or as a `backendRef` (where we attach to the "Service Backend").

In general, when a request is sent to a Service frontend (ex: `curl svc`), it should utilize a Route bound to that Service.
However, when sent to a Service backend (ex: `curl pod-ip`), it would not.

Similarly, if we have multiple "levels" of Routes defined, only the first will be used, as that is the only one that accesses the Service frontend.

Consider a cluster with routes for a Service in both a Gateway, consumer namespace, and producer namespace:

* Requests from the Gateway will utilize the (possibly merged) set of routes attached to the Gateway
* Requests from a namespace with consumer routes will utilize the (possibly merged) set of routes in the consumer namespace
* Requests from other namespaces will utilize the (possibly merged) set of routes in the producer namespace

The merging of routes occurs only within groups of the same type of routes (Gateway bound, producer, or consumer), and follows the standard route merging behavior already defined.

Note: a possible future extension is to allow `backendRefs` to explicitly target a "frontend" or "backend". This could allow chaining multiple routes together. However, this is out of scope for the current GEP.

### Drawbacks

* The fact that this pattern is used for mesh configuration is implicit - this may benefit from some additional configuration to map the `HTTPRoute` to a particular mesh implementation rather than being picked up by any or all GAMMA meshes present in a cluster. Possible approaches include:
* [GEP-1282: Describing Backend Properties](https://gateway-api.sigs.k8s.io/geps/gep-1282/) may be one path to associating a `Service` with a mesh, but likely wouldn't be able to handle the application of multiple `HTTPRoutes` for the same `Service`, but each intended for different mesh implementations
    * It's currently unclear how relevant this constraint may be, but associating an `HTTPRoute` with a mesh by this method would additionally require an extra graph traversal step.
* Expecting a `Mesh` `parentRef` or similar reference as proposed in [GEP-1291: Mesh Representation](https://docs.google.com/document/d/1oyA9uUH7pNNxxwy3WZGSWx-edHDBLrujcezr8q3el70/edit#) may be a preferred eventual path forward, but wouldn't be required initially, with the assumption that only one mesh should typically be present in a cluster.
* No mechanism for egress redirection of traffic from arbitrary hostnames to a mesh service within this approach (but could still be implemented separately).

## Alternatives

### New `MeshService` (or `HttpService`, `VirtualService`, `ServiceBinding`) resource as `parentRef`

Introduce a new resource to represent the "frontend" role of a service as defined in [GEP-1291: Mesh Representation](https://docs.google.com/document/d/1oyA9uUH7pNNxxwy3WZGSWx-edHDBLrujcezr8q3el70/edit#).

#### Controller manages new DNS hostname

A controller could create a matching selector-less `Service` (i.e. no endpoints), to create a `.cluster.local` name, or could interact with [external-dns](https://github.com/kubernetes-sigs/external-dns) to create a DNS name in an owned domain.

Ownership/trust would remain based on naming pattern: `serviceName.namespace.svc.[USER_DOMAIN]`

Separate `HttpService`, `TlsService` and `TcpService` resources could have the benefit of allowing us to define protocol specific elements to the spec along with an embedded `CommonServiceSpec`, similar to [`CommonRouteSpec`](https://gateway-api.sigs.k8s.io/v1alpha2/reference/spec/#gateway.networking.k8s.io/v1.CommonRouteSpec), and keep similar patterns as `Service`.

##### Drawbacks

* May require reconfiguring existing applications to point to a new mesh service hostname - adoption wouldn't be "transparent".
* The pattern of creating a new pure routing construct would still be implementable following the proposed approach, by manually creating and targeting a new `Service` without selectors as a `parentRef`, without the overhead of introducing a new resource.

#### Manage DNS by binding to an existing `Service`

A new `ServiceBinding` resource would directly reference an existing `Service` to determine which traffic should be intercepted and redirected following configured service mesh routing rules and facilitate "transparent proxy" functionality. This resource could possibly share similar responsibilities as the need identified in [GEP-1282: Describing Backend Properties](https://gateway-api.sigs.k8s.io/geps/gep-1282/).

```
kind: ServiceBinding
metadata:
  name: foo_binding
spec:
  parentRefs:
  - kind: Service
    name: foo
---
spec:
  parentRefs:
  - kind: ServiceBinding
    name: foo_binding
  rules:
    backendRefs:
    - kind: Service
      name: foo
      weight: 90
    - kind: Service
      name: foo_v2
      weight: 10
```

While the `HTTPRoute` does not directly reference a particular mesh implementation in this approach, it would be possible to design the `ServiceBinding` resource to specify that.

##### Drawbacks

* Introduces an extra layer of abstraction while still having several of the same fundamental drawbacks as a direct `parentRef` binding to `Service`.
* May require reconfiguring `Gateway` `HTTPRoutes` to specify `ServiceBindings` as `backendRefs`.

#### Drawbacks

* The split frontend/backend role of `Service` is fundamentally an issue with the `Service` resource, and while upstream changes may be quite slow, this would likely be best addressed through an upstream KEP - introducing a new resource to GAMMA now would likely result in API churn if we expect a similar proposal to be upstreamed eventually.
* Adopting the proposed `Service` as `parentRef` approach wouldn't foreclose the possibility of migrating to a new frontend-only resource in the future, and wouldn't even require a breaking change to `HTTPRoute`, just adding support for a new `parentRef` `Kind`.
* Would be less clear how to integrate with transparent proxy functionality - it may be possible to design some way to select a `Service` or hostname to intercept, but abstraction through a separate resource would make configuration more complex.

### `Mesh` resource as `parentRef`

This binds an `HTTPRoute` directly to a cluster-scoped `Mesh` object as defined in [GEP-1291: Mesh Representation](https://docs.google.com/document/d/1oyA9uUH7pNNxxwy3WZGSWx-edHDBLrujcezr8q3el70/edit#).

```
spec:
  parentRefs:
  - kind: Mesh
    name: cool-mesh
```

It is currently undefined how this approach may interact with either explicitly configured [`hostnames`](https://gateway-api.sigs.k8s.io/v1alpha2/reference/spec/#gateway.networking.k8s.io/v1.HTTPRouteSpec) or implicit "transparent proxy" routing for Kubernetes `Services` to determine how traffic should be intercepted and redirected.

This approach is not entirely abandoned, as it could supplement the proposed approach if explicit attachment to a specific mesh is deemed necessary. Additionally, this approach may offer a future option for attaching an `HTTPRoute` to a mesh, but not a specific service (e.g. to implement mesh-wide egress functionality for all requests to a specific hostname).

#### Peer to `Service` resource `parentRef`

An `HTTPRoute` could specify a `Mesh` resource `parentRef` as a peer to a `Service` resource `parentRef`.

```
spec:
  parentRefs:
  - kind: Mesh
    name: cool-mesh
  - kind: Service
    name: foo
```

##### Drawbacks

* Would require separate `HTTPRoute` resources to explicitly define _different_ traffic routing rules for the same service on different meshes.

#### Nested `services` and `hostnames` fields in [`ParentReference`](https://gateway-api.sigs.k8s.io/v1alpha2/reference/spec/#gateway.networking.k8s.io/v1.ParentReference)

In core conformance, the `services` would only be valid for `Mesh` types, and `hostnames` field only for `Gateway`. Mesh implementations could still use a `Host` header match if they wanted limit rules to specific hostnames.

```
parentRefs:
- kind: Mesh
  name: coolmesh
  services:
  - name: foo
    kind: Service
- kind: Gateway
  name: staging
  hostnames: [staging.example.com]
- kind: Gateway
  name: prod
  hostnames: [prod.example.com]
# Top level hostnames field removed
```

Moving the `hostnames` field from `HTTPRoute` to `ParentReference` might introduce a clean path for concurrently using a route across North/South and mesh use cases,  even without introducing the `services` field or a new `Mesh` resource, and even makes pure North/South implementations more flexible by allowing a hostname-per-`Gateway` scope.

##### Drawbacks

* Substantial API change, impacting even North/South use cases
* Extending this functionality to support mesh-wide egress or arbitrary redirection may still require some sort of bidirectional handshake with a `Hostname` resource to support configuration across namespaces and limit conflicting configuration.


#### Phantom `parentRef`
```
spec:
  parentRefs:
  - kind: Mesh
    name: istio
```

This is done by configuring the `parentRef`, to point to the `istio` `Mesh`. This resource does not actually exist in the cluster and is only used to signal that the Istio mesh should be used. In Istio's [experimental implementation](https://istio.io/latest/docs/tasks/traffic-management/ingress/gateway-api/#mesh-traffic), the `hostnames` field on `HTTPRoute` is used to match mesh service traffic to the routing rules.

### New field on `HTTPRoute` for `Service` binding

A new field `serviceBinding` would be added to `HTTPRoute` to attach to the `Service`. Alternatively, this could be a new field in [`HTTPRouteMatch`](https://gateway-api.sigs.k8s.io/v1alpha2/reference/spec/#gateway.networking.k8s.io/v1.HTTPRouteMatch). As with the proposed implementation, this approach could be combined with a `Mesh` resource or similar as the `parentRef`, which would just define that the route would be applied to a mesh.

```
spec:
  serviceBinding:
    name: my-service
```

OR

```
spec:
  matches:
    service:
      name: my-service
```

For either implementation, the type of the `serviceBinding` or `service` field should likely be a struct with `Group` (defaulting to the Kubernetes core API group when unspecified), `Kind` (defaulting to `Service` when unspecified) and `Name` fields, to allow for extensibility to `ServiceImport` or custom mesh service types.

#### Drawbacks

* API addition required, which is somewhat awkwardly ignored for North/South use cases, and could complicate potential for concurrent use of an `HTTPRoute` across both North/South and mesh use cases.
* Adding new fields to a relatively stable resource like `HTTPRoute` could be difficult to do in an experimental way.
* Following this pattern may lead to subsequent additional fields to further clarify or extend behavior.

### `Gateway` resource with `class: mesh` as `parentRef`

To support arbitrary DNS names (owned by a "domain owner persona") we would need a similar mechanism to what `Gateway` is using for delegating management of `HTTPRoutes` to namespaces. Instead of duplicating everything - we could use `Gateway` as is, with `class: mesh` (or matching the mesh implementation desired name).

```
kind: Gateway
spec:
  class: mesh
  listeners:
  - name: example
    hostname: "example.com"
---
kind: HTTPRoute
spec:
  parentRefs:
    name: foo_gateway
    sectionName: example
  hostnames: ["example.com", "foo.svc.cluster.local"]
```

Functionally such a mesh could be implemented using the existing gateway spec - a GAMMA implementation would only remove the extra hop through the `Gateway`, using sidecars, or it may use a specialized per-namespace gateway to isolate the mesh traffic (like [Istio Ambient](https://istio.io/latest/blog/2022/introducing-ambient-mesh/)). Proxyless gRPC could also use this to route directly.

This solution could work well for both non-`cluster.local` names but also for egress, where a `Gateway` with `class: egress` could define names that are external to the mesh and need to either have policies applied or go to a dedicated egress gateway.

#### Drawbacks

* Using the `HTTPRoute` `hostnames` field to match mesh traffic breaks from the typical Gateway API pattern of explicit Kubernetes resource references, is extremely implicit, and could reduce portability of configuration.
* Potentially unclear translation between conceptual resource and concrete implementation, particularly for "proxyless" mesh implementations.
* Service meshes may wish to express [egress](https://istio.io/latest/docs/tasks/traffic-management/egress/egress-gateway/) or [other "in-mesh" gateways](https://www.consul.io/docs/connect/gateways) through an API like this, and it could be confusing to overload this resource too much or conflate different personas who may wish to manage mesh service traffic routing as an application owner separately from egress rules as a service consumer or cluster operator.

### `ServiceProjection` resource as `parentRef` and `backendRefs`

This approach is similar to the above `ServiceBinding` proposal with a couple of major differences:

* `ServiceProjection` encapsulates both "frontend" and "backend" roles of the `Service` resource
* `ServiceProjection` could handle the full responsibilities described in [GEP-1282: Describing Backend Properties](https://gateway-api.sigs.k8s.io/geps/gep-1282/)

```
kind: ServiceProjection
metadata:
    name: foo
    namespace: store
spec:
    serviceRef:
        name: foo
        kind: Service|ServiceImport
    roles:
        frontend:
       backend:
            loadbalancerConfig:
                strategy: RoundRobin
             clientTLS:
                secretRef:
                    ...
---
kind: HTTPRoute
metadata:
  name: foo_route
  namespace: store
spec:
  parentRefs:
  - kind: ServiceProjection
    name: foo
    role: frontend
  rules:
    backendRefs:
    - kind: ServiceProjection
      name: foo
      role: backend
      weight: 90
    - kind: ServiceProjection
      role: backend
      name: foo_v2
      weight: 10
```

For convenience, `ServiceProjection` could have a `meshRef` field that, when set instead of `serviceRef`, makes all configuration within the `ServiceProjection` apply to all services in the mesh (the mesh control plane would need to read the `Mesh` resource). Pursuant to the changes to status semantics in [GEP-1364: Status and Conditions Update](https://gateway-api.sigs.k8s.io/geps/gep-1364/), it is necessary for the route to attach to something; in this case, the route attaches to the specific role or profile of the `ServiceProjection` and the mesh control plane should update the route status to reflect that.

#### Drawbacks

* May require reconfiguring `Gateway` `HTTPRoutes` to specify `ServiceProjections` as `backendRefs`.
* Verbose boilerplate for each service.

### Implicit backendRef

An initial iteration of this GEP had the ability to omit a `backendRef` and have it implicitly be set the same as the `parentRef`.
This has been removed due to inconsistency with Gateway `parentRefs` and tight coupling of the "frontend" and "backend" roles.

Implementations MUST respect the standard `backendRef` rules as defined by the existing spec.
