# GEP-91: Client Certificate Validation for TLS terminating at the Gateway Listener

* Issue: [#91](https://github.com/kubernetes-sigs/gateway-api/issues/91)
* Status: Implementable

(See definitions in [GEP Status][/contributing/gep#status].)

## TLDR

This GEP proposes a way to validate the TLS certificate presented by the downstream client to the server
(Gateway Listener in this case) during a [TLS Handshake Protocol][].

## Goals

* Define an API field to specify the CA Certificate within the Gateway Listener configuration that can be used as a trust anchor to validate the certificates presented by the client. This use case has been highlighted in the [TLS Configuration GEP][] under segment 1 and in the [Gateway API TLS Use Cases][] document under point 7.

## Non-Goals
* Define other fields that can be used to verify the client certificate such as the Certificate Hash.

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
| NGINX Ingress Controller | [ingressMTLS](https://docs.nginx.com/nginx-ingress-controller/configuration/policy-resource/#ingressmtls)    |

### API

* Introduce a `FrontendValidation` field of type `FrontEndTLSValidationContext` within [GatewayTLSConfig][] that can be used to validate the peer(frontend) with which the TLS connection is being made.
* Introduce a `caCertificateRefs` field within `FrontEndTLSValidationContext` that can be used to specify a list of CA Certificates that can be used as a trust anchor to validate the certificates presented by the client.
* This new field is mutually exclusive with the [BackendTLSPolicy][] configuation which is used to validate the TLS certificate presented by the peer on the connection between the Gateway and the backend, and this GEP is adding support for validating the TLS certificate presented by the peer on the connection between the Gateway and the frontend (downstream client).

#### GO

```go

type GatewayTLSConfig struct {
......
    // FrontendValidation holds configuration for validating the frontend (client).
    // Setting this field will require clients to send a client certificate
    // required for validation. In browsers this may result in a dialog appearing 
    // that requests a user to specify the client certificate.
    // The maximum depth of a certificate chain accepted in verification is Implementation specific.
    FrontendValidation *FrontEndTLSValidationContext `json:"frontEndValidation,omitempty"`
}

// FrontEndTLSValidationContext holds configuration that can be used to validate the frontend in the TLS connection
type FrontEndTLSValidationContext struct {
    // CACertificateRefs contains one or more references to
    // Kubernetes objects that contain TLS certificates of
    // the Certificate Authorities that can be used
    // as a trust anchor to validate the certificates presented by the client.
    //
    // A single CA certificate reference to a Kubernetes ConfigMap
    // has "Core" support.
    // Implementations MAY choose to support attaching multiple CA certificates to
    // a Listener, but this behavior is implementation-specific.
    //
    // Support: Core - An optional single reference to a single Kubernetes Secret
    // and to a single Kubernetes ConfigMap with the CA certificate in a key named `ca.crt`.
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
    // +kubebuilder:validation:MinItems=1
    CACertificateRefs []SecretObjectReference `json:"caCertificateRefs,omitempty"`
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
      frontEndValidation:
        caCertificateRefs:
        - kind: ConfigMap
          group: ""
          name: foo-example-com-ca-cert
```

## Deferred

This section highlights use cases that may be covered in a future iteration of this GEP

* Using system CA certificates as the trust anchor to validate the certificates presented by the client.
* Supporting a mode where validating client certficates is optional, useful for debugging and migrating to strict TLS.
* Supporting an optional `subjectAltNames` field within `ClientValidationContext` that can be used to specify one or more alternate names to verify the subject identity in the certificate presented by the client. This field falls under authorization and will be revisited when authorization is tackled as a whole in the project.
* Specifying the verification depth in the client certificates chain


## References

[TLS Handshake Protocol]: https://www.rfc-editor.org/rfc/rfc5246#section-7.4
[Certificate Path Validation]: https://www.rfc-editor.org/rfc/rfc5280#section-6
[GatewayTLSConfig]: https://gateway-api.sigs.k8s.io/references/spec/#gateway.networking.k8s.io/v1.GatewayTLSConfig
[BackendTLSPolicy]: https://gateway-api.sigs.k8s.io/api-types/backendtlspolicy/
[TLS Configuration GEP]: https://gateway-api.sigs.k8s.io/geps/gep-2907/
[Gateway API TLS Use Cases]: https://docs.google.com/document/d/17sctu2uMJtHmJTGtBi_awGB0YzoCLodtR6rUNmKMCs8/edit?pli=1#heading=h.cxuq8vo8pcxm
