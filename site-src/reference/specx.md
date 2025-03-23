# API Reference

## Packages
- [gateway.networking.x-k8s.io/v1alpha1](#gatewaynetworkingx-k8siov1alpha1)


## gateway.networking.x-k8s.io/v1alpha1

Package v1alpha1 contains API Schema definitions for the gateway.networking.k8s-x.io
API group.


### Resource Types
- [XBackendTrafficPolicy](#xbackendtrafficpolicy)
- [XListenerSet](#xlistenerset)





#### BackendTrafficPolicySpec



BackendTrafficPolicySpec define the desired state of BackendTrafficPolicy
Note: there is no Override or Default policy configuration.



_Appears in:_
- [XBackendTrafficPolicy](#xbackendtrafficpolicy)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `targetRefs` _[LocalPolicyTargetReference](#localpolicytargetreference) array_ | TargetRefs identifies API object(s) to apply this policy to.<br />Currently, Backends (A grouping of like endpoints such as Service,<br />ServiceImport, or any implementation-specific backendRef) are the only<br />valid API target references.<br />Currently, a TargetRef can not be scoped to a specific port on a<br />Service. |  | MaxItems: 16 <br />MinItems: 1 <br /> |
| `retryConstraint` _[RetryConstraint](#retryconstraint)_ | RetryConstraint defines the configuration for when to allow or prevent<br />further retries to a target backend, by dynamically calculating a 'retry<br />budget'. This budget is calculated based on the percentage of incoming<br />traffic composed of retries over a given time interval. Once the budget<br />is exceeded, additional retries will be rejected.<br />For example, if the retry budget interval is 10 seconds, there have been<br />1000 active requests in the past 10 seconds, and the allowed percentage<br />of requests that can be retried is 20% (the default), then 200 of those<br />requests may be composed of retries. Active requests will only be<br />considered for the duration of the interval when calculating the retry<br />budget. Retrying the same original request multiple times within the<br />retry budget interval will lead to each retry being counted towards<br />calculating the budget.<br />Configuring a RetryConstraint in BackendTrafficPolicy is compatible with<br />HTTPRoute Retry settings for each HTTPRouteRule that targets the same<br />backend. While the HTTPRouteRule Retry stanza can specify whether a<br />request will be retried, and the number of retry attempts each client<br />may perform, RetryConstraint helps prevent cascading failures such as<br />retry storms during periods of consistent failures.<br />After the retry budget has been exceeded, additional retries to the<br />backend MUST return a 503 response to the client.<br />Additional configurations for defining a constraint on retries MAY be<br />defined in the future.<br />Support: Extended<br /><gateway:experimental> |  |  |
| `sessionPersistence` _[SessionPersistence](#sessionpersistence)_ | SessionPersistence defines and configures session persistence<br />for the backend.<br />Support: Extended |  |  |


#### BudgetDetails



BudgetDetails specifies the details of the budget configuration, like
the percentage of requests in the budget, and the interval between
checks.



_Appears in:_
- [RetryConstraint](#retryconstraint)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `percent` _integer_ | BudgetPercent defines the maximum percentage of active requests that may<br />be made up of retries.<br />Support: Extended | 20 | Maximum: 100 <br />Minimum: 0 <br /> |
| `interval` _[Duration](#duration)_ | BudgetInterval defines the duration in which requests will be considered<br />for calculating the budget for retries.<br />Support: Extended | 10s |  |












#### ListenerEntry







_Appears in:_
- [ListenerSetSpec](#listenersetspec)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `name` _[SectionName](#sectionname)_ | Name is the name of the Listener. This name MUST be unique within a<br />ListenerSet.<br />Name is not required to be unique across a Gateway and ListenerSets.<br />Routes can attach to a Listener by having a ListenerSet as a parentRef<br />and setting the SectionName |  |  |
| `hostname` _[Hostname](#hostname)_ | Hostname specifies the virtual hostname to match for protocol types that<br />define this concept. When unspecified, all hostnames are matched. This<br />field is ignored for protocols that don't require hostname based<br />matching.<br />Implementations MUST apply Hostname matching appropriately for each of<br />the following protocols:<br />* TLS: The Listener Hostname MUST match the SNI.<br />* HTTP: The Listener Hostname MUST match the Host header of the request.<br />* HTTPS: The Listener Hostname SHOULD match at both the TLS and HTTP<br />  protocol layers as described above. If an implementation does not<br />  ensure that both the SNI and Host header match the Listener hostname,<br />  it MUST clearly document that.<br />For HTTPRoute and TLSRoute resources, there is an interaction with the<br />`spec.hostnames` array. When both listener and route specify hostnames,<br />there MUST be an intersection between the values for a Route to be<br />accepted. For more information, refer to the Route specific Hostnames<br />documentation.<br />Hostnames that are prefixed with a wildcard label (`*.`) are interpreted<br />as a suffix match. That means that a match for `*.example.com` would match<br />both `test.example.com`, and `foo.test.example.com`, but not `example.com`. |  |  |
| `port` _[PortNumber](#portnumber)_ | Port is the network port. Multiple listeners may use the<br />same port, subject to the Listener compatibility rules. |  |  |
| `protocol` _[ProtocolType](#protocoltype)_ | Protocol specifies the network protocol this listener expects to receive. |  |  |
| `tls` _[GatewayTLSConfig](#gatewaytlsconfig)_ | TLS is the TLS configuration for the Listener. This field is required if<br />the Protocol field is "HTTPS" or "TLS". It is invalid to set this field<br />if the Protocol field is "HTTP", "TCP", or "UDP".<br />The association of SNIs to Certificate defined in GatewayTLSConfig is<br />defined based on the Hostname field for this listener.<br />The GatewayClass MUST use the longest matching SNI out of all<br />available certificates for any TLS handshake. |  |  |
| `allowedRoutes` _[AllowedRoutes](#allowedroutes)_ | AllowedRoutes defines the types of routes that MAY be attached to a<br />Listener and the trusted namespaces where those Route resources MAY be<br />present.<br />Although a client request may match multiple route rules, only one rule<br />may ultimately receive the request. Matching precedence MUST be<br />determined in order of the following criteria:<br />* The most specific match as defined by the Route type.<br />* The oldest Route based on creation timestamp. For example, a Route with<br />  a creation timestamp of "2020-09-08 01:02:03" is given precedence over<br />  a Route with a creation timestamp of "2020-09-08 01:02:04".<br />* If everything else is equivalent, the Route appearing first in<br />  alphabetical order (namespace/name) should be given precedence. For<br />  example, foo/bar is given precedence over foo/baz.<br />All valid rules within a Route attached to this Listener should be<br />implemented. Invalid Route rules can be ignored (sometimes that will mean<br />the full Route). If a Route rule transitions from valid to invalid,<br />support for that Route rule should be dropped to ensure consistency. For<br />example, even if a filter specified by a Route rule is invalid, the rest<br />of the rules within that Route should still be supported. | \{ namespaces:map[from:Same] \} |  |






#### ListenerEntryStatus



ListenerStatus is the status associated with a Listener.



_Appears in:_
- [ListenerSetStatus](#listenersetstatus)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `name` _[SectionName](#sectionname)_ | Name is the name of the Listener that this status corresponds to. |  |  |
| `port` _[PortNumber](#portnumber)_ | Port is the network port the listener is configured to listen on. |  |  |
| `supportedKinds` _[RouteGroupKind](#routegroupkind) array_ | SupportedKinds is the list indicating the Kinds supported by this<br />listener. This MUST represent the kinds an implementation supports for<br />that Listener configuration.<br />If kinds are specified in Spec that are not supported, they MUST NOT<br />appear in this list and an implementation MUST set the "ResolvedRefs"<br />condition to "False" with the "InvalidRouteKinds" reason. If both valid<br />and invalid Route kinds are specified, the implementation MUST<br />reference the valid Route kinds that have been specified. |  | MaxItems: 8 <br /> |
| `attachedRoutes` _integer_ | AttachedRoutes represents the total number of Routes that have been<br />successfully attached to this Listener.<br />Successful attachment of a Route to a Listener is based solely on the<br />combination of the AllowedRoutes field on the corresponding Listener<br />and the Route's ParentRefs field. A Route is successfully attached to<br />a Listener when it is selected by the Listener's AllowedRoutes field<br />AND the Route has a valid ParentRef selecting the whole Gateway<br />resource or a specific Listener as a parent resource (more detail on<br />attachment semantics can be found in the documentation on the various<br />Route kinds ParentRefs fields). Listener or Route status does not impact<br />successful attachment, i.e. the AttachedRoutes field count MUST be set<br />for Listeners with condition Accepted: false and MUST count successfully<br />attached Routes that may themselves have Accepted: false conditions.<br />Uses for this field include troubleshooting Route attachment and<br />measuring blast radius/impact of changes to a Listener. |  |  |
| `conditions` _[Condition](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.32/#condition-v1-meta) array_ | Conditions describe the current condition of this listener. |  | MaxItems: 8 <br /> |






#### ListenerSetSpec



ListenerSetSpec defines the desired state of a ListenerSet.



_Appears in:_
- [XListenerSet](#xlistenerset)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `parentRef` _[ParentGatewayReference](#parentgatewayreference)_ | ParentRef references the Gateway that the listeners are attached to. |  |  |
| `listeners` _[ListenerEntry](#listenerentry) array_ | Listeners associated with this ListenerSet. Listeners define<br />logical endpoints that are bound on this referenced parent Gateway's addresses.<br />Listeners in a `Gateway` and their attached `ListenerSets` are concatenated<br />as a list when programming the underlying infrastructure. Each listener<br />name does not need to be unique across the Gateway and ListenerSets.<br />See ListenerEntry.Name for more details.<br />Implementations MUST treat the parent Gateway as having the merged<br />list of all listeners from itself and attached ListenerSets using<br />the following precedence:<br />1. "parent" Gateway<br />2. ListenerSet ordered by creation time (oldest first)<br />3. ListenerSet ordered alphabetically by “\{namespace\}/\{name\}”.<br />An implementation MAY reject listeners by setting the ListenerEntryStatus<br />`Accepted`` condition to False with the Reason `TooManyListeners`<br />If a listener has a conflict, this will be reported in the<br />Status.ListenerEntryStatus setting the `Conflicted` condition to True.<br />Implementations SHOULD be cautious about what information from the<br />parent or siblings are reported to avoid accidentally leaking<br />sensitive information that the child would not otherwise have access<br />to. This can include contents of secrets etc. |  | MaxItems: 64 <br />MinItems: 1 <br /> |


#### ListenerSetStatus







_Appears in:_
- [XListenerSet](#xlistenerset)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `conditions` _[Condition](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.32/#condition-v1-meta) array_ | Conditions describe the current conditions of the ListenerSet.<br />Implementations MUST express ListenerSet conditions using the<br />`ListenerSetConditionType` and `ListenerSetConditionReason`<br />constants so that operators and tools can converge on a common<br />vocabulary to describe ListenerSet state.<br />Known condition types are:<br />* "Accepted"<br />* "Programmed" | [map[lastTransitionTime:1970-01-01T00:00:00Z message:Waiting for controller reason:Pending status:Unknown type:Accepted] map[lastTransitionTime:1970-01-01T00:00:00Z message:Waiting for controller reason:Pending status:Unknown type:Programmed]] | MaxItems: 8 <br /> |
| `listeners` _[ListenerEntryStatus](#listenerentrystatus) array_ | Listeners provide status for each unique listener port defined in the Spec. |  | MaxItems: 64 <br /> |








#### ParentGatewayReference



ParentGatewayReference identifies an API object including its namespace,
defaulting to Gateway.



_Appears in:_
- [ListenerSetSpec](#listenersetspec)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `group` _[Group](#group)_ | Group is the group of the referent. | gateway.networking.k8s.io |  |
| `kind` _[Kind](#kind)_ | Kind is kind of the referent. For example "Gateway". | Gateway |  |
| `name` _[ObjectName](#objectname)_ | Name is the name of the referent. |  |  |
| `namespace` _[Namespace](#namespace)_ | Namespace is the namespace of the referent.  If not present,<br />the namespace of the referent is assumed to be the same as<br />the namespace of the referring object. |  |  |








#### RequestRate



RequestRate expresses a rate of requests over a given period of time.



_Appears in:_
- [RetryConstraint](#retryconstraint)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `count` _integer_ | Count specifies the number of requests per time interval.<br />Support: Extended |  | Maximum: 1e+06 <br />Minimum: 1 <br /> |
| `interval` _[Duration](#duration)_ | Interval specifies the divisor of the rate of requests, the amount of<br />time during which the given count of requests occur.<br />Support: Extended |  |  |


#### RetryConstraint



RetryConstraint defines the configuration for when to retry a request.



_Appears in:_
- [BackendTrafficPolicySpec](#backendtrafficpolicyspec)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `budget` _[BudgetDetails](#budgetdetails)_ | Budget holds the details of the retry budget configuration. | \{ interval:10s percent:20 \} |  |
| `minRetryRate` _[RequestRate](#requestrate)_ | MinRetryRate defines the minimum rate of retries that will be allowable<br />over a specified duration of time.<br />The effective overall minimum rate of retries targeting the backend<br />service may be much higher, as there can be any number of clients which<br />are applying this setting locally.<br />This ensures that requests can still be retried during periods of low<br />traffic, where the budget for retries may be calculated as a very low<br />value.<br />Support: Extended | \{ count:10 interval:1s \} |  |








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


