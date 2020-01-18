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
content of this document is taken from the [API sketch][api-sketch].

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
  Examples include: the cloud provider (AWS, Azure, GCP, ...), the PaaS provider
  in a company.
* **Cluster operator**: The cluster operator (ops) is responsible for
  administration of entire clusters. They manage policies, network access,
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

## Resource model

> Note: Resource will initially live in the `networking.x-k8s.io` API group as
> Custom Resource Definitions (CRDs). Unqualified resource names will implicitly
> be assumed to be part of this API group.

Our resource model is based around a separation of concerns for a service
producer. Each resource is intended to be (mostly) independently-evolvable and
self-consistent:

* A way to group pods (backends) into a set via label selection: Kubernetes
  `core.Service`, independent of application-level routing.
* A way to describe application-level routing: `xxxxRoute`, e.g. `HTTPRoute`,
  `TCPRoute` independent of traffic access and consumption.
* A way to describe traffic access and consumption: `Gateway`, independent of
  implementation.
* A way to describe which implementations of traffic access are available:
  `GatewayClass`.

The combination of `GatewayClass`, `Gateway`, `xxxxRoute` and `Service`(s) will
define an implementable load-balancer. The diagram below illustrates the
relationships between the different resources:

<!-- source: https://docs.google.com/document/d/1BxYbDovMwnEqe8lj8JwHo8YxHAt3oC7ezhlFsG_tyag/edit#heading=h.8du598fded3c -->
![schema](schema-uml.svg)

### Design considerations

There are some general design guidelines used throughout this API.

#### Single resource consistency

The Kubernetes API guarantees consistency only on a single resource level. There
are a couple of consequences for complex resource graphs as opposed to single
resources:

*   Error checking of properties spanning multiple resource will be asynchronous
    and eventually consistent. Simple syntax checks will be possible at the
    single resource level, but cross resource dependencies will need to be
    handled by the controller.
*   Controllers will need to handle broken links between resources and/or
    mismatched configuration.

#### Conflicts

Separation and delegation of responsibility among independent actors (e.g
between cluster ops and application developers) can result in conflicts in the
configuration. For example, two application teams may inadvertently submit
configuration for the same HTTP path. There are several different strategies for
handling this:

* TODO

#### Extensibility

TODO

### GatewayClass

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

#### GatewayClass parameters

Providers of the `Gateway API` may need to pass parameters to their controller
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

#### GatewayClass status

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

### Gateway

A `Gateway` is 1:1 with the life cycle of the configuration of
infrastructure. When a user creates a `Gateway`, a load balancer is provisioned
(see below for details) by the `GatewayClass` controller. `Gateway` is the
resource that triggers actions in this API. Other resources in this API are
configuration snippets until a Gateway has been created to link the resources
together.

The `Gateway` spec defines the following:

*   `GatewayClass` used to instantiate this Gateway.
*   Listener bindings, which define addresses and ports, protocol termination,
    TLS-settings. Listener configuration requested by a Gateway definition can
    be incompatible with a given `GatewayClass` (e.g. port/protocol combination
    is not supported)

Listener configuration requested by a Gateway definition can be incompatible
with a given GatewayClass (e.g. port/protocol combination is not supported). In
this case, the Gateway will be in an error state, signaled by the status field.
Routes, which point to a set of protocol-specific routing served by the Gateway.
A Gateway can point directly to Kubernetes Service if no advanced routing is
required.

#### Deployment models

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

#### Gateway Status

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
