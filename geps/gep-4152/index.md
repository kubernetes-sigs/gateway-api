# GEP-4152: Extending TLS Validation in BackendTLSPolicy

* Issue: [#4152](https://github.com/kubernetes-sigs/gateway-api/issues/4152)
* Status: Implementable

## TLDR

The ability for the `BackendTLSPolicy` to skip TLS verification or to trust
specific backend certificates.

## Motivation

The current `BackendTLSPolicy` follows a secure-by-default approach that requires
users to provide a trusted CA certificate bundle or rely on the system’s default
certificate store (which typically includes root CAs) to validate backend server
certificates. However, real-world deployments include cases where strict
certificate validation may not be possible or practical, e.g., Development and
testing environments that use self-signed certificates generated dynamically at
runtime.

In such scenarios, users may need the flexibility to disable certificate
verification or to use certificate pinning. Certificate pinning offers a safer
and more controlled alternative, instead of bypassing TLS validation, the gateway
verifies that the backend’s certificate matches a known fingerprint or public key
hash. This preserves the confidentiality and integrity guarantees of TLS while
removing the operational overhead of managing full certificate chains or trusted
CA bundles.

### User Stories

* As an application developer, I want the option to disable backend TLS
  certificate verification on a per-backend basis, so I can connect to services
  using dynamically generated or self-signed certificates during development or
  testing.
* As an application developer, I want the option to skip SAN validation when 
  connecting to services using certificates that only contain a CommonName (CN),
  such as those issued by internal or legacy PKI systems that don't generate SANs.
* As an application developer, I want secure-by-default behavior, ensuring that
  certificate verification is always enabled unless I explicitly opt out, to
  prevent accidentally deploying insecure configurations to production.
* As an application developer, I want to securely communicate with backends
  using self-signed certificates without managing CA bundles or disabling
  verification, by trusting specific certificates directly.
* As a gateway operator, I want to control whether skipping TLS validation is
  permitted for specific Gateways.
* As a security officer, I want transparency and auditability into where TLS
  verification has been disabled.

## Goals

* Enable connecting to backends over TLS with full, partial, or no
  certificate verification.
* Support certificate pinning as an alternative to disabling verification or
  relying on CA trust chains.
* Maintain a secure-by-default approach, with certificate verification enabled
  unless explicitly opted out.
* Provide operator-level controls so Gateway constraints can restrict or permit
  the use of skip-verify.
* Provide clear runtime indicators that security is degraded when TLS validation
  is disabled.

## API

This GEP extends `BackendTLSPolicy.spec.validation` with two new optional fields:
* `mode` selects the certificate verification strategy.
* `trustedCertificates` an alternative trust source to `caCertificateRefs`
  and `wellKnownCACertificates`, allowing backend certificates to be trusted 
  directly without requiring a CA chain.

All features introduced by this GEP are Extended support. Only the default
values, which preserve the existing behavior, are considered Core.

### BackendTLSPolicy Changes

#### BackendTLSPolicy Validation Mode
The `mode` field selects one of the following verification strategies:

* `Standard` _(default)_ performs full TLS verification. The backend certificate 
  chain is validated against the trust anchors configured in validation (either 
  `caCertificateRefs` or `wellKnownCACertificates`). As an alternative,
  `trustedCertificates` can be used instead of  `caCertificateRefs` or 
  `wellKnownCACertificates` to establish trust. Connections to backends presenting
  invalid, expired, untrusted, SAN-mismatch, or hostname-mismatched certificates 
  are rejected.
* `InsecureAllowExpired` performs full TLS verification against the configured
  trust anchors (`caCertificateRefs`, `wellKnownCACertificates`, or
  `trustedCertificates`), including SAN-matching, but accepts certificates
  that are expired.
* `InsecureSkipOnlySANVerification` performs chain validation against the configured
  trust anchors (`caCertificateRefs`, `wellKnownCACertificates`, or
  `trustedCertificates`), but skips SAN-matching.
* `InsecureSkipVerification` disables all certificate verification. Neither the
  certificate chain nor the hostname is validated, and the backend may present
  any certificate, including self-signed, expired, or attacker-controlled ones.
  TLS is still negotiated, so traffic is encrypted in transit, but the gateway
  has no cryptographic assurance that it is talking to the intended backend.
  This mode is intended for development and testing environments and SHOULD NOT
  be used in production.

#### Trusted Certificates
`trustedCertificates` is an alternative to `caCertificateRefs` and
`wellKnownCACertificates`. It references one or more Kubernetes objects
(i.e., ConfigMap or Secret) containing PEM-encoded certificates that
are trusted directly as leaf certificates. When set, the backend certificate
is accepted if it matches one of the referenced certificates exactly, with no
CA chain validation required.

Implementations MAY compute the SHA-256 hash of the DER-encoded representation
of the referenced PEM-encoded certificate and use that hash to match the
backend certificate, rather than comparing the full certificate.

#### API Validations

* When `mode` is `Standard` or `InsecureSkipOnlySANVerification`, at least one of
  `caCertificateRefs`, `wellKnownCACertificates`, or `trustedCertificates`
  MUST be set. `caCertificateRefs` and `wellKnownCACertificates` remain mutually
  exclusive and only at most one of them can be set.
* When `mode` is `InsecureSkipVerification`, `caCertificateRefs`,
  `wellKnownCACertificates`, and `trustedCertificates` MUST all be empty.
* `subjectAltNames` MUST be empty when `mode` is `InsecureSkipOnlySANVerification`
  or `InsecureSkipVerification`, as SAN matching is not performed in these
  modes.
* `hostname` MUST be set in all modes. It is used as the SNI value for
  certificate selection during the TLS handshake, independent of whether it
  is subsequently matched against the presented certificate.

#### Examples
BackendTLSPolicy with trusted certificates:
```yaml
apiVersion: gateway.networking.k8s.io/v1
kind: BackendTLSPolicy
metadata:
  name: example
spec:
  targetRefs:
  - group: ""
    kind: Service
    name: example
  validation:
    mode: Standard
    trustedCertificates:
    - group: ""
      kind: ConfigMap
      name: example-cert-1
    - group: ""
      kind: Secret
      name: example-cert-2  
    hostname: internal.example.com
```

BackendTLSPolicy with insecure verification:
```yaml
apiVersion: gateway.networking.k8s.io/v1
kind: BackendTLSPolicy
metadata:
  name: example
spec:
  targetRefs:
  - group: ""
    kind: Service
    name: example
  validation:
    mode: InsecureSkipVerification
    hostname: internal.example.com
```

### Restricting Insecure Verification Modes via ValidatingAdmissionPolicy

Gateway operators can restrict insecure backend TLS verification modes using a 
ValidatingAdmissionPolicy (VAP), which rejects any `BackendTLSPolicy` that sets 
`mode` to `InsecureSkipVerification` or `InsecureSkipOnlySANVerification`.

This approach keeps policy enforcement in a standard Kubernetes mechanism
rather than extending the Gateway API surface.

#### Example
```yaml
apiVersion: admissionregistration.k8s.io/v1
kind: ValidatingAdmissionPolicy
metadata:
  name: disallow-insecure-backend-tls-verification
spec:
  failurePolicy: Fail
  matchConstraints:
    resourceRules:
      - apiGroups:   ["gateway.networking.k8s.io"]
        apiVersions: ["v1alpha3", "v1"]
        operations:  ["CREATE", "UPDATE"]
        resources:   ["backendtlspolicies"]
  validations:
    - expression: >
        !has(object.spec.validation.mode) ||
        (object.spec.validation.mode != "InsecureSkipVerification" &&
         object.spec.validation.mode != "InsecureAllowExpired" &&
         object.spec.validation.mode != "InsecureSkipSANVerification")
      reason: Forbidden
      messageExpression: >
        "BackendTLSPolicy '" + object.metadata.name +
        "' uses insecure verification mode '" +
        object.spec.validation.mode +
        "'. Only Standard verification is permitted in this cluster."
```

## References

* [GEP-1897: BackendTLSPolicy - Explicit Backend TLS Connection Configuration](https://gateway-api.sigs.k8s.io/geps/gep-1897/)
