---
title: "Backend"
weight: 11
---

{{< details title="Experimental Channel" color="purple" >}}
The `Backend` resource is part of the Experimental Channel.
For more information on release channels, refer to our
[versioning guide](/docs/concepts/versioning/).
{{< /details >}}

[Backend][backend] is a Gateway API type used to define a backend destination
and backend-specific connection behavior for Gateways when they act as a client.

## Background

A `Backend` provides a Gateway-native resource for describing where traffic
should go and how the Gateway should connect to that destination.

This is especially useful for:

- Defining external destinations without synthetic `ExternalName` Services.
- Setting backend connection protocol expectations.
- Defining backend TLS settings directly on the backend destination.

## ExternalHostname vs ExternalName

`Backend` supports `ExternalHostname` for external FQDN destinations instead of
requiring users to create an additional `Service` with `type: ExternalName`.

These options are similar in that both ultimately depend on DNS, but
`ExternalHostname` gives Gateway API a clearer and safer model for egress
configuration.

One important reason is DNS rebinding risk. `ExternalName` has a known history
of DNS rebinding concerns (for example,
[CVE-2021-25740](https://github.com/kubernetes/kubernetes/issues/103675)),
where a hostname can resolve to unexpected addresses over time. To reduce this
risk, `ExternalHostname` validation in this API does not allow hostnames ending
in `.cluster.local`, which helps prevent references that look like in-cluster
service names.

Because DNS trust is still required for external destinations, implementations
should add additional guardrails such as egress domain allow-lists and pair
this with admission and network-level controls.

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
