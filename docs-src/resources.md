# API Resources

## GatewayClass

`GatewayClass` ([source code][gatewayclass-src]) is cluster-scoped resource
defined by the infrastructure provider. This resource represents a class of
Gateways that can be instantiated.

[gatewayclass-src]: https://github.com/kubernetes-sigs/service-apis/blob/master/apis/v1alpha1/gatewayclass_types.go

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
    group: acme.io/v1alpha1
    kind: Config
    name: internet-gateway-config
---
kind: Config
apiVersion: acme.io/v1alpha1
metadata:
  name: internet-gateway-config
spec:
  ip-address-pool: internet-vips
  ...
```

Using a Custom Resource for `GatewayClass.spec.parametersRef` is encouraged
but implementations may resort to using a ConfigMap if needed.

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

### GatewayClass controller selection

The `GatewayClass.spec.controller` field is used to determine whether
or not a given `GatewayClass` is managed by the controller.
This format of this field is opaque and specific to a particular controller.
Which GatewayClass is selected by a given controller field depends on how
various controller(s) in the cluster interpret this field.

It is RECOMMENDED that controller authors/deployments make their
selection unique by using a domain / path combination under their
administrative control (e.g. controller managing of all `controller`s
starting with `acme.io` is the owner of the `acme.io` domain) to avoid
conflicts.

Controller versioning can be done by encoding the version of a
controller into the path portion. An example scheme could be (similar
to container URIs):

```text
acme.io/gateway/v1   // Use version 1
acme.io/gateway/v2   // Use version 2
acme.io/gateway      // Use the default version
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

<!---
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
--->
