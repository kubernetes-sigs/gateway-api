# GEP-4748: EgressGateway Resource

* Issue: [#4748](https://github.com/kubernetes-sigs/gateway-api/issues/4748)
* Status: Provisional

(See [status definitions](../overview.md#gep-states).)

[Chihiro]: https://gateway-api.sigs.k8s.io/concepts/roles-and-personas/#key-roles-and-personas
[Ian]: https://gateway-api.sigs.k8s.io/concepts/roles-and-personas/#key-roles-and-personas
[Ana]: https://gateway-api.sigs.k8s.io/concepts/roles-and-personas/#key-roles-and-personas

The key words "MUST", "MUST NOT", "REQUIRED", "SHALL", "SHALL NOT", "SHOULD",
"SHOULD NOT", "RECOMMENDED", "NOT RECOMMENDED", "MAY", and "OPTIONAL" in this
document are to be interpreted as described in BCP 14 ([RFC8174]) when, and
only when, they appear in all capitals, as shown here.

[RFC8174]: https://www.rfc-editor.org/rfc/rfc8174

> **Note**: This GEP is Provisional. It proposes a dedicated `EgressGateway`
> resource as an alternative to reusing the existing `Gateway` resource for
> egress (see companion [GEP-4747]). Community feedback on both approaches is
> explicitly requested.

[GEP-4747]: ../gep-4747/index.md

## TLDR

Introduce a dedicated `EgressGateway` resource for L7 reverse-proxy egress
traffic management. While Gateway semantics are largely identical for ingress
and egress at the field level, this GEP argues that a separate resource
provides clearer user experience, stronger guardrails, and design space for
egress-specific concerns that may emerge as the pattern matures.

## User Stories

### Platform Operator

> **[Ian] needs to deploy an egress gateway that is structurally distinct from
> ingress gateways, preventing misconfiguration and making egress
> infrastructure immediately identifiable in cluster inventory.**

### Application Developer

> **[Ana] needs to understand at a glance whether a gateway handles inbound or
> outbound traffic, without inspecting GatewayClass parameters or controller
> documentation.**

### Cluster Administrator

> **[Chihiro] needs to apply RBAC policies that distinguish between who can
> create ingress gateways vs egress gateways, using standard Kubernetes
> resource-level permissions rather than GatewayClass-specific admission
> control.**

## Goals

* Introduce `EgressGateway` as a dedicated resource for egress traffic
* Reuse Gateway API types maximally -- share listener, address, and route
  attachment semantics with Gateway
* Provide structural guardrails against ingress/egress misconfiguration
* Reserve design space for egress-specific fields as the pattern matures
* Enable simple RBAC separation between ingress and egress gateway creation

## Non-Goals

* Define the Backend resource (see [PR #4488](https://github.com/kubernetes-sigs/gateway-api/pull/4488))
* Define significantly different listener or routing semantics from Gateway
* Address forward-proxy egress, L3/L4 egress, or mesh-attached egress
* Deprecate or replace Gateway for any existing use case including in
  existing egress implementations.
* Prescribe a single egress architecture (this resource supports both
  Endpoint and Parent routing modes)

## Introduction

### The Case for a Separate Resource

[GEP-4747] demonstrates that Gateway field semantics are largely equivalent for ingress
and egress, though some fields carry different contextual meanings. This GEP does not dispute that analysis. Instead, it argues that
equivalent semantics are necessary but not sufficient -- user experience, RBAC,
and future extensibility warrant a dedicated resource.

#### 1. User Clarity

Gateway API serves multiple personas ([Chihiro], [Ian], [Ana]). When [Ian]
runs `kubectl get gateways`, a mix of ingress and egress gateways appears with
no structural distinction. Labels and GatewayClass names are conventions, not
guarantees. A dedicated `EgressGateway` resource makes the distinction
first-class:

```bash
$ kubectl get gateways
NAME              CLASS    ADDRESS          READY
ingress-gateway   nginx    203.0.113.10     True

$ kubectl get egressgateways
NAME              CLASS    ADDRESS            READY
egress-gateway    egress   10.96.100.50       True
```

#### 2. RBAC Separation

With a single Gateway resource, controlling who can create egress vs ingress
gateways requires admission webhooks or policy engines that inspect
GatewayClass references. A dedicated resource enables standard Kubernetes RBAC:

```yaml
# Allow team to create egress gateways but not ingress gateways
apiVersion: rbac.authorization.k8s.io/v1
kind: Role
rules:
- apiGroups: ["gateway.networking.k8s.io"]
  resources: ["egressgateways"]
  verbs: ["create", "get", "list", "watch"]
# No rule for "gateways" -- team cannot create ingress gateways
```

#### 3. Structural Guardrails

Certain Gateway configurations are valid for ingress but problematic for
egress. A dedicated resource can structurally prevent these:

- **Address types**: Egress gateways need ClusterIP, not LoadBalancer.
  An EgressGateway could default or restrict address types.
- **TLS mode**: Egress listeners may have different default TLS
  expectations than ingress listeners.

#### 4. Design Space

Even if no egress-specific fields are needed today, reserving a resource
allows future additions without modifying the Gateway resource:

- Default DNS resolver configuration for external FQDNs
- Egress-specific status conditions (e.g., upstream reachability)
- Source identity requirements (which ServiceAccounts can use this gateway)
- Proxy mode (explicit vs transparent -- if transparent egress is added later)

## API

### EgressGateway Resource

`EgressGateway` is derived from `Gateway`, re-using all of the same
underlying types. The exact field set, validation, and defaults will be
finalized if the community selects this approach. The key design
principle is maximal type reuse: `EgressGateway` SHOULD re-use
types that already exist in Gateway API where applicable.

### Key Differences From Gateway

| Aspect | Gateway | EgressGateway |
|--------|---------|---------------|
| Default address type | Implementation-specific (often LoadBalancer) | SHOULD default to ClusterIP |
| Typical listener count | Multiple (one per vhost) | One (wildcard) |
| Route attachment | HTTPRoute parentRef targets Gateway | HTTPRoute parentRef targets EgressGateway |
| RBAC resource | `gateways` | `egressgateways` |

### HTTPRoute Attachment

HTTPRoute already supports heterogeneous parentRef kinds. Attaching to an
EgressGateway requires only setting the `kind` field:

```yaml
apiVersion: gateway.networking.k8s.io/v1
kind: HTTPRoute
metadata:
  name: openai-route
  namespace: app-team
spec:
  parentRefs:
  - name: egress-gateway
    namespace: gateway-system
    kind: EgressGateway
    group: gateway.networking.k8s.io
  hostnames:
  - "api.openai.com"
  rules:
  - backendRefs:
    - group: gateway.networking.k8s.io
      kind: Backend
      name: openai-backend
```

### Full Example

```yaml
apiVersion: gateway.networking.k8s.io/v1
kind: GatewayClass
metadata:
  name: egress
spec:
  controllerName: example.com/egress-controller
---
apiVersion: gateway.networking.k8s.io/v1alpha1
kind: EgressGateway
metadata:
  name: egress-gw
  namespace: gateway-system
spec:
  gatewayClassName: egress
  listeners:
  - name: proxy
    port: 8443
    protocol: HTTPS
    tls:
      mode: Terminate
      certificateRefs:
      - name: egress-gw-cert
    allowedRoutes:
      namespaces:
        from: All
---
apiVersion: gateway.networking.k8s.io/v1
kind: HTTPRoute
metadata:
  name: openai
  namespace: ml-team
spec:
  parentRefs:
  - name: egress-gw
    namespace: gateway-system
    kind: EgressGateway
    group: gateway.networking.k8s.io
  hostnames:
  - "api.openai.com"
  rules:
  - backendRefs:
    - group: gateway.networking.k8s.io
      kind: Backend
      name: openai-api
---
apiVersion: gateway.networking.k8s.io/v1alpha1
kind: Backend
metadata:
  name: openai-api
  namespace: ml-team
spec:
  destination:
    type: Hostname
    hostname:
      address: api.openai.com
    ports:
    - number: 443
      protocol: HTTP2
      tls:
        mode: Simple
        validation:
          hostname: api.openai.com
```

## Trade-offs

This section honestly presents the trade-offs of this approach. See [GEP-4747]
for the alternative.

### Costs of a Separate Resource

1. **API surface area**: A new CRD, new RBAC resources, new conformance tests.
   More for implementations to support.

2. **Type duplication**: Even though `EgressGatewaySpec` reuses underlying
   types, it is a new top-level type that must be maintained.

3. **Ecosystem fragmentation**: Implementations must decide whether to support
   Gateway-for-egress, EgressGateway, or both.

4. **No existing API divergence**: Every Gateway field has identical
   or equivalent semantics for ingress and egress. The case for a
   separate resource is UX, API design space, and RBAC, not semantic
   necessity.

### Benefits of a Separate Resource

1. **Clear UX**: `kubectl get egressgateways` immediately distinguishes egress.

2. **Native RBAC**: Standard Kubernetes RBAC without admission webhooks.

3. **Structural defaults**: Sensible egress defaults (ClusterIP, wildcard
   listeners) without special-casing Gateway behavior based on GatewayClass.

4. **Design space**: Future egress-specific fields can be added without
   modifying Gateway.

## Conformance

EgressGateway uses the same `Egress` conformance profile defined in
[GEP-4747]. Whether the community chooses Gateway reuse or EgressGateway,
the conformance tests verify the same egress behavior. The only difference
is which resource types the tests target.

If EgressGateway is chosen, conformance tests MUST verify:

- EgressGateway creation and status reporting
- HTTPRoute attachment via `parentRef.kind: EgressGateway`
- Routing to Backend resources ([PR #4488](https://github.com/kubernetes-sigs/gateway-api/pull/4488))
- ClusterIP addressing by default
- All core and extended egress conformance features from [GEP-4747]

## Security Considerations

Same as [GEP-4747], plus:

- **RBAC boundary**: EgressGateway provides a natural RBAC boundary.
  Organizations can grant `egressgateways` permissions to platform teams
  without granting `gateways` permissions, reducing blast radius.

## Open Questions

### 1. Should EgressGateway Share GatewayClass?

Should `EgressGateway.spec.gatewayClassName` reference the same `GatewayClass`
resource as `Gateway`? Or should there be a separate `EgressGatewayClass`?

**Recommendation**: Reuse `GatewayClass`. The controller distinction is
already expressed via `controllerName`. Adding `EgressGatewayClass` would
be excessive fragmentation.

### 2. How Much Type Reuse?

Should `EgressGatewaySpec` literally embed `GatewaySpec` (and add/remove
fields via validation) or define its own parallel struct with shared
sub-types?

**Recommendation**: Own struct with shared sub-types (Listener, GatewaySpecAddress,
etc.). This allows egress-specific defaults and documentation without import
cycles.

## Alternatives Considered

### Reuse Existing Gateway (GEP-4747)

[GEP-4747] proposes using existing Gateway with GatewayClass to distinguish
egress. It argues no new resource is needed because field semantics are
identical.

**This GEP's response**: Identical semantics are necessary but not sufficient.
UX, RBAC, and design space justify a dedicated resource.

### EgressRoute (Prior GEP #1971)

A [previous attempt](https://github.com/kubernetes-sigs/gateway-api/pull/1971)
proposed an `EgressRoute` resource. This GEP takes a different approach: the
new resource is the gateway, not the route. Routes (HTTPRoute, GRPCRoute)
are reused unchanged.

## Dependencies

| Dependency | Status | Impact |
|---|---|---|
| [PR #4488: Backend Resource](https://github.com/kubernetes-sigs/gateway-api/pull/4488) | PR open | Required -- egress routes need Backend destinations |
| [#1651: Gateway Routability](https://github.com/kubernetes-sigs/gateway-api/issues/1651) | Issue open | Nice-to-have -- EgressGateway could default to ClusterIP addressing |
| [GEP-4747: Egress Gateway Support](../gep-4747/index.md) | Companion | Community must choose between this GEP and GEP-4747 |

## References

* [WG AI Gateway egress proposal](https://github.com/kubernetes-sigs/wg-ai-gateway/blob/main/proposals/10-egress-gateways.md)
* [EgressGateway prototype (wg-ai-gateway PR #45)](https://github.com/kubernetes-sigs/wg-ai-gateway/pull/45)
* [PR #4488: Backend Resource](https://github.com/kubernetes-sigs/gateway-api/pull/4488)
* [GEP-4747: Egress Gateway Support](../gep-4747/index.md)

## Graduation Criteria

### Provisional -> Implementable

- [ ] Community decision on EgressGateway (this GEP) vs Gateway reuse ([GEP-4747])
- [ ] [PR #4488](https://github.com/kubernetes-sigs/gateway-api/pull/4488) (Backend) reaches at least Provisional status
- [ ] Open questions resolved (GatewayClass reuse, type structure)
- [ ] EgressGateway CRD schema finalized

### Alpha (Experimental)

- [ ] EgressGateway CRD in gateway-api repository
- [ ] Conformance tests for EgressGateway
- [ ] At least one implementation

### Beta

- [ ] At least two implementations
- [ ] Production usage reports
- [ ] No major API changes for 3+ months

### GA (Standard)

- [ ] Three implementations
- [ ] Stable for 6+ months
- [ ] Security review complete

