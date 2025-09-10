# GEP-4080: Certificate Revocation Lists for Certificate Validiation

* Issue: [#4080](https://github.com/kubernetes-sigs/gateway-api/issues/4080)
* Status: Provisional

(See [status definitions](../overview.md#gep-states).)

The key words "MUST", "MUST NOT", "REQUIRED", "SHALL", "SHALL NOT", "SHOULD",
"SHOULD NOT", "RECOMMENDED", "NOT RECOMMENDED", "MAY", and "OPTIONAL" in this
document are to be interpreted as described in BCP 14 ([RFC8174]) when, and
only when, they appear in all capitals, as shown here.

[RFC8174]: https://www.rfc-editor.org/rfc/rfc8174

## What
Extend the TLS certificate validation mechanisms defined in [GEP-91](../gep-91/index.md) (Frontend TLS Validation) and [GEP-1897](../gep-1897/index.md) (Backend TLS Validation) by introducing support for Certificate Revocation Lists (CRLs).  

With this enhancement, operators and application developpers can configure CRLs that Gateways use during TLS authentication, both when verifying clients connecting to the Gateway and when verifying backends that the Gateway connects to. In both cases, the Gateway will check the presented certificate against the configured CRLs to ensure it has not been revoked by the issuing Certificate Authority (CA).

## Why
While [GEP-91](../gep-91/index.md) and [GEP-1897](../gep-1897/index.md) define how certificates are validated against trusted CAs, neither addresses the issue of revocation. In practice, certificates may need to be revoked long before expiration (e.g., if private keys are compromised, a device is decommissioned, or access rights are withdrawn). Without revocation checks, Gateways may continue to trust certificates that should no longer be valid, exposing clusters to unauthorized access.  

This limitation is critical in two directions:  
* On the client side, a revoked certificate could still authenticate and gain access if CRLs are not enforced.  
* On the backend side, the Gateway could continue to route requests to services or workloads using revoked credentials, undermining the security of service-to-service communication.  

## Who: Beneficiaries
* **Application Developers**: Gain stronger guarantees that their applications are protected from unauthorized clients and backends using revoked certificates.  
* **Platform Operators/Administrators**: Without CRL support, the only way to remain secure in the face of a compromised certificate is to rotate the entire CA and reissue all certificates, which is an expensive, disruptive, and often impractical operation. CRL support eliminates this burden by allowing platform operators to revoke only the compromised certificates while leaving the rest of the trust hierarchy intact.  
