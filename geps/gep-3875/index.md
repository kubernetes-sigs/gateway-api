# GEP-3875: BackendTLS consumer overrides and context awareness

* Issue: [#3875](https://github.com/kubernetes-sigs/gateway-api/issues/3875)
* Status: Provisional


## TLDR

This GEP aims to make the configuration of `BackendTLSPolicy` more flexible by allowing users to specify when the policy should be applied.
This extends the current state of the API, where this can only be controlled based on Service but the application developer persona.

## Goals

* Allow [service consumers](https://gateway-api.sigs.k8s.io/concepts/glossary/?h=gloss#consumer-route) to define their own TLS policies (the current state would be described as a [service producer policy](https://gateway-api.sigs.k8s.io/concepts/glossary/?h=gloss#producer-route)).
* Allow contextual selection of a TLS policy. For instance, allow usage of TLS from a specific Gateway or HTTPRoute but not others.

## Non-Goals

(What is out of scope for this proposal.)

## Introduction

(Can link to external doc -- but we should bias towards copying
the content into the GEP as online documents are easier to lose
-- e.g. owner messes up the permissions, accidental deletion)

## API

This proposal includes a few API changes:

1. A new `namespaces` field will be added to `spec.targetRefs`.
2. A new `from` field will be added to `spec.targetRefs`.
3. A new `mode` field will be added to `spec`.

### `TargetRef.Namespace`

Implementation target: Experimental in 1.4.

The `namespace` field allows a service consumer in another namespace to override the BackendTLSPolicy to use when connecting to a Service.

* `BackendTLSPolicy` can be applied to a `Service` in the same namespace to provide defaults to clients.
* `BackendTLSPolicy` can be applied to a `Service` in a different namespace to provide explicit configuration for clients in that namespace.

In a complex setup, there are up to 4 namespaces involved: the Gateway, ListenerSet, Route, and Service.
However, `BackendTLSPolicy` will only be consulted in the `Gateway` and `Service` namespace.
For mesh use cases, there are only 2 relevant namespaces: the calling client namespace and the `Service` namespace; both of these will be consulted.

If a `BackendTLSPolicy` exists for a `Service` in both namespaces, the consumer namespace will take precedence.

For example, given the following (condensed) configuration:

```yaml
kind: Gateway
namespace: gateway-ns
---
kind: BackendTLSPolicy
name: gateway-pol
namespace: gateway-ns
targetRef:
  name: svc
  namespace: app
---
kind: Service
name: svc
namespace: app
---
kind: BackendTLSPolicy
name: app-pol
namespace: gateway-ns
targetRef:
  name: svc
```

For requests from Gateways in the `gateway-ns` to `svc.app`, the `gateway-pol` will be used. For requests from Gateways in other namespaces, the `app-pol` will be used.

There will be no merging of policies. Exactly one policy will apply.
This mirrors the [producer and consumer route semantics](https://gateway-api.sigs.k8s.io/concepts/glossary/?h=gloss#producer-route).

This will be implemented by a new API type, `PolicyTargetReferenceWithSectionName`, as there is no existing policy that utilizes a cross-namespace reference.

### `TargetRef.From`

This field is expected to be implemented in a release after [`Namespace`](#targetrefnamespace).

[Tracking Issue](https://github.com/kubernetes-sigs/gateway-api/issues/3856)

While `Namespace` scoping enables some override use cases, the granularity is coarse.
For example, we cannot apply different policies when requests come from the same namespace.
It is (unfortunately) common to run many applications in a single namespace, or even to want different policies from a single workload.

This proposal adds a new `from` selector on the `targetRef`, which would allow referencing an object in the same namespace:

```go
type LocalFromReference struct {
  // Group is the group of the resource.
  Group Group `json:"group"`
  // Kind is kind of the resource.
  Kind Kind `json:"kind"`
  // Name is the name of the resource.
  Name ObjectName `json:"name"`
}
```

An example usage, to attach a BackendTLSPolicy for calls from Gateway `my-gateway` when calling the `app` Service in another namespace:

```yaml
kind: BackendTLSPolicy
metadata:
  name: app-pol
  namespace: gateway-ns
targetRef:
  namespace: app-namespace
  name: app
  from:
    kind: Gateway
    name: my-gateway
```

An example, doing the same but for a specific HTTPRoute:

```yaml
kind: BackendTLSPolicy
metadata:
  name: app-pol
  namespace: gateway-ns
targetRef:
  namespace: app-namespace
  name: app
  from:
    kind: HTTPRoute
    name: my-route
```

Including `SectionName` is also a likely addition that could be added for fine-grained control.

The following `from` types will be `Core`:

* Gateway
* ListenerSet
* All Route types where TLS is applicable (not UDPRoute)

Implementations may have implementation specific `from` types as well.

The inclusion of 'From' introduces a more complex hierarchy of policies.
In the event multiple policies could apply to a certain traffic path, the most precise one will win.
This means, in order of highest to lowest priority:

* Route
* ListenerSet
* Gateway
* `from` unset, in the consumer namespace
* `from` unset, in the producer namespace

In the event that multiple policies apply at the same level, [standard conflict resolution](/geps/gep-713#conflict-resolution) applies.

### `Spec.Mode`

With the ability to override `BackendTLSPolicy` based on context, users can modify the properties of the TLS handshake, such as changing validation parameters.
However, there is no way to *disable* the `BackendTLSPolicy` entirely.

The new `mode` field will enable this by allowing users to specify a `BackendTLSPolicy` that indicates "Do not send TLS".

The `BackendTLSMode` will be an enum with values: `None`, `TLS`.
The default for `Mode` will be `TLS`, to maintain backwards compatibility (and behavioral expectations).

When `None` is specified, the only permitted field is `targetRefs`; all other fields must be empty.

While this GEP only proposed these 2 fields, it is plausible the `Mode` could be extended for future use cases such as Mutual TLS.

## Conformance Details

Existing `BackendTLSPolicy` conformance tests apply policies to Services and test that the policy is used when reaching targeted services, and not used when reaching non-targeted Services.
This same approach will be used with these new matching criteria to ensure implementations are choosing the correct policy to apply.
In particular, scenarios where multiple applicable policies *could* apply will be tested to ensure proper precedence.

#### Feature Names

(Does it require separate feature(s) for mesh? Please add them if necessary)

Yes, the following features will be added:

1. BackendTLSOverride

## Alternatives


## References

Previous discussions:

* [BackendTLSPolicy's Service attachment is problematic](https://github.com/kubernetes-sigs/gateway-api/issues/3554)
* [Support a "from" field in targetRef for policy attachment](https://github.com/kubernetes-sigs/gateway-api/issues/3856)
