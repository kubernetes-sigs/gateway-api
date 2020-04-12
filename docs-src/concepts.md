<!--
Copyright 2019 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

     http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
-->

# API Concepts

This document is a deep dive into the reasoning and design for the API. The
content of this document was originally taken from the [API sketch][api-sketch].

> We will try to keep the two documents in sync as the sketch document has to
> lowest bar to contribution, but this document is easier to format well and
> review.

[api-sketch]: https://docs.google.com/document/d/1BxYbDovMwnEqe8lj8JwHo8YxHAt3oC7ezhlFsG_tyag/preview

## Roles and personas

In the original design of Kubernetes, the Ingress and Service resources were
based on a self-service model of usage; developers who create Services and
Ingresses control all aspects of defining and exposing their applications to
their users.

We have found that the self-service model does not fully capture some of the
more complex deployment and team structures that our users are seeing. The
Gateway/Routes API will target the following personas:

* **Infrastructure provider**: The infrastructure provider (infra) is
  responsible for the overall environment that the cluster(s) are operating in.
  Examples include public cloud providers (AWS, Azure, GCP, ...), or PaaS providers
  within an organization.
* **Cluster operator**: The cluster operator (ops) is responsible for
  administration of entire clusters. They manage policies, network access, and
  application permissions.
* **Application developer**: The application developer (dev) is responsible for
  defining their application configuration (e.g. timeouts, request
  matching/filter) and Service composition (e.g. path routing to backends).

We expect that each persona will map approximately to a `Role` in the Kubernetes
Role-Based Authentication (RBAC) system and will define resource model
responsibility and separation.

Depending on the environment, multiple roles can map to the same user.
For example, giving the user all of the above role replicates the self-service
model.

For more information on the roles and personas considered in the Service API
design, refer to the [Security Model](/security-model). 

## Resource model

> Note: Resource will initially live in the `networking.x-k8s.io` API group as
> Custom Resource Definitions (CRDs). Unqualified resource names will implicitly
> be assumed to be part of this API group.


There are three main types of object in our resource model:

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

Various types of Route objects define how traffic via the Gateway is mapped to Kubernetes Services.

Currently the two Route object types are `HTTPRoute` and `TCPRoute`, to cover the two most common uses for this sort of traffic.

This design is intended to be extensible at this point - it is possible that the service-apis team may create a UDPRoute, or even an IPRoute in the future.

### Combined types

The combination of `GatewayClass`, `Gateway`, `xxxxRoute` and `Service`(s) will
define an implementable load-balancer. The diagram below illustrates the
relationships between the different resources:

<!-- source: https://docs.google.com/document/d/1BxYbDovMwnEqe8lj8JwHo8YxHAt3oC7ezhlFsG_tyag/edit#heading=h.8du598fded3c -->
![schema](schema-uml.svg)

## Request flow

