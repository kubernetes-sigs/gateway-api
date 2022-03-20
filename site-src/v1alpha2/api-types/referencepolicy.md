# ReferencePolicy

A ReferencePolicy can be used to enable cross namespace references within
Gateway API. In particular, Routes may forward traffic to backends in other
namespaces, or Gateways may refer to Secrets in another namespace.

![Reference Policy](/v1alpha2/images/referencepolicy-simple.png)

In the past, we've seen that forwarding traffic across namespace boundaries is a
desired feature, but without a safeguard like ReferencePolicy,
[vulnerabilities](https://github.com/kubernetes/kubernetes/issues/103675) can
emerge.

If an object is referred to from outside its namespace, the object's owner must
create a ReferencePolicy resource to explicitly allow that reference. Without a
ReferencePolicy, a cross namespace reference is invalid.

## Structure
Fundamentally a ReferencePolicy is made up of two lists, a list of resources
references may come from, and a list of resources that may be referenced.

The `from` list allows you to specify the group, kind, and namespace of
resources that may reference items described in the `to` list.

The `to` list allows you to specify the group and kind of resources that may be
referenced by items described in the `from` list. The namespace is not necessary
in the `to` list because a ReferencePolicy can only be used to allow references
to resources in the same namespace as the ReferencePolicy.

## Example
The following example shows how a HTTPRoute in namespace `foo` can reference a
Service in namespace `bar`. In this example a ReferencePolicy in the `bar`
namespace explicitly allows references to Services from HTTPRoutes in the `foo`
namespace.

```yaml
kind: HTTPRoute
metadata:
  name: foo
  namespace: foo
spec:
  rules:
  - matches:
    - path: /bar
    forwardTo:
      backend:
      - name: bar
        namespace: bar
---
kind: ReferencePolicy
metadata:
  name: bar
  namespace: bar
spec:
  from:
  - group: networking.gateway.k8s.io
    kind: HTTPRoute
    namespace: foo
  to:
  - group: ""
    kind: Service
```

## API design decisions
While the API is simplistic in nature, it comes with a few notable decisions:

1. Each ReferencePolicy only supports a single From and To section. Additional
   trust relationships must be modeled with additional ReferencePolicy
   resources.
1. Resource names are intentionally excluded from this policy for simplicity and
   because they rarely provide any meaningful protection. A user that is able to
   write to resources of a certain kind within a namespace can always rename
   resources or change the structure of the resources to match a given policy.
1. A single Namespace is allowed per "From" struct. Although a selector would be
   more powerful, it encourages unnecessarily insecure configuration.
1. The effect of these resources is purely additive, they stack on top of each
   other. This makes it impossible for them to conflict with each other.

Please see the [API
Specification](/v1alpha2/references/spec#gateway.networking.k8s.io/v1alpha2.ReferencePolicy)
for more details on how specific ReferencePolicy fields are interpreted.

## Exceptions
Cross namespace Route -> Gateway binding follows a slightly different pattern
where the handshake mechanism is built into the Gateway resource. For more
information on that approach, refer to the relevant [Security Model
documentation](/concepts/security-model). Although conceptually similar to
ReferencePolicy, this configuration is built directly into Gateway Listeners,
and allows for fine-grained per Listener configuration that would not be
possible with ReferencePolicy.

There are some situations where it MAY be acceptable to ignore ReferencePolicy
in favor of some other security mechanism. This MAY only be done if other
mechanisms like NetworkPolicy can effectively limit cross-namespace references
by the implementation.

An implementation choosing to make this exception MUST clearly document that
ReferencePolicy is not honored by their implementations and detail which
alternative safeguards are available. Note that this is unlikely to apply to
ingress implementations of the API and will not apply to all mesh
implementations.

For an example of the risks involved in cross-namespace references, refer to
[CVE-2021-25740](https://github.com/kubernetes/kubernetes/issues/103675).
Implementations of this API need to be very careful to avoid confused deputy
attacks. ReferencePolicy provides a safeguard for that. Exceptions MUST only be
made by implementations that are absolutely certain that other equally effective
safeguards are in place.

## Conformance Level
ReferencePolicy support is a "CORE" conformance level requirement for
cross-namespace references that originate from the following objects:

- HTTPRoute
- TLSRoute
- TCPRoute
- UDPRoute

That is, all implementations MUST use this flow for any cross namespace
references in any of the core xRoute types, except as noted in the Exceptions
section above.

Other "ImplementationSpecific" objects and references MUST also use this flow
for cross-namespace references, except as noted in the Exceptions section above.
