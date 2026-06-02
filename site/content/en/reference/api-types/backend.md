---
title: "Backend"
weight: 3
---

{{< details title="Experimental Channel" color="purple" >}}
The `Backend` resource is part of the Experimental Channel.
For more information on release channels, refer to our
[versioning guide](/docs/concepts/versioning/).
{{< /details >}}

[Backend][backend] is a Gateway API type used to define a backend destination
and backend-specific connection behavior for Gateway clients.

## Background

A `Backend` provides a Gateway-native resource for describing where traffic
should go and how the Gateway should connect to that destination.

This is especially useful for:

- Defining external destinations without synthetic `ExternalName` Services.
- Setting backend connection protocol expectations.
- Defining backend TLS settings directly on the backend destination.

## Spec

The specification of a [Backend][backend] consists of:

- `Type` - Defines the backend type. Currently, `ExternalHostname` is supported.
- `Port` - Defines the destination port used for backend connections.
- `ExternalHostname` - Defines the external FQDN destination when `type` is
  `ExternalHostname`.
- `Protocol` - Defines the protocol the implementation should use when
  connecting to the backend.
- `TLS` - Defines TLS settings for backend connections.

## Example

The following example defines an external hostname backend and references it
from an `HTTPRoute`.

```yaml
apiVersion: gateway.networking.x-k8s.io/v1alpha1
kind: XBackend
metadata:
  name: openai-api
spec:
  type: ExternalHostname
  port:
    port: 443
  externalHostname:
    hostname: api.openai.com
  tls:
    mode: ServerOnly
    validation:
      hostname: api.openai.com
```

```yaml
apiVersion: gateway.networking.k8s.io/v1
kind: HTTPRoute
metadata:
  name: external-api-route
spec:
  parentRefs:
  - name: my-gateway
  rules:
  - backendRefs:
    - group: gateway.networking.x-k8s.io
      kind: XBackend
      name: openai-api
      port: 443
```

## Status

`Backend` status is reported through ancestor conditions in `status.ancestors`.
These conditions indicate whether parent resources have accepted and programmed
this backend destination.

[backend]: /reference/api-spec/main/specx/#xbackend