A typical client/gateway API request flow for a gateway implemented using a reverse proxy is:

 1. A client makes a request to a FQDN, i.e. "foo.example.com".
 2. The FQDN gets resolved to `gateway.status.listeners[x].address`.
 3. The request is received by the Gateway implementation, i.e. reverse proxy, on `gateway.status.listeners[x].address`
 and `gateway.spec.listeners[x].port`.
 4. If the request uses TLS, then `gateway.spec.listeners[x].tls` is used for establishing the connection.
 __Note:__ The details for modeling a "virtual host" is still [under review](https://github.com/kubernetes-sigs/service-apis/issues/49).
 Therefore, how TLS is configured may change in the future.
 5. If the `Gateway` is configured to terminate the TLS connection, an `HTTPRoute` is selected based on the request’s
 [Host header](https://tools.ietf.org/html/rfc7230#section-5.4), i.e. FQDN in step 1, matching
 `httpRoute.spec.hosts[x].hostname`. If the `Gateway` is configured to pass the TLS connection through to the backend
 object, i.e. Service, SNI is used to match the request with an `HTTPRoute` based on `httpRoute.spec.hosts[x].hostname`.
 __Note:__ Whether `hosts` should be singular is still [under review](https://github.com/kubernetes-sigs/service-apis/pull/43).
 6. The Gateway implementation performs filtering (optional) and forwarding based on
 `httpRoute.spec.hosts[x].rules[x].match` . The match can be based on the request path and/or header.
 7. Lastly, the request is forwarded to an object within the cluster.

## TLS Configuration

TLS configuration is tied to Gateway listeners. Although adding the option to
configure TLS on other resources was considered, ultimately TLS configuration on
Gateway listeners was deemed sufficient for the following reasons:

* In most cases, users that are configuring TLS will naturally also have access
  to Gateways.
* In other cases, TLS config could be implemented with a controller watching
  Routes and adding generated certs to corresponding Gateways.
* This does not solve the use case for users wanting to provide their own certs
  for Routes while not having access to a Gateway resource. This seems like it
  would be a rare edge case and is not worth supporting at this point. The
  security model outlined a potential approach to enable this in the future, but
  there does not seem to be a sufficient reason to work towards that now.

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
configuration for the same HTTP path. There are several different strategies for
handling this:

* TODO

### Extensibility

TODO

## GatewayClass

`GatewayClass` ([source code][gatewayclass-src]) is cluster-scoped resource
defined by the infrastructure provider. This resource represents a class of
Gateways that can be instantiated.

[gatewayclass-src]: https://github.com/kubernetes-sigs/service-apis/blob/master/api/v1alpha1/gatewayclass_types.go

> Note: this serves the same function as the [`networking.IngressClass` resource][ingress-class-api].

[ingress-class-api]: https://github.com/kubernetes/enhancements/blob/master/keps/sig-network/20190125-ingress-api-group.md#ingressclass-resource

```yaml
kind: GatewayClass
metadata:
  name: cluster-gateway
spec:
  controller: "acme.io/gateway-controller"
```

We expect that one or more `GatewayClasses` will be created by the
infrastructure provider for the user. It allows decoupling of which mechanism
(e.g. controller) implements the `Gateways` from the user. For instance, an
infrastructure provider may create two `GatewayClasses` named `internet` and
`private` to reflect `Gateways` that define Internet-facing vs private, internal
applications.

```yaml
kind: GatewayClass
metadata:
  name: internet
  ...
---
kind: GatewayClass
metadata:
  name: private
  ...
```

The user of the classes will not need to know *how* `internet` and `private` are
implemented. Instead, the user will only need to understand the resulting
properties of the class that the `Gateway` was created with.

### GatewayClass parameters

Providers of the `Gateway` API may need to pass parameters to their controller
as part of the class definition. This is done using the
`GatewayClass.spec.parametersRef` field:

```yaml
# GatewayClass for Gateways that define Internet-facing applications.
kind: GatewayClass
metadata:
  name: internet
spec:
  controller: "acme.io/gateway-controller"
  parametersRef:
    apiVersion: core/v1
    kind: ConfigMap
    namespace: acme-system
    name: internet-gateway
---
kind: ConfigMap
metadata:
  name: internet-gateway
  namespace: acme-system
data:
  ip-address-pool: internet-vips
  ...
```

The type of object referenced by `GatewayClass.spec.parametersRef` will depend
on the provider itself. A `core.ConfigMap` is used in the example above, but
controllers may opt to use a `CustomResource` for better schema validation.

### GatewayClass status

`GatewayClasses` MUST be validated by the provider to ensure that the configured
parameters are valid. The validity of the class will be signaled to the user via
`GatewayClass.status`:

```yaml
kind: GatewayClass
...
status:
  conditions:
  - type: InvalidParameters
    status: Unknown
    ...
```

A new `GatewayClass` will start with the `InvalidParameters` condition set to
`Unknown`. At this point the controller has not seen the configuration. Once the
controller has processed the configuration, the condition will be set to
`False`:

```yaml
kind: GatewayClass
...
status:
  conditions:
  - type: InvalidParameters
    status: False
    ...
```

If there is an error in the `GatewayClass.spec`, the conditions will be
non-empty and contain information about the error.

```yaml
kind: GatewayClass
...
status:
  conditions:
  - type: InvalidParameters
    status: True
    Reason: BadFooBar
    Message: "foobar" is an FooBar.
```

## Gateway

A `Gateway` is 1:1 with the life cycle of the configuration of
infrastructure. When a user creates a `Gateway`, some load balancing
infrastructure is provisioned or configured
(see below for details) by the `GatewayClass` controller. `Gateway` is the
resource that triggers actions in this API. Other resources in this API are
configuration snippets until a Gateway has been created to link the resources
together.

The `Gateway` spec defines the following:

*   The `GatewayClass` used to instantiate this Gateway.
*   The Listener bindings, which define addresses and ports, protocol termination,
    and TLS settings. The Listener configuration requested by a Gateway definition can
    be incompatible with a given `GatewayClass` (e.g. port/protocol combination
    is not supported).
*   The Routes, which describe how traffic is processed and forwarded.

If the Listener configuration requested by a Gateway definition is incompatible
with a given GatewayClass, the Gateway will be in an error state, signaled by the status field.

### Deployment models

Depending on the `GatewayClass`, the creation of the `Gateway` could do any of
the following actions:

* Use cloud APIs to create an LB instance.
* Spawn a new instance of a software LB (in this or another cluster).
* Add a configuration stanza to an already instantiated LB to handle the new
  routes.
* Program the SDN to implement the configuration.
* Something else we haven’t thought of yet...

The API does not specify which one of these actions will be taken. Note that a
GatewayClass controller that manages in-cluster proxy processes MAY restrict
Gateway configuration scope, e.g. only be served in the same namespace.

### Gateway Status

Gateways track status for the `Gateway` resource as a whole as well as each
`Listener` it contains. The status for a specific Route is reported in the
status of the `Route` resource. Within `GatewayStatus`, Listeners will have
status entries corresponding to their name. Both `GatewayStatus` and
`ListenerStatus` follow the conditions pattern used elsewhere in Kubernetes.
This is a list that includes a type of condition, the status of that condition,
and the last time this condition changed.

#### Listeners

TODO

### Routes

TODO

#### `HTTPRoute`

TODO

#### `TCPRoute`

TODO

#### Generic routing

TODO

#### Delegation/inclusion

TODO

### Destinations

TODO
