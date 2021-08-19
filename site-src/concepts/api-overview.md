# API Overview

This document provides an overview of Gateway API.

## Roles and personas.

There are 3 primary roles in Gateway API:

- Infrastructure Provider
- Cluster Operator
- Application Developer

There could be a fourth role of Application Admin in some use cases.

Please refer to the [roles and personas](/concepts/security-model#roles-and-personas)
section in the Security model for details.

## Resource model

!!! note
    As of v1alpha2, resources are in the `gateway.networking.k8s.io` API group as
    Custom Resource Definitions (CRDs). Unqualified resource names below will implicitly
    be in this API group. Prior to v1alpha1, the apigroup was `networking.x-k8s.io`.

There are three main types of objects in our resource model:

*GatewayClass* defines a set of gateways with a common configuration and
behavior.

*Gateway* requests a point where traffic can be translated to Services within
the cluster.

*Routes* describe how traffic coming via the Gateway maps to the Services.

### GatewayClass

GatewayClass defines a set of Gateways that share a common configuration and
behaviour. Each GatewayClass will be handled by a single controller, although
controllers MAY handle more than one GatewayClass.

GatewayClass is a cluster-scoped resource. There MUST be at least one
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
GatewayClass’ configuration and behaviour contract. The resource MAY be created
by an operator directly, or MAY be created by a controller handling a
GatewayClass.

As the Gateway spec captures user intent, it may not contain a complete
specification for all attributes in the spec. For example, the user may omit
fields such as addresses, TLS settings. This allows the controller
managing the GatewayClass to provide these settings for the user, resulting in a
more portable spec. This behaviour will be made clear using the GatewayClass
Status object.

A Gateway MAY contain one or more *Route references which serve to direct
traffic for a subset of traffic to a specific service.*

### Route objects

Route objects define protocol-specific rules for mapping requests from a Gateway
to Kubernetes Services.

As of v1alpha2, there are four Route types defined in this repo, although it's
expected that in the future implementations may create their own custom Route
types.

#### HTTPRoute
HTTPRoute is for multiplexing HTTP or terminated HTTPS connections onto a single port or set of ports. It's intended for use in cases where you want to inspect the HTTP stream and use HTTP-level data for either routing or modification. (An example here is using HTTP Headers for routing, or modifying them in-flight).

#### TLSRoute
TLSRoute is for multiplexing TLS connections, discriminated via SNI, onto a single port or set or ports. It's intended for where you want to use the SNI as the main routing method, and are not interested in properties of the underlying connection (such as whether or not it is HTTP).

#### TCPRoute and UDPRoute
TCPRoute (and UDPRoute) are intended for use as a mapping between a single port or set of ports and a single backend. In this case, there is no discriminator you can use to choose different backends on the same port, so each TCPRoute really needs a different port on the listener (in general, anyway).

#### Route summary table
The "Routing Discriminator" column below refers to what information can be used to allow multiple Routes to share ports on the Listener.

|Object|OSI Layer|Routing Discriminator|TLS Support|Purpose|
|------|---------|---------------------|-----------|-------|
|HTTPRoute| Layer 7 | Anything in the HTTP Protocol | Terminated only, can be reencrypted| HTTP and HTTPS Routing|
|TLSRoute| Somewhere between layer 4 and 7| SNI or other TLS properties| Passthrough or terminated, can be reencrypted if terminated. | Routing of TLS protocols including HTTPS where inspection of the HTTP stream is not required.|
|TCPRoute| Layer 4| None | None (but passthrough will work because the connection is forwarded) | Allows for forwarding of a TCP stream from the Listener to the Backends |
|UDPRoute| Layer 4| None | None | Allows for forwarding of a UDP stream from the Listener to the Backends. |


### Route binding

When a Route binds to a Gateway it represents configuration that is applied on
the Gateway that configures the underlying load balancer or proxy. How and which
Routes bind to Gateways is controlled by the resources themselves. Route and
Gateway resources have built-in controls to permit or constrain how they select
valid partners. This is useful for enforcing organizational policies for how Routes
are exposed and on which Gateways. Consider the following example:

> A Kubernetes cluster admin has deployed a Gateway “shared-gw” in the “Infra”
> Namespace to be used by different application teams for exposing their
> applications outside the cluster. Teams A and B (in Namespaces “A” and “B”
> respectively) bind their Routes to this Gateway. They are unaware of each other
> and as long as their Route rules do not conflict with each other they can
> continue operating in isolation. Team C has special networking needs (perhaps
> performance, security, or criticality) and they need a dedicated Gateway to
> proxy their application to the outside world. Team C deploys their own Gateway
> “dedicated-gw”  in the “C” Namespace that can only be used by apps in the "C"
> Namespace.

<!-- source: https://docs.google.com/presentation/d/1neBkFDTZ__vRoDXIWvAcxk2Pb7-evdBT6ykw_frf9QQ/edit?usp=sharing -->
![route binding](/images/gateway-route-binding.png)

There is a lot of flexibility in how Routes can bind to Gateways to achieve
different organizational policies and scopes of responsibility. These are
different relationships that Gateways and Routes can have:

- **One-to-one** - A Gateway and Route may be deployed and used by a single
  owner and have a one-to-one relationship. Team C is an example of this.
- **One-to-many** - A Gateway can have many Routes bound to it that are owned by
  different teams from across different Namespaces. Teams A and B are an example
  of this.
- **Many-to-one** - Routes can also be bound to more than one Gateway, allowing
  a single Route to control application exposure simultaneously across different
  IPs, load balancers, or networks.

*In summary, Routes attach to Gateways and Gateways choose what attachments to
allow. When a Route tries to attach to a Gateway that does not prevent it, then
the Route will bind to the Gateway. When Routes are bound to a Gateway it means
their collective routing rules are configured on the underlying load balancers
or proxies that are managed by that Gateway. Thus, a Gateway is a logical
representation of a networking data plane that can be configured through
Routes.*

#### Route binding handshake

A Route *must* select what Gateway it wants to attach to, based on the
`parentRefs` field, which allows the selection of the Group, Kind, name,
and namespace of the object. Cluster-scoped objects can also be selected by
changing the Scope to `Cluster`. Although only Gateways are currently supported,
this is intended to allow for later extension. Additionally, the `parentRefs`
stanza is a list, so a Route may request to attach to more than one Gateway
(or other parent object).

Additionally, Gateways *may* specify what kind of Routes they support
(defaults to Routes that match the Listener protocol if not specified), and
where those Routes can be (defaults to same namespace). If a Route wants
to attach to a Gateway in another namespace, that Gateway must *explicitly*
allow Routes from its namespace for the binding to succeed.

The Route becomes attached only when the Gateway and Route specifications intersect.
Note that this means that the binding requires bidirectional agreement between
the two objects. This is a critical part of the API's role-based structure.

#### Gateway - Route binding examples

The following `my-route` Route wants to attach to the `foo-gateway` in the
`foo-namespace` and will not bind with any other Gateways. Note that
`foo-gateway` is in a different Namespace. The `foo-gateway` must allow
bindings from HTTPRoutes in the namespace `bar-namespace`.

```yaml
kind: HTTPRoute
metadata:
  name: my-route
  namespace: bar-namespace
spec:
  parentRefs:
  - kind: Gateway
    name: foo-gateway
    namespace: foo-namespace
...
```

This `foo-gateway` allows the `my-route` HTTPRoute to bind.

```yaml
kind: Gateway
metadata:
  name: foo-gateway
  namespace: foo-namespace
spec:
  listeners:
  - name: prod-web
    routes:
      kinds:
      - HTTPRoute
      namespaces:
      - from: bar-namespace
```

For a more permissive example, the below Gateway will allow all HTTPRoute resources
to attach from Namespaces with the "expose-apps: true" label.

```yaml
kind: Gateway
...
spec:
  listeners:
  - name: prod-web
    routes:
      kinds:
      - HTTPRoute
      namespaces:
      - from: Selector
        selector:
          matchLabels:
            expose-apps: "true"
```

It may not always be apparent from the resource specifications which Gateways
and Routes are bound, but binding can be determined from the resource status.
The [Route status](/api-types/httproute#routestatus) will list all of the Gateways that
a Route is bound to and any relevant conditions for the binding.

### Combined types

The combination of `GatewayClass`, `Gateway`, `xRoute` and `Service`(s)
defines an implementable load-balancer. The diagram below illustrates the
relationships between the different resources:

<!-- source: https://docs.google.com/document/d/1BxYbDovMwnEqe8lj8JwHo8YxHAt3oC7ezhlFsG_tyag/edit#heading=h.8du598fded3c -->
![schema](/images/schema-uml.svg)

## Request flow

A typical client/gateway API request flow for a gateway implemented using a
reverse proxy is:

 1. A client makes a request to http://foo.example.com.
 2. DNS resolves the name to a `Gateway` address.
 3. The reverse proxy receives the request on a `Listener` and uses the [Host
 header](https://tools.ietf.org/html/rfc7230#section-5.4) to match an
 `HTTPRoute`.
 5. Optionally, the reverse proxy can perform request header and/or path
 matching based on `match` rules of the `HTTPRoute`.
 6. Optionally, the reverse proxy can modify the request, i.e. add/remove
 headers, based on `filter` rules of the `HTTPRoute`.
 7. Lastly, the reverse proxy forwards the request to one or more objects, i.e.
 `Service`, in the cluster based on `forwardTo` rules of the `HTTPRoute`.

## TLS Configuration

TLS is configured on Gateway listeners, and may be referred to across namespaces.

Please refer to the [TLS details](/v1alpha2/guides/tls) guide for a deep dive on TLS.


## Extension points

A number of extension points are provided in the API to provide flexibility in
addressing the large number of use-cases that cannot be addressed by a general
purpose API.

Here is a summary of extension points in the API:

- **XRouteMatch.ExtensionRef**: This extension point should be used to extend
  the match semantics of a specific core Route. This is an experimental
  extension point and will be iterated on in future based on feedback.
- **XForwardTo.BackendRef**: This extension point should be used for forwarding
  traffic to network endpoints other than core Kubernetes Service resource.
  Examples include an S3 bucket, Lambda function, a file-server, etc.
- **HTTPRouteFilter**: This API type in HTTPRoute provides a way to hook into
the request/response lifecycle of an HTTP request.
- **Custom Routes**: If none of the above extensions points suffice for a use
  case, Implementers can chose to create custom Route resources for protocols
  that are not currently supported in the API.

Whenever you are using an extension point without any prior art, please let
the community know. As we learn more about usage of extension points, we would
like to find the common denominators and promote the features to core/extended
API conformance.

