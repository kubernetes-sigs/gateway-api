# GEP-1911: Backend Protocol Selection

* Issue: [#1911](https://github.com/kubernetes-sigs/gateway-api/issues/1911)
* Status: Standard

(See [status definitions](../overview.md#gep-states).)

## TLDR

Not all implementations support automatic protocol selection. Even in some cases protocols are disabled without an explicit opt-in (eg. websockets with Contour & NGINX). Thus application developers need the ability to specify the protocol(s) that their application supports.

## Goals

- Support protocols that can have a Gateway `*Route` resource as a frontend
- Standardize Gateway API implementations on the protocols & constants defined by the Kubernetes [Standard Application Protocols (KEP-3726)][kep-3726]
- Support backends with multiple protocols on the same port (ie. tcp/udp)

## Non-Goals

- Backend TLS (covered in [GEP-1897](https://github.com/kubernetes-sigs/gateway-api/issues/1897))
- Additional protocol specific configuration
- Disabling Protocols

## Introduction

Since Kubernetes 1.20 the [`core/v1.Service`][k8s-service] and [`core/v1.EndpointSlice`][k8s-endpointslices] resource has a stable `appProtocol` field. It's purpose is to allow end-users to specify an application protocol (L7) for each service port.

Originally the use of this field in the Gateway API was rejected in [GEP-1282](../gep-1282/index.md#non-goals):
> v1.Service’s appProtocol field is not fit for purpose, because it is defined as accepting values either from the IANA Service Name registry, or domain-prefixed values and we need more flexibility than that.

Since then a Kubernetes enhancement proposal was created [KEP-3726][kep-3726] to repurpose `appProtocol` to include a convention for protocols that are not IANA service names. This would involve prefixing protocol names with `kubernetes.io/*`.

Note: Kubernetes will automatically create `EndpointSlices` for `Services` that have a selector. [Custom `EndpointSlices`](https://kubernetes.io/docs/concepts/services-networking/service/#custom-endpointslices) can manually be created.

## API Semantics

A Gateway implementation MUST recognize the Kubernetes Standard Application Protocols ([KEP-3726][kep-3726]) for specifying the protocol for a backend reference in a Gateway API `*Route` resource

Thus when a `*Route` points to a Kubernetes Service, implementations SHOULD honor the appProtocol field if it
is set for the target Service Port.


At the moment there exists three defined constants:

- `kubernetes.io/h2c` - HTTP/2 over cleartext as described in [RFC7540](https://www.rfc-editor.org/rfc/rfc7540)
- `kubernetes.io/ws` - WebSocket over cleartext as described in [RFC6445](https://www.rfc-editor.org/rfc/rfc6455)
- `kubernetes.io/wss` - WebSocket over TLS as described in [RFC6455](https://www.rfc-editor.org/rfc/rfc6455)

### New Protocols & Reserved Prefix

To add support for a new protocol it should first become a Kubernetes Standard Application Protocol by updating the [KEP-3726][kep-3726]. [KEP-3726][kep-3726] also states the `appProtocol` field accepts a domain-prefixed implementation specific value. Thus, if the suggested protocol is not suited to have a `kubernetes.io/*` prefix, then the Gateway API MAY support the new protocol using its own prefix `gateway.networking.k8s.io/*`.  Please make a PR to this GEP.

For example we may want to add a sentinel `appProtocol` value that prevents Gateway implementations from discovering the protocol of the application. Instead they should just refer to the Service's `protocol` field. Such a constant was rejected upstream (https://github.com/kubernetes/enhancements/pull/4106) but as an example it could be defined in a future addition to this GEP as `gateway.networking.k8s.io/no-sniff`.

### Default Protocols

If a Service `appProtocol` isn't specified an implementation MAY infer the backend protocol through its own means. Implementations MAY infer the protocol from the `Route` type referring to the backend Service.

Absence of the `appProtocol` field does not imply the implementation should disable any features (eg. websocket upgrades).

### Multiple Protocols on the Same Port

Only the Kubernetes `Service` `protocol` field supports multiple protocols on the same port. See the details in [KEP-1435][kep-1435].

Implementations MAY support Kubernetes Service BackendRefs that are multiplexing TCP and UDP on the same port. Otherwise implementations MUST set *Route ResolvedRefs condition to False with the "UnsupportedProtocol" Reason with a clear message that multiplexing is not supported.

Currently Kubernetes `Service` API only allows different `appProtocol` values for the same port when `protocol` fields differs. At this time there seems to be interest in changing `appProtocol` to be a list in order to facilitate this use-case.

### Supporting Protocols

If a Route is not able to send traffic to the backend using the specified protocol then the backend is considered invalid. Implementations MUST set `ResolvedRefs` condition to `False` with the Reason `UnsupportedProtocol`.

Implementations MAY support the following combinations below:

ServicePort Protocol | ServicePort AppProtocol | Route Type | Supported
-|-|-|-
`TCP` | `kubernetes.io/h2c` | `GRPCRoute` | Yes [1]
`TCP` | `kubernetes.io/h2c` | `HTTPRoute` | Yes
`TCP` | `kubernetes.io/ws`  | `HTTPRoute` | Yes
`TCP` | `kubernetes.io/wss` | `TLSRoute`  | Yes

1. GRPC works over h2c - so a GRPCRoute should be able to connect to an h2c backend

Implementations MAY support the following combinations below:

ServicePort Protocol | ServicePort AppProtocol | Route Type | Supported
-|-|-|-
`TCP`  | `kubernetes.io/wss` | `HTTPRoute` | Conditional [1]

1. Only if there is a corresponding `BackendTLSPolicy` - see [GEP-1897](../gep-1897/index.md)

## Open Questions

1. TLSRoute & UDP protocol

TLS over UDP seems to be a thing via QUIC/HTTP3 [ref](https://www.smashingmagazine.com/2021/08/http3-core-concepts-part1/).
Likewise there's also [DTLS](https://en.wikipedia.org/wiki/Datagram_Transport_Layer_Security). But it's unclear if Gateway's TLSRoute
applies to an underlying UDP protocol.

2. Websockets & HTTP/2/3

Should we upstream new constants for websocket over [HTTP/2](https://www.rfc-editor.org/rfc/rfc8441.html) & [HTTP/3](https://www.rfc-editor.org/rfc/rfc9220.html) ? HTTP/3 makes things more complicated since its supports UDP as the underlying protocol.

## Alternatives

### Single Meta-resource

The first pass of this GEP proposed a new meta-resource [GEP-713](../gep-713/index.md) called `BackendProtocol`.

This allows end-users to specify a list of ports and a list of corresponding protocols that that single
port supports.

This was dropped in favour of supporting Kubernetes Standard Application Protocols.

### Multiple Protocol Meta-resources

Rather than bundle protocol details into a single resource an alternative would be to create distinct meta resources.
ie. `HTTP2Backend`, `GRPCBackend`, `WebsocketBackend`.

The advantages of this approach are:

- Easy to introduce new protocols
- Definitions/types would be simpler

The disadvantages of this approach are:

- N resources for N protocols need to be created to describe a single backend
- No easy mechanic to specify priority of protocols

### Adding Properties on Gateway Route Objects

From [GEP-1282](../gep-1282/index.md#tldr):
> some types of configuration requested by users are more about defining functionality that describes capabilities of the backend more than the route you take to get to the backend.

Backend protocol is specifying capabilities. This configuration is less about routing.

### Kubernetes Service - Expanding Protocol field

The `protocol` field on a Kubernetes service is used to specify a L4 protocol over IP. This field isn't appropriate to describe protocols
that operate at a higher 'application' level (eg. HTTP/GRPC etc.)

### Extending Kubernetes Service

This is considered untenable due to the 'the turnaround time for those changes can be years.' ([ref-1282](../gep-1282/index.md#non-goals))

### Unstructured Data/Special Values

Unstructured data refers to using labels and annotations.

From [GEP-1282](../gep-1282/index.md#non-goals):
> these are very sticky and hard to get rid of once you start using them.

Special values refers to using special strings in existing Kubernetes Resources.
For example Istio allows for protocol to be specified by prefixing the Kubernetes
Service's port name with the protocol (ie. `http-`, `grpc-`). This approach is
limiting as it doesn't allow for multiple protocols on the same port and future
configuration per protocol. One protocol per port may be relaxed in the future see
[KEP 1435][kep-1435]

Additionally, annotations are not self-documenting unlike CRD fields which can display
documentation via `kubectl explain`

## References

- [GitHub Discussion](https://github.com/kubernetes-sigs/gateway-api/discussions/1244)
- GEP-1282 - Describing Backend Properties
    - [GEP](../gep-1282/index.md)
    - [Issue](https://github.com/kubernetes-sigs/gateway-api/issues/1911)
- [GEP-713 - Metaresources](../gep-713/index.md)
- [Linkerd Protocol Detection](https://linkerd.io/2.12/features/protocol-detection/)
- [Istio Protocol Selection](https://istio.io/latest/docs/ops/configuration/traffic-management/protocol-selection/)
- Contour Protocol Selection
    - [Websockets](https://projectcontour.io/docs/1.24/config/websockets/)
    - [GRPC](https://projectcontour.io/docs/1.24/guides/grpc/#httpproxy-configuration)
- [AWS Gateway Protocol Selection](https://github.com/aws/aws-application-networking-k8s/blob/a277fb39449383f53cd7d1e5576b4fa190a1a853/config/crds/bases/application-networking.k8s.aws_targetgrouppolicies.yaml#L109)
- [Google GKE AppProtocol Selection](https://cloud.google.com/kubernetes-engine/docs/concepts/ingress-xlb#https_tls_between_load_balancer_and_your_application)

[k8s-service]: https://kubernetes.io/docs/concepts/services-networking/service/
[k8s-endpointslices]: https://kubernetes.io/docs/concepts/services-networking/endpoint-slices/
[kep-3726]: https://github.com/kubernetes/enhancements/tree/master/keps/sig-network/3726-standard-application-protocols
[kep-1435]: https://github.com/kubernetes/enhancements/tree/master/keps/sig-network/1435-mixed-protocol-lb
