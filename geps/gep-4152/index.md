# GEP-4152: Extending TLS Validation in BackendTLSPolicy

* Issue: [#4152](https://github.com/kubernetes-sigs/gateway-api/issues/4152)
* Status: Implementable

## TLDR

The ability for the `BackendTLSPolicy` to skip TLS verification or to validate
certificates based on their fingerprint or public key hash.

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
* As an application developer, I want an alternative to disabling verification,
  such as certificate or SPKI pinning, so I can securely communicate with
  backends using self-signed certificates without managing CA bundles.
* As a gateway operator, I want to control whether skipping TLS validation is
  permitted for specific Gateways.
* As a security officer, I want transparency and auditability into where TLS
  verification has been disabled.

## Goals

* Enable connecting to backends over TLS with full, partial, or no
  certificate verification.
* Support certificate and SPKI pinning as alternatives to disabling verification
  or relying on CA trust chains.
* Maintain a secure-by-default approach, with certificate verification enabled
  unless explicitly opted out.
* Provide operator-level controls so Gateway constraints can restrict or permit
  the use of skip-verify.
* Provide clear runtime indicators that security is degraded when TLS validation
  is disabled.

## API

This GEP extends `BackendTLSPolicy.spec.validation` with two new optional fields:
* `mode` selects the certificate verification strategy.
* `trustedCertificateHashes` an alternative trust source to `caCertificateRefs`
  and `wellKnownCACertificates`, allowing backends to be trusted directly by
  certificate or public-key hash without requiring a CA chain.

All features introduced by this GEP are Extended support. Only the default
values, which preserve the existing behavior, are considered Core.

### BackendTLSPolicy Changes

#### BackendTLSPolicy Validation Mode
The `mode` field selects one of the following verification strategies:

* `Strict` _(default)_ performs full TLS verification. The backend certificate 
  chain is validated against the trust anchors configured in validation (either 
  `caCertificateRefs` or `wellKnownCACertificates`). As an alternative,
  `trustedCertificateHashes` can be used instead of  `caCertificateRefs` or 
  `wellKnownCACertificates` to establish trust. Connections to backends presenting
  invalid, expired, untrusted, SAN-mismatch, or hostname-mismatched certificates 
  are rejected.
* `InsecureSkipSANVerification` performs chain validation against the configured
  trust anchors (`caCertificateRefs`, `wellKnownCACertificates`, or
  `trustedCertificateHashes`), but skips SAN-matching.
* `InsecureSkipVerification` disables all certificate verification. Neither the
  certificate chain nor the hostname is validated, and the backend may present
  any certificate, including self-signed, expired, or attacker-controlled ones.
  TLS is still negotiated, so traffic is encrypted in transit, but the gateway
  has no cryptographic assurance that it is talking to the intended backend.
  This mode is intended for development and testing environments.

#### Trusted Certificate Hashes
`trustedCertificateHashes` is an alternative to `caCertificateRefs` and
`wellKnownCACertificates`. When set, the backend certificate is accepted if it
(or its Subject Public Key Info) matches one of the listed hashes, no CA
chain is required.

#### API Validations

* When `mode` is `Strict` or `InsecureSkipSANVerification`, at least one of
  `caCertificateRefs`, `wellKnownCACertificates`, or `trustedCertificateHashes`
  MUST be set.
* When `mode` is `InsecureSkipVerification`, `caCertificateRefs`,
  `wellKnownCACertificates`, and `trustedCertificateHashes` MUST all be empty.
* `subjectAltNames` MUST be empty when `mode` is `InsecureSkipSANVerification`
  or `InsecureSkipVerification`, as SAN matching is not performed in these
  modes.
* `hostname` MUST be set in all modes. It is used as the SNI value for
  certificate selection during the TLS handshake, independent of whether it
  is subsequently matched against the presented certificate.

#### Examples
**BackendTLSPolicy with trusted certificate hashes:**
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
    mode: Strict
    trustedCertificateHashes:
    - type: SPKI
      algorithm: SHA256
      value: "qrvdF0L7Kp5l3H8k0m3x7VZq3p5O6s4L4kC2Z7tZt+Q="
    - type: Certificate
      algorithm: SHA256
      value: "9C:5E:AD:EE:F4:38:A4:BC:4D:99:65:7B:12:C8:3D:7E:21:B7:40:8F:2E:91:5C:6A:D3:B0:88:19:42:60:AF:33"
    hostname: internal.example.com
```

**BackendTLSPolicy with insecure verification:**
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

### Gateway-Level Changes
A new `mode` field is added to `Gateway.spec.tls.backend.validation` to let
operators control which `BackendTLSPolicy` verification modes are permitted
for backends behind this Gateway.

* `AllowStrictOnly` _(default)_ permits only `BackendTLSPolicy` objects
  with `mode: Strict`. Policies requesting `InsecureSkipSANVerification` or
  `InsecureSkipVerification` are rejected with the `Accepted` condition on the 
  policy ancestor status set to `False` with `Reason: InsecureVerificationNotAllowed`.
  In this case implementations MUST NOT fall back to unencrypted (plaintext) or 
  insecure TLS connections, and the client MUST receive an HTTP 5xx error
  response.
* `AllowInsecureVerification` permits all verification modes, including 
  `InsecureSkipSANVerification` and `InsecureSkipVerification`.

#### Status and Observability
When `mode` is changed from `AllowStrictOnly` to `AllowInsecureVerification`,
the `InsecureBackendValidationMode` condition MUST be set to `True` with
Reason `ConfigurationChanged` on the Gateway. This condition is removed as
soon as `mode` is changed back to `AllowStrictOnly`.

#### Example

**Gateway permitting insecure backend verification modes:**

```yaml
apiVersion: gateway.networking.k8s.io/v1
kind: Gateway
metadata:
  name: example
spec:
  tls:
    backend:
      validation:
        mode: AllowInsecureVerification
```

## References

* [GEP-1897: BackendTLSPolicy - Explicit Backend TLS Connection Configuration](https://gateway-api.sigs.k8s.io/geps/gep-1897/)
