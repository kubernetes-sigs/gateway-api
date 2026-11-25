---
title: "GEP-5489: Gateways for egress"
---

* Issue: [#5489](https://github.com/kubernetes-sigs/gateway-api/issues/5489). Related: [Issue #4747](https://github.com/kubernetes-sigs/gateway-api/issues/4747), [Issue #4704](https://github.com/kubernetes-sigs/gateway-api/issues/4704), [Issue #4748](https://github.com/kubernetes-sigs/gateway-api/issues/4748)
* Status: Provisional

(See [status definitions](../overview.md#gep-states).)

> **Note on history.** The present GEP continues the work of GEP-4747 ([PR #4746](https://github.com/kubernetes-sigs/gateway-api/pull/4746)). PR #4746 was closed in June 2026 to be [split into smaller, more focused proposals](https://github.com/kubernetes-sigs/gateway-api/pull/4746#issuecomment-4803827055). The first of them, address reachability, is now [GEP-5093](../gep-5093/index.md) (Experimental). The present GEP covers another, the indication of ingress or egress intent in the API, together with the forward-proxy model raised in [Issue #4704](https://github.com/kubernetes-sigs/gateway-api/issues/4704).

## TL;DR

Agree that a `Gateway` can be used for egress, and add an optional `direction` field to `Gateway` so that egress Gateways are explicit and discoverable. Define the two egress models of a Gateway: a **reverse proxy**, where Routes select pre-declared external destinations, and a **forward proxy**, where the Gateway itself is the proxy and needs no Routes. Authorization comes later, in a separate proposal.

## Goals

* Establish egress as a supported use of the `Gateway` resource. Egress needs no new Gateway-like resource.
* Make the traffic direction of a `Gateway` an explicit, portable property: a Gateway that declares a direction handles only that direction.
* Let users, agents and other tools find egress Gateways (and their cluster-internal addresses, via [GEP-5093](../gep-5093/index.md)) from the Gateway resource alone. This answers the user story [raised on PR #4746](https://github.com/kubernetes-sigs/gateway-api/pull/4746#issuecomment-4349943298): "As a user who wishes to use an Egress Gateway, how do I tell which Gateways are for egress?"
* Define **reverse-proxy egress**: an egress Gateway whose `HTTPRoute`s select declared external destinations. The Gateway processes requests as an ingress Gateway does, and the destinations are `Backend` resources of type `ExternalHostname` ([GEP-4894](../gep-4894/index.md)).
* Define **forward-proxy egress**: an egress Gateway that is itself a forward proxy. The destination comes from the request, and users create no Route or backend to represent the proxy or its destinations.

## Non-Goals

Out of scope for this GEP:

* Authorization policy. This GEP neither defines one nor specifies how one is evaluated on an egress Gateway.
* Requiring implementations to support a particular egress model. Each model is a separate conformance feature, so an implementation can claim one without the other.
* Mechanisms that force workloads to use the egress Gateway (NetworkPolicy, CNI, sidecars).
* The way a workload is steered to the egress Gateway: DNS, mesh or CNI redirection, client proxy settings. The present GEP defines what the Gateway does with a request that arrives at its address. [GEP-5093](../gep-5093/index.md) covers address reachability.
* Forward-proxy mechanisms other than the HTTP proxy protocol, for example forwarding by Host header or SNI, or transparent proxying. Implementations MAY offer them as implementation-specific behavior.
* Egress gateway chaining ([GEP-4751](https://github.com/kubernetes-sigs/gateway-api/issues/4751)).
* A new `EgressGateway` resource ([GEP-4748](https://github.com/kubernetes-sigs/gateway-api/issues/4748)); see [Alternatives](#alternatives).
* Defining or changing the `Backend` resource ([GEP-4894](../gep-4894/index.md)), including TLS to the destination (e.g. [Issue #5302](https://github.com/kubernetes-sigs/gateway-api/issues/5302)).
* Backend kinds other than `Backend` of type `ExternalHostname` on a reverse-proxy egress Gateway. Implementations MAY accept them as implementation-specific behavior.
* Gateways that handle both ingress and egress (mixed Gateways). This GEP standardizes Gateways that declare one direction.
* Route kinds other than `HTTPRoute` on egress Gateways (`GRPCRoute`, `TLSRoute`, `TCPRoute`, `UDPRoute`), and TLS passthrough.
* A listener that terminates TLS for an external hostname (TLS interception).
* Changing the behavior of any existing Gateway. The new `direction` field is optional, and Gateways that do not set it behave as they do today.
* A portable API for modifying requests on a forward-proxy Gateway (for example rewriting the Host from a custom header). Some implementations do this with Route filters today and MAY keep offering it as implementation-specific behavior.

## Introduction/Overview

### Terminology

* **Ingress Gateway**: accepts connections from clients and forwards them to backends that are exposed through Routes. The backends typically run in the cluster and belong to application teams (Ana). This is how Gateway API has worked so far.
* **Egress Gateway**: accepts connections from clients inside the cluster (or the mesh, or a private network) and forwards them to destinations *outside* the trust or administrative domain of those clients, usually the internet or third-party SaaS and AI providers.
* **Reverse-proxy egress**: Routes attached to the Gateway select a declared external destination. Clients can reach only destinations that have a Route.
* **Forward-proxy egress**: the Gateway is a forward proxy with dynamic routing. It takes the destination from the request and resolves it at request time, so users do not declare destinations one by one.
* **Explicit forward proxy**: a forward proxy that clients address on purpose, typically through proxy settings such as `HTTPS_PROXY` and `HTTP_PROXY`, and to which they send HTTP proxy requests.

### The problem

Applications, and AI agents above all, need managed and audited access to destinations outside the cluster: model providers, MCP servers, SaaS APIs, package registries. Implementations already use `Gateway` for this, each in its own way, and the API says nothing about it.

Since GEP-4747 was split, two building blocks have landed: [GEP-4894](../gep-4894/index.md) `Backend` (Experimental, with an Extended `ExternalHostname` type for external destinations) and [GEP-5093](../gep-5093/index.md) address `routability` (Experimental, with a `Cluster` value so that a Gateway can request and report a cluster-internal address).

The present GEP addresses three gaps:

1. **Agreement.** Two options have been proposed: reusing `Gateway` for egress (GEP-4747, [PR #4746](https://github.com/kubernetes-sigs/gateway-api/pull/4746)) and a dedicated `EgressGateway` resource (GEP-4748, [PR #4749](https://github.com/kubernetes-sigs/gateway-api/pull/4749)). No decision has been made between them.
2. **Intent.** Nothing on a Gateway says "this is an egress Gateway", although the two directions differ. For instance, they differ trust-wise: An ingress Gateway typically forwards to in-cluster workloads, whose certificates often come from a private CA, while an egress Gateway forwards to third parties, whose certificates typically come from public CAs.
3. **A model for forward proxies.** [GEP-4747 v1](https://github.com/kubernetes-sigs/gateway-api/pull/4746) and the [WG AI Gateway egress proposal](https://github.com/kubernetes-sigs/wg-ai-gateway/blob/main/proposals/10-egress-gateways.md) described reverse-proxy egress and left forward proxies for later ([Issue #4704](https://github.com/kubernetes-sigs/gateway-api/issues/4704)). Reverse-proxy egress fits destinations that are known in advance. Gateway API needs a way to represent a proxy for use cases where Chihiro cannot list the destinations ahead of time, such as agents and general-purpose workloads.

### What this GEP proposes

This GEP proposes the use of the `Gateway` resource for egress, with:

* A new optional field on `Gateway` that declares the direction of its traffic: ingress or egress. A Gateway that declares a direction handles only that direction.
* Two egress models for a Gateway (described below) and a new field that selects the model. Each model is a separate conformance feature. An implementation MAY support either model or both. An egress Gateway operates in one model, which applies to all of its listeners.
    * Reverse proxy: Routes select declared external destinations. For connection mechanisms, the existing `HTTP` and `HTTPS` listener protocols suffice.
    * Forward proxy: the Gateway itself is the proxy. This GEP standardizes the explicit forward proxy, to which clients send HTTP proxy requests: `CONNECT` requests, and [absolute-form](https://www.rfc-editor.org/info/rfc9112/#name-absolute-form) requests with an `http://` target. The existing `HTTP` and `HTTPS` listener protocols suffice for it as well.

### Egress models

|  | Reverse-proxy egress | Forward-proxy egress |
| :---- | :---- | :---- |
| What represents the proxy | The `Gateway` | The `Gateway` |
| What represents destinations | `HTTPRoute`s + `Backend` of type `ExternalHostname` ([GEP-4894](../gep-4894/index.md)) | Nothing: the destination comes from the request |
| Routes | Required: one per destination or group of destinations | Not needed |
| What limits where traffic can go | The set of Routes | Policy attached to the Gateway |
| Fits | Known, curated destinations | Open-ended destinations (agents, general-purpose workloads) |

#### Reverse-proxy egress

* A client sends an ordinary HTTP request to an address of the egress Gateway. The Gateway processes it as an ingress Gateway does: it matches the request against listeners and `HTTPRoute`s (`hostnames`, matches, filters) and forwards it to the backend that the matching rule references.
* The Gateway does not forward a request that matches no Route, as for ingress.
* The destination is a `Backend` of type `ExternalHostname`. [GEP-4894](../gep-4894/index.md) defines how the Gateway connects to it, including TLS. The present GEP adds nothing to it.
* An `HTTPRoute` can send requests to an external destination in two ways. This GEP covers both:
    * The request carries the external hostname in its Host header, and the Route's `hostnames` match it.
    * The request carries a hostname of the Gateway, and the Route rewrites Host to the external hostname with the existing `URLRewrite` filter.
* Routes that reference other backend kinds on an egress Gateway are implementation-specific. An implementation that does not accept such a reference reports it through the Route's `ResolvedRefs` condition, as today.

#### Forward-proxy egress

* The Gateway itself is the forward proxy. Its listener accepts requests from clients, and the Gateway determines the destination from each request and resolves it at request time.
* Clients use the Gateway as an explicit HTTP proxy, typically through proxy settings such as `HTTPS_PROXY` and `HTTP_PROXY`. This is the one forward-proxy mechanism that this GEP standardizes. A listener accepts two request forms:
    * `CONNECT` requests, which most clients send for `https://` destinations. The request names the destination host and port, and the Gateway opens a tunnel to it.
    * Absolute-form requests with an `http://` target (`GET http://host/path`), which most clients send for `http://` destinations. The Gateway forwards the request to the host named in the request target.
    * The behavior for other requests (e.g. absolute-form requests with an `https://` target) is implementation-specific.
* Before this GEP reaches Experimental, it will define the remaining details of this behavior, such as error responses.
* The Gateway forwards traffic without a Route or a backend: a forward proxy has no per-destination configuration for them to express.
* Policy attached to the Gateway restricts which clients may reach which destinations. That policy is either a standard policy, which a follow-up to the present GEP is to define (see [Future work](#future-work)), or an implementation-specific one. Whether the Gateway forwards requests when no policy is attached is an open question (see [Open questions](#open-questions)).
* Whether the Gateway can reach destinations inside the cluster is implementation-specific. This GEP standardizes behavior for external destinations.
* The listener of a forward-proxy Gateway sets no `hostname`. The behavior of a listener with a `hostname` is implementation-specific.
* The existing HTTPRoute text, "When no rules matching a request have been successfully attached to the parent a request is coming from, a HTTP 404 status code MUST be returned", dates from when Gateway API covered only ingress. This GEP clarifies that the text describes route-based request processing and does not apply to an egress Gateway operating as a forward proxy.
* This GEP does not forbid Routes on a forward-proxy Gateway. An implementation MAY let users attach Routes to it for implementation-specific purposes. Those Routes are outside the forward-proxy model that this GEP standardizes.

## Purpose (why and who)

* **Chihiro** needs to tell ingress and egress Gateways apart by their API, and give separate RBAC and review to egress Gateways, since those can reach the internet and may hold credentials. This should not depend on GatewayClass naming conventions.
* **Chihiro** also needs to offer an egress Gateway that fits the destinations of the workloads:
    * For a small set of well-known providers: a curated Gateway, with Routes and one `Backend` per provider.
    * For agents and other workloads that may reach thousands of hosts: a Gateway that they use as a forward proxy, without a Route and a backend for each host.
* **Ana** needs her tooling to find which Gateway in her cluster is the egress Gateway and which address to use. Today she has to read controller documentation.

### Why a `direction` field

A way to mark a Gateway as an egress Gateway has been requested in review of the earlier proposals (comments on PR #4746: [1](https://github.com/kubernetes-sigs/gateway-api/pull/4746#issuecomment-4349943298), [2](https://github.com/kubernetes-sigs/gateway-api/pull/4746#discussion_r3111659210); on PR #4749: [3](https://github.com/kubernetes-sigs/gateway-api/pull/4749#issuecomment-4412210184)). A field on `Gateway` does this without a new resource kind ([GEP-4748](https://github.com/kubernetes-sigs/gateway-api/issues/4748)), which would duplicate the whole Gateway API surface.

Some implementations let one Gateway serve both directions. Gateways that leave the `direction` field unset can keep doing that as implementation-specific behavior. Such mixed Gateways are out of scope: this GEP defines portable meaning for Gateways that declare one direction.

### Why forward-proxy egress needs no Routes

* **The Gateway already models a proxy.** Listeners, addresses and TLS settings describe where and how clients connect. A forward proxy needs nothing more: it has no set of backends to choose between, so a Route would have nothing to route.
* **A forward proxy leaves no destination for a Route to declare.** In reverse-proxy egress, a Route and its `Backend` declare each destination, so the set of Routes is the allowlist. A forward proxy accepts any destination. The decision that remains is which clients may reach which destinations, and it usually depends on the caller's identity, which Route matches cannot express. Policy attached to the Gateway expresses it, with the same mechanism as for ingress and mesh.
* **It keeps the two models easy to tell apart.** A Gateway with Routes reaches what its Routes name. A Gateway without Routes reaches what policy allows.

### Forward proxies configured through Routes today

Some implementations already offer a forward proxy with dynamic routing through Routes: an `HTTPRoute` whose single `backendRef` points to an implementation-specific "dynamic" backend ([Envoy Gateway](https://github.com/envoyproxy/gateway/blob/main/site/content/en/latest/tasks/traffic/backend.md#dynamic-forward-proxy) `Backend` of `type: DynamicResolver`, [agentgateway](https://github.com/agentgateway/website/blob/main/assets/agw-docs/pages/traffic-management/dfp.md) `dynamicForwardProxy: {}`). This GEP does not standardize that shape (see [Alternatives](#d-forward-proxy-as-a-backend-behind-a-route)), and it does not forbid it either: such implementations remain valid, and their Route-based configuration stays implementation-specific.

## Example API draft

This section is required from Implementable onwards. The sketch below is illustrative and non-binding; names and details are open (see [Open questions](#open-questions)).

```yaml
apiVersion: gateway.networking.k8s.io/v1
kind: Gateway
metadata:
  name: egress
  namespace: egress-system
spec:
  gatewayClassName: example-egress
  direction: Egress              # new, optional: Ingress | Egress
  egress:                        # required for `direction: Egress`
    model: ExplicitForwardProxy  # new: ReverseProxy | ExplicitForwardProxy
  addresses:
  - type: IPAddress
    routability: Cluster         # GEP-5093
  listeners:
  - name: proxy
    protocol: HTTP
    port: 3128
```

This Gateway is a complete forward proxy with no Route attached. A reverse-proxy egress Gateway sets `model: ReverseProxy` and has Routes attached.

Intended semantics:

* The `direction` field is optional. When it is unset, behavior is implementation-specific, so existing Gateways behave as before.
* A Gateway with a direction handles only that direction on all of its listeners.
* An implementation that does not support the requested direction for a given GatewayClass rejects the Gateway through status conditions.
* A GatewayClass MAY support only some `direction`s or only some egress `model`s. For example, an implementation can offer separate GatewayClasses for ingress and egress.
* An `Egress` Gateway SHOULD have addresses that are not reachable from the internet (for example `routability: Cluster`).
* A Gateway with `direction: Egress` selects its egress model in `egress.model`. The model applies to the whole Gateway. An implementation that does not support the requested model for a given GatewayClass rejects the Gateway through status conditions.
* The listeners of an egress Gateway use the existing `HTTP` or `HTTPS` protocols in both models. The listener protocol describes the connection between the client and the Gateway, not the scheme of the destination.
* On an `ExplicitForwardProxy` Gateway, an `HTTP` listener accepts `CONNECT` requests and absolute-form requests with an `http://` target. Whether `HTTPS` listeners are part of the conformance feature is an open question (see [Open questions](#open-questions)).
* Implementations MAY offer other forward-proxy mechanisms through implementation-specific values of `egress.model`, which use a domain prefix.
* One listener is enough for a fully functional egress Gateway, and conformance requires no more: tests use single-listener Gateways. On an egress Gateway with more listeners, each listener behaves as it would on a single-listener Gateway.

### Gateway for ingress (north-south)

Unchanged. `Ingress` makes today's semantics explicit.

### Gateway for mesh (east-west)

Unchanged. Gateways that some meshes use as waypoints or east-west gateways are neither ingress nor egress Gateways, and they leave the `direction` field unset (see [Open questions](#open-questions)).

## Conformance details

To be completed before Experimental. The intent is two features with Extended support: one for reverse-proxy egress and one for forward-proxy egress, so that an implementation can claim one model without the other. Each feature covers the `direction` field for its model. The reverse-proxy feature depends on the `Backend` `ExternalHostname` feature of [GEP-4894](../gep-4894/index.md). Tests allow an implementation to use separate GatewayClasses for ingress and egress.

## Future work

Authorization is the follow-up to the present GEP. In forward-proxy egress, a policy attached to the Gateway decides where traffic may go. Candidates for a standard policy include an extension of [GEP-3779](../gep-3779/index.md) to Gateways and the [kube-agentic-networking `AccessPolicy`](https://github.com/kubernetes-sigs/kube-agentic-networking/blob/main/api/v1alpha1/accesspolicy_types.go). Choosing or defining one is out of scope here.

## Alternatives

### A. Dedicated `EgressGateway` resource ([GEP-4748](https://github.com/kubernetes-sigs/gateway-api/issues/4748))

*For*: a distinct UX for egress, resource-level RBAC, room for egress-only fields. *Against*: duplicates the Gateway surface (listeners, addresses, status, policy attachment, Route parentRefs) and splits implementations and tooling. As GEP-4747 v1 showed, Gateway fields mean the same thing at the field level in both directions. A `direction` field keeps most of the benefits: intent is discoverable, and admins can restrict who creates egress Gateways with ValidatingAdmissionPolicy.

### B. Infer egress from backends (GEP-4747 v1 stance)

Rejected by maintainer feedback: users cannot tell from the Gateway itself whether its Routes target external destinations.

### C. Distinguish egress by `GatewayClass` only

A GatewayClass can signal egress through naming conventions. This would require implementations to introduce new GatewayClasses. Naming conventions are not self-documenting and are less discoverable than an explicit API field.

### D. Forward proxy as a backend behind a Route

Existing implementations ([Envoy Gateway](https://github.com/envoyproxy/gateway/blob/main/site/content/en/latest/tasks/traffic/backend.md#dynamic-forward-proxy), [agentgateway](https://github.com/agentgateway/website/blob/main/assets/agw-docs/pages/traffic-management/dfp.md)) configure a forward proxy with dynamic routing as an `HTTPRoute` with a catch-all rule whose `backendRef` is an implementation-specific dynamic backend. We propose not to standardize this, for these reasons:

1. **The Route does no routing.** The rule's `backendRef` stands for the proxy itself, not for a remote host, so one is enough and there is nothing to choose between. The Route and the backend restate what the Gateway already is.
2. **A dynamic backend names no destination.** In [GEP-4894](../gep-4894/index.md), a `Backend` describes how the Gateway connects to a destination: its address, port and TLS settings. A dynamic backend has none of these, because the destination comes from each request. It configures how the proxy behaves, which is the role of the Gateway.
3. **Limiting destinations is authorization.** Route matches cover request attributes and cannot refer to the caller's identity, so "workload A may reach X, workload B may reach Y" needs a policy in any case. Envoy Gateway's own documentation points to `SecurityPolicy` authorization for restricting destinations of its dynamic resolver; agentgateway's documentation likewise advises to "restrict clients with an authorization policy".
4. **Ownership.** With a catch-all Route to a dynamic backend, creating a Route is what opens egress to arbitrary destinations, which makes RBAC on Routes the security boundary. Gateway API delegates Route creation to Ana, while egress filtering should be owned by Chihiro.
5. **Complexity for users.** Users would configure a Gateway, a Route and a backend of a new destination-less type, and still need an authorization policy for identity-based control. The model proposed here needs only the Gateway and the authorization policy.
6. **API fragmentation and redundancy.** The new destination-less backend type would be specific to forward proxies. It would add an API type with a single use, and it would repeat what the `egress.model` field of the Gateway already says.

This model gives up Route filters that modify requests on their way through a forward proxy (for example Envoy Gateway's host rewriting from a custom header). The same functionality can be configured at the Gateway level, through an implementation-specific API or a future standard one (see [Non-Goals](#non-goals)). This GEP does not forbid Routes on a forward-proxy Gateway, so Route filters on it remain implementation-specific.

## Open questions

1. **Names and values.** `direction`, `trafficDirection` or `type`? Should the field accept only `Ingress` and `Egress`, or also implementation-specific values with a domain prefix, as `routability` does in [GEP-5093](../gep-5093/index.md)? Prefixed values would let implementations label Gateways that are neither, such as mesh waypoints. What should the `egress.model` field and its values be called, and should the field accept implementation-specific values with a domain prefix as well?
2. **Mutability.** Should Gateway API make the `direction` field and the `egress.model` field immutable once set, allow only `unset → value`, or leave mutability to implementations?
3. **Default without a policy.** Should a forward-proxy Gateway with no policy attached forward requests (allow-by-default) or reject them (deny-by-default)? With allow-by-default, conformance tests can run without an authorization policy.
4. **TLS to the Gateway.** Should the forward-proxy conformance feature cover `HTTPS` listeners (TLS between the client and the Gateway) or only `HTTP` listeners?
5. **A feature for the `direction` field alone.** The two conformance features cover the `direction` field only for implementations that support an egress model. Is a third feature needed just for the field, possibly with Core support, so that an implementation without egress support is tested for rejecting `direction: Egress` rather than treating the Gateway as an ingress Gateway? Should that feature also cover status that reports the effective direction of Gateways that leave the `direction` field unset?

## References

* [GEP-4747](https://github.com/kubernetes-sigs/gateway-api/issues/4747): L7 Reverse-Proxy Egress Gateway Support ([PR #4746](https://github.com/kubernetes-sigs/gateway-api/pull/4746); [comment on its split into separate GEPs](https://github.com/kubernetes-sigs/gateway-api/pull/4746#issuecomment-4803827055))
* [GEP-4748](https://github.com/kubernetes-sigs/gateway-api/issues/4748): EgressGateway Resource ([PR #4749](https://github.com/kubernetes-sigs/gateway-api/pull/4749))
* [Issue #4704](https://github.com/kubernetes-sigs/gateway-api/issues/4704): Egress in the forward proxy model with dynamic routing
* [GEP-5093](../gep-5093/index.md): Gateway Address Routability (Experimental)
* [GEP-4894](../gep-4894/index.md): Backend Resource (Experimental)
    * [Issue #5302](https://github.com/kubernetes-sigs/gateway-api/issues/5302): Interaction between Gateway and XBackend `clientCertificateRef` (related; tracked there)
* [GEP-3779](../gep-3779/index.md): Identity Based Authz for east-west traffic (Implementable)
* [kube-agentic-networking `AccessPolicy`](https://github.com/kubernetes-sigs/kube-agentic-networking/blob/main/api/v1alpha1/accesspolicy_types.go)
* [WG AI Gateway egress proposal](https://github.com/kubernetes-sigs/wg-ai-gateway/blob/main/proposals/10-egress-gateways.md): Egress Gateways (reverse-proxy egress model)
* [Istio egress gateway](https://github.com/istio/istio.io/blob/master/content/en/docs/tasks/traffic-management/egress/egress-gateway/index.md): reverse-proxy egress with Routes and `ServiceEntry`
* Forward proxies configured through Routes: [Envoy Gateway](https://github.com/envoyproxy/gateway/blob/main/site/content/en/latest/tasks/traffic/backend.md#dynamic-forward-proxy), [agentgateway](https://github.com/agentgateway/website/blob/main/assets/agw-docs/pages/traffic-management/dfp.md)
