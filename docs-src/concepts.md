# API Concepts

This document is a deep dive into the reasoning and design for Service APIs.

## Roles and personas.

There are 3 primary roles in the API:

- Infrastructure Provider
- Cluster Operator
- Application Developer

There could be a fourth role of Application Admin in some use cases.

Please refer to the [roles and personas](security-model.md#roles-and-personas) section
in the Security model for details.

## Resource model

> Note: Resources will initially live in the `networking.x-k8s.io` API group as
> Custom Resource Definitions (CRDs). Unqualified resource names will implicitly
> be in this API group.

There are three main types of objects in our resource model:

*GatewayClass* defines a set of gateways with a common configuration and behavior.

*Gateway* requests a point where traffic can be translated to Services within the cluster.

*Routes* describe how traffic coming via the Gateway maps to the Services.

### GatewayClass

GatewayClass defines a set of Gateways that share a common configuration and behaviour.
Each GatewayClass will be handled by a single controller, although controllers MAY handle more than one.

GatewayClass is a cluster-scoped resource.
There MUST be at least one GatewayClass defined in order to be able to have functional Gateways.
A controller that implements the Gateway API does so by providing an associated GatewayClass resource that the user can reference from their Gateway(s).

This is similar to [IngressClass](https://github.com/kubernetes/enhancements/blob/master/keps/sig-network/20190125-ingress-api-group.md#ingress-class) for Ingress and [StorageClass](https://kubernetes.io/docs/concepts/storage/storage-classes/) for PersistentVolumes.
In Ingress v1beta1, the closest analog to GatewayClass is the `ingress-class` annotation, and in IngressV1, the closest analog is the IngressClass object.

### Gateway

A Gateway describes how traffic can be translated to Services within the cluster.
That is, it defines a request for a way to translate traffic from somewhere that does not know about Kubernetes to somewhere that does.
For example, traffic sent to a Kubernetes Services by a cloud load balancer, an in-cluster proxy or external hardware load balancer.
While many use cases have client traffic originating “outside” the cluster, this is not a requirement.

It defines a request for a specific load balancer config that implements the GatewayClass’ configuration and behaviour contract.
The resource MAY be created by an operator directly, or MAY be created by a controller handling a GatewayClass.

As the Gateway spec captures user intent, it may not contain a complete specification for all attributes in the spec.
For example, the user may omit fields such as addresses, ports, TLS settings.
This allows the controller managing the GatewayClass to provide these settings for the user, resulting in a more portable spec.
This behaviour will be made clear using the GatewayClass Status object.

A Gateway MAY contain one or more *Route references which serve to direct traffic for a subset of traffic to a specific service.

### {HTTP,TCP,Foo}Route

Route objects define protocol-specific rules for mapping requests from a Gateway to Kubernetes Services.

`HTTPRoute` and `TCPRoute` are currently the only defined Route objects. Additional protocol-specific Route
objects may be added in the future.

### BackendPolicy

BackendPolicy provides a way to configure connections between a Gateway and a
backend. For the purpose of this API, a backend is any resource that a route can
forward traffic to. A common example of a backend is a Service. Configuration at
this level is currently limited to TLS, but will expand in the future to support
more advanced policies such as health checking.

Some backend configuration may vary depending on the Route that is targeting the
backend. In those cases, configuration fields will be placed on Routes and not
BackendPolicy. For more information on what may be configured with this resource
in the future, refer to the related [GitHub
issue](https://github.com/kubernetes-sigs/service-apis/issues/196).

### Combined types

The combination of `GatewayClass`, `Gateway`, `xRoute` and `Service`(s) will
define an implementable load-balancer. The diagram below illustrates the
relationships between the different resources:

<!-- source: https://docs.google.com/document/d/1BxYbDovMwnEqe8lj8JwHo8YxHAt3oC7ezhlFsG_tyag/edit#heading=h.8du598fded3c -->
![schema](schema-uml.svg)

## Request flow

A typical client/gateway API request flow for a gateway implemented using a reverse proxy is:

 1. A client makes a request to http://foo.example.com.
 2. DNS resolves the name to a `Gateway` address.
 3. The reverse proxy receives the request on a `Listener` and uses the
 [Host header](https://tools.ietf.org/html/rfc7230#section-5.4) to match an `HTTPRoute`.
 5. Optionally, the reverse proxy can perform request header and/or path matching based
 on `match` rules of the `HTTPRoute`.
 6. Optionally, the reverse proxy can modify the request, i.e. add/remove headers, based
 on `filter` rules of the `HTTPRoute`.
 7. Lastly, the reverse proxy forwards the request to one or more objects, i.e. `Service`,
 in the cluster based on `forwardTo` rules of the `HTTPRoute`.

## TLS Configuration

TLS is configured on Gateway listeners. Additionally, TLS certificates
can be configured on route objects for certain self-service use cases.

Please refer to [TLS details](tls.md) for a deep dive on TLS.

## Design considerations

There are some general design guidelines used throughout this API.

### Single resource consistency

The Kubernetes API guarantees consistency only on a single resource level. There
are a couple of consequences for complex resource graphs as opposed to single
resources:

*   Error checking of properties spanning multiple resource will be asynchronous
    and eventually consistent. Simple syntax checks will be possible at the
    single resource level, but cross resource dependencies will need to be
    handled by the controller.
*   Controllers will need to handle broken links between resources and/or
    mismatched configuration.

### Conflicts

Separation and delegation of responsibility among independent actors (e.g
between cluster ops and application developers) can result in conflicts in the
configuration. For example, two application teams may inadvertently submit
configuration for the same HTTP path.

In most cases, guidance for conflict resolution is provided along with the
documentation for fields that may have a conflict. If a conflict does not have a
prescribed resolution, the following guiding principles should be applied:

* Prefer not to break things that are working.
* Drop as little traffic as possible.
* Provide a consistent experience when conflicts occur.
* Make it clear which path has been chosen when a conflict has been identified.
  Where possible, this should be communicated by setting appropriate status
  conditions on relevant resources.
* More specific matches should be given precedence over less specific ones.
* The resource with the oldest creation timestamp wins.
* If everything else is equivalent (including creation timestamp), precedences
  should be given to the resource appearing first in alphabetical order
  (namespace/name). For example, foo/bar would be given precedence over foo/baz.

### Conformance

As this API aims to cover a wide set of implementations and use cases,
it will not be possible for all implementations to support *all*
features at the present. However, we do expect the set of features
supported to converge eventually. For a given feature, users will be
guaranteed that features in the API will be portable between providers
if the feature is supported.

To model this in the API, we are taking a similar approach as with
[sig-arch][sig-arch-bdd] work on conformance profiles. Features as
described in the API spec will be divided into three major categories:

[sig-arch-bdd]: https://github.com/kubernetes/enhancements/tree/master/keps/sig-architecture/960-conformance-behaviors

* **CORE** features will be portable and we expect that there is a
  reasonable roadmap for ALL implementations towards support of APIs
  in this category.
* **EXTENDED** features are those that are portable but not
  universally supported across implementations. Those implementations
  that support the feature will have the same behavior and
  semantics. It is expected that some number of EXTENDED features will
  eventually migrate into the CORE. EXTENDED features will be part of
  the API types and schema.
* **CUSTOM** features are those that are not portable and are
  vendor-specific. CUSTOM features will not have API types and schema
  except via generic extension points.

Behavior and feature in the CORE and EXTENDED set will be defined and
validated via behavior-driven conformance tests. CUSTOM features will
not be covered by conformance tests.

By including and standardizing EXTENDED features in the API spec, we
expect to be able to converge on portable subsets of the API among
implementations without compromising overall API support. Lack of
universal support will not be a blocker towards developing portable
feature sets. Standardizing on spec will make it easier to eventually
graduate to CORE when support is widespread.

#### Conformance expectations

We expect there will be varying levels of conformance among the
different providers in the early days of this API. Users can use the
results of the conformance tests to understand areas where there may
be differences in behavior from the spec.

### Extension points

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
- **HTTPRouteFilter**: This API type in HTTPoute provides a way to hook into the
  request/response lifecycle of an HTTP request.
- **Custom Routes**: If none of the above extensions points suffice for a use
  case, Implementers can chose to create custom Route resources for protocols
  that are not currently supported in the API.

Whenever you are using an extension point without any prior art, please let
the community know. As we learn more about usage of extension points, we would
like to find the common denominators and promote the features to core/extended
API conformance.

