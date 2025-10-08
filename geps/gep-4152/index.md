# GEP-4152: Extending TLS Validation in BackendTLSPolicy

* Issue: [#4152](https://github.com/kubernetes-sigs/gateway-api/issues/4152)
* Status: Provisional

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

* Enable connecting to backends over TLS without requiring certificate
  verification.
* Support certificate and SPKI pinning as alternatives to disabling verification
  or relying on CA trust chains.
* Maintain a secure-by-default approach, with certificate verification enabled
  unless explicitly opted out.
* Provide operator-level controls so Gateway constraints can restrict or permit
  the use of skip-verify.
* Provide clear runtime indicators that security is degraded when TLS validation
  is disabled.

## API

**TODO**: First PR will not include any implementation details, in favor of
building consensus on the motivation, goals and non-goals first. _"How?"_ we
implement shall be left open-ended until _"What?"_ and _"Why?"_ are solid.

## References

* [GEP-1897: BackendTLSPolicy - Explicit Backend TLS Connection Configuration](https://gateway-api.sigs.k8s.io/geps/gep-1897/)
