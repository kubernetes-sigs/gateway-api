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
- Defining internal (in-cluster) destinations without directly configuring a
  `Service`.
- Setting backend connection protocol expectations.
- Defining backend TLS settings directly on the backend destination.

## Internal (in-cluster) destinations

`Backend` uses `spec.endpointSelector` for defining internal destinations. To
use it, set `spec.type` to `EndpointSelector` and provide the labels of the
workload Pods in `spec.endpointSelector.matchLabels`. Currently,
`matchLabels` is the only supported endpoint selection mechanism.

An `EndpointSelector` Backend decorates internal endpoints with Gateway-specific
configuration, such as TLS and session persistence config. It is an additional
way to represent an internal destination, not a replacement for `Service`. Routes
can continue to target `Service` backends.

Users configure only the `Backend` for the internal destinations. There is no
need to create or manage a separate `Service` or endpoint resource. While the end
goal is to use [EndpointSelector](https://github.com/kubernetes/enhancements/issues/6116),
it is still being developed. Until then, the controller may use a `Service` behind
the scenes to resolve endpoints.

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

- `Type` - Defines the backend type: `EndpointSelector` for internal
  destinations or `ExternalHostname` for external destinations.
- `Port` - Defines the destination port used for backend connections.
- `EndpointSelector` - Selects workload Pods by label when `type` is
  `EndpointSelector`.
- `ExternalHostname` - Defines the external FQDN destination when `type` is
  `ExternalHostname`.
- `SessionPersistence` - Defines session persistence across endpoints when
  `type` is `EndpointSelector`.
- `Protocol` - Defines the protocol the implementation should use when
  connecting to the backend.
- `TLS` - Defines TLS settings for backend connections.

## Examples

### External Destinations

The following example defines an external hostname backend.

{{< readfile file="/examples/experimental/xbackend/external.yaml" code="true" lang="yaml" >}}

### Internal (In-Cluster) Destinations

The following example defines an internal backend by selecting workload Pods
with the label `app: store`.

{{< readfile file="/examples/experimental/xbackend/internal.yaml" code="true" lang="yaml" >}}

## Status

`Backend` status is reported for each associated ancestor in
`status.ancestors`. The `Accepted` condition indicates whether the Backend has
been accepted by the controller, while `ResolvedRefs` indicates whether all
references in the Backend have been resolved.

[backend]: /reference/api-spec/main/specx/#xbackend
