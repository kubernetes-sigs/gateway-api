# API Reference

## Packages
- [gateway.networking.k8s.io/v1](#gatewaynetworkingk8siov1)
- [gateway.networking.k8s.io/v1alpha2](#gatewaynetworkingk8siov1alpha2)
- [gateway.networking.k8s.io/v1alpha3](#gatewaynetworkingk8siov1alpha3)
- [gateway.networking.k8s.io/v1beta1](#gatewaynetworkingk8siov1beta1)


## gateway.networking.k8s.io/v1

Package v1 contains API Schema definitions for the gateway.networking.k8s.io
API group.


### Resource Types
- [GRPCRoute](#grpcroute)
- [Gateway](#gateway)
- [GatewayClass](#gatewayclass)
- [HTTPRoute](#httproute)



#### AbsoluteURI

_Underlying type:_ _string_

The AbsoluteURI MUST NOT be a relative URI, and it MUST follow the URI syntax and
encoding rules specified in RFC3986.  The AbsoluteURI MUST include both a
scheme (e.g., "http" or "spiffe") and a scheme-specific-part.  URIs that
include an authority MUST include a fully qualified domain name or
IP address as the host.
<gateway:util:excludeFromCRD> The below regex is taken from the regex section in RFC 3986 with a slight modification to enforce a full URI and not relative. </gateway:util:excludeFromCRD>

_Validation:_
- MaxLength: 253
- MinLength: 1
- Pattern: `^(([^:/?#]+):)(//([^/?#]*))([^?#]*)(\?([^#]*))?(#(.*))?`

_Appears in:_
- [HTTPCORSFilter](#httpcorsfilter)
- [SubjectAltName](#subjectaltname)



#### AddressType

_Underlying type:_ _string_

AddressType defines how a network address is represented as a text string.
This may take two possible forms:

* A predefined CamelCase string identifier (currently limited to `IPAddress` or `Hostname`)
* A domain-prefixed string identifier (like `acme.io/CustomAddressType`)

Values `IPAddress` and `Hostname` have Extended support.

The `NamedAddress` value has been deprecated in favor of implementation
specific domain-prefixed strings.

All other values, including domain-prefixed values have Implementation-specific support,
which are used in implementation-specific behaviors. Support for additional
predefined CamelCase identifiers may be added in future releases.

_Validation:_
- MaxLength: 253
- MinLength: 1
- Pattern: `^Hostname|IPAddress|NamedAddress|[a-z0-9]([-a-z0-9]*[a-z0-9])?(\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*\/[A-Za-z0-9\/\-._~%!$&'()*+,;=:]+$`

_Appears in:_
- [GatewaySpecAddress](#gatewayspecaddress)
- [GatewayStatusAddress](#gatewaystatusaddress)

| Field | Description |
| --- | --- |
| `IPAddress` | A textual representation of a numeric IP address. IPv4<br />addresses must be in dotted-decimal form. IPv6 addresses<br />must be in a standard IPv6 text representation<br />(see [RFC 5952](https://tools.ietf.org/html/rfc5952)).<br />This type is intended for specific addresses. Address ranges are not<br />supported (e.g. you cannot use a CIDR range like 127.0.0.0/24 as an<br />IPAddress).<br />Support: Extended<br /> |
| `Hostname` | A Hostname represents a DNS based ingress point. This is similar to the<br />corresponding hostname field in Kubernetes load balancer status. For<br />example, this concept may be used for cloud load balancers where a DNS<br />name is used to expose a load balancer.<br />Support: Extended<br /> |
| `NamedAddress` | A NamedAddress provides a way to reference a specific IP address by name.<br />For example, this may be a name or other unique identifier that refers<br />to a resource on a cloud provider such as a static IP.<br />The `NamedAddress` type has been deprecated in favor of implementation<br />specific domain-prefixed strings.<br />Support: Implementation-specific<br /> |


#### AllowedListeners



AllowedListeners defines which ListenerSets can be attached to this Gateway.



_Appears in:_
- [GatewaySpec](#gatewayspec)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `namespaces` _[ListenerNamespaces](#listenernamespaces)_ | Namespaces defines which namespaces ListenerSets can be attached to this Gateway.<br />While this feature is experimental, the default value is to allow no ListenerSets. | \{ from:None \} |  |


#### AllowedRoutes



AllowedRoutes defines which Routes may be attached to this Listener.



_Appears in:_
- [Listener](#listener)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `namespaces` _[RouteNamespaces](#routenamespaces)_ | Namespaces indicates namespaces from which Routes may be attached to this<br />Listener. This is restricted to the namespace of this Gateway by default.<br />Support: Core | \{ from:Same \} |  |
| `kinds` _[RouteGroupKind](#routegroupkind) array_ | Kinds specifies the groups and kinds of Routes that are allowed to bind<br />to this Gateway Listener. When unspecified or empty, the kinds of Routes<br />selected are determined using the Listener protocol.<br />A RouteGroupKind MUST correspond to kinds of Routes that are compatible<br />with the application protocol specified in the Listener's Protocol field.<br />If an implementation does not support or recognize this resource type, it<br />MUST set the "ResolvedRefs" condition to False for this Listener with the<br />"InvalidRouteKinds" reason.<br />Support: Core |  | MaxItems: 8 <br /> |


#### AnnotationKey

_Underlying type:_ _string_

AnnotationKey is the key of an annotation in Gateway API. This is used for
validation of maps such as TLS options. This matches the Kubernetes
"qualified name" validation that is used for annotations and other common
values.

Valid values include:

* example
* example.com
* example.com/path
* example.com/path.html

Invalid values include:

* example~ - "~" is an invalid character
* example.com. - cannot start or end with "."

_Validation:_
- MaxLength: 253
- MinLength: 1
- Pattern: `^([a-z0-9]([-a-z0-9]*[a-z0-9])?(\\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*/)?([A-Za-z0-9][-A-Za-z0-9_.]{0,61})?[A-Za-z0-9]$`

_Appears in:_
- [BackendTLSPolicySpec](#backendtlspolicyspec)
- [GatewayInfrastructure](#gatewayinfrastructure)
- [GatewayTLSConfig](#gatewaytlsconfig)



#### AnnotationValue

_Underlying type:_ _string_

AnnotationValue is the value of an annotation in Gateway API. This is used
for validation of maps such as TLS options. This roughly matches Kubernetes
annotation validation, although the length validation in that case is based
on the entire size of the annotations struct.

_Validation:_
- MaxLength: 4096
- MinLength: 0

_Appears in:_
- [BackendTLSPolicySpec](#backendtlspolicyspec)
- [GatewayInfrastructure](#gatewayinfrastructure)
- [GatewayTLSConfig](#gatewaytlsconfig)



#### BackendObjectReference



BackendObjectReference defines how an ObjectReference that is
specific to BackendRef. It includes a few additional fields and features
than a regular ObjectReference.

Note that when a namespace different than the local namespace is specified, a
ReferenceGrant object is required in the referent namespace to allow that
namespace's owner to accept the reference. See the ReferenceGrant
documentation for details.

The API object must be valid in the cluster; the Group and Kind must
be registered in the cluster for this reference to be valid.

References to objects with invalid Group and Kind are not valid, and must
be rejected by the implementation, with appropriate Conditions set
on the containing object.



_Appears in:_
- [BackendRef](#backendref)
- [GRPCBackendRef](#grpcbackendref)
- [HTTPBackendRef](#httpbackendref)
- [HTTPRequestMirrorFilter](#httprequestmirrorfilter)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `group` _[Group](#group)_ | Group is the group of the referent. For example, "gateway.networking.k8s.io".<br />When unspecified or empty string, core API group is inferred. |  | MaxLength: 253 <br />Pattern: `^$\|^[a-z0-9]([-a-z0-9]*[a-z0-9])?(\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*$` <br /> |
| `kind` _[Kind](#kind)_ | Kind is the Kubernetes resource kind of the referent. For example<br />"Service".<br />Defaults to "Service" when not specified.<br />ExternalName services can refer to CNAME DNS records that may live<br />outside of the cluster and as such are difficult to reason about in<br />terms of conformance. They also may not be safe to forward to (see<br />CVE-2021-25740 for more information). Implementations SHOULD NOT<br />support ExternalName Services.<br />Support: Core (Services with a type other than ExternalName)<br />Support: Implementation-specific (Services with type ExternalName) | Service | MaxLength: 63 <br />MinLength: 1 <br />Pattern: `^[a-zA-Z]([-a-zA-Z0-9]*[a-zA-Z0-9])?$` <br /> |
| `name` _[ObjectName](#objectname)_ | Name is the name of the referent. |  | MaxLength: 253 <br />MinLength: 1 <br /> |
| `namespace` _[Namespace](#namespace)_ | Namespace is the namespace of the backend. When unspecified, the local<br />namespace is inferred.<br />Note that when a namespace different than the local namespace is specified,<br />a ReferenceGrant object is required in the referent namespace to allow that<br />namespace's owner to accept the reference. See the ReferenceGrant<br />documentation for details.<br />Support: Core |  | MaxLength: 63 <br />MinLength: 1 <br />Pattern: `^[a-z0-9]([-a-z0-9]*[a-z0-9])?$` <br /> |
| `port` _[PortNumber](#portnumber)_ | Port specifies the destination port number to use for this resource.<br />Port is required when the referent is a Kubernetes Service. In this<br />case, the port number is the service port number, not the target port.<br />For other resources, destination port might be derived from the referent<br />resource or this field. |  | Maximum: 65535 <br />Minimum: 1 <br /> |


#### BackendRef



BackendRef defines how a Route should forward a request to a Kubernetes
resource.

Note that when a namespace different than the local namespace is specified, a
ReferenceGrant object is required in the referent namespace to allow that
namespace's owner to accept the reference. See the ReferenceGrant
documentation for details.

<gateway:experimental:description>

When the BackendRef points to a Kubernetes Service, implementations SHOULD
honor the appProtocol field if it is set for the target Service Port.

Implementations supporting appProtocol SHOULD recognize the Kubernetes
Standard Application Protocols defined in KEP-3726.

If a Service appProtocol isn't specified, an implementation MAY infer the
backend protocol through its own means. Implementations MAY infer the
protocol from the Route type referring to the backend Service.

If a Route is not able to send traffic to the backend using the specified
protocol then the backend is considered invalid. Implementations MUST set the
"ResolvedRefs" condition to "False" with the "UnsupportedProtocol" reason.

</gateway:experimental:description>

Note that when the BackendTLSPolicy object is enabled by the implementation,
there are some extra rules about validity to consider here. See the fields
where this struct is used for more information about the exact behavior.



_Appears in:_
- [GRPCBackendRef](#grpcbackendref)
- [HTTPBackendRef](#httpbackendref)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `group` _[Group](#group)_ | Group is the group of the referent. For example, "gateway.networking.k8s.io".<br />When unspecified or empty string, core API group is inferred. |  | MaxLength: 253 <br />Pattern: `^$\|^[a-z0-9]([-a-z0-9]*[a-z0-9])?(\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*$` <br /> |
| `kind` _[Kind](#kind)_ | Kind is the Kubernetes resource kind of the referent. For example<br />"Service".<br />Defaults to "Service" when not specified.<br />ExternalName services can refer to CNAME DNS records that may live<br />outside of the cluster and as such are difficult to reason about in<br />terms of conformance. They also may not be safe to forward to (see<br />CVE-2021-25740 for more information). Implementations SHOULD NOT<br />support ExternalName Services.<br />Support: Core (Services with a type other than ExternalName)<br />Support: Implementation-specific (Services with type ExternalName) | Service | MaxLength: 63 <br />MinLength: 1 <br />Pattern: `^[a-zA-Z]([-a-zA-Z0-9]*[a-zA-Z0-9])?$` <br /> |
| `name` _[ObjectName](#objectname)_ | Name is the name of the referent. |  | MaxLength: 253 <br />MinLength: 1 <br /> |
| `namespace` _[Namespace](#namespace)_ | Namespace is the namespace of the backend. When unspecified, the local<br />namespace is inferred.<br />Note that when a namespace different than the local namespace is specified,<br />a ReferenceGrant object is required in the referent namespace to allow that<br />namespace's owner to accept the reference. See the ReferenceGrant<br />documentation for details.<br />Support: Core |  | MaxLength: 63 <br />MinLength: 1 <br />Pattern: `^[a-z0-9]([-a-z0-9]*[a-z0-9])?$` <br /> |
| `port` _[PortNumber](#portnumber)_ | Port specifies the destination port number to use for this resource.<br />Port is required when the referent is a Kubernetes Service. In this<br />case, the port number is the service port number, not the target port.<br />For other resources, destination port might be derived from the referent<br />resource or this field. |  | Maximum: 65535 <br />Minimum: 1 <br /> |
| `weight` _integer_ | Weight specifies the proportion of requests forwarded to the referenced<br />backend. This is computed as weight/(sum of all weights in this<br />BackendRefs list). For non-zero values, there may be some epsilon from<br />the exact proportion defined here depending on the precision an<br />implementation supports. Weight is not a percentage and the sum of<br />weights does not need to equal 100.<br />If only one backend is specified and it has a weight greater than 0, 100%<br />of the traffic is forwarded to that backend. If weight is set to 0, no<br />traffic should be forwarded for this entry. If unspecified, weight<br />defaults to 1.<br />Support for this field varies based on the context where used. | 1 | Maximum: 1e+06 <br />Minimum: 0 <br /> |


#### CommonRouteSpec



CommonRouteSpec defines the common attributes that all Routes MUST include
within their spec.



_Appears in:_
- [GRPCRouteSpec](#grpcroutespec)
- [HTTPRouteSpec](#httproutespec)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `parentRefs` _[ParentReference](#parentreference) array_ | ParentRefs references the resources (usually Gateways) that a Route wants<br />to be attached to. Note that the referenced parent resource needs to<br />allow this for the attachment to be complete. For Gateways, that means<br />the Gateway needs to allow attachment from Routes of this kind and<br />namespace. For Services, that means the Service must either be in the same<br />namespace for a "producer" route, or the mesh implementation must support<br />and allow "consumer" routes for the referenced Service. ReferenceGrant is<br />not applicable for governing ParentRefs to Services - it is not possible to<br />create a "producer" route for a Service in a different namespace from the<br />Route.<br />There are two kinds of parent resources with "Core" support:<br />* Gateway (Gateway conformance profile)<br />* Service (Mesh conformance profile, ClusterIP Services only)<br />This API may be extended in the future to support additional kinds of parent<br />resources.<br />ParentRefs must be _distinct_. This means either that:<br />* They select different objects.  If this is the case, then parentRef<br />  entries are distinct. In terms of fields, this means that the<br />  multi-part key defined by `group`, `kind`, `namespace`, and `name` must<br />  be unique across all parentRef entries in the Route.<br />* They do not select different objects, but for each optional field used,<br />  each ParentRef that selects the same object must set the same set of<br />  optional fields to different values. If one ParentRef sets a<br />  combination of optional fields, all must set the same combination.<br />Some examples:<br />* If one ParentRef sets `sectionName`, all ParentRefs referencing the<br />  same object must also set `sectionName`.<br />* If one ParentRef sets `port`, all ParentRefs referencing the same<br />  object must also set `port`.<br />* If one ParentRef sets `sectionName` and `port`, all ParentRefs<br />  referencing the same object must also set `sectionName` and `port`.<br />It is possible to separately reference multiple distinct objects that may<br />be collapsed by an implementation. For example, some implementations may<br />choose to merge compatible Gateway Listeners together. If that is the<br />case, the list of routes attached to those resources should also be<br />merged.<br />Note that for ParentRefs that cross namespace boundaries, there are specific<br />rules. Cross-namespace references are only valid if they are explicitly<br />allowed by something in the namespace they are referring to. For example,<br />Gateway has the AllowedRoutes field, and ReferenceGrant provides a<br />generic way to enable other kinds of cross-namespace reference.<br /><gateway:experimental:description><br />ParentRefs from a Route to a Service in the same namespace are "producer"<br />routes, which apply default routing rules to inbound connections from<br />any namespace to the Service.<br />ParentRefs from a Route to a Service in a different namespace are<br />"consumer" routes, and these routing rules are only applied to outbound<br />connections originating from the same namespace as the Route, for which<br />the intended destination of the connections are a Service targeted as a<br />ParentRef of the Route.<br /></gateway:experimental:description><br /><gateway:standard:validation:XValidation:message="sectionName must be specified when parentRefs includes 2 or more references to the same parent",rule="self.all(p1, self.all(p2, p1.group == p2.group && p1.kind == p2.kind && p1.name == p2.name && (((!has(p1.__namespace__) \|\| p1.__namespace__ == '') && (!has(p2.__namespace__) \|\| p2.__namespace__ == '')) \|\| (has(p1.__namespace__) && has(p2.__namespace__) && p1.__namespace__ == p2.__namespace__ )) ? ((!has(p1.sectionName) \|\| p1.sectionName == '') == (!has(p2.sectionName) \|\| p2.sectionName == '')) : true))"><br /><gateway:standard:validation:XValidation:message="sectionName must be unique when parentRefs includes 2 or more references to the same parent",rule="self.all(p1, self.exists_one(p2, p1.group == p2.group && p1.kind == p2.kind && p1.name == p2.name && (((!has(p1.__namespace__) \|\| p1.__namespace__ == '') && (!has(p2.__namespace__) \|\| p2.__namespace__ == '')) \|\| (has(p1.__namespace__) && has(p2.__namespace__) && p1.__namespace__ == p2.__namespace__ )) && (((!has(p1.sectionName) \|\| p1.sectionName == '') && (!has(p2.sectionName) \|\| p2.sectionName == '')) \|\| (has(p1.sectionName) && has(p2.sectionName) && p1.sectionName == p2.sectionName))))"><br /><gateway:experimental:validation:XValidation:message="sectionName or port must be specified when parentRefs includes 2 or more references to the same parent",rule="self.all(p1, self.all(p2, p1.group == p2.group && p1.kind == p2.kind && p1.name == p2.name && (((!has(p1.__namespace__) \|\| p1.__namespace__ == '') && (!has(p2.__namespace__) \|\| p2.__namespace__ == '')) \|\| (has(p1.__namespace__) && has(p2.__namespace__) && p1.__namespace__ == p2.__namespace__)) ? ((!has(p1.sectionName) \|\| p1.sectionName == '') == (!has(p2.sectionName) \|\| p2.sectionName == '') && (!has(p1.port) \|\| p1.port == 0) == (!has(p2.port) \|\| p2.port == 0)): true))"><br /><gateway:experimental:validation:XValidation:message="sectionName or port must be unique when parentRefs includes 2 or more references to the same parent",rule="self.all(p1, self.exists_one(p2, p1.group == p2.group && p1.kind == p2.kind && p1.name == p2.name && (((!has(p1.__namespace__) \|\| p1.__namespace__ == '') && (!has(p2.__namespace__) \|\| p2.__namespace__ == '')) \|\| (has(p1.__namespace__) && has(p2.__namespace__) && p1.__namespace__ == p2.__namespace__ )) && (((!has(p1.sectionName) \|\| p1.sectionName == '') && (!has(p2.sectionName) \|\| p2.sectionName == '')) \|\| ( has(p1.sectionName) && has(p2.sectionName) && p1.sectionName == p2.sectionName)) && (((!has(p1.port) \|\| p1.port == 0) && (!has(p2.port) \|\| p2.port == 0)) \|\| (has(p1.port) && has(p2.port) && p1.port == p2.port))))"> |  | MaxItems: 32 <br /> |


#### CookieConfig



CookieConfig defines the configuration for cookie-based session persistence.



_Appears in:_
- [SessionPersistence](#sessionpersistence)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `lifetimeType` _[CookieLifetimeType](#cookielifetimetype)_ | LifetimeType specifies whether the cookie has a permanent or<br />session-based lifetime. A permanent cookie persists until its<br />specified expiry time, defined by the Expires or Max-Age cookie<br />attributes, while a session cookie is deleted when the current<br />session ends.<br />When set to "Permanent", AbsoluteTimeout indicates the<br />cookie's lifetime via the Expires or Max-Age cookie attributes<br />and is required.<br />When set to "Session", AbsoluteTimeout indicates the<br />absolute lifetime of the cookie tracked by the gateway and<br />is optional.<br />Defaults to "Session".<br />Support: Core for "Session" type<br />Support: Extended for "Permanent" type | Session | Enum: [Permanent Session] <br /> |


#### CookieLifetimeType

_Underlying type:_ _string_



_Validation:_
- Enum: [Permanent Session]

_Appears in:_
- [CookieConfig](#cookieconfig)

| Field | Description |
| --- | --- |
| `Session` | SessionCookieLifetimeType specifies the type for a session<br />cookie.<br />Support: Core<br /> |
| `Permanent` | PermanentCookieLifetimeType specifies the type for a permanent<br />cookie.<br />Support: Extended<br /> |


#### Duration

_Underlying type:_ _string_

Duration is a string value representing a duration in time. The format is as specified
in GEP-2257, a strict subset of the syntax parsed by Golang time.ParseDuration.

_Validation:_
- Pattern: `^([0-9]{1,5}(h|m|s|ms)){1,4}$`

_Appears in:_
- [HTTPRouteRetry](#httprouteretry)
- [HTTPRouteTimeouts](#httproutetimeouts)
- [SessionPersistence](#sessionpersistence)



#### FeatureName

_Underlying type:_ _string_

FeatureName is used to describe distinct features that are covered by
conformance tests.



_Appears in:_
- [SupportedFeature](#supportedfeature)



#### Fraction







_Appears in:_
- [HTTPRequestMirrorFilter](#httprequestmirrorfilter)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `numerator` _integer_ |  |  | Minimum: 0 <br /> |
| `denominator` _integer_ |  | 100 | Minimum: 1 <br /> |


#### FromNamespaces

_Underlying type:_ _string_

FromNamespaces specifies namespace from which Routes/ListenerSets may be attached to a
Gateway.



_Appears in:_
- [ListenerNamespaces](#listenernamespaces)
- [RouteNamespaces](#routenamespaces)

| Field | Description |
| --- | --- |
| `All` | Routes/ListenerSets in all namespaces may be attached to this Gateway.<br /> |
| `Selector` | Only Routes/ListenerSets in namespaces selected by the selector may be attached to<br />this Gateway.<br /> |
| `Same` | Only Routes/ListenerSets in the same namespace as the Gateway may be attached to this<br />Gateway.<br /> |
| `None` | No Routes/ListenerSets may be attached to this Gateway.<br /> |


#### FrontendTLSValidation



FrontendTLSValidation holds configuration information that can be used to validate
the frontend initiating the TLS connection



_Appears in:_
- [GatewayTLSConfig](#gatewaytlsconfig)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `caCertificateRefs` _[ObjectReference](#objectreference) array_ | CACertificateRefs contains one or more references to<br />Kubernetes objects that contain TLS certificates of<br />the Certificate Authorities that can be used<br />as a trust anchor to validate the certificates presented by the client.<br />A single CA certificate reference to a Kubernetes ConfigMap<br />has "Core" support.<br />Implementations MAY choose to support attaching multiple CA certificates to<br />a Listener, but this behavior is implementation-specific.<br />Support: Core - A single reference to a Kubernetes ConfigMap<br />with the CA certificate in a key named `ca.crt`.<br />Support: Implementation-specific (More than one reference, or other kinds<br />of resources).<br />References to a resource in a different namespace are invalid UNLESS there<br />is a ReferenceGrant in the target namespace that allows the certificate<br />to be attached. If a ReferenceGrant does not allow this reference, the<br />"ResolvedRefs" condition MUST be set to False for this listener with the<br />"RefNotPermitted" reason. |  | MaxItems: 8 <br />MinItems: 1 <br /> |


#### GRPCBackendRef



GRPCBackendRef defines how a GRPCRoute forwards a gRPC request.

Note that when a namespace different than the local namespace is specified, a
ReferenceGrant object is required in the referent namespace to allow that
namespace's owner to accept the reference. See the ReferenceGrant
documentation for details.

<gateway:experimental:description>

When the BackendRef points to a Kubernetes Service, implementations SHOULD
honor the appProtocol field if it is set for the target Service Port.

Implementations supporting appProtocol SHOULD recognize the Kubernetes
Standard Application Protocols defined in KEP-3726.

If a Service appProtocol isn't specified, an implementation MAY infer the
backend protocol through its own means. Implementations MAY infer the
protocol from the Route type referring to the backend Service.

If a Route is not able to send traffic to the backend using the specified
protocol then the backend is considered invalid. Implementations MUST set the
"ResolvedRefs" condition to "False" with the "UnsupportedProtocol" reason.

</gateway:experimental:description>



_Appears in:_
- [GRPCRouteRule](#grpcrouterule)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `group` _[Group](#group)_ | Group is the group of the referent. For example, "gateway.networking.k8s.io".<br />When unspecified or empty string, core API group is inferred. |  | MaxLength: 253 <br />Pattern: `^$\|^[a-z0-9]([-a-z0-9]*[a-z0-9])?(\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*$` <br /> |
| `kind` _[Kind](#kind)_ | Kind is the Kubernetes resource kind of the referent. For example<br />"Service".<br />Defaults to "Service" when not specified.<br />ExternalName services can refer to CNAME DNS records that may live<br />outside of the cluster and as such are difficult to reason about in<br />terms of conformance. They also may not be safe to forward to (see<br />CVE-2021-25740 for more information). Implementations SHOULD NOT<br />support ExternalName Services.<br />Support: Core (Services with a type other than ExternalName)<br />Support: Implementation-specific (Services with type ExternalName) | Service | MaxLength: 63 <br />MinLength: 1 <br />Pattern: `^[a-zA-Z]([-a-zA-Z0-9]*[a-zA-Z0-9])?$` <br /> |
| `name` _[ObjectName](#objectname)_ | Name is the name of the referent. |  | MaxLength: 253 <br />MinLength: 1 <br /> |
| `namespace` _[Namespace](#namespace)_ | Namespace is the namespace of the backend. When unspecified, the local<br />namespace is inferred.<br />Note that when a namespace different than the local namespace is specified,<br />a ReferenceGrant object is required in the referent namespace to allow that<br />namespace's owner to accept the reference. See the ReferenceGrant<br />documentation for details.<br />Support: Core |  | MaxLength: 63 <br />MinLength: 1 <br />Pattern: `^[a-z0-9]([-a-z0-9]*[a-z0-9])?$` <br /> |
| `port` _[PortNumber](#portnumber)_ | Port specifies the destination port number to use for this resource.<br />Port is required when the referent is a Kubernetes Service. In this<br />case, the port number is the service port number, not the target port.<br />For other resources, destination port might be derived from the referent<br />resource or this field. |  | Maximum: 65535 <br />Minimum: 1 <br /> |
| `weight` _integer_ | Weight specifies the proportion of requests forwarded to the referenced<br />backend. This is computed as weight/(sum of all weights in this<br />BackendRefs list). For non-zero values, there may be some epsilon from<br />the exact proportion defined here depending on the precision an<br />implementation supports. Weight is not a percentage and the sum of<br />weights does not need to equal 100.<br />If only one backend is specified and it has a weight greater than 0, 100%<br />of the traffic is forwarded to that backend. If weight is set to 0, no<br />traffic should be forwarded for this entry. If unspecified, weight<br />defaults to 1.<br />Support for this field varies based on the context where used. | 1 | Maximum: 1e+06 <br />Minimum: 0 <br /> |
| `filters` _[GRPCRouteFilter](#grpcroutefilter) array_ | Filters defined at this level MUST be executed if and only if the<br />request is being forwarded to the backend defined here.<br />Support: Implementation-specific (For broader support of filters, use the<br />Filters field in GRPCRouteRule.) |  | MaxItems: 16 <br /> |


#### GRPCHeaderMatch



GRPCHeaderMatch describes how to select a gRPC route by matching gRPC request
headers.



_Appears in:_
- [GRPCRouteMatch](#grpcroutematch)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `type` _[GRPCHeaderMatchType](#grpcheadermatchtype)_ | Type specifies how to match against the value of the header. | Exact | Enum: [Exact RegularExpression] <br /> |
| `name` _[GRPCHeaderName](#grpcheadername)_ | Name is the name of the gRPC Header to be matched.<br />If multiple entries specify equivalent header names, only the first<br />entry with an equivalent name MUST be considered for a match. Subsequent<br />entries with an equivalent header name MUST be ignored. Due to the<br />case-insensitivity of header names, "foo" and "Foo" are considered<br />equivalent. |  | MaxLength: 256 <br />MinLength: 1 <br />Pattern: `^[A-Za-z0-9!#$%&'*+\-.^_\x60\|~]+$` <br /> |
| `value` _string_ | Value is the value of the gRPC Header to be matched. |  | MaxLength: 4096 <br />MinLength: 1 <br /> |


#### GRPCHeaderMatchType

_Underlying type:_ _string_

GRPCHeaderMatchType specifies the semantics of how GRPC header values should
be compared. Valid GRPCHeaderMatchType values, along with their conformance
levels, are:

* "Exact" - Core
* "RegularExpression" - Implementation Specific

Note that new values may be added to this enum in future releases of the API,
implementations MUST ensure that unknown values will not cause a crash.

Unknown values here MUST result in the implementation setting the Accepted
Condition for the Route to `status: False`, with a Reason of
`UnsupportedValue`.

_Validation:_
- Enum: [Exact RegularExpression]

_Appears in:_
- [GRPCHeaderMatch](#grpcheadermatch)

| Field | Description |
| --- | --- |
| `Exact` |  |
| `RegularExpression` |  |


#### GRPCHeaderName

_Underlying type:_ _[HeaderName](#headername)_



_Validation:_
- MaxLength: 256
- MinLength: 1
- Pattern: `^[A-Za-z0-9!#$%&'*+\-.^_\x60|~]+$`

_Appears in:_
- [GRPCHeaderMatch](#grpcheadermatch)



#### GRPCMethodMatch



GRPCMethodMatch describes how to select a gRPC route by matching the gRPC
request service and/or method.

At least one of Service and Method MUST be a non-empty string.



_Appears in:_
- [GRPCRouteMatch](#grpcroutematch)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `type` _[GRPCMethodMatchType](#grpcmethodmatchtype)_ | Type specifies how to match against the service and/or method.<br />Support: Core (Exact with service and method specified)<br />Support: Implementation-specific (Exact with method specified but no service specified)<br />Support: Implementation-specific (RegularExpression) | Exact | Enum: [Exact RegularExpression] <br /> |
| `service` _string_ | Value of the service to match against. If left empty or omitted, will<br />match any service.<br />At least one of Service and Method MUST be a non-empty string. |  | MaxLength: 1024 <br /> |
| `method` _string_ | Value of the method to match against. If left empty or omitted, will<br />match all services.<br />At least one of Service and Method MUST be a non-empty string. |  | MaxLength: 1024 <br /> |


#### GRPCMethodMatchType

_Underlying type:_ _string_

MethodMatchType specifies the semantics of how gRPC methods and services are compared.
Valid MethodMatchType values, along with their conformance levels, are:

* "Exact" - Core
* "RegularExpression" - Implementation Specific

Exact methods MUST be syntactically valid:

- Must not contain `/` character

_Validation:_
- Enum: [Exact RegularExpression]

_Appears in:_
- [GRPCMethodMatch](#grpcmethodmatch)

| Field | Description |
| --- | --- |
| `Exact` | Matches the method or service exactly and with case sensitivity.<br /> |
| `RegularExpression` | Matches if the method or service matches the given regular expression with<br />case sensitivity.<br />Since `"RegularExpression"` has implementation-specific conformance,<br />implementations can support POSIX, PCRE, RE2 or any other regular expression<br />dialect.<br />Please read the implementation's documentation to determine the supported<br />dialect.<br /> |


#### GRPCRoute



GRPCRoute provides a way to route gRPC requests. This includes the capability
to match requests by hostname, gRPC service, gRPC method, or HTTP/2 header.
Filters can be used to specify additional processing steps. Backends specify
where matching requests will be routed.

GRPCRoute falls under extended support within the Gateway API. Within the
following specification, the word "MUST" indicates that an implementation
supporting GRPCRoute must conform to the indicated requirement, but an
implementation not supporting this route type need not follow the requirement
unless explicitly indicated.

Implementations supporting `GRPCRoute` with the `HTTPS` `ProtocolType` MUST
accept HTTP/2 connections without an initial upgrade from HTTP/1.1, i.e. via
ALPN. If the implementation does not support this, then it MUST set the
"Accepted" condition to "False" for the affected listener with a reason of
"UnsupportedProtocol".  Implementations MAY also accept HTTP/2 connections
with an upgrade from HTTP/1.

Implementations supporting `GRPCRoute` with the `HTTP` `ProtocolType` MUST
support HTTP/2 over cleartext TCP (h2c,
https://www.rfc-editor.org/rfc/rfc7540#section-3.1) without an initial
upgrade from HTTP/1.1, i.e. with prior knowledge
(https://www.rfc-editor.org/rfc/rfc7540#section-3.4). If the implementation
does not support this, then it MUST set the "Accepted" condition to "False"
for the affected listener with a reason of "UnsupportedProtocol".
Implementations MAY also accept HTTP/2 connections with an upgrade from
HTTP/1, i.e. without prior knowledge.



_Appears in:_
- [GRPCRoute](#grpcroute)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `apiVersion` _string_ | `gateway.networking.k8s.io/v1` | | |
| `kind` _string_ | `GRPCRoute` | | |
| `metadata` _[ObjectMeta](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.32/#objectmeta-v1-meta)_ | Refer to Kubernetes API documentation for fields of `metadata`. |  |  |
| `spec` _[GRPCRouteSpec](#grpcroutespec)_ | Spec defines the desired state of GRPCRoute. |  |  |
| `status` _[GRPCRouteStatus](#grpcroutestatus)_ | Status defines the current state of GRPCRoute. |  |  |


#### GRPCRouteFilter



GRPCRouteFilter defines processing steps that must be completed during the
request or response lifecycle. GRPCRouteFilters are meant as an extension
point to express processing that may be done in Gateway implementations. Some
examples include request or response modification, implementing
authentication strategies, rate-limiting, and traffic shaping. API
guarantee/conformance is defined based on the type of the filter.



_Appears in:_
- [GRPCBackendRef](#grpcbackendref)
- [GRPCRouteRule](#grpcrouterule)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `type` _[GRPCRouteFilterType](#grpcroutefiltertype)_ | Type identifies the type of filter to apply. As with other API fields,<br />types are classified into three conformance levels:<br />- Core: Filter types and their corresponding configuration defined by<br />  "Support: Core" in this package, e.g. "RequestHeaderModifier". All<br />  implementations supporting GRPCRoute MUST support core filters.<br />- Extended: Filter types and their corresponding configuration defined by<br />  "Support: Extended" in this package, e.g. "RequestMirror". Implementers<br />  are encouraged to support extended filters.<br />- Implementation-specific: Filters that are defined and supported by specific vendors.<br />  In the future, filters showing convergence in behavior across multiple<br />  implementations will be considered for inclusion in extended or core<br />  conformance levels. Filter-specific configuration for such filters<br />  is specified using the ExtensionRef field. `Type` MUST be set to<br />  "ExtensionRef" for custom filters.<br />Implementers are encouraged to define custom implementation types to<br />extend the core API with implementation-specific behavior.<br />If a reference to a custom filter type cannot be resolved, the filter<br />MUST NOT be skipped. Instead, requests that would have been processed by<br />that filter MUST receive a HTTP error response.<br /><gateway:experimental:validation:Enum=ResponseHeaderModifier;RequestHeaderModifier;RequestMirror;ExtensionRef> |  | Enum: [ResponseHeaderModifier RequestHeaderModifier RequestMirror ExtensionRef] <br /> |
| `requestHeaderModifier` _[HTTPHeaderFilter](#httpheaderfilter)_ | RequestHeaderModifier defines a schema for a filter that modifies request<br />headers.<br />Support: Core |  |  |
| `responseHeaderModifier` _[HTTPHeaderFilter](#httpheaderfilter)_ | ResponseHeaderModifier defines a schema for a filter that modifies response<br />headers.<br />Support: Extended |  |  |
| `requestMirror` _[HTTPRequestMirrorFilter](#httprequestmirrorfilter)_ | RequestMirror defines a schema for a filter that mirrors requests.<br />Requests are sent to the specified destination, but responses from<br />that destination are ignored.<br />This filter can be used multiple times within the same rule. Note that<br />not all implementations will be able to support mirroring to multiple<br />backends.<br />Support: Extended |  |  |
| `extensionRef` _[LocalObjectReference](#localobjectreference)_ | ExtensionRef is an optional, implementation-specific extension to the<br />"filter" behavior.  For example, resource "myroutefilter" in group<br />"networking.example.net"). ExtensionRef MUST NOT be used for core and<br />extended filters.<br />Support: Implementation-specific<br />This filter can be used multiple times within the same rule. |  |  |


#### GRPCRouteFilterType

_Underlying type:_ _string_

GRPCRouteFilterType identifies a type of GRPCRoute filter.



_Appears in:_
- [GRPCRouteFilter](#grpcroutefilter)

| Field | Description |
| --- | --- |
| `RequestHeaderModifier` | GRPCRouteFilterRequestHeaderModifier can be used to add or remove a gRPC<br />header from a gRPC request before it is sent to the upstream target.<br />Support in GRPCRouteRule: Core<br />Support in GRPCBackendRef: Extended<br /> |
| `ResponseHeaderModifier` | GRPCRouteFilterRequestHeaderModifier can be used to add or remove a gRPC<br />header from a gRPC response before it is sent to the client.<br />Support in GRPCRouteRule: Core<br />Support in GRPCBackendRef: Extended<br /> |
| `RequestMirror` | GRPCRouteFilterRequestMirror can be used to mirror gRPC requests to a<br />different backend. The responses from this backend MUST be ignored by<br />the Gateway.<br />Support in GRPCRouteRule: Extended<br />Support in GRPCBackendRef: Extended<br /> |
| `ExtensionRef` | GRPCRouteFilterExtensionRef should be used for configuring custom<br />gRPC filters.<br />Support in GRPCRouteRule: Implementation-specific<br />Support in GRPCBackendRef: Implementation-specific<br /> |


#### GRPCRouteMatch



GRPCRouteMatch defines the predicate used to match requests to a given
action. Multiple match types are ANDed together, i.e. the match will
evaluate to true only if all conditions are satisfied.

For example, the match below will match a gRPC request only if its service
is `foo` AND it contains the `version: v1` header:

```
matches:
  - method:
    type: Exact
    service: "foo"
    headers:
  - name: "version"
    value "v1"

```



_Appears in:_
- [GRPCRouteRule](#grpcrouterule)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `method` _[GRPCMethodMatch](#grpcmethodmatch)_ | Method specifies a gRPC request service/method matcher. If this field is<br />not specified, all services and methods will match. |  |  |
| `headers` _[GRPCHeaderMatch](#grpcheadermatch) array_ | Headers specifies gRPC request header matchers. Multiple match values are<br />ANDed together, meaning, a request MUST match all the specified headers<br />to select the route. |  | MaxItems: 16 <br /> |


#### GRPCRouteRule



GRPCRouteRule defines the semantics for matching a gRPC request based on
conditions (matches), processing it (filters), and forwarding the request to
an API object (backendRefs).



_Appears in:_
- [GRPCRouteSpec](#grpcroutespec)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `name` _[SectionName](#sectionname)_ | Name is the name of the route rule. This name MUST be unique within a Route if it is set.<br />Support: Extended<br /><gateway:experimental> |  | MaxLength: 253 <br />MinLength: 1 <br />Pattern: `^[a-z0-9]([-a-z0-9]*[a-z0-9])?(\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*$` <br /> |
| `matches` _[GRPCRouteMatch](#grpcroutematch) array_ | Matches define conditions used for matching the rule against incoming<br />gRPC requests. Each match is independent, i.e. this rule will be matched<br />if **any** one of the matches is satisfied.<br />For example, take the following matches configuration:<br />```<br />matches:<br />- method:<br />    service: foo.bar<br />  headers:<br />    values:<br />      version: 2<br />- method:<br />    service: foo.bar.v2<br />```<br />For a request to match against this rule, it MUST satisfy<br />EITHER of the two conditions:<br />- service of foo.bar AND contains the header `version: 2`<br />- service of foo.bar.v2<br />See the documentation for GRPCRouteMatch on how to specify multiple<br />match conditions to be ANDed together.<br />If no matches are specified, the implementation MUST match every gRPC request.<br />Proxy or Load Balancer routing configuration generated from GRPCRoutes<br />MUST prioritize rules based on the following criteria, continuing on<br />ties. Merging MUST not be done between GRPCRoutes and HTTPRoutes.<br />Precedence MUST be given to the rule with the largest number of:<br />* Characters in a matching non-wildcard hostname.<br />* Characters in a matching hostname.<br />* Characters in a matching service.<br />* Characters in a matching method.<br />* Header matches.<br />If ties still exist across multiple Routes, matching precedence MUST be<br />determined in order of the following criteria, continuing on ties:<br />* The oldest Route based on creation timestamp.<br />* The Route appearing first in alphabetical order by<br />  "\{namespace\}/\{name\}".<br />If ties still exist within the Route that has been given precedence,<br />matching precedence MUST be granted to the first matching rule meeting<br />the above criteria. |  | MaxItems: 64 <br /> |
| `filters` _[GRPCRouteFilter](#grpcroutefilter) array_ | Filters define the filters that are applied to requests that match<br />this rule.<br />The effects of ordering of multiple behaviors are currently unspecified.<br />This can change in the future based on feedback during the alpha stage.<br />Conformance-levels at this level are defined based on the type of filter:<br />- ALL core filters MUST be supported by all implementations that support<br />  GRPCRoute.<br />- Implementers are encouraged to support extended filters.<br />- Implementation-specific custom filters have no API guarantees across<br />  implementations.<br />Specifying the same filter multiple times is not supported unless explicitly<br />indicated in the filter.<br />If an implementation cannot support a combination of filters, it must clearly<br />document that limitation. In cases where incompatible or unsupported<br />filters are specified and cause the `Accepted` condition to be set to status<br />`False`, implementations may use the `IncompatibleFilters` reason to specify<br />this configuration error.<br />Support: Core |  | MaxItems: 16 <br /> |
| `backendRefs` _[GRPCBackendRef](#grpcbackendref) array_ | BackendRefs defines the backend(s) where matching requests should be<br />sent.<br />Failure behavior here depends on how many BackendRefs are specified and<br />how many are invalid.<br />If *all* entries in BackendRefs are invalid, and there are also no filters<br />specified in this route rule, *all* traffic which matches this rule MUST<br />receive an `UNAVAILABLE` status.<br />See the GRPCBackendRef definition for the rules about what makes a single<br />GRPCBackendRef invalid.<br />When a GRPCBackendRef is invalid, `UNAVAILABLE` statuses MUST be returned for<br />requests that would have otherwise been routed to an invalid backend. If<br />multiple backends are specified, and some are invalid, the proportion of<br />requests that would otherwise have been routed to an invalid backend<br />MUST receive an `UNAVAILABLE` status.<br />For example, if two backends are specified with equal weights, and one is<br />invalid, 50 percent of traffic MUST receive an `UNAVAILABLE` status.<br />Implementations may choose how that 50 percent is determined.<br />Support: Core for Kubernetes Service<br />Support: Implementation-specific for any other resource<br />Support for weight: Core |  | MaxItems: 16 <br /> |
| `sessionPersistence` _[SessionPersistence](#sessionpersistence)_ | SessionPersistence defines and configures session persistence<br />for the route rule.<br />Support: Extended<br /><gateway:experimental> |  |  |


#### GRPCRouteSpec



GRPCRouteSpec defines the desired state of GRPCRoute



_Appears in:_
- [GRPCRoute](#grpcroute)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `parentRefs` _[ParentReference](#parentreference) array_ | ParentRefs references the resources (usually Gateways) that a Route wants<br />to be attached to. Note that the referenced parent resource needs to<br />allow this for the attachment to be complete. For Gateways, that means<br />the Gateway needs to allow attachment from Routes of this kind and<br />namespace. For Services, that means the Service must either be in the same<br />namespace for a "producer" route, or the mesh implementation must support<br />and allow "consumer" routes for the referenced Service. ReferenceGrant is<br />not applicable for governing ParentRefs to Services - it is not possible to<br />create a "producer" route for a Service in a different namespace from the<br />Route.<br />There are two kinds of parent resources with "Core" support:<br />* Gateway (Gateway conformance profile)<br />* Service (Mesh conformance profile, ClusterIP Services only)<br />This API may be extended in the future to support additional kinds of parent<br />resources.<br />ParentRefs must be _distinct_. This means either that:<br />* They select different objects.  If this is the case, then parentRef<br />  entries are distinct. In terms of fields, this means that the<br />  multi-part key defined by `group`, `kind`, `namespace`, and `name` must<br />  be unique across all parentRef entries in the Route.<br />* They do not select different objects, but for each optional field used,<br />  each ParentRef that selects the same object must set the same set of<br />  optional fields to different values. If one ParentRef sets a<br />  combination of optional fields, all must set the same combination.<br />Some examples:<br />* If one ParentRef sets `sectionName`, all ParentRefs referencing the<br />  same object must also set `sectionName`.<br />* If one ParentRef sets `port`, all ParentRefs referencing the same<br />  object must also set `port`.<br />* If one ParentRef sets `sectionName` and `port`, all ParentRefs<br />  referencing the same object must also set `sectionName` and `port`.<br />It is possible to separately reference multiple distinct objects that may<br />be collapsed by an implementation. For example, some implementations may<br />choose to merge compatible Gateway Listeners together. If that is the<br />case, the list of routes attached to those resources should also be<br />merged.<br />Note that for ParentRefs that cross namespace boundaries, there are specific<br />rules. Cross-namespace references are only valid if they are explicitly<br />allowed by something in the namespace they are referring to. For example,<br />Gateway has the AllowedRoutes field, and ReferenceGrant provides a<br />generic way to enable other kinds of cross-namespace reference.<br /><gateway:experimental:description><br />ParentRefs from a Route to a Service in the same namespace are "producer"<br />routes, which apply default routing rules to inbound connections from<br />any namespace to the Service.<br />ParentRefs from a Route to a Service in a different namespace are<br />"consumer" routes, and these routing rules are only applied to outbound<br />connections originating from the same namespace as the Route, for which<br />the intended destination of the connections are a Service targeted as a<br />ParentRef of the Route.<br /></gateway:experimental:description><br /><gateway:standard:validation:XValidation:message="sectionName must be specified when parentRefs includes 2 or more references to the same parent",rule="self.all(p1, self.all(p2, p1.group == p2.group && p1.kind == p2.kind && p1.name == p2.name && (((!has(p1.__namespace__) \|\| p1.__namespace__ == '') && (!has(p2.__namespace__) \|\| p2.__namespace__ == '')) \|\| (has(p1.__namespace__) && has(p2.__namespace__) && p1.__namespace__ == p2.__namespace__ )) ? ((!has(p1.sectionName) \|\| p1.sectionName == '') == (!has(p2.sectionName) \|\| p2.sectionName == '')) : true))"><br /><gateway:standard:validation:XValidation:message="sectionName must be unique when parentRefs includes 2 or more references to the same parent",rule="self.all(p1, self.exists_one(p2, p1.group == p2.group && p1.kind == p2.kind && p1.name == p2.name && (((!has(p1.__namespace__) \|\| p1.__namespace__ == '') && (!has(p2.__namespace__) \|\| p2.__namespace__ == '')) \|\| (has(p1.__namespace__) && has(p2.__namespace__) && p1.__namespace__ == p2.__namespace__ )) && (((!has(p1.sectionName) \|\| p1.sectionName == '') && (!has(p2.sectionName) \|\| p2.sectionName == '')) \|\| (has(p1.sectionName) && has(p2.sectionName) && p1.sectionName == p2.sectionName))))"><br /><gateway:experimental:validation:XValidation:message="sectionName or port must be specified when parentRefs includes 2 or more references to the same parent",rule="self.all(p1, self.all(p2, p1.group == p2.group && p1.kind == p2.kind && p1.name == p2.name && (((!has(p1.__namespace__) \|\| p1.__namespace__ == '') && (!has(p2.__namespace__) \|\| p2.__namespace__ == '')) \|\| (has(p1.__namespace__) && has(p2.__namespace__) && p1.__namespace__ == p2.__namespace__)) ? ((!has(p1.sectionName) \|\| p1.sectionName == '') == (!has(p2.sectionName) \|\| p2.sectionName == '') && (!has(p1.port) \|\| p1.port == 0) == (!has(p2.port) \|\| p2.port == 0)): true))"><br /><gateway:experimental:validation:XValidation:message="sectionName or port must be unique when parentRefs includes 2 or more references to the same parent",rule="self.all(p1, self.exists_one(p2, p1.group == p2.group && p1.kind == p2.kind && p1.name == p2.name && (((!has(p1.__namespace__) \|\| p1.__namespace__ == '') && (!has(p2.__namespace__) \|\| p2.__namespace__ == '')) \|\| (has(p1.__namespace__) && has(p2.__namespace__) && p1.__namespace__ == p2.__namespace__ )) && (((!has(p1.sectionName) \|\| p1.sectionName == '') && (!has(p2.sectionName) \|\| p2.sectionName == '')) \|\| ( has(p1.sectionName) && has(p2.sectionName) && p1.sectionName == p2.sectionName)) && (((!has(p1.port) \|\| p1.port == 0) && (!has(p2.port) \|\| p2.port == 0)) \|\| (has(p1.port) && has(p2.port) && p1.port == p2.port))))"> |  | MaxItems: 32 <br /> |
| `hostnames` _[Hostname](#hostname) array_ | Hostnames defines a set of hostnames to match against the GRPC<br />Host header to select a GRPCRoute to process the request. This matches<br />the RFC 1123 definition of a hostname with 2 notable exceptions:<br />1. IPs are not allowed.<br />2. A hostname may be prefixed with a wildcard label (`*.`). The wildcard<br />   label MUST appear by itself as the first label.<br />If a hostname is specified by both the Listener and GRPCRoute, there<br />MUST be at least one intersecting hostname for the GRPCRoute to be<br />attached to the Listener. For example:<br />* A Listener with `test.example.com` as the hostname matches GRPCRoutes<br />  that have either not specified any hostnames, or have specified at<br />  least one of `test.example.com` or `*.example.com`.<br />* A Listener with `*.example.com` as the hostname matches GRPCRoutes<br />  that have either not specified any hostnames or have specified at least<br />  one hostname that matches the Listener hostname. For example,<br />  `test.example.com` and `*.example.com` would both match. On the other<br />  hand, `example.com` and `test.example.net` would not match.<br />Hostnames that are prefixed with a wildcard label (`*.`) are interpreted<br />as a suffix match. That means that a match for `*.example.com` would match<br />both `test.example.com`, and `foo.test.example.com`, but not `example.com`.<br />If both the Listener and GRPCRoute have specified hostnames, any<br />GRPCRoute hostnames that do not match the Listener hostname MUST be<br />ignored. For example, if a Listener specified `*.example.com`, and the<br />GRPCRoute specified `test.example.com` and `test.example.net`,<br />`test.example.net` MUST NOT be considered for a match.<br />If both the Listener and GRPCRoute have specified hostnames, and none<br />match with the criteria above, then the GRPCRoute MUST NOT be accepted by<br />the implementation. The implementation MUST raise an 'Accepted' Condition<br />with a status of `False` in the corresponding RouteParentStatus.<br />If a Route (A) of type HTTPRoute or GRPCRoute is attached to a<br />Listener and that listener already has another Route (B) of the other<br />type attached and the intersection of the hostnames of A and B is<br />non-empty, then the implementation MUST accept exactly one of these two<br />routes, determined by the following criteria, in order:<br />* The oldest Route based on creation timestamp.<br />* The Route appearing first in alphabetical order by<br />  "\{namespace\}/\{name\}".<br />The rejected Route MUST raise an 'Accepted' condition with a status of<br />'False' in the corresponding RouteParentStatus.<br />Support: Core |  | MaxItems: 16 <br />MaxLength: 253 <br />MinLength: 1 <br />Pattern: `^(\*\.)?[a-z0-9]([-a-z0-9]*[a-z0-9])?(\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*$` <br /> |
| `rules` _[GRPCRouteRule](#grpcrouterule) array_ | Rules are a list of GRPC matchers, filters and actions.<br /><gateway:experimental:validation:XValidation:message="Rule name must be unique within the route",rule="self.all(l1, !has(l1.name) \|\| self.exists_one(l2, has(l2.name) && l1.name == l2.name))"> |  | MaxItems: 16 <br /> |


#### GRPCRouteStatus



GRPCRouteStatus defines the observed state of GRPCRoute.



_Appears in:_
- [GRPCRoute](#grpcroute)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `parents` _[RouteParentStatus](#routeparentstatus) array_ | Parents is a list of parent resources (usually Gateways) that are<br />associated with the route, and the status of the route with respect to<br />each parent. When this route attaches to a parent, the controller that<br />manages the parent must add an entry to this list when the controller<br />first sees the route and should update the entry as appropriate when the<br />route or gateway is modified.<br />Note that parent references that cannot be resolved by an implementation<br />of this API will not be added to this list. Implementations of this API<br />can only populate Route status for the Gateways/parent resources they are<br />responsible for.<br />A maximum of 32 Gateways will be represented in this list. An empty list<br />means the route has not been attached to any Gateway. |  | MaxItems: 32 <br /> |


#### Gateway



Gateway represents an instance of a service-traffic handling infrastructure
by binding Listeners to a set of IP addresses.



_Appears in:_
- [Gateway](#gateway)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `apiVersion` _string_ | `gateway.networking.k8s.io/v1` | | |
| `kind` _string_ | `Gateway` | | |
| `metadata` _[ObjectMeta](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.32/#objectmeta-v1-meta)_ | Refer to Kubernetes API documentation for fields of `metadata`. |  |  |
| `spec` _[GatewaySpec](#gatewayspec)_ | Spec defines the desired state of Gateway. |  |  |
| `status` _[GatewayStatus](#gatewaystatus)_ | Status defines the current state of Gateway. | \{ conditions:[map[lastTransitionTime:1970-01-01T00:00:00Z message:Waiting for controller reason:Pending status:Unknown type:Accepted] map[lastTransitionTime:1970-01-01T00:00:00Z message:Waiting for controller reason:Pending status:Unknown type:Programmed]] \} |  |


#### GatewayBackendTLS



GatewayBackendTLS describes backend TLS configuration for gateway.



_Appears in:_
- [GatewaySpec](#gatewayspec)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `clientCertificateRef` _[SecretObjectReference](#secretobjectreference)_ | ClientCertificateRef is a reference to an object that contains a Client<br />Certificate and the associated private key.<br />References to a resource in different namespace are invalid UNLESS there<br />is a ReferenceGrant in the target namespace that allows the certificate<br />to be attached. If a ReferenceGrant does not allow this reference, the<br />"ResolvedRefs" condition MUST be set to False for this listener with the<br />"RefNotPermitted" reason.<br />ClientCertificateRef can reference to standard Kubernetes resources, i.e.<br />Secret, or implementation-specific custom resources.<br />This setting can be overridden on the service level by use of BackendTLSPolicy.<br />Support: Core<br /><gateway:experimental> |  |  |


#### GatewayClass



GatewayClass describes a class of Gateways available to the user for creating
Gateway resources.

It is recommended that this resource be used as a template for Gateways. This
means that a Gateway is based on the state of the GatewayClass at the time it
was created and changes to the GatewayClass or associated parameters are not
propagated down to existing Gateways. This recommendation is intended to
limit the blast radius of changes to GatewayClass or associated parameters.
If implementations choose to propagate GatewayClass changes to existing
Gateways, that MUST be clearly documented by the implementation.

Whenever one or more Gateways are using a GatewayClass, implementations SHOULD
add the `gateway-exists-finalizer.gateway.networking.k8s.io` finalizer on the
associated GatewayClass. This ensures that a GatewayClass associated with a
Gateway is not deleted while in use.

GatewayClass is a Cluster level resource.



_Appears in:_
- [GatewayClass](#gatewayclass)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `apiVersion` _string_ | `gateway.networking.k8s.io/v1` | | |
| `kind` _string_ | `GatewayClass` | | |
| `metadata` _[ObjectMeta](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.32/#objectmeta-v1-meta)_ | Refer to Kubernetes API documentation for fields of `metadata`. |  |  |
| `spec` _[GatewayClassSpec](#gatewayclassspec)_ | Spec defines the desired state of GatewayClass. |  |  |
| `status` _[GatewayClassStatus](#gatewayclassstatus)_ | Status defines the current state of GatewayClass.<br />Implementations MUST populate status on all GatewayClass resources which<br />specify their controller name. | \{ conditions:[map[lastTransitionTime:1970-01-01T00:00:00Z message:Waiting for controller reason:Pending status:Unknown type:Accepted]] \} |  |






#### GatewayClassSpec



GatewayClassSpec reflects the configuration of a class of Gateways.



_Appears in:_
- [GatewayClass](#gatewayclass)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `controllerName` _[GatewayController](#gatewaycontroller)_ | ControllerName is the name of the controller that is managing Gateways of<br />this class. The value of this field MUST be a domain prefixed path.<br />Example: "example.net/gateway-controller".<br />This field is not mutable and cannot be empty.<br />Support: Core |  | MaxLength: 253 <br />MinLength: 1 <br />Pattern: `^[a-z0-9]([-a-z0-9]*[a-z0-9])?(\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*\/[A-Za-z0-9\/\-._~%!$&'()*+,;=:]+$` <br /> |
| `parametersRef` _[ParametersReference](#parametersreference)_ | ParametersRef is a reference to a resource that contains the configuration<br />parameters corresponding to the GatewayClass. This is optional if the<br />controller does not require any additional configuration.<br />ParametersRef can reference a standard Kubernetes resource, i.e. ConfigMap,<br />or an implementation-specific custom resource. The resource can be<br />cluster-scoped or namespace-scoped.<br />If the referent cannot be found, refers to an unsupported kind, or when<br />the data within that resource is malformed, the GatewayClass SHOULD be<br />rejected with the "Accepted" status condition set to "False" and an<br />"InvalidParameters" reason.<br />A Gateway for this GatewayClass may provide its own `parametersRef`. When both are specified,<br />the merging behavior is implementation specific.<br />It is generally recommended that GatewayClass provides defaults that can be overridden by a Gateway.<br />Support: Implementation-specific |  |  |
| `description` _string_ | Description helps describe a GatewayClass with more details. |  | MaxLength: 64 <br /> |


#### GatewayClassStatus



GatewayClassStatus is the current status for the GatewayClass.



_Appears in:_
- [GatewayClass](#gatewayclass)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `conditions` _[Condition](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.32/#condition-v1-meta) array_ | Conditions is the current status from the controller for<br />this GatewayClass.<br />Controllers should prefer to publish conditions using values<br />of GatewayClassConditionType for the type of each Condition. | [map[lastTransitionTime:1970-01-01T00:00:00Z message:Waiting for controller reason:Pending status:Unknown type:Accepted]] | MaxItems: 8 <br /> |
| `supportedFeatures` _[SupportedFeature](#supportedfeature) array_ | SupportedFeatures is the set of features the GatewayClass support.<br />It MUST be sorted in ascending alphabetical order by the Name key.<br /><gateway:experimental> |  | MaxItems: 64 <br /> |






#### GatewayController

_Underlying type:_ _string_

GatewayController is the name of a Gateway API controller. It must be a
domain prefixed path.

Valid values include:

* "example.com/bar"

Invalid values include:

* "example.com" - must include path
* "foo.example.com" - must include path

_Validation:_
- MaxLength: 253
- MinLength: 1
- Pattern: `^[a-z0-9]([-a-z0-9]*[a-z0-9])?(\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*\/[A-Za-z0-9\/\-._~%!$&'()*+,;=:]+$`

_Appears in:_
- [GatewayClassSpec](#gatewayclassspec)
- [RouteParentStatus](#routeparentstatus)



#### GatewayInfrastructure



GatewayInfrastructure defines infrastructure level attributes about a Gateway instance.



_Appears in:_
- [GatewaySpec](#gatewayspec)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `labels` _object (keys:[LabelKey](#labelkey), values:[LabelValue](#labelvalue))_ | Labels that SHOULD be applied to any resources created in response to this Gateway.<br />For implementations creating other Kubernetes objects, this should be the `metadata.labels` field on resources.<br />For other implementations, this refers to any relevant (implementation specific) "labels" concepts.<br />An implementation may chose to add additional implementation-specific labels as they see fit.<br />If an implementation maps these labels to Pods, or any other resource that would need to be recreated when labels<br />change, it SHOULD clearly warn about this behavior in documentation.<br />Support: Extended |  | MaxProperties: 8 <br /> |
| `annotations` _object (keys:[AnnotationKey](#annotationkey), values:[AnnotationValue](#annotationvalue))_ | Annotations that SHOULD be applied to any resources created in response to this Gateway.<br />For implementations creating other Kubernetes objects, this should be the `metadata.annotations` field on resources.<br />For other implementations, this refers to any relevant (implementation specific) "annotations" concepts.<br />An implementation may chose to add additional implementation-specific annotations as they see fit.<br />Support: Extended |  | MaxProperties: 8 <br /> |
| `parametersRef` _[LocalParametersReference](#localparametersreference)_ | ParametersRef is a reference to a resource that contains the configuration<br />parameters corresponding to the Gateway. This is optional if the<br />controller does not require any additional configuration.<br />This follows the same semantics as GatewayClass's `parametersRef`, but on a per-Gateway basis<br />The Gateway's GatewayClass may provide its own `parametersRef`. When both are specified,<br />the merging behavior is implementation specific.<br />It is generally recommended that GatewayClass provides defaults that can be overridden by a Gateway.<br />If the referent cannot be found, refers to an unsupported kind, or when<br />the data within that resource is malformed, the Gateway SHOULD be<br />rejected with the "Accepted" status condition set to "False" and an<br />"InvalidParameters" reason.<br />Support: Implementation-specific |  |  |


#### GatewaySpec



GatewaySpec defines the desired state of Gateway.

Not all possible combinations of options specified in the Spec are
valid. Some invalid configurations can be caught synchronously via CRD
validation, but there are many cases that will require asynchronous
signaling via the GatewayStatus block.



_Appears in:_
- [Gateway](#gateway)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `gatewayClassName` _[ObjectName](#objectname)_ | GatewayClassName used for this Gateway. This is the name of a<br />GatewayClass resource. |  | MaxLength: 253 <br />MinLength: 1 <br /> |
| `listeners` _[Listener](#listener) array_ | Listeners associated with this Gateway. Listeners define<br />logical endpoints that are bound on this Gateway's addresses.<br />At least one Listener MUST be specified.<br />## Distinct Listeners<br />Each Listener in a set of Listeners (for example, in a single Gateway)<br />MUST be _distinct_, in that a traffic flow MUST be able to be assigned to<br />exactly one listener. (This section uses "set of Listeners" rather than<br />"Listeners in a single Gateway" because implementations MAY merge configuration<br />from multiple Gateways onto a single data plane, and these rules _also_<br />apply in that case).<br />Practically, this means that each listener in a set MUST have a unique<br />combination of Port, Protocol, and, if supported by the protocol, Hostname.<br />Some combinations of port, protocol, and TLS settings are considered<br />Core support and MUST be supported by implementations based on the objects<br />they support:<br />HTTPRoute<br />1. HTTPRoute, Port: 80, Protocol: HTTP<br />2. HTTPRoute, Port: 443, Protocol: HTTPS, TLS Mode: Terminate, TLS keypair provided<br />TLSRoute<br />1. TLSRoute, Port: 443, Protocol: TLS, TLS Mode: Passthrough<br />"Distinct" Listeners have the following property:<br />**The implementation can match inbound requests to a single distinct<br />Listener**.<br />When multiple Listeners share values for fields (for<br />example, two Listeners with the same Port value), the implementation<br />can match requests to only one of the Listeners using other<br />Listener fields.<br />When multiple listeners have the same value for the Protocol field, then<br />each of the Listeners with matching Protocol values MUST have different<br />values for other fields.<br />The set of fields that MUST be different for a Listener differs per protocol.<br />The following rules define the rules for what fields MUST be considered for<br />Listeners to be distinct with each protocol currently defined in the<br />Gateway API spec.<br />The set of listeners that all share a protocol value MUST have _different_<br />values for _at least one_ of these fields to be distinct:<br />* **HTTP, HTTPS, TLS**: Port, Hostname<br />* **TCP, UDP**: Port<br />One **very** important rule to call out involves what happens when an<br />implementation:<br />* Supports TCP protocol Listeners, as well as HTTP, HTTPS, or TLS protocol<br />  Listeners, and<br />* sees HTTP, HTTPS, or TLS protocols with the same `port` as one with TCP<br />  Protocol.<br />In this case all the Listeners that share a port with the<br />TCP Listener are not distinct and so MUST NOT be accepted.<br />If an implementation does not support TCP Protocol Listeners, then the<br />previous rule does not apply, and the TCP Listeners SHOULD NOT be<br />accepted.<br />Note that the `tls` field is not used for determining if a listener is distinct, because<br />Listeners that _only_ differ on TLS config will still conflict in all cases.<br />### Listeners that are distinct only by Hostname<br />When the Listeners are distinct based only on Hostname, inbound request<br />hostnames MUST match from the most specific to least specific Hostname<br />values to choose the correct Listener and its associated set of Routes.<br />Exact matches MUST be processed before wildcard matches, and wildcard<br />matches MUST be processed before fallback (empty Hostname value)<br />matches. For example, `"foo.example.com"` takes precedence over<br />`"*.example.com"`, and `"*.example.com"` takes precedence over `""`.<br />Additionally, if there are multiple wildcard entries, more specific<br />wildcard entries must be processed before less specific wildcard entries.<br />For example, `"*.foo.example.com"` takes precedence over `"*.example.com"`.<br />The precise definition here is that the higher the number of dots in the<br />hostname to the right of the wildcard character, the higher the precedence.<br />The wildcard character will match any number of characters _and dots_ to<br />the left, however, so `"*.example.com"` will match both<br />`"foo.bar.example.com"` _and_ `"bar.example.com"`.<br />## Handling indistinct Listeners<br />If a set of Listeners contains Listeners that are not distinct, then those<br />Listeners are _Conflicted_, and the implementation MUST set the "Conflicted"<br />condition in the Listener Status to "True".<br />The words "indistinct" and "conflicted" are considered equivalent for the<br />purpose of this documentation.<br />Implementations MAY choose to accept a Gateway with some Conflicted<br />Listeners only if they only accept the partial Listener set that contains<br />no Conflicted Listeners.<br />Specifically, an implementation MAY accept a partial Listener set subject to<br />the following rules:<br />* The implementation MUST NOT pick one conflicting Listener as the winner.<br />  ALL indistinct Listeners must not be accepted for processing.<br />* At least one distinct Listener MUST be present, or else the Gateway effectively<br />  contains _no_ Listeners, and must be rejected from processing as a whole.<br />The implementation MUST set a "ListenersNotValid" condition on the<br />Gateway Status when the Gateway contains Conflicted Listeners whether or<br />not they accept the Gateway. That Condition SHOULD clearly<br />indicate in the Message which Listeners are conflicted, and which are<br />Accepted. Additionally, the Listener status for those listeners SHOULD<br />indicate which Listeners are conflicted and not Accepted.<br />## General Listener behavior<br />Note that, for all distinct Listeners, requests SHOULD match at most one Listener.<br />For example, if Listeners are defined for "foo.example.com" and "*.example.com", a<br />request to "foo.example.com" SHOULD only be routed using routes attached<br />to the "foo.example.com" Listener (and not the "*.example.com" Listener).<br />This concept is known as "Listener Isolation", and it is an Extended feature<br />of Gateway API. Implementations that do not support Listener Isolation MUST<br />clearly document this, and MUST NOT claim support for the<br />`GatewayHTTPListenerIsolation` feature.<br />Implementations that _do_ support Listener Isolation SHOULD claim support<br />for the Extended `GatewayHTTPListenerIsolation` feature and pass the associated<br />conformance tests.<br />## Compatible Listeners<br />A Gateway's Listeners are considered _compatible_ if:<br />1. They are distinct.<br />2. The implementation can serve them in compliance with the Addresses<br />   requirement that all Listeners are available on all assigned<br />   addresses.<br />Compatible combinations in Extended support are expected to vary across<br />implementations. A combination that is compatible for one implementation<br />may not be compatible for another.<br />For example, an implementation that cannot serve both TCP and UDP listeners<br />on the same address, or cannot mix HTTPS and generic TLS listens on the same port<br />would not consider those cases compatible, even though they are distinct.<br />Implementations MAY merge separate Gateways onto a single set of<br />Addresses if all Listeners across all Gateways are compatible.<br />In a future release the MinItems=1 requirement MAY be dropped.<br />Support: Core |  | MaxItems: 64 <br />MinItems: 1 <br /> |
| `addresses` _[GatewaySpecAddress](#gatewayspecaddress) array_ | Addresses requested for this Gateway. This is optional and behavior can<br />depend on the implementation. If a value is set in the spec and the<br />requested address is invalid or unavailable, the implementation MUST<br />indicate this in the associated entry in GatewayStatus.Addresses.<br />The Addresses field represents a request for the address(es) on the<br />"outside of the Gateway", that traffic bound for this Gateway will use.<br />This could be the IP address or hostname of an external load balancer or<br />other networking infrastructure, or some other address that traffic will<br />be sent to.<br />If no Addresses are specified, the implementation MAY schedule the<br />Gateway in an implementation-specific manner, assigning an appropriate<br />set of Addresses.<br />The implementation MUST bind all Listeners to every GatewayAddress that<br />it assigns to the Gateway and add a corresponding entry in<br />GatewayStatus.Addresses.<br />Support: Extended<br /><gateway:validateIPAddress> |  | MaxItems: 16 <br /> |
| `infrastructure` _[GatewayInfrastructure](#gatewayinfrastructure)_ | Infrastructure defines infrastructure level attributes about this Gateway instance.<br />Support: Extended |  |  |
| `backendTLS` _[GatewayBackendTLS](#gatewaybackendtls)_ | BackendTLS configures TLS settings for when this Gateway is connecting to<br />backends with TLS.<br />Support: Core<br /><gateway:experimental> |  |  |
| `allowedListeners` _[AllowedListeners](#allowedlisteners)_ | AllowedListeners defines which ListenerSets can be attached to this Gateway.<br />While this feature is experimental, the default value is to allow no ListenerSets.<br /><gateway:experimental> |  |  |


#### GatewaySpecAddress



GatewaySpecAddress describes an address that can be bound to a Gateway.



_Appears in:_
- [GatewaySpec](#gatewayspec)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `type` _[AddressType](#addresstype)_ | Type of the address. | IPAddress | MaxLength: 253 <br />MinLength: 1 <br />Pattern: `^Hostname\|IPAddress\|NamedAddress\|[a-z0-9]([-a-z0-9]*[a-z0-9])?(\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*\/[A-Za-z0-9\/\-._~%!$&'()*+,;=:]+$` <br /> |
| `value` _string_ | When a value is unspecified, an implementation SHOULD automatically<br />assign an address matching the requested type if possible.<br />If an implementation does not support an empty value, they MUST set the<br />"Programmed" condition in status to False with a reason of "AddressNotAssigned".<br />Examples: `1.2.3.4`, `128::1`, `my-ip-address`. |  | MaxLength: 253 <br /> |


#### GatewayStatus



GatewayStatus defines the observed state of Gateway.



_Appears in:_
- [Gateway](#gateway)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `addresses` _[GatewayStatusAddress](#gatewaystatusaddress) array_ | Addresses lists the network addresses that have been bound to the<br />Gateway.<br />This list may differ from the addresses provided in the spec under some<br />conditions:<br />  * no addresses are specified, all addresses are dynamically assigned<br />  * a combination of specified and dynamic addresses are assigned<br />  * a specified address was unusable (e.g. already in use)<br /><gateway:validateIPAddress> |  | MaxItems: 16 <br /> |
| `conditions` _[Condition](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.32/#condition-v1-meta) array_ | Conditions describe the current conditions of the Gateway.<br />Implementations should prefer to express Gateway conditions<br />using the `GatewayConditionType` and `GatewayConditionReason`<br />constants so that operators and tools can converge on a common<br />vocabulary to describe Gateway state.<br />Known condition types are:<br />* "Accepted"<br />* "Programmed"<br />* "Ready" | [map[lastTransitionTime:1970-01-01T00:00:00Z message:Waiting for controller reason:Pending status:Unknown type:Accepted] map[lastTransitionTime:1970-01-01T00:00:00Z message:Waiting for controller reason:Pending status:Unknown type:Programmed]] | MaxItems: 8 <br /> |
| `listeners` _[ListenerStatus](#listenerstatus) array_ | Listeners provide status for each unique listener port defined in the Spec. |  | MaxItems: 64 <br /> |


#### GatewayStatusAddress



GatewayStatusAddress describes a network address that is bound to a Gateway.



_Appears in:_
- [GatewayStatus](#gatewaystatus)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `type` _[AddressType](#addresstype)_ | Type of the address. | IPAddress | MaxLength: 253 <br />MinLength: 1 <br />Pattern: `^Hostname\|IPAddress\|NamedAddress\|[a-z0-9]([-a-z0-9]*[a-z0-9])?(\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*\/[A-Za-z0-9\/\-._~%!$&'()*+,;=:]+$` <br /> |
| `value` _string_ | Value of the address. The validity of the values will depend<br />on the type and support by the controller.<br />Examples: `1.2.3.4`, `128::1`, `my-ip-address`. |  | MaxLength: 253 <br />MinLength: 1 <br /> |


#### GatewayTLSConfig



GatewayTLSConfig describes a TLS configuration.



_Appears in:_
- [Listener](#listener)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `mode` _[TLSModeType](#tlsmodetype)_ | Mode defines the TLS behavior for the TLS session initiated by the client.<br />There are two possible modes:<br />- Terminate: The TLS session between the downstream client and the<br />  Gateway is terminated at the Gateway. This mode requires certificates<br />  to be specified in some way, such as populating the certificateRefs<br />  field.<br />- Passthrough: The TLS session is NOT terminated by the Gateway. This<br />  implies that the Gateway can't decipher the TLS stream except for<br />  the ClientHello message of the TLS protocol. The certificateRefs field<br />  is ignored in this mode.<br />Support: Core | Terminate | Enum: [Terminate Passthrough] <br /> |
| `certificateRefs` _[SecretObjectReference](#secretobjectreference) array_ | CertificateRefs contains a series of references to Kubernetes objects that<br />contains TLS certificates and private keys. These certificates are used to<br />establish a TLS handshake for requests that match the hostname of the<br />associated listener.<br />A single CertificateRef to a Kubernetes Secret has "Core" support.<br />Implementations MAY choose to support attaching multiple certificates to<br />a Listener, but this behavior is implementation-specific.<br />References to a resource in different namespace are invalid UNLESS there<br />is a ReferenceGrant in the target namespace that allows the certificate<br />to be attached. If a ReferenceGrant does not allow this reference, the<br />"ResolvedRefs" condition MUST be set to False for this listener with the<br />"RefNotPermitted" reason.<br />This field is required to have at least one element when the mode is set<br />to "Terminate" (default) and is optional otherwise.<br />CertificateRefs can reference to standard Kubernetes resources, i.e.<br />Secret, or implementation-specific custom resources.<br />Support: Core - A single reference to a Kubernetes Secret of type kubernetes.io/tls<br />Support: Implementation-specific (More than one reference or other resource types) |  | MaxItems: 64 <br /> |
| `frontendValidation` _[FrontendTLSValidation](#frontendtlsvalidation)_ | FrontendValidation holds configuration information for validating the frontend (client).<br />Setting this field will require clients to send a client certificate<br />required for validation during the TLS handshake. In browsers this may result in a dialog appearing<br />that requests a user to specify the client certificate.<br />The maximum depth of a certificate chain accepted in verification is Implementation specific.<br />Support: Extended<br /><gateway:experimental> |  |  |
| `options` _object (keys:[AnnotationKey](#annotationkey), values:[AnnotationValue](#annotationvalue))_ | Options are a list of key/value pairs to enable extended TLS<br />configuration for each implementation. For example, configuring the<br />minimum TLS version or supported cipher suites.<br />A set of common keys MAY be defined by the API in the future. To avoid<br />any ambiguity, implementation-specific definitions MUST use<br />domain-prefixed names, such as `example.com/my-custom-option`.<br />Un-prefixed names are reserved for key names defined by Gateway API.<br />Support: Implementation-specific |  | MaxProperties: 16 <br /> |


#### Group

_Underlying type:_ _string_

Group refers to a Kubernetes Group. It must either be an empty string or a
RFC 1123 subdomain.

This validation is based off of the corresponding Kubernetes validation:
https://github.com/kubernetes/apimachinery/blob/02cfb53916346d085a6c6c7c66f882e3c6b0eca6/pkg/util/validation/validation.go#L208

Valid values include:

* "" - empty string implies core Kubernetes API group
* "gateway.networking.k8s.io"
* "foo.example.com"

Invalid values include:

* "example.com/bar" - "/" is an invalid character

_Validation:_
- MaxLength: 253
- Pattern: `^$|^[a-z0-9]([-a-z0-9]*[a-z0-9])?(\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*$`

_Appears in:_
- [BackendObjectReference](#backendobjectreference)
- [BackendRef](#backendref)
- [GRPCBackendRef](#grpcbackendref)
- [HTTPBackendRef](#httpbackendref)
- [LocalObjectReference](#localobjectreference)
- [LocalParametersReference](#localparametersreference)
- [ObjectReference](#objectreference)
- [ParametersReference](#parametersreference)
- [ParentReference](#parentreference)
- [RouteGroupKind](#routegroupkind)
- [SecretObjectReference](#secretobjectreference)



#### HTTPBackendRef



HTTPBackendRef defines how a HTTPRoute forwards a HTTP request.

Note that when a namespace different than the local namespace is specified, a
ReferenceGrant object is required in the referent namespace to allow that
namespace's owner to accept the reference. See the ReferenceGrant
documentation for details.

<gateway:experimental:description>

When the BackendRef points to a Kubernetes Service, implementations SHOULD
honor the appProtocol field if it is set for the target Service Port.

Implementations supporting appProtocol SHOULD recognize the Kubernetes
Standard Application Protocols defined in KEP-3726.

If a Service appProtocol isn't specified, an implementation MAY infer the
backend protocol through its own means. Implementations MAY infer the
protocol from the Route type referring to the backend Service.

If a Route is not able to send traffic to the backend using the specified
protocol then the backend is considered invalid. Implementations MUST set the
"ResolvedRefs" condition to "False" with the "UnsupportedProtocol" reason.

</gateway:experimental:description>



_Appears in:_
- [HTTPRouteRule](#httprouterule)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `group` _[Group](#group)_ | Group is the group of the referent. For example, "gateway.networking.k8s.io".<br />When unspecified or empty string, core API group is inferred. |  | MaxLength: 253 <br />Pattern: `^$\|^[a-z0-9]([-a-z0-9]*[a-z0-9])?(\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*$` <br /> |
| `kind` _[Kind](#kind)_ | Kind is the Kubernetes resource kind of the referent. For example<br />"Service".<br />Defaults to "Service" when not specified.<br />ExternalName services can refer to CNAME DNS records that may live<br />outside of the cluster and as such are difficult to reason about in<br />terms of conformance. They also may not be safe to forward to (see<br />CVE-2021-25740 for more information). Implementations SHOULD NOT<br />support ExternalName Services.<br />Support: Core (Services with a type other than ExternalName)<br />Support: Implementation-specific (Services with type ExternalName) | Service | MaxLength: 63 <br />MinLength: 1 <br />Pattern: `^[a-zA-Z]([-a-zA-Z0-9]*[a-zA-Z0-9])?$` <br /> |
| `name` _[ObjectName](#objectname)_ | Name is the name of the referent. |  | MaxLength: 253 <br />MinLength: 1 <br /> |
| `namespace` _[Namespace](#namespace)_ | Namespace is the namespace of the backend. When unspecified, the local<br />namespace is inferred.<br />Note that when a namespace different than the local namespace is specified,<br />a ReferenceGrant object is required in the referent namespace to allow that<br />namespace's owner to accept the reference. See the ReferenceGrant<br />documentation for details.<br />Support: Core |  | MaxLength: 63 <br />MinLength: 1 <br />Pattern: `^[a-z0-9]([-a-z0-9]*[a-z0-9])?$` <br /> |
| `port` _[PortNumber](#portnumber)_ | Port specifies the destination port number to use for this resource.<br />Port is required when the referent is a Kubernetes Service. In this<br />case, the port number is the service port number, not the target port.<br />For other resources, destination port might be derived from the referent<br />resource or this field. |  | Maximum: 65535 <br />Minimum: 1 <br /> |
| `weight` _integer_ | Weight specifies the proportion of requests forwarded to the referenced<br />backend. This is computed as weight/(sum of all weights in this<br />BackendRefs list). For non-zero values, there may be some epsilon from<br />the exact proportion defined here depending on the precision an<br />implementation supports. Weight is not a percentage and the sum of<br />weights does not need to equal 100.<br />If only one backend is specified and it has a weight greater than 0, 100%<br />of the traffic is forwarded to that backend. If weight is set to 0, no<br />traffic should be forwarded for this entry. If unspecified, weight<br />defaults to 1.<br />Support for this field varies based on the context where used. | 1 | Maximum: 1e+06 <br />Minimum: 0 <br /> |
| `filters` _[HTTPRouteFilter](#httproutefilter) array_ | Filters defined at this level should be executed if and only if the<br />request is being forwarded to the backend defined here.<br />Support: Implementation-specific (For broader support of filters, use the<br />Filters field in HTTPRouteRule.) |  | MaxItems: 16 <br /> |


#### HTTPCORSFilter



HTTPCORSFilter defines a filter that that configures Cross-Origin Request
Sharing (CORS).



_Appears in:_
- [HTTPRouteFilter](#httproutefilter)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `allowOrigins` _[AbsoluteURI](#absoluteuri) array_ | AllowOrigins indicates whether the response can be shared with requested<br />resource from the given `Origin`.<br />The `Origin` consists of a scheme and a host, with an optional port, and<br />takes the form `<scheme>://<host>(:<port>)`.<br />Valid values for scheme are: `http` and `https`.<br />Valid values for port are any integer between 1 and 65535 (the list of<br />available TCP/UDP ports). Note that, if not included, port `80` is<br />assumed for `http` scheme origins, and port `443` is assumed for `https`<br />origins. This may affect origin matching.<br />The host part of the origin may contain the wildcard character `*`. These<br />wildcard characters behave as follows:<br />* `*` is a greedy match to the _left_, including any number of<br />  DNS labels to the left of its position. This also means that<br />  `*` will include any number of period `.` characters to the<br />  left of its position.<br />* A wildcard by itself matches all hosts.<br />An origin value that includes _only_ the `*` character indicates requests<br />from all `Origin`s are allowed.<br />When the `AllowOrigins` field is configured with multiple origins, it<br />means the server supports clients from multiple origins. If the request<br />`Origin` matches the configured allowed origins, the gateway must return<br />the given `Origin` and sets value of the header<br />`Access-Control-Allow-Origin` same as the `Origin` header provided by the<br />client.<br />The status code of a successful response to a "preflight" request is<br />always an OK status (i.e., 204 or 200).<br />If the request `Origin` does not match the configured allowed origins,<br />the gateway returns 204/200 response but doesn't set the relevant<br />cross-origin response headers. Alternatively, the gateway responds with<br />403 status to the "preflight" request is denied, coupled with omitting<br />the CORS headers. The cross-origin request fails on the client side.<br />Therefore, the client doesn't attempt the actual cross-origin request.<br />The `Access-Control-Allow-Origin` response header can only use `*`<br />wildcard as value when the `AllowCredentials` field is unspecified.<br />When the `AllowCredentials` field is specified and `AllowOrigins` field<br />specified with the `*` wildcard, the gateway must return a single origin<br />in the value of the `Access-Control-Allow-Origin` response header,<br />instead of specifying the `*` wildcard. The value of the header<br />`Access-Control-Allow-Origin` is same as the `Origin` header provided by<br />the client.<br />Support: Extended |  | MaxItems: 64 <br />MaxLength: 253 <br />MinLength: 1 <br />Pattern: `^(([^:/?#]+):)(//([^/?#]*))([^?#]*)(\?([^#]*))?(#(.*))?` <br /> |
| `allowCredentials` _[TrueField](#truefield)_ | AllowCredentials indicates whether the actual cross-origin request allows<br />to include credentials.<br />The only valid value for the `Access-Control-Allow-Credentials` response<br />header is true (case-sensitive).<br />If the credentials are not allowed in cross-origin requests, the gateway<br />will omit the header `Access-Control-Allow-Credentials` entirely rather<br />than setting its value to false.<br />Support: Extended |  | Enum: [true] <br /> |
| `allowMethods` _[HTTPMethodWithWildcard](#httpmethodwithwildcard) array_ | AllowMethods indicates which HTTP methods are supported for accessing the<br />requested resource.<br />Valid values are any method defined by RFC9110, along with the special<br />value `*`, which represents all HTTP methods are allowed.<br />Method names are case sensitive, so these values are also case-sensitive.<br />(See https://www.rfc-editor.org/rfc/rfc2616#section-5.1.1)<br />Multiple method names in the value of the `Access-Control-Allow-Methods`<br />response header are separated by a comma (",").<br />A CORS-safelisted method is a method that is `GET`, `HEAD`, or `POST`.<br />(See https://fetch.spec.whatwg.org/#cors-safelisted-method) The<br />CORS-safelisted methods are always allowed, regardless of whether they<br />are specified in the `AllowMethods` field.<br />When the `AllowMethods` field is configured with one or more methods, the<br />gateway must return the `Access-Control-Allow-Methods` response header<br />which value is present in the `AllowMethods` field.<br />If the HTTP method of the `Access-Control-Request-Method` request header<br />is not included in the list of methods specified by the response header<br />`Access-Control-Allow-Methods`, it will present an error on the client<br />side.<br />The `Access-Control-Allow-Methods` response header can only use `*`<br />wildcard as value when the `AllowCredentials` field is unspecified.<br />When the `AllowCredentials` field is specified and `AllowMethods` field<br />specified with the `*` wildcard, the gateway must specify one HTTP method<br />in the value of the Access-Control-Allow-Methods response header. The<br />value of the header `Access-Control-Allow-Methods` is same as the<br />`Access-Control-Request-Method` header provided by the client. If the<br />header `Access-Control-Request-Method` is not included in the request,<br />the gateway will omit the `Access-Control-Allow-Methods` response header,<br />instead of specifying the `*` wildcard. A Gateway implementation may<br />choose to add implementation-specific default methods.<br />Support: Extended |  | Enum: [GET HEAD POST PUT DELETE CONNECT OPTIONS TRACE PATCH *] <br />MaxItems: 9 <br /> |
| `allowHeaders` _[HTTPHeaderName](#httpheadername) array_ | AllowHeaders indicates which HTTP request headers are supported for<br />accessing the requested resource.<br />Header names are not case sensitive.<br />Multiple header names in the value of the `Access-Control-Allow-Headers`<br />response header are separated by a comma (",").<br />When the `AllowHeaders` field is configured with one or more headers, the<br />gateway must return the `Access-Control-Allow-Headers` response header<br />which value is present in the `AllowHeaders` field.<br />If any header name in the `Access-Control-Request-Headers` request header<br />is not included in the list of header names specified by the response<br />header `Access-Control-Allow-Headers`, it will present an error on the<br />client side.<br />If any header name in the `Access-Control-Allow-Headers` response header<br />does not recognize by the client, it will also occur an error on the<br />client side.<br />A wildcard indicates that the requests with all HTTP headers are allowed.<br />The `Access-Control-Allow-Headers` response header can only use `*`<br />wildcard as value when the `AllowCredentials` field is unspecified.<br />When the `AllowCredentials` field is specified and `AllowHeaders` field<br />specified with the `*` wildcard, the gateway must specify one or more<br />HTTP headers in the value of the `Access-Control-Allow-Headers` response<br />header. The value of the header `Access-Control-Allow-Headers` is same as<br />the `Access-Control-Request-Headers` header provided by the client. If<br />the header `Access-Control-Request-Headers` is not included in the<br />request, the gateway will omit the `Access-Control-Allow-Headers`<br />response header, instead of specifying the `*` wildcard. A Gateway<br />implementation may choose to add implementation-specific default headers.<br />Support: Extended |  | MaxItems: 64 <br />MaxLength: 256 <br />MinLength: 1 <br />Pattern: `^[A-Za-z0-9!#$%&'*+\-.^_\x60\|~]+$` <br /> |
| `exposeHeaders` _[HTTPHeaderName](#httpheadername) array_ | ExposeHeaders indicates which HTTP response headers can be exposed<br />to client-side scripts in response to a cross-origin request.<br />A CORS-safelisted response header is an HTTP header in a CORS response<br />that it is considered safe to expose to the client scripts.<br />The CORS-safelisted response headers include the following headers:<br />`Cache-Control`<br />`Content-Language`<br />`Content-Length`<br />`Content-Type`<br />`Expires`<br />`Last-Modified`<br />`Pragma`<br />(See https://fetch.spec.whatwg.org/#cors-safelisted-response-header-name)<br />The CORS-safelisted response headers are exposed to client by default.<br />When an HTTP header name is specified using the `ExposeHeaders` field,<br />this additional header will be exposed as part of the response to the<br />client.<br />Header names are not case sensitive.<br />Multiple header names in the value of the `Access-Control-Expose-Headers`<br />response header are separated by a comma (",").<br />A wildcard indicates that the responses with all HTTP headers are exposed<br />to clients. The `Access-Control-Expose-Headers` response header can only<br />use `*` wildcard as value when the `AllowCredentials` field is<br />unspecified.<br />Support: Extended |  | MaxItems: 64 <br />MaxLength: 256 <br />MinLength: 1 <br />Pattern: `^[A-Za-z0-9!#$%&'*+\-.^_\x60\|~]+$` <br /> |
| `maxAge` _integer_ | MaxAge indicates the duration (in seconds) for the client to cache the<br />results of a "preflight" request.<br />The information provided by the `Access-Control-Allow-Methods` and<br />`Access-Control-Allow-Headers` response headers can be cached by the<br />client until the time specified by `Access-Control-Max-Age` elapses.<br />The default value of `Access-Control-Max-Age` response header is 5<br />(seconds). | 5 | Minimum: 1 <br /> |


#### HTTPHeader



HTTPHeader represents an HTTP Header name and value as defined by RFC 7230.



_Appears in:_
- [HTTPHeaderFilter](#httpheaderfilter)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `name` _[HTTPHeaderName](#httpheadername)_ | Name is the name of the HTTP Header to be matched. Name matching MUST be<br />case-insensitive. (See https://tools.ietf.org/html/rfc7230#section-3.2).<br />If multiple entries specify equivalent header names, the first entry with<br />an equivalent name MUST be considered for a match. Subsequent entries<br />with an equivalent header name MUST be ignored. Due to the<br />case-insensitivity of header names, "foo" and "Foo" are considered<br />equivalent. |  | MaxLength: 256 <br />MinLength: 1 <br />Pattern: `^[A-Za-z0-9!#$%&'*+\-.^_\x60\|~]+$` <br /> |
| `value` _string_ | Value is the value of HTTP Header to be matched. |  | MaxLength: 4096 <br />MinLength: 1 <br /> |


#### HTTPHeaderFilter



HTTPHeaderFilter defines a filter that modifies the headers of an HTTP
request or response. Only one action for a given header name is permitted.
Filters specifying multiple actions of the same or different type for any one
header name are invalid and will be rejected by CRD validation.
Configuration to set or add multiple values for a header must use RFC 7230
header value formatting, separating each value with a comma.



_Appears in:_
- [GRPCRouteFilter](#grpcroutefilter)
- [HTTPRouteFilter](#httproutefilter)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `set` _[HTTPHeader](#httpheader) array_ | Set overwrites the request with the given header (name, value)<br />before the action.<br />Input:<br />  GET /foo HTTP/1.1<br />  my-header: foo<br />Config:<br />  set:<br />  - name: "my-header"<br />    value: "bar"<br />Output:<br />  GET /foo HTTP/1.1<br />  my-header: bar |  | MaxItems: 16 <br /> |
| `add` _[HTTPHeader](#httpheader) array_ | Add adds the given header(s) (name, value) to the request<br />before the action. It appends to any existing values associated<br />with the header name.<br />Input:<br />  GET /foo HTTP/1.1<br />  my-header: foo<br />Config:<br />  add:<br />  - name: "my-header"<br />    value: "bar,baz"<br />Output:<br />  GET /foo HTTP/1.1<br />  my-header: foo,bar,baz |  | MaxItems: 16 <br /> |
| `remove` _string array_ | Remove the given header(s) from the HTTP request before the action. The<br />value of Remove is a list of HTTP header names. Note that the header<br />names are case-insensitive (see<br />https://datatracker.ietf.org/doc/html/rfc2616#section-4.2).<br />Input:<br />  GET /foo HTTP/1.1<br />  my-header1: foo<br />  my-header2: bar<br />  my-header3: baz<br />Config:<br />  remove: ["my-header1", "my-header3"]<br />Output:<br />  GET /foo HTTP/1.1<br />  my-header2: bar |  | MaxItems: 16 <br /> |


#### HTTPHeaderMatch



HTTPHeaderMatch describes how to select a HTTP route by matching HTTP request
headers.



_Appears in:_
- [HTTPRouteMatch](#httproutematch)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `type` _[HeaderMatchType](#headermatchtype)_ | Type specifies how to match against the value of the header.<br />Support: Core (Exact)<br />Support: Implementation-specific (RegularExpression)<br />Since RegularExpression HeaderMatchType has implementation-specific<br />conformance, implementations can support POSIX, PCRE or any other dialects<br />of regular expressions. Please read the implementation's documentation to<br />determine the supported dialect. | Exact | Enum: [Exact RegularExpression] <br /> |
| `name` _[HTTPHeaderName](#httpheadername)_ | Name is the name of the HTTP Header to be matched. Name matching MUST be<br />case-insensitive. (See https://tools.ietf.org/html/rfc7230#section-3.2).<br />If multiple entries specify equivalent header names, only the first<br />entry with an equivalent name MUST be considered for a match. Subsequent<br />entries with an equivalent header name MUST be ignored. Due to the<br />case-insensitivity of header names, "foo" and "Foo" are considered<br />equivalent.<br />When a header is repeated in an HTTP request, it is<br />implementation-specific behavior as to how this is represented.<br />Generally, proxies should follow the guidance from the RFC:<br />https://www.rfc-editor.org/rfc/rfc7230.html#section-3.2.2 regarding<br />processing a repeated header, with special handling for "Set-Cookie". |  | MaxLength: 256 <br />MinLength: 1 <br />Pattern: `^[A-Za-z0-9!#$%&'*+\-.^_\x60\|~]+$` <br /> |
| `value` _string_ | Value is the value of HTTP Header to be matched. |  | MaxLength: 4096 <br />MinLength: 1 <br /> |


#### HTTPHeaderName

_Underlying type:_ _[HeaderName](#headername)_

HTTPHeaderName is the name of an HTTP header.

Valid values include:

* "Authorization"
* "Set-Cookie"

Invalid values include:

  - ":method" - ":" is an invalid character. This means that HTTP/2 pseudo
    headers are not currently supported by this type.
  - "/invalid" - "/ " is an invalid character

_Validation:_
- MaxLength: 256
- MinLength: 1
- Pattern: `^[A-Za-z0-9!#$%&'*+\-.^_\x60|~]+$`

_Appears in:_
- [HTTPCORSFilter](#httpcorsfilter)
- [HTTPHeader](#httpheader)
- [HTTPHeaderMatch](#httpheadermatch)
- [HTTPQueryParamMatch](#httpqueryparammatch)



#### HTTPMethod

_Underlying type:_ _string_

HTTPMethod describes how to select a HTTP route by matching the HTTP
method as defined by
[RFC 7231](https://datatracker.ietf.org/doc/html/rfc7231#section-4) and
[RFC 5789](https://datatracker.ietf.org/doc/html/rfc5789#section-2).
The value is expected in upper case.

Note that values may be added to this enum, implementations
must ensure that unknown values will not cause a crash.

Unknown values here must result in the implementation setting the
Accepted Condition for the Route to `status: False`, with a
Reason of `UnsupportedValue`.

_Validation:_
- Enum: [GET HEAD POST PUT DELETE CONNECT OPTIONS TRACE PATCH]

_Appears in:_
- [HTTPRouteMatch](#httproutematch)

| Field | Description |
| --- | --- |
| `GET` |  |
| `HEAD` |  |
| `POST` |  |
| `PUT` |  |
| `DELETE` |  |
| `CONNECT` |  |
| `OPTIONS` |  |
| `TRACE` |  |
| `PATCH` |  |


#### HTTPMethodWithWildcard

_Underlying type:_ _string_



_Validation:_
- Enum: [GET HEAD POST PUT DELETE CONNECT OPTIONS TRACE PATCH *]

_Appears in:_
- [HTTPCORSFilter](#httpcorsfilter)



#### HTTPPathMatch



HTTPPathMatch describes how to select a HTTP route by matching the HTTP request path.



_Appears in:_
- [HTTPRouteMatch](#httproutematch)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `type` _[PathMatchType](#pathmatchtype)_ | Type specifies how to match against the path Value.<br />Support: Core (Exact, PathPrefix)<br />Support: Implementation-specific (RegularExpression) | PathPrefix | Enum: [Exact PathPrefix RegularExpression] <br /> |
| `value` _string_ | Value of the HTTP path to match against. | / | MaxLength: 1024 <br /> |


#### HTTPPathModifier



HTTPPathModifier defines configuration for path modifiers.



_Appears in:_
- [HTTPRequestRedirectFilter](#httprequestredirectfilter)
- [HTTPURLRewriteFilter](#httpurlrewritefilter)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `type` _[HTTPPathModifierType](#httppathmodifiertype)_ | Type defines the type of path modifier. Additional types may be<br />added in a future release of the API.<br />Note that values may be added to this enum, implementations<br />must ensure that unknown values will not cause a crash.<br />Unknown values here must result in the implementation setting the<br />Accepted Condition for the Route to `status: False`, with a<br />Reason of `UnsupportedValue`. |  | Enum: [ReplaceFullPath ReplacePrefixMatch] <br /> |
| `replaceFullPath` _string_ | ReplaceFullPath specifies the value with which to replace the full path<br />of a request during a rewrite or redirect. |  | MaxLength: 1024 <br /> |
| `replacePrefixMatch` _string_ | ReplacePrefixMatch specifies the value with which to replace the prefix<br />match of a request during a rewrite or redirect. For example, a request<br />to "/foo/bar" with a prefix match of "/foo" and a ReplacePrefixMatch<br />of "/xyz" would be modified to "/xyz/bar".<br />Note that this matches the behavior of the PathPrefix match type. This<br />matches full path elements. A path element refers to the list of labels<br />in the path split by the `/` separator. When specified, a trailing `/` is<br />ignored. For example, the paths `/abc`, `/abc/`, and `/abc/def` would all<br />match the prefix `/abc`, but the path `/abcd` would not.<br />ReplacePrefixMatch is only compatible with a `PathPrefix` HTTPRouteMatch.<br />Using any other HTTPRouteMatch type on the same HTTPRouteRule will result in<br />the implementation setting the Accepted Condition for the Route to `status: False`.<br />Request Path \| Prefix Match \| Replace Prefix \| Modified Path |  | MaxLength: 1024 <br /> |


#### HTTPPathModifierType

_Underlying type:_ _string_

HTTPPathModifierType defines the type of path redirect or rewrite.



_Appears in:_
- [HTTPPathModifier](#httppathmodifier)

| Field | Description |
| --- | --- |
| `ReplaceFullPath` | This type of modifier indicates that the full path will be replaced<br />by the specified value.<br /> |
| `ReplacePrefixMatch` | This type of modifier indicates that any prefix path matches will be<br />replaced by the substitution value. For example, a path with a prefix<br />match of "/foo" and a ReplacePrefixMatch substitution of "/bar" will have<br />the "/foo" prefix replaced with "/bar" in matching requests.<br />Note that this matches the behavior of the PathPrefix match type. This<br />matches full path elements. A path element refers to the list of labels<br />in the path split by the `/` separator. When specified, a trailing `/` is<br />ignored. For example, the paths `/abc`, `/abc/`, and `/abc/def` would all<br />match the prefix `/abc`, but the path `/abcd` would not.<br /> |


#### HTTPQueryParamMatch



HTTPQueryParamMatch describes how to select a HTTP route by matching HTTP
query parameters.



_Appears in:_
- [HTTPRouteMatch](#httproutematch)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `type` _[QueryParamMatchType](#queryparammatchtype)_ | Type specifies how to match against the value of the query parameter.<br />Support: Extended (Exact)<br />Support: Implementation-specific (RegularExpression)<br />Since RegularExpression QueryParamMatchType has Implementation-specific<br />conformance, implementations can support POSIX, PCRE or any other<br />dialects of regular expressions. Please read the implementation's<br />documentation to determine the supported dialect. | Exact | Enum: [Exact RegularExpression] <br /> |
| `name` _[HTTPHeaderName](#httpheadername)_ | Name is the name of the HTTP query param to be matched. This must be an<br />exact string match. (See<br />https://tools.ietf.org/html/rfc7230#section-2.7.3).<br />If multiple entries specify equivalent query param names, only the first<br />entry with an equivalent name MUST be considered for a match. Subsequent<br />entries with an equivalent query param name MUST be ignored.<br />If a query param is repeated in an HTTP request, the behavior is<br />purposely left undefined, since different data planes have different<br />capabilities. However, it is *recommended* that implementations should<br />match against the first value of the param if the data plane supports it,<br />as this behavior is expected in other load balancing contexts outside of<br />the Gateway API.<br />Users SHOULD NOT route traffic based on repeated query params to guard<br />themselves against potential differences in the implementations. |  | MaxLength: 256 <br />MinLength: 1 <br />Pattern: `^[A-Za-z0-9!#$%&'*+\-.^_\x60\|~]+$` <br /> |
| `value` _string_ | Value is the value of HTTP query param to be matched. |  | MaxLength: 1024 <br />MinLength: 1 <br /> |


#### HTTPRequestMirrorFilter



HTTPRequestMirrorFilter defines configuration for the RequestMirror filter.



_Appears in:_
- [GRPCRouteFilter](#grpcroutefilter)
- [HTTPRouteFilter](#httproutefilter)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `backendRef` _[BackendObjectReference](#backendobjectreference)_ | BackendRef references a resource where mirrored requests are sent.<br />Mirrored requests must be sent only to a single destination endpoint<br />within this BackendRef, irrespective of how many endpoints are present<br />within this BackendRef.<br />If the referent cannot be found, this BackendRef is invalid and must be<br />dropped from the Gateway. The controller must ensure the "ResolvedRefs"<br />condition on the Route status is set to `status: False` and not configure<br />this backend in the underlying implementation.<br />If there is a cross-namespace reference to an *existing* object<br />that is not allowed by a ReferenceGrant, the controller must ensure the<br />"ResolvedRefs"  condition on the Route is set to `status: False`,<br />with the "RefNotPermitted" reason and not configure this backend in the<br />underlying implementation.<br />In either error case, the Message of the `ResolvedRefs` Condition<br />should be used to provide more detail about the problem.<br />Support: Extended for Kubernetes Service<br />Support: Implementation-specific for any other resource |  |  |
| `percent` _integer_ | Percent represents the percentage of requests that should be<br />mirrored to BackendRef. Its minimum value is 0 (indicating 0% of<br />requests) and its maximum value is 100 (indicating 100% of requests).<br />Only one of Fraction or Percent may be specified. If neither field<br />is specified, 100% of requests will be mirrored. |  | Maximum: 100 <br />Minimum: 0 <br /> |
| `fraction` _[Fraction](#fraction)_ | Fraction represents the fraction of requests that should be<br />mirrored to BackendRef.<br />Only one of Fraction or Percent may be specified. If neither field<br />is specified, 100% of requests will be mirrored. |  |  |


#### HTTPRequestRedirectFilter



HTTPRequestRedirect defines a filter that redirects a request. This filter
MUST NOT be used on the same Route rule as a HTTPURLRewrite filter.



_Appears in:_
- [HTTPRouteFilter](#httproutefilter)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `scheme` _string_ | Scheme is the scheme to be used in the value of the `Location` header in<br />the response. When empty, the scheme of the request is used.<br />Scheme redirects can affect the port of the redirect, for more information,<br />refer to the documentation for the port field of this filter.<br />Note that values may be added to this enum, implementations<br />must ensure that unknown values will not cause a crash.<br />Unknown values here must result in the implementation setting the<br />Accepted Condition for the Route to `status: False`, with a<br />Reason of `UnsupportedValue`.<br />Support: Extended |  | Enum: [http https] <br /> |
| `hostname` _[PreciseHostname](#precisehostname)_ | Hostname is the hostname to be used in the value of the `Location`<br />header in the response.<br />When empty, the hostname in the `Host` header of the request is used.<br />Support: Core |  | MaxLength: 253 <br />MinLength: 1 <br />Pattern: `^[a-z0-9]([-a-z0-9]*[a-z0-9])?(\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*$` <br /> |
| `path` _[HTTPPathModifier](#httppathmodifier)_ | Path defines parameters used to modify the path of the incoming request.<br />The modified path is then used to construct the `Location` header. When<br />empty, the request path is used as-is.<br />Support: Extended |  |  |
| `port` _[PortNumber](#portnumber)_ | Port is the port to be used in the value of the `Location`<br />header in the response.<br />If no port is specified, the redirect port MUST be derived using the<br />following rules:<br />* If redirect scheme is not-empty, the redirect port MUST be the well-known<br />  port associated with the redirect scheme. Specifically "http" to port 80<br />  and "https" to port 443. If the redirect scheme does not have a<br />  well-known port, the listener port of the Gateway SHOULD be used.<br />* If redirect scheme is empty, the redirect port MUST be the Gateway<br />  Listener port.<br />Implementations SHOULD NOT add the port number in the 'Location'<br />header in the following cases:<br />* A Location header that will use HTTP (whether that is determined via<br />  the Listener protocol or the Scheme field) _and_ use port 80.<br />* A Location header that will use HTTPS (whether that is determined via<br />  the Listener protocol or the Scheme field) _and_ use port 443.<br />Support: Extended |  | Maximum: 65535 <br />Minimum: 1 <br /> |
| `statusCode` _integer_ | StatusCode is the HTTP status code to be used in response.<br />Note that values may be added to this enum, implementations<br />must ensure that unknown values will not cause a crash.<br />Unknown values here must result in the implementation setting the<br />Accepted Condition for the Route to `status: False`, with a<br />Reason of `UnsupportedValue`.<br />Support: Core | 302 | Enum: [301 302] <br /> |


#### HTTPRoute



HTTPRoute provides a way to route HTTP requests. This includes the capability
to match requests by hostname, path, header, or query param. Filters can be
used to specify additional processing steps. Backends specify where matching
requests should be routed.



_Appears in:_
- [HTTPRoute](#httproute)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `apiVersion` _string_ | `gateway.networking.k8s.io/v1` | | |
| `kind` _string_ | `HTTPRoute` | | |
| `metadata` _[ObjectMeta](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.32/#objectmeta-v1-meta)_ | Refer to Kubernetes API documentation for fields of `metadata`. |  |  |
| `spec` _[HTTPRouteSpec](#httproutespec)_ | Spec defines the desired state of HTTPRoute. |  |  |
| `status` _[HTTPRouteStatus](#httproutestatus)_ | Status defines the current state of HTTPRoute. |  |  |


#### HTTPRouteFilter



HTTPRouteFilter defines processing steps that must be completed during the
request or response lifecycle. HTTPRouteFilters are meant as an extension
point to express processing that may be done in Gateway implementations. Some
examples include request or response modification, implementing
authentication strategies, rate-limiting, and traffic shaping. API
guarantee/conformance is defined based on the type of the filter.

<gateway:experimental:validation:XValidation:message="filter.cors must be nil if the filter.type is not CORS",rule="!(has(self.cors) && self.type != 'CORS')">
<gateway:experimental:validation:XValidation:message="filter.cors must be specified for CORS filter.type",rule="!(!has(self.cors) && self.type == 'CORS')">



_Appears in:_
- [HTTPBackendRef](#httpbackendref)
- [HTTPRouteRule](#httprouterule)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `type` _[HTTPRouteFilterType](#httproutefiltertype)_ | Type identifies the type of filter to apply. As with other API fields,<br />types are classified into three conformance levels:<br />- Core: Filter types and their corresponding configuration defined by<br />  "Support: Core" in this package, e.g. "RequestHeaderModifier". All<br />  implementations must support core filters.<br />- Extended: Filter types and their corresponding configuration defined by<br />  "Support: Extended" in this package, e.g. "RequestMirror". Implementers<br />  are encouraged to support extended filters.<br />- Implementation-specific: Filters that are defined and supported by<br />  specific vendors.<br />  In the future, filters showing convergence in behavior across multiple<br />  implementations will be considered for inclusion in extended or core<br />  conformance levels. Filter-specific configuration for such filters<br />  is specified using the ExtensionRef field. `Type` should be set to<br />  "ExtensionRef" for custom filters.<br />Implementers are encouraged to define custom implementation types to<br />extend the core API with implementation-specific behavior.<br />If a reference to a custom filter type cannot be resolved, the filter<br />MUST NOT be skipped. Instead, requests that would have been processed by<br />that filter MUST receive a HTTP error response.<br />Note that values may be added to this enum, implementations<br />must ensure that unknown values will not cause a crash.<br />Unknown values here must result in the implementation setting the<br />Accepted Condition for the Route to `status: False`, with a<br />Reason of `UnsupportedValue`.<br /><gateway:experimental:validation:Enum=RequestHeaderModifier;ResponseHeaderModifier;RequestMirror;RequestRedirect;URLRewrite;ExtensionRef;CORS> |  | Enum: [RequestHeaderModifier ResponseHeaderModifier RequestMirror RequestRedirect URLRewrite ExtensionRef] <br /> |
| `requestHeaderModifier` _[HTTPHeaderFilter](#httpheaderfilter)_ | RequestHeaderModifier defines a schema for a filter that modifies request<br />headers.<br />Support: Core |  |  |
| `responseHeaderModifier` _[HTTPHeaderFilter](#httpheaderfilter)_ | ResponseHeaderModifier defines a schema for a filter that modifies response<br />headers.<br />Support: Extended |  |  |
| `requestMirror` _[HTTPRequestMirrorFilter](#httprequestmirrorfilter)_ | RequestMirror defines a schema for a filter that mirrors requests.<br />Requests are sent to the specified destination, but responses from<br />that destination are ignored.<br />This filter can be used multiple times within the same rule. Note that<br />not all implementations will be able to support mirroring to multiple<br />backends.<br />Support: Extended |  |  |
| `requestRedirect` _[HTTPRequestRedirectFilter](#httprequestredirectfilter)_ | RequestRedirect defines a schema for a filter that responds to the<br />request with an HTTP redirection.<br />Support: Core |  |  |
| `urlRewrite` _[HTTPURLRewriteFilter](#httpurlrewritefilter)_ | URLRewrite defines a schema for a filter that modifies a request during forwarding.<br />Support: Extended |  |  |
| `cors` _[HTTPCORSFilter](#httpcorsfilter)_ | CORS defines a schema for a filter that responds to the<br />cross-origin request based on HTTP response header.<br />Support: Extended<br /><gateway:experimental> |  |  |
| `extensionRef` _[LocalObjectReference](#localobjectreference)_ | ExtensionRef is an optional, implementation-specific extension to the<br />"filter" behavior.  For example, resource "myroutefilter" in group<br />"networking.example.net"). ExtensionRef MUST NOT be used for core and<br />extended filters.<br />This filter can be used multiple times within the same rule.<br />Support: Implementation-specific |  |  |


#### HTTPRouteFilterType

_Underlying type:_ _string_

HTTPRouteFilterType identifies a type of HTTPRoute filter.



_Appears in:_
- [HTTPRouteFilter](#httproutefilter)

| Field | Description |
| --- | --- |
| `RequestHeaderModifier` | HTTPRouteFilterRequestHeaderModifier can be used to add or remove an HTTP<br />header from an HTTP request before it is sent to the upstream target.<br />Support in HTTPRouteRule: Core<br />Support in HTTPBackendRef: Extended<br /> |
| `ResponseHeaderModifier` | HTTPRouteFilterResponseHeaderModifier can be used to add or remove an HTTP<br />header from an HTTP response before it is sent to the client.<br />Support in HTTPRouteRule: Extended<br />Support in HTTPBackendRef: Extended<br /> |
| `RequestRedirect` | HTTPRouteFilterRequestRedirect can be used to redirect a request to<br />another location. This filter can also be used for HTTP to HTTPS<br />redirects. This may not be used on the same Route rule or BackendRef as a<br />URLRewrite filter.<br />Support in HTTPRouteRule: Core<br />Support in HTTPBackendRef: Extended<br /> |
| `URLRewrite` | HTTPRouteFilterURLRewrite can be used to modify a request during<br />forwarding. At most one of these filters may be used on a Route rule.<br />This may not be used on the same Route rule or BackendRef as a<br />RequestRedirect filter.<br />Support in HTTPRouteRule: Extended<br />Support in HTTPBackendRef: Extended<br /> |
| `RequestMirror` | HTTPRouteFilterRequestMirror can be used to mirror HTTP requests to a<br />different backend. The responses from this backend MUST be ignored by<br />the Gateway.<br />Support in HTTPRouteRule: Extended<br />Support in HTTPBackendRef: Extended<br /> |
| `CORS` | HTTPRouteFilterCORS can be used to add CORS headers to an<br />HTTP response before it is sent to the client.<br />Support in HTTPRouteRule: Extended<br />Support in HTTPBackendRef: Extended<br /><gateway:experimental><br /> |
| `ExtensionRef` | HTTPRouteFilterExtensionRef should be used for configuring custom<br />HTTP filters.<br />Support in HTTPRouteRule: Implementation-specific<br />Support in HTTPBackendRef: Implementation-specific<br /> |


#### HTTPRouteMatch



HTTPRouteMatch defines the predicate used to match requests to a given
action. Multiple match types are ANDed together, i.e. the match will
evaluate to true only if all conditions are satisfied.

For example, the match below will match a HTTP request only if its path
starts with `/foo` AND it contains the `version: v1` header:

```
match:

	path:
	  value: "/foo"
	headers:
	- name: "version"
	  value "v1"

```



_Appears in:_
- [HTTPRouteRule](#httprouterule)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `path` _[HTTPPathMatch](#httppathmatch)_ | Path specifies a HTTP request path matcher. If this field is not<br />specified, a default prefix match on the "/" path is provided. | \{ type:PathPrefix value:/ \} |  |
| `headers` _[HTTPHeaderMatch](#httpheadermatch) array_ | Headers specifies HTTP request header matchers. Multiple match values are<br />ANDed together, meaning, a request must match all the specified headers<br />to select the route. |  | MaxItems: 16 <br /> |
| `queryParams` _[HTTPQueryParamMatch](#httpqueryparammatch) array_ | QueryParams specifies HTTP query parameter matchers. Multiple match<br />values are ANDed together, meaning, a request must match all the<br />specified query parameters to select the route.<br />Support: Extended |  | MaxItems: 16 <br /> |
| `method` _[HTTPMethod](#httpmethod)_ | Method specifies HTTP method matcher.<br />When specified, this route will be matched only if the request has the<br />specified method.<br />Support: Extended |  | Enum: [GET HEAD POST PUT DELETE CONNECT OPTIONS TRACE PATCH] <br /> |


#### HTTPRouteRetry



HTTPRouteRetry defines retry configuration for an HTTPRoute.

Implementations SHOULD retry on connection errors (disconnect, reset, timeout,
TCP failure) if a retry stanza is configured.



_Appears in:_
- [HTTPRouteRule](#httprouterule)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `codes` _[HTTPRouteRetryStatusCode](#httprouteretrystatuscode) array_ | Codes defines the HTTP response status codes for which a backend request<br />should be retried.<br />Support: Extended |  | Maximum: 599 <br />Minimum: 400 <br /> |
| `attempts` _integer_ | Attempts specifies the maximum number of times an individual request<br />from the gateway to a backend should be retried.<br />If the maximum number of retries has been attempted without a successful<br />response from the backend, the Gateway MUST return an error.<br />When this field is unspecified, the number of times to attempt to retry<br />a backend request is implementation-specific.<br />Support: Extended |  |  |
| `backoff` _[Duration](#duration)_ | Backoff specifies the minimum duration a Gateway should wait between<br />retry attempts and is represented in Gateway API Duration formatting.<br />For example, setting the `rules[].retry.backoff` field to the value<br />`100ms` will cause a backend request to first be retried approximately<br />100 milliseconds after timing out or receiving a response code configured<br />to be retryable.<br />An implementation MAY use an exponential or alternative backoff strategy<br />for subsequent retry attempts, MAY cap the maximum backoff duration to<br />some amount greater than the specified minimum, and MAY add arbitrary<br />jitter to stagger requests, as long as unsuccessful backend requests are<br />not retried before the configured minimum duration.<br />If a Request timeout (`rules[].timeouts.request`) is configured on the<br />route, the entire duration of the initial request and any retry attempts<br />MUST not exceed the Request timeout duration. If any retry attempts are<br />still in progress when the Request timeout duration has been reached,<br />these SHOULD be canceled if possible and the Gateway MUST immediately<br />return a timeout error.<br />If a BackendRequest timeout (`rules[].timeouts.backendRequest`) is<br />configured on the route, any retry attempts which reach the configured<br />BackendRequest timeout duration without a response SHOULD be canceled if<br />possible and the Gateway should wait for at least the specified backoff<br />duration before attempting to retry the backend request again.<br />If a BackendRequest timeout is _not_ configured on the route, retry<br />attempts MAY time out after an implementation default duration, or MAY<br />remain pending until a configured Request timeout or implementation<br />default duration for total request time is reached.<br />When this field is unspecified, the time to wait between retry attempts<br />is implementation-specific.<br />Support: Extended |  | Pattern: `^([0-9]\{1,5\}(h\|m\|s\|ms))\{1,4\}$` <br /> |


#### HTTPRouteRetryStatusCode

_Underlying type:_ _integer_

HTTPRouteRetryStatusCode defines an HTTP response status code for
which a backend request should be retried.

Implementations MUST support the following status codes as retryable:

* 500
* 502
* 503
* 504

Implementations MAY support specifying additional discrete values in the
500-599 range.

Implementations MAY support specifying discrete values in the 400-499 range,
which are often inadvisable to retry.

<gateway:experimental>

_Validation:_
- Maximum: 599
- Minimum: 400

_Appears in:_
- [HTTPRouteRetry](#httprouteretry)



#### HTTPRouteRule



HTTPRouteRule defines semantics for matching an HTTP request based on
conditions (matches), processing it (filters), and forwarding the request to
an API object (backendRefs).



_Appears in:_
- [HTTPRouteSpec](#httproutespec)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `name` _[SectionName](#sectionname)_ | Name is the name of the route rule. This name MUST be unique within a Route if it is set.<br />Support: Extended<br /><gateway:experimental> |  | MaxLength: 253 <br />MinLength: 1 <br />Pattern: `^[a-z0-9]([-a-z0-9]*[a-z0-9])?(\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*$` <br /> |
| `matches` _[HTTPRouteMatch](#httproutematch) array_ | Matches define conditions used for matching the rule against incoming<br />HTTP requests. Each match is independent, i.e. this rule will be matched<br />if **any** one of the matches is satisfied.<br />For example, take the following matches configuration:<br />```<br />matches:<br />- path:<br />    value: "/foo"<br />  headers:<br />  - name: "version"<br />    value: "v2"<br />- path:<br />    value: "/v2/foo"<br />```<br />For a request to match against this rule, a request must satisfy<br />EITHER of the two conditions:<br />- path prefixed with `/foo` AND contains the header `version: v2`<br />- path prefix of `/v2/foo`<br />See the documentation for HTTPRouteMatch on how to specify multiple<br />match conditions that should be ANDed together.<br />If no matches are specified, the default is a prefix<br />path match on "/", which has the effect of matching every<br />HTTP request.<br />Proxy or Load Balancer routing configuration generated from HTTPRoutes<br />MUST prioritize matches based on the following criteria, continuing on<br />ties. Across all rules specified on applicable Routes, precedence must be<br />given to the match having:<br />* "Exact" path match.<br />* "Prefix" path match with largest number of characters.<br />* Method match.<br />* Largest number of header matches.<br />* Largest number of query param matches.<br />Note: The precedence of RegularExpression path matches are implementation-specific.<br />If ties still exist across multiple Routes, matching precedence MUST be<br />determined in order of the following criteria, continuing on ties:<br />* The oldest Route based on creation timestamp.<br />* The Route appearing first in alphabetical order by<br />  "\{namespace\}/\{name\}".<br />If ties still exist within an HTTPRoute, matching precedence MUST be granted<br />to the FIRST matching rule (in list order) with a match meeting the above<br />criteria.<br />When no rules matching a request have been successfully attached to the<br />parent a request is coming from, a HTTP 404 status code MUST be returned. | [map[path:map[type:PathPrefix value:/]]] | MaxItems: 64 <br /> |
| `filters` _[HTTPRouteFilter](#httproutefilter) array_ | Filters define the filters that are applied to requests that match<br />this rule.<br />Wherever possible, implementations SHOULD implement filters in the order<br />they are specified.<br />Implementations MAY choose to implement this ordering strictly, rejecting<br />any combination or order of filters that cannot be supported. If implementations<br />choose a strict interpretation of filter ordering, they MUST clearly document<br />that behavior.<br />To reject an invalid combination or order of filters, implementations SHOULD<br />consider the Route Rules with this configuration invalid. If all Route Rules<br />in a Route are invalid, the entire Route would be considered invalid. If only<br />a portion of Route Rules are invalid, implementations MUST set the<br />"PartiallyInvalid" condition for the Route.<br />Conformance-levels at this level are defined based on the type of filter:<br />- ALL core filters MUST be supported by all implementations.<br />- Implementers are encouraged to support extended filters.<br />- Implementation-specific custom filters have no API guarantees across<br />  implementations.<br />Specifying the same filter multiple times is not supported unless explicitly<br />indicated in the filter.<br />All filters are expected to be compatible with each other except for the<br />URLRewrite and RequestRedirect filters, which may not be combined. If an<br />implementation cannot support other combinations of filters, they must clearly<br />document that limitation. In cases where incompatible or unsupported<br />filters are specified and cause the `Accepted` condition to be set to status<br />`False`, implementations may use the `IncompatibleFilters` reason to specify<br />this configuration error.<br />Support: Core |  | MaxItems: 16 <br /> |
| `backendRefs` _[HTTPBackendRef](#httpbackendref) array_ | BackendRefs defines the backend(s) where matching requests should be<br />sent.<br />Failure behavior here depends on how many BackendRefs are specified and<br />how many are invalid.<br />If *all* entries in BackendRefs are invalid, and there are also no filters<br />specified in this route rule, *all* traffic which matches this rule MUST<br />receive a 500 status code.<br />See the HTTPBackendRef definition for the rules about what makes a single<br />HTTPBackendRef invalid.<br />When a HTTPBackendRef is invalid, 500 status codes MUST be returned for<br />requests that would have otherwise been routed to an invalid backend. If<br />multiple backends are specified, and some are invalid, the proportion of<br />requests that would otherwise have been routed to an invalid backend<br />MUST receive a 500 status code.<br />For example, if two backends are specified with equal weights, and one is<br />invalid, 50 percent of traffic must receive a 500. Implementations may<br />choose how that 50 percent is determined.<br />When a HTTPBackendRef refers to a Service that has no ready endpoints,<br />implementations SHOULD return a 503 for requests to that backend instead.<br />If an implementation chooses to do this, all of the above rules for 500 responses<br />MUST also apply for responses that return a 503.<br />Support: Core for Kubernetes Service<br />Support: Extended for Kubernetes ServiceImport<br />Support: Implementation-specific for any other resource<br />Support for weight: Core |  | MaxItems: 16 <br /> |
| `timeouts` _[HTTPRouteTimeouts](#httproutetimeouts)_ | Timeouts defines the timeouts that can be configured for an HTTP request.<br />Support: Extended |  |  |
| `retry` _[HTTPRouteRetry](#httprouteretry)_ | Retry defines the configuration for when to retry an HTTP request.<br />Support: Extended<br /><gateway:experimental> |  |  |
| `sessionPersistence` _[SessionPersistence](#sessionpersistence)_ | SessionPersistence defines and configures session persistence<br />for the route rule.<br />Support: Extended<br /><gateway:experimental> |  |  |


#### HTTPRouteSpec



HTTPRouteSpec defines the desired state of HTTPRoute



_Appears in:_
- [HTTPRoute](#httproute)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `parentRefs` _[ParentReference](#parentreference) array_ | ParentRefs references the resources (usually Gateways) that a Route wants<br />to be attached to. Note that the referenced parent resource needs to<br />allow this for the attachment to be complete. For Gateways, that means<br />the Gateway needs to allow attachment from Routes of this kind and<br />namespace. For Services, that means the Service must either be in the same<br />namespace for a "producer" route, or the mesh implementation must support<br />and allow "consumer" routes for the referenced Service. ReferenceGrant is<br />not applicable for governing ParentRefs to Services - it is not possible to<br />create a "producer" route for a Service in a different namespace from the<br />Route.<br />There are two kinds of parent resources with "Core" support:<br />* Gateway (Gateway conformance profile)<br />* Service (Mesh conformance profile, ClusterIP Services only)<br />This API may be extended in the future to support additional kinds of parent<br />resources.<br />ParentRefs must be _distinct_. This means either that:<br />* They select different objects.  If this is the case, then parentRef<br />  entries are distinct. In terms of fields, this means that the<br />  multi-part key defined by `group`, `kind`, `namespace`, and `name` must<br />  be unique across all parentRef entries in the Route.<br />* They do not select different objects, but for each optional field used,<br />  each ParentRef that selects the same object must set the same set of<br />  optional fields to different values. If one ParentRef sets a<br />  combination of optional fields, all must set the same combination.<br />Some examples:<br />* If one ParentRef sets `sectionName`, all ParentRefs referencing the<br />  same object must also set `sectionName`.<br />* If one ParentRef sets `port`, all ParentRefs referencing the same<br />  object must also set `port`.<br />* If one ParentRef sets `sectionName` and `port`, all ParentRefs<br />  referencing the same object must also set `sectionName` and `port`.<br />It is possible to separately reference multiple distinct objects that may<br />be collapsed by an implementation. For example, some implementations may<br />choose to merge compatible Gateway Listeners together. If that is the<br />case, the list of routes attached to those resources should also be<br />merged.<br />Note that for ParentRefs that cross namespace boundaries, there are specific<br />rules. Cross-namespace references are only valid if they are explicitly<br />allowed by something in the namespace they are referring to. For example,<br />Gateway has the AllowedRoutes field, and ReferenceGrant provides a<br />generic way to enable other kinds of cross-namespace reference.<br /><gateway:experimental:description><br />ParentRefs from a Route to a Service in the same namespace are "producer"<br />routes, which apply default routing rules to inbound connections from<br />any namespace to the Service.<br />ParentRefs from a Route to a Service in a different namespace are<br />"consumer" routes, and these routing rules are only applied to outbound<br />connections originating from the same namespace as the Route, for which<br />the intended destination of the connections are a Service targeted as a<br />ParentRef of the Route.<br /></gateway:experimental:description><br /><gateway:standard:validation:XValidation:message="sectionName must be specified when parentRefs includes 2 or more references to the same parent",rule="self.all(p1, self.all(p2, p1.group == p2.group && p1.kind == p2.kind && p1.name == p2.name && (((!has(p1.__namespace__) \|\| p1.__namespace__ == '') && (!has(p2.__namespace__) \|\| p2.__namespace__ == '')) \|\| (has(p1.__namespace__) && has(p2.__namespace__) && p1.__namespace__ == p2.__namespace__ )) ? ((!has(p1.sectionName) \|\| p1.sectionName == '') == (!has(p2.sectionName) \|\| p2.sectionName == '')) : true))"><br /><gateway:standard:validation:XValidation:message="sectionName must be unique when parentRefs includes 2 or more references to the same parent",rule="self.all(p1, self.exists_one(p2, p1.group == p2.group && p1.kind == p2.kind && p1.name == p2.name && (((!has(p1.__namespace__) \|\| p1.__namespace__ == '') && (!has(p2.__namespace__) \|\| p2.__namespace__ == '')) \|\| (has(p1.__namespace__) && has(p2.__namespace__) && p1.__namespace__ == p2.__namespace__ )) && (((!has(p1.sectionName) \|\| p1.sectionName == '') && (!has(p2.sectionName) \|\| p2.sectionName == '')) \|\| (has(p1.sectionName) && has(p2.sectionName) && p1.sectionName == p2.sectionName))))"><br /><gateway:experimental:validation:XValidation:message="sectionName or port must be specified when parentRefs includes 2 or more references to the same parent",rule="self.all(p1, self.all(p2, p1.group == p2.group && p1.kind == p2.kind && p1.name == p2.name && (((!has(p1.__namespace__) \|\| p1.__namespace__ == '') && (!has(p2.__namespace__) \|\| p2.__namespace__ == '')) \|\| (has(p1.__namespace__) && has(p2.__namespace__) && p1.__namespace__ == p2.__namespace__)) ? ((!has(p1.sectionName) \|\| p1.sectionName == '') == (!has(p2.sectionName) \|\| p2.sectionName == '') && (!has(p1.port) \|\| p1.port == 0) == (!has(p2.port) \|\| p2.port == 0)): true))"><br /><gateway:experimental:validation:XValidation:message="sectionName or port must be unique when parentRefs includes 2 or more references to the same parent",rule="self.all(p1, self.exists_one(p2, p1.group == p2.group && p1.kind == p2.kind && p1.name == p2.name && (((!has(p1.__namespace__) \|\| p1.__namespace__ == '') && (!has(p2.__namespace__) \|\| p2.__namespace__ == '')) \|\| (has(p1.__namespace__) && has(p2.__namespace__) && p1.__namespace__ == p2.__namespace__ )) && (((!has(p1.sectionName) \|\| p1.sectionName == '') && (!has(p2.sectionName) \|\| p2.sectionName == '')) \|\| ( has(p1.sectionName) && has(p2.sectionName) && p1.sectionName == p2.sectionName)) && (((!has(p1.port) \|\| p1.port == 0) && (!has(p2.port) \|\| p2.port == 0)) \|\| (has(p1.port) && has(p2.port) && p1.port == p2.port))))"> |  | MaxItems: 32 <br /> |
| `hostnames` _[Hostname](#hostname) array_ | Hostnames defines a set of hostnames that should match against the HTTP Host<br />header to select a HTTPRoute used to process the request. Implementations<br />MUST ignore any port value specified in the HTTP Host header while<br />performing a match and (absent of any applicable header modification<br />configuration) MUST forward this header unmodified to the backend.<br />Valid values for Hostnames are determined by RFC 1123 definition of a<br />hostname with 2 notable exceptions:<br />1. IPs are not allowed.<br />2. A hostname may be prefixed with a wildcard label (`*.`). The wildcard<br />   label must appear by itself as the first label.<br />If a hostname is specified by both the Listener and HTTPRoute, there<br />must be at least one intersecting hostname for the HTTPRoute to be<br />attached to the Listener. For example:<br />* A Listener with `test.example.com` as the hostname matches HTTPRoutes<br />  that have either not specified any hostnames, or have specified at<br />  least one of `test.example.com` or `*.example.com`.<br />* A Listener with `*.example.com` as the hostname matches HTTPRoutes<br />  that have either not specified any hostnames or have specified at least<br />  one hostname that matches the Listener hostname. For example,<br />  `*.example.com`, `test.example.com`, and `foo.test.example.com` would<br />  all match. On the other hand, `example.com` and `test.example.net` would<br />  not match.<br />Hostnames that are prefixed with a wildcard label (`*.`) are interpreted<br />as a suffix match. That means that a match for `*.example.com` would match<br />both `test.example.com`, and `foo.test.example.com`, but not `example.com`.<br />If both the Listener and HTTPRoute have specified hostnames, any<br />HTTPRoute hostnames that do not match the Listener hostname MUST be<br />ignored. For example, if a Listener specified `*.example.com`, and the<br />HTTPRoute specified `test.example.com` and `test.example.net`,<br />`test.example.net` must not be considered for a match.<br />If both the Listener and HTTPRoute have specified hostnames, and none<br />match with the criteria above, then the HTTPRoute is not accepted. The<br />implementation must raise an 'Accepted' Condition with a status of<br />`False` in the corresponding RouteParentStatus.<br />In the event that multiple HTTPRoutes specify intersecting hostnames (e.g.<br />overlapping wildcard matching and exact matching hostnames), precedence must<br />be given to rules from the HTTPRoute with the largest number of:<br />* Characters in a matching non-wildcard hostname.<br />* Characters in a matching hostname.<br />If ties exist across multiple Routes, the matching precedence rules for<br />HTTPRouteMatches takes over.<br />Support: Core |  | MaxItems: 16 <br />MaxLength: 253 <br />MinLength: 1 <br />Pattern: `^(\*\.)?[a-z0-9]([-a-z0-9]*[a-z0-9])?(\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*$` <br /> |
| `rules` _[HTTPRouteRule](#httprouterule) array_ | Rules are a list of HTTP matchers, filters and actions.<br /><gateway:experimental:validation:XValidation:message="Rule name must be unique within the route",rule="self.all(l1, !has(l1.name) \|\| self.exists_one(l2, has(l2.name) && l1.name == l2.name))"> | [map[matches:[map[path:map[type:PathPrefix value:/]]]]] | MaxItems: 16 <br /> |


#### HTTPRouteStatus



HTTPRouteStatus defines the observed state of HTTPRoute.



_Appears in:_
- [HTTPRoute](#httproute)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `parents` _[RouteParentStatus](#routeparentstatus) array_ | Parents is a list of parent resources (usually Gateways) that are<br />associated with the route, and the status of the route with respect to<br />each parent. When this route attaches to a parent, the controller that<br />manages the parent must add an entry to this list when the controller<br />first sees the route and should update the entry as appropriate when the<br />route or gateway is modified.<br />Note that parent references that cannot be resolved by an implementation<br />of this API will not be added to this list. Implementations of this API<br />can only populate Route status for the Gateways/parent resources they are<br />responsible for.<br />A maximum of 32 Gateways will be represented in this list. An empty list<br />means the route has not been attached to any Gateway. |  | MaxItems: 32 <br /> |


#### HTTPRouteTimeouts



HTTPRouteTimeouts defines timeouts that can be configured for an HTTPRoute.
Timeout values are represented with Gateway API Duration formatting.



_Appears in:_
- [HTTPRouteRule](#httprouterule)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `request` _[Duration](#duration)_ | Request specifies the maximum duration for a gateway to respond to an HTTP request.<br />If the gateway has not been able to respond before this deadline is met, the gateway<br />MUST return a timeout error.<br />For example, setting the `rules.timeouts.request` field to the value `10s` in an<br />`HTTPRoute` will cause a timeout if a client request is taking longer than 10 seconds<br />to complete.<br />Setting a timeout to the zero duration (e.g. "0s") SHOULD disable the timeout<br />completely. Implementations that cannot completely disable the timeout MUST<br />instead interpret the zero duration as the longest possible value to which<br />the timeout can be set.<br />This timeout is intended to cover as close to the whole request-response transaction<br />as possible although an implementation MAY choose to start the timeout after the entire<br />request stream has been received instead of immediately after the transaction is<br />initiated by the client.<br />The value of Request is a Gateway API Duration string as defined by GEP-2257. When this<br />field is unspecified, request timeout behavior is implementation-specific.<br />Support: Extended |  | Pattern: `^([0-9]\{1,5\}(h\|m\|s\|ms))\{1,4\}$` <br /> |
| `backendRequest` _[Duration](#duration)_ | BackendRequest specifies a timeout for an individual request from the gateway<br />to a backend. This covers the time from when the request first starts being<br />sent from the gateway to when the full response has been received from the backend.<br />Setting a timeout to the zero duration (e.g. "0s") SHOULD disable the timeout<br />completely. Implementations that cannot completely disable the timeout MUST<br />instead interpret the zero duration as the longest possible value to which<br />the timeout can be set.<br />An entire client HTTP transaction with a gateway, covered by the Request timeout,<br />may result in more than one call from the gateway to the destination backend,<br />for example, if automatic retries are supported.<br />The value of BackendRequest must be a Gateway API Duration string as defined by<br />GEP-2257.  When this field is unspecified, its behavior is implementation-specific;<br />when specified, the value of BackendRequest must be no more than the value of the<br />Request timeout (since the Request timeout encompasses the BackendRequest timeout).<br />Support: Extended |  | Pattern: `^([0-9]\{1,5\}(h\|m\|s\|ms))\{1,4\}$` <br /> |


#### HTTPURLRewriteFilter



HTTPURLRewriteFilter defines a filter that modifies a request during
forwarding. At most one of these filters may be used on a Route rule. This
MUST NOT be used on the same Route rule as a HTTPRequestRedirect filter.

Support: Extended



_Appears in:_
- [HTTPRouteFilter](#httproutefilter)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `hostname` _[PreciseHostname](#precisehostname)_ | Hostname is the value to be used to replace the Host header value during<br />forwarding.<br />Support: Extended |  | MaxLength: 253 <br />MinLength: 1 <br />Pattern: `^[a-z0-9]([-a-z0-9]*[a-z0-9])?(\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*$` <br /> |
| `path` _[HTTPPathModifier](#httppathmodifier)_ | Path defines a path rewrite.<br />Support: Extended |  |  |


#### HeaderMatchType

_Underlying type:_ _string_

HeaderMatchType specifies the semantics of how HTTP header values should be
compared. Valid HeaderMatchType values, along with their conformance levels, are:

* "Exact" - Core
* "RegularExpression" - Implementation Specific

Note that values may be added to this enum, implementations
must ensure that unknown values will not cause a crash.

Unknown values here must result in the implementation setting the
Accepted Condition for the Route to `status: False`, with a
Reason of `UnsupportedValue`.

_Validation:_
- Enum: [Exact RegularExpression]

_Appears in:_
- [HTTPHeaderMatch](#httpheadermatch)

| Field | Description |
| --- | --- |
| `Exact` |  |
| `RegularExpression` |  |


#### HeaderName

_Underlying type:_ _string_

HeaderName is the name of a header or query parameter.

_Validation:_
- MaxLength: 256
- MinLength: 1
- Pattern: `^[A-Za-z0-9!#$%&'*+\-.^_\x60|~]+$`

_Appears in:_
- [GRPCHeaderName](#grpcheadername)
- [HTTPHeaderName](#httpheadername)



#### Hostname

_Underlying type:_ _string_

Hostname is the fully qualified domain name of a network host. This matches
the RFC 1123 definition of a hostname with 2 notable exceptions:

 1. IPs are not allowed.
 2. A hostname may be prefixed with a wildcard label (`*.`). The wildcard
    label must appear by itself as the first label.

Hostname can be "precise" which is a domain name without the terminating
dot of a network host (e.g. "foo.example.com") or "wildcard", which is a
domain name prefixed with a single wildcard label (e.g. `*.example.com`).

Note that as per RFC1035 and RFC1123, a *label* must consist of lower case
alphanumeric characters or '-', and must start and end with an alphanumeric
character. No other punctuation is allowed.

_Validation:_
- MaxLength: 253
- MinLength: 1
- Pattern: `^(\*\.)?[a-z0-9]([-a-z0-9]*[a-z0-9])?(\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*$`

_Appears in:_
- [GRPCRouteSpec](#grpcroutespec)
- [HTTPRouteSpec](#httproutespec)
- [Listener](#listener)
- [SubjectAltName](#subjectaltname)



#### Kind

_Underlying type:_ _string_

Kind refers to a Kubernetes Kind.

Valid values include:

* "Service"
* "HTTPRoute"

Invalid values include:

* "invalid/kind" - "/" is an invalid character

_Validation:_
- MaxLength: 63
- MinLength: 1
- Pattern: `^[a-zA-Z]([-a-zA-Z0-9]*[a-zA-Z0-9])?$`

_Appears in:_
- [BackendObjectReference](#backendobjectreference)
- [BackendRef](#backendref)
- [GRPCBackendRef](#grpcbackendref)
- [HTTPBackendRef](#httpbackendref)
- [LocalObjectReference](#localobjectreference)
- [LocalParametersReference](#localparametersreference)
- [ObjectReference](#objectreference)
- [ParametersReference](#parametersreference)
- [ParentReference](#parentreference)
- [RouteGroupKind](#routegroupkind)
- [SecretObjectReference](#secretobjectreference)



#### LabelKey

_Underlying type:_ _string_

LabelKey is the key of a label in the Gateway API. This is used for validation
of maps such as Gateway infrastructure labels. This matches the Kubernetes
"qualified name" validation that is used for labels.

Valid values include:

* example
* example.com
* example.com/path
* example.com/path.html

Invalid values include:

* example~ - "~" is an invalid character
* example.com. - cannot start or end with "."

_Validation:_
- MaxLength: 253
- MinLength: 1
- Pattern: `^([a-z0-9]([-a-z0-9]*[a-z0-9])?(\\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*/)?([A-Za-z0-9][-A-Za-z0-9_.]{0,61})?[A-Za-z0-9]$`

_Appears in:_
- [GatewayInfrastructure](#gatewayinfrastructure)



#### LabelValue

_Underlying type:_ _string_

LabelValue is the value of a label in the Gateway API. This is used for validation
of maps such as Gateway infrastructure labels. This matches the Kubernetes
label validation rules:
* must be 63 characters or less (can be empty),
* unless empty, must begin and end with an alphanumeric character ([a-z0-9A-Z]),
* could contain dashes (-), underscores (_), dots (.), and alphanumerics between.

Valid values include:

* MyValue
* my.name
* 123-my-value

_Validation:_
- MaxLength: 63
- MinLength: 0
- Pattern: `^(([A-Za-z0-9][-A-Za-z0-9_.]*)?[A-Za-z0-9])?$`

_Appears in:_
- [GatewayInfrastructure](#gatewayinfrastructure)



#### Listener



Listener embodies the concept of a logical endpoint where a Gateway accepts
network connections.



_Appears in:_
- [GatewaySpec](#gatewayspec)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `name` _[SectionName](#sectionname)_ | Name is the name of the Listener. This name MUST be unique within a<br />Gateway.<br />Support: Core |  | MaxLength: 253 <br />MinLength: 1 <br />Pattern: `^[a-z0-9]([-a-z0-9]*[a-z0-9])?(\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*$` <br /> |
| `hostname` _[Hostname](#hostname)_ | Hostname specifies the virtual hostname to match for protocol types that<br />define this concept. When unspecified, all hostnames are matched. This<br />field is ignored for protocols that don't require hostname based<br />matching.<br />Implementations MUST apply Hostname matching appropriately for each of<br />the following protocols:<br />* TLS: The Listener Hostname MUST match the SNI.<br />* HTTP: The Listener Hostname MUST match the Host header of the request.<br />* HTTPS: The Listener Hostname SHOULD match both the SNI and Host header.<br />  Note that this does not require the SNI and Host header to be the same.<br />  The semantics of this are described in more detail below.<br />To ensure security, Section 11.1 of RFC-6066 emphasizes that server<br />implementations that rely on SNI hostname matching MUST also verify<br />hostnames within the application protocol.<br />Section 9.1.2 of RFC-7540 provides a mechanism for servers to reject the<br />reuse of a connection by responding with the HTTP 421 Misdirected Request<br />status code. This indicates that the origin server has rejected the<br />request because it appears to have been misdirected.<br />To detect misdirected requests, Gateways SHOULD match the authority of<br />the requests with all the SNI hostname(s) configured across all the<br />Gateway Listeners on the same port and protocol:<br />* If another Listener has an exact match or more specific wildcard entry,<br />  the Gateway SHOULD return a 421.<br />* If the current Listener (selected by SNI matching during ClientHello)<br />  does not match the Host:<br />    * If another Listener does match the Host the Gateway SHOULD return a<br />      421.<br />    * If no other Listener matches the Host, the Gateway MUST return a<br />      404.<br />For HTTPRoute and TLSRoute resources, there is an interaction with the<br />`spec.hostnames` array. When both listener and route specify hostnames,<br />there MUST be an intersection between the values for a Route to be<br />accepted. For more information, refer to the Route specific Hostnames<br />documentation.<br />Hostnames that are prefixed with a wildcard label (`*.`) are interpreted<br />as a suffix match. That means that a match for `*.example.com` would match<br />both `test.example.com`, and `foo.test.example.com`, but not `example.com`.<br />Support: Core |  | MaxLength: 253 <br />MinLength: 1 <br />Pattern: `^(\*\.)?[a-z0-9]([-a-z0-9]*[a-z0-9])?(\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*$` <br /> |
| `port` _[PortNumber](#portnumber)_ | Port is the network port. Multiple listeners may use the<br />same port, subject to the Listener compatibility rules.<br />Support: Core |  | Maximum: 65535 <br />Minimum: 1 <br /> |
| `protocol` _[ProtocolType](#protocoltype)_ | Protocol specifies the network protocol this listener expects to receive.<br />Support: Core |  | MaxLength: 255 <br />MinLength: 1 <br />Pattern: `^[a-zA-Z0-9]([-a-zA-Z0-9]*[a-zA-Z0-9])?$\|[a-z0-9]([-a-z0-9]*[a-z0-9])?(\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*\/[A-Za-z0-9]+$` <br /> |
| `tls` _[GatewayTLSConfig](#gatewaytlsconfig)_ | TLS is the TLS configuration for the Listener. This field is required if<br />the Protocol field is "HTTPS" or "TLS". It is invalid to set this field<br />if the Protocol field is "HTTP", "TCP", or "UDP".<br />The association of SNIs to Certificate defined in GatewayTLSConfig is<br />defined based on the Hostname field for this listener.<br />The GatewayClass MUST use the longest matching SNI out of all<br />available certificates for any TLS handshake.<br />Support: Core |  |  |
| `allowedRoutes` _[AllowedRoutes](#allowedroutes)_ | AllowedRoutes defines the types of routes that MAY be attached to a<br />Listener and the trusted namespaces where those Route resources MAY be<br />present.<br />Although a client request may match multiple route rules, only one rule<br />may ultimately receive the request. Matching precedence MUST be<br />determined in order of the following criteria:<br />* The most specific match as defined by the Route type.<br />* The oldest Route based on creation timestamp. For example, a Route with<br />  a creation timestamp of "2020-09-08 01:02:03" is given precedence over<br />  a Route with a creation timestamp of "2020-09-08 01:02:04".<br />* If everything else is equivalent, the Route appearing first in<br />  alphabetical order (namespace/name) should be given precedence. For<br />  example, foo/bar is given precedence over foo/baz.<br />All valid rules within a Route attached to this Listener should be<br />implemented. Invalid Route rules can be ignored (sometimes that will mean<br />the full Route). If a Route rule transitions from valid to invalid,<br />support for that Route rule should be dropped to ensure consistency. For<br />example, even if a filter specified by a Route rule is invalid, the rest<br />of the rules within that Route should still be supported.<br />Support: Core | \{ namespaces:map[from:Same] \} |  |






#### ListenerNamespaces



ListenerNamespaces indicate which namespaces ListenerSets should be selected from.



_Appears in:_
- [AllowedListeners](#allowedlisteners)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `from` _[FromNamespaces](#fromnamespaces)_ | From indicates where ListenerSets can attach to this Gateway. Possible<br />values are:<br />* Same: Only ListenerSets in the same namespace may be attached to this Gateway.<br />* Selector: ListenerSets in namespaces selected by the selector may be attached to this Gateway.<br />* All: ListenerSets in all namespaces may be attached to this Gateway.<br />* None: Only listeners defined in the Gateway's spec are allowed<br />While this feature is experimental, the default value None | None | Enum: [All Selector Same None] <br /> |
| `selector` _[LabelSelector](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.32/#labelselector-v1-meta)_ | Selector must be specified when From is set to "Selector". In that case,<br />only ListenerSets in Namespaces matching this Selector will be selected by this<br />Gateway. This field is ignored for other values of "From". |  |  |


#### ListenerStatus



ListenerStatus is the status associated with a Listener.



_Appears in:_
- [GatewayStatus](#gatewaystatus)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `name` _[SectionName](#sectionname)_ | Name is the name of the Listener that this status corresponds to. |  | MaxLength: 253 <br />MinLength: 1 <br />Pattern: `^[a-z0-9]([-a-z0-9]*[a-z0-9])?(\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*$` <br /> |
| `supportedKinds` _[RouteGroupKind](#routegroupkind) array_ | SupportedKinds is the list indicating the Kinds supported by this<br />listener. This MUST represent the kinds an implementation supports for<br />that Listener configuration.<br />If kinds are specified in Spec that are not supported, they MUST NOT<br />appear in this list and an implementation MUST set the "ResolvedRefs"<br />condition to "False" with the "InvalidRouteKinds" reason. If both valid<br />and invalid Route kinds are specified, the implementation MUST<br />reference the valid Route kinds that have been specified. |  | MaxItems: 8 <br /> |
| `attachedRoutes` _integer_ | AttachedRoutes represents the total number of Routes that have been<br />successfully attached to this Listener.<br />Successful attachment of a Route to a Listener is based solely on the<br />combination of the AllowedRoutes field on the corresponding Listener<br />and the Route's ParentRefs field. A Route is successfully attached to<br />a Listener when it is selected by the Listener's AllowedRoutes field<br />AND the Route has a valid ParentRef selecting the whole Gateway<br />resource or a specific Listener as a parent resource (more detail on<br />attachment semantics can be found in the documentation on the various<br />Route kinds ParentRefs fields). Listener or Route status does not impact<br />successful attachment, i.e. the AttachedRoutes field count MUST be set<br />for Listeners with condition Accepted: false and MUST count successfully<br />attached Routes that may themselves have Accepted: false conditions.<br />Uses for this field include troubleshooting Route attachment and<br />measuring blast radius/impact of changes to a Listener. |  |  |
| `conditions` _[Condition](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.32/#condition-v1-meta) array_ | Conditions describe the current condition of this listener. |  | MaxItems: 8 <br /> |


#### LocalObjectReference



LocalObjectReference identifies an API object within the namespace of the
referrer.
The API object must be valid in the cluster; the Group and Kind must
be registered in the cluster for this reference to be valid.

References to objects with invalid Group and Kind are not valid, and must
be rejected by the implementation, with appropriate Conditions set
on the containing object.



_Appears in:_
- [BackendTLSPolicyValidation](#backendtlspolicyvalidation)
- [GRPCRouteFilter](#grpcroutefilter)
- [HTTPRouteFilter](#httproutefilter)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `group` _[Group](#group)_ | Group is the group of the referent. For example, "gateway.networking.k8s.io".<br />When unspecified or empty string, core API group is inferred. |  | MaxLength: 253 <br />Pattern: `^$\|^[a-z0-9]([-a-z0-9]*[a-z0-9])?(\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*$` <br /> |
| `kind` _[Kind](#kind)_ | Kind is kind of the referent. For example "HTTPRoute" or "Service". |  | MaxLength: 63 <br />MinLength: 1 <br />Pattern: `^[a-zA-Z]([-a-zA-Z0-9]*[a-zA-Z0-9])?$` <br /> |
| `name` _[ObjectName](#objectname)_ | Name is the name of the referent. |  | MaxLength: 253 <br />MinLength: 1 <br /> |


#### LocalParametersReference



LocalParametersReference identifies an API object containing controller-specific
configuration resource within the namespace.



_Appears in:_
- [GatewayInfrastructure](#gatewayinfrastructure)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `group` _[Group](#group)_ | Group is the group of the referent. |  | MaxLength: 253 <br />Pattern: `^$\|^[a-z0-9]([-a-z0-9]*[a-z0-9])?(\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*$` <br /> |
| `kind` _[Kind](#kind)_ | Kind is kind of the referent. |  | MaxLength: 63 <br />MinLength: 1 <br />Pattern: `^[a-zA-Z]([-a-zA-Z0-9]*[a-zA-Z0-9])?$` <br /> |
| `name` _string_ | Name is the name of the referent. |  | MaxLength: 253 <br />MinLength: 1 <br /> |


#### Namespace

_Underlying type:_ _string_

Namespace refers to a Kubernetes namespace. It must be a RFC 1123 label.

This validation is based off of the corresponding Kubernetes validation:
https://github.com/kubernetes/apimachinery/blob/02cfb53916346d085a6c6c7c66f882e3c6b0eca6/pkg/util/validation/validation.go#L187

This is used for Namespace name validation here:
https://github.com/kubernetes/apimachinery/blob/02cfb53916346d085a6c6c7c66f882e3c6b0eca6/pkg/api/validation/generic.go#L63

Valid values include:

* "example"

Invalid values include:

* "example.com" - "." is an invalid character

_Validation:_
- MaxLength: 63
- MinLength: 1
- Pattern: `^[a-z0-9]([-a-z0-9]*[a-z0-9])?$`

_Appears in:_
- [BackendObjectReference](#backendobjectreference)
- [BackendRef](#backendref)
- [GRPCBackendRef](#grpcbackendref)
- [HTTPBackendRef](#httpbackendref)
- [ObjectReference](#objectreference)
- [ParametersReference](#parametersreference)
- [ParentReference](#parentreference)
- [SecretObjectReference](#secretobjectreference)



#### ObjectName

_Underlying type:_ _string_

ObjectName refers to the name of a Kubernetes object.
Object names can have a variety of forms, including RFC 1123 subdomains,
RFC 1123 labels, or RFC 1035 labels.

_Validation:_
- MaxLength: 253
- MinLength: 1

_Appears in:_
- [BackendObjectReference](#backendobjectreference)
- [BackendRef](#backendref)
- [GRPCBackendRef](#grpcbackendref)
- [GatewaySpec](#gatewayspec)
- [HTTPBackendRef](#httpbackendref)
- [LocalObjectReference](#localobjectreference)
- [ObjectReference](#objectreference)
- [ParentReference](#parentreference)
- [SecretObjectReference](#secretobjectreference)



#### ObjectReference



ObjectReference identifies an API object including its namespace.

The API object must be valid in the cluster; the Group and Kind must
be registered in the cluster for this reference to be valid.

References to objects with invalid Group and Kind are not valid, and must
be rejected by the implementation, with appropriate Conditions set
on the containing object.



_Appears in:_
- [FrontendTLSValidation](#frontendtlsvalidation)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `group` _[Group](#group)_ | Group is the group of the referent. For example, "gateway.networking.k8s.io".<br />When set to the empty string, core API group is inferred. |  | MaxLength: 253 <br />Pattern: `^$\|^[a-z0-9]([-a-z0-9]*[a-z0-9])?(\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*$` <br /> |
| `kind` _[Kind](#kind)_ | Kind is kind of the referent. For example "ConfigMap" or "Service". |  | MaxLength: 63 <br />MinLength: 1 <br />Pattern: `^[a-zA-Z]([-a-zA-Z0-9]*[a-zA-Z0-9])?$` <br /> |
| `name` _[ObjectName](#objectname)_ | Name is the name of the referent. |  | MaxLength: 253 <br />MinLength: 1 <br /> |
| `namespace` _[Namespace](#namespace)_ | Namespace is the namespace of the referenced object. When unspecified, the local<br />namespace is inferred.<br />Note that when a namespace different than the local namespace is specified,<br />a ReferenceGrant object is required in the referent namespace to allow that<br />namespace's owner to accept the reference. See the ReferenceGrant<br />documentation for details.<br />Support: Core |  | MaxLength: 63 <br />MinLength: 1 <br />Pattern: `^[a-z0-9]([-a-z0-9]*[a-z0-9])?$` <br /> |


#### ParametersReference



ParametersReference identifies an API object containing controller-specific
configuration resource within the cluster.



_Appears in:_
- [GatewayClassSpec](#gatewayclassspec)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `group` _[Group](#group)_ | Group is the group of the referent. |  | MaxLength: 253 <br />Pattern: `^$\|^[a-z0-9]([-a-z0-9]*[a-z0-9])?(\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*$` <br /> |
| `kind` _[Kind](#kind)_ | Kind is kind of the referent. |  | MaxLength: 63 <br />MinLength: 1 <br />Pattern: `^[a-zA-Z]([-a-zA-Z0-9]*[a-zA-Z0-9])?$` <br /> |
| `name` _string_ | Name is the name of the referent. |  | MaxLength: 253 <br />MinLength: 1 <br /> |
| `namespace` _[Namespace](#namespace)_ | Namespace is the namespace of the referent.<br />This field is required when referring to a Namespace-scoped resource and<br />MUST be unset when referring to a Cluster-scoped resource. |  | MaxLength: 63 <br />MinLength: 1 <br />Pattern: `^[a-z0-9]([-a-z0-9]*[a-z0-9])?$` <br /> |


#### ParentReference



ParentReference identifies an API object (usually a Gateway) that can be considered
a parent of this resource (usually a route). There are two kinds of parent resources
with "Core" support:

* Gateway (Gateway conformance profile)
* Service (Mesh conformance profile, ClusterIP Services only)

This API may be extended in the future to support additional kinds of parent
resources.

The API object must be valid in the cluster; the Group and Kind must
be registered in the cluster for this reference to be valid.



_Appears in:_
- [CommonRouteSpec](#commonroutespec)
- [GRPCRouteSpec](#grpcroutespec)
- [HTTPRouteSpec](#httproutespec)
- [RouteParentStatus](#routeparentstatus)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `group` _[Group](#group)_ | Group is the group of the referent.<br />When unspecified, "gateway.networking.k8s.io" is inferred.<br />To set the core API group (such as for a "Service" kind referent),<br />Group must be explicitly set to "" (empty string).<br />Support: Core | gateway.networking.k8s.io | MaxLength: 253 <br />Pattern: `^$\|^[a-z0-9]([-a-z0-9]*[a-z0-9])?(\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*$` <br /> |
| `kind` _[Kind](#kind)_ | Kind is kind of the referent.<br />There are two kinds of parent resources with "Core" support:<br />* Gateway (Gateway conformance profile)<br />* Service (Mesh conformance profile, ClusterIP Services only)<br />Support for other resources is Implementation-Specific. | Gateway | MaxLength: 63 <br />MinLength: 1 <br />Pattern: `^[a-zA-Z]([-a-zA-Z0-9]*[a-zA-Z0-9])?$` <br /> |
| `namespace` _[Namespace](#namespace)_ | Namespace is the namespace of the referent. When unspecified, this refers<br />to the local namespace of the Route.<br />Note that there are specific rules for ParentRefs which cross namespace<br />boundaries. Cross-namespace references are only valid if they are explicitly<br />allowed by something in the namespace they are referring to. For example:<br />Gateway has the AllowedRoutes field, and ReferenceGrant provides a<br />generic way to enable any other kind of cross-namespace reference.<br /><gateway:experimental:description><br />ParentRefs from a Route to a Service in the same namespace are "producer"<br />routes, which apply default routing rules to inbound connections from<br />any namespace to the Service.<br />ParentRefs from a Route to a Service in a different namespace are<br />"consumer" routes, and these routing rules are only applied to outbound<br />connections originating from the same namespace as the Route, for which<br />the intended destination of the connections are a Service targeted as a<br />ParentRef of the Route.<br /></gateway:experimental:description><br />Support: Core |  | MaxLength: 63 <br />MinLength: 1 <br />Pattern: `^[a-z0-9]([-a-z0-9]*[a-z0-9])?$` <br /> |
| `name` _[ObjectName](#objectname)_ | Name is the name of the referent.<br />Support: Core |  | MaxLength: 253 <br />MinLength: 1 <br /> |
| `sectionName` _[SectionName](#sectionname)_ | SectionName is the name of a section within the target resource. In the<br />following resources, SectionName is interpreted as the following:<br />* Gateway: Listener name. When both Port (experimental) and SectionName<br />are specified, the name and port of the selected listener must match<br />both specified values.<br />* Service: Port name. When both Port (experimental) and SectionName<br />are specified, the name and port of the selected listener must match<br />both specified values.<br />Implementations MAY choose to support attaching Routes to other resources.<br />If that is the case, they MUST clearly document how SectionName is<br />interpreted.<br />When unspecified (empty string), this will reference the entire resource.<br />For the purpose of status, an attachment is considered successful if at<br />least one section in the parent resource accepts it. For example, Gateway<br />listeners can restrict which Routes can attach to them by Route kind,<br />namespace, or hostname. If 1 of 2 Gateway listeners accept attachment from<br />the referencing Route, the Route MUST be considered successfully<br />attached. If no Gateway listeners accept attachment from this Route, the<br />Route MUST be considered detached from the Gateway.<br />Support: Core |  | MaxLength: 253 <br />MinLength: 1 <br />Pattern: `^[a-z0-9]([-a-z0-9]*[a-z0-9])?(\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*$` <br /> |
| `port` _[PortNumber](#portnumber)_ | Port is the network port this Route targets. It can be interpreted<br />differently based on the type of parent resource.<br />When the parent resource is a Gateway, this targets all listeners<br />listening on the specified port that also support this kind of Route(and<br />select this Route). It's not recommended to set `Port` unless the<br />networking behaviors specified in a Route must apply to a specific port<br />as opposed to a listener(s) whose port(s) may be changed. When both Port<br />and SectionName are specified, the name and port of the selected listener<br />must match both specified values.<br /><gateway:experimental:description><br />When the parent resource is a Service, this targets a specific port in the<br />Service spec. When both Port (experimental) and SectionName are specified,<br />the name and port of the selected port must match both specified values.<br /></gateway:experimental:description><br />Implementations MAY choose to support other parent resources.<br />Implementations supporting other types of parent resources MUST clearly<br />document how/if Port is interpreted.<br />For the purpose of status, an attachment is considered successful as<br />long as the parent resource accepts it partially. For example, Gateway<br />listeners can restrict which Routes can attach to them by Route kind,<br />namespace, or hostname. If 1 of 2 Gateway listeners accept attachment<br />from the referencing Route, the Route MUST be considered successfully<br />attached. If no Gateway listeners accept attachment from this Route,<br />the Route MUST be considered detached from the Gateway.<br />Support: Extended |  | Maximum: 65535 <br />Minimum: 1 <br /> |


#### PathMatchType

_Underlying type:_ _string_

PathMatchType specifies the semantics of how HTTP paths should be compared.
Valid PathMatchType values, along with their support levels, are:

* "Exact" - Core
* "PathPrefix" - Core
* "RegularExpression" - Implementation Specific

PathPrefix and Exact paths must be syntactically valid:

- Must begin with the `/` character
- Must not contain consecutive `/` characters (e.g. `/foo///`, `//`).

Note that values may be added to this enum, implementations
must ensure that unknown values will not cause a crash.

Unknown values here must result in the implementation setting the
Accepted Condition for the Route to `status: False`, with a
Reason of `UnsupportedValue`.

_Validation:_
- Enum: [Exact PathPrefix RegularExpression]

_Appears in:_
- [HTTPPathMatch](#httppathmatch)

| Field | Description |
| --- | --- |
| `Exact` | Matches the URL path exactly and with case sensitivity. This means that<br />an exact path match on `/abc` will only match requests to `/abc`, NOT<br />`/abc/`, `/Abc`, or `/abcd`.<br /> |
| `PathPrefix` | Matches based on a URL path prefix split by `/`. Matching is<br />case-sensitive and done on a path element by element basis. A<br />path element refers to the list of labels in the path split by<br />the `/` separator. When specified, a trailing `/` is ignored.<br />For example, the paths `/abc`, `/abc/`, and `/abc/def` would all match<br />the prefix `/abc`, but the path `/abcd` would not.<br />"PathPrefix" is semantically equivalent to the "Prefix" path type in the<br />Kubernetes Ingress API.<br /> |
| `RegularExpression` | Matches if the URL path matches the given regular expression with<br />case sensitivity.<br />Since `"RegularExpression"` has implementation-specific conformance,<br />implementations can support POSIX, PCRE, RE2 or any other regular expression<br />dialect.<br />Please read the implementation's documentation to determine the supported<br />dialect.<br /> |


#### PortNumber

_Underlying type:_ _integer_

PortNumber defines a network port.

_Validation:_
- Maximum: 65535
- Minimum: 1

_Appears in:_
- [BackendObjectReference](#backendobjectreference)
- [BackendRef](#backendref)
- [GRPCBackendRef](#grpcbackendref)
- [HTTPBackendRef](#httpbackendref)
- [HTTPRequestRedirectFilter](#httprequestredirectfilter)
- [Listener](#listener)
- [ParentReference](#parentreference)



#### PreciseHostname

_Underlying type:_ _string_

PreciseHostname is the fully qualified domain name of a network host. This
matches the RFC 1123 definition of a hostname with 1 notable exception that
numeric IP addresses are not allowed.

Note that as per RFC1035 and RFC1123, a *label* must consist of lower case
alphanumeric characters or '-', and must start and end with an alphanumeric
character. No other punctuation is allowed.

_Validation:_
- MaxLength: 253
- MinLength: 1
- Pattern: `^[a-z0-9]([-a-z0-9]*[a-z0-9])?(\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*$`

_Appears in:_
- [BackendTLSPolicyValidation](#backendtlspolicyvalidation)
- [HTTPRequestRedirectFilter](#httprequestredirectfilter)
- [HTTPURLRewriteFilter](#httpurlrewritefilter)



#### ProtocolType

_Underlying type:_ _string_

ProtocolType defines the application protocol accepted by a Listener.
Implementations are not required to accept all the defined protocols. If an
implementation does not support a specified protocol, it MUST set the
"Accepted" condition to False for the affected Listener with a reason of
"UnsupportedProtocol".

Core ProtocolType values are listed in the table below.

Implementations can define their own protocols if a core ProtocolType does not
exist. Such definitions must use prefixed name, such as
`mycompany.com/my-custom-protocol`. Un-prefixed names are reserved for core
protocols. Any protocol defined by implementations will fall under
Implementation-specific conformance.

Valid values include:

* "HTTP" - Core support
* "example.com/bar" - Implementation-specific support

Invalid values include:

* "example.com" - must include path if domain is used
* "foo.example.com" - must include path if domain is used

_Validation:_
- MaxLength: 255
- MinLength: 1
- Pattern: `^[a-zA-Z0-9]([-a-zA-Z0-9]*[a-zA-Z0-9])?$|[a-z0-9]([-a-z0-9]*[a-z0-9])?(\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*\/[A-Za-z0-9]+$`

_Appears in:_
- [Listener](#listener)

| Field | Description |
| --- | --- |
| `HTTP` | Accepts cleartext HTTP/1.1 sessions over TCP. Implementations MAY also<br />support HTTP/2 over cleartext. If implementations support HTTP/2 over<br />cleartext on "HTTP" listeners, that MUST be clearly documented by the<br />implementation.<br /> |
| `HTTPS` | Accepts HTTP/1.1 or HTTP/2 sessions over TLS.<br /> |
| `TLS` | Accepts TLS sessions over TCP.<br /> |
| `TCP` | Accepts TCP sessions.<br /> |
| `UDP` | Accepts UDP packets.<br /> |


#### QueryParamMatchType

_Underlying type:_ _string_

QueryParamMatchType specifies the semantics of how HTTP query parameter
values should be compared. Valid QueryParamMatchType values, along with their
conformance levels, are:

* "Exact" - Core
* "RegularExpression" - Implementation Specific

Note that values may be added to this enum, implementations
must ensure that unknown values will not cause a crash.

Unknown values here must result in the implementation setting the
Accepted Condition for the Route to `status: False`, with a
Reason of `UnsupportedValue`.

_Validation:_
- Enum: [Exact RegularExpression]

_Appears in:_
- [HTTPQueryParamMatch](#httpqueryparammatch)

| Field | Description |
| --- | --- |
| `Exact` |  |
| `RegularExpression` |  |






#### RouteGroupKind



RouteGroupKind indicates the group and kind of a Route resource.



_Appears in:_
- [AllowedRoutes](#allowedroutes)
- [ListenerStatus](#listenerstatus)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `group` _[Group](#group)_ | Group is the group of the Route. | gateway.networking.k8s.io | MaxLength: 253 <br />Pattern: `^$\|^[a-z0-9]([-a-z0-9]*[a-z0-9])?(\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*$` <br /> |
| `kind` _[Kind](#kind)_ | Kind is the kind of the Route. |  | MaxLength: 63 <br />MinLength: 1 <br />Pattern: `^[a-zA-Z]([-a-zA-Z0-9]*[a-zA-Z0-9])?$` <br /> |


#### RouteNamespaces



RouteNamespaces indicate which namespaces Routes should be selected from.



_Appears in:_
- [AllowedRoutes](#allowedroutes)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `from` _[FromNamespaces](#fromnamespaces)_ | From indicates where Routes will be selected for this Gateway. Possible<br />values are:<br />* All: Routes in all namespaces may be used by this Gateway.<br />* Selector: Routes in namespaces selected by the selector may be used by<br />  this Gateway.<br />* Same: Only Routes in the same namespace may be used by this Gateway.<br />Support: Core | Same | Enum: [All Selector Same] <br /> |
| `selector` _[LabelSelector](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.32/#labelselector-v1-meta)_ | Selector must be specified when From is set to "Selector". In that case,<br />only Routes in Namespaces matching this Selector will be selected by this<br />Gateway. This field is ignored for other values of "From".<br />Support: Core |  |  |


#### RouteParentStatus



RouteParentStatus describes the status of a route with respect to an
associated Parent.



_Appears in:_
- [GRPCRouteStatus](#grpcroutestatus)
- [HTTPRouteStatus](#httproutestatus)
- [RouteStatus](#routestatus)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `parentRef` _[ParentReference](#parentreference)_ | ParentRef corresponds with a ParentRef in the spec that this<br />RouteParentStatus struct describes the status of. |  |  |
| `controllerName` _[GatewayController](#gatewaycontroller)_ | ControllerName is a domain/path string that indicates the name of the<br />controller that wrote this status. This corresponds with the<br />controllerName field on GatewayClass.<br />Example: "example.net/gateway-controller".<br />The format of this field is DOMAIN "/" PATH, where DOMAIN and PATH are<br />valid Kubernetes names<br />(https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#names).<br />Controllers MUST populate this field when writing status. Controllers should ensure that<br />entries to status populated with their ControllerName are cleaned up when they are no<br />longer necessary. |  | MaxLength: 253 <br />MinLength: 1 <br />Pattern: `^[a-z0-9]([-a-z0-9]*[a-z0-9])?(\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*\/[A-Za-z0-9\/\-._~%!$&'()*+,;=:]+$` <br /> |
| `conditions` _[Condition](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.32/#condition-v1-meta) array_ | Conditions describes the status of the route with respect to the Gateway.<br />Note that the route's availability is also subject to the Gateway's own<br />status conditions and listener status.<br />If the Route's ParentRef specifies an existing Gateway that supports<br />Routes of this kind AND that Gateway's controller has sufficient access,<br />then that Gateway's controller MUST set the "Accepted" condition on the<br />Route, to indicate whether the route has been accepted or rejected by the<br />Gateway, and why.<br />A Route MUST be considered "Accepted" if at least one of the Route's<br />rules is implemented by the Gateway.<br />There are a number of cases where the "Accepted" condition may not be set<br />due to lack of controller visibility, that includes when:<br />* The Route refers to a nonexistent parent.<br />* The Route is of a type that the controller does not support.<br />* The Route is in a namespace the controller does not have access to. |  | MaxItems: 8 <br />MinItems: 1 <br /> |


#### RouteStatus



RouteStatus defines the common attributes that all Routes MUST include within
their status.



_Appears in:_
- [GRPCRouteStatus](#grpcroutestatus)
- [HTTPRouteStatus](#httproutestatus)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `parents` _[RouteParentStatus](#routeparentstatus) array_ | Parents is a list of parent resources (usually Gateways) that are<br />associated with the route, and the status of the route with respect to<br />each parent. When this route attaches to a parent, the controller that<br />manages the parent must add an entry to this list when the controller<br />first sees the route and should update the entry as appropriate when the<br />route or gateway is modified.<br />Note that parent references that cannot be resolved by an implementation<br />of this API will not be added to this list. Implementations of this API<br />can only populate Route status for the Gateways/parent resources they are<br />responsible for.<br />A maximum of 32 Gateways will be represented in this list. An empty list<br />means the route has not been attached to any Gateway. |  | MaxItems: 32 <br /> |


#### SecretObjectReference



SecretObjectReference identifies an API object including its namespace,
defaulting to Secret.

The API object must be valid in the cluster; the Group and Kind must
be registered in the cluster for this reference to be valid.

References to objects with invalid Group and Kind are not valid, and must
be rejected by the implementation, with appropriate Conditions set
on the containing object.



_Appears in:_
- [GatewayBackendTLS](#gatewaybackendtls)
- [GatewayTLSConfig](#gatewaytlsconfig)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `group` _[Group](#group)_ | Group is the group of the referent. For example, "gateway.networking.k8s.io".<br />When unspecified or empty string, core API group is inferred. |  | MaxLength: 253 <br />Pattern: `^$\|^[a-z0-9]([-a-z0-9]*[a-z0-9])?(\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*$` <br /> |
| `kind` _[Kind](#kind)_ | Kind is kind of the referent. For example "Secret". | Secret | MaxLength: 63 <br />MinLength: 1 <br />Pattern: `^[a-zA-Z]([-a-zA-Z0-9]*[a-zA-Z0-9])?$` <br /> |
| `name` _[ObjectName](#objectname)_ | Name is the name of the referent. |  | MaxLength: 253 <br />MinLength: 1 <br /> |
| `namespace` _[Namespace](#namespace)_ | Namespace is the namespace of the referenced object. When unspecified, the local<br />namespace is inferred.<br />Note that when a namespace different than the local namespace is specified,<br />a ReferenceGrant object is required in the referent namespace to allow that<br />namespace's owner to accept the reference. See the ReferenceGrant<br />documentation for details.<br />Support: Core |  | MaxLength: 63 <br />MinLength: 1 <br />Pattern: `^[a-z0-9]([-a-z0-9]*[a-z0-9])?$` <br /> |


#### SectionName

_Underlying type:_ _string_

SectionName is the name of a section in a Kubernetes resource.

In the following resources, SectionName is interpreted as the following:

* Gateway: Listener name
* HTTPRoute: HTTPRouteRule name
* Service: Port name

Section names can have a variety of forms, including RFC 1123 subdomains,
RFC 1123 labels, or RFC 1035 labels.

This validation is based off of the corresponding Kubernetes validation:
https://github.com/kubernetes/apimachinery/blob/02cfb53916346d085a6c6c7c66f882e3c6b0eca6/pkg/util/validation/validation.go#L208

Valid values include:

* "example"
* "foo-example"
* "example.com"
* "foo.example.com"

Invalid values include:

* "example.com/bar" - "/" is an invalid character

_Validation:_
- MaxLength: 253
- MinLength: 1
- Pattern: `^[a-z0-9]([-a-z0-9]*[a-z0-9])?(\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*$`

_Appears in:_
- [GRPCRouteRule](#grpcrouterule)
- [HTTPRouteRule](#httprouterule)
- [Listener](#listener)
- [ListenerStatus](#listenerstatus)
- [ParentReference](#parentreference)



#### SessionPersistence



SessionPersistence defines the desired state of SessionPersistence.



_Appears in:_
- [GRPCRouteRule](#grpcrouterule)
- [HTTPRouteRule](#httprouterule)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `sessionName` _string_ | SessionName defines the name of the persistent session token<br />which may be reflected in the cookie or the header. Users<br />should avoid reusing session names to prevent unintended<br />consequences, such as rejection or unpredictable behavior.<br />Support: Implementation-specific |  | MaxLength: 128 <br /> |
| `absoluteTimeout` _[Duration](#duration)_ | AbsoluteTimeout defines the absolute timeout of the persistent<br />session. Once the AbsoluteTimeout duration has elapsed, the<br />session becomes invalid.<br />Support: Extended |  | Pattern: `^([0-9]\{1,5\}(h\|m\|s\|ms))\{1,4\}$` <br /> |
| `idleTimeout` _[Duration](#duration)_ | IdleTimeout defines the idle timeout of the persistent session.<br />Once the session has been idle for more than the specified<br />IdleTimeout duration, the session becomes invalid.<br />Support: Extended |  | Pattern: `^([0-9]\{1,5\}(h\|m\|s\|ms))\{1,4\}$` <br /> |
| `type` _[SessionPersistenceType](#sessionpersistencetype)_ | Type defines the type of session persistence such as through<br />the use a header or cookie. Defaults to cookie based session<br />persistence.<br />Support: Core for "Cookie" type<br />Support: Extended for "Header" type | Cookie | Enum: [Cookie Header] <br /> |
| `cookieConfig` _[CookieConfig](#cookieconfig)_ | CookieConfig provides configuration settings that are specific<br />to cookie-based session persistence.<br />Support: Core |  |  |


#### SessionPersistenceType

_Underlying type:_ _string_



_Validation:_
- Enum: [Cookie Header]

_Appears in:_
- [SessionPersistence](#sessionpersistence)

| Field | Description |
| --- | --- |
| `Cookie` | CookieBasedSessionPersistence specifies cookie-based session<br />persistence.<br />Support: Core<br /> |
| `Header` | HeaderBasedSessionPersistence specifies header-based session<br />persistence.<br />Support: Extended<br /> |


#### SupportedFeature







_Appears in:_
- [GatewayClassStatus](#gatewayclassstatus)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `name` _[FeatureName](#featurename)_ |  |  |  |


#### TLSModeType

_Underlying type:_ _string_

TLSModeType type defines how a Gateway handles TLS sessions.

_Validation:_
- Enum: [Terminate Passthrough]

_Appears in:_
- [GatewayTLSConfig](#gatewaytlsconfig)

| Field | Description |
| --- | --- |
| `Terminate` | In this mode, TLS session between the downstream client<br />and the Gateway is terminated at the Gateway.<br /> |
| `Passthrough` | In this mode, the TLS session is NOT terminated by the Gateway. This<br />implies that the Gateway can't decipher the TLS stream except for<br />the ClientHello message of the TLS protocol.<br />Note that SSL passthrough is only supported by TLSRoute.<br /> |


#### TrueField

_Underlying type:_ _boolean_

TrueField is a boolean value that can only be set to true

_Validation:_
- Enum: [true]

_Appears in:_
- [HTTPCORSFilter](#httpcorsfilter)




## gateway.networking.k8s.io/v1alpha2

Package v1alpha2 contains API Schema definitions for the
gateway.networking.k8s.io API group.


### Resource Types
- [GRPCRoute](#grpcroute)
- [ReferenceGrant](#referencegrant)
- [TCPRoute](#tcproute)
- [TLSRoute](#tlsroute)
- [UDPRoute](#udproute)

















#### GRPCRoute

_Underlying type:_ _[GRPCRoute](#grpcroute)_







| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `apiVersion` _string_ | `gateway.networking.k8s.io/v1alpha2` | | |
| `kind` _string_ | `GRPCRoute` | | |
| `metadata` _[ObjectMeta](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.32/#objectmeta-v1-meta)_ | Refer to Kubernetes API documentation for fields of `metadata`. |  |  |
| `spec` _[GRPCRouteSpec](#grpcroutespec)_ | Spec defines the desired state of GRPCRoute. |  |  |
| `status` _[GRPCRouteStatus](#grpcroutestatus)_ | Status defines the current state of GRPCRoute. |  |  |












#### LocalPolicyTargetReference



LocalPolicyTargetReference identifies an API object to apply a direct or
inherited policy to. This should be used as part of Policy resources
that can target Gateway API resources. For more information on how this
policy attachment model works, and a sample Policy resource, refer to
the policy attachment documentation for Gateway API.



_Appears in:_
- [LocalPolicyTargetReferenceWithSectionName](#localpolicytargetreferencewithsectionname)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `group` _[Group](#group)_ | Group is the group of the target resource. |  | MaxLength: 253 <br />Pattern: `^$\|^[a-z0-9]([-a-z0-9]*[a-z0-9])?(\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*$` <br /> |
| `kind` _[Kind](#kind)_ | Kind is kind of the target resource. |  | MaxLength: 63 <br />MinLength: 1 <br />Pattern: `^[a-zA-Z]([-a-zA-Z0-9]*[a-zA-Z0-9])?$` <br /> |
| `name` _[ObjectName](#objectname)_ | Name is the name of the target resource. |  | MaxLength: 253 <br />MinLength: 1 <br /> |


#### LocalPolicyTargetReferenceWithSectionName



LocalPolicyTargetReferenceWithSectionName identifies an API object to apply a
direct policy to. This should be used as part of Policy resources that can
target single resources. For more information on how this policy attachment
mode works, and a sample Policy resource, refer to the policy attachment
documentation for Gateway API.

Note: This should only be used for direct policy attachment when references
to SectionName are actually needed. In all other cases,
LocalPolicyTargetReference should be used.



_Appears in:_
- [BackendTLSPolicySpec](#backendtlspolicyspec)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `group` _[Group](#group)_ | Group is the group of the target resource. |  | MaxLength: 253 <br />Pattern: `^$\|^[a-z0-9]([-a-z0-9]*[a-z0-9])?(\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*$` <br /> |
| `kind` _[Kind](#kind)_ | Kind is kind of the target resource. |  | MaxLength: 63 <br />MinLength: 1 <br />Pattern: `^[a-zA-Z]([-a-zA-Z0-9]*[a-zA-Z0-9])?$` <br /> |
| `name` _[ObjectName](#objectname)_ | Name is the name of the target resource. |  | MaxLength: 253 <br />MinLength: 1 <br /> |
| `sectionName` _[SectionName](#sectionname)_ | SectionName is the name of a section within the target resource. When<br />unspecified, this targetRef targets the entire resource. In the following<br />resources, SectionName is interpreted as the following:<br />* Gateway: Listener name<br />* HTTPRoute: HTTPRouteRule name<br />* Service: Port name<br />If a SectionName is specified, but does not exist on the targeted object,<br />the Policy must fail to attach, and the policy implementation should record<br />a `ResolvedRefs` or similar Condition in the Policy's status. |  | MaxLength: 253 <br />MinLength: 1 <br />Pattern: `^[a-z0-9]([-a-z0-9]*[a-z0-9])?(\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*$` <br /> |










#### PolicyAncestorStatus



PolicyAncestorStatus describes the status of a route with respect to an
associated Ancestor.

Ancestors refer to objects that are either the Target of a policy or above it
in terms of object hierarchy. For example, if a policy targets a Service, the
Policy's Ancestors are, in order, the Service, the HTTPRoute, the Gateway, and
the GatewayClass. Almost always, in this hierarchy, the Gateway will be the most
useful object to place Policy status on, so we recommend that implementations
SHOULD use Gateway as the PolicyAncestorStatus object unless the designers
have a _very_ good reason otherwise.

In the context of policy attachment, the Ancestor is used to distinguish which
resource results in a distinct application of this policy. For example, if a policy
targets a Service, it may have a distinct result per attached Gateway.

Policies targeting the same resource may have different effects depending on the
ancestors of those resources. For example, different Gateways targeting the same
Service may have different capabilities, especially if they have different underlying
implementations.

For example, in BackendTLSPolicy, the Policy attaches to a Service that is
used as a backend in a HTTPRoute that is itself attached to a Gateway.
In this case, the relevant object for status is the Gateway, and that is the
ancestor object referred to in this status.

Note that a parent is also an ancestor, so for objects where the parent is the
relevant object for status, this struct SHOULD still be used.

This struct is intended to be used in a slice that's effectively a map,
with a composite key made up of the AncestorRef and the ControllerName.



_Appears in:_
- [PolicyStatus](#policystatus)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `ancestorRef` _[ParentReference](#parentreference)_ | AncestorRef corresponds with a ParentRef in the spec that this<br />PolicyAncestorStatus struct describes the status of. |  |  |
| `controllerName` _[GatewayController](#gatewaycontroller)_ | ControllerName is a domain/path string that indicates the name of the<br />controller that wrote this status. This corresponds with the<br />controllerName field on GatewayClass.<br />Example: "example.net/gateway-controller".<br />The format of this field is DOMAIN "/" PATH, where DOMAIN and PATH are<br />valid Kubernetes names<br />(https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#names).<br />Controllers MUST populate this field when writing status. Controllers should ensure that<br />entries to status populated with their ControllerName are cleaned up when they are no<br />longer necessary. |  | MaxLength: 253 <br />MinLength: 1 <br />Pattern: `^[a-z0-9]([-a-z0-9]*[a-z0-9])?(\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*\/[A-Za-z0-9\/\-._~%!$&'()*+,;=:]+$` <br /> |
| `conditions` _[Condition](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.32/#condition-v1-meta) array_ | Conditions describes the status of the Policy with respect to the given Ancestor. |  | MaxItems: 8 <br />MinItems: 1 <br /> |






#### PolicyStatus



PolicyStatus defines the common attributes that all Policies should include within
their status.



_Appears in:_
- [BackendTLSPolicy](#backendtlspolicy)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `ancestors` _[PolicyAncestorStatus](#policyancestorstatus) array_ | Ancestors is a list of ancestor resources (usually Gateways) that are<br />associated with the policy, and the status of the policy with respect to<br />each ancestor. When this policy attaches to a parent, the controller that<br />manages the parent and the ancestors MUST add an entry to this list when<br />the controller first sees the policy and SHOULD update the entry as<br />appropriate when the relevant ancestor is modified.<br />Note that choosing the relevant ancestor is left to the Policy designers;<br />an important part of Policy design is designing the right object level at<br />which to namespace this status.<br />Note also that implementations MUST ONLY populate ancestor status for<br />the Ancestor resources they are responsible for. Implementations MUST<br />use the ControllerName field to uniquely identify the entries in this list<br />that they are responsible for.<br />Note that to achieve this, the list of PolicyAncestorStatus structs<br />MUST be treated as a map with a composite key, made up of the AncestorRef<br />and ControllerName fields combined.<br />A maximum of 16 ancestors will be represented in this list. An empty list<br />means the Policy is not relevant for any ancestors.<br />If this slice is full, implementations MUST NOT add further entries.<br />Instead they MUST consider the policy unimplementable and signal that<br />on any related resources such as the ancestor that would be referenced<br />here. For example, if this list was full on BackendTLSPolicy, no<br />additional Gateways would be able to reference the Service targeted by<br />the BackendTLSPolicy. |  | MaxItems: 16 <br /> |






#### ReferenceGrant

_Underlying type:_ _[ReferenceGrant](#referencegrant)_

ReferenceGrant identifies kinds of resources in other namespaces that are
trusted to reference the specified kinds of resources in the same namespace
as the policy.

Each ReferenceGrant can be used to represent a unique trust relationship.
Additional Reference Grants can be used to add to the set of trusted
sources of inbound references for the namespace they are defined within.

A ReferenceGrant is required for all cross-namespace references in Gateway API
(with the exception of cross-namespace Route-Gateway attachment, which is
governed by the AllowedRoutes configuration on the Gateway, and cross-namespace
Service ParentRefs on a "consumer" mesh Route, which defines routing rules
applicable only to workloads in the Route namespace). ReferenceGrants allowing
a reference from a Route to a Service are only applicable to BackendRefs.

ReferenceGrant is a form of runtime verification allowing users to assert
which cross-namespace object references are permitted. Implementations that
support ReferenceGrant MUST NOT permit cross-namespace references which have
no grant, and MUST respond to the removal of a grant by revoking the access
that the grant allowed.





| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `apiVersion` _string_ | `gateway.networking.k8s.io/v1alpha2` | | |
| `kind` _string_ | `ReferenceGrant` | | |
| `metadata` _[ObjectMeta](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.32/#objectmeta-v1-meta)_ | Refer to Kubernetes API documentation for fields of `metadata`. |  |  |
| `spec` _[ReferenceGrantSpec](#referencegrantspec)_ | Spec defines the desired state of ReferenceGrant. |  |  |






















#### TCPRoute



TCPRoute provides a way to route TCP requests. When combined with a Gateway
listener, it can be used to forward connections on the port specified by the
listener to a set of backends specified by the TCPRoute.





| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `apiVersion` _string_ | `gateway.networking.k8s.io/v1alpha2` | | |
| `kind` _string_ | `TCPRoute` | | |
| `metadata` _[ObjectMeta](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.32/#objectmeta-v1-meta)_ | Refer to Kubernetes API documentation for fields of `metadata`. |  |  |
| `spec` _[TCPRouteSpec](#tcproutespec)_ | Spec defines the desired state of TCPRoute. |  |  |
| `status` _[TCPRouteStatus](#tcproutestatus)_ | Status defines the current state of TCPRoute. |  |  |


#### TCPRouteRule



TCPRouteRule is the configuration for a given rule.



_Appears in:_
- [TCPRouteSpec](#tcproutespec)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `name` _[SectionName](#sectionname)_ | Name is the name of the route rule. This name MUST be unique within a Route if it is set.<br />Support: Extended |  | MaxLength: 253 <br />MinLength: 1 <br />Pattern: `^[a-z0-9]([-a-z0-9]*[a-z0-9])?(\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*$` <br /> |
| `backendRefs` _[BackendRef](#backendref) array_ | BackendRefs defines the backend(s) where matching requests should be<br />sent. If unspecified or invalid (refers to a nonexistent resource or a<br />Service with no endpoints), the underlying implementation MUST actively<br />reject connection attempts to this backend. Connection rejections must<br />respect weight; if an invalid backend is requested to have 80% of<br />connections, then 80% of connections must be rejected instead.<br />Support: Core for Kubernetes Service<br />Support: Extended for Kubernetes ServiceImport<br />Support: Implementation-specific for any other resource<br />Support for weight: Extended |  | MaxItems: 16 <br />MinItems: 1 <br /> |


#### TCPRouteSpec



TCPRouteSpec defines the desired state of TCPRoute



_Appears in:_
- [TCPRoute](#tcproute)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `rules` _[TCPRouteRule](#tcprouterule) array_ | Rules are a list of TCP matchers and actions.<br /><gateway:experimental:validation:XValidation:message="Rule name must be unique within the route",rule="self.all(l1, !has(l1.name) \|\| self.exists_one(l2, has(l2.name) && l1.name == l2.name))"> |  | MaxItems: 16 <br />MinItems: 1 <br /> |


#### TCPRouteStatus



TCPRouteStatus defines the observed state of TCPRoute



_Appears in:_
- [TCPRoute](#tcproute)



#### TLSRoute



The TLSRoute resource is similar to TCPRoute, but can be configured
to match against TLS-specific metadata. This allows more flexibility
in matching streams for a given TLS listener.

If you need to forward traffic to a single target for a TLS listener, you
could choose to use a TCPRoute with a TLS listener.





| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `apiVersion` _string_ | `gateway.networking.k8s.io/v1alpha2` | | |
| `kind` _string_ | `TLSRoute` | | |
| `metadata` _[ObjectMeta](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.32/#objectmeta-v1-meta)_ | Refer to Kubernetes API documentation for fields of `metadata`. |  |  |
| `spec` _[TLSRouteSpec](#tlsroutespec)_ | Spec defines the desired state of TLSRoute. |  |  |
| `status` _[TLSRouteStatus](#tlsroutestatus)_ | Status defines the current state of TLSRoute. |  |  |


#### TLSRouteRule



TLSRouteRule is the configuration for a given rule.



_Appears in:_
- [TLSRouteSpec](#tlsroutespec)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `name` _[SectionName](#sectionname)_ | Name is the name of the route rule. This name MUST be unique within a Route if it is set.<br />Support: Extended |  | MaxLength: 253 <br />MinLength: 1 <br />Pattern: `^[a-z0-9]([-a-z0-9]*[a-z0-9])?(\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*$` <br /> |
| `backendRefs` _[BackendRef](#backendref) array_ | BackendRefs defines the backend(s) where matching requests should be<br />sent. If unspecified or invalid (refers to a nonexistent resource or<br />a Service with no endpoints), the rule performs no forwarding; if no<br />filters are specified that would result in a response being sent, the<br />underlying implementation must actively reject request attempts to this<br />backend, by rejecting the connection or returning a 500 status code.<br />Request rejections must respect weight; if an invalid backend is<br />requested to have 80% of requests, then 80% of requests must be rejected<br />instead.<br />Support: Core for Kubernetes Service<br />Support: Extended for Kubernetes ServiceImport<br />Support: Implementation-specific for any other resource<br />Support for weight: Extended |  | MaxItems: 16 <br />MinItems: 1 <br /> |


#### TLSRouteSpec



TLSRouteSpec defines the desired state of a TLSRoute resource.



_Appears in:_
- [TLSRoute](#tlsroute)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `hostnames` _[Hostname](#hostname) array_ | Hostnames defines a set of SNI names that should match against the<br />SNI attribute of TLS ClientHello message in TLS handshake. This matches<br />the RFC 1123 definition of a hostname with 2 notable exceptions:<br />1. IPs are not allowed in SNI names per RFC 6066.<br />2. A hostname may be prefixed with a wildcard label (`*.`). The wildcard<br />   label must appear by itself as the first label.<br />If a hostname is specified by both the Listener and TLSRoute, there<br />must be at least one intersecting hostname for the TLSRoute to be<br />attached to the Listener. For example:<br />* A Listener with `test.example.com` as the hostname matches TLSRoutes<br />  that have either not specified any hostnames, or have specified at<br />  least one of `test.example.com` or `*.example.com`.<br />* A Listener with `*.example.com` as the hostname matches TLSRoutes<br />  that have either not specified any hostnames or have specified at least<br />  one hostname that matches the Listener hostname. For example,<br />  `test.example.com` and `*.example.com` would both match. On the other<br />  hand, `example.com` and `test.example.net` would not match.<br />If both the Listener and TLSRoute have specified hostnames, any<br />TLSRoute hostnames that do not match the Listener hostname MUST be<br />ignored. For example, if a Listener specified `*.example.com`, and the<br />TLSRoute specified `test.example.com` and `test.example.net`,<br />`test.example.net` must not be considered for a match.<br />If both the Listener and TLSRoute have specified hostnames, and none<br />match with the criteria above, then the TLSRoute is not accepted. The<br />implementation must raise an 'Accepted' Condition with a status of<br />`False` in the corresponding RouteParentStatus.<br />Support: Core |  | MaxItems: 16 <br />MaxLength: 253 <br />MinLength: 1 <br />Pattern: `^(\*\.)?[a-z0-9]([-a-z0-9]*[a-z0-9])?(\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*$` <br /> |
| `rules` _[TLSRouteRule](#tlsrouterule) array_ | Rules are a list of TLS matchers and actions.<br /><gateway:experimental:validation:XValidation:message="Rule name must be unique within the route",rule="self.all(l1, !has(l1.name) \|\| self.exists_one(l2, has(l2.name) && l1.name == l2.name))"> |  | MaxItems: 16 <br />MinItems: 1 <br /> |


#### TLSRouteStatus



TLSRouteStatus defines the observed state of TLSRoute



_Appears in:_
- [TLSRoute](#tlsroute)



#### UDPRoute



UDPRoute provides a way to route UDP traffic. When combined with a Gateway
listener, it can be used to forward traffic on the port specified by the
listener to a set of backends specified by the UDPRoute.





| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `apiVersion` _string_ | `gateway.networking.k8s.io/v1alpha2` | | |
| `kind` _string_ | `UDPRoute` | | |
| `metadata` _[ObjectMeta](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.32/#objectmeta-v1-meta)_ | Refer to Kubernetes API documentation for fields of `metadata`. |  |  |
| `spec` _[UDPRouteSpec](#udproutespec)_ | Spec defines the desired state of UDPRoute. |  |  |
| `status` _[UDPRouteStatus](#udproutestatus)_ | Status defines the current state of UDPRoute. |  |  |


#### UDPRouteRule



UDPRouteRule is the configuration for a given rule.



_Appears in:_
- [UDPRouteSpec](#udproutespec)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `name` _[SectionName](#sectionname)_ | Name is the name of the route rule. This name MUST be unique within a Route if it is set.<br />Support: Extended |  | MaxLength: 253 <br />MinLength: 1 <br />Pattern: `^[a-z0-9]([-a-z0-9]*[a-z0-9])?(\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*$` <br /> |
| `backendRefs` _[BackendRef](#backendref) array_ | BackendRefs defines the backend(s) where matching requests should be<br />sent. If unspecified or invalid (refers to a nonexistent resource or a<br />Service with no endpoints), the underlying implementation MUST actively<br />reject connection attempts to this backend. Packet drops must<br />respect weight; if an invalid backend is requested to have 80% of<br />the packets, then 80% of packets must be dropped instead.<br />Support: Core for Kubernetes Service<br />Support: Extended for Kubernetes ServiceImport<br />Support: Implementation-specific for any other resource<br />Support for weight: Extended |  | MaxItems: 16 <br />MinItems: 1 <br /> |


#### UDPRouteSpec



UDPRouteSpec defines the desired state of UDPRoute.



_Appears in:_
- [UDPRoute](#udproute)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `rules` _[UDPRouteRule](#udprouterule) array_ | Rules are a list of UDP matchers and actions.<br /><gateway:experimental:validation:XValidation:message="Rule name must be unique within the route",rule="self.all(l1, !has(l1.name) \|\| self.exists_one(l2, has(l2.name) && l1.name == l2.name))"> |  | MaxItems: 16 <br />MinItems: 1 <br /> |


#### UDPRouteStatus



UDPRouteStatus defines the observed state of UDPRoute.



_Appears in:_
- [UDPRoute](#udproute)




## gateway.networking.k8s.io/v1alpha3

Package v1alpha3 contains API Schema definitions for the
gateway.networking.k8s.io API group.


### Resource Types
- [BackendTLSPolicy](#backendtlspolicy)



#### BackendTLSPolicy



BackendTLSPolicy provides a way to configure how a Gateway
connects to a Backend via TLS.





| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `apiVersion` _string_ | `gateway.networking.k8s.io/v1alpha3` | | |
| `kind` _string_ | `BackendTLSPolicy` | | |
| `metadata` _[ObjectMeta](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.32/#objectmeta-v1-meta)_ | Refer to Kubernetes API documentation for fields of `metadata`. |  |  |
| `spec` _[BackendTLSPolicySpec](#backendtlspolicyspec)_ | Spec defines the desired state of BackendTLSPolicy. |  |  |
| `status` _[PolicyStatus](#policystatus)_ | Status defines the current state of BackendTLSPolicy. |  |  |


#### BackendTLSPolicySpec



BackendTLSPolicySpec defines the desired state of BackendTLSPolicy.

Support: Extended



_Appears in:_
- [BackendTLSPolicy](#backendtlspolicy)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `targetRefs` _[LocalPolicyTargetReferenceWithSectionName](#localpolicytargetreferencewithsectionname) array_ | TargetRefs identifies an API object to apply the policy to.<br />Only Services have Extended support. Implementations MAY support<br />additional objects, with Implementation Specific support.<br />Note that this config applies to the entire referenced resource<br />by default, but this default may change in the future to provide<br />a more granular application of the policy.<br />TargetRefs must be _distinct_. This means either that:<br />* They select different targets. If this is the case, then targetRef<br />  entries are distinct. In terms of fields, this means that the<br />  multi-part key defined by `group`, `kind`, and `name` must<br />  be unique across all targetRef entries in the BackendTLSPolicy.<br />* They select different sectionNames in the same target.<br />Support: Extended for Kubernetes Service<br />Support: Implementation-specific for any other resource |  | MaxItems: 16 <br />MinItems: 1 <br /> |
| `validation` _[BackendTLSPolicyValidation](#backendtlspolicyvalidation)_ | Validation contains backend TLS validation configuration. |  |  |
| `options` _object (keys:[AnnotationKey](#annotationkey), values:[AnnotationValue](#annotationvalue))_ | Options are a list of key/value pairs to enable extended TLS<br />configuration for each implementation. For example, configuring the<br />minimum TLS version or supported cipher suites.<br />A set of common keys MAY be defined by the API in the future. To avoid<br />any ambiguity, implementation-specific definitions MUST use<br />domain-prefixed names, such as `example.com/my-custom-option`.<br />Un-prefixed names are reserved for key names defined by Gateway API.<br />Support: Implementation-specific |  | MaxProperties: 16 <br /> |


#### BackendTLSPolicyValidation



BackendTLSPolicyValidation contains backend TLS validation configuration.



_Appears in:_
- [BackendTLSPolicySpec](#backendtlspolicyspec)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `caCertificateRefs` _[LocalObjectReference](#localobjectreference) array_ | CACertificateRefs contains one or more references to Kubernetes objects that<br />contain a PEM-encoded TLS CA certificate bundle, which is used to<br />validate a TLS handshake between the Gateway and backend Pod.<br />If CACertificateRefs is empty or unspecified, then WellKnownCACertificates must be<br />specified. Only one of CACertificateRefs or WellKnownCACertificates may be specified,<br />not both. If CACertificateRefs is empty or unspecified, the configuration for<br />WellKnownCACertificates MUST be honored instead if supported by the implementation.<br />References to a resource in a different namespace are invalid for the<br />moment, although we will revisit this in the future.<br />A single CACertificateRef to a Kubernetes ConfigMap kind has "Core" support.<br />Implementations MAY choose to support attaching multiple certificates to<br />a backend, but this behavior is implementation-specific.<br />Support: Core - An optional single reference to a Kubernetes ConfigMap,<br />with the CA certificate in a key named `ca.crt`.<br />Support: Implementation-specific (More than one reference, or other kinds<br />of resources). |  | MaxItems: 8 <br /> |
| `wellKnownCACertificates` _[WellKnownCACertificatesType](#wellknowncacertificatestype)_ | WellKnownCACertificates specifies whether system CA certificates may be used in<br />the TLS handshake between the gateway and backend pod.<br />If WellKnownCACertificates is unspecified or empty (""), then CACertificateRefs<br />must be specified with at least one entry for a valid configuration. Only one of<br />CACertificateRefs or WellKnownCACertificates may be specified, not both. If an<br />implementation does not support the WellKnownCACertificates field or the value<br />supplied is not supported, the Status Conditions on the Policy MUST be<br />updated to include an Accepted: False Condition with Reason: Invalid.<br />Support: Implementation-specific |  | Enum: [System] <br /> |
| `hostname` _[PreciseHostname](#precisehostname)_ | Hostname is used for two purposes in the connection between Gateways and<br />backends:<br />1. Hostname MUST be used as the SNI to connect to the backend (RFC 6066).<br />2. Hostname MUST be used for authentication and MUST match the certificate served by the matching backend, unless SubjectAltNames is specified.<br />   authentication and MUST match the certificate served by the matching<br />   backend.<br />Support: Core |  | MaxLength: 253 <br />MinLength: 1 <br />Pattern: `^[a-z0-9]([-a-z0-9]*[a-z0-9])?(\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*$` <br /> |
| `subjectAltNames` _[SubjectAltName](#subjectaltname) array_ | SubjectAltNames contains one or more Subject Alternative Names.<br />When specified the certificate served from the backend MUST<br />have at least one Subject Alternate Name matching one of the specified SubjectAltNames.<br />Support: Extended |  | MaxItems: 5 <br /> |


#### SubjectAltName



SubjectAltName represents Subject Alternative Name.



_Appears in:_
- [BackendTLSPolicyValidation](#backendtlspolicyvalidation)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `type` _[SubjectAltNameType](#subjectaltnametype)_ | Type determines the format of the Subject Alternative Name. Always required.<br />Support: Core |  | Enum: [Hostname URI] <br /> |
| `hostname` _[Hostname](#hostname)_ | Hostname contains Subject Alternative Name specified in DNS name format.<br />Required when Type is set to Hostname, ignored otherwise.<br />Support: Core |  | MaxLength: 253 <br />MinLength: 1 <br />Pattern: `^(\*\.)?[a-z0-9]([-a-z0-9]*[a-z0-9])?(\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*$` <br /> |
| `uri` _[AbsoluteURI](#absoluteuri)_ | URI contains Subject Alternative Name specified in a full URI format.<br />It MUST include both a scheme (e.g., "http" or "ftp") and a scheme-specific-part.<br />Common values include SPIFFE IDs like "spiffe://mycluster.example.com/ns/myns/sa/svc1sa".<br />Required when Type is set to URI, ignored otherwise.<br />Support: Core |  | MaxLength: 253 <br />MinLength: 1 <br />Pattern: `^(([^:/?#]+):)(//([^/?#]*))([^?#]*)(\?([^#]*))?(#(.*))?` <br /> |


#### SubjectAltNameType

_Underlying type:_ _string_

SubjectAltNameType is the type of the Subject Alternative Name.

_Validation:_
- Enum: [Hostname URI]

_Appears in:_
- [SubjectAltName](#subjectaltname)

| Field | Description |
| --- | --- |
| `Hostname` | HostnameSubjectAltNameType specifies hostname-based SAN.<br />Support: Core<br /> |
| `URI` | URISubjectAltNameType specifies URI-based SAN, e.g. SPIFFE id.<br />Support: Core<br /> |


#### WellKnownCACertificatesType

_Underlying type:_ _string_

WellKnownCACertificatesType is the type of CA certificate that will be used
when the caCertificateRefs field is unspecified.

_Validation:_
- Enum: [System]

_Appears in:_
- [BackendTLSPolicyValidation](#backendtlspolicyvalidation)

| Field | Description |
| --- | --- |
| `System` | WellKnownCACertificatesSystem indicates that well known system CA certificates should be used.<br /> |



## gateway.networking.k8s.io/v1beta1

Package v1beta1 contains API Schema definitions for the
gateway.networking.k8s.io API group.


### Resource Types
- [Gateway](#gateway)
- [GatewayClass](#gatewayclass)
- [HTTPRoute](#httproute)
- [ReferenceGrant](#referencegrant)





















#### Gateway

_Underlying type:_ _[Gateway](#gateway)_

Gateway represents an instance of a service-traffic handling infrastructure
by binding Listeners to a set of IP addresses.





| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `apiVersion` _string_ | `gateway.networking.k8s.io/v1beta1` | | |
| `kind` _string_ | `Gateway` | | |
| `metadata` _[ObjectMeta](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.32/#objectmeta-v1-meta)_ | Refer to Kubernetes API documentation for fields of `metadata`. |  |  |
| `spec` _[GatewaySpec](#gatewayspec)_ | Spec defines the desired state of Gateway. |  |  |
| `status` _[GatewayStatus](#gatewaystatus)_ | Status defines the current state of Gateway. | \{ conditions:[map[lastTransitionTime:1970-01-01T00:00:00Z message:Waiting for controller reason:Pending status:Unknown type:Accepted] map[lastTransitionTime:1970-01-01T00:00:00Z message:Waiting for controller reason:Pending status:Unknown type:Programmed]] \} |  |


#### GatewayClass

_Underlying type:_ _[GatewayClass](#gatewayclass)_

GatewayClass describes a class of Gateways available to the user for creating
Gateway resources.

It is recommended that this resource be used as a template for Gateways. This
means that a Gateway is based on the state of the GatewayClass at the time it
was created and changes to the GatewayClass or associated parameters are not
propagated down to existing Gateways. This recommendation is intended to
limit the blast radius of changes to GatewayClass or associated parameters.
If implementations choose to propagate GatewayClass changes to existing
Gateways, that MUST be clearly documented by the implementation.

Whenever one or more Gateways are using a GatewayClass, implementations SHOULD
add the `gateway-exists-finalizer.gateway.networking.k8s.io` finalizer on the
associated GatewayClass. This ensures that a GatewayClass associated with a
Gateway is not deleted while in use.

GatewayClass is a Cluster level resource.





| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `apiVersion` _string_ | `gateway.networking.k8s.io/v1beta1` | | |
| `kind` _string_ | `GatewayClass` | | |
| `metadata` _[ObjectMeta](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.32/#objectmeta-v1-meta)_ | Refer to Kubernetes API documentation for fields of `metadata`. |  |  |
| `spec` _[GatewayClassSpec](#gatewayclassspec)_ | Spec defines the desired state of GatewayClass. |  |  |
| `status` _[GatewayClassStatus](#gatewayclassstatus)_ | Status defines the current state of GatewayClass.<br />Implementations MUST populate status on all GatewayClass resources which<br />specify their controller name. | \{ conditions:[map[lastTransitionTime:1970-01-01T00:00:00Z message:Waiting for controller reason:Pending status:Unknown type:Accepted]] \} |  |


















































#### HTTPRoute

_Underlying type:_ _[HTTPRoute](#httproute)_

HTTPRoute provides a way to route HTTP requests. This includes the capability
to match requests by hostname, path, header, or query param. Filters can be
used to specify additional processing steps. Backends specify where matching
requests should be routed.





| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `apiVersion` _string_ | `gateway.networking.k8s.io/v1beta1` | | |
| `kind` _string_ | `HTTPRoute` | | |
| `metadata` _[ObjectMeta](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.32/#objectmeta-v1-meta)_ | Refer to Kubernetes API documentation for fields of `metadata`. |  |  |
| `spec` _[HTTPRouteSpec](#httproutespec)_ | Spec defines the desired state of HTTPRoute. |  |  |
| `status` _[HTTPRouteStatus](#httproutestatus)_ | Status defines the current state of HTTPRoute. |  |  |




















































#### ReferenceGrant



ReferenceGrant identifies kinds of resources in other namespaces that are
trusted to reference the specified kinds of resources in the same namespace
as the policy.

Each ReferenceGrant can be used to represent a unique trust relationship.
Additional Reference Grants can be used to add to the set of trusted
sources of inbound references for the namespace they are defined within.

All cross-namespace references in Gateway API (with the exception of cross-namespace
Gateway-route attachment) require a ReferenceGrant.

ReferenceGrant is a form of runtime verification allowing users to assert
which cross-namespace object references are permitted. Implementations that
support ReferenceGrant MUST NOT permit cross-namespace references which have
no grant, and MUST respond to the removal of a grant by revoking the access
that the grant allowed.



_Appears in:_
- [ReferenceGrant](#referencegrant)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `apiVersion` _string_ | `gateway.networking.k8s.io/v1beta1` | | |
| `kind` _string_ | `ReferenceGrant` | | |
| `metadata` _[ObjectMeta](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.32/#objectmeta-v1-meta)_ | Refer to Kubernetes API documentation for fields of `metadata`. |  |  |
| `spec` _[ReferenceGrantSpec](#referencegrantspec)_ | Spec defines the desired state of ReferenceGrant. |  |  |


#### ReferenceGrantFrom



ReferenceGrantFrom describes trusted namespaces and kinds.



_Appears in:_
- [ReferenceGrantSpec](#referencegrantspec)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `group` _[Group](#group)_ | Group is the group of the referent.<br />When empty, the Kubernetes core API group is inferred.<br />Support: Core |  | MaxLength: 253 <br />Pattern: `^$\|^[a-z0-9]([-a-z0-9]*[a-z0-9])?(\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*$` <br /> |
| `kind` _[Kind](#kind)_ | Kind is the kind of the referent. Although implementations may support<br />additional resources, the following types are part of the "Core"<br />support level for this field.<br />When used to permit a SecretObjectReference:<br />* Gateway<br />When used to permit a BackendObjectReference:<br />* GRPCRoute<br />* HTTPRoute<br />* TCPRoute<br />* TLSRoute<br />* UDPRoute |  | MaxLength: 63 <br />MinLength: 1 <br />Pattern: `^[a-zA-Z]([-a-zA-Z0-9]*[a-zA-Z0-9])?$` <br /> |
| `namespace` _[Namespace](#namespace)_ | Namespace is the namespace of the referent.<br />Support: Core |  | MaxLength: 63 <br />MinLength: 1 <br />Pattern: `^[a-z0-9]([-a-z0-9]*[a-z0-9])?$` <br /> |


#### ReferenceGrantSpec



ReferenceGrantSpec identifies a cross namespace relationship that is trusted
for Gateway API.



_Appears in:_
- [ReferenceGrant](#referencegrant)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `from` _[ReferenceGrantFrom](#referencegrantfrom) array_ | From describes the trusted namespaces and kinds that can reference the<br />resources described in "To". Each entry in this list MUST be considered<br />to be an additional place that references can be valid from, or to put<br />this another way, entries MUST be combined using OR.<br />Support: Core |  | MaxItems: 16 <br />MinItems: 1 <br /> |
| `to` _[ReferenceGrantTo](#referencegrantto) array_ | To describes the resources that may be referenced by the resources<br />described in "From". Each entry in this list MUST be considered to be an<br />additional place that references can be valid to, or to put this another<br />way, entries MUST be combined using OR.<br />Support: Core |  | MaxItems: 16 <br />MinItems: 1 <br /> |


#### ReferenceGrantTo



ReferenceGrantTo describes what Kinds are allowed as targets of the
references.



_Appears in:_
- [ReferenceGrantSpec](#referencegrantspec)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `group` _[Group](#group)_ | Group is the group of the referent.<br />When empty, the Kubernetes core API group is inferred.<br />Support: Core |  | MaxLength: 253 <br />Pattern: `^$\|^[a-z0-9]([-a-z0-9]*[a-z0-9])?(\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*$` <br /> |
| `kind` _[Kind](#kind)_ | Kind is the kind of the referent. Although implementations may support<br />additional resources, the following types are part of the "Core"<br />support level for this field:<br />* Secret when used to permit a SecretObjectReference<br />* Service when used to permit a BackendObjectReference |  | MaxLength: 63 <br />MinLength: 1 <br />Pattern: `^[a-zA-Z]([-a-zA-Z0-9]*[a-zA-Z0-9])?$` <br /> |
| `name` _[ObjectName](#objectname)_ | Name is the name of the referent. When unspecified, this policy<br />refers to all resources of the specified Group and Kind in the local<br />namespace. |  | MaxLength: 253 <br />MinLength: 1 <br /> |




















