# GEP-4661: In-Cluster: Provisioned service scope and optimizations

* Issue: [#4661](https://github.com/kubernetes-sigs/gateway-api/issues/4661)
* Status: Provisional

## TLDR (What)

This GEP enables Gateway owners to portably select the Kubernetes Service type provisioned by an in-cluster Gateway implementation, 
and establishes production-ready defaults for each service type so that common best practices are applied automatically.
In-cluster Gateways differ from traditional Gateways because they are used exclusively inside a cluster that has not been provisioned for cloud access.

Concretely, this GEP:

* Introduces new `AddressType` values that allow users to portably specify
  the scope of a service provisioned by an `In-Cluster` implementation —
  whether the provisioned Service should be a `ClusterIP` or an optimized
  `LoadBalancer` — as an explicit opt-in, without changing existing behavior.
* Defines normative requirements per address type so that implementations
  ship with consistent, well-defined defaults (e.g. `externalTrafficPolicy`).

### Goals

* Introduce a new `AddressType` value that allows Gateway owners to portably
  provision a Gateway scoped for in-cluster-only visibility (ClusterIP)
* Introduce a new `AddressType` value that allows Gateway owners to opt in to
  production-ready LoadBalancer defaults (e.g. `externalTrafficPolicy: Local`) without changing the behavior of existing Gateways
* Define normative requirements per `AddressType` so implementations ship
  consistent, well-defined behavior for each service scope

### Non-Goals

* Replicate the full Kubernetes Service API inside Gateway API
* Change existing LoadBalancer provisioning behavior — optimized defaults are
  opt-in via new address types, not retroactively applied
* Replace `infrastructure.parametersRef` for implementation-specific Service
  customization
* Support `ExternalName` Service types

## Motivation (Why)

[GEP-1762](https://gateway-api.sigs.k8s.io/geps/gep-1762/) established the foundation for in-cluster Gateway deployments and acknowledged that Service type matters — its [Gateway IP](https://gateway-api.sigs.k8s.io/geps/gep-1762/#gateway-ip) section references both `ClusterIP` and `LoadBalancer` services — but did not provide a portable mechanism to choose between them. Instead, this was deferred to "arbitrary customization" via `infrastructure.parametersRef` ([GEP-1867](https://gateway-api.sigs.k8s.io/geps/gep-1867/)).

In practice, this means that every implementation has solved service type selection differently — through custom annotations, implementation-specific parameters, or other ad-hoc mechanisms. 
This reproduces the same fragmentation that Gateway API was designed to eliminate: users must learn each implementation's particular approach for what is fundamentally a portable concern.

By promoting service type selection into the Gateway API itself, this GEP gives users a single, declarative way to express their intent. Implementations can then apply well-defined defaults for each service type, ensuring that a newly provisioned Gateway is production-ready without requiring additional configuration.

This GEP does not aim to replicate the full Kubernetes Service API. The scope is deliberately narrow: service type selection and normative defaults for the most impactful fields. Additional Service-level customization remains available through `infrastructure.parametersRef` for implementation-specific needs.

## Who

This GEP benefits Chihiro, the cluster operator as they:

* need to choose the right service type for their workload without learning implementation-specific configuration.
* want consistent, production-ready defaults across Gateway deployments in their clusters.

### Use Cases

* A Gateway owner provisions a Gateway for
  [inference extension](https://gateway-api-inference-extension.sigs.k8s.io/) and wants it reachable only within the cluster. Today, making the provisioned Service a `ClusterIP` requires implementation-specific knowledge. With this GEP, the owner can express this intent portably.
* A Gateway owner provisions a Gateway exposed via `LoadBalancer` and expects production-ready traffic routing out of the box — with `externalTrafficPolicy` set to `Local` to preserve client source IP and avoid unnecessary cross-node hops.
* A cluster operator migrating from implementation A to implementation B expects their Gateway manifests to work without modification, because the service type and scope are expressed portably rather than through implementation-specific annotations.


## API

This GEP introduces two new predefined `AddressType` values for
`spec.addresses`: `ClusterIPAddress` and `OptimizedLoadBalancerAddress`
(name subject to change). These new types allow Gateway owners to express
service scope and opt in to production-ready defaults without any change
to the existing `GatewaySpecAddress` struct — only the set of recognized
`AddressType` values is extended.

No new fields are added to `GatewaySpecAddress`. The `value` field remains
optional as defined in the current API.

When a user specifies an address entry without an explicit `type`, the CRD
defaulting sets it to `IPAddress` (via `+kubebuilder:default=IPAddress` on
`GatewaySpecAddress.Type`). This means the new address types are always
opt-in — existing Gateways and Gateways that omit the `type` field continue
to behave exactly as they do today, with the implementation deciding the
Service type and configuration.

### API Changes

The only API change is the addition of two new `AddressType` constants:

```go
// A ClusterIPAddress requests that the implementation provisions a
// ClusterIP Service for this Gateway, making it reachable only within
// the cluster. The user MUST NOT set a value for this address type;
// the ClusterIP is allocated by Kubernetes when the Service is created.
//
// When a Gateway is provisioned with a ClusterIPAddress, it is also
// reachable via the internal DNS name of the provisioned Service
// (e.g. <service-name>.<namespace>.svc.cluster.local).
//
// Support: Extended
ClusterIPAddressType AddressType = "ClusterIPAddress"

// An OptimizedLoadBalancerAddress requests that the implementation
// provisions a LoadBalancer Service with production-ready defaults.
// Implementations MUST set externalTrafficPolicy to Local when using this mode.
//
// The value field is optional. When empty, the external address is
// assigned by the load balancer provider. When set, it requests that
// specific address from the provider (subject to provider support).
//
// This type exists to allow opting in to best-practice defaults
// without changing the behavior of Gateways that do not specify
// any address or that use the existing IPAddress / Hostname types.
//
// Support: Extended
OptimizedLoadBalancerAddressType AddressType = "OptimizedLoadBalancerAddress"
```

### Normative Requirements

#### Status Addresses

The new address types (`ClusterIPAddress`, `OptimizedLoadBalancerAddress`)
express user intent in `spec.addresses` — they control how the Gateway is
provisioned. However, `status.addresses` describes the *observed state*: the
actual network endpoints where the Gateway is reachable. Echoing the
spec type into status is not useful to consumers, because it does not tell
them whether the address is an IP or a hostname.

Implementations MUST report `status.addresses` using primitive types
(`IPAddress` or `Hostname`) that describe the concrete address, not the
requested address type. The `GatewayStatus` documentation already states
that the status address list may differ from the spec, so this is consistent
with the existing API contract.

#### ClusterIPAddress

* The implementation MUST provision a `ClusterIP` Service for the Gateway.
* The `value` field MUST be empty. If a user specifies a value, the
  implementation MUST set the `Programmed` condition to `False` with reason
  `AddressNotAssigned`.
* The ClusterIP is allocated by Kubernetes when the Service is created.
  The implementation MUST report the allocated ClusterIP in
  `status.addresses` with `type: IPAddress`.
* The Gateway is also reachable via the internal DNS name of the provisioned
  Service (e.g. `<service-name>.<namespace>.svc.cluster.local`).

#### OptimizedLoadBalancerAddress

* The implementation MUST provision a `LoadBalancer` Service for the Gateway.
* The `value` field is optional. When empty, the external address is assigned
  by the load balancer provider. When set, the implementation SHOULD request
  that specific address from the load balancer provider. If the provider does
  not support static address assignment, the implementation MUST set the
  `Programmed` condition to `False` with reason `AddressNotAssigned`.
* Implementations SHOULD set `externalTrafficPolicy: Local`.
* The implementation MUST report the assigned external address in
  `status.addresses` using the appropriate primitive type (`IPAddress` or
  `Hostname`) depending on what the load balancer provider assigns.

### Precedence

`infrastructure.parametersRef` takes precedence over the service type
expressed via `spec.addresses`. This preserves backward compatibility for
existing deployments that rely on `parametersRef` to control the provisioned
Service.

### Examples

#### Current behavior: no addresses specified

When `spec.addresses` is omitted, implementations provision a Gateway using
their default behavior — typically a `LoadBalancer` Service without any
guaranteed best-practice defaults:

```yaml
apiVersion: gateway.networking.k8s.io/v1
kind: Gateway
metadata:
  name: default-gateway
spec:
  gatewayClassName: example
  listeners:
  - name: http
    port: 80
    protocol: HTTP
```

The implementation decides the Service type and configuration. The resulting
behavior is implementation-specific and may vary across providers.

#### ClusterIP: in-cluster-only Gateway

To provision a Gateway reachable only within the cluster, the owner specifies
the `ClusterIPAddress` type with no value:

```yaml
apiVersion: gateway.networking.k8s.io/v1
kind: Gateway
metadata:
  name: internal-gateway
spec:
  gatewayClassName: example
  addresses:
  - type: ClusterIPAddress
  listeners:
  - name: http
    port: 80
    protocol: HTTP
```

The implementation provisions a `ClusterIP` Service. Kubernetes allocates
the ClusterIP and the implementation reports it in status. The Gateway is
reachable both by its ClusterIP and by the internal DNS name of the
provisioned Service (e.g. `internal-gateway.default.svc.cluster.local`):

```yaml
status:
  addresses:
  - type: IPAddress
    value: "10.96.42.7"
  conditions:
  - type: Programmed
    status: "True"
```

#### OptimizedLoadBalancer: production-ready external Gateway

To provision a Gateway with a LoadBalancer Service that applies best-practice
defaults, the owner specifies the `OptimizedLoadBalancerAddress` type:

```yaml
apiVersion: gateway.networking.k8s.io/v1
kind: Gateway
metadata:
  name: production-gateway
spec:
  gatewayClassName: example
  addresses:
  - type: OptimizedLoadBalancerAddress
  listeners:
  - name: http
    port: 80
    protocol: HTTP
```

The implementation provisions a `LoadBalancer` Service with
`externalTrafficPolicy: Local`.
The external address assigned by the load balancer provider is reported in
status using the primitive type — `IPAddress` when the provider assigns an IP,
or `Hostname` when the provider assigns a DNS name (e.g. cloud load balancers):

```yaml
# Provider assigns an IP address
status:
  addresses:
  - type: IPAddress
    value: "203.0.113.10"
  conditions:
  - type: Programmed
    status: "True"
```

```yaml
# Provider assigns a hostname (e.g. AWS ELB)
status:
  addresses:
  - type: Hostname
    value: "abc123.us-east-1.elb.amazonaws.com"
  conditions:
  - type: Programmed
    status: "True"
```

When a specific external address is desired, the owner can request it via the
`value` field (subject to load balancer provider support):

```yaml
apiVersion: gateway.networking.k8s.io/v1
kind: Gateway
metadata:
  name: production-gateway-static
spec:
  gatewayClassName: example
  addresses:
  - type: OptimizedLoadBalancerAddress
    value: "203.0.113.50"
  listeners:
  - name: http
    port: 80
    protocol: HTTP
```

### Cluster Policy Enforcement with ValidatingAdmissionPolicy

Cluster administrators can use Kubernetes
[ValidatingAdmissionPolicy](https://kubernetes.io/docs/reference/access-authn-authz/validating-admission-policy/)
to enforce organizational constraints on which address types are allowed.

#### Example: Block external LoadBalancers (ClusterIP only)

This policy ensures that every Gateway in the cluster must use
`ClusterIPAddress`, preventing the creation of externally-exposed Gateways:

```yaml
apiVersion: admissionregistration.k8s.io/v1
kind: ValidatingAdmissionPolicy
metadata:
  name: "gateway-clusterip-only.example.com"
spec:
  failurePolicy: Fail
  matchConstraints:
    resourceRules:
    - apiGroups:   ["gateway.networking.k8s.io"]
      apiVersions: ["v1"]
      operations:  ["CREATE", "UPDATE"]
      resources:   ["gateways"]
  validations:
    - expression: >-
        has(object.spec.addresses) &&
        object.spec.addresses.size() > 0 &&
        object.spec.addresses.all(a, a.type == 'ClusterIPAddress')
      message: "Gateways in this cluster must use ClusterIPAddress. External LoadBalancers are not allowed."
      reason: Forbidden
---
apiVersion: admissionregistration.k8s.io/v1
kind: ValidatingAdmissionPolicyBinding
metadata:
  name: gateway-clusterip-only.example.com
spec:
  policyName: gateway-clusterip-only.example.com
  validationActions: [Deny]
  matchResources:
    resourceRules:
    - apiGroups:   ["gateway.networking.k8s.io"]
      apiVersions: ["v1"]
      resources:   ["gateways"]
      operations:  ["CREATE", "UPDATE"]
```

#### Example: Enforce OptimizedLoadBalancer for external Gateways

This policy allows both `ClusterIPAddress` and `OptimizedLoadBalancerAddress`
but blocks Gateways that omit `spec.addresses` (which would get
implementation-default LoadBalancer behavior without best-practice defaults):

```yaml
apiVersion: admissionregistration.k8s.io/v1
kind: ValidatingAdmissionPolicy
metadata:
  name: "gateway-optimized-lb.example.com"
spec:
  failurePolicy: Fail
  matchConstraints:
    resourceRules:
    - apiGroups:   ["gateway.networking.k8s.io"]
      apiVersions: ["v1"]
      operations:  ["CREATE", "UPDATE"]
      resources:   ["gateways"]
  validations:
    - expression: >-
        has(object.spec.addresses) &&
        object.spec.addresses.size() > 0 &&
        object.spec.addresses.all(a,
          a.type == 'ClusterIPAddress' ||
          a.type == 'OptimizedLoadBalancerAddress'
        )
      message: "Gateways must specify an explicit address type (ClusterIPAddress or OptimizedLoadBalancerAddress). Default LoadBalancer provisioning without optimized defaults is not allowed."
      reason: Forbidden
---
apiVersion: admissionregistration.k8s.io/v1
kind: ValidatingAdmissionPolicyBinding
metadata:
  name: gateway-optimized-lb.example.com
spec:
  policyName: gateway-optimized-lb.example.com
  validationActions: [Deny]
  matchResources:
    resourceRules:
    - apiGroups:   ["gateway.networking.k8s.io"]
      apiVersions: ["v1"]
      resources:   ["gateways"]
      operations:  ["CREATE", "UPDATE"]
```

## Open Questions

* **Mixed address types**: When a Gateway specifies multiple entries in
  `spec.addresses` that combine different types (e.g. both `ClusterIPAddress`
  and `OptimizedLoadBalancerAddress`, or a new type alongside an existing
  `IPAddress`), should the implementation create one Service per type? And
  when multiple addresses of the same type are specified, should the
  implementation aggregate them into a single Service? Optimally this would
  be implementation-specific behavior, as long as `status.addresses`
  accurately reflects the addresses that were provisioned and are reachable.

* **Mutability of address type and value**: The current API spec does not
  enforce immutability on `spec.addresses` — there are no CEL validation
  rules comparing against `oldObject`, meaning a user can freely change the
  `type` or `value` of an existing address entry after Gateway creation (see
  `Addresses` field definition in
  [`gateway_types.go`](https://github.com/kubernetes-sigs/gateway-api/blob/main/apis/v1/gateway_types.go)).
  For the new address types, this raises questions: for example, switching
  from `ClusterIPAddress` to `OptimizedLoadBalancerAddress` would require
  the implementation to deprovision one Service type and provision another.
  It needs to be defined whether this is allowed, and if so, what the
  expected implementation behavior and status transitions should be during
  the change.

* **Naming and specification of address types**: The names `ClusterIPAddress`
  and `OptimizedLoadBalancerAddress` are working names used throughout this GEP
  to convey intent. The final naming, as well as the precise specification of
  each type (support level, normative requirements), must still be discussed
  and agreed upon before this GEP moves to Implementable.

## Alternatives Considered

### ServiceRef or external API references

An earlier approach considered adding a field like `serviceRef` to
`GatewaySpec` that would point to an existing Kubernetes Service, allowing
users to bring their own pre-configured Service rather than having the
implementation provision one. This was discarded for several reasons
(discussed in detail in [SIG-Network Gateway API meeting](https://www.youtube.com/watch?v=zopZ5r_97uU&t=892s)):

* **Escape hatches become impossible to remove**: Once an escape hatch like
  `serviceRef` exists in a stable API, it becomes a permanent fixture.
  Users and implementations grow to depend on it, making it effectively
  impossible to deprecate even when better alternatives are available.

* **Loss of incentive for proper solutions**: When a catch-all escape hatch
  exists, there is reduced motivation to design focused, well-specified API
  extensions. Problems that should be solved portably within Gateway API end
  up being addressed through the escape hatch instead, leaving the portable
  API incomplete.

* **Fragmentation over unification**: A `serviceRef` would allow users to
  configure Service-level details outside of Gateway API's control, leading
  to divergent patterns across implementations. This directly undermines the
  project's goal of providing a unified, portable API — the same kind of
  fragmentation that Gateway API was created to eliminate.

* **Complex secondary API surface**: Referencing an external Service
  introduces questions around ownership, lifecycle management, conflict
  resolution, and status reporting that would significantly increase the
  API's complexity without proportional benefit.

This GEP instead takes the approach of expressing intent within the existing
`spec.addresses` mechanism via new `AddressType` values, keeping the API
surface narrow and leaving implementation-specific customization to
`infrastructure.parametersRef`.

## References

