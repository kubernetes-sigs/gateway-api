# GEP-746: Replace Cert Refs on HTTPRoute with Cross Namespace Refs from Gateway

* Issue: [#746](https://github.com/kubernetes-sigs/gateway-api/issues/746)
* Status: Standard

## TLDR

This GEP proposes that we should remove TLS Certificate references from
HTTPRoute and replace them with Cross Namespace Certificate references from
Gateways. Although that is not a complete replacement on its own, this GEP shows
how a controller could provide the rest of the functionality with this approach.

## Goals

* Remove a confusing and underspecified part of the API - cert refs on
  HTTPRoute.
* Add the ability to reference certificates in other namespaces from Gateways
  to replace much of the functionality that was enabled by cert refs on
  HTTPRoute.
* Describe how a controller could automate self service cert attachment to
  Gateway listeners.

## Non-Goals

* Actually provide a core implementation of a controller that can enable self
  service cert attachment. This may be worth considering at a later point, but
  is out of scope for this GEP.

## Introduction

TLS Certificate references on HTTPRoute have always been a confusing part of the
Gateway API. In the v1alpha2 release, we should consider removing this feature
while we still can. This GEP proposes an alternative that is simpler to work
with and understand, while also leaving sufficient room to enable all the same
capabilities that certificate references on HTTPRoute enabled.

### Attaching TLS Certificates with Routes is Confusing
One of the most confusing parts of the Gateway API is how certificates can be
attached to Routes. There are a variety of different factors that lead to
confusion here:

* It can be natural to assume that a certificate attached to a Route only
  applies to that Route. In reality, it applies to the entire listener(s)
  associated with that Route.
* This means that a Route can affect any other Routes attached to the same
  Gateway Listener. By attaching a Route to a Gateway Listener, you’re
  implicitly trusting all other Routes attached to that Gateway Listener.
* When multiple Routes specify a certificate for the same Listener, it’s
  possible that they will conflict and create more confusion.

### Why We Did It
To understand how we ended up with the ability to attach TLS certificates with
Routes, it’s helpful to look at the use cases for this capability:

1. Some users want Route owners to be able to attach arbitrary domains and certs
   to a Gateway listener.
   [#103](https://github.com/kubernetes-sigs/gateway-api/issues/103)
1. Some users want Route owners to control certs for their applications.

### Alternative Solutions

#### 1. Automation with tools like Cert-Manager
When automation is acceptable, the first use case is entirely possible with
tools like cert-manager that can watch Routes, generate certs for them, and
attach them to a Gateway.

#### 2. Cross Namespace Cert Direct References from Gateways
With the already established ReferenceGrant concept, we have established a safe
way to reference resources across namespaces. Although this would require some
coordination between Gateway and App owners, it would enable App owners to
retain full control of the certs used by their app without the extra confusion
that certs in HTTPRoute have led to.

### Enabling Self-Service Certificate Attachment for App Owners
Although this dramatically simplifies the API, it does not completely replace
the functionality that certs attached to HTTPRoutes enabled. Most notably, it
would be difficult to attach arbitrary self-provided certificates to a Gateway
listener without requiring manual changes from a Gateway admin.

There are a couple potential solutions here:

#### 1. Implement a selector for cert references instead of direct references
Although the simplicity of this approach is nice, it ends up with many of the
same problems as certificates attached to Routes have and feels inconsistent
with how Routes attach to Gateways.

#### 2. Implement a controller that attaches certificates to Gateway listeners
Similar to cert-manager, it could be possible to implement a controller that
watches for Secrets with a certain label, and attaches those to the specified
Gateway. Although it's out of scope for this GEP to completely define what a
controller like this could look like, it would likely need to include at least
one of the following safeguards:

1. A way to configure which namespaces could attach certificates for each
   domain.
2. A way to configure which namespaces could attach certificates to each
   Gateway (or Listener).
3. A way to use ReferenceGrant to indicate where references from Secrets to
   Gateways were trusted from and to.

## API

The API changes proposed here are quite small, mostly removing fields.

### Changes
1. The `LocalObjectReference` used for the `CertificateRef` field in
   `GatewayTLSConfig` would be replaced with an `ObjectReference`.
1. `ReferenceGrant` would be updated to note that references from Gateways to
   Secrets were part of the Core support level.

### Removals

From HTTPRouteSpec:
```go
    // TLS defines the TLS certificate to use for Hostnames defined in this
    // Route. This configuration only takes effect if the AllowRouteOverride
    // field is set to true in the associated Gateway resource.
    //
    // Collisions can happen if multiple HTTPRoutes define a TLS certificate
    // for the same hostname. In such a case, conflict resolution guiding
    // principles apply, specifically, if hostnames are same and two different
    // certificates are specified then the certificate in the
    // oldest resource wins.
    //
    // Please note that HTTP Route-selection takes place after the
    // TLS Handshake (ClientHello). Due to this, TLS certificate defined
    // here will take precedence even if the request has the potential to
    // match multiple routes (in case multiple HTTPRoutes share the same
    // hostname).
    //
    // Support: Core
    //
    // +optional
    TLS *RouteTLSConfig `json:"tls,omitempty"`
```

And the associated struct:
```go
// RouteTLSConfig describes a TLS configuration defined at the Route level.
type RouteTLSConfig struct {
    // CertificateRef is a reference to a Kubernetes object that contains a TLS
    // certificate and private key. This certificate is used to establish a TLS
    // handshake for requests that match the hostname of the associated HTTPRoute.
    // The referenced object MUST reside in the same namespace as HTTPRoute.
    //
    // CertificateRef can reference a standard Kubernetes resource, i.e. Secret,
    // or an implementation-specific custom resource.
    //
    // Support: Core (Kubernetes Secrets)
    //
    // Support: Implementation-specific (Other resource types)
    //
    CertificateRef LocalObjectReference `json:"certificateRef"`
}
```

From GatewayTlsConfig:
```go
    // RouteOverride dictates if TLS settings can be configured
    // via Routes or not.
    //
    // CertificateRef must be defined even if `routeOverride.certificate` is
    // set to 'Allow' as it will be used as the default certificate for the
    // listener.
    //
    // Support: Core
    //
    // +optional
    // +kubebuilder:default={certificate:Deny}
    RouteOverride *TLSOverridePolicy `json:"routeOverride,omitempty"`
```

And the associated types:
```go
type TLSRouteOverrideType string

const (
    // Allows the parameter to be configured from all routes.
    TLSROuteOVerrideAllow TLSRouteOverrideType = "Allow"

    // Prohibits the parameter from being configured from any route.
    TLSRouteOverrideDeny TLSRouteOverrideType = "Deny"
)

// TLSOverridePolicy defines a schema for overriding TLS settings at the Route
// level.
type TLSOverridePolicy struct {
    // Certificate dictates if TLS certificates can be configured
    // via Routes. If set to 'Allow', a TLS certificate for a hostname
    // defined in a Route takes precedence over the certificate defined in
    // Gateway.
    //
    // Support: Core
    //
    // +optional
    // +kubebuilder:default=Deny
    Certificate *TLSRouteOverrideType `json:"certificate,omitempty"`
}
```

## Prior Art

OpenShift already supports configuring TLS certificates on Routes. Although
largely similar to the Gateway API approach, there are some notable differences:

* Each Route can specify a maximum of 1 hostname
* When a Route is attached to a hostname, newer Routes can't use the same
  hostname unless all of the following are true:
    * The Routes are in the same namespace or the Router is configured to allow
      sharing hostnames across namespaces
    * The Routes have unique, non-overlapping paths specified
    * The Routes are not TCP or TLS routes

A typical configuration would involve a Router with `*.example.com` that has a
wildcard cert. Routes could be attached within those constraints without the
need for a cert. Routes can also use a different hostname if they also provide a
cert.

## Alternatives

### 1. Improved Documentation + Extended Support Level
My first attempt to improve this was to create a
[PR](https://github.com/kubernetes-sigs/gateway-api/pull/739) that would clarify
the documentation around how this works and lower the support level to extended.

Trying to improve the documentation around this feature made it clear how easy
it would be to get confused by how it worked. It would be only natural to assume
that a cert attached to a Route would only apply to that Route. The conflict
resolution semantics associated with this were both complicated and difficult to
surface to a user through status or other means.

Lowering the support level from core to extended also didn't make sense.
Although some implementers were uncomfortable with supporting this feature due
to the potential for vulnerabilities, that was not a sufficient reason to lower
the support level. An extended support level should only be used for features
that cannot be universally supported. That was not the case here. Instead there
were just very real questions around the safety of the feature.

The combination of those 2 factors led me to believe that this feature was not
well thought out and should be removed. Since this was essentially just a
shortcut to attaching certificates to a Gateway listener from different sources,
it seemed like there had to be a way that was both safer and easier to
understand. That led to this proposal.

### 2. Implement Hostname Restrictions
Similar to the OpenShift approach described above, we could enforce the
following:

1. Only a single hostname may be specified for HTTPRoutes with a certificate
   reference.
1. The oldest HTTPRoute to attach a certificate to a hostname would effectively
   own that hostname. No other HTTPRoutes could be attached with the same
   hostname unless they were explicitly allowed by that HTTPRoute.

The second condition would be difficult to validate. As we've seen elsewhere in
the API, it's difficult to determine which resource was first to claim a
hostname or path. Instead we have to rely on the oldest resource, which can
result in some weird and potentially breaking changes if an older resource
chooses to claim a hostname.

## References

Docs:

* [Gateway API: Replacing TLS Certificates in Routes](https://docs.google.com/document/d/1Cv95XFCL6S_9pIyS0drnsDLsfinWc2tHOFl_x3-_SWI/edit)