# GEP-1897: BackendTLSPolicy - Explicit Backend TLS Connection Configuration

* Issue: [#1897](https://github.com/kubernetes-sigs/gateway-api/issues/1897)
* Status: Standard

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
2. In this GEP, only the application developer persona will have control over TLS settings. This does not preclude adding other personas in future GEPs.
3. The solution should consider client TLS settings used in the TLS handshake **from
Gateway to backend**, such as server name indication and trusted CA certificates.
4. Both Gateway and Mesh use cases may be supported, depending on the implementation, and will be covered by features in each case.

## Longer Term Goals

These are worthy goals, but deserve a different GEP for proper attention.  This GEP is concerned entirely with the
hop between gateway client and backend.
1. [TCPRoute](../../apis/v1alpha2/tcproute_types.go) use cases are not addressed here, because at this point in time
TCPRoute is not graduated to beta.
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
5. GRPCRoute use cases
6. Controlling TLS versions or cipher suites used in TLS handshakes. (Use case #5 in [Gateway API TLS Use Cases](#references))
7. Controlling certificates used by more than one workload (#6 in [Gateway API TLS Use Cases](#references))
8. Client certificate settings used in TLS **from external clients to the
Listener** (#7 in [Gateway API TLS Use Cases](#references))
9. Service Mesh "mesh transport security".
10. Providing a mechanism for the cluster operator to override gateway to backend TLS settings.

> It is very common for service mesh implementations to implement some form of transparent transport security, whether
> that is WireGuard, mTLS, or others. This is completely orthogonal to the use cases being tackled by this GEP.
> * The "mesh transport security" is something invisible to the user's application, and is simply used to secure
> communication between components in the mesh.
> * This proposal, instead, explicitly calls for sending TLS **to the user's application**.
> However, this does not mean service meshes are outside of scope for this proposal, merely that only the
> application-level TLS configuration is in scope.

![](images/mesh.png "Mesh transport")

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
* CA certificates to trust
* other properties of the TLS handshake, such as SNI and SAN validation
* client certificate of the gateway (outside of scope for this GEP)

## Purpose - why do we want to do this?

This proposal is _very_ tightly scoped because we have tried and failed to address this well-known
gap in the API specification. The lack of support for this fundamental concept is holding back
Gateway API adoption by users that require a solution to the use case. 
Another reason for the tight scope
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
[GEP-713: Metaresources and PolicyAttachment](../gep-713/index.md).
Because naming is hard, a new name may be
substituted without blocking acceptance of the content of the API change.

The selection of the applicable Gateway API persona is important in the design of BackendTLSPolicy, because it provides
a way to explicitly describe the _expectations_ of the connection to the application.
In this GEP, BackendTLSPolicy will be configured only by the application developer Gateway API persona to tell gateway
clients how to connect to the application, from a TLS perspective.
Future iterations *may* expand this to additionally allow consumer overrides; see [Future plans](#future-plans).

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
- A reference to one or more CA certificates (which could include "system certificates") to validate the server's TLS certificates.

BackendTLSPolicy is defined as a Direct Policy Attachment without defaults or overrides, applied to a Service that
accesses the backend in question, where the BackendTLSPolicy resides in the same namespace as the Service it is
applied to.  For now, the BackendTLSPolicy and the Service must reside in the same namespace in order to prevent the
complications involved with sharing trust across namespace boundaries (see [Future plans](#future-plans)).  We chose the Service
resource as a target,
rather than the Route resource, so that we can reuse the same BackendTLSPolicy for all the different Routes that
might point to this Service.

One of the areas of concern for this API is that we need to indicate how and when the API implementations should use the
backend destination certificate authority.  This solution proposes, as introduced in
[GEP-713](../gep-713/index.md), that the implementation
should watch the connections to the specified TargetRefs (Services), and if a Service matches a BackendTLSPolicy, then
assume the connection is TLS, and verify that the TargetRef’s certificate can be validated by the client (Gateway) using
the provided certificates and hostname before the connection is made. On the question of how to signal
that there was a failure in the certificate validation, this is left up to the implementation to return a response error
that is appropriate, such as one of the HTTP error codes: 400 (Bad Request), 401 (Unauthorized), 403 (Forbidden), or
other signal that makes the failure sufficiently clear to the requester without revealing too much about the transaction,
based on established security requirements.

BackendTLSPolicy applies only to TCP traffic. If a policy explicitly attaches to a UDP port of a Service (that is, the
`targetRef` has a `sectionName` specifying a single port or the service has only 1 port), the `Accepted: False` Condition
with `Reason: Invalid` MUST be set. If the policy attaches to a mix of TCP and UDP ports, implementations SHOULD include
a warning in the `Accepted` condition message (`ancestors.conditions`); the policy will only be effective for the TCP ports.

All policy resources must include `TargetRefs` with the fields specified
in [PolicyTargetReference](https://github.com/kubernetes-sigs/gateway-api/blob/a33a934af9ec6997b34fd9b00d2ecd13d143e48b/apis/v1alpha2/policy_types.go#L24-L41).
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
If any of the CACertificateRefs cannot be resolved (e.g., the referenced resource does not exist) or is misconfigured
(e.g., ConfigMap does not contain a key named `ca.crt`), the `ResolvedRefs` status condition MUST be set to `False` with
`Reason: InvalidCACertificateRef`. Connections using that CACertificateRef MUST fail, and the client MUST receive an
HTTP 5xx error response.
References to objects with an unsupported Group and Kind are not valid, and MUST be rejected by the implementation with
the `ResolvedRefs` status condition set to `False` and `Reason: InvalidKind`.
Implementations MAY perform further validation of the certificate content (i.e., checking expiry or enforcing specific
formats). If they do, they MUST ensure that the `ResolvedRefs` Condition is `False` and use an implementation-specific
`Reason`, like `ExpiredCertificate` or similar.
If `ResolvedRefs` Condition is `False` implementations SHOULD include a message specifying which references are invalid
and explaining why.

If all CertificateRefs cannot be resolved, the BackendTLSPolicy is considered invalid and the implementation MUST set
the `Accepted` Condition to `False`, with a reason of `NoValidCACertificate` and a message explaining this.

WellKnownCACertificates is an optional enum that allows users to specify whether to use the set of CA certificates trusted by the
Gateway (WellKnownCACertificates specified as "System"), or to use the existing CACertificateRefs (WellKnownCACertificates
specified as "").  The use and definition of system certificates is implementation-dependent, and the intent is that
these certificates are obtained from the underlying operating system. CACertificateRefs contains one or more
references to Kubernetes objects that contain PEM-encoded TLS certificates, which are used to establish a TLS handshake
between the gateway and backend pod. References to a resource in a different namespace are invalid.
If ClientCertificateRefs is unspecified, then WellKnownCACertificates must be set to "System" for a valid configuration.
If WellKnownCACertificates is unspecified, then CACertificateRefs must be specified with at least one entry for a valid
configuration.
If an implementation does not support the WellKnownCACertificates, or the provided value is unsupported,the
BackendTLSPolicy is considered invalid, and the implementation MUST set the `Accepted` Condition to `False`, with a
reason of `Invalid` and a message explaining this.

For an invalid BackendTLSPolicy, implementations MUST NOT fall back to unencrypted (plaintext) connections. 
Instead, the corresponding TLS connection MUST fail, and the client MUST receive an HTTP 5xx error response.

Implementations MUST NOT modify any status other than their own. Ownership of a status is determined by the `controllerName`,
which identifies the responsible controller.

The `Hostname` field is required and is to be used to configure the SNI the Gateway should use to connect to the backend.
Implementations must validate that at least one name in the certificate served by the backend matches this field.
We originally proposed using a list of allowed Subject Alternative Names, but determined that this was [not needed in
the first round](https://github.com/kubernetes-sigs/gateway-api/pull/2113#issuecomment-1696127092),
but may be added in the future.

We originally proposed allowing the configuration of expected TLS versions, but determined that this was [not needed in
the first round](https://github.com/kubernetes-sigs/gateway-api/pull/2113#issuecomment-1696127092).

Thus, the following additions would be made to Gateway API.  See the
[BackendTLSPolicy API](https://kubernetes-sigs/gateway-api/blob/main/apis/v1/backendtlspolicy_types.go) for more
details.

```go
// +genclient
// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:storageversion
// +kubebuilder:resource:categories=gateway-api,shortName=btlspolicy
// +kubebuilder:printcolumn:name="Age",type=date,JSONPath=`.metadata.creationTimestamp`
//
// BackendTLSPolicy is a Direct Attached Policy.
// +kubebuilder:metadata:labels="gateway.networking.k8s.io/policy=Direct"

// BackendTLSPolicy provides a way to configure how a Gateway
// connects to a Backend via TLS.
type BackendTLSPolicy struct {
	metav1.TypeMeta `json:",inline"`
	// +optional
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// Spec defines the desired state of BackendTLSPolicy.
	// +required
	Spec BackendTLSPolicySpec `json:"spec"`

	// Status defines the current state of BackendTLSPolicy.
	// +optional
	Status PolicyStatus `json:"status,omitempty"`
}

// BackendTLSPolicyList contains a list of BackendTLSPolicies
// +kubebuilder:object:root=true
type BackendTLSPolicyList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []BackendTLSPolicy `json:"items"`
}

// BackendTLSPolicySpec defines the desired state of BackendTLSPolicy.
//
// Support: Extended
type BackendTLSPolicySpec struct {
	// TargetRefs identifies an API object to apply the policy to.
	// Only Services have Extended support. Implementations MAY support
	// additional objects, with Implementation Specific support.
	// Note that this config applies to the entire referenced resource
	// by default, but this default may change in the future to provide
	// a more granular application of the policy.
	//
	// TargetRefs must be _distinct_. This means either that:
	//
	// * They select different targets. If this is the case, then targetRef
	//   entries are distinct. In terms of fields, this means that the
	//   multi-part key defined by `group`, `kind`, and `name` must
	//   be unique across all targetRef entries in the BackendTLSPolicy.
	// * They select different sectionNames in the same target.
	//
	//
	// When more than one BackendTLSPolicy selects the same target and
	// sectionName, implementations MUST determine precedence using the
	// following criteria, continuing on ties:
	//
	// * The older policy by creation timestamp takes precedence. For
	//   example, a policy with a creation timestamp of "2021-07-15
	//   01:02:03" MUST be given precedence over a policy with a
	//   creation timestamp of "2021-07-15 01:02:04".
	// * The policy appearing first in alphabetical order by {name}.
	//   For example, a policy named `bar` is given precedence over a
	//   policy named `baz`.
	//
	// For any BackendTLSPolicy that does not take precedence, the
	// implementation MUST ensure the `Accepted` Condition is set to
	// `status: False`, with Reason `Conflicted`.
	//
	// Support: Extended for Kubernetes Service
	//
	// Support: Implementation-specific for any other resource
	//
	// +required
	// +listType=atomic
	// +kubebuilder:validation:MinItems=1
	// +kubebuilder:validation:MaxItems=16
	// +kubebuilder:validation:XValidation:message="sectionName must be specified when targetRefs includes 2 or more references to the same target",rule="self.all(p1, self.all(p2, p1.group == p2.group && p1.kind == p2.kind && p1.name == p2.name ? ((!has(p1.sectionName) || p1.sectionName == '') == (!has(p2.sectionName) || p2.sectionName == '')) : true))"
	// +kubebuilder:validation:XValidation:message="sectionName must be unique when targetRefs includes 2 or more references to the same target",rule="self.all(p1, self.exists_one(p2, p1.group == p2.group && p1.kind == p2.kind && p1.name == p2.name && (((!has(p1.sectionName) || p1.sectionName == '') && (!has(p2.sectionName) || p2.sectionName == '')) || (has(p1.sectionName) && has(p2.sectionName) && p1.sectionName == p2.sectionName))))"
	TargetRefs []LocalPolicyTargetReferenceWithSectionName `json:"targetRefs"`

	// Validation contains backend TLS validation configuration.
	// +required
	Validation BackendTLSPolicyValidation `json:"validation"`

	// Options are a list of key/value pairs to enable extended TLS
	// configuration for each implementation. For example, configuring the
	// minimum TLS version or supported cipher suites.
	//
	// A set of common keys MAY be defined by the API in the future. To avoid
	// any ambiguity, implementation-specific definitions MUST use
	// domain-prefixed names, such as `example.com/my-custom-option`.
	// Un-prefixed names are reserved for key names defined by Gateway API.
	//
	// Support: Implementation-specific
	//
	// +optional
	// +kubebuilder:validation:MaxProperties=16
	Options map[AnnotationKey]AnnotationValue `json:"options,omitempty"`
}

// BackendTLSPolicyValidation contains backend TLS validation configuration.
// +kubebuilder:validation:XValidation:message="must not contain both CACertificateRefs and WellKnownCACertificates",rule="!(has(self.caCertificateRefs) && size(self.caCertificateRefs) > 0 && has(self.wellKnownCACertificates) && self.wellKnownCACertificates != \"\")"
// +kubebuilder:validation:XValidation:message="must specify either CACertificateRefs or WellKnownCACertificates",rule="(has(self.caCertificateRefs) && size(self.caCertificateRefs) > 0 || has(self.wellKnownCACertificates) && self.wellKnownCACertificates != \"\")"
type BackendTLSPolicyValidation struct {
	// CACertificateRefs contains one or more references to Kubernetes objects that
	// contain a PEM-encoded TLS CA certificate bundle, which is used to
	// validate a TLS handshake between the Gateway and backend Pod.
	//
	// If CACertificateRefs is empty or unspecified, then WellKnownCACertificates must be
	// specified. Only one of CACertificateRefs or WellKnownCACertificates may be specified,
	// not both. If CACertificateRefs is empty or unspecified, the configuration for
	// WellKnownCACertificates MUST be honored instead if supported by the implementation.
	//
	// A CACertificateRef is invalid if:
	//
	// * It refers to a resource that cannot be resolved (e.g., the referenced resource
	//   does not exist) or is misconfigured (e.g., a ConfigMap does not contain a key
	//   named `ca.crt`). In this case, the Reason must be set to `InvalidCACertificateRef`
	//   and the Message of the Condition must indicate which reference is invalid and why.
	//
	// * It refers to an unknown or unsupported kind of resource. In this case, the Reason
	//   must be set to `InvalidKind` and the Message of the Condition must explain which
	//   kind of resource is unknown or unsupported.
	//
	// * It refers to a resource in another namespace. This may change in future
	//   spec updates.
	//
	// Implementations MAY choose to perform further validation of the certificate
	// content (e.g., checking expiry or enforcing specific formats). In such cases,
	// an implementation-specific Reason and Message must be set for the invalid reference.
	//
	// In all cases, the implementation MUST ensure the `ResolvedRefs` Condition on
	// the BackendTLSPolicy is set to `status: False`, with a Reason and Message
	// that indicate the cause of the error. Connections using an invalid
	// CACertificateRef MUST fail, and the client MUST receive an HTTP 5xx error
	// response. If ALL CACertificateRefs are invalid, the implementation MUST also
	// ensure the `Accepted` Condition on the BackendTLSPolicy is set to
	// `status: False`, with a Reason `NoValidCACertificate`.
	//
	//
	// A single CACertificateRef to a Kubernetes ConfigMap kind has "Core" support.
	// Implementations MAY choose to support attaching multiple certificates to
	// a backend, but this behavior is implementation-specific.
	//
	// Support: Core - An optional single reference to a Kubernetes ConfigMap,
	// with the CA certificate in a key named `ca.crt`.
	//
	// Support: Implementation-specific - More than one reference, other kinds
	// of resources, or a single reference that includes multiple certificates.
	//
	// +optional
	// +listType=atomic
	// +kubebuilder:validation:MaxItems=8
	CACertificateRefs []LocalObjectReference `json:"caCertificateRefs,omitempty"`

	// WellKnownCACertificates specifies whether system CA certificates may be used in
	// the TLS handshake between the gateway and backend pod.
	//
	// If WellKnownCACertificates is unspecified or empty (""), then CACertificateRefs
	// must be specified with at least one entry for a valid configuration. Only one of
	// CACertificateRefs or WellKnownCACertificates may be specified, not both.
	// If an implementation does not support the WellKnownCACertificates field, or
	// the supplied value is not recognized, the implementation MUST ensure the
	// `Accepted` Condition on the BackendTLSPolicy is set to `status: False`, with
	// a Reason `Invalid`.
	//
	// Support: Implementation-specific
	//
	// +optional
	// +listType=atomic
	WellKnownCACertificates *WellKnownCACertificatesType `json:"wellKnownCACertificates,omitempty"`

	// Hostname is used for two purposes in the connection between Gateways and
	// backends:
	//
	// 1. Hostname MUST be used as the SNI to connect to the backend (RFC 6066).
	// 2. Hostname MUST be used for authentication and MUST match the certificate
	//    served by the matching backend, unless SubjectAltNames is specified.
	// 3. If SubjectAltNames are specified, Hostname can be used for certificate selection
	//    but MUST NOT be used for authentication. If you want to use the value
	//    of the Hostname field for authentication, you MUST add it to the SubjectAltNames list.
	//
	// Support: Core
	//
	// +required
	Hostname PreciseHostname `json:"hostname"`

	// SubjectAltNames contains one or more Subject Alternative Names.
	// When specified the certificate served from the backend MUST
	// have at least one Subject Alternate Name matching one of the specified SubjectAltNames.
	//
	// Support: Extended
	//
	// +optional
	// +listType=atomic
	// +kubebuilder:validation:MaxItems=5
	SubjectAltNames []SubjectAltName `json:"subjectAltNames,omitempty"`
}

// SubjectAltName represents Subject Alternative Name.
// +kubebuilder:validation:XValidation:message="SubjectAltName element must contain Hostname, if Type is set to Hostname",rule="!(self.type == \"Hostname\" && (!has(self.hostname) || self.hostname == \"\"))"
// +kubebuilder:validation:XValidation:message="SubjectAltName element must not contain Hostname, if Type is not set to Hostname",rule="!(self.type != \"Hostname\" && has(self.hostname) && self.hostname != \"\")"
// +kubebuilder:validation:XValidation:message="SubjectAltName element must contain URI, if Type is set to URI",rule="!(self.type == \"URI\" && (!has(self.uri) || self.uri == \"\"))"
// +kubebuilder:validation:XValidation:message="SubjectAltName element must not contain URI, if Type is not set to URI",rule="!(self.type != \"URI\" && has(self.uri) && self.uri != \"\")"
type SubjectAltName struct {
	// Type determines the format of the Subject Alternative Name. Always required.
	//
	// Support: Core
	//
	// +required
	Type SubjectAltNameType `json:"type"`

	// Hostname contains Subject Alternative Name specified in DNS name format.
	// Required when Type is set to Hostname, ignored otherwise.
	//
	// Support: Core
	//
	// +optional
	Hostname Hostname `json:"hostname,omitempty"`

	// URI contains Subject Alternative Name specified in a full URI format.
	// It MUST include both a scheme (e.g., "http" or "ftp") and a scheme-specific-part.
	// Common values include SPIFFE IDs like "spiffe://mycluster.example.com/ns/myns/sa/svc1sa".
	// Required when Type is set to URI, ignored otherwise.
	//
	// Support: Core
	//
	// +optional
	URI AbsoluteURI `json:"uri,omitempty"`
}

// WellKnownCACertificatesType is the type of CA certificate that will be used
// when the caCertificateRefs field is unspecified.
// +kubebuilder:validation:Enum=System
type WellKnownCACertificatesType string

const (
	// WellKnownCACertificatesSystem indicates that well known system CA certificates should be used.
	WellKnownCACertificatesSystem WellKnownCACertificatesType = "System"
)

// SubjectAltNameType is the type of the Subject Alternative Name.
// +kubebuilder:validation:Enum=Hostname;URI
type SubjectAltNameType string

const (
	// HostnameSubjectAltNameType specifies hostname-based SAN.
	//
	// Support: Core
	HostnameSubjectAltNameType SubjectAltNameType = "Hostname"

	// URISubjectAltNameType specifies URI-based SAN, e.g. SPIFFE id.
	//
	// Support: Core
	URISubjectAltNameType SubjectAltNameType = "URI"
)

const (
	// This reason is used with the "Accepted" condition when it is
	// set to false because all CACertificateRefs of the
	// BackendTLSPolicy are invalid.
	BackendTLSPolicyReasonNoValidCACertificate PolicyConditionReason = "NoValidCACertificate"
)

const (
	// This condition indicates whether the controller was able to resolve all
	// object references for the BackendTLSPolicy.
	//
	// Possible reasons for this condition to be True are:
	//
	// * "ResolvedRefs"
	//
	// Possible reasons for this condition to be False are:
	//
	// * "InvalidCACertificateRef"
	// * "InvalidKind"
	//
	// Controllers may raise this condition with other reasons, but should
	// prefer to use the reasons listed above to improve interoperability.
	BackendTLSPolicyConditionResolvedRefs PolicyConditionType = "ResolvedRefs"

	// This reason is used with the "ResolvedRefs" condition when the condition
	// is true.
	BackendTLSPolicyReasonResolvedRefs PolicyConditionReason = "ResolvedRefs"

	// This reason is used with the "ResolvedRefs" condition when one of the
	// BackendTLSPolicy's CACertificateRefs is invalid.
	// A CACertificateRef is considered invalid when it refers to a nonexistent
	// resource or when the data within that resource is malformed.
	BackendTLSPolicyReasonInvalidCACertificateRef PolicyConditionReason = "InvalidCACertificateRef"

	// This reason is used with the "ResolvedRefs" condition when one of the
	// BackendTLSPolicy's CACertificateRefs references an unknown or unsupported
	// Group and/or Kind.
	BackendTLSPolicyReasonInvalidKind PolicyConditionReason = "InvalidKind"
)
```

## How a client behaves

The `BackendTLSPolicy` tells a client "Connect to this service using TLS".
This is unconditional to the type of traffic the gateway client is forwarding.

For instance, the following will all have the gateway client add TLS if the backend is targeted by a BackendTLSPolicy:

* A Gateway accepts traffic on an HTTP listener
* A Gateway accepts and terminates TLS on an HTTPS listener
* A Gateway accepts traffic on a TCP listener

There is no need for a Gateway that accepts traffic with `Mode: Passthrough` to do anything differently here, but
implementations MAY choose to treat TLS passthrough as a special case. Implementations that do this SHOULD clearly
document their approach if BackendTLSPolicy is treated differently for TLS passthrough.

Note that there are cases where these patterns may result in multiple layers of TLS on a single connection.
There may be even cases where the gateway implementation is unaware of this; for example, processing TCPRoute traffic --
the traffic may or may not be TLS, and the gateway would be unaware.
This is intentional to allow full fidelity of the API, as this is commonly desired for tunneling scenarios.
When users do not want this, they should ensure that the BackendTLSPolicy is not incorrectly applied to traffic that is
already TLS.
The [Future Plans](#future-plans) include more controls over the API to make this easier to manage.

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

## Future plans

In order to scope this GEP, some changes are deferred to a near-future GEP.
This GEP intends to add the ability for additional control by gateway clients to override TLS settings, following previously
established patterns of [consumer and producer policies]([glossary](https://gateway-api.sigs.k8s.io/concepts/glossary/?h=gloss#producer-route)).
Additionally, more contextual control over when to apply the policies will be explored, to enable use cases like "apply
TLS only from this route" ([issue](https://github.com/kubernetes-sigs/gateway-api/issues/3856)).

While the details of these plans are out of scope for this GEP it is important to be aware of the future plans for the
API to ensure the immediate-term plans are future-proofed against the proposed plans.

Implementations should plan for the existence of future fields that may be added that will control where the TLS policy
applies.
These may include, but are not limited to:

* `spec.targetRefs.namespace`
* `spec.targetRefs.from`
* `spec.mode`

While in some cases adding new fields may be seen as a backwards compatibility risk, due to older implementations not
knowing to respect the fields, these fields (or similar, should future GEPs decide on new names) are pre-approved to be
added in a future release, should the GEPs to add them are approved in the first place.

## Outstanding issues

### Multiple TargetRefs rolling up to the same Gateway cannot be represented in status

It is possible to have a BackendTLSPolicy target multiple, different Services that are used in HTTPRoutes that attach
to the same Gateway.

As written, the Status section of BackendTLSPolicy does not have a way to represent these separate statuses, as the
status is namespaced by `controllerName` and `ancestorRef` (where "ancestor" is the Gateway in this case).

We need to decide if this is enough of an issue to change the status design, or if we record this as a design decision
and accept the tradeoff.


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
* A Policy attachment can be provided when certification validation is required that is called egressMTLS (egress from the proxy to the upstream).  This can be tuned to perform various certificate validation tests.  It was created as a Policy because it implies some type of AuthN/AuthZ due to the additional checks.  This was also compatible with Open Service Mesh and NGINX Service Mesh and removed the need for a sidecar at the ingress controller.
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

A. This was most recently addressed by adding hostname for SNI and removing allowed SANs.

Q. Rob Scott is interested in extending the TargetRef to optionally include port, since we are targeting the entirety
of a Service. See the discussion in https://github.com/kubernetes-sigs/gateway-api/pull/2113/files#r1231594914,
and follow-up issue in https://github.com/kubernetes-sigs/gateway-api/issues/2147.

A. TargetRef has been changed to `LocalPolicyTargetReferenceWithSectionName`, wherein the `SectionName` field is
interpreted as a port name for a Service.

## Graduation Criteria

This section is to record issues that were requested for discussion in the API section before this GEP graduates
out of `Provisional` status.

1. Michael Pleshakov asked about conflicts that could arise when multiple implementations are running in a cluster.
This is a gap in our policy attachment model that needs to be addressed.  See the discussion in
https://github.com/kubernetes-sigs/gateway-api/pull/2113/files#r1235750540. Graduating this GEP to implementable
requires an update to the Policy GEP to define how status can be nested to support multiple implementations. This will
likely look very similar to Route status.

>This question has been converted to a Gateway API enhancement request for Policy:
https://github.com/kubernetes-sigs/gateway-api/issues/4098.

2. Rob Scott [wanted to note](https://github.com/kubernetes-sigs/gateway-api/pull/2113#issuecomment-1696127092) that
when this graduates to the standard channel, implementations of HTTPRoute may also be
required to watch the BackendTLSPolicy. If one of these policies is attached to a Service targeted by an HTTPRoute,
the implementation would be required to fully implement the policy or mark the backend invalid.

>This comment may be added to the release notes for v1.4.0 of Gateway API, along with other special notes for the introduction
of the first standard Policy type.

## References

[Gateway API TLS Use Cases](https://docs.google.com/document/d/17sctu2uMJtHmJTGtBi_awGB0YzoCLodtR6rUNmKMCs8/edit#heading=h.cxuq8vo8pcxm)

[GEP-713: Metaresources and PolicyAttachment](../gep-713/index.md)

[Gateway API TLS](../../guides/tls.md)

[SIG-NET Gateway API: TLS to the K8s.Service/Backend](https://docs.google.com/document/d/1RTYh2brg_vLX9o3pTcrWxtZSsf8Y5NQvIG52lpFcZlo)

[SAN vs SNI](https://serverfault.com/questions/807959/what-is-the-difference-between-san-and-sni-ssl-certificates)
