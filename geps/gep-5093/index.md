---
title: "GEP-5093: Gateway Address Routability"
---

* Issue: [\#5093](https://github.com/kubernetes-sigs/gateway-api/issues/5093)
* Status: Implementable

(See [status definitions](../overview.md#gep-states).)

This GEP obsoletes [GEP-1651: Gateway Routability](../gep-1651/). See [Background](#background) for the relationship and prior iterations.

## TLDR

Add a `routability` field to Gateway addresses (`spec.addresses` and `status.addresses`) so that users can request, and implementations can report, the reachability scope of each address a Gateway uses.

## Motivation

Gateway API currently treats all addresses as opaque values with a `type` (`IPAddress`, `Hostname`) but no indication of where those addresses are reachable from. There is no portable way to say "give me an address that is only reachable inside the cluster" or to discover whether a provisioned address is public, cluster-internal, or somewhere in between.

This gap blocks several use cases:

* **Internal-only gateways.** Knative and similar projects need to deploy Gateways that are reachable within the cluster but not from the public internet ([#1651](https://github.com/kubernetes-sigs/gateway-api/issues/1651)). Today this requires implementation-specific annotations or out-of-band Service manipulation.

* **Egress gateways.** Workloads that route outbound traffic through a Gateway need a cluster-internal address to connect to. Without a portable way to request or identify such an address, egress patterns cannot be standardized. Standardizing such patterns has been requested by the wg-ai-gateway in service of generative AI use cases: for example, a `Gateway` with a  `Cluster` scoped address and a `Backend` pointing to an external inference provider should generally not be reachable from the open internet, to avoid injecting inference credentials into arbitrary requests. (see [#4746](https://github.com/kubernetes-sigs/gateway-api/pull/4746) discussions on "open relays".)

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
* Support prefixed custom routability values for implementation-specific scopes.

## Non-Goals

* Defining enforcement mechanisms (e.g. NetworkPolicy) for restricting traffic to or from a Gateway.
* Validating that an `External` address is reachable from a particular network. The `routability` field expresses intent and expected reachability, except that reported `IPAddress` values are checked against the ServiceCIDR for the `Cluster` and `External` scopes.
* Exposing Service-level fields (`loadBalancerClass`, `sessionAffinity`, etc.) on Gateway. Those concerns belong in a separate effort.
* Recording or enforcing Gateway intent (e.g. an ingress/egress `type` field). This GEP defines *reachability* of addresses; it takes no opinion on whether a Gateway is intended for ingress, egress, or both, or on how that intent is recorded or enforced. That is pursued in a separate GEP split from \#4746.
* Providing a way to direct traffic to a Gateway's internal addresses without reading status, e.g. via a Service or EndpointSelector. This is a natural follow-on for discovery ergonomics and is deferred to [KEP-6116](https://github.com/kubernetes/enhancements/issues/6116) rather than pursued here.

## API

### Routability Field

A new optional `routability` field is added to both `GatewaySpecAddress` and `GatewayStatusAddress`. It uses a new `GatewayAddressRoutability` string type:

```go
// GatewayAddressRoutability describes where a Gateway address is expected to
// be reachable from.
//
// Valid values are `External`, `Cluster`, or a prefixed implementation-specific
// value. The `gateway.networking.k8s.io` prefix is reserved and cannot be used
// until Gateway API defines a value for it.
//
// +kubebuilder:validation:MinLength=1
// +kubebuilder:validation:MaxLength=253
// Prefixed values require only a non-empty prefix and path; the prefix is not
// validated as a domain name.
// +kubebuilder:validation:XValidation:message="Routability must be External, Cluster, or an implementation-specific prefixed path; gateway.networking.k8s.io is reserved",rule="self == 'External' || self == 'Cluster' || (self.matches(r\"\"\"^[^/]+/.+$\"\"\") && !self.startsWith('gateway.networking.k8s.io/'))"
type GatewayAddressRoutability string

const (
	GatewayAddressRoutabilityExternal GatewayAddressRoutability = "External"
	GatewayAddressRoutabilityCluster  GatewayAddressRoutability = "Cluster"
)

type GatewaySpecAddress struct {
	// Existing fields omitted.

	// Routability specifies the requested reachability scope of this address.
	// When unset, this field requests External.
	//
	// <gateway:util:excludeFromCRD>
	// Notes for implementors:
	//
	// Implementations claiming GatewayAddressRoutability MUST use and report
	// External when this field is unset.
	// </gateway:util:excludeFromCRD>
	//
	// +optional
	// <gateway:experimental>
	Routability GatewayAddressRoutability `json:"routability,omitempty,omitzero"`
}

type GatewayStatusAddress struct {
	// Existing fields omitted.

	// Routability reports the reachability scope of this address. When unset,
	// this field is interpreted as External.
	//
	// <gateway:util:excludeFromCRD>
	// Notes for implementors:
	//
	// Implementations claiming GatewayAddressRoutability MUST set this field
	// for every status address.
	// </gateway:util:excludeFromCRD>
	//
	// +optional
	// <gateway:experimental>
	Routability *GatewayAddressRoutability `json:"routability,omitempty"`
}
```

CRD validation MUST accept the two well-known values and implementation-specific values with a non-empty prefix and path. It MUST reject other unprefixed values and values using the reserved `gateway.networking.k8s.io` prefix. Prefixed values are implementation-specific: a valid value is not necessarily supported by every implementation.

### Well-Known Values

The set of well-known values is intentionally open-ended. Because the field is a string and consumers must already tolerate values they do not recognize (including prefixed ones), new well-known scopes can be added in future revisions without breaking compatibility.

The two-value model below is a portable starting point, not a ceiling: if experience (e.g. multi-network Kubernetes, [KEP-3700](https://github.com/kubernetes/enhancements/pull/3700)) or expansion of LoadBalancer semantics ([KEP-6128](https://github.com/kubernetes/enhancements/pull/6129)) shows that additional scopes are needed, they can be introduced without disrupting existing Gateways.

* **`External`**: The address is routable from outside the cluster and consumers MUST treat it as externally reachable. For `IPAddress` addresses explicitly reported with `routability: External`, the reported provisioned address MUST NOT be in the cluster's ServiceCIDR. This is a conservative security posture, not a guarantee that a particular address is reachable from every network on the Internet. No portable validation is defined for actual external reachability. How the implementation internally provisions that address (for example, a LoadBalancer backed by a Service with a ClusterIP) is out of scope.

* **`Cluster`**: For `IPAddress` addresses, the reported address MUST be a ClusterIP in Kubernetes Service terms: it MUST be in the cluster's ServiceCIDR and MUST be reachable from within the cluster. It MAY be routable outside the cluster at the network administrator's discretion. It SHOULD use a non-globally-routable address (for example, RFC 1918 or RFC 4193) unless the cluster, including its ServiceCIDR, uses globally routable addresses.

The field also accepts prefixed values (for example, `example.com/CorpWan` or `example.com/PublicVPC`) for implementation-specific scopes or internal address ranges (RFC 1918, RFC 4193, RFC 6598). Unlike `AddressType`, the prefix is not validated as a domain name. These values have no portability guarantee and are defined by the implementation that supports them. Values using the `gateway.networking.k8s.io` prefix are invalid until Gateway API defines a corresponding well-known value.

### Spec Semantics

When `routability` is set in `spec.addresses`, it forms a **requirement** on any address the implementation provisions for that entry--the same way specifying an exact `value` does. When unset, it requests `External`. An implementation claiming `GatewayAddressRoutability` MUST honor that request and report `External`.

For an `IPAddress` result, `Cluster` MUST use a ClusterIP from the ServiceCIDR and `External` MUST NOT.

An implementation MUST NOT satisfy an entry with an address of a different reachability value and MUST treat an unrecognized routability value as unsatisfiable.

If a requested routability cannot be satisfied, the correct behavior is to leave that entry unsatisfied and report it. Implementations MUST NOT substitute a different scope.

`spec.addresses` MAY contain entries with different routability values and may combine them with requests for specific addresses. Implementations MUST evaluate each address request separately.

When `spec.addresses` is nonempty, `status.addresses` MUST contain exactly one distinct matching address for every satisfied spec entry and MUST NOT contain an address that does not match a spec entry. List order has no semantic meaning. A status address matches an entry when its effective type matches, its value matches when requested, and its routability matches. A claiming implementation MUST report `External` for an entry with unset routability. If every spec entry is satisfied, `status.addresses` contains the same number of entries as `spec.addresses`. If only some entries are satisfied, `status.addresses` contains only the successful entries.

When `spec.addresses` is empty, implementations continue to populate `status.addresses` as they do today. Implementations claiming `GatewayAddressRoutability` MUST populate routability for every status address.

#### Full and Partially Accepted Address Entry Semantics

A new `AddressesAssigned` condition is added to Gateway status to surface address assignment outcomes independently of `Programmed`:

```go
const (
	GatewayConditionAddressesAssigned GatewayConditionType = "AddressesAssigned"

	GatewayReasonAddressesAssigned          GatewayConditionReason = "Assigned"
	GatewayReasonAddressesPartiallyAssigned GatewayConditionReason = "PartiallyAssigned"
	GatewayReasonAddressesNotAssigned       GatewayConditionReason = "NotAssigned"
)
```

The condition MUST be omitted until the controller has reconciled address assignment. It is not added to the default Gateway conditions. After reconciliation, it applies both to explicit requests and to default address selection when `spec.addresses` is empty.

If ***all*** requested entries can be satisfied, or if default address selection has completed when `spec.addresses` is empty:

* MUST set `AddressesAssigned=True` with reason `Assigned`

If ***some***, but not all, entries can be satisfied, the implementation SHOULD program the Gateway using the addresses it can satisfy. In either case, it:

* MUST set `AddressesAssigned=False` with reason `PartiallyAssigned`, with a message enumerating the unsatisfied entries.
* MUST display *only* satisfied addresses in `status.addresses`.

Vendors that opt to reject partially satisfied address entries MUST follow the same semantics as the "no entries can be satisfied" behavior below.

If ***no*** entries can be satisfied, the Gateway MUST NOT be programmed. The implementation

* MUST set `Programmed=False` with reason `AddressNotAssigned`
* MUST set `AddressesAssigned=False` with reason `NotAssigned`

`Programmed` otherwise retains its existing meaning: it reports whether the proxy is actually deployed and ready. A Gateway with all addresses assigned may still have `Programmed=False` for an unrelated reason.

### Status Semantics

Each address in `status.addresses` from an implementation claiming `GatewayAddressRoutability` MUST have `routability` set. An unset `routability` in status is understood by consumers as `External`, preserving compatibility with implementations that do not claim the feature.

**Addressing Backward Compatibility**

For backward compatibility, an unset `status.addresses[].routability` is interpreted by consumers as `External`. Implementations claiming `GatewayAddressRoutability` MUST report the actual routability for every status address.

**Why Defaulting to External Makes Sense**

* A `Cluster` address surfacing as `External` overstates exposure. Mislabeling an address as `Cluster` *understates* exposure by implying that clients who reach the Gateway exist within a privileged address space.
* `External` gives consumers the stricter security posture, while a value like `Unspecified` carries the same implication but at the cost of an extra enum value.

### Address Equivalence

When a Gateway supplies multiple addresses that share the same effective attributes -- routability value, IP family (IPv4 or IPv6), and any future or implementation-specific per-address attributes -- traffic to any of those addresses SHOULD produce equivalent results. Implementations SHOULD NOT specialize listener or routing behavior within such a set. This definition is intentionally extensible: as new per-address attributes are introduced, they narrow the equivalence class rather than conflicting with it.

**Exceptions to Address Equivalence**

* Draining and rotating load balancers -- all listeners may not drain at the same rate.
* Making equivalence a MUST implies an admission policy which is out of scope for this proposal. If an operator wishes to deny clients access to a particular address on a particular listener, this should be allowed.
* Per-address or per-client authn/authz decisions remain permitted.

### Hostname Addresses

For addresses of `type: Hostname`, the `routability` value is expected to apply to any addresses the hostname resolves to. That is, a `type: Hostname` address with `routability: Cluster` carries the same reachability expectations as a `type: IPAddress` with `routability: Cluster`.

`Cluster` has portable semantics only for `IPAddress`. An implementation MAY support `Cluster` for a `Hostname`, but in that case it MUST guarantee that every address returned by ordinary in-cluster DNS resolution is in the cluster's ServiceCIDR. Otherwise, it MUST leave the request unsatisfied. This behavior is implementation-specific and is not covered by portable conformance. `NamedAddress` and implementation-specific address types are likewise implementation-specific and have no portable `Cluster` guarantee.

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

### Routability mutability

* Should `routability` be immutable once an address request is created? `spec.addresses` is an atomic list without a stable per-entry key, so a field-level immutability rule cannot cleanly distinguish an edit from an entry being removed, added, or reordered. This GEP does not yet prescribe an update model.

### Per-address attributes

* Some implementations may need per-address configuration beyond `routability` (e.g. load balancer class, traffic policy). If so, the principle should be that addresses sharing the same routability value and IP family MUST produce equivalent routing results, but MAY carry distinct implementation-specific metadata. Whether and how to expose such metadata (and how to avoid an open-ended `map[string]string`) is currently deferred to a follow-on proposal.

## Conformance Details

### Feature Names

`GatewayAddressRoutability` is an Extended feature. A GatewayClass that claims it through `status.supportedFeatures` MUST support both `External` and `Cluster` for `IPAddress` requests, report routability on every status address, and implement the assignment-condition semantics in this GEP. Prefixed scopes and non-IP `Cluster` support remain implementation-specific.

### Conformance test scenarios

Conformance tests for `GatewayAddressRoutability` will cover the following scenarios:

* API validation accepts `External`, `Cluster`, and prefixed values with a non-empty prefix and path, and rejects an unknown bare value, malformed prefixed values, an explicit empty string, and values using the reserved `gateway.networking.k8s.io` prefix. The validation is tested on both spec and status addresses.
* A `Cluster` `IPAddress` request is reported with `routability: Cluster`, is in the configured or discovered ServiceCIDR, and is reachable from an in-cluster client.
* An `External` `IPAddress` request is reported with `routability: External` and is outside the ServiceCIDR. Tests do not validate its reachability from external networks.
* A Gateway with one `External` and one `Cluster` request reports exactly one matching status address for each, regardless of list order, and `AddressesAssigned=True` with reason `Assigned`.
* A `Cluster` `IPAddress` request with a static value outside the ServiceCIDR is unsatisfied. Combined with a satisfiable `External` request, it reports only the successful address and `AddressesAssigned=False` with reason `PartiallyAssigned`; on its own, it reports `Programmed=False` with reason `AddressNotAssigned` and `AddressesAssigned=False` with reason `NotAssigned`. This scenario requires `SupportGatewayStaticAddresses`.
* A Gateway with `spec.addresses` unset has `AddressesAssigned=True` with reason `Assigned` after default address selection. A claiming implementation reports `routability: External` for every resulting status address.
* A claiming implementation with `spec.addresses[].routability` unset (not explicitly set to an empty string) reports `routability: External` for that satisfied status address.

The suite may discover ServiceCIDRs from the cluster or receive them through conformance configuration. Tests for valid but unsupported prefixed values require configuration identifying a prefix that the implementation does not support and are not mandatory portable conformance tests. `Hostname`, `NamedAddress`, and implementation-specific address types are excluded from portable conformance for this feature.

## Alternatives Considered

* **Gateway-level `spec.infrastructure.routability` (GEP-1651).** The predecessor design placed a single routability field on the Gateway rather than on each address. GEP-1651's own "Alternatives" section anticipated per-address routability but recorded concerns about "complicating the Gateway's purpose" by allowing multiple scopes. This GEP takes the position that multi-address, mixed-scope Gateways are a real and important use case (notably for egress, where a Gateway may need both a cluster-internal and a broader reachability scope), and that the per-address model is the cleaner way to express it. See [Background](#background) for the full rationale.

* **Adding `ClusterIP` as a new `AddressType`.** This conflates the reachability scope with the address format. A ClusterIP is an `IPAddress` with cluster scope, not a different type of address. Keeping `type` and `routability` orthogonal is cleaner and more extensible.

* **An explicit `Unspecified` status value.** Considered as a way to avoid ascribing `External` to legacy Gateways that never set the field. Rejected: it would permanently enshrine a transition-period edge case, as the vast majority of addresses would likely carry `Unspecified` long after implementations adopt the field. Meanwhile, interpreting an omitted value as `External` gives consumers the same conservative security posture while encouraging adoption of correctly labeled routability.

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
