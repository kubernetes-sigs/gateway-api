# Cross namespace references and ReferencePolicy

In the Gateway API, it is possible to have references between objects cross namespace boundaries.
In particular, it's expected that Gateways and Routes may exist in different namespaces,
or that Services may be referred to by Routes in a another namespace.

However, this significantly violates the idea of a namespace as the edge of a trust domain.
In order to bring cross-namespace references under the control of the owner of the referent object's namespace,
the Gateway API has a ReferencePolicy object that must be created in the referent namespace for the reference to be successful.

To put this another way, if an object is referred to from outside its namespace,
the object's owner must create a ReferencePolicy object that describes how that reference is allowed.
This page explains how this process works.

In the past, we've seen that forwarding traffic across namespace boundaries is a desired feature,
but without the kinds of safeguards proposed here, [vulnerabilities](https://github.com/kubernetes/kubernetes/issues/103675)
can emerge.

## ReferencePolicy

To ensure that Gateway API is able to safely provide this functionality,
we need to enforce a handshake mechanism that requires resources in both namespaces to agree to this reference.
To accomplish that, a ReferencePolicy resource has been be introduced.

![Reference Policy](images/referencepolicy.png)

With this model, Routes are able to directly reference Routes and Services in other namespaces.
These references are only considered valid if a ReferencePolicy in the target namespace explicitly allows it.

The following example shows how a HTTPRoute in namespace `foo` can reference a Service in namespace `bar`.
In this example a ReferencePolicy in the `bar` namespace explicitly allows references to Services from HTTPRoutes in the `foo` namespace.

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
  - group: core
    kind: Service
```

### API design decisions
This proposed API is fairly straightforward, but comes with a few notable decisions:

1. Each ReferencePolicy only supports a single From and To section.
   Additional trust relationships must be modeled with additional ReferencePolicy resources.
1. Resource names are intentionally excluded from this policy for simplicity and because they rarely provide any meaningful protection.
   A user that is able to write to resources of a certain kind within a namespace can always rename
   resources or change the structure of the resources to match a given policy.
1. A single Namespace is allowed per "From" struct.
   Although a selector would be more powerful it may encourage unnecessarily insecure configuration.

```go
// ReferencePolicy identifies kinds of resources in other namespaces that are
// trusted to reference the specified kinds of resources in the local namespace.
// Each ReferencePolicy can be used to represent a unique trust relationship.
// Additional Reference Policies can be used to add to the set of trusted
// sources of inbound references for the namespace they are defined within.
type ReferencePolicy struct {
    metav1.TypeMeta   `json:",inline"`
    metav1.ObjectMeta `json:"metadata,omitempty"`

    // Spec defines the desired state of ReferencePolicy.
    Spec ReferencePolicySpec `json:"spec,omitempty"`
}


// ReferencePolicySpec identifies a cross namespace relationship that is trusted
// for Gateway API.
type ReferencePolicySpec struct {
    // From describes the trusted namespaces and kinds that can reference the
    // resources described in "To". Each entry in this list must be considered
    // to be an additional place that references can be valid from, or to put
    // this another way, entries must be combined using OR.
    //
    // Support: Core
    //
    // +kubebuilder:validation:MinItems=1
    From []ReferencePolicyFrom `json:"from"`

    // To describes the resources that may be referenced by the resources
    // described in "From". Each entry in this list must be considered to be an
    // additional place that references can be valid to, or to put this another
    // way, entries must be combined using OR.
    //
    // Support: Core
    //
    // +kubebuilder:validation:MinItems=1
    To []ReferencePolicyTo `json:"to"`
}

// ReferencePolicyFrom describes trusted namespaces and kinds.
type ReferencePolicyFrom struct {
    // Group is the group of the referent.
    //
    // Support: Core
    //
    // +kubebuilder:validation:MinLength=1
    // +kubebuilder:validation:MaxLength=253
    Group string `json:"group"`

    // Kind is the kind of the referent. Although implementations may support
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

    // Namespace is the namespace of the referent.
    //
    // Support: Core
    //
    // +kubebuilder:validation:MinLength=1
    // +kubebuilder:validation:MaxLength=253
    Namespace string `json:"namespace,omitempty"`
}

// ReferencePolicyTo describes what Kinds are allowed as targets of the
// references.
type ReferencePolicyTo struct {
    // Group is the group of the referent.
    //
    // Support: Core
    //
    // +kubebuilder:validation:MinLength=1
    // +kubebuilder:validation:MaxLength=253
    Group string `json:"group"`

    // Kind is the kind of the referent. Although implementations may support
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
* A single ReferencePolicy resource can be used for a namespace in place of
  separate handshake config on each Service or Route resource.

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

## ForwardTo

To enable cross-namespace forwarding, we have created a new `ObjectReference` struct that can be used in places
where cross-namespace references are possible, and have updated the HTTPRoute ForwardTo field `BackendRef` to use the new type.

```go

// Fron object_reference_types.go

// ObjectReference identifies an API object including its namespace.
type ObjectReference struct {
	// Group is the group of the referent.
	//
	// +kubebuilder:validation:MaxLength=253
	Group string `json:"group"`

	// Kind is kind of the referent.
	//
	// +kubebuilder:validation:MinLength=1
	// +kubebuilder:validation:MaxLength=253
	Kind string `json:"kind"`

	// Name is the name of the referent.
	//
	// +kubebuilder:validation:MinLength=1
	// +kubebuilder:validation:MaxLength=253
	Name string `json:"name"`

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


// From httproute_types.go

	// BackendRef is a reference to a backend to forward matched requests to. If
	// both BackendRef and ServiceName are specified, ServiceName will be given
	// precedence.
	//
	// If the referent cannot be found, the route must be dropped
	// from the Gateway. The controller should raise the "ResolvedRefs"
	// condition on the Gateway with the "DegradedRoutes" reason.
	// The gateway status for this route should be updated with a
	// condition that describes the error more specifically.
	//
	// If there is a cross-namespace reference to an *existing* object
	// with no ReferencePolicy, the controller must ensure the "ResolvedRefs"
	// condition on the Gateway is set to `status: true`, with the "RefNotPermitted"
	// reason.
	//
	// Support: Custom
	//
	// +optional
	BackendRef *ObjectReference `json:"backendRef,omitempty"`

    // The BackendRef in the more general RouteForwardTo object in shared_types.go
    // has also been updated.
```
### Conformance Level

ReferencePolicy support is a "CORE" coformance level for the following
objects:
- HTTPRoute
- TLSRoute
- TCPRoute
- UDPRoute

Other "ImplementationSpecific" objects and references are *strongly recommended* to also use this flow for cross-namespace references. 