# GEP-91: Client Certificate Validation for TLS terminating at the Gateway Listener

* Issue: [#91](https://github.com/kubernetes-sigs/gateway-api/issues/91)
* Status: Provisional

(See definitions in [GEP Status][/contributing/gep#status].)

## TLDR

This GEP proposes a way to validate the TLS certificate presented by the downstream client to the server
(Gateway Listener in this case) during a [TLS Handshake Protocol][], also commonly referred to as mutual TLS (mTLS).

## Goals
- Define an API field to specify the CA Certificate within the Gateway Listener configuration that can be used as a trusted anchor to validate the certificates presented by the client.

## Non-Goals
- Define other fields that can be used to verify the client certificate such as the Cerificate Hash or Subject Alt Name. 

## References

[TLS Handshake Protocol]: https://www.rfc-editor.org/rfc/rfc5246#section-7.4
[Certificate Path Validation]: https://www.rfc-editor.org/rfc/rfc5280#section-6
