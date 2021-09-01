# GatewayClass

[GatewayClass][gatewayclass] is cluster-scoped resource defined by the
infrastructure provider. This resource represents a class of Gateways that can
be instantiated.

[gatewayclass]: /v1alpha1/references/spec/#networking.x-k8s.io/v1alpha1.GatewayClass

> Note: GatewayClass serves the same function as the
> [`networking.IngressClass` resource][ingress-class-api].

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
apiVersion: acme.io/v1alpha1
kind: Config
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
  - type: Admitted
    status: False
    ...
```

A new `GatewayClass` will start with the `Admitted` condition set to
`False`. At this point the controller has not seen the configuration. Once the
controller has processed the configuration, the condition will be set to
`True`:

```yaml
kind: GatewayClass
...
status:
  conditions:
  - type: Admitted
    status: True
    ...
```

If there is an error in the `GatewayClass.spec`, the conditions will be
non-empty and contain information about the error.

```yaml
kind: GatewayClass
...
status:
  conditions:
  - type: Admitted
    status: False
    Reason: BadFooBar
    Message: "foobar" is an FooBar.
```

### GatewayClass controller selection

The `GatewayClass.spec.controller` field determines the controller implementation
responsible for managing the `GatewayClass`. The format of the field is opaque
and specific to a particular controller. The GatewayClass selected by a given
controller field depends on how various controller(s) in the cluster interpret
this field.

It is RECOMMENDED that controller authors/deployments make their selection
unique by using a domain / path combination under their administrative control
(e.g. controller managing of all `controller`s starting with `acme.io` is the
owner of the `acme.io` domain) to avoid conflicts.

Controller versioning can be done by encoding the version of a controller into
the path portion. An example scheme could be (similar to container URIs):

```text
acme.io/gateway/v1   // Use version 1
acme.io/gateway/v2   // Use version 2
acme.io/gateway      // Use the default version
```

[ingress-class-api]: https://kubernetes.io/docs/concepts/services-networking/ingress/#ingress-class
