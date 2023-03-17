# API Overview

This document provides an overview of Gateway API.

## Roles and personas

There are 3 primary roles in Gateway API:

- Infrastructure Provider
- Cluster Operator
- Application Developer

There could be a fourth role of Application Admin in some use cases.

Please refer to the [roles and personas](/concepts/security-model#roles-and-personas)
section in the Security model for details.

## Resource model

!!! note
    Gateway API resources are in the `gateway.networking.k8s.io` API group as
    Custom Resource Definitions (CRDs). Unqualified resource names below will
    implicitly be in this API group.

There are three main types of objects in our resource model:

*GatewayClass* defines a set of gateways with a common configuration and
behavior.

*Gateway* requests a point where traffic can be translated to Services within
the cluster.

*Routes* describe how traffic coming via the Gateway maps to the Services.

### GatewayClass

GatewayClass defines a set of Gateways that share a common configuration and
behaviour. Each GatewayClass will be handled by a single controller, although
controllers may handle more than one GatewayClass.

GatewayClass is a cluster-scoped resource. There must be at least one
GatewayClass defined in order to be able to have functional Gateways. A
controller that implements the Gateway API does so by providing an associated
GatewayClass resource that the user can reference from their Gateway(s).

This is similar to
[IngressClass](https://kubernetes.io/docs/concepts/services-networking/ingress/#ingress-class)
for Ingress and
[StorageClass](https://kubernetes.io/docs/concepts/storage/storage-classes/) for
PersistentVolumes. In Ingress v1beta1, the closest analog to GatewayClass is the
`ingress-class` annotation, and in IngressV1, the closest analog is the
IngressClass object.

### Gateway

A Gateway describes how traffic can be translated to Services within the
cluster. That is, it defines a request for a way to translate traffic from
somewhere that does not know about Kubernetes to somewhere that does. For
example, traffic sent to a Kubernetes Service by a cloud load balancer, an
in-cluster proxy, or an external hardware load balancer. While many use cases
have client traffic originating “outside” the cluster, this is not a
requirement.

It defines a request for a specific load balancer config that implements the
GatewayClass’ configuration and behaviour contract. The resource may be created
by an operator directly, or may be created by a controller handling a
GatewayClass.

As the Gateway spec captures user intent, it may not contain a complete
specification for all attributes in the spec. For example, the user may omit
fields such as addresses, TLS settings. This allows the controller
managing the GatewayClass to provide these settings for the user, resulting in a
more portable spec. This behaviour will be made clear using the GatewayClass
Status object.

A Gateway may be attached to one or more *Route references* which serve to direct
traffic for a subset of traffic *to a specific service.*

### Route Resources

Route resources define protocol-specific rules for mapping requests from a Gateway
to Kubernetes Services.

As of v1alpha2, four Route resource types are included with the API. Custom Route types
that are implementation-specific are encouraged for other protocols. New route
types may be added to the API in future.

#### HTTPRoute

HTTPRoute is for multiplexing HTTP or terminated HTTPS connections. It's intended
for use in cases where you want to inspect the HTTP stream and use HTTP request data
for either routing or modification, for example using HTTP Headers for routing, or
modifying them in-flight.

#### TLSRoute

!!! info "Experimental Channel"

    The `TLSRoute` resource described below is currently only included in the
    "Experimental" channel of Gateway API. For more information on release
    channels, refer to the [related documentation](https://gateway-api.sigs.k8s.io/concepts/versioning).

TLSRoute is for multiplexing TLS connections, discriminated via SNI. It's intended
for where you want to use the SNI as the main routing method, and are not interested
in properties of the higher-level protocols like HTTP.  The byte stream of the
connection is proxied without any inspection to the backend.

#### TCPRoute and UDPRoute

!!! info "Experimental Channel"

    The `TCPRoute` and `UDPRoute` resources described below are currently only included in the
    "Experimental" channel of Gateway API. For more information on release
    channels, refer to the [related documentation](https://gateway-api.sigs.k8s.io/concepts/versioning).

TCPRoute (and UDPRoute) are intended for use for mapping one or more ports
to a single backend. In this case, there is no discriminator you can
use to choose different backends on the same port, so each TCPRoute really needs a
different port on the listener (in general, anyway). You can terminate TLS in
which case the unencrypted byte stream is passed through to the backend.
You can choose to not terminate TLS, in which case the encrypted byte stream
is passed through to the backend.

#### GRPCRoute

!!! info "Experimental Channel"

    The `GRPCRoute` resource described below is currently only included in the
    "Experimental" channel of Gateway API. For more information on release
    channels, refer to the [related documentation](https://gateway-api.sigs.k8s.io/concepts/versioning).

GRPCRoute is for idiomatically routing gRPC traffic. Gateways supporting
GRPCRoute are required to support HTTP/2 without an initial upgrade from HTTP/1,
so gRPC traffic is guaranteed to flow properly.

#### Route summary table

The "Routing Discriminator" column below refers to what information can be used
to allow multiple Routes to share ports on the Listener.

|Object|OSI Layer|Routing Discriminator|TLS Support|Purpose|
|------|---------|---------------------|-----------|-------|
|HTTPRoute| Layer 7 | Anything in the HTTP Protocol | Terminated only | HTTP and HTTPS Routing|
|TLSRoute| Somewhere between layer 4 and 7| SNI or other TLS properties| Passthrough or Terminated | Routing of TLS protocols including HTTPS where inspection of the HTTP stream is not required.|
|TCPRoute| Layer 4| destination port | Passthrough or Terminated | Allows for forwarding of a TCP stream from the Listener to the Backends |
|UDPRoute| Layer 4| destination port | None | Allows for forwarding of a UDP stream from the Listener to the Backends. |
|GRPCRoute| Layer 7 | Anything in the gRPC Protocol | Terminated only | gRPC Routing over HTTP/2 and HTTP/2 cleartext|

Note that traffic routed via HTTPRoute and TCPRoute can be encrypted between the
Gateway and backend (commonly referred to as reencryption). It is not possible
to configure that with existing Gateway API resources, but implementations may
provide custom configuration for this until there is a standardized approach
defined by Gateway API.

### Attaching Routes to Gateways

!!! note
    This section has changed significantly between v1alpha1 and v1alpha2. This
    section describes the v1alpha2 approach.

When a Route attaches to a Gateway, it represents configuration that is applied
on the Gateway that configures the underlying load balancer or proxy. How and
which Routes attach to Gateways is controlled by the resources themselves. Route
and Gateway resources have built-in controls to permit or constrain how they are
attached. Together with Kubernetes RBAC, these allow organizations to enforce
policies for how Routes are exposed and on which Gateways.

There is a lot of flexibility in how Routes can attach to Gateways to achieve
different organizational policies and scopes of responsibility. These are
different relationships that Gateways and Routes can have:

- **One-to-one** - A Gateway and Route may be deployed and used by a single
  owner and have a one-to-one relationship.
- **One-to-many** - A Gateway can have many Routes bound to it that are owned by
  different teams from across different Namespaces.
- **Many-to-one** - Routes can also be bound to more than one Gateway, allowing
  a single Route to control application exposure simultaneously across different
  IPs, load balancers, or networks.

### Example

A Kubernetes cluster admin has deployed a Gateway `shared-gw` in the `Infra`
Namespace to be used by different application teams for exposing their
applications outside the cluster. Teams A and B (in Namespaces `A` and `B`
respectively) attach their Routes to this Gateway. They are unaware of each
other and as long as their Route rules do not conflict with each other they
can continue operating in isolation. Team C has special networking needs
(perhaps performance, security, or criticality) and they need a dedicated
Gateway to proxy their application to the outside world. Team C deploys their
own Gateway `dedicated-gw`  in the `C` Namespace that can only be used by apps
in the `C` Namespace.

<!-- source: https://docs.google.com/presentation/d/1neBkFDTZ__vRoDXIWvAcxk2Pb7-evdBT6ykw_frf9QQ/edit?usp=sharing -->
![route binding](/images/gateway-route-binding.png)

### How it Works

The following is required for a Route to be attached to a Gateway:

1. The Route needs an entry in its `parentRefs` field referencing the Gateway.
2. At least one listener on the Gateway needs to allow this attachment.

#### Referencing Gateways

!!! info "Experimental Channel"

    The `Port` field described below is currently only included in the
    "Experimental" channel of Gateway API. For more information on release
    channels, refer to the [related documentation](https://gateway-api.sigs.k8s.io/concepts/versioning/#adding-experimental-fields).

A Route can reference a Gateway by specifying the namespace (optional if the
Route and the Gateway are in the same namespace) and name of the Gateway in
a `parentRef`. A Route can further select a subset of listeners under the
Gateway using the following fields in `parentRef`:

1. **SectionName** When `sectionName` is set, the Route selects the listener
   with the specified name.
2. **Port** When `port` is set, the Route selects all listeners listening on
   the specified port and with protocol compatible with this kind of Route.

When multiple fields in `parentRef` are set, the Route selects listeners that
satisfy all conditions specified in those fields. For example, when both
`sectionName` and `port` are set, the Route selects listeners with the specified
name AND listening on the specified port.

#### Restricting Route Attachment

Each Gateway listener can restrict which Routes can be attached with the
following mechanisms:

1. **Hostname:** When the `hostname` field on a listener is set, attached Routes
   that specify a `hostnames` field must have at least one overlapping value.
2. **Namespaces:** The `allowedRoutes.namespaces` field on a listener can be
   used to restrict where Routes may be attached from. The `namespaces.from`
   field supports the following values:
    * `SameNamespace` is the default option. Only Routes in the same namespace
      as this Gateway may be attached.
    * `All` will allow Routes from all Namespaces to be attached.
    * `Selector` means that Routes from a subset of Namespaces selected by a
      Namespace label selector may be attached to this Gateway. When `Selector`
      is used, the `namespaces.selector` field must be used to specify label
      selectors. This field is not supported with `All` or `SameNamespace`.
3. **Kinds:** The `allowedRoutes.kinds` field on a listener can be used to
   restrict the kinds of Routes that may be attached.

If none of the above are specified, a Gateway listener will trust Routes
attached from the same namespace that support the listener protocol.

#### Further Gateway - Route attachment examples

The following `my-route` Route wants to attach to the `foo-gateway` in the
`gateway-api-example-ns1` and will not attach to any other Gateways. Note that
`foo-gateway` is in a different Namespace. The `foo-gateway` must allow
attachment from HTTPRoutes in the namespace `gateway-api-example-ns2`.

```yaml
{% include 'standard/http-route-attachment/httproute.yaml' %}
```

This `foo-gateway` allows the `my-route` HTTPRoute to attach.

```yaml
{% include 'standard/http-route-attachment/gateway-strict.yaml' %}
```

For a more permissive example, the below Gateway will allow all HTTPRoute resources
to attach from Namespaces with the "expose-apps: true" label.

```yaml
{% include 'standard/http-route-attachment/gateway-namespaces.yaml' %}
```

### Combined types

The combination of `GatewayClass`, `Gateway`, `xRoute` and `Service`(s)
defines an implementable load-balancer. The diagram below illustrates the
relationships between the different resources:

<!-- source: https://docs.google.com/document/d/1BxYbDovMwnEqe8lj8JwHo8YxHAt3oC7ezhlFsG_tyag/edit#heading=h.8du598fded3c -->
![schema](/images/schema-uml.svg)

## Request flow

A typical client/gateway API request flow for a gateway implemented using a
reverse proxy is:

 1. A client makes a request to <http://foo.example.com>.
 2. DNS resolves the name to a `Gateway` address.
 3. The reverse proxy receives the request on a `Listener` and uses the [Host
 header](https://tools.ietf.org/html/rfc7230#section-5.4) to match an
 `HTTPRoute`.
 4. Optionally, the reverse proxy can perform request header and/or path
 matching based on `match` rules of the `HTTPRoute`.
 5. Optionally, the reverse proxy can modify the request, i.e. add/remove
 headers, based on `filter` rules of the `HTTPRoute`.
 6. Lastly, the reverse proxy forwards the request to one or more objects, i.e.
 `Service`, in the cluster based on `backendRefs` rules of the `HTTPRoute`.

## TLS Configuration

TLS is configured on Gateway listeners, and may be referred to across namespaces.

Please refer to the [TLS details](/guides/tls) guide for a deep dive on TLS.

## Extension points

A number of extension points are provided in the API to provide flexibility in
addressing the large number of use-cases that cannot be addressed by a general
purpose API.

Here is a summary of extension points in the API:

- **BackendRefs**: This extension point should be used for forwarding
  traffic to network endpoints other than core Kubernetes Service resource.
  Examples include an S3 bucket, Lambda function, a file-server, etc.
- **HTTPRouteFilter**: This API type in HTTPRoute provides a way to hook into
the request/response lifecycle of an HTTP request.
- **Custom Routes**: If none of the above extensions points suffice for a use
  case, Implementers can choose to create custom Route resources for protocols
  that are not currently supported in the API. Custom Route types need to share
  the same fields that core Route types do. These are contained within
  CommonRouteSpec and RouteStatus.

Whenever you are using an extension point without any prior art, please let
the community know. As we learn more about usage of extension points, we would
like to find the common denominators and promote the features to core/extended
API conformance.
