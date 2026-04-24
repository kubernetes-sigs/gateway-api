# GEP-4751: Egress Gateway Chaining (Parent Mode)

* Issue: [#4751](https://github.com/kubernetes-sigs/gateway-api/issues/4751)
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

> **Note**: This GEP is a companion to [GEP-4747], which defines how
> Gateway API resources compose for L7 reverse-proxy egress. This GEP
> extends that work by defining how egress gateways can be chained.

[GEP-4747]: https://github.com/kubernetes-sigs/gateway-api/issues/4747
[GEP-4748]: https://github.com/kubernetes-sigs/gateway-api/issues/4748

## TLDR

Define how egress gateways can be chained -- routing traffic through a local
egress gateway to an upstream gateway -- for multi-cluster, multi-zone, and
compliance use cases.

## User Stories

### Multi-Cluster Operator

> **[Ian] operates multiple clusters and needs egress traffic from workloads in
> regional clusters to route through a central egress cluster, ensuring
> consistent policy enforcement and a single exit point for external traffic.**

### Compliance Officer

> **[Chihiro] needs all outbound traffic to exit through an audited chokepoint,
> regardless of which cluster or zone the workload runs in, to satisfy
> regulatory requirements.**

### Multi-Zone Platform Operator

> **[Ian] runs workloads across availability zones and needs regional egress
> gateways to route through a global exit point, reducing the number of
> external firewall rules and IP allowlists that must be maintained.**

## Goals

* Define how egress gateways chain through an upstream gateway (parent mode)
* Specify retry and loop-prevention requirements for chained gateways
* Determine whether gateway chaining warrants conformance tests

## Non-Goals

* Define the core egress gateway model (see [GEP-4747])
* Define the Backend resource (see [PR #4488](https://github.com/kubernetes-sigs/gateway-api/pull/4488))
* Address forward-proxy egress (see [#4704](https://github.com/kubernetes-sigs/gateway-api/issues/4704))

## Gateway Chaining

Traffic flows through a local egress gateway to an upstream gateway:

```
Workload --> Local Egress GW --> Upstream GW --> External API
```

Use cases:

- Multi-cluster: local cluster routes through a central egress cluster
- Multi-zone: regional gateways route through a global exit point
- Compliance: all traffic exits through an audited chokepoint

Requirements:

- Local retries MUST be limited to establishing the upstream connection
- Request-level retries MUST be performed by the upstream gateway
- Implementations MUST tag forwarded requests (e.g., via header) to prevent
  retry loops between chained gateways

## Open Questions

> These are explicitly marked as open for community feedback during
> Provisional status.

### 1. Conformance

Is gateway chaining common enough to warrant conformance tests, or should it
remain Implementation-Specific?

### 2. Standardized Tagging

Should the mechanism for tagging forwarded requests (to prevent retry loops)
be standardized (e.g., a well-known header), or left to implementations?

### 3. Depth Limits

Should the specification limit chaining depth (e.g., at most two hops), or
allow arbitrary depth with loop detection?

## Dependencies

| Dependency | Status | Impact |
|---|---|---|
| [GEP-4747: L7 Reverse-Proxy Egress Gateway][GEP-4747] | PR open | Required -- defines the base egress model this GEP extends |
| [PR #4488: Backend Resource](https://github.com/kubernetes-sigs/gateway-api/pull/4488) | PR open | Required -- egress routes need Backend destinations |

## References

* [GEP-4747: L7 Reverse-Proxy Egress Gateway][GEP-4747]
* [GEP-4748: EgressGateway Resource][GEP-4748]
* [WG AI Gateway egress proposal](https://github.com/kubernetes-sigs/wg-ai-gateway/blob/main/proposals/10-egress-gateways.md)

## Graduation Criteria

### Provisional -> Implementable

- [ ] [GEP-4747] reaches at least Implementable status
- [ ] Open questions resolved
- [ ] Retry and loop-prevention requirements validated by at least one implementation

### Alpha (Experimental)

- [ ] At least one implementation supports gateway chaining
- [ ] Conformance tests implemented (if warranted)

### Beta

- [ ] At least two implementations support gateway chaining
- [ ] Production usage reports

### GA (Standard)

- [ ] Three implementations passing conformance
- [ ] No API-level changes needed for 6+ months
