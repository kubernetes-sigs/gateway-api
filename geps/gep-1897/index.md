# GEP-1897: BackendTLSPolicy - Explicit Backend TLS Connection Configuration

* Issue: [#1897](https://github.com/kubernetes-sigs/gateway-api/issues/1897)
* Status: Experimental

## TLDR

This document specifically addresses the topic of conveying HTTPS from the Gateway
dataplane to the backend (backend TLS termination), and intends to satisfy the single
use case “As a client implementation of Gateway API, I need to know how to connect to
a backend pod that has its own certificate”. TLS configuration can be a nebulous topic,
so in order to drive resolution this GEP focuses only on this single piece of functionality.

Furthermore, for Gateway API to handle the case where the service or backend owner is doing their own TLS, _and_
the service or backend owner wants to validate the clients connecting to it, two things need to happen:

- The service or backend owner has to provide a method for the Gateway owner to retrieve a certificate.
- Gateway API has to provide a way for the Gateway to configure and apply the validation options.

## Immediate Goals

1. The solution must satisfy the following use case: the backend pod has its own
certificate and the gateway implementation client needs to know how to connect to the
backend pod. (Use case #4 in [Gateway API TLS Use Cases](#references))
2. In terms of the Gateway API personas, only the application developer persona applies in this
solution. The application developer should control the gateway to backend TLS settings,
not the cluster operator, as requiring a cluster operator to manage certificate renewals
and revocations would be extremely cumbersome.
3. The solution should consider client certificate settings used in the TLS handshake **from
Gateway to backend**, such as server name indication, trusted certificates,
and CA certificates.

## Longer Term Goals

These are worthy goals, but deserve a different GEP for proper attention.  This GEP is concerned entirely with the
controlplane, i.e. the hop between gateway and backend.

1. [TCPRoute](https://gateway-api.sigs.k8s.io/reference/spec/#gateway.networking.k8s.io/v1alpha2.TCPRoute) and
[GRPCRoute](https://gateway-api.sigs.k8s.io/reference/spec/#gateway.networking.k8s.io/v1alpha2.GRPCRoute) use cases
are not addressed here, because at this point in time these two route types are not graduated to beta.
2. Mutual TLS (mTLS) use cases are intentionally out of scope for this GEP for two reasons.  First, the design of Gateway
API is backend-attached and does not currently support mutual authentication, and also because this GEP does not
address the case where connections to TLS are **implicitly configured** on behalf of the user, which is the norm for mTLS.
This GEP is about the case where an application developer needs to **explicitly express** that they expect TLS when
there is no automatic, implicit configuration available.
3. Service mesh use cases are not addressed here because this GEP is specifically concerned with the connection between
Gateways and Backends, not Service to Service.  Service mesh use cases should ignore the design components described in
this proposal.

## Non-Goals

These are worthy goals, but will not be covered by this GEP.

1. Changes to the existing mechanisms for edge or passthrough TLS termination
2. Providing a mechanism to decorate multiple route instances
3. TLSRoute use cases
4. UDPRoute use cases
5. Controlling TLS versions or cipher suites used in TLS handshakes. (Use case #5 in [Gateway API TLS Use Cases](#references))
6. Controlling certificates used by more than one workload (#6 in [Gateway API TLS Use Cases](#references))
7. Client certificate settings used in TLS **from external clients to the
Listener** (#7 in [Gateway API TLS Use Cases](#references))
8. Providing a mechanism for the cluster operator to override gateway to backend TLS settings.

## Already Solved TLS Use Cases

These are worthy goals that are already solved and thus will not be modified by the implementation.

1. Termination of TLS for HTTP routing (#1 in [Gateway API TLS Use Cases](#references))
2. HTTPS passthrough use cases (#2 in [Gateway API TLS Use Cases](#references))
3. Termination of TLS for non-HTTP TCP streams (#3 in [Gateway API TLS Use Cases](#references))

## Overview - what do we want to do?

Given that the current ingress solution specifies **edge** TLS termination (from the client to
the gateway), and how to handle **passthrough** TLS (from the client to the backend pod), this
proposed ingress solution specifies TLS origination to the **backend** (from the gateway to the
backend pod).  As mentioned, this solution satisfies the use case in which the backend pod
has its own certificate and the gateway client needs to know how to connect to the backend pod.

![image depicting TLS termination types](images/1897-TLStermtypes.png "TLS termination types")

Gateway API is missing a mechanism for separately providing the details for the backend TLS handshake,
including (but not limited to):

* intent to use TLS on the backend hop
* client certificate of the gateway
* system certificates to use in the absence of client certificates

## Purpose - why do we want to do this?

This proposal is _very_ tightly scoped because we have tried and failed to address this well-known
gap in the API specification. The lack of support for this fundamental concept is holding back
Gateway API adoption by users that require a solution to the use case. One of the recurring themes
that has held up the prior art has been interest related to service mesh, and as such this proposal
focuses explicitly on the ingress use case in the initial round.  Another reason for the tight scope
is that we have been too focused on a generic representation of everything that TLS can do, which
covers too much ground to address in a single GEP.

## The history of backend TLS

Work on this topic has spanned over three years, as documented in our repositories and other references,
and summarized below.

In January 2020, in issue [TLS Termination Policy #52](https://github.com/kubernetes-sigs/gateway-api/issues/52),
this use case was discussed.  The discussion ended after being diverted by
[KEP: Adding AppProtocol to Services and Endpoints #1422](https://github.com/kubernetes/enhancements/pull/1422),
which was implemented and later reverted.

In February 2020, [HTTPRoute: Add Reencrypt #81](https://github.com/kubernetes-sigs/gateway-api/pull/81)
added the dataplane feature as “reencrypt”, but it went stale and was closed in favor of the work done in the
next paragraph, which unfortunately didn’t implement the backend TLS termination feature.

In August 2020, it resurfaced with a [comment](https://github.com/kubernetes-sigs/gateway-api/pull/256/files#r472734392)
on this pull request: [tls: introduce mode and sni to cert matching behavior](https://github.com/kubernetes-sigs/gateway-api/pull/256/files#top).
The backend TLS termination feature was deferred at that time.  Other TLS discussion was documented in
[[SIG-NETWORK] TLS config in service-apis](https://docs.google.com/document/d/15fkzMrhN_7tA-i2mHKwZpqcjN1o2Pe9Am9Qt828x1lo/edit#heading=h.wym7wehwll44)
, a list of TLS features that had been collected in June 2020, itself based on spreadsheet
[Service API: TLS related issues](https://docs.google.com/spreadsheets/d/18KE61Y6InCmoQHZcbrYYRZS5Cnt7n33s5dTxUlhHgIA/edit#gid=0).

In December 2021, this was discussed as a beta blocker in issue
[Docs mentions Reencrypt for HTTPRoute and TLSRoute is available #968](https://github.com/kubernetes-sigs/gateway-api/issues/968).

A March 2022 issue documents another request for it: [Provide a way to configure TLS from a Gateway to Backends #1067](https://github.com/kubernetes-sigs/gateway-api/issues/1067)

A June 2022 issue documents a documentation issue related to it:
[Unclear how to specify upstream (webserver) HTTP protocol #1244](https://github.com/kubernetes-sigs/gateway-api/discussions/1244)

A July 2022 discussion [Specify Re-encrypt TLS Termination (i.e., Upstream TLS) #1285](https://github.com/kubernetes-sigs/gateway-api/discussions/1285)
collected most of the historical context preceding the backend TLS termination feature, with the intention of
collecting evidence that this feature is still unresolved.  This was followed by
[GEP: Describe Backend Properties #1282](https://github.com/kubernetes-sigs/gateway-api/issues/1282).

In August 2022, [Add Provisional GEP-1282 document #1333](https://github.com/kubernetes-sigs/gateway-api/pull/1333)
was created, and in October 2022, a GEP update with proposed implementation
[GEP-1282 Backend Properties - Update implementation #1430](https://github.com/kubernetes-sigs/gateway-api/pull/1430)
was followed by intense discussion and closed in favor of a downsize in scope.

In January 2023 we closed GEP-1282 and began a new discussion on enumerating TLS use cases in
[Gateway API TLS Use Cases](#references), for the purposes of a clear definition and separation of concerns.
This GEP is the outcome of the TLS use case #4 in
[Gateway API TLS Use Cases](#references) as mentioned in the Immediate Goals section above.

## API

To allow the gateway client to know how to connect to the backend pod, when the backend pod has its own
certificate, we implement a metaresource named `BackendTLSPolicy`, that was previously introduced with the name
`TLSConnectionPolicy` as a hypothetical Direct Policy Attachment example in
[GEP-713: Metaresources and PolicyAttachment](https://gateway-api.sigs.k8s.io/geps/gep-713/).
Because naming is hard, a new name may be
substituted without blocking acceptance of the content of the API change.

The selection of the applicable Gateway API persona is important in the design of BackendTLSPolicy, because it provides
a way to explicitly describe the _expectations_ of the connection to the application.  BackendTLSPolicy is configured
by the application developer Gateway API persona to signal what the application developer _expects_ in connections to
the application, from a TLS perspective.  Only the application developer can know what the application expects, so it is
important that this configuration be managed by that persona.

During the course of discussion of this proposal, we did consider allowing the cluster operator persona to have some access
to Gateway cert validation, but as mentioned, BackendTLSPolicy is used primarily to signal what the application
developer expects in the connection.  Granting this expectation to any other role would blur the lines between role
responsibilities, which compromises the role-oriented design principle of Gateway API. As mentioned in Non-goal #8,
providing a mechanism for the cluster operator gateway role to override gateway to backend TLS settings is not covered
by this proposal, but should be addressed in a future update.  One idea is to use two types: ApplicationBackendTLSPolicy,
and GatewayBackendTLSPolicy, where the application developer is responsible for the former, the cluster operator is
responsible for the latter, and the cluster operator may configure whether certain settings may be overridden by
application developers.

The BackendTLSPolicy must contain these configuration items to allow the Gateway to operate successfully
as a TLS Client:

- An explicit signal that TLS should be used by this connection.
- A hostname the Gateway should use to connect to the backend.
- A reference to one or more certificates to use in the TLS handshake, signed by a CA or self-signed.
- An indication that system certificates may be used.

BackendTLSPolicy is defined as a Direct Policy Attachment without defaults or overrides, applied to a Service that
accesses the backend in question, where the BackendTLSPolicy resides in the same namespace as the Service it is
applied to.  The BackendTLSPolicy and the Service must reside in the same namespace in order to prevent the
complications involved with sharing trust across namespace boundaries.  We chose the Service resource as a target,
rather than the Route resource, so that we can reuse the same BackendTLSPolicy for all the different Routes that
might point to this Service.
For the use case where certificates are stored in their own namespace, users may create Secrets and use ReferenceGrants
for a BackendTLSPolicy-to-Secret binding.  Implementations must respect a ReferenceGrant for cross-namespace Secret
sharing to BackendTLSPolicy, even if they don't for other cross-namespace sharing.

One of the areas of concern for this API is that we need to indicate how and when the API implementations should use the
backend destination certificate authority.  This solution proposes, as introduced in
[GEP-713](https://gateway-api.sigs.k8s.io/geps/gep-713/), that the implementation
should watch the connections to the specified TargetRefs (Services), and if a Service matches a BackendTLSPolicy, then
assume the connection is TLS, and verify that the TargetRef’s certificate can be validated by the client (Gateway) using
the provided certificates and hostname before the connection is made. On the question of how to signal
that there was a failure in the certificate validation, this is left up to the implementation to return a response error
that is appropriate, such as one of the HTTP error codes: 400 (Bad Request), 401 (Unauthorized), 403 (Forbidden), or
other signal that makes the failure sufficiently clear to the requester without revealing too much about the transaction,
based on established security requirements.

All policy resources must include `TargetRefs` with the fields specified
[here](https://github.com/kubernetes-sigs/gateway-api/blob/a33a934af9ec6997b34fd9b00d2ecd13d143e48b/apis/v1alpha2/policy_types.go#L24-L41).
In an upcoming [extension](https://github.com/kubernetes-sigs/gateway-api/issues/2147) to TargetRefs, policy resources
_may_ also choose to include `SectionName` and/or `Port` in the target reference following the same mechanics as `ParentRef`.

BackendTLSPolicySpec contains the `TargetRefs` and `Validation` fields.  The `Validation` field is a
`BackendTLSPolicyValidation` and contains `CACertificateRefs`, `WellKnownCACertificates`, and `Hostname`.
The names of the fields were chosen to facilitate discussion, but may be substituted without blocking acceptance of the
content of the API change. In fact, the `CertRefs` field name was changed to CACertRefs and then to
CACertificateRefs as of April 2024.

The `CACertificateRefs` and `WellKnownCACertificates` fields are both optional, but one of them must be set for a valid TLS
configuration. CACertificateRefs is an implementation-specific slice of
named object references, each containing a single cert. We originally proposed to follow the convention established by the
[CertificateRefs field on Gateway](https://github.com/kubernetes-sigs/gateway-api/blob/18e79909f7310aafc625ba7c862dfcc67b385250/apis/v1beta1/gateway_types.go#L340)
, but the CertificateRef requires both a tls.key and tls.crt and a certificate reference only requires the tls.crt.
WellKnownCACertificates is an optional enum that allows users to specify whether to use the set of CA certificates trusted by the
Gateway (WellKnownCACertificates specified as "System"), or to use the existing CACertificateRefs (WellKnownCACertificates
specified as "").  The use and definition of system certificates is implementation-dependent, and the intent is that
these certificates are obtained from the underlying operating system. CACertificateRefs contains one or more
references to Kubernetes objects that contain PEM-encoded TLS certificates, which are used to establish a TLS handshake
between the gateway and backend pod. References to a resource in a different namespace are invalid.
If ClientCertifcateRefs is unspecified, then WellKnownCACertificates must be set to "System" for a valid configuration.
If WellKnownCACertificates is unspecified, then CACertificateRefs must be specified with at least one entry for a valid configuration.
If WellKnownCACertficates is set to "System" and there are no system trusted certificates or the implementation doesn't define system
trusted certificates, then the associated TLS connection must fail.

The `Hostname` field is required and is to be used to configure the SNI the Gateway should use to connect to the backend.
Implementations must validate that at least one name in the certificate served by the backend matches this field.
We originally proposed using a list of allowed Subject Alternative Names, but determined that this was [not needed in
the first round](https://github.com/kubernetes-sigs/gateway-api/pull/2113#issuecomment-1696127092),
but may be added in the future.

We originally proposed allowing the configuration of expected TLS versions, but determined that this was [not needed in
the first round](https://github.com/kubernetes-sigs/gateway-api/pull/2113#issuecomment-1696127092).

Thus, the following additions would be made to the Gateway API:

```go
//TODO: Will update this section once API changes from PR 2955 are approved.
```

## How a client behaves

This table describes the effect that a BackendTLSPolicy has on a Route.  There are only two cases where the
BackendTLSPolicy will signal a Route to connect to a backend using TLS, an HTTPRoute with a backend that is targeted
by a BackendTLSPolicy, either with or without listener TLS configured.  (There are a few other cases where it may be
possible, but is implementation dependent.)

Every implementation that claims supports for BackendTLSPolicy should document for which Routes it is being implemented.

| Route Type | Gateway Config             | Backend is targeted by a BackendTLSPolicy? | Connect to backend  with TLS? |
|------------|----------------------------|-----------------------------------------------|-------------------------------|
| HTTPRoute  | Listener tls               | Yes                                           | **Yes**                       |
| HTTPRoute  | No listener tls            | Yes                                           | **Yes**                       |
| HTTPRoute  | Listener tls               | No                                            | No                            |
| HTTPRoute  | No listener tls            | No                                            | No                            |
| TLSRoute   | Listener Mode: Passthrough | Yes                                           | No                            |
| TLSRoute   | Listener Mode: Terminate   | Yes                                           | Implementation-dependent      |
| TLSRoute   | Listener Mode: Passthrough | No                                            | No                            |
| TLSRoute   | Listener Mode: Terminate   | No                                            | No                            |
| TCPRoute   | Listener TLS               | Yes                                           | Implementation-dependent      |
| TCPRoute   | No listener TLS            | Yes                                           | Implementation-dependent      |
| TCPRoute   | Listener TLS               | No                                            | No                            |
| TCPRoute   | No listener TLS            | No                                            | No                            |
| UDPRoute   | Listener TLS               | Yes                                           | No                            |
| UDPRoute   | No listener TLS            | Yes                                           | No                            |
| UDPRoute   | Listener TLS               | No                                            | No                            |
| UDPRoute   | No listener TLS            | No                                            | No                            |
| GRPCRoute  | Listener TLS               | Yes                                           | Implementation-dependent      |
| GRPCRoute  | No Listener TLS            | Yes                                           | Implementation-dependent      |
| GRPCRoute  | Listener TLS               | No                                            | No                            |
| GRPCRoute  | No Listener TLS            | No                                            | No                            |

## Request Flow

Step 6 would be changed in the typical client/gateway API request flow for a gateway implemented using a
reverse proxy. This is shown as **bolded** additions in step 6 below.

1. A client makes a request to http://foo.example.com.
2. DNS resolves the name to a Gateway address.
3. The reverse proxy receives the request on a Listener and uses the Host header to match an HTTPRoute.
4. Optionally, the reverse proxy can perform request header and/or path matching based on match rules of the HTTPRoute.
5. Optionally, the reverse proxy can modify the request, i.e. add/remove headers, based on filter rules of the HTTPRoute.
6. Lastly, the reverse proxy **optionally performs a TLS handshake** and forwards the request to one or more objects,
i.e. Service, in the cluster based on backendRefs rules of the HTTPRoute **and the TargetRefs of the BackendTLSPolicy**.

## Alternatives
Most alternatives are enumerated in the section "The history of backend TLS".  A couple of additional
alternatives are also listed here.

1. Expand BackendRef, which is already an expansion point.  At first, it seems logical that since listeners are handling
the client-gateway certs, BackendRefs could handle the gateway-backend certs.  However, when multiple Routes to target
the same Service, there would be unnecessary copying of the BackendRef every time the Service was targeted.  As well,
there could be multiple bBackendRefs with multiple rules on a rRoute, each of which might need the gateway-backend cert
configuration, so it is not the appropriate pattern.
2. Extend HTTPRoute to indicate TLS backend support. Extending HTTPRoute would interfere with deployed implementations
too much to be a practical solution.
3. Add a new type of Route for backend TLS.  This is impractical because we might want to enable backend TLS on other
route types in the future, and because we might want to have both TLS listeners and backend TLS on a single route.

## Prior Art

TLS from gateway to backend for ingress exists in several implementations, and was developed independently.

### Istio Gateway supports this with a DestinationRule:

* A secret representing a certificate/key pair, where the certificate is valid for the route host
* Set Gateway spec.servers[].port.protocol: HTTPS, spec.servers[].tls.mode=SIMPLE, spec.servers[].tls.credentialName
* Set DestinationRule spec.trafficPolicy.tls.mode: SIMPLE

Ref: [Istio / Understanding TLS Configuration](https://istio.io/latest/docs/ops/configuration/traffic-management/tls-configuration/#gateways)
and [Istio / Destination Rule](https://istio.io/latest/docs/reference/config/networking/destination-rule/#ClientTLSSettings)

### OpenShift Route (comparable to GW API Gateway) supports this with the following route configuration items:

* A certificate/key pair, where the certificate is valid for the route host
* A separate destination CA certificate enables the Ingress Controller to trust the destination’s certificate
* An optional, separate CA certificate that completes the certificate chain

Ref: [Secured routes - Configuring Routes | Networking | OpenShift Container Platform 4.12](https://docs.openshift.com/container-platform/4.12/networking/routes/secured-routes.html#nw-ingress-creating-a-reencrypt-route-with-a-custom-certificate_secured-routes)

### Contour supports this from Envoy to the backend using:

* An Envoy client certificate
* A CA certificate and SubjectName which are both used to verify the backend endpoint’s identity
* Kubernetes Service annotation: projectcontour.io/upstream-protocol.tls

Ref: [Upstream TLS](https://projectcontour.io/docs/v1.21.1/config/upstream-tls/)

### GKE supports a way to encrypt traffic to the backend pods using:

* `AppProtocol` on Service set to HTTPS
* Load balancer does not verify the certificate used by backend pods

Ref: [Secure a Gateway](https://cloud.google.com/kubernetes-engine/docs/how-to/secure-gateway#load-balancer-tls)

### Emissary supports encrypted traffic to services

* In the `Mapping` definition, set https:// in the spec.service field
* A spec.tls in the `Mapping` definition, with the name of a `TLSContext`
* A `TLSContext` to provide a client certificate, set minimum TLS version support, SNI

Ref: [TLS Origination](https://www.getambassador.io/docs/emissary/latest/topics/running/tls/origination)

### NGINX implementation through CRDs (Comparable to Route or Policy of Gateway API) supports both TLS and mTLS

* In the Upstream section of a VirtualServer or VirtualServerRoute (equivalent to HTTPRoute) there is a simple toggle to enable TLS.  This does not validate the certificate of the backend and implicitly trusts the backend in order to form the SSL tunnel.  This is not about validating the certificate but obfuscating the traffic with TLS/SSL.
* A Policy attachment can be provided when certification validation is required that is called egressMTLS (egress from the proxy to the upstream).  This can be tuned to perform various certificate validation tests.  It was created as a Policy becuase it implies some type of AuthN/AuthZ due to the additional checks.  This was also compatible with Open Service Mesh and NGINX Service Mesh and removed the need for a sidecar at the ingress controller.
* A corresponding 'IngressMTLS' policy also exists for mTLS verification of client connections to the proxy.  The Policy object is used for anything that implies AuthN/AuthZ.

Ref: [Upstream.TLS](https://docs.nginx.com/nginx-ingress-controller/configuration/virtualserver-and-virtualserverroute-resources/#upstreamtls)

Ref: [EgressMTLS](https://docs.nginx.com/nginx-ingress-controller/configuration/policy-resource/#egressmtls)

Ref: [IngressMTLS](https://docs.nginx.com/nginx-ingress-controller/configuration/policy-resource/#ingressmtls)

## Answered Questions

Q. Bowei recommended that we mention the approach of cross-namespace referencing between Route and Service.
Be explicit about using the standard rules with respect to attaching policies to resources.

A. This is mentioned in the
API section.

Q. Costin recommended that Gateway SHOULD authenticate with either a JWT with audience or client cert
or some other means - so gateway added headers can be trusted, amongst other things.

A. This is out of scope for this
proposal, which centers around application developer persona resources such as HTTPRoute and Service.

Q. Costin mentioned we need to answer the question - is configuring the connection to a backend and TLS
something the route author decides - or the backend owner?

A. This is decided by the application developer persona,
which would more likely, but not exclusively, be the backend owner.

Q.Costin continued, same for SAN (Subject Alternative Name) certificates.
The backend owner is the application developer, and the route owner will have to collaborate with the application
developer to provide the appropriate configuration for TLS.  The implementation would need to take the certificate
provided by the application and verify that it satisfies the requirements of the route-as-client, including SAN
information.  Sometimes the backend owner and route owner are the same entity.

A. This was most recently addressed by
adding hostname for SNI and removing allowed SANs.

## Graduation Criteria

This section is to record issues that were requested for discussion in the API section before this GEP graduates
out of `Provisional` status.

1. Rob Scott is interested in extending the TargetRef to optionally include port, since we are targeting the entirety
of a Service. See the discussion in https://github.com/kubernetes-sigs/gateway-api/pull/2113/files#r1231594914,
and follow up issue in https://github.com/kubernetes-sigs/gateway-api/issues/2147
2. Michael Pleshakov asked about conflicts that could arise when multiple implementations are running in a cluster.
This is a gap in our policy attachment model that needs to be addressed.  See the discussion in
https://github.com/kubernetes-sigs/gateway-api/pull/2113/files#r1235750540. Graduating this GEP to implementable
requires an update to the Policy GEP to define how status can be nested to support multiple implementations. This will
likely look very similar to Route status.
See [comment](https://github.com/kubernetes-sigs/gateway-api/pull/2113#issuecomment-1696127092).
3. Rob Scott [wanted to note](https://github.com/kubernetes-sigs/gateway-api/pull/2113#issuecomment-1696127092) that
when this graduates to the standard channel, implementations of HTTPRoute may also be
required to watch the BackendTLSPolicy. If one of these policies is attached to a Service targeted by an HTTPRoute,
the implementation would be required to fully implement the policy or mark the backend invalid.

## References

[Gateway API TLS Use Cases](https://docs.google.com/document/d/17sctu2uMJtHmJTGtBi_awGB0YzoCLodtR6rUNmKMCs8/edit#heading=h.cxuq8vo8pcxm)

[GEP-713: Metaresources and PolicyAttachment](https://gateway-api.sigs.k8s.io/geps/gep-713/)

[Policy Attachment](https://gateway-api.sigs.k8s.io/reference/policy-attachment/#direct-policy-attachment)

[Gateway API TLS](https://gateway-api.sigs.k8s.io/guides/tls/)

[SIG-NET Gateway API: TLS to the K8s.Service/Backend](https://docs.google.com/document/d/1RTYh2brg_vLX9o3pTcrWxtZSsf8Y5NQvIG52lpFcZlo)

[SAN vs SNI](https://serverfault.com/questions/807959/what-is-the-difference-between-san-and-sni-ssl-certificates)
