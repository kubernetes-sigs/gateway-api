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

**1. Add a new `Backend` field at GatewayTLSConfig struct located in Gateway**

```go
// GatewayTLSConfig specifies frontend and backend tls configuration for gateway.
type GatewayTLSConfig struct {
	// Backend describes TLS configuration for gateway when connecting
	// to backends.
	//
	// Note that this contains only details for the Gateway as a TLS client,
	// and does _not_ imply behavior about how to choose which backend should
	// get a TLS connection. That is determined by the presence of a BackendTLSPolicy.
	//
	// Support: Core
	//
	// +optional
	// <gateway:experimental>
	Backend *GatewayBackendTLS `json:"backend,omitempty"`
    ...
}
type GatewayBackendTLS struct {
  // ClientCertificateRef references an object that contains a client certificate 
  // and its associated private key. It can reference standard Kubernetes resources,
  // i.e., Secret, or implementation-specific custom resources.
  //
  // A ClientCertificateRef is considered invalid if:
  //
  // * It refers to a resource that cannot be resolved (e.g., the referenced resource
  //   does not exist) or is misconfigured (e.g., a Secret does not contain the keys
  //   named `tls.crt` and `tls.key`). In this case, the `ResolvedRefs` condition 
  //   on the Gateway MUST be set to False with the Reason `InvalidClientCertificateRef`
  //   and the Message of the Condition MUST indicate why the reference is invalid.
  //
  // * It refers to a resource in another namespace UNLESS there is a ReferenceGrant
  //   in the target namespace that allows the certificate to be attached.
  //   If a ReferenceGrant does not allow this reference, the `ResolvedRefs` condition 
  //   on the Gateway MUST be set to False with the Reason `RefNotPermitted`.
  //
  // Implementations MAY choose to perform further validation of the certificate
  // content (e.g., checking expiry or enforcing specific formats). In such cases,
  // an implementation-specific Reason and Message MUST be set.
  //
  // Support: Core - Reference to a Kubernetes TLS Secret (with the type `kubernetes.io/tls`).
  // Support: Implementation-specific - Other resource kinds or Secrets with a
  // different type (e.g., `Opaque`).
  // +optional
  // <gateway:experimental>
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
  // If SubjectAltNames are specified, Hostname MUST NOT be used for authentication,
  // even if this would cause a failure in the case that the SubjectAltNames do not match.
  // If you want to use Hostname for authentication, you must add Hostname to the SubjectAltNames list.
  //
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
