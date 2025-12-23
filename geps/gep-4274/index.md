# GEP-4274: BackendTLSPolicy Support for TLSRoute

* Issue: [#4274](https://github.com/kubernetes-sigs/gateway-api/issues/4274)
* Status: Provisional

## TLDR

Extend BackendTLSPolicy to support TLSRoute in termination mode, enabling
explicit backend TLS configuration for TLS-routed traffic from the Gateway to
backend services.

## Motivation

BackendTLSPolicy, as defined in [GEP-1897](../gep-1897/index.md), provides a
mechanism for configuring how a Gateway connects to backends via TLS.
Currently, BackendTLSPolicy is only supported for HTTPRoute. TLSRoute was
explicitly excluded from the initial implementation (see
[GEP-1897 Non-Goals](#references)).

TLSRoute is used for routing TLS traffic based on SNI and other TLS-specific
metadata. When configured in termination mode, the Gateway terminates the
incoming TLS connection and routes traffic based on the decrypted information.
In this scenario, there are legitimate use cases where the backend itself
requires TLS connections with specific validation requirements:

1. **TLS Re-encryption**: A Gateway terminates incoming TLS connections and
   needs to establish new TLS connections to backends that have their own
   certificates, ensuring end-to-end encryption.
2. **Compliance Requirements**: Organizations may require encryption between
   the Gateway and backends even after TLS termination, necessitating separate
   backend TLS connections with proper certificate validation.
3. **Zero-Trust Architecture**: In zero-trust environments, every network hop
   must be encrypted and authenticated, including the Gateway-to-backend
   connection after TLS termination.

Currently, users of TLSRoute cannot leverage BackendTLSPolicy to configure
these backend TLS connections, creating an inconsistency with HTTPRoute and
limiting deployment flexibility for TLS-based routing scenarios.

### User Stories

* As an application developer using TLSRoute in termination mode, I want to
  configure backend TLS connections with specific certificate validation
  requirements, so I can ensure secure communication with my backend services
  that have their own certificates.
* As a platform engineer, I want consistent TLS configuration capabilities
  across HTTPRoute and TLSRoute, so I can apply uniform security policies
  regardless of the routing mechanism used.
* As an application I want to maintain my existing BackendTLSPolicy configurations, 
  so I can preserve my backend security posture without rewriting policies.

## Goals

* Enable BackendTLSPolicy to attach to Services referenced by TLSRoute in
  termination mode, following the same pattern established for HTTPRoute.
* Maintain consistency with existing BackendTLSPolicy semantics and behavior
  defined in GEP-1897.
* Ensure that TLSRoute implementations can configure backend TLS connections
  with:
  * CA certificate validation
  * SNI configuration
  * Hostname verification
* Provide clear guidance on how BackendTLSPolicy interacts with TLSRoute in
  termination mode.

## API

**TODO**: First PR will not include any implementation details, in favor of
building consensus on the motivation, goals and non-goals first. _"How?"_ we
implement shall be left open-ended until _"What?"_ and _"Why?"_ are solid.

## References

* [GEP-1897: BackendTLSPolicy - Explicit Backend TLS Connection Configuration](https://gateway-api.sigs.k8s.io/geps/gep-1897/)
* [Issue #4274: BackendTLSPolicy support for TLSRoute](https://github.com/kubernetes-sigs/gateway-api/issues/4274)
