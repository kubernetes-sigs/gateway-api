# API Overview

This document provides an overview of Service APIs.

## Roles and personas.

There are 3 primary roles in Service APIs:

- Infrastructure Provider
- Cluster Operator
- Application Developer

There could be a fourth role of Application Admin in some use cases.

Please refer to the [roles and personas](security-model.md#roles-and-personas) 
section in the Security model for details.

## Resource model

> Note: Resources will initially live in the `networking.x-k8s.io` API group as
> Custom Resource Definitions (CRDs). Unqualified resource names will implicitly
> be in this API group.

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
[IngressClass](https://github.com/kubernetes/enhancements/blob/master/keps/sig-network/20190125-ingress-api-group.md#ingress-class)
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
fields such as addresses, ports, TLS settings. This allows the controller
managing the GatewayClass to provide these settings for the user, resulting in a
more portable spec. This behaviour will be made clear using the GatewayClass
Status object.

A Gateway MAY contain one or more *Route references which serve to direct
traffic for a subset of traffic to a specific service.*

### {HTTP,TCP,Foo}Route

Route objects define protocol-specific rules for mapping requests from a Gateway
to Kubernetes Services.

`HTTPRoute` and `TCPRoute` are currently the only defined Route objects.
Additional protocol-specific Route objects may be added in the future.

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

### Route binding

When a Route binds to a Gateway it represents configuration that is applied on
the Gateway that configures the underlying load balancer or proxy. How and which
Routes bind to Gateways is controlled by the resources themselves. Route and
Gateway resources have built-in controls to permit or constrain how they select
each other. This is useful for enforcing organizational policies for how Routes
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
![route binding](images/gateway-route-binding.png)

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

*In summary, Gateways select Routes and Routes control their exposure. When a
Gateway selects a Route that allows itself to be exposed, then the Route will
bind to the Gateway. When Routes are bound to a Gateway it means their
collective routing rules are configured on the underlying load balancers or
proxies that are managed by that Gateway. Thus, a Gateway is a logical
representation of a networking data plane that can be configured through
Routes.*

#### Route Selection

A Gateway selects routes based on the Route metadata, specifically the kind,
namespace, and labels of Route resources. Routes are actually bound to specific
listeners within the Gateway so each listener has a `listener.routes` field
which selects Routes by one or more of the following criterea:

- **Label** - A Gateway can select Routes via labels that exist on the 
resource (similar to how Services select Pods via Pod labels).
- **Kind** - A Gateway listener can only select a single type of Route 
resource. This could be an HTTPRoute, TCPRoute, or a custom Route type.
- **Namespace** - A Gateway can also control from which Namespaces Routes can be
selected via the `namespaces.from` field. It supports three possible values:
    - `SameNamespace` is the default option. Only Routes in the same namespace 
      as this Gateway will be selected. 
    - `All` will select Routes from all Namespaces.
    - `Selector` means that Routes from a subset of Namespaces selected by a 
      Namespace label selector will be selected by this Gateway. When `Selector` 
      is used then the `listeners.routes.namespaces.selector` field can be used 
      to specify label selectors. This field is not supported with `All` or 
      `SameNamespace`.

The below Gateway will select all HTTPRoute resources with the `expose:
prod-web-gw` across all Namespaces in the cluster.

```
kind: Gateway
...
spec:
  listeners:  
  - routes:
      kind: HTTPRoute
      selector:
        matchLabels:
          expose: prod-web-gw 
      namespaces:
        from: All
```

#### Route Exposure

Routes can determine how they are exposed through Gateways. The `gateways.allow`
field supports three values:

- `All` is the default value if none is specified. This leaves all binding 
to the Route label and Namespace selectors on the Gateway. 
- `SameNamespace` only allows this Route to bind with Gateways from the 
same Namespace.
- `FromList` allows an explicit list of Gateways to be specifiied that a 
Route will bind with. `gateways.gatewayRefs` is only supported with this option. 

The following `my-route` Route selects only the `foo-gateway` in the
`foo-namespace` and will not be able to bind with any other Gateways. Note that
`foo-gateway` is in a different Namespace. If the `foo-gateway` allows
cross-Namespace binding and also selects this Route then `my-route` will bind to
it. 

```yaml
kind: HTTPRoute
metadata:
  name: my-route
  namespace: bar-namespace
spec:
  gateways:
    allow: FromList
    gatewayRefs:
    - name: foo-gateway
      namespace: foo-namespace
```

Note that Gateway and Route binding is bi-directional. This means that both
resources must select each other for them to bind. If a Gateway has Route label
selectors that do not match any existing Route then nothing will bind to it even
if a Route's `spec.gateways.allow = All`. Similarly, if a Route references a
specific Gateway, but that Gateway is not selecting the Route's Namespace, then
they will not bind. A binding will only take place if both resources select each
other.

It may not always be apparent from the resource specifications which Gateways
and Routes are bound, but binding can be determined from the resource status.
The [Route status](httproute.md#routestatus) will list all of the Gateways that
a Route is bound to and any relevant conditions for the binding.

### Combined types

The combination of `GatewayClass`, `Gateway`, `xRoute` and `Service`(s) will
define an implementable load-balancer. The diagram below illustrates the
relationships between the different resources:

<!-- source: https://docs.google.com/document/d/1BxYbDovMwnEqe8lj8JwHo8YxHAt3oC7ezhlFsG_tyag/edit#heading=h.8du598fded3c -->
![schema](images/schema-uml.svg)

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

TLS is configured on Gateway listeners. Additionally, TLS certificates
can be configured on route objects for certain self-service use cases.

Please refer to [TLS details](tls.md) for a deep dive on TLS.


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

