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
- Define other fields that can be used to verify the client certificate such as the Certificate Hash or Subject Alt Name. 

### API

* Introduce a `clientValidation` field with [Gateway.Tls][] that can be used to validate the client intiating the TLS connection
to the Gateway
* Introduce a `caCerficateRefs` field within `clientValidation` that can be used to specify a list of CA Certificates that
can be used as a trusted anchor to validate the certificates presented by the client

#### GO

```go
// ClientValidationContext holds configuration that can be used to validate the client intiating the TLS connection
// to the Gateway.
// By default, no client specific configuration is validated.
type ClientValidationContext struct {
    // CACertificateRefs contains one or more references to
    // Kubernetes objects that contain TLS certificates of
    // the Certificate Authorities that can be used to
    // as a trusted anchor to validate the certificates presented by the client.
    //
    // A single CACertificateRef to a Kubernetes ConfigMap with a key called `ca.crt`
    // has "Core" support.
    // Implementations MAY choose to support attaching multiple certificates to
    // a Listener, but this behavior is implementation-specific.
    //
    // References to a resource in different namespace are invalid UNLESS there
    // is a ReferenceGrant in the target namespace that allows the certificate
    // to be attached. If a ReferenceGrant does not allow this reference, the
    // "ResolvedRefs" condition MUST be set to False for this listener with the
    // "RefNotPermitted" reason.
    //
    // +kubebuilder:validation:MaxItems=64
    // +optional
    CACertificateRefs []corev1.ObjectReference `json:”caCertificateRefs,omitempty”`
}

```

#### YAML

```yaml
apiVersion: gateway.networking.k8s.io/v1beta1
kind: Gateway
metadata:
  name: mtls-basic
spec:
  gatewayClassName: acme-lb
  listeners:
  - name: foo-https
    protocol: HTTPS
    port: 443
    hostname: foo.example.com
    tls:
      certificateRefs:
      - kind: Secret
        group: ""
        name: foo-example-com-cert
      clientValidation:
        caCertificateRefs:
        - kind: ConfigMap
          group: ""
          name: foo-example-com-ca-cert
```

## References

[TLS Handshake Protocol]: https://www.rfc-editor.org/rfc/rfc5246#section-7.4
[Certificate Path Validation]: https://www.rfc-editor.org/rfc/rfc5280#section-6
[Gateway.TLS]: https://gateway-api.sigs.k8s.io/references/spec/#gateway.networking.k8s.io/v1beta1.GatewayTLSConfig
