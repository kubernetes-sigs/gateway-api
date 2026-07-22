# GEP-5093: Gateway Address Routability

* Issue: [\#5093](https://github.com/kubernetes-sigs/gateway-api/issues/5093)
* Status: Provisional

(See [status definitions](../overview.md#gep-states).)

This GEP obsoletes [GEP-1651: Gateway Routability](https://gateway-api.sigs.k8s.io/geps/gep-1651/). See [Background](#background) for the relationship and prior iterations.

## TLDR

Add a `routability` field to Gateway addresses (`spec.addresses` and `status.addresses`) so that users can request, and implementations can report, the reachability scope of each address a Gateway uses.

## Motivation

Gateway API currently treats all addresses as opaque values with a `type` (`IPAddress`, `Hostname`) but no indication of where those addresses are reachable from. There is no portable way to say "give me an address that is only reachable inside the cluster" or to discover whether a provisioned address is public, cluster-internal, or somewhere in between.

This gap blocks several use cases:

* **Internal-only gateways.** Knative and similar projects need to deploy Gateways that are reachable within the cluster but not from the public internet (\#1651). Today this requires implementation-specific annotations or out-of-band Service manipulation.

* **Egress gateways.** Workloads that route outbound traffic through a Gateway need a cluster-internal address to connect to. Without a portable way to request or identify such an address, egress patterns cannot be standardized. Standardizing such patterns has been requested by the wg-ai-gateway in service of generative AI use cases: for example, a `Gateway` with a  `Cluster` scoped address and a `Backend` pointing to an external inference provider should generally not be reachable from the open internet, to avoid injecting inference credentials into arbitrary requests. (see [\#4746](https://github.com/kubernetes-sigs/gateway-api/pull/4746) discussions on "open relays".)

* **Multi-address gateways.** A Gateway may be provisioned with both a public and an internal address. Clients currently have no way to determine which address is appropriate for their context.

## Background

This GEP has two lines of origin.

### GEP-1651: Gateway-level routability

This GEP obsoletes GEP-1651, which proposed a `routability` field under `spec.infrastructure.routability` on the Gateway. That design treated routability as a Gateway-level concern: one scope per Gateway, drawn from a `Public` / `Private` / `Cluster` enum, with status addresses required to be semantically no wider than the requested scope.

GEP-1651 correctly identified the problem and `Cluster` is preserved here. Its "Alternatives" section even anticipated per-address routability and recorded the trade-offs. We are indebted to dprotaso and sunjayBhatia for the foundational work.

GEP-1651 did not progress past Provisional. The blocking concerns were:

* **`Private` needs decomposition.** It was defined as "routable inside a private network larger than a single cluster (e.g. VPC) and MAY include RFC1918 address space." Reviewer comments reveal a need to decompose `Private` into two distinct scopes--a VPC-internal address and a cluster-internal address--with different operational implications (whether kube-proxy captures traffic, whether the address is reachable from adjacent clusters). Without that decomposition, a precise definition could not be agreed.
* **Multi-network Kubernetes.** Concurrent work on multi-network support upstream ([KEP-3700](https://github.com/kubernetes/enhancements/pull/3700)) made any binary classification of "private" feel premature, because the boundary of "the private network" is itself becoming an administrator-defined concept.
* **Single scope per Gateway is too coarse.** Real deployments provision Gateways with both a public and a cluster-internal address (e.g. an ingress Gateway also reachable from inside the cluster for mesh-internal callers). A Gateway-level field cannot express this.
* **No portable per-address discovery.** Consumers (workloads, agents, egress tooling) had no reliable way to read the reachability of an *individual* address from `status.addresses`, because the scope lived on the Gateway rather than on each address.

### GEP-4747: L7 reverse-proxy egress

The second line of origin is GEP-4747 ([PR \#4746](https://github.com/kubernetes-sigs/gateway-api/pull/4746)), which proposed L7 reverse-proxy egress using the existing `Gateway` resource. During review, a portable way to request a cluster-internal address was listed by Gateway API maintainers as the GEP's "biggest requirement"--agents need an IP to connect to and must be able to *programmatically* determine its reachability from the Gateway's status ([\#4746 comment](https://github.com/kubernetes-sigs/gateway-api/pull/4746#issuecomment-4349943298)).

An initial attempt to absorb that requirement into GEP-4747 (by adding `ClusterIP` as an address type and a Gateway `type` field) caused the GEP to accumulate unrelated debates--TLS policy scoping, open-relay prevention, ingress/egress intent encoding--that were valuable in their own right but not load-bearing for the core egress model. At [howardjohn's](https://github.com/howardjohn) suggestion ([review](https://github.com/kubernetes-sigs/gateway-api/pull/4746#pullrequestreview-4573981917)), the GEP was closed and split into smaller proposals, of which this is the first ([closing comment](https://github.com/kubernetes-sigs/gateway-api/pull/4746#issuecomment-4367626623)).

The split was deliberate: reachability is a prerequisite for egress, but its scope is broader than egress alone (it also covers internal-only ingress, multi-address Gateways, and discovery by arbitrary consumers). Pursuing it here lets GEP-4747 and any companion proposals reference a settled reachability model rather than re-litigating it.

### User Stories

* As a platform operator, I want to request a cluster-internal address for a Gateway so that it is not exposed to the public internet.
* As a workload developer, I want to discover from a Gateway's status whether its address is reachable from inside the cluster, from the public internet or within an implementation-defined scope (e.g., a VPC).
* As an implementation author, I want to express implementation-specific routability scopes without waiting for upstream API changes.
* As a workload developer, I want to direct traffic at a suitable (ideally cluster-internal) Gateway address without consuming the status subobject, e.g. by targeting a Service or an EndpointSelector. (Out of scope for this GEP; see Non-Goals and [KEP-6116](https://github.com/kubernetes/enhancements/issues/6116).)

## Goals

* Define a `routability` field on `GatewaySpecAddress` and `GatewayStatusAddress` with well-known values that cover the most common scopes.
* Allow implementations to report routability in `status` even when the user did not request a specific scope in `spec`.
* Support domain-prefixed custom routability values for implementation-specific scopes.

## Non-Goals

* Defining enforcement mechanisms (e.g. NetworkPolicy) for restricting traffic to or from a Gateway.
* Validating that an address actually falls within a particular IP range. The `routability` field expresses intent and expected reachability, not a CIDR validation rule.
* Exposing Service-level fields (`loadBalancerClass`, `sessionAffinity`, etc.) on Gateway. Those concerns belong in a separate effort.
* Recording or enforcing Gateway intent (e.g. an ingress/egress `type` field). This GEP defines *reachability* of addresses; it takes no opinion on whether a Gateway is intended for ingress, egress, or both, or on how that intent is recorded or enforced. That is pursued in a separate GEP split from \#4746.
* Providing a way to direct traffic to a Gateway's internal addresses without reading status, e.g. via a Service or EndpointSelector. This is a natural follow-on for discovery ergonomics and is deferred to [KEP-6116](https://github.com/kubernetes/enhancements/issues/6116) rather than pursued here.

## API

**TODO**: Concrete type definitions will be added once there is consensus on the motivation, values, and field placement described below. This includes the `routability` field on `GatewaySpecAddress` and `GatewayStatusAddress`, as well as a new `GatewayConditionReason` for `AddressesPartiallyAssigned`.

### Routability Field

A new optional `routability` field is added to both `GatewaySpecAddress` and `GatewayStatusAddress`. The field is a string with two well-known values and support for domain-prefixed extensibility.

### Well-Known Values

The set of well-known values is intentionally open-ended. Because the field is a string and consumers must already tolerate values they do not recognize (including domain-prefixed ones), importantly, new well-known scopes can be added in future revisions without breaking compatibility.

The two-value model below is a portable starting point, not a ceiling: if experience (e.g. multi-network Kubernetes, [KEP-3700](https://github.com/kubernetes/enhancements/pull/3700)) or expansion of LoadBalancer semantics ([KEP-6128](https://github.com/kubernetes/enhancements/pull/6129)) shows that additional scopes are needed, they can be introduced without disrupting existing Gateways.

* **`External`**: The address is considered routable from outside of the Cluster and must be treated as accessible from anywhere in the world from a routing perspective. The reported address MUST NOT be an address from the cluster's defined service networking range. How the implementation internally provisions that address (e.g. a LoadBalancer backed by a Service with a ClusterIP) is out of scope; the constraint applies to the value surfaced in `status.addresses`. Because the security posture toward this assumption is correct regardless of the scope, this is the default when routability is unspecified.

* **`Cluster`**: The address is routable inside the cluster the Gateway is provisioned in at a minimum. It MUST use an address from the cluster's defined service networking range (a ClusterIP in Service terms). The address MAY be routable outside the cluster at the network administrator's discretion. It SHOULD use a non-globally-routable address (e.g. RFC 1918, RFC 4193) unless the cluster is provisioned with globally routable addresses as a whole, including the service network.

The field also accepts domain-prefixed values (e.g. `example.com/CorpWan, example.com/PublicVPC`) for implementation-specific scopes or internal address ranges (RFC 1918, RFC 4193, RFC 6598), following the same extensibility pattern used by `AddressType`. Domain-prefixed values are vendor-specific and carry no portability guarantee; their behavior is defined entirely by the implementation that supports them.

### Spec Semantics

When `routability` is set in `spec.addresses`, it forms a **requirement** on any address the implementation provisions for that entry--the same way specifying an exact `value` does.

`Cluster` MUST use an address from the cluster's service networking range; `External` MUST NOT.

An implementation MUST NOT satisfy an entry with an address of a different reachability value and MUST treat an unrecognized routability value as unsatisfiable.

If a requested routability cannot be satisfied, the correct behavior is to leave that entry unsatisfied and report it. Implementations MUST NOT substitute a different scope.

`spec.addresses` MAY contain references to different routability types, with a different type on each requested address (and this MAY be combined with requesting specific addresses). In this case, implementations MUST evaluate each address request separately according to the rules above, and MUST populate `status.addresses` (including `routability`) for each configured address.
**Full and Partially Accepted Address Entry Semantics**

If ***all*** spec address entries can be satisfied, the implementation programs the Gateway normally.

If ***some***, but not all, entries can be satisfied, the implementation SHOULD program the Gateway using the addresses it can satisfy. If the Gateway is partially programmed:

* MUST set `Programmed=True`
* MUST surface the partial failure(s) with reason `AddressesPartiallyAssigned`, with a message enumerating the unsatisfied entries.
* MUST display *only* satisfied addresses in `status.addresses`.

Vendors that opt to reject partially satisfied address entries MUST follow the same semantics as the "no entries can be satisfied" behavior below.

If ***no*** entries can be satisfied, the Gateway MUST NOT be programmed. The implementation

* MUST set `Programmed=False`
* MUST use the existing reason, `AddressNotAssigned`

When `routability` is omitted from a spec address entry, the implementation MAY provision whatever routability it supports.

### Status Semantics

Each address in `status.addresses` SHOULD have `routability` set. An unset `routability` in status is understood as `External`.

**Addressing Backward Compatibility**

Some existing implementations may currently assign non-globally-routable addresses (e.g., RFC-1918 address space). Via the status semantics above, by default, those will conservatively read as `External`. This scope carries the assumption of global reachability by default. Implementations reserve the ability to assign a more specific scope e.g. example.com/Internal in order to signal that this is not the case.

**Why Defaulting to External Makes Sense**

* A `Cluster` address surfacing as `External` overstates exposure. Mislabeling an address as `Cluster` *understates* exposure by implying that clients who reach the Gateway exist within a privileged address space.
* `External` implies a stricter security posture, while a value like `Unspecified` carries the same implication but at the cost of an extra enum value.

### Address Equivalence

When a Gateway supplies multiple addresses with the same routability value and the same IP family (IPv4 or IPv6), traffic to any of those addresses SHOULD produce equivalent results. Implementations SHOULD NOT specialize listener or routing behavior within such a set.

**Exceptions to Address Equivalence**

* Draining and rotating load balancers -- all listeners may not drain at the same rate.
* Making equivalence a MUST implies an admission policy which is out of scope for this proposal. If an operator wishes to deny clients access to a particular address on a particular listener, this should be allowed.
* Per-address or per-client authn/authz decisions remain permitted.

**Routing to Equivalent Addresses**

Implementations SHOULD support round-robin as a strategy for distributing traffic equally between listeners.

Round-robin among such addresses is a viable strategy; this establishes a floor of interchangeability, not a recommended algorithm.

### Hostname Addresses

For addresses of `type: Hostname`, the `routability` value is expected to apply to any addresses the hostname resolves to. That is, a `type: Hostname` address with `routability: Cluster` carries the same reachability expectations as a `type: IPAddress` with `routability: Cluster`. Operators are responsible for ensuring that the hostname's resolution is consistent with the declared scope.

### Examples

Request a cluster-internal-only Gateway:

```yaml
apiVersion: gateway.networking.k8s.io/v1
kind: Gateway
metadata:
  name: internal-gw
spec:
  gatewayClassName: example
  addresses:
    - type: IPAddress
      routability: Cluster
  listeners:
    - name: http
      port: 80
      protocol: HTTP
```

Status showing a multi-address Gateway:

```yaml
status:
  addresses:
    - type: IPAddress
      value: "203.0.113.10"
      routability: External
    - type: IPAddress
      value: "10.96.0.42"
      routability: Cluster
```

## Open Questions

* Scopes between `Cluster` and `External` (e.g. VPC-internal, corporate WAN) are left to domain-prefixed values. Should the spec recommend a common domain prefix (e.g. `gateway.networking.k8s.io/VPC`, `gateway.networking.k8s.io/Internal`) for widely-used intermediate scopes, so that implementations converge on shared names rather than each inventing their own? This would provide a middle ground between the two well-known values and fully vendor-specific prefixes.

* The `External` definition requires that the reported address MUST NOT be from the cluster's service networking range, while explicitly scoping out how the implementation internally provisions that address (e.g. a LoadBalancer backed by a Service with a ClusterIP). Is this distinction clear enough, or should the spec say more about the boundary between the reported address and the implementation's internal plumbing?

## Conformance

**TODO**: Conformance details will be developed as the proposal matures beyond Provisional. At a minimum, the following should be testable:

* An implementation that cannot satisfy **any** requested routability sets `Programmed=False` with reason `AddressNotAssigned`.
* An implementation that satisfies **some but not all** requested routabilities programs the Gateway with the satisfiable addresses and surfaces the partial failure in status (reason `AddressesPartiallyAssigned`).
* An implementation never provisions an address whose routability differs from the one requested for that entry.

Support is expected to be an Extended conformance feature.

## Alternatives Considered

* **Gateway-level `spec.infrastructure.routability` (GEP-1651).** The predecessor design placed a single routability field on the Gateway rather than on each address. GEP-1651's own "Alternatives" section anticipated per-address routability but recorded concerns about "complicating the Gateway's purpose" by allowing multiple scopes. This GEP takes the position that multi-address, mixed-scope Gateways are a real and important use case (notably for egress, where a Gateway may need both a cluster-internal and a broader reachability scope), and that the per-address model is the cleaner way to express it. See [Background](#background) for the full rationale.

* **Adding `ClusterIP` as a new `AddressType`.** This conflates the reachability scope with the address format. A ClusterIP is an `IPAddress` with cluster scope, not a different type of address. Keeping `type` and `routability` orthogonal is cleaner and more extensible.

* **An explicit `Unspecified` status value.** Considered as a way to avoid ascribing `External` to legacy Gateways that never set the field. Rejected: it would permanently enshrine a transition-period edge case, as the vast majority of addresses would likely carry `Unspecified` long after implementations adopt the field. Meanwhile, defaulting to `External` carries equivalent security implications while encouraging adoption of correctly labeled routability.

## References

* [GEP-1651: Gateway Routability (obsoleted by this GEP)](https://gateway-api.sigs.k8s.io/geps/gep-1651/)
* [Issue \#1651: GEP: Gateway Routability](https://github.com/kubernetes-sigs/gateway-api/issues/1651)
* [PR \#4746: GEP-4747 L7 Reverse-Proxy Egress Gateway Support (closed)](https://github.com/kubernetes-sigs/gateway-api/pull/4746)
* [KEP-6116: Gateway API Service Mesh](https://github.com/kubernetes/enhancements/issues/6116)
* [KEP-6128: (alpha) LoadBalancer resource for explicitly managing and monitoring load balancers](https://github.com/kubernetes/enhancements/pull/6129)
* [wg-ai-gateway egress proposal](https://github.com/kubernetes-sigs/wg-ai-gateway/blob/main/proposals/10-egress-gateways.md)
* [KEP-3700: Multi-Network Kubernetes](https://github.com/kubernetes/enhancements/pull/3700)
* [RFC 1918: Address Allocation for Private Internets](https://tools.ietf.org/html/rfc1918)
* [RFC 4193: Unique Local IPv6 Unicast Addresses](https://tools.ietf.org/html/rfc4193)
* [RFC 6598: IANA-Reserved IPv4 Prefix for Shared Address Space](https://tools.ietf.org/html/rfc6598)

