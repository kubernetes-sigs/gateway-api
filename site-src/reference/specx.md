# API Reference

## Packages
- [gateway.networking.x-k8s.io/v1alpha1](#gatewaynetworkingx-k8siov1alpha1)


## gateway.networking.x-k8s.io/v1alpha1

Package v1alpha1 contains API Schema definitions for the gateway.networking.k8s-x.io
API group.


### Resource Types
- [XBackendTrafficPolicy](#xbackendtrafficpolicy)
- [XListenerSet](#xlistenerset)



#### AllowedRoutes







_Appears in:_
- [ListenerEntry](#listenerentry)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `namespaces` _[RouteNamespaces](#routenamespaces)_ | Namespaces indicates namespaces from which Routes may be attached to this<br />Listener. This is restricted to the namespace of this Gateway by default.<br /><br />Support: Core | \{ from:Same \} |  |
| `kinds` _[RouteGroupKind](#routegroupkind) array_ | Kinds specifies the groups and kinds of Routes that are allowed to bind<br />to this Gateway Listener. When unspecified or empty, the kinds of Routes<br />selected are determined using the Listener protocol.<br /><br />A RouteGroupKind MUST correspond to kinds of Routes that are compatible<br />with the application protocol specified in the Listener's Protocol field.<br />If an implementation does not support or recognize this resource type, it<br />MUST set the "ResolvedRefs" condition to False for this Listener with the<br />"InvalidRouteKinds" reason.<br /><br />Support: Core |  | MaxItems: 8 <br /> |


#### BackendTrafficPolicySpec



BackendTrafficPolicySpec define the desired state of BackendTrafficPolicy
Note: there is no Override or Default policy configuration.



_Appears in:_
- [XBackendTrafficPolicy](#xbackendtrafficpolicy)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `targetRefs` _LocalPolicyTargetReference array_ | TargetRefs identifies API object(s) to apply this policy to.<br />Currently, Backends (A grouping of like endpoints such as Service,<br />ServiceImport, or any implementation-specific backendRef) are the only<br />valid API target references.<br /><br />Currently, a TargetRef can not be scoped to a specific port on a<br />Service. |  | MaxItems: 16 <br />MinItems: 1 <br /> |
| `retryConstraint` _[RetryConstraint](#retryconstraint)_ | RetryConstraint defines the configuration for when to allow or prevent<br />further retries to a target backend, by dynamically calculating a 'retry<br />budget'. This budget is calculated based on the percentage of incoming<br />traffic composed of retries over a given time interval. Once the budget<br />is exceeded, additional retries will be rejected.<br /><br />For example, if the retry budget interval is 10 seconds, there have been<br />1000 active requests in the past 10 seconds, and the allowed percentage<br />of requests that can be retried is 20% (the default), then 200 of those<br />requests may be composed of retries. Active requests will only be<br />considered for the duration of the interval when calculating the retry<br />budget. Retrying the same original request multiple times within the<br />retry budget interval will lead to each retry being counted towards<br />calculating the budget.<br /><br />Configuring a RetryConstraint in BackendTrafficPolicy is compatible with<br />HTTPRoute Retry settings for each HTTPRouteRule that targets the same<br />backend. While the HTTPRouteRule Retry stanza can specify whether a<br />request will be retried, and the number of retry attempts each client<br />may perform, RetryConstraint helps prevent cascading failures such as<br />retry storms during periods of consistent failures.<br /><br />After the retry budget has been exceeded, additional retries to the<br />backend MUST return a 503 response to the client.<br /><br />Additional configurations for defining a constraint on retries MAY be<br />defined in the future.<br /><br />Support: Extended<br /><br /><gateway:experimental> |  |  |
| `sessionPersistence` _[SessionPersistence](#sessionpersistence)_ | SessionPersistence defines and configures session persistence<br />for the backend.<br /><br />Support: Extended |  |  |


#### BudgetDetails



BudgetDetails specifies the details of the budget configuration, like
the percentage of requests in the budget, and the interval between
checks.



_Appears in:_
- [RetryConstraint](#retryconstraint)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `percent` _integer_ | BudgetPercent defines the maximum percentage of active requests that may<br />be made up of retries.<br /><br />Support: Extended | 20 | Maximum: 100 <br />Minimum: 0 <br /> |
| `interval` _[Duration](#duration)_ | BudgetInterval defines the duration in which requests will be considered<br />for calculating the budget for retries.<br /><br />Support: Extended | 10s | Pattern: `^([0-9]\{1,5\}(h\|m\|s\|ms))\{1,4\}$` <br /> |


#### Duration

_Underlying type:_ _string_





_Appears in:_
- [BudgetDetails](#budgetdetails)
- [RequestRate](#requestrate)
- SessionPersistence



#### GatewayTLSConfig







_Appears in:_
- [ListenerEntry](#listenerentry)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `mode` _[TLSModeType](#tlsmodetype)_ | Mode defines the TLS behavior for the TLS session initiated by the client.<br />There are two possible modes:<br /><br />- Terminate: The TLS session between the downstream client and the<br />  Gateway is terminated at the Gateway. This mode requires certificates<br />  to be specified in some way, such as populating the certificateRefs<br />  field.<br />- Passthrough: The TLS session is NOT terminated by the Gateway. This<br />  implies that the Gateway can't decipher the TLS stream except for<br />  the ClientHello message of the TLS protocol. The certificateRefs field<br />  is ignored in this mode.<br /><br />Support: Core | Terminate | Enum: [Terminate Passthrough] <br /> |
| `certificateRefs` _[SecretObjectReference](#secretobjectreference) array_ | CertificateRefs contains a series of references to Kubernetes objects that<br />contains TLS certificates and private keys. These certificates are used to<br />establish a TLS handshake for requests that match the hostname of the<br />associated listener.<br /><br />A single CertificateRef to a Kubernetes Secret has "Core" support.<br />Implementations MAY choose to support attaching multiple certificates to<br />a Listener, but this behavior is implementation-specific.<br /><br />References to a resource in different namespace are invalid UNLESS there<br />is a ReferenceGrant in the target namespace that allows the certificate<br />to be attached. If a ReferenceGrant does not allow this reference, the<br />"ResolvedRefs" condition MUST be set to False for this listener with the<br />"RefNotPermitted" reason.<br /><br />This field is required to have at least one element when the mode is set<br />to "Terminate" (default) and is optional otherwise.<br /><br />CertificateRefs can reference to standard Kubernetes resources, i.e.<br />Secret, or implementation-specific custom resources.<br /><br />Support: Core - A single reference to a Kubernetes Secret of type kubernetes.io/tls<br /><br />Support: Implementation-specific (More than one reference or other resource types) |  | MaxItems: 64 <br /> |
| `frontendValidation` _[FrontendTLSValidation](#frontendtlsvalidation)_ | FrontendValidation holds configuration information for validating the frontend (client).<br />Setting this field will require clients to send a client certificate<br />required for validation during the TLS handshake. In browsers this may result in a dialog appearing<br />that requests a user to specify the client certificate.<br />The maximum depth of a certificate chain accepted in verification is Implementation specific.<br /><br />Support: Extended<br /><br /><gateway:experimental> |  |  |
| `options` _object (keys:[AnnotationKey](#annotationkey), values:[AnnotationValue](#annotationvalue))_ | Options are a list of key/value pairs to enable extended TLS<br />configuration for each implementation. For example, configuring the<br />minimum TLS version or supported cipher suites.<br /><br />A set of common keys MAY be defined by the API in the future. To avoid<br />any ambiguity, implementation-specific definitions MUST use<br />domain-prefixed names, such as `example.com/my-custom-option`.<br />Un-prefixed names are reserved for key names defined by Gateway API.<br /><br />Support: Implementation-specific |  | MaxProperties: 16 <br /> |


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
- LocalPolicyTargetReference
- [ObjectReference](#objectreference)
- [ParentGatewayReference](#parentgatewayreference)
- [ParentReference](#parentreference)
- [RouteGroupKind](#routegroupkind)
- [SecretObjectReference](#secretobjectreference)



#### Hostname

_Underlying type:_ _string_





_Appears in:_
- [ListenerEntry](#listenerentry)



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
- LocalPolicyTargetReference
- [ObjectReference](#objectreference)
- [ParentGatewayReference](#parentgatewayreference)
- [ParentReference](#parentreference)
- [RouteGroupKind](#routegroupkind)
- [SecretObjectReference](#secretobjectreference)



#### ListenerEntry







_Appears in:_
- [ListenerSetSpec](#listenersetspec)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `name` _[SectionName](#sectionname)_ | Name is the name of the Listener. This name MUST be unique within a<br />ListenerSet.<br /><br />Name is not required to be unique across a Gateway and ListenerSets.<br />Routes can attach to a Listener by having a ListenerSet as a parentRef<br />and setting the SectionName |  |  |
| `hostname` _[Hostname](#hostname)_ | Hostname specifies the virtual hostname to match for protocol types that<br />define this concept. When unspecified, all hostnames are matched. This<br />field is ignored for protocols that don't require hostname based<br />matching.<br /><br />Implementations MUST apply Hostname matching appropriately for each of<br />the following protocols:<br /><br />* TLS: The Listener Hostname MUST match the SNI.<br />* HTTP: The Listener Hostname MUST match the Host header of the request.<br />* HTTPS: The Listener Hostname SHOULD match at both the TLS and HTTP<br />  protocol layers as described above. If an implementation does not<br />  ensure that both the SNI and Host header match the Listener hostname,<br />  it MUST clearly document that.<br /><br />For HTTPRoute and TLSRoute resources, there is an interaction with the<br />`spec.hostnames` array. When both listener and route specify hostnames,<br />there MUST be an intersection between the values for a Route to be<br />accepted. For more information, refer to the Route specific Hostnames<br />documentation.<br /><br />Hostnames that are prefixed with a wildcard label (`*.`) are interpreted<br />as a suffix match. That means that a match for `*.example.com` would match<br />both `test.example.com`, and `foo.test.example.com`, but not `example.com`. |  |  |
| `port` _[PortNumber](#portnumber)_ | Port is the network port. Multiple listeners may use the<br />same port, subject to the Listener compatibility rules. |  |  |
| `protocol` _[ProtocolType](#protocoltype)_ | Protocol specifies the network protocol this listener expects to receive. |  |  |
| `tls` _[GatewayTLSConfig](#gatewaytlsconfig)_ | TLS is the TLS configuration for the Listener. This field is required if<br />the Protocol field is "HTTPS" or "TLS". It is invalid to set this field<br />if the Protocol field is "HTTP", "TCP", or "UDP".<br /><br />The association of SNIs to Certificate defined in GatewayTLSConfig is<br />defined based on the Hostname field for this listener.<br /><br />The GatewayClass MUST use the longest matching SNI out of all<br />available certificates for any TLS handshake. |  |  |
| `allowedRoutes` _[AllowedRoutes](#allowedroutes)_ | AllowedRoutes defines the types of routes that MAY be attached to a<br />Listener and the trusted namespaces where those Route resources MAY be<br />present.<br /><br />Although a client request may match multiple route rules, only one rule<br />may ultimately receive the request. Matching precedence MUST be<br />determined in order of the following criteria:<br /><br />* The most specific match as defined by the Route type.<br />* The oldest Route based on creation timestamp. For example, a Route with<br />  a creation timestamp of "2020-09-08 01:02:03" is given precedence over<br />  a Route with a creation timestamp of "2020-09-08 01:02:04".<br />* If everything else is equivalent, the Route appearing first in<br />  alphabetical order (namespace/name) should be given precedence. For<br />  example, foo/bar is given precedence over foo/baz.<br /><br />All valid rules within a Route attached to this Listener should be<br />implemented. Invalid Route rules can be ignored (sometimes that will mean<br />the full Route). If a Route rule transitions from valid to invalid,<br />support for that Route rule should be dropped to ensure consistency. For<br />example, even if a filter specified by a Route rule is invalid, the rest<br />of the rules within that Route should still be supported. | \{ namespaces:map[from:Same] \} |  |






#### ListenerEntryStatus



ListenerStatus is the status associated with a Listener.



_Appears in:_
- [ListenerSetStatus](#listenersetstatus)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `name` _[SectionName](#sectionname)_ | Name is the name of the Listener that this status corresponds to. |  |  |
| `port` _[PortNumber](#portnumber)_ | Port is the network port the listener is configured to listen on. |  |  |
| `supportedKinds` _[RouteGroupKind](#routegroupkind) array_ | SupportedKinds is the list indicating the Kinds supported by this<br />listener. This MUST represent the kinds an implementation supports for<br />that Listener configuration.<br /><br />If kinds are specified in Spec that are not supported, they MUST NOT<br />appear in this list and an implementation MUST set the "ResolvedRefs"<br />condition to "False" with the "InvalidRouteKinds" reason. If both valid<br />and invalid Route kinds are specified, the implementation MUST<br />reference the valid Route kinds that have been specified. |  | MaxItems: 8 <br /> |
| `attachedRoutes` _integer_ | AttachedRoutes represents the total number of Routes that have been<br />successfully attached to this Listener.<br /><br />Successful attachment of a Route to a Listener is based solely on the<br />combination of the AllowedRoutes field on the corresponding Listener<br />and the Route's ParentRefs field. A Route is successfully attached to<br />a Listener when it is selected by the Listener's AllowedRoutes field<br />AND the Route has a valid ParentRef selecting the whole Gateway<br />resource or a specific Listener as a parent resource (more detail on<br />attachment semantics can be found in the documentation on the various<br />Route kinds ParentRefs fields). Listener or Route status does not impact<br />successful attachment, i.e. the AttachedRoutes field count MUST be set<br />for Listeners with condition Accepted: false and MUST count successfully<br />attached Routes that may themselves have Accepted: false conditions.<br /><br />Uses for this field include troubleshooting Route attachment and<br />measuring blast radius/impact of changes to a Listener. |  |  |
| `conditions` _[Condition](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.32/#condition-v1-meta) array_ | Conditions describe the current condition of this listener. |  | MaxItems: 8 <br /> |






#### ListenerSetSpec



ListenerSetSpec defines the desired state of a ListenerSet.



_Appears in:_
- [XListenerSet](#xlistenerset)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `parentRef` _[ParentGatewayReference](#parentgatewayreference)_ | ParentRef references the Gateway that the listeners are attached to. |  |  |
| `listeners` _[ListenerEntry](#listenerentry) array_ | Listeners associated with this ListenerSet. Listeners define<br />logical endpoints that are bound on this referenced parent Gateway's addresses.<br /><br />Listeners in a `Gateway` and their attached `ListenerSets` are concatenated<br />as a list when programming the underlying infrastructure. Each listener<br />name does not need to be unique across the Gateway and ListenerSets.<br />See ListenerEntry.Name for more details.<br /><br />Implementations MUST treat the parent Gateway as having the merged<br />list of all listeners from itself and attached ListenerSets using<br />the following precedence:<br /><br />1. "parent" Gateway<br />2. ListenerSet ordered by creation time (oldest first)<br />3. ListenerSet ordered alphabetically by “\{namespace\}/\{name\}”.<br /><br />An implementation MAY reject listeners by setting the ListenerEntryStatus<br />`Accepted`` condition to False with the Reason `TooManyListeners`<br /><br />If a listener has a conflict, this will be reported in the<br />Status.ListenerEntryStatus setting the `Conflicted` condition to True.<br /><br />Implementations SHOULD be cautious about what information from the<br />parent or siblings are reported to avoid accidentally leaking<br />sensitive information that the child would not otherwise have access<br />to. This can include contents of secrets etc. |  | MaxItems: 64 <br />MinItems: 1 <br /> |


#### ListenerSetStatus







_Appears in:_
- [XListenerSet](#xlistenerset)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `conditions` _[Condition](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.32/#condition-v1-meta) array_ | Conditions describe the current conditions of the ListenerSet.<br /><br />Implementations MUST express ListenerSet conditions using the<br />`ListenerSetConditionType` and `ListenerSetConditionReason`<br />constants so that operators and tools can converge on a common<br />vocabulary to describe ListenerSet state.<br /><br />Known condition types are:<br /><br />* "Accepted"<br />* "Programmed" | [map[lastTransitionTime:1970-01-01T00:00:00Z message:Waiting for controller reason:Pending status:Unknown type:Accepted] map[lastTransitionTime:1970-01-01T00:00:00Z message:Waiting for controller reason:Pending status:Unknown type:Programmed]] | MaxItems: 8 <br /> |
| `listeners` _[ListenerEntryStatus](#listenerentrystatus) array_ | Listeners provide status for each unique listener port defined in the Spec. |  | MaxItems: 64 <br /> |


#### LocalPolicyTargetReference







_Appears in:_
- [BackendTrafficPolicySpec](#backendtrafficpolicyspec)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `group` _[Group](#group)_ | Group is the group of the target resource. |  | MaxLength: 253 <br />Pattern: `^$\|^[a-z0-9]([-a-z0-9]*[a-z0-9])?(\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*$` <br /> |
| `kind` _[Kind](#kind)_ | Kind is kind of the target resource. |  | MaxLength: 63 <br />MinLength: 1 <br />Pattern: `^[a-zA-Z]([-a-zA-Z0-9]*[a-zA-Z0-9])?$` <br /> |
| `name` _[ObjectName](#objectname)_ | Name is the name of the target resource. |  | MaxLength: 253 <br />MinLength: 1 <br /> |


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
- [ObjectReference](#objectreference)
- [ParentGatewayReference](#parentgatewayreference)
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
- LocalPolicyTargetReference
- [ObjectReference](#objectreference)
- [ParentGatewayReference](#parentgatewayreference)
- [ParentReference](#parentreference)
- [SecretObjectReference](#secretobjectreference)



#### ParentGatewayReference



ParentGatewayReference identifies an API object including its namespace,
defaulting to Gateway.



_Appears in:_
- [ListenerSetSpec](#listenersetspec)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `group` _[Group](#group)_ | Group is the group of the referent. | gateway.networking.k8s.io | MaxLength: 253 <br />Pattern: `^$\|^[a-z0-9]([-a-z0-9]*[a-z0-9])?(\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*$` <br /> |
| `kind` _[Kind](#kind)_ | Kind is kind of the referent. For example "Gateway". | Gateway | MaxLength: 63 <br />MinLength: 1 <br />Pattern: `^[a-zA-Z]([-a-zA-Z0-9]*[a-zA-Z0-9])?$` <br /> |
| `name` _[ObjectName](#objectname)_ | Name is the name of the referent. |  | MaxLength: 253 <br />MinLength: 1 <br /> |
| `namespace` _[Namespace](#namespace)_ | Namespace is the namespace of the referent.  If not present,<br />the namespace of the referent is assumed to be the same as<br />the namespace of the referring object. |  | MaxLength: 63 <br />MinLength: 1 <br />Pattern: `^[a-z0-9]([-a-z0-9]*[a-z0-9])?$` <br /> |


#### PolicyStatus







_Appears in:_
- [XBackendTrafficPolicy](#xbackendtrafficpolicy)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `ancestors` _[PolicyAncestorStatus](#policyancestorstatus) array_ | Ancestors is a list of ancestor resources (usually Gateways) that are<br />associated with the policy, and the status of the policy with respect to<br />each ancestor. When this policy attaches to a parent, the controller that<br />manages the parent and the ancestors MUST add an entry to this list when<br />the controller first sees the policy and SHOULD update the entry as<br />appropriate when the relevant ancestor is modified.<br /><br />Note that choosing the relevant ancestor is left to the Policy designers;<br />an important part of Policy design is designing the right object level at<br />which to namespace this status.<br /><br />Note also that implementations MUST ONLY populate ancestor status for<br />the Ancestor resources they are responsible for. Implementations MUST<br />use the ControllerName field to uniquely identify the entries in this list<br />that they are responsible for.<br /><br />Note that to achieve this, the list of PolicyAncestorStatus structs<br />MUST be treated as a map with a composite key, made up of the AncestorRef<br />and ControllerName fields combined.<br /><br />A maximum of 16 ancestors will be represented in this list. An empty list<br />means the Policy is not relevant for any ancestors.<br /><br />If this slice is full, implementations MUST NOT add further entries.<br />Instead they MUST consider the policy unimplementable and signal that<br />on any related resources such as the ancestor that would be referenced<br />here. For example, if this list was full on BackendTLSPolicy, no<br />additional Gateways would be able to reference the Service targeted by<br />the BackendTLSPolicy. |  | MaxItems: 16 <br /> |


#### PortNumber

_Underlying type:_ _integer_





_Appears in:_
- [ListenerEntry](#listenerentry)
- [ListenerEntryStatus](#listenerentrystatus)
- [ParentReference](#parentreference)



#### ProtocolType

_Underlying type:_ _string_





_Appears in:_
- [ListenerEntry](#listenerentry)

| Field | Description |
| --- | --- |
| `HTTP` | Accepts cleartext HTTP/1.1 sessions over TCP. Implementations MAY also<br />support HTTP/2 over cleartext. If implementations support HTTP/2 over<br />cleartext on "HTTP" listeners, that MUST be clearly documented by the<br />implementation.<br /> |
| `HTTPS` | Accepts HTTP/1.1 or HTTP/2 sessions over TLS.<br /> |
| `TLS` | Accepts TLS sessions over TCP.<br /> |
| `TCP` | Accepts TCP sessions.<br /> |
| `UDP` | Accepts UDP packets.<br /> |


#### RequestRate



RequestRate expresses a rate of requests over a given period of time.



_Appears in:_
- [RetryConstraint](#retryconstraint)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `count` _integer_ | Count specifies the number of requests per time interval.<br /><br />Support: Extended |  | Maximum: 1e+06 <br />Minimum: 1 <br /> |
| `interval` _[Duration](#duration)_ | Interval specifies the divisor of the rate of requests, the amount of<br />time during which the given count of requests occur.<br /><br />Support: Extended |  | Pattern: `^([0-9]\{1,5\}(h\|m\|s\|ms))\{1,4\}$` <br /> |


#### RetryConstraint



RetryConstraint defines the configuration for when to retry a request.



_Appears in:_
- [BackendTrafficPolicySpec](#backendtrafficpolicyspec)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `budget` _[BudgetDetails](#budgetdetails)_ | Budget holds the details of the retry budget configuration. | \{ interval:10s percent:20 \} |  |
| `minRetryRate` _[RequestRate](#requestrate)_ | MinRetryRate defines the minimum rate of retries that will be allowable<br />over a specified duration of time.<br /><br />The effective overall minimum rate of retries targeting the backend<br />service may be much higher, as there can be any number of clients which<br />are applying this setting locally.<br /><br />This ensures that requests can still be retried during periods of low<br />traffic, where the budget for retries may be calculated as a very low<br />value.<br /><br />Support: Extended | \{ count:10 interval:1s \} |  |


#### RouteGroupKind



RouteGroupKind indicates the group and kind of a Route resource.



_Appears in:_
- AllowedRoutes
- [ListenerEntryStatus](#listenerentrystatus)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `group` _[Group](#group)_ | Group is the group of the Route. | gateway.networking.k8s.io | MaxLength: 253 <br />Pattern: `^$\|^[a-z0-9]([-a-z0-9]*[a-z0-9])?(\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*$` <br /> |
| `kind` _[Kind](#kind)_ | Kind is the kind of the Route. |  | MaxLength: 63 <br />MinLength: 1 <br />Pattern: `^[a-zA-Z]([-a-zA-Z0-9]*[a-zA-Z0-9])?$` <br /> |


#### SectionName

_Underlying type:_ _string_





_Appears in:_
- [ListenerEntry](#listenerentry)
- [ListenerEntryStatus](#listenerentrystatus)
- [ParentReference](#parentreference)



#### SessionPersistence







_Appears in:_
- [BackendTrafficPolicySpec](#backendtrafficpolicyspec)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `sessionName` _string_ | SessionName defines the name of the persistent session token<br />which may be reflected in the cookie or the header. Users<br />should avoid reusing session names to prevent unintended<br />consequences, such as rejection or unpredictable behavior.<br /><br />Support: Implementation-specific |  | MaxLength: 128 <br /> |
| `absoluteTimeout` _[Duration](#duration)_ | AbsoluteTimeout defines the absolute timeout of the persistent<br />session. Once the AbsoluteTimeout duration has elapsed, the<br />session becomes invalid.<br /><br />Support: Extended |  | Pattern: `^([0-9]\{1,5\}(h\|m\|s\|ms))\{1,4\}$` <br /> |
| `idleTimeout` _[Duration](#duration)_ | IdleTimeout defines the idle timeout of the persistent session.<br />Once the session has been idle for more than the specified<br />IdleTimeout duration, the session becomes invalid.<br /><br />Support: Extended |  | Pattern: `^([0-9]\{1,5\}(h\|m\|s\|ms))\{1,4\}$` <br /> |
| `type` _[SessionPersistenceType](#sessionpersistencetype)_ | Type defines the type of session persistence such as through<br />the use a header or cookie. Defaults to cookie based session<br />persistence.<br /><br />Support: Core for "Cookie" type<br /><br />Support: Extended for "Header" type | Cookie | Enum: [Cookie Header] <br /> |
| `cookieConfig` _[CookieConfig](#cookieconfig)_ | CookieConfig provides configuration settings that are specific<br />to cookie-based session persistence.<br /><br />Support: Core |  |  |


#### XBackendTrafficPolicy



XBackendTrafficPolicy defines the configuration for how traffic to a
target backend should be handled.





| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `apiVersion` _string_ | `gateway.networking.x-k8s.io/v1alpha1` | | |
| `kind` _string_ | `XBackendTrafficPolicy` | | |
| `metadata` _[ObjectMeta](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.32/#objectmeta-v1-meta)_ | Refer to Kubernetes API documentation for fields of `metadata`. |  |  |
| `spec` _[BackendTrafficPolicySpec](#backendtrafficpolicyspec)_ | Spec defines the desired state of BackendTrafficPolicy. |  |  |
| `status` _[PolicyStatus](#policystatus)_ | Status defines the current state of BackendTrafficPolicy. |  |  |


#### XListenerSet



XListenerSet defines a set of additional listeners
to attach to an existing Gateway.





| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `apiVersion` _string_ | `gateway.networking.x-k8s.io/v1alpha1` | | |
| `kind` _string_ | `XListenerSet` | | |
| `metadata` _[ObjectMeta](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.32/#objectmeta-v1-meta)_ | Refer to Kubernetes API documentation for fields of `metadata`. |  |  |
| `spec` _[ListenerSetSpec](#listenersetspec)_ | Spec defines the desired state of ListenerSet. |  |  |
| `status` _[ListenerSetStatus](#listenersetstatus)_ | Status defines the current state of ListenerSet. | \{ conditions:[map[lastTransitionTime:1970-01-01T00:00:00Z message:Waiting for controller reason:Pending status:Unknown type:Accepted] map[lastTransitionTime:1970-01-01T00:00:00Z message:Waiting for controller reason:Pending status:Unknown type:Programmed]] \} |  |


