# GEP-1867: Per-Gateway Infrastructure

* Status: Standard
* Issue: [#1867](https://github.com/kubernetes-sigs/gateway-api/issues/1867)

## Overview

`Gateway`s represent a piece of infrastructure implemented by cloud load balancers, in-cluster deployments, or other mechanisms.
These often need vendor-specific configuration outside the scope of existing APIs (e.g. "size" or "version" of the infrastructure to provision).

Today `GatewayClass.spec.parametersRef` is available to attach arbitrary configuration to a `GatewayClass`.

This GEP will explain why that is not sufficient to meet common use cases, and introduce a new field - `infrastructure` - to address these cases.

Related discussions:
* [Support cluster-local Gateways](https://github.com/kubernetes-sigs/gateway-api/discussions/1247)
* [Scaling Gateway Resources](https://github.com/kubernetes-sigs/gateway-api/discussions/1355)
* [Manual deployments](https://github.com/kubernetes-sigs/gateway-api/issues/1687)
* [Merging Gateways](https://github.com/kubernetes-sigs/gateway-api/pull/1863)
* [In Cluster Gateway Deployments](https://github.com/kubernetes-sigs/gateway-api/pull/1757)

## Goals

* Provide the ability to configure arbitrary (implementation specific) attributes about a **specific Gateway**.
* Provide the ability to configure a standardized set of attributes about a **specific Gateway**.

## Why not GatewayClass parameters?

`GatewayClass.spec.parametersRef` is the existing mechanism to configure arbitrary fields on a Gateway.
However, this introduces operational challenges when configuring Gateways.

### Scope

As a `Gateway` manager (with RBAC permissions to a specific `Gateway`) I should be able to declaratively make changes to that `Gateway` without the need for access to cluster-scoped resources (`GatewayClass`) and without affecting other `Gateways` managed by the same `GatewayClass`.
This has been previously discussed in [this issue](https://github.com/kubernetes-sigs/gateway-api/issues/567).

As a cluster scoped resource, `GatewayClass` does not meet this requirement.
This restricts customization use cases to either a few pre-provisioned classes by the admin, or running in an environment where the "Infrastructure Provider" and "Cluster Operator" are the same roles.
The distinction between these roles is explicitly called out on the [homepage](https://gateway-api.sigs.k8s.io/#what-is-the-gateway-api).

### Custom Resource

`parametersRef` is entirely a generic implementation-specific meaning.
This means implementations will either need a custom CRD or use untyped resources like ConfigMap.
Neither of these have any consistency between implementations.
While there will always be some vendor-specific requirements, there are also a number of configuration aspects of a Gateway that are common between implementations.
However, these cannot currently be expressed in a vendor-neutral way.

The original motivation behind `parametersRef` was for implementation specific concepts, while portable comments could be added into the API as first-class fields, but this has not been done (yet).

Additionally, there is hesitancy to use a CRD (which leads to CRD proliferation), which pushes users towards untyped ConfigMaps which are not much better than annotations.
The scoping, as mentioned above, is also a bit awkward of a cluster scoped resource pointing to a namespaced object.

### Separation of concerns

While there is value out of providing class-wide options as defaults, there is also value in providing these options on the object (Gateway) directly.

Some parallels in existing APIs:

[Policy Attachment](https://gateway-api.sigs.k8s.io/reference/policy-attachment) offers a hierarchy of defaults and overrides, allowing attachment to GatewayClass and Gateway.
This is similar to our needs here, but representing infrastructure configuration as a "Policy" is a bit problematic, and the existing mechanisms have no hierarchy.

In core Kubernetes, Pods declare their requirements (for example, CPU requests) inline in the Pod resource; there is not a `ResourceClass` API that abstracts these further.
These higher level abstractions are handled by layered APIs (whether this is a CRD, an admission webhook, CI/CD tooling, etc).
This allows users the flexibility to easily configure things per-pod basis.
If the infrastructure admin wants to impose defaults or requirements on this flexibility, they are able to do so (in fact, `LimitRanger` provides a built in mechanism to do so).

### Dynamic Changes

Currently, the spec recommends `GatewayClass` to be used as a *template*.
Changes to it are not expected to change deployed `Gateway`s.

This makes usage problematic in a declarative way.
For example, if I wanted to represent a `version` field and change that to trigger an upgrade, I would need to create an entirely new `Gateway`.

## API

In order to address the concerns above, I propose a standard `infrastructure` API is added to `Gateway` and `GatewayClass`.
Note the important part of this is the `Gateway` change; the `GatewayClass` aspect is mostly for consistency.

The exact fields are out of scope for this GEP and will be handled by additional GEPs.
One example GEP already depending on this is [GEP-1651](/geps/gep-1651).

The fields as defined below are, of course, not useful.
This is intended as a basis for other PRs, not to provide value on its own.
This GEP will remain in provisional until at least one field is ready to be promoted.

```go
type GatewaySpec struct {
  // Infrastructure defines infrastructure level attributes about this Gateway instance.
  Infrastructure GatewayInfrastructure `json:"infrastructure"`
  // ...
}

type GatewayClassSpec struct {
  // Infrastructure defines infrastructure level attributes for all Gateways in this class.
  // A Gateway may provide configuration for the same values; as all fields in GatewayInfrastructure are implementation specific,
  // the merging logic between these is as well. However, the GatewayClass is generally expected to be providing defaults
  // rather than overrides.
  Infrastructure GatewayClassInfrastructure `json:"infrastructure"`
  // ...
}

type GatewayInfrastructure struct {
  // ParametersRef provides a arbitrary implementation-specific configuration for
  // fields not expressed directly in this struct.
  // This follows the same semantics as GatewayClass's ParametersRef, but lives on the Gateway.
  ParametersRef ParametersReference
}

type GatewayClassInfrastructure struct {
}
```

### API Principles

For any given field, we will need to make two decisions:
* whether this should be a first-class field or a generic `parametersRef`.
* whether this field should be configurable on a Gateway and/or GatewayClass level

The choice to use an extension (`parametersRef`) or first-class field is a well known problem across the API, and the same logic will be used here.
Fields that are generally portable across implementations and have wide-spread demand and use cases will be promoted to first-class fields,
while vendor specific or niche fields will remain extensions.
Because infrastructure is somewhat inherently implementation specific, it is likely most fields will be Extended or ImplementationSpecific.
However, there are still a variety of concepts that have some meaning between implementations that can provide value to users.

Introduction at Gateway or GatewayClass level will depend on the specific field and use cases for the field.
In general, it makes sense to provide defaults (GatewayClass) and specific settings (Gateway) for most fields, but
this will be evaluated on a case-by-case basis.

### Status

The API should likely expose some status. However, it is not yet clear what that will look like.
This will be addressed prior to promotion beyond "Provisional".
