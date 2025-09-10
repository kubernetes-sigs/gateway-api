# GEP-4078: Certificate Pinning for Client Certificate Validation

* Issue: [#4078](https://github.com/kubernetes-sigs/gateway-api/issues/4078)
* Status: Provisional

(See [status definitions](../overview.md#gep-states).)

The key words "MUST", "MUST NOT", "REQUIRED", "SHALL", "SHALL NOT", "SHOULD",
"SHOULD NOT", "RECOMMENDED", "NOT RECOMMENDED", "MAY", and "OPTIONAL" in this
document are to be interpreted as described in BCP 14 ([RFC8174]) when, and
only when, they appear in all capitals, as shown here.

[RFC8174]: https://www.rfc-editor.org/rfc/rfc8174

## What
Enhance the existing Client Certificate Validation defined in [GEP-91](../gep-91/index.md) by introducing support for certificate pinning. This allows specifying one or more certificate or public key hashes (SPKI) that are considered valid for client connections. During TLS client authentication, the Gateway will validate not only against the configured CAs, but also against the pinned certificates or keys. This provides a mechanism to restrict allowed clients to a narrowly defined set of certificates, even if the CA trust domain is broad.

## Why
While [GEP-91](../gep-91/index.md) enables client certificate validation against a CA, in practice many operators rely on large organizational CAs or even on public CAs. These often cover many certificates and, as a result, unintentionally expand the trust boundary. This limitation becomes particularly significant when client certificates are used not only for encryption but also for authentication. In these cases, operators may want to restrict access to an explicit set of identities, rather than the entire population served by a CA.  
In addition, certificate pinning strengthens security by reducing exposure to rogue certificates and mitigating the risk of man-in-the-middle (MITM) attacks.

## Who: Beneficiaries
* **Application Developers**: Can use certificate pinning as a lightweight alternative to JWTs or other token systems, authenticating users or devices directly through TLS without additional infrastructure.  
* **Platform Operators/Administrators**: Gain a standardized way to enforce fine-grained client restrictions, even when using broad or public CAs, which improves security.  
* **Security and Compliance Teams**: Benefit from a certificate-bound authentication mechanism that limits access to explicitly approved identities, aligning with regulatory and organizational policies.