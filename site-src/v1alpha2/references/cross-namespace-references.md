# Cross namespace references and ReferencePolicy

## Introduction
In the Gateway API, it is possible to have references between objects cross
namespace boundaries. In particular, Services may be referred to by Routes
in a another namespace, or possibly Secrets may be referred to by Gateways or
Routes in another namespace.

In the past, we've seen that forwarding traffic across namespace boundaries is a
desired feature, but a safeguard like ReferencePolicy,
[vulnerabilities](https://github.com/kubernetes/kubernetes/issues/103675) can emerge.

!!! note
    When discussing the process of creating cross-namespace object references, this
    document and the documentation on the API itself talk about the object being
    referred to using the name "the referent object", using the
    [meaning](https://dictionary.cambridge.org/dictionary/english/referent)
    of "referent" to be "the person, thing, or idea that a word, phrase, or object
    refers to".

In order to bring cross-namespace references under the control
of the owner of the referent object's namespace, the Gateway API has a
ReferencePolicy object that must be created in the referent namespace for the
reference to be successful.

To put this another way, if an object is referred to from outside its namespace,
the object's owner must create a ReferencePolicy object that describes how that
reference is allowed. This page explains how this process works.

## ReferencePolicy

To ensure that Gateway API is able to safely provide this functionality,
we need to enforce a handshake mechanism that requires resources in both
namespaces to agree to this reference. To accomplish that, a ReferencePolicy
resource has been introduced.

![Reference Policy](/v1alpha2/references/images/referencepolicy.png)

With this model, Routes are able to directly reference Services in other namespaces.
These references are only considered valid if a ReferencePolicy in the target
namespace explicitly allows it.

The following example shows how a HTTPRoute in namespace `foo` can reference a
Service in namespace `bar`. In this example a ReferencePolicy in the `bar` namespace
explicitly allows references to Services from HTTPRoutes in the `foo` namespace.

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

### API design decisions
While the API is simplistic in nature, it comes with a few notable decisions:

1. Each ReferencePolicy only supports a single From and To section.
   Additional trust relationships must be modeled with additional ReferencePolicy
   resources.
1. Resource names are intentionally excluded from this policy for simplicity
   and because they rarely provide any meaningful protection. A user that is
   able to write to resources of a certain kind within a namespace can always
   rename resources or change the structure of the resources to match a given
   policy.
1. A single Namespace is allowed per "From" struct.
   Although a selector would be more powerful, it encourages unnecessarily
   insecure configuration.

Please see the [API Specification](/v1alpha2/references/spec#gateway.networking.k8s.io/v1alpha2.ReferencePolicy)
for the details of the object's behavior.

### Exceptions
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
attacks. ReferencePolicy provides a safeguard for that. Exceptions MUST only
be made by implementations that are absolutely certain that other equally
effective safeguards are in place.

### Conformance Level

ReferencePolicy support is a "CORE" conformance level requirement for
cross-namespace references that originate from the following objects:
- HTTPRoute
- TLSRoute
- TCPRoute
- UDPRoute

That is, all implemenations MUST use this flow for any cross namespaces
in any of the core xRoute types, except as noted in the Exceptions section
above.

Other "ImplementationSpecific" objects and references MUST also use this flow
for cross-namespace references, except as noted in the Exceptions section
above.