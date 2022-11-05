# GEP-709: Cross Namespace References from Routes

* Issue: [#709](https://github.com/kubernetes-sigs/gateway-api/issues/709)
* Status: Standard

!!! note
    This resource was originally named "ReferencePolicy". It was renamed
    to "ReferenceGrant" to avoid any confusion with policy attachment.

## TLDR

This GEP attempts to enable cross namespace forwarding from Routes and provide a
way to simplify adding Route inclusion (Routes including other Routes) in the
future. These are closely related concepts that can be solved with a new
ReferenceGrant resource that enables app admins to describe where they trust
references from.

## Motivation/User Journeys/Background

This GEP keeps same namespace references simple while enabling the following
capabilities for cross namespace references:

1. Retaining full control of Gateway and Routes in an infra namespace, while
   targeting apps in different namespaces.
1. Traffic splitting between Services in different namespaces.
1. Mesh overrides to target Services in different namespaces. (For more info,
   see GEP [#713](https://github.com/kubernetes-sigs/gateway-api/issues/713))

## ReferenceGrant

Anytime we allow crossing a namespace boundary, we need to be very cautious.
In the past, we've seen that forwarding traffic across namespace boundaries is
a desired feature, but without the kinds of safeguards proposed here,
[vulnerabilities](https://github.com/kubernetes/kubernetes/issues/103675) can
emerge.

To ensure that Gateway API is able to safely provide this functionality, we need
to enforce a handshake mechanism that requires resources in both namespaces to
agree to this reference. To accomplish that, a new ReferenceGrant resource
should be introduced.

![Reference Policy](images/709-referencegrant.png)

With this model, Routes would be able to directly reference Routes and Services
in other namespaces. These references would only be considered valid if a
ReferenceGrant in the target namespace explicitly allowed it.

The following example shows how a HTTPRoute in namespace foo could reference
a Service in namespace bar. In this example a ReferenceGrant in the bar
namespace explicitly allows references to Services from HTTPRoutes in the foo
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
kind: ReferenceGrant
metadata:
  name: bar
  namespace: bar
spec:
  from:
  - group: networking.gateway.k8s.io
    kind: HTTPRoute
    namespace: foo
  to:
  - group: core
    kind: Service
```

### API
This proposed API is fairly straightforward, but comes with a few notable
decisions:

1. Each ReferenceGrant only supports a single From and To section. Additional
   trust relationships can be modeled with additional ReferenceGrant resources.
1. Resource names are intentionally excluded from this policy for simplicity and
   because they rarely provide any meaningful protection. A user that is able
   to write to resources of a certain kind within a namespace can always rename
   resources or change the structure of the resources to match a given policy.
1. A single Namespace is allowed per "From" struct. Although a selector would be
   more powerful it may encourage unnecessarily insecure configuration.

```go
// ReferenceGrant identifies kinds of resources in other namespaces that are
// trusted to reference the specified kinds of resources in the local namespace.
// Each ReferenceGrant can be used to represent a unique trust relationship.
// Additional ReferenceGrants can be used to add to the set of trusted
// sources of inbound references for the namespace they are defined within.
type ReferenceGrant struct {
    metav1.TypeMeta   `json:",inline"`
    metav1.ObjectMeta `json:"metadata,omitempty"`

    // Spec defines the desired state of ReferenceGrant.
    Spec ReferenceGrantSpec `json:"spec,omitempty"`
}


// ReferenceGrantSpec identifies a cross namespace relationship that is trusted
// for Gateway API.
type ReferenceGrantSpec struct {
    // From describes the trusted namespaces and kinds that can reference the
    // resources described in "To". Each entry in this list must be considered
    // to be an additional place that references can be valid from, or to put
    // this another way, entries must be combined using OR.
    //
    // Support: Core
    //
    // +kubebuilder:validation:MinItems=1
    From []ReferenceGrantFrom `json:"from"`

    // To describes the resources that may be referenced by the resources
    // described in "From". Each entry in this list must be considered to be an
    // additional place that references can be valid to, or to put this another
    // way, entries must be combined using OR.
    //
    // Support: Core
    //
    // +kubebuilder:validation:MinItems=1
    To []ReferenceGrantTo `json:"to"`
}

// ReferenceGrantFrom describes trusted namespaces and kinds.
type ReferenceGrantFrom struct {
    // Group is the group of the referrent.
    //
    // Support: Core
    //
    // +kubebuilder:validation:MinLength=1
    // +kubebuilder:validation:MaxLength=253
    Group string `json:"group"`

    // Kind is the kind of the referrent. Although implementations may support
    // additional resources, the following Route types are part of the "Core"
    // support level for this field:
    //
    // * HTTPRoute
    // * TCPRoute
    // * TLSRoute
    // * UDPRoute
    //
    // +kubebuilder:validation:MinLength=1
    // +kubebuilder:validation:MaxLength=253
    Kind string `json:"kind"`

    // Namespace is the namespace of the referrent.
    //
    // Support: Core
    //
    // +kubebuilder:validation:MinLength=1
    // +kubebuilder:validation:MaxLength=253
    Namespace string `json:"namespace,omitempty"`
}

// ReferenceGrantTo describes what Kinds are allowed as targets of the
// references.
type ReferenceGrantTo struct {
    // Group is the group of the referrent.
    //
    // Support: Core
    //
    // +kubebuilder:validation:MinLength=1
    // +kubebuilder:validation:MaxLength=253
    Group string `json:"group"`

    // Kind is the kind of the referrent. Although implementations may support
    // additional resources, the following types are part of the "Core"
    // support level for this field:
    //
    // * Service
    // * HTTPRoute
    // * TCPRoute
    // * TLSRoute
    // * UDPRoute
    //
    // +kubebuilder:validation:MinLength=1
    // +kubebuilder:validation:MaxLength=253
    Kind string `json:"kind"`
}
```


### Benefits

* Conceptually similar to NetworkPolicy.
* A separate resource enables admins to restrict who can allow cross namespace
  references.
* Provides consistent way to control references to any resource from a Route.
* Can be extended in the future for additional use cases.
* A single ReferenceGrant resource can be used for a namespace in place of
  separate handshake config on each Service or Route resource.

#### Exceptions
There are some situations where it MAY be acceptable to ignore ReferenceGrant
in favor of some other security mechanism. This MAY only be done if other
mechanisms like NetworkPolicy can effectively limit cross-namespace references
by the implementation.

An implementation choosing to make this exception MUST clearly document that
ReferenceGrant is not honored by their implementations and detail which
alternative safeguards are available. Note that this is unlikely to apply to
ingress implementations of the API and will not apply to all mesh
implementations.

For an example of the risks involved in cross-namespace references, refer to
[CVE-2021-25740](https://github.com/kubernetes/kubernetes/issues/103675).
Implementations of this API need to be very careful to avoid confused deputy
attacks. ReferenceGrant provides a safeguard for that. Exceptions MUST only
be made by implementations that are absolutely certain that other equally
effective safeguards are in place.

## ForwardTo

To enable cross-namespace forwarding, we'll need to add an optional `namespace`
field to the ForwardTo BackendRef struct.

```go
type BackendRef struct {
    // ...

    // Namespace is the namespace of the backend. When unspecified, the local
    // namespace is inferred.
    //
    // Support: Core
    //
    // +kubebuilder:validation:MinLength=1
    // +kubebuilder:validation:MaxLength=253
    // +optional
    Namespace *string `json:"namespace,omitempty"`
}
```

## Alternatives

### Inline Config
Instead of ReferenceGrant, it is possible to represent these relationships
inline.
![Inline](images/709-inline.png)

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
kind: Service
metadata:
  name: baz
  namespace: baz
  annotations:
    gateway.networking.k8s.io/accept-forwarding-from: bar
```

Although this requires less YAML for the simple case, it is less flexible.
Annotations have real limitations and don't provide any room for RBAC
differentiation. Although it's possible that we could eventually add a proper
field to the Service API to represent this, it would be impossible to add this
concept to all potential backend types.

## Out of scope

* Although closely related, this GEP does not attempt to improve the
  Gateway->Route relationship. That will instead be covered by a future GEP.
* Although this GEP explores how ReferenceGrant could enable Route inclusion,
  the details of that feature will be left for a future GEP.

## References

**GitHub Issues:**

* [#411: Clarify how RouteGateways would work if we supported Route->Route
  delegation](https://github.com/kubernetes-sigs/gateway-api/issues/411)
* [#582: Allow cross namespace
  references](https://github.com/kubernetes-sigs/gateway-api/issues/582)
* [#634: Request Filtering Between Gateways and Namespaced
  Routes](https://github.com/kubernetes-sigs/gateway-api/issues/634)

**Docs:**

* [Gateway API Reference
  Policy](https://docs.google.com/document/d/18MoabVA-fr5XL9cYdf6cxclqRwFpOvHUXV_UYzSiooY/edit)
* [Selection Policy
  Proposal](https://docs.google.com/document/d/1S9t4YiDBwe1X7q915zKO0meZ8O_UPa8bzBLWBY8_XdM/edit?usp=sharing)
* [Route Inclusion
  Proposal](https://docs.google.com/document/d/1-0mgRRAY784OgGQ1_LCOshpLLbeAtIr4eXd0YVYK4RY/edit#heading=h.8cfxzle5tmqb)
* [Cross Namespace Forwarding
  Proposal](https://docs.google.com/document/d/1_B1G9JcNw3skNYLtdK7lTTzOeyz5w2hpa84cKA_MGKk/edit)
