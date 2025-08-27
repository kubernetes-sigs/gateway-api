# GEP-3155: Complete Backend mutual TLS Configuration

* Issue: [#3155](https://github.com/kubernetes-sigs/gateway-api/issues/3155)
* Status: Experimental

## TLDR

This GEP aims to complete the configuration required for Backend mutual TLS in Gateway
API. This includes the following new capabilities:

1. Configuration for the client certificate Gateways should use when connecting
   to Backends
1. Ability to specify SANs on BackendTLSPolicy
1. Add TLS options to BackendTLSPolicy to mirror TLS config on Gateways

## Goals

* Add sufficient configuration that basic mutual TLS is possible between Gateways and
  Backends
* Enable the optional use of SPIFFE for Backend mutual TLS

## Non-Goals

* Define how automatic mTLS should be implemented with Gateway API

## Introduction

This is a wide ranging GEP intending to cover three additions to the API that all
have a shared goal - enabling backend mutual TLS with Gateway API. Although this
specific GEP focuses on manual configuration across the board, the hope is that
it will also enable higher level automation to simplify this process for users.

## API

### Client Certs on Gateways

A key requirement of mutual TLS is that the Gateway can provide a client cert to the
backend. This adds that configuration to both Gateway and Service (via
BackendTLSPolicy).

#### Gateway-level (Core support)
Specifying credentials at the gateway level is the default operation mode, where all
backends will be presented with a single gateway certificate. Per-service overrides are
subject for consideration as the future work.

**1. Add a new `BackendValidation` field at TLSConfig struct located in GatewayTLSConfig.Default field**

```go
// TLSConfig describes TLS configuration that can apply to multiple Listeners
// within this Gateway.
type TLSConfig struct {
    ...
	// GatewayBackendTLS describes TLS configuration for gateway when connecting
	// to backends.
	// Support: Core
	//
	// +optional
	// <gateway:experimental>
	BackendValidation *GatewayBackendTLS `json:"backendValidation,omitempty"`
}
type GatewayBackendTLS struct {
  // ClientCertificateRef is a reference to an object that contains a Client
  // Certificate and the associated private key.
  //
  // References to a resource in different namespace are invalid UNLESS there
  // is a ReferenceGrant in the target namespace that allows the certificate
  // to be attached. If a ReferenceGrant does not allow this reference, the
  // "ResolvedRefs" condition MUST be set to False for this listener with the
  // "RefNotPermitted" reason.
  //
  // ClientCertificateRef can reference to standard Kubernetes resources, i.e.
  // Secret, or implementation-specific custom resources.
  //
  // This setting can be overridden on the service level by use of BackendTLSPolicy.
  ClientCertificateRef SecretObjectReference `json:"clientCertificateRef,omitempty"`
}
```

### SANs on BackendTLSPolicy

This change enables the backend certificate to have a different identity than the SNI
(both are currently tied to the hostname field). This is particularly useful
when using SPIFFE, which relies on URI Subject Names which are not valid SNIs
as per https://www.rfc-editor.org/rfc/rfc6066.html#section-3.

In such case either connection properties or an arbitrary SNI, like cluster-local
service name could be used for certificate selection, while the identity validation
will be done based on SubjectAltNames field.

When specified, the certificate served from the backend MUST have at least one Subject
Alternate Name matching one of the specified SubjectAltNames.



**1. Add a new `SubjectAltNames` field to `BackendTLSPolicyValidation`**

```go
type BackendTLSPolicyValidation struct {
  // SubjectAltNames contains one or more Subject Alternative Names.
  // When specified, the certificate served from the backend MUST have at least one
  // Subject Alternate Name matching one of the specified SubjectAltNames.
  // +kubebuilder:validation:MaxItems=5
  SubjectAltNames []SubjectAltName `json:"subjectAltNames,omitempty"`
}

// +kubebuilder:validation:Enum=Cookie;Header
type SubjectAltNameType string

const (
   // HostnameSubjectAltNameType specifies hostname-based SAN.
   //
   // Support: Core
   HostnameSubjectAltNameType SubjectAltNameType = "Hostname"

   // URISubjectAltNameType specifies URI-based SAN, e.g. SPIFFE id.
   //
   // Support: Core
   URISubjectAltNameType SubjectAltNameType = "URI"
)


type SubjectAltName struct {
  // Type determines the format of the Subject Alternative Name. Always required.
  Type SubjectAltNameType `json:"type"`

  // Hostname contains Subject Alternative Name specified in DNS name format. Required when Type is set to Hostname, ignored otherwise.
  Hostname v1.PreciseHostname `json:"hostname,omitempty"`

  // URI contains Subject Alternative Name specified in URI format. Required when Type is set to URI, ignored otherwise.
  URI string `json:"uri,omitempty"`
}
```

**2. Modify Spec for `BackendTLSPolicyValidation` `Hostname`**

Before:
```go
  // 2. Hostname MUST be used for authentication and MUST match the certificate
  //    served by the matching backend.
```

After:
```go
  // 2. Only if SubjectAltNames is not specified, Hostname MUST be used for
  //    authentication and MUST match the certificate served by the matching
  //    backend.
```

### Allow per-service TLS settings BackendTLSPolicy

Gateway level TLS configuration already includes an `options` field. This has
been helpful for implementation-specific TLS configurations, or simply features
that have not made it to the core API yet. It would be similarly useful to have
an identical field on BackendTLSPolicy.

Examples:
- configuration options for vendor-specific mTLS automation
- restrictions on the minimum supported TLS version or supported cipher suites

```go
type BackendTLSPolicySpec struct {
  // Options are a list of key/value pairs to enable extended TLS
  // configuration for each implementation. For example, configuring the
  // minimum TLS version or supported cipher suites.
  //
  // A set of common keys MAY be defined by the API in the future. To avoid
  // any ambiguity, implementation-specific definitions MUST use
  // domain-prefixed names, such as `example.com/my-custom-option`.
  // Un-prefixed names are reserved for key names defined by Gateway API.
  //
  // Support: Implementation-specific
  //
  // +optional
  // +kubebuilder:validation:MaxProperties=16
  Options map[AnnotationKey]AnnotationValue `json:"options,omitempty"`
}
```

## Conformance Details

Conformance tests will be written to ensure the following:

1. When SubjectAltNames are specified in BackendTLSPolicy:
    - The hostname field is still used as SNI, if specified
    - A certificate with at least one matching SubjectAltName is accepted
    - A certificate without a matching SubjectAltName is rejected

2. When a Client Certificate is specified on a Gateway:
    - It is applied to all services.
    - The appropriate status condition is populated if the reference is invalid

## Future work
This GEP does not cover per-service overrides for client certificate. This is mostly for two reasons:

- it supports only the niche use cases - it should be reconsidered in future
- in current model, where BackendTLSPolicy shares namespace with the service instead of the Gateway, there are non-trivial security implications to adding client certificate configuration at this level - therefore ownership and colocation of BackendTLSPolicy (Service vs Gateway) needs to be figured out first

## References

This is a natural continuation of
[GEP-2907](../gep-2907/index.md), the memorandum GEP
that provided the overall vision for where TLS configuration should fit
throughout Gateway API.
