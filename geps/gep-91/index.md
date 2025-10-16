# GEP-91: Client Certificate Validation for TLS terminating at the Gateway

* Issue: [#91](https://github.com/kubernetes-sigs/gateway-api/issues/91)
* Status: Experimental

(See definitions in [GEP States](../overview.md#gep-states).)

## TLDR

This GEP proposes a way to validate the TLS certificate presented by the frontend client to the server
(Gateway in this case) during a [TLS Handshake Protocol][].

## Connection coalescing security issue

Gateway API standard defines a `Listener` as "the concept of a logical endpoint where a Gateway accepts network connections." This statement is incorrect because once the connection is established (this applies to both HTTP and TLS traffic) it can be reused across multiple listeners sharing the same port. This might lead to bypassing client certificate validation configuration for a given `Listener`. Those concerns were raised in [GEP-3567](). To provide the best experience for gateway users and secure their applications, client certificate configuration needs to be introduced with finer granularity per-port (binding to TCP connection).

## Goals

* Define an API field to specify the CA Certificate within the Gateway configuration that can be used as a trust anchor to validate the certificates presented by the client.
This use case has been highlighted in the [TLS Configuration GEP][] under segment 1 and in the [Gateway API TLS Use Cases][] document under point 7.
* Introduce explicit client certificate validation modes that reflect common TLS behaviors (e.g., optional vs. required client certs)
* Ensure the configuration mitigates the authentication bypass risks associated with HTTP/2 connection coalesing as described in [GEP-3567](https://gateway-api.sigs.k8s.io/geps/gep-3567/#interaction-with-client-cert-validation).

## Non-Goals
* Define other fields that can be used to verify the client certificate such as the Certificate Hash.

### API

* Introduce new structs: `GatewayTLSConfig`, `TLSConfig`, `TLSPortConfig`, `FrontendTLSValidation` allowing for the definition of certificate validation used to authenticate the peer (frontend) in a TLS connection. A new `tls` field with gateway tls configuration will be added to the gateway object.
* `TLSConfig` will allow defining client certificate validation per port which will be applied to all Listeners matching this port. We might want to extend this struct with other tls configurations.
* `TLSPortConfig` will allow defining client certificate validation per port which will be applied to all Listeners matching this port.
* `GatewayTLSConfig` struct contains default and (optional) per port configuration. Default configuration will apply to all Listeners which are not matching per port override.
* This new field is separate from the existing [BackendTLSPolicy][] configuration. [BackendTLSPolicy][] controls TLS certificate validation for connections *from* the Gateway to the backend service.
This proposal adds the ability to validate the TLS certificate presented by the *client* connecting to the Gateway (the frontend).
These two validation mechanisms operate independently and can be used simultaneously.
* Introduce a `caCertificateRefs` field within `FrontendTLSValidation` that can be used to specify a list of CA Certificates that can be used as a trust anchor to validate the certificates presented by the client.
* Add a new `FrontendValidationModeType` enum within `FrontendTLSValidation` indicating how gateway should validate client certificates. As for now we support following values but it might change in the future:
  1) `AllowValidOnly` (Core Support)
  2) `AllowInsecureFallback` (Extended Support)

    `AllowInsecureFallback` mode indicates the gateway will accept connections even if the client certificate is not presented or fails verification.
	This approach delegates client authorization to the backend and introduce a significant security risk. It should be used in testing environments or
	on a temporary basis in non-testing environments.
    When `FrontendValidationModeType` is changed from `AllowValidOnly` to `AllowInsecureFallback` the `InsecureFrontendValidationMode` condition MUST be set to `True` with Reason `ConfigurationChanged` on gateway.
* Introduce a `ObjectReference` structure that can be used to specify `caCertificateRefs` references.

### Impact on listeners

This proposal removes frontendTLSValidation from Listener's TLS configuration and introduces gateways level per port configuration. This is a breaking change for exisitng implementation which uses this feature from Experimental API.
 Once gateway level TLS is configured (either by default or for a specific port), the TLS settings will apply to all existing and newly created Listeners serving HTTPS that match the configuration.

#### GO

```go
// ObjectReference identifies an API object including its namespace.
//
// The API object must be valid in the cluster; the Group and Kind must
// be registered in the cluster for this reference to be valid.
//
// References to objects with invalid Group and Kind are not valid, and must
// be rejected by the implementation, with appropriate Conditions set
// on the containing object.
type ObjectReference struct {
	// Group is the group of the referent. For example, "gateway.networking.k8s.io".
	// When unspecified or empty string, core API group is inferred.
	Group Group `json:"group"`

	// Kind is kind of the referent. For example "ConfigMap" or "Service".
	Kind Kind `json:"kind"`

	// Name is the name of the referent.
	Name ObjectName `json:"name"`

	// Namespace is the namespace of the referenced object. When unspecified, the local
	// namespace is inferred.
	//
	// Note that when a namespace different than the local namespace is specified,
	// a ReferenceGrant object is required in the referent namespace to allow that
	// namespace's owner to accept the reference. See the ReferenceGrant
	// documentation for details.
	//
	// Support: Core
	//
	// +optional
	Namespace *Namespace `json:"namespace,omitempty"`
}

// GatewayTLSConfig specifies frontend tls configuration for gateway.
type GatewayTLSConfig struct {
    // default specifies the default client certificate validation configuration
    // for all Listeners handling HTTPS traffic, unless a per-port configuration
    // is defined.
    //
    // support: Core
    //
    // +required
    // <gateway:experimental>
    Default FrontendTLSValidation `json:"default"`

    // PerPort specifies tls configuration assigned per port.
    // Per port configuration is optional. Once set this configuration overrides
    // the default configuration for all Listeners handling HTTPS traffic
    // that match this port.
    // Each override port requires a unique TLS configuration.
    //
    // support: Core
    //
    PerPort []TLSConfig `json:"PerPort,omitempty"`
}

// TLSConfig describes a TLS configuration. Currently, it stores only the client
// certificate validation configuration, but this may be extended in the future.
type TLSConfig struct {
	// FrontendValidation holds configuration information for validating the frontend (client).
	// Setting this field will result in mutual authentication when connecting to the gateway.
	// In browsers this may result in a dialog appearing
	// that requests a user to specify the client certificate.
	// The maximum depth of a certificate chain accepted in verification is Implementation specific.
	//
	// Support: Extended
	//
	// +required
	// <gateway:experimental>
	FrontendValidation FrontendTLSValidation `json:"frontendValidation"`
}

type TLSPortConfig struct {
	// The Port indicates the Port Number to which the TLS configuration will be
	// applied. This configuration will be applied to all Listeners handling HTTPS
	// traffic that match this port.
	//
	// Support: Core
	//
	// +required
	// <gateway:experimental>
	Port PortNumber `json:"port"`
	// TLS store the configuration that will be applied to all Listeners handling
	// HTTPS traffic and matching given port.
	TLS TLSConfig `json:"tls"`
}

// FrontendTLSValidation holds configuration information that can be used to validate
// the frontend initiating the TLS connection
type FrontendTLSValidation struct {
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
	// Support: Core - A single reference to a Kubernetes ConfigMap
	// with the CA certificate in a key named `ca.crt`.
	//
	// Support: Implementation-specific (More than one certificate in a ConfigMap with different keys or more than one reference, or other kinds of resources).
	//
	// References to a resource in a different namespace are invalid UNLESS there
	// is a ReferenceGrant in the target namespace that allows the certificate
	// to be attached. If a ReferenceGrant does not allow this reference, the
	// "ResolvedRefs" condition MUST be set to False for this listener with the
	// "RefNotPermitted" reason.
	//
	// +kubebuilder:validation:MaxItems=8
	// +kubebuilder:validation:MinItems=1
	CACertificateRefs []ObjectReference `json:"caCertificateRefs,omitempty"`

	// FrontendValidationMode defines the mode for validating the client certificate.
	// There are two possible modes:
	//
	// - AllowValidOnly: In this mode, the gateway will accept connections only if
	//   the client presents a valid certificate. This certificate must successfully
	//   pass validation against the CA certificates specified in `CACertificateRefs`.
	// - AllowInsecureFallback: In this mode, the gateway will accept connections
	//   even if the client certificate is not presented or fails verification.
	//
	//   This approach delegates client authorization to the backend and introduce
	//   a significant security risk. It should be used in testing environments or
	//   on a temporary basis in non-testing environments.
	//
	// Defaults to AllowValidOnly.
	//
	// Support: Core
	//
	// +optional
	// +kubebuilder:default=AllowValidOnly
	Mode FrontendValidationModeType `json:"mode,omitempty"`
}

// FrontendValidationModeType type defines how a Gateway validates client certificates.
//
// +kubebuilder:validation:Enum=AllowValidOnly;AllowInsecureFallback
type FrontendValidationModeType string

const (
	// AllowValidOnly indicates that a client certificate is required
	// during the TLS handshake and MUST pass validation.
	AllowValidOnly FrontendValidationModeType = "AllowValidOnly"

	// AllowInsecureFallback indicates that a client certificate may not be
	// presented during the handshake or the validation against CA certificates may fail.
	AllowInsecureFallback FrontendValidationModeType = "AllowInsecureFallback"
)

type GatewaySpec struct {
    ...
    // GatewayTLSConfig specifies frontend tls configuration for gateway.
	//
	// Support: Core
	//
	// +optional
	// <gateway:experimental>
	TLS *GatewayTLSConfig `json:"tls,omitempty"`
}

```

#### YAML

1. Setting default `frontendValidation` config.

```yaml
apiVersion: gateway.networking.k8s.io/v1beta1
kind: Gateway
metadata:
  name: client-validation-basic
spec:
  gatewayClassName: acme-lb
  tls:
    default:
    frontendValidation:
      caCertificateRefs:
      - kind: ConfigMap
        group: ""
        name: default-cert
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
  - name: bar-https
    protocol: HTTPS
    port: 8443
    hostname: bar.example.com
    tls:
      certificateRefs:
      - kind: Secret
        group: ""
        name: bar-example-com-cert
```

2. Setting default and per port `frontendValidation` configs.

```yaml
apiVersion: gateway.networking.k8s.io/v1beta1
kind: Gateway
metadata:
  name: client-validation-basic
spec:
  gatewayClassName: acme-lb
  tls:
    default:
      frontendValidation:
        caCertificateRefs:
          - kind: ConfigMap
            group: ""
            name: default-cert
        mode: AllowInsecureFallback
    perPort:
      - port: 443
        frontendValidation:
          caCertificateRefs:
          - kind: ConfigMap
            group: ""
            name: foo-example-com-ca-cert
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
  - name: bar-https
    protocol: HTTPS
    port: 8443
    hostname: bar.example.com
    tls:
      certificateRefs:
      - kind: Secret
        group: ""
        name: bar-example-com-cert
```

## Deferred

This section highlights use cases that may be covered in a future iteration of this GEP

* Using system CA certificates as the trust anchor to validate the certificates presented by the frontend client.
* Supporting an optional `subjectAltNames` field within `FrontendTLSValidation` that can be used to specify one or more alternate names to verify the subject identity in the certificate presented by the client. This field falls under Authorization, the initial focus here is on Client Authentication and will be revisited when Authorization is tackled as a whole in the project.
* Specifying the verification depth in the client certificate chain. This is being deferred because the default verification depth differs across implementations.

## Existing support in Implementations

This feature is already widely supported by implementations that conform to the Gateway API.
The table below summarizes current support. Please feel free to add any implementations that are missing.
This GEP aims to standardize this behavior as an official part of the upstream specification.

| Implementation                      | Support                                                                                                                                                     | Granularity                  | Inline vs Policy |
|------------------------------------|-------------------------------------------------------------------------------------------------------------------------------------------------------------|------------------------------|------------------|
| Acnodal EPIC                       |                                                                                                                                                             |                              |                  |
| Airlock Microgateway               | [SidecarGateway.spec.applications[].downstream.tls.clientCertificate](https://docs.airlock.com/microgateway/latest/index/api/crds/sidecar-gateway/v1alpha1/index.html#sidecargatewayspecapplicationsdownstreamtls) | Per Listener                 | Inline           |
| Amazon Elastic Kubernetes Service  |                                                                                                                                                             |                              |                  |
| Apache APISIX                      | [ApisixTls.Client.CASecret](https://apisix.apache.org/docs/ingress-controller/tutorials/mtls/#mutual-authentication)                                       | Per SNI                      |                  |
| Avi Kubernetes Operator            |                                                                                                                                                             |                              |                  |
| Azure Application Gateway for Containers | [FrontendTLSPolicy](https://learn.microsoft.com/en-us/azure/application-gateway/for-containers/api-specification-kubernetes#alb.networking.azure.io/v1.FrontendTLSPolicy) | Per Gateway & Per Listener  | Policy           |
| Cilium                             |                                                                                                                                                             |                              |                  |
| Contour                            | [HTTPProxy.Spec.VirtualHost.Tls.ClientValidation.CASecret](https://projectcontour.io/docs/v1.17.1/config/tls-termination/)                                 | Per SNI                      | Inline           |
| Easegress                          |                                                                                                                                                             |                              |                  |
| Emissary Ingress                   | [TlSContext.Spec.CASecret](https://www.getambassador.io/docs/emissary/latest/howtos/client-cert-validation)                                                | Per SNI                      | Policy           |
| Envoy Gateway                      | [ClientTrafficPolicy.Spec.TLS.ClientValidation](https://gateway.envoyproxy.io/docs/api/extension_types/#clientvalidationcontext)                           | Per Gateway & Per Listener  | Policy           |
| Flomesh Service Mesh               |                                                                                                                                                             |                              |                  |
| Gloo Gateway                       | [VirtualService.Spec.SSLConfig.SecretRef](https://docs.solo.io/gloo-edge/latest/guides/security/tls/server_tls/#configuring-downstream-mtls-in-a-virtual-service) | Per SNI                      | Inline           |
| Google Cloud Service Mesh          |                                                                                                                                                             |                              |                  |
| Google Kubernetes Engine           |                                                                                                                                                             |                              |                  |
| HAProxy Ingress                    |                                                                                                                                                             |                              |                  |
| HAProxy Kubernetes Ingress Controller | [ca-file](https://www.haproxy.com/documentation/haproxy-configuration-tutorials/security/authentication/client-certificate-authentication/#sidebar)      | Per SNI                      | Inline           |
| HashiCorp Consul                   | [file-system-certificate](https://developer.hashicorp.com/consul/docs/north-south/api-gateway/secure-traffic/encrypt)                                      | Per Listener                 | Policy           |
| Istio                              | [Gateway.Spec.Servers.TLS.Mode](https://istio.io/latest/docs/tasks/traffic-management/ingress/secure-ingress/#configure-a-mutual-tls-ingress-gateway)      | Per Server                  | Inline           |
| kgateway                           |                                                                                                                                                             |                              |                  |
| Kong Kubernetes Ingress Controller | [mTLS Plugin](https://docs.konghq.com/hub/kong-inc/mtls-auth/)                                                                                              | Per HTTP Proxy (Host/Port)   | Policy           |
| Kong Gateway Operator              | [mTLS Plugin](https://docs.konghq.com/hub/kong-inc/mtls-auth/)                                                                                              | Per HTTP Proxy (Host/Port)   | Policy           |
| Kuma                               |                                                                                                                                                             |                              |                  |
| Linkerd                            |                                                                                                                                                             |                              |                  |
| LiteSpeed Ingress Controller       |                                                                                                                                                             |                              |                  |
| LoxiLB                             |                                                                                                                                                             |                              |                  |
| NGINX Gateway Fabric               | [ingressMTLS](https://docs.nginx.com/nginx-ingress-controller/configuration/policy-resource/#ingressmtls)                                                  | Per Listener                 | Policy           |
| ngrok Kubernetes Operator          | [TrafficPolicy.Terminate-TLS.Config.MutualTLSCertificateAuthorities](https://ngrok.com/docs/traffic-policy/actions/terminate-tls/#configuration-reference) | Per Endpoint (Host:Port)     | Policy           |
| STUNner                            |                                                                                                                                                             |                              |                  |
| Traefik Proxy                      | [TLSOption.Spec.ClientAuth](https://doc.traefik.io/traefik/https/tls/#client-authentication-mtls)                                                          | Per EntryPoint               | Inline           |
| Tyk                                | [Enable Client Certificate](https://tyk.io/docs/basic-config-and-security/security/mutual-tls/client-mtls/#why-use-mutual-tls)                             | Per Gateway                  | Policy           |
| WSO2 APK                           | [Authentication.Spec.Default.AuthTypes.MTLS](https://apk.docs.wso2.com/en/latest/catalogs/crds/authentication_types/#dp.wso2.com/v1alpha2.MutualSSLConfig) | Per API                      | Policy           |
| Ingress-NGINX                      | [nginx.ingress.kubernetes.io/auth-tls-verify-client](https://kubernetes.github.io/ingress-nginx/user-guide/nginx-configuration/annotations/#client-certificate-authentication) | Per Ingress                  | Inline           |

## References

[TLS Handshake Protocol]: https://www.rfc-editor.org/rfc/rfc5246#section-7.4
[Certificate Path Validation]: https://www.rfc-editor.org/rfc/rfc5280#section-6
[BackendTLSPolicy]: ../../api-types/backendtlspolicy.md
[TLS Configuration GEP]: ../gep-2907/index.md
[Gateway API TLS Use Cases]: https://docs.google.com/document/d/17sctu2uMJtHmJTGtBi_awGB0YzoCLodtR6rUNmKMCs8/edit?pli=1#heading=h.cxuq8vo8pcxm
[GEP-3567]: https://gateway-api.sigs.k8s.io/geps/gep-3567/
