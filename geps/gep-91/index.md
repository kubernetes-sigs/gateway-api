# GEP-91: Client Certificate Validation for TLS terminating at the Gateway Listener

* Issue: [#91](https://github.com/kubernetes-sigs/gateway-api/issues/91)
* Status: Implementable 

(See definitions in [GEP Status][/contributing/gep#status].)

## TLDR

This GEP proposes a way to validate the TLS certificate presented by the downstream client to the server
(Gateway Listener in this case) during a [TLS Handshake Protocol][].

## Goals
- Define an API field to specify the CA Certificate within the Gateway Listener configuration that can be used as a trust anchor to validate the certificates presented by the client. This use case has been been highlighted in the [Gateway API TLS Use Cases][] document under point 7.

## Non-Goals
- Define other fields that can be used to verify the client certificate such as the Certificate Hash or Subject Alt Name. 

## Existing support in Implementations

This feature is widely supported in implementations that support Gateway API.
This table highlights the support. Please feel free to add any missing implementations not mentioned below.

| Implementation | Support       |
|----------------|------------|
| Apache APISIX  | [ApisixTls.Client.CASecret](https://apisix.apache.org/docs/ingress-controller/tutorials/mtls/#mutual-authentication)      |
| Contour        | [HTTPProxy.Spec.VirtualHost.Tls.ClientValidation.CASecret](https://projectcontour.io/docs/v1.17.1/config/tls-termination/)      |
| Emissary Ingress| [TlSContext.Spec.Secret](https://www.getambassador.io/docs/emissary/latest/topics/running/tls/mtls)     |
| Gloo Edge      | [VirtualService.Spec.SSLConfig.SecretRef](https://docs.solo.io/gloo-edge/latest/guides/security/tls/server_tls/#configuring-downstream-mtls-in-a-virtual-service)      |
| Istio          | [Gateway.Spec.Servers.TLS.Mode](https://istio.io/latest/docs/tasks/traffic-management/ingress/secure-ingress/#configure-a-mutual-tls-ingress-gateway)      |
| Kong           | [mTLS Plugin](https://docs.konghq.com/hub/kong-inc/mtls-auth/)      |
| Traefik        | [TLSOption.Spec.ClientAuth](https://doc.traefik.io/traefik/https/tls/#client-authentication-mtls)    |

### API

* Introduce a `clientValidation` field of type `ClientValidationContext` within [GatewayTLSConfig][] that can be used to validate the client initiating the TLS connection
to the Gateway.
* Introduce a `caCertificateRefs` field within `ClientValidationContext` that can be used to specify a list of CA Certificates that
can be used as a trust anchor to validate the certificates presented by the client.
* Add CEL validation to ensure that `caCertificateRefs` cannot be empty. This validation will be removed once more fields are added
into `clientValidation`.
* This new field is mutually exclusive with the [BackendTLSPolicy][] configuation which is used to validate the TLS certificate presented by the peer on the connection between the Gateway and the backend, and this GEP is adding support for validating the TLS certificate presented by the peer on the connection between the Gateway and the downstream client.

#### GO

```go
// ClientValidationContext holds configuration that can be used to validate the client intiating the TLS connection
// to the Gateway.
// By default, no client specific configuration is validated.
type ClientValidationContext struct {
    // CACertificateRefs contains one or more references to
    // Kubernetes objects that contain TLS certificates of
    // the Certificate Authorities that can be used
    // as a trust anchor to validate the certificates presented by the client.
    //
    // A single CA certificate reference to a Kubernetes ConfigMap kind has "Core" support.
    // Implementations MAY choose to support attaching multiple CA certificates to
    // a Listener, but this behavior is implementation-specific.
    //
    // Support: Core - An optional single reference to a Kubernetes ConfigMap,
    // with the CA certificate in a key named `ca.crt`.
    //
    // Support: Implementation-specific (More than one reference, or other kinds
    // of resources).
    //
    // References to a resource in a different namespace are invalid UNLESS there
    // is a ReferenceGrant in the target namespace that allows the certificate
    // to be attached. If a ReferenceGrant does not allow this reference, the
    // "ResolvedRefs" condition MUST be set to False for this listener with the
    // "RefNotPermitted" reason.
    //
    // +kubebuilder:validation:MaxItems=8
    // +optional
    CACertificateRefs []SecretObjectReference `json:”caCertificateRefs,omitempty”`
}

```

#### YAML

```yaml
apiVersion: gateway.networking.k8s.io/v1beta1
kind: Gateway
metadata:
  name: client-validation-basic
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

## Deferred

This section highlights use cases that may be covered in a future iteration of this GEP

* Using system CA certificates as the trust anchor to validate the certificates presented by the client.

## References

[TLS Handshake Protocol]: https://www.rfc-editor.org/rfc/rfc5246#section-7.4
[Certificate Path Validation]: https://www.rfc-editor.org/rfc/rfc5280#section-6
[GatewayTLSConfig]: https://gateway-api.sigs.k8s.io/references/spec/#gateway.networking.k8s.io/v1.GatewayTLSConfig
[BackendTLSPolicy]: https://gateway-api.sigs.k8s.io/api-types/backendtlspolicy/
[Gateway API TLS Use Cases]: https://docs.google.com/document/d/17sctu2uMJtHmJTGtBi_awGB0YzoCLodtR6rUNmKMCs8/edit?pli=1#heading=h.cxuq8vo8pcxm
