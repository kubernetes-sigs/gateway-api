# GEP-851: Allow Multiple Certificate Refs per Gateway Listener

* Issue: [#851](https://github.com/kubernetes-sigs/gateway-api/issues/851)
* Status: Standard

## TLDR

Replace `CertificateRef` field with a `CertificateRefs` field in Gateway
Listeners.

## Goals

Provide a path for implementations to support:

* RSA and ECDSA certs on the same Listener.
* Referencing certificates for different hostnames from the same Listener (maybe
  as part of a self-service TLS approach).
* Including new and old certificates as part of renewal process.
* Certificate pinning: enable implementations that require certain certificates
  for legacy clients while exposing a "normal" certificate to non-legacy
  clients.

## Non-Goals

* Define how implementations should support these features. Many implementations
  have limited control as far as how certs are handled. Some implementations
  just pass certs directly to the dataplane and rely on that implementation
  specific behavior to determine which certs are used for a given request. That
  makes it difficult to define any truly portable handling of multiple
  certificates.

## Introduction

As described above, there are a number of potential use cases for attaching
multiple certificates to a Gateway Listener. The most straightforward reason for
that involves attaching RSA and ECDSA certs. Although this is not a very common
use case, it is a clearly understood and broadly supported example of why this
change would be helpful. This change will enable implementations to support
these more advanced use cases while still providing a portable core.

## API

The `CertificateRef` field in `GatewayTLSConfig` would be replaced with the
following `CertificateRefs` field:

```go
    // CertificateRefs contains a series of references to Kubernetes objects that
    // contains TLS certificates and private keys. These certificates are used to
    // establish a TLS handshake for requests that match the hostname of the
    // associated listener.
    //
    // A single CertificateRef to a Kubernetes Secret has "Core" support.
    // Implementations MAY choose to support attaching multiple certificates to
    // a Listener, but this behavior is implementation-specific.
    //
    // References to a resource in different namespace are invalid UNLESS there
    // is a ReferenceGrant in the target namespace that allows the certificate
    // to be attached. If a ReferenceGrant does not allow this reference, the
    // "ResolvedRefs" condition MUST be set to False for this listener with the
    // "InvalidCertificateRef" reason.
    //
    // This field is required to have at least one element when the mode is set
    // to "Terminate" (default) and is optional otherwise.
    //
    // CertificateRefs can reference to standard Kubernetes resources, i.e.
    // Secret, or implementation-specific custom resources.
    //
    // Support: Core - A single reference to a Kubernetes Secret
    //
    // Support: Implementation-specific (More than one reference or other resource types)
    //
    // +optional
    // +kubebuilder:validation:MaxItems=64
    CertificateRefs []*SecretObjectReference `json:"certificateRefs,omitempty"`
```

## Alternatives

N/A
