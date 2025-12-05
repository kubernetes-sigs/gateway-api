# GEP-1651: Gateway Routability

* Issue: [#1651](https://github.com/kubernetes-sigs/gateway-api/issues/1651)
* Status: Provisional

(See [status definitions](../overview.md#gep-states).)

## TLDR

Allow users to configure a Gateway so that it is only routable within
a specific scope (ie. public/private/cluster)

## Goals

- Define a mechanic to set the routability on a Gateway
- Provide a default set of routability options
- Provide a way for vendors to support custom options

## Non-Goals

- Per-request/route scope
- Not a lightweight service mesh

## Introduction

One of the early feature requests for Knative was the ability to deploy an
application using Knative's HTTP routing support, but make it only available
within the cluster. I want to be able to specify both the "cluster"
(service.namespace.svc) and "external" (service.namespace.example.com).
Gateways using the same GatewayClass on the cluster, but ensure that the
"cluster" service is only routable within the cluster. This would greatly
simplify deployment for users over the instructions we have today.

Likewise another use case is to provide load balancing capabilities within a virtual
private network. Different IaaS providers offer private load balancers to support
these use cases.

## API

We propose adding a new `routability` field under the `spec.infrastructure` stanza of a Gateway.

### Predefined Routability Values

Implementations MAY implement the following values for 'routability' and MUST abide by
their defined semantics.

Value | Scope
-|-
`Public`|The address is routable on the public internet
`Private`|The address is routable inside a private network larger than a single cluster (ie. VPC) and MAY include RFC1918 address space
`Cluster`|The address is routable inside the [cluster's network](https://kubernetes.io/docs/concepts/cluster-administration/networking/#how-to-implement-the-kubernetes-network-model)

Values can be compared semantically - `Public` has a larger scope than `Private`, while `Private` has a larger scope than `Cluster`.

### Vendor prefixed values

Implementations can define custom 'routability' values by specifying a vendor prefix followed
by a slash `/` and a custom name ie. `com.example.com/my-routability`.

Comparing vendor prefixed scopes with the pre-defined ones in implementation specific.

### Default Routability

The default value of `routability` is implementation specific. It is RECOMMENDED that
the default `routability` remains consistent for Gateways with the same
`gatewayClassName`.

Implementations MUST signal the default routability using the Gateway's `status.addresses`. See 'Status Addresses`
for more details.

### Mutability

Implementations MAY prevent end-users from updating the `routability` value of a Gateway. If
updates are allowed the semantics and behaviour will depend on the underlying implementation.

If a Gateway is mutated but does not support the desired routability it MUST set the conditions
`Accepted`, `Programmed` to `False` with `Reason` set to `UnsupportedRoutability`. Implementations
MAY choose to leave the old Gateway running with the previous generation's configuration.

### Go

```go

// GatewayRoutability represents the routability of a Gateway
//
// The pre-defined values listed in this package can be compared semantically.
// `Public` has a larger scope than `Private`, while `Private` has a larger scope than
// `Cluster`.
//
// Implementations can define custom routability values by specifying a vendor
// prefix followed by a slash '/' and a custom name ie. `dev.example.com/my-routability`.
//
// +kubebuilder:validation:MinLength=1
// +kubebuilder:validation:MaxLength=253
// +kubebuilder:validation:Pattern=`^Public|Private|Cluster|[a-z0-9]([-a-z0-9]*[a-z0-9])?(\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*\/[A-Za-z0-9\/\-_]+$`
type GatewayRoutability string

const (
  // GatewayRoutabilityPublic means the Gateway's address MUST
  // be routable on the public internet
  //
  // Implementations MAY support this routability
  GatewayRoutabilityPublic  GatewayRoutability  = "Public"

  // GatewayRoutabilityPrivate means the Gateway's address MUST
  // only be routable inside a private network larger than a single
  // cluster (ie. VPC) and MAY include the RFC1918 address space
  //
  // Implementations MAY support this routability
  GatewayRoutabilityPrivate GatewayRoutability  = "Private"

  // GatewayRoutabilityCluster means the Gateway's address MUST
  // only be routable inside the [cluster's network]
  //
  // Implementations MAY support this routability
  //
  // [cluster's network](https://kubernetes.io/docs/concepts/cluster-administration/networking/#how-to-implement-the-kubernetes-network-model)
  GatewayRoutabilityCluster GatewayRoutability  = "Cluster"
)

type GatewaySpec struct {
  // Infrastructure defines infrastructure level attributes about this Gateway instance.
  Infrastructure GatewayInfrastructure `json:"infrastructure"`
  // ...
}
type GatewayInfrastructure struct {
  // Routability allows the Gateway to specify the accessibility of its addresses. Setting
  // this property will override the default value defined by the GatewayClass.
  //
  // If the desired Gateway routability is incompatible with the GatewayClass implementations
  // MUST set the condition `Accepted` to `False` with `Reason` set to `UnsupportedRoutability`.

  // The default value of routability is implementation specific and  MUST remains consistent for
  // Gateways with the same gatewayClassName
  //
  // Implementations MAY prevent end-users from updating the routability value of a Gateway.
  // If updates are allowed the semantics and behaviour will depend on the underlying implementation.
  // If a Gateway is mutated but does not support the desired routability it MUST set `Accepted` 
  // and  `Programmed` conditions to `False` with `Reason` set to `UnsupportedRoutability`.
  //
  // It is RECOMMENDED that in-cluster gateways SHOULD NOT support 'Private' routability.
  // Kubernetes doesn't have a concept of 'Private' routability for Services. In the future this may
  // change upstream.
  //
  // +optional
  Routability *GatewayRoutability `json:"routability,omitempty"`
}

type GatewayStatus struct {
  // Addresses lists the IP addresses that have actually been
  // bound to the Gateway. These addresses may differ from the
  // addresses in the Spec, e.g. if the Gateway automatically
  // assigns an address from a reserved pool.
  //
  // Implementations that support Gateway routability MUST include an address
  // that has the same routable semantics as defined in the Gateway spec.
  //
  // Implementations MAY add additional addresses in status, but they MUST be
  // semantically less than the scope of the requested scope. For example if a
  // user requests a `Private` routable Gateway then an additional address MAY
  // have a routability of `Cluster` but MUST NOT include `Public`.
  //
  // +optional
  // +kubebuilder:validation:MaxItems=16
  Addresses []GatewayStatusAddress `json:"addresses,omitempty"`
  // ...
}

type GatewayStatusAddress struct {
  // Routability specifies the routable bounds of this address
  // Predefined values are: 'Private', 'Public', Cluster
  // Other values MUST have a vendor prefix.
  //
  // Implementations that support Routability MUST populate this
  // field
  //
  // +optional
  Routability *GatewayRoutability `json:"routability,omitempty"`

  // ...
}

type GatewayClassStatus struct {
  // Routabilities specifies a list of supported routabilities offered by
  // the GatewayClass. The first entry in this list will be the default
  // routability used when Gateways of this class are created.
  //
  // Implementations MAY provide a pre-defined set of GatewayClasses that
  // limit the routability choices of a Gateway.
  //
  // Implementations that support routability MUST populate this list with
  // a subset of the pre-defined GatewayRoutability values or vendored
  // prefix values.
  //
  // +optional
  // +kubebuilder:validation:MaxItems=8
  // <gateway:experimental>
  Routabilities []GatewayRoutability `json:"routabilities"`
}
```

### YAML
```yaml
apiVersion: gateway.networking.k8s.io/v1beta1
kind: Gateway
metadata:
  name: prod-web
spec:
  gatewayClassName: example
  infrastructure:
    routability: Public
  listeners:
  - protocol: HTTP
    port: 80
```

### Semantics

#### Interaction with GatewayClass

An infrastructure provider MAY provide a pre-defined set of GatewayClasses that limit the
routability choices of a Gateway. If the desired Gateway routability is incompatible with the
GatewayClass it MUST set the condition `Accepted` to `False` with `Reason` set to `UnsupportedRoutability`.

If an implementation supports 'routability' then the GatewayClass MUST list the supported
routabilities in the status stanza. The `status.routabilities` MUST contain either
a subset of the pre-defined values mentioned above or contain vendored prefixed values.

The first value in the list will be used as the default value when Gateways of this class
are created. This can be overridden by setting the Gateway's `spec.infrastructure.routability`.

#### Unsupported routability & address values

If a Gateway is unable to provide an address for the desired routability it MUST set the condition `Accepted`
to `False` with `Reason` set to `UnsupportedRoutability`

#### Status.Addresses

If a Gateway supports the desired 'routability' implementations MUST populate the `status.addresses` with
an address that has the same routable semantics. The GatewayAddress field `routability` MUST be populated.

Implementations MAY add additional addresses in status, but they MUST be semantically less than the scope
of the requested scope. For example if a user requests a `Cluster` routable Gateway then the list of addresses
MUST NOT have a routability of `Public` or `Private`.

We plan on introducing a new type `GatewayStatusAddress` and change Gateway's `status.addresses` to be
`[]GatewayStatusAddress`. This will allow the status address type to evolve separately from the spec address.

#### In-cluster Gateways and 'Private' Routability

It is RECOMMENDED that in-cluster gateways SHOULD NOT support 'Private' routability. Kubernetes doesn't have
a concept of 'Private' routability for Services. In the future this may change upstream.

#### Interaction with Multi-Network Kubernetes

[Multi-Network Kubernetes](https://github.com/kubernetes/enhancements/pull/3700) 
is a sibling SIG working on adding multi-network support to Pods. After reaching out and having a discussion with about this GEP
the consensus is that a Gateway most likely in the future can be tied to a single PodNetwork. Defining this is out of scope for this GEP.

A second consensus is the Routabilities defined in this GEP don't impact PodNetworks but instead are indicators to LB implementations
on how they should behave.

## Examples

#### 1. Request a GatewayAddress that is routable within the same cluster

```yaml
apiVersion: gateway.networking.k8s.io/v1beta1
kind: Gateway
metadata:
  name: prod-web
spec:
  gatewayClassName: example
  infrastructure:
    routability: Cluster
  listeners:
  - protocol: HTTP
    port: 80
```

#### 2. Request a GatewayAddress with a specific routability and address
```yaml
apiVersion: gateway.networking.k8s.io/v1beta1
kind: Gateway
metadata:
  name: prod-web
spec:
  gatewayClassName: example
  infrastructure:
    routability: Cluster
  listeners:
  - protocol: HTTP
    port: 80
  addresses:
  - value: 10.0.0.8
```

#### 3. Request a GatewayAddress that is routable on the public internet
```yaml
apiVersion: gateway.networking.k8s.io/v1beta1
kind: Gateway
metadata:
  name: prod-web
spec:
  gatewayClassName: example
  infrastructure:
    routability: Public
  listeners:
  - protocol: HTTP
    port: 80
```

#### 4. Request a GatewayAddress that is a cloud provider's VPC
```yaml
apiVersion: gateway.networking.k8s.io/v1beta1
kind: Gateway
metadata:
  name: prod-web
spec:
  gatewayClassName: example
  infrastructure:
    routability: Private
  listeners:
  - protocol: HTTP
    port: 80
```

## Alternatives

### Introducing new GatewayAddress Types

We could introduce additional `AddressTypes` (ie. `ClusterLocalIPAddress`) but
this would lead to a combinatorial explosion as new dimensions (ie. IPv6) are
introduced.

From: [https://github.com/kubernetes-sigs/gateway-api/pull/1653#issuecomment-1451246877](https://github.com/kubernetes-sigs/gateway-api/pull/1653#issuecomment-1451246877)

> Although this makes sense in isolation, I'm worried about the long term impacts this could have. In my opinion, ClusterLocal is a modifier, not exactly an address type. For example, it's possible in the future that we'll have a way to provision cluster-local DNS names, we may want to use the same kind of mechanism to request a ClusterLocal DNS name for the Gateway.
>
> It's also possible that users will want to explicitly request an IP Families (v4, v6, or both). I'd really hate to get into a situation where we have the following options:
>
> - IPAddress
> - IPv4Address
> - IPv6Address
> - ClusterLocalIPAddress
> - ClusterLocalIPv4Address
> - ClusterLocalIPv6Address
>
> For each dimension we avoid adding a separate field for and instead try to embed into a single name, we risk this kind of name explosion. Of course, none of the above even begins to cover my idea of NetworkLocal which could further complicate this.

### Scope/reachability/routability field on GatewayAddress

This would allow Gateways to have multiple scopes.

From: [https://github.com/kubernetes-sigs/gateway-api/pull/1653#issuecomment-1486271913](https://github.com/kubernetes-sigs/gateway-api/pull/1653#issuecomment-1486271913)
> The obvious application for multiple scopes seems to be saving on boilerplate, which is a win, but are there are any other advantages to allowing one Gateway to have multiple scopes?
>
> Multiple scopes Pros:
>
> Allows a single Gateway to express multiple networks, saving on needing to attach HTTPRoutes to multiple Gateways for each network scope.
>
> Multiple scopes Cons:
>
> Complicates the Gateway's purpose. Instead of one Gateway being one set of Listeners, now a Gateway is two sets of listeners that have a totally different scope (and presumably, security context). Personally, I'm also concerned how this will interact with other features like merging and preprovisioning that GEP-1867: Per-Gateway Infrastructure #1868 will allow.

### Adding `routability` attribute to GatewayClass

See [Prior Art - Multiple Gateways Classes](#multiple-gateway-classes)

## Survey of Prior Art

These alternatives are a survey of existing approaches to support cluster
local Gateways. Most are implementation specific and are not portable.

### Special annotation/label

Istio let's you specify an annotation `networking.istio.io/service-type` to
change the underlying Kubernetes Service type to make it a ClusterIP type.

### Re-use of AddressType Hostname

Istio let's you re-use existing Gateway deployments by setting the address
type to `Hostname` and the value to the Istio ingress Kubernetes Service. If an
operator configures the Istio deployment to support cluster local traffic a
Gateway implementation can select it using the `HostName` attribute.

### Multiple Gateway Classes

Some implementations support multiple deployments on a single cluster where each maps to a
GatewayClass. One of these deployments can be configured to serve cluster local traffic. This is
sub-optimal because this is implementation specific and the end-user is effectively managing the
deployments themselves rather than infrastructure being automatically provisioned.

Likewise, infrastructure providers may provide a fixed set of GatewayClasses with unique and fixed
routability. Thus GatewayClass name is a viable option to control routability. There may be a
non-zero cost when requiring additional GatewayClasses - but this depends on the implementation.

Additionally, if more attributes are added to GatewayClass to constrain Gateways in some
form this leads to a combinatorial number of GatewayClassNames. For example, `foo-public` and
`foo-cluster` are two GatewayClasses surfacing the values of a single attribute `routability`.
Let's say we want to enforce address types to just IP then our `gatewayClassName` would be:

- `foo-public-ipv4`
- `foo-public-ipv6`
- `foo-cluster-ipv4`
- `foo-cluster-ipv6`

This may not be as flexible for end-users compared to configuring `routability` when creating
a Gateway.

As mentioned in [howardjohn's comment on GEP-1651: Gateway Routability](https://github.com/kubernetes-sigs/gateway-api/pull/1653#issuecomment-1429992160):
> having the ability to configure things at a higher level seems nice for Gateway, but being able to configure them on a per-Gateway basis remains important.

## References

- [Knative - Private Services](https://knative.dev/docs/serving/services/private-services/#configuring-private-services)
- [Initial Gateway GitHub Discussion](https://github.com/kubernetes-sigs/gateway-api/discussions/1247)
- [Istio Support for Private Gateways](https://istio.io/latest/docs/tasks/traffic-management/ingress/gateway-api/#automated-deployment)
- [Envoy Gateway Support for Private Gateways](https://gateway.envoyproxy.io/latest/api/config_types.html#kubernetesservicespec)

