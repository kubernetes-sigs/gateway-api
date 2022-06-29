# Changelog

## Table of Contents

- [v0.5.0-rc1](#v050-rc1)
- [v0.4.3](#v043)
- [v0.4.2](#v042)
- [v0.4.1](#v041)
- [v0.4.0](#v040)
- [v0.4.0-rc2](#v040-rc2)
- [v0.4.0-rc1](#v040-rc1)
- [v0.3.0](#v030)
- [v0.2.0](#v020)
- [v0.1.0](#v010)
- [v0.1.0-rc2](#v010-rc2)
- [v0.1.0-rc1](#v010-rc1)

## v0.5.0-rc1

API versions: v1beta1, v1alpha2

This release is all about stability.

Changes in this release can largely be divided into the following categories:

- Release Channels
- Resources graduating to beta
- New experimental features
- Bug Fixes
- General Improvements
- Breaking Changes
  - Validation improvements
  - Internal type cleanup

### Release channels

In this release, we've made two release channels available, `experimental` and
`standard`.

The `experimental` channel contains all resources and fields, while `standard`
contains only resources that mave moved to beta status.

We've also added a way to flag particular fields within a resource as
experimental, and any fields marked in this way are only present in the
`experimental` channel. Please see the [versioning][vers] docs for a more
detailed explanation.

One caveat for the standard channel - due to work on the new ReferenceGrant
resource: conformance tests may not pass with the `standard` set of CRDs.

[vers]:https://gateway-api.sigs.k8s.io/concepts/versioning/

### Resources Graduating to BETA

The following APIs have been promoted to a `v1beta1` maturity:

- `GatewayClass`
- `Gateway`
- `HTTPRoute`

[#1192](https://github.com/kubernetes-sigs/gateway-api/pull/1192)

### New Experimental Features

- Routes can now select `Gateway` listeners by port number
  [#1002](https://github.com/kubernetes-sigs/gateway-api/pull/1002)
- Gateway API now includes "Experimental" release channel. Consequently, CRDs now
  include `gateway.networking.k8s.io/bundle-version` and
  `gateway.networking.k8s.io/channel` annotations.
  [#945](https://github.com/kubernetes-sigs/gateway-api/pull/945)
- URL Rewrites and Path redirects have been added as new "Experimental" features
  [#945](https://github.com/kubernetes-sigs/gateway-api/pull/945)

### Bug Fixes

- Fixes a problem that would cause webhook deployment to fail on Kubernetes
  v1.22 and greater.
  [#991](https://github.com/kubernetes-sigs/gateway-api/pull/991)
- Fixes a bug where the `Namespace` could be unspecified in `ReferencePolicy`
  [#964](https://github.com/kubernetes-sigs/gateway-api/pull/964)
- Fixes a bug where v1alpha2 GatewayClass controller names were not being
  shown in the output of `kubectl get gatewayclasses`
  [#909](https://github.com/kubernetes-sigs/gateway-api/pull/909)

### General Improvements

- Conformance tests were introduced with [GEP-917][gep-917] and multiple
  conformance tests were added from a variety of contributors under the
  `conformance/` directory.
- The status of the GatewayClass "Accepted" condition for the `GatewayClass`
  is now present in `kubectl get` output.
  [#1168](https://github.com/kubernetes-sigs/gateway-api/pull/1168)
- New `RouteConditionReason` types `RouteReasonNotAllowedByListeners` and
  `RouteReasonNoMatchingListenerHostname` were added.
  [#1155](https://github.com/kubernetes-sigs/gateway-api/pull/1155)
- New `RouteConditionReason` type added with `RouteReasonAccepted`,
  `RouteReasonResolvedRefs` and `RouteReasonRefNotPermitted` constants.
  [#1114](https://github.com/kubernetes-sigs/gateway-api/pull/1114)
- Introduced PreciseHostname which prevents wildcard characters in relevant
  Hostname values.
  [#956](https://github.com/kubernetes-sigs/gateway-api/pull/956)

[gep-917]:https://gateway-api.sigs.k8s.io/geps/gep-917/

### Validation Improvements

- Webhook validation now ensures that a path match exists when required by path
  modifier in filter.
  [#1171](https://github.com/kubernetes-sigs/gateway-api/pull/1171)
- Webhook validation was added to ensure that only type-appropriate fields are
  set in `HTTPPathModifier`.
  [#1124](https://github.com/kubernetes-sigs/gateway-api/pull/1124)
- The Gateway API webhook is now deployed in a `gateway-system` namespace
  instead of `gateway-api`.
  [#1051](https://github.com/kubernetes-sigs/gateway-api/pull/1051)

### Breaking Changes

- The v1alpha1 API version was deprecated and removed.
  [#1197](https://github.com/kubernetes-sigs/gateway-api/pull/1197)
  [#906](https://github.com/kubernetes-sigs/gateway-api/issues/906)
- The `NamedAddress` value for `Gateway`'s `spec.addresses[].type` field has
  been deprecated, and support for domain-prefixed values (like
  `example.com/NamedAddress`) has been added instead to better represent the
  custom nature of this support.
  [#1178](https://github.com/kubernetes-sigs/gateway-api/pull/1178)
- Implementations are now expected to use `500` instead of `503` responses when
  the data-plane has no matching route.
  [#1151](https://github.com/kubernetes-sigs/gateway-api/pull/1151)

#### UX and Status Improvements

The following are **breaking changes** related to status updates and end-user
experience changes.

- The `UnsupportedExtension` named `ListenerConditionReason` has been removed.
  [#1146](https://github.com/kubernetes-sigs/gateway-api/pull/1146)
- The `RouteConflict` named `ListenerConditionReason` has been removed.
  [#1145](https://github.com/kubernetes-sigs/gateway-api/pull/1145)

#### Internal Type Cleanup

These changes will only affect implementations. Implementors will need to adjust
for the type changes when updating the Gateway API dependency in their projects.

**NOTE**: These kinds of changes are not always present in the CHANGELOG so
          please be aware that the CHANGELOG is not an exhaustive list of Go
          type changes. In this case there were a significant number of changes
          in a single release, so we included them for extra visibility for
          implementors.

- `ReferencePolicy` has been renamed to `ReferenceGrant`.
  [#1179](https://github.com/kubernetes-sigs/gateway-api/pull/1179)
- `GatewayTLSConfig`'s `CertificateRefs` field is now a slice of pointers to
  structs instead of the structs directly.
  [#1176](https://github.com/kubernetes-sigs/gateway-api/pull/1176)
- `HTTPPathModifer` field `Absolute` renamed to `ReplaceFullPath`
  [#1124](https://github.com/kubernetes-sigs/gateway-api/pull/1124)
- the `ParentRef` type was renamed to `ParentReference`
  [#982](https://github.com/kubernetes-sigs/gateway-api/pull/982)
- Types `ConditionRouteAccepted` and `ConditionRouteResolvedRefs` are now
  deprecated in favor of `RouteConditionAccepted` & `RouteConditionResolvedRefs`
  [#1114](https://github.com/kubernetes-sigs/gateway-api/pull/1114)

## v0.4.3

API version: v1alpha2

This release includes improvements to our webhook, including:

* Migrating kube-webhook-certgen to k8s.gcr.io/ingress-nginx:v1.1.1.
  [#1126](https://github.com/kubernetes-sigs/gateway-api/pull/1126)
* New validation to ensure that a HTTPRouterFilter Type matches its value
  [#1071](https://github.com/kubernetes-sigs/gateway-api/pull/1071)
* A fix to ensure that Path match validation actually works
  [#1071](https://github.com/kubernetes-sigs/gateway-api/pull/1071)

## v0.4.2

API version: v1alpha2

This release is intended to verify our webhook image tagging process.

### Bug Fixes

* Update image generation process with more consistent naming
  [#1034](https://github.com/kubernetes-sigs/gateway-api/pull/1034)

## v0.4.1

API version: v1alpha2

This release contains minor bug fixes for v1alpha2.

### Bug Fixes

* ControllerName now prints correctly in kubectl output for GatewayClass
  [#909](https://github.com/kubernetes-sigs/gateway-api/pull/909)
* Namespace can no longer be left unspecified in ReferencePolicy
  [#964](https://github.com/kubernetes-sigs/gateway-api/pull/964)
* Wildcard characters can no longer be used in redirect Hostname values
  [#956](https://github.com/kubernetes-sigs/gateway-api/pull/956)

## v0.4.0

API version: v1alpha2

This release contains significant breaking changes as we strive for a concise
API. We anticipate that this API will be very similar to a future v1beta1
release.

The following changes have been made since v0.3.0:

### Major Changes

* The Gateway API APIGroup has moved from `networking.x-k8s.io` to
  `gateway.networking.k8s.io`. This means that, as far as the apiserver is
  concerned, this version is wholly distinct from v1alpha1, and automatic
  conversion is not possible. As part of this process, Gateway API is now
  subject to Kubernetes API review, the same as changes made to core API
  resources. More details in
  [#780](https://github.com/kubernetes-sigs/gateway-api/pull/780) and
  [#716](https://github.com/kubernetes-sigs/gateway-api/issues/716).

* Gateway-Route binding changes ([GEP-724](https://gateway-api.sigs.k8s.io/geps/gep-724/)):
  In v1alpha1, Gateways chose which Routes were attached using a combination of
  object and namespace selectors, with the option of also specifying object
  names. This resulted in a very complex config, that's easy to misinterpret. As
  part of v1alpha2, we're changing to:
  * Gateways *may* specify what kind of Routes they support (defaults to same
    protocol if not specified), and where those Routes can be (defaults to same
    namespace).
  * Routes *must* directly reference the Gateways the want to attach to, this is
    a list, so a Route can attach to more than one Gateway.
  * The Route becomes attached only when the specifications intersect.

  We believe this is quite a bit easier to understand, and still gives good
  flexibility for most use cases.
  GEP added in [#725](https://github.com/kubernetes-sigs/gateway-api/pull/725).
  Implemented in [#754](https://github.com/kubernetes-sigs/gateway-api/pull/754).
  Further documentation was added in [#762](https://github.com/kubernetes-sigs/gateway-api/pull/762).

* Safer cross-namespace references ([GEP-709](https://gateway-api.sigs.k8s.io/geps/gep-709/)): 
  This concerns (currently), references from Routes to Backends, and Gateways to
  Secrets. The new behavior is:
  * By default, references across namespaces are not permitted; creating a
    reference across a namespace (like a Route referencing a Service in another
    namespace) must be rejected by implementations.
  * These references can be accepted by creating a ReferencePolicy in the
    referent (target) namespace, that specifies what Kind is allowed to accept
    incoming references, and from what namespace and Kind the references may be.

  The intent here is that the owner of the referent namespace must explicitly
  accept incoming references, otherwise we can run into all sorts of bad things
  from breaking the namespace security model.
  Implemented in [#741](https://github.com/kubernetes-sigs/gateway-api/pull/741).

* Attaching Policy to objects ([GEP-713](https://gateway-api.sigs.k8s.io/geps/gep-713/)):
  This has been added so that we have an extensible mechanism for adding a
  cascading set of policy to Gateway API objects.

  What policy? Well, it's kind of up to the implementations, but the best example
  to begin with is timeout policy.

  Timeout policy for HTTP connections is highly depedent on how the underlying
  implementation handles policy - it's very difficult to extract commonalities.

  This is intended to allow things like:
  * Attach a policy that specifies the default connection timeout for backends
    to a GatewayClass. All Gateways that are part of that Class will have Routes
    get that default connection timeout unless they specify differently.
  * If a Gateway that's a member of the GatewayClass has a different default
    attached, then that will beat the GatewayClass (for defaults, more specific
    object beats less specific object).
  * Alternatively, a Policy that mandates that you can't set the client timeout
    to "no timeout" can be attached to a GatewayClass as an override. An
    override will always take effect, with less specific beating more specific.

  This one is a bit complex, but will allow implementations to solve some things
  that currently require tools like admission control.
  Implemented in [#736](https://github.com/kubernetes-sigs/gateway-api/pull/736).

* As part of GEP-713, `BackendPolicy` has been removed, as its functionality is
  now better handled using that mechanism.
  [#732](https://github.com/kubernetes-sigs/gateway-api/pull/732).

* Removal of certificate references from HTTPRoutes ([GEP-746](https://gateway-api.sigs.k8s.io/geps/gep-746/)):
  In v1alpha1, HTTPRoute objects have a stanza that allows referencing a TLS
  keypair, intended to allow people to have a more self-service model, where an
  app owner can provision a TLS keypair inside their own namespace, attach it to
  a HTTPRoute they control, and then have that used to secure their app.
  When implementing this, however, there are a large number of edge cases that
  are complex, hard to handle, and poorly defined - about checking SNI, hostname,
  and overrides, that made even writing a spec on how to implement this very
  difficult, let alone actually implementing it.

  In removing certificate references from HTTPRoute, we're using the
  ReferencePolicy from GEP-709 to allow Gateways to securely create a
  cross-namespace reference to TLS keypairs in app namespaces.
  We're hopeful that this will hit most of the self-service use case, and even
  if not, provide a basis to build from to meet it eventually.
  GEP added in [#749](https://github.com/kubernetes-sigs/gateway-api/pull/749).
  Implemented in [#768](https://github.com/kubernetes-sigs/gateway-api/pull/768).

  [GEP-851](https://github.com/kubernetes-sigs/gateway-api/blob/main/site-src/geps/gep-851.md)
  was a follow up on this change that allowed multiple Certificate Refs per
  Gateway Listener. This was implemented in
  [#852](https://github.com/kubernetes-sigs/gateway-api/pull/852).

* The `RouteForwardTo` (YAML: `routeForwardTo`) struct/stanza has been reworked
  into the `BackendRef` (YAML: `backendRef`) struct/stanza,
  [GEP-718](https://gateway-api.sigs.k8s.io/geps/gep-718/). As part of this
  change, the `ServiceName` (YAML: `serviceName`) field has been removed, and
  Service references must instead now use the `BackendRef`/`backendRef`
  struct/stanza.

### Small Changes
* Extension points within match blocks from all Routes have been removed
  [#829](https://github.com/kubernetes-sigs/gateway-api/pull/829). Implements
  [GEP-820](https://github.com/kubernetes-sigs/gateway-api/blob/main/site-src/geps/gep-820.md).
  These extension points have been removed because they are currently not used,
  are poorly understood, and we don't have good use cases for them. We may
  consider re-adding them in the future.

* Controller is now a required field in Gateway references from Route status.
  [#671](https://github.com/kubernetes-sigs/gateway-api/pull/671).

* Header Matching, Query Param Matching, and HTTPRequestHeaderFilter now use
  named subobjects instead of maps.
  [#657](https://github.com/kubernetes-sigs/gateway-api/pull/657) and
  [#681](https://github.com/kubernetes-sigs/gateway-api/pull/681)

* [#796](https://github.com/kubernetes-sigs/gateway-api/pull/796) API Review
  suggestions:
  * listener.routes has been renamed to listener.allowedRoutes
  * The `NoSuchGatewayClass` has been removed after it was deprecated in
    v1alpha1
  * `*` is no longer a valid hostname. Instead, leaving hostname unspecified is
    interpreted as `*`.

* The `scope` field has been removed from all object references.
  [#882](https://github.com/kubernetes-sigs/gateway-api/pull/882)

* "Controller" has been renamed to "ControllerName"
  [#839](https://github.com/kubernetes-sigs/gateway-api/pull/839)

* "Admitted" condition has been renamed to "Accepted" and now defaults to an
  "Unknown" state instead of "False"
  [#839](https://github.com/kubernetes-sigs/gateway-api/pull/839)

* HTTPRequestRedirectFilter's Protocol field has been renamed to Scheme.
  [#863](https://github.com/kubernetes-sigs/gateway-api/pull/863)

* ImplementationSpecific match types in HTTPRoute's path, query, and header
  matches have been removed.
  [#850](https://github.com/kubernetes-sigs/gateway-api/pull/850)

* The "Prefix" path match type has been renamed "PathPrefix".
  [#898](https://github.com/kubernetes-sigs/gateway-api/pull/898)

### Small Additions
* HTTP Method matching is now added into HTTPRoute, with Extended support:
  [#733](https://github.com/kubernetes-sigs/gateway-api/pull/733).

* GatewayClass now has a 'Description' field that is printed as a column in
  `kubectl get` output. You can now end up with output that looks like this:
  ```shell
  $> kubectl get gatewayclass
  NAME       CONTROLLER                            DESCRIPTION
  internal   gateway-controller-internal   For non-internet-facing Gateways.
  external   gateway-controller-external   For internet-facing Gateways.
  ```
  See [#610](https://github.com/kubernetes-sigs/gateway-api/issues/610) and
  [#653](https://github.com/kubernetes-sigs/gateway-api/pull/653) for the
  details.

### Validation changes
* Ensure TLSConfig is empty when the protocol is HTTP, TCP, or UDP
  [#886](https://github.com/kubernetes-sigs/gateway-api/pull/886)
* Ensure Hostname is empty when the protocol is TCP or UDP.
  [#886](https://github.com/kubernetes-sigs/gateway-api/pull/886)
* Listener ProtocolType now has validation.
  [#871](https://github.com/kubernetes-sigs/gateway-api/pull/871)
* HTTP Path match values are now validated for PathMatchExact and
  PathMatchPrefix match types.
  [#894](https://github.com/kubernetes-sigs/gateway-api/pull/894)
* TLS options keys are now subject to the same validation as Kubernetes
  annotations. [#886](https://github.com/kubernetes-sigs/gateway-api/pull/886)
* TLS options values now have a max length of 4096 characters.
  [#886](https://github.com/kubernetes-sigs/gateway-api/pull/886)
* Make `MirrorFilter.BackendRef` a required field when the mirror filter is used
  [#837](https://github.com/kubernetes-sigs/gateway-api/pull/837).

### Clarifications
* Updated guidance on how HTTP and TLS Route status should be populated when
  hostnames do not match.
  [#859](https://github.com/kubernetes-sigs/gateway-api/pull/859)
* Aligned path prefix matching with Ingress by clarifying that it is a prefix of
  path elements. [#869](https://github.com/kubernetes-sigs/gateway-api/pull/869)
* HTTP listeners may now be used for Cleartext HTTP/2.
  [#879](https://github.com/kubernetes-sigs/gateway-api/pull/879)
* Added clarification that implementation-specific TLS options MUST be
  domain-prefixed.
  [#899](https://github.com/kubernetes-sigs/gateway-api/pull/899)

### Documentation Updates
* [#782](https://github.com/kubernetes-sigs/gateway-api/pull/782) : Restructure docs and split into versioned and unversioned
* [#777](https://github.com/kubernetes-sigs/gateway-api/pull/777) : Fix typo
* [#765](https://github.com/kubernetes-sigs/gateway-api/pull/765) : document multi-value headers as undefined
* [#761](https://github.com/kubernetes-sigs/gateway-api/pull/761) : minor improvements to navigation on docs site
* [#760](https://github.com/kubernetes-sigs/gateway-api/pull/760) : Remove references of vendor configurations in GatewayTLSConfig
* [#756](https://github.com/kubernetes-sigs/gateway-api/pull/756) : Clarify docs on invalid serviceName
* [#755](https://github.com/kubernetes-sigs/gateway-api/pull/755) : Document the supported kubernetes versions
* [#745](https://github.com/kubernetes-sigs/gateway-api/pull/745) : Remove RouteTLSConfig requirement for gateway TLS passthrough.
* [#744](https://github.com/kubernetes-sigs/gateway-api/pull/744) : automate nav for GEPs
* [#743](https://github.com/kubernetes-sigs/gateway-api/pull/743) : Add READY and ADDRESS to gateway printer columns
* [#742](https://github.com/kubernetes-sigs/gateway-api/pull/742) : Moving method match to v1alpha2 example
* [#729](https://github.com/kubernetes-sigs/gateway-api/pull/729) : Adding suggested reasons for when conditions are healthy
* [#728](https://github.com/kubernetes-sigs/gateway-api/pull/728) : Fixing wording in enhancement template
* [#723](https://github.com/kubernetes-sigs/gateway-api/pull/723) : Clarifying Redirect Support levels
* [#756](https://github.com/kubernetes-sigs/gateway-api/pull/756) : Clarify docs on invalid serviceName
* [#880](https://github.com/kubernetes-sigs/gateway-api/pull/880) : Reworking Policy vs. Filter Documentation
* [#878](https://github.com/kubernetes-sigs/gateway-api/pull/878) : Clarifying the fields that all Route types must include
* [#875](https://github.com/kubernetes-sigs/gateway-api/pull/875) : Fix HTTP path match documentation.
* [#864](https://github.com/kubernetes-sigs/gateway-api/pull/864) : Merging v1alpha2 concepts docs into unversioned docs
* [#858](https://github.com/kubernetes-sigs/gateway-api/pull/858) : Fixing broken link to spec page
* [#857](https://github.com/kubernetes-sigs/gateway-api/pull/857) : Adding missing references pages to docs navigation
* [#853](https://github.com/kubernetes-sigs/gateway-api/pull/853) : docs: Use v0.4.0-rc1 in "Getting started with Gateway APIs" for v1alpha2
* [#845](https://github.com/kubernetes-sigs/gateway-api/pull/845) : Fix markdown list formatting.
* [#844](https://github.com/kubernetes-sigs/gateway-api/pull/844) : docs: add ssl passthrough note in FAQ
* [#843](https://github.com/kubernetes-sigs/gateway-api/pull/843) : Add APISIX implementation
* [#834](https://github.com/kubernetes-sigs/gateway-api/pull/834) : Fixes some broken links
* [#807](https://github.com/kubernetes-sigs/gateway-api/pull/807) : docs: update multiple-ns guide for v1alpha2
* [#888](https://github.com/kubernetes-sigs/gateway-api/pull/888) : Corrected broken getting started
* [#885](https://github.com/kubernetes-sigs/gateway-api/pull/885) : Fix incorrect urls
* [#890](https://github.com/kubernetes-sigs/gateway-api/pull/890) : Updating HTTPRoute docs for v1alpha2
* [#870](https://github.com/kubernetes-sigs/gateway-api/pull/870) : Adding guidance on Kind vs. Resource in implementation guidelines
* [#865](https://github.com/kubernetes-sigs/gateway-api/pull/865) : Route cleanup for v1alpha2 sig-network review

### Tooling and infra updates
* [#766](https://github.com/kubernetes-sigs/gateway-api/pull/766) : comment out the GEP notice
* [#758](https://github.com/kubernetes-sigs/gateway-api/pull/758) : bump up mkdocs and deps
* [#751](https://github.com/kubernetes-sigs/gateway-api/pull/751) : bump up deps to k8s v1.22
* [#748](https://github.com/kubernetes-sigs/gateway-api/pull/748) : fix kustomize to install v1a2 crds
* [#747](https://github.com/kubernetes-sigs/gateway-api/pull/747) : Cleaning up GEP Template
* [#889](https://github.com/kubernetes-sigs/gateway-api/pull/889) : remove outdated version label
* [#883](https://github.com/kubernetes-sigs/gateway-api/pull/883) : validating webhook cleanup
* [#872](https://github.com/kubernetes-sigs/gateway-api/pull/872) : Remove duplicate validation from CRD & Webhook

## v0.4.0-rc2

API version: v1alpha2

The group expects that this release candidate has no changes before we release
v1alpha2 final, but are cutting here to allow implementations a chance to check
before we go to the final release.

In general, most of the changes below have been made to reduce the complexity of
the API for v1alpha2, on the assumption that we can add functionality in later
in the API's lifecycle, but cannot remove it.

The following changes have been made since v0.4.0-rc1:

### GEP implementations
* Replace `CertificateRef` field with `CertificateRefs` in `GatewayTLSConfig`.
[#852](https://github.com/kubernetes-sigs/gateway-api/pull/852). This implements
[GEP-851](https://github.com/kubernetes-sigs/gateway-api/blob/main/site-src/geps/gep-851.md),
Allow Multiple Certificate Refs per Gateway Listener.
* Extension points within match blocks from all Routes have been removed
[#829](https://github.com/kubernetes-sigs/gateway-api/pull/829). Implements
[GEP-820](https://github.com/kubernetes-sigs/gateway-api/blob/main/site-src/geps/gep-820.md).
These extension points have been removed because they are currently not used,
are poorly understood, and we don't have good use cases for them. We may
consider re-adding them in the future.

### Field changes
* Make `MirrorFilter.BackendRef` a required field when the mirror filter is used
[#837](https://github.com/kubernetes-sigs/gateway-api/pull/837).
* ImplementationSpecific match types in HTTPRoute's path, query, and header
matches have been removed.
[#850](https://github.com/kubernetes-sigs/gateway-api/pull/850)
* The "Prefix" path match type has been renamed "PathPrefix".
* The "ClassName" field in PolicyTargetReference has been removed.
* A new optional "Name" field has been added to ReferencePolicyTo.
[#898](https://github.com/kubernetes-sigs/gateway-api/pull/898)

### Field Renames
* "Controller" has been renamed to "ControllerName"
* "Admitted" condition has been renamed to "Accepted" and now defaults to an
"Unknown" state instead of "False" [#839](https://github.com/kubernetes-sigs/gateway-api/pull/839)
* HTTPRequestRedirectFilter's Protocol field has been renamed to Scheme.
[#863](https://github.com/kubernetes-sigs/gateway-api/pull/863)


### Validation changes
*  Validation: Ensure TLSConfig is empty when the protocol is HTTP, TCP, or UDP
[#886](https://github.com/kubernetes-sigs/gateway-api/pull/886)
*  Validation: Ensure Hostname is empty when the protocol is TCP or UDP.
[#886](https://github.com/kubernetes-sigs/gateway-api/pull/886)
* Validation: Listener ProtocolType now has validation.
[#871](https://github.com/kubernetes-sigs/gateway-api/pull/871)
* Validation: HTTP Path match values are now validated for PathMatchExact and
PathMatchPrefix match types. [#894](https://github.com/kubernetes-sigs/gateway-api/pull/894)

### Documentation and specification updates
* Updated guidance on how HTTP and TLS Route status should be populated when
hostnames do not match.
[#859](https://github.com/kubernetes-sigs/gateway-api/pull/859)
* Aligned path prefix matching with Ingress by clarifying that it is a prefix of
path elements. [#869](https://github.com/kubernetes-sigs/gateway-api/pull/869)
* HTTP listeners may now be used for Cleartext HTTP/2.
[#879](https://github.com/kubernetes-sigs/gateway-api/pull/879)
* The `scope` field has been removed from all object references.
* ParentRefs can no longer refer to cluster-scoped resources.
[#882](https://github.com/kubernetes-sigs/gateway-api/pull/882)
* TLS options keys are now subject to the same validation as Kubernetes
annotations. [#886](https://github.com/kubernetes-sigs/gateway-api/pull/886)
* TLS options values now have a max length of 4096 characters.
[#886](https://github.com/kubernetes-sigs/gateway-api/pull/886)
* Added clarification that implementation-specific TLS options MUST be domain-prefixed.
[#899](https://github.com/kubernetes-sigs/gateway-api/pull/899)

### Other changes
* [#890](https://github.com/kubernetes-sigs/gateway-api/pull/890) : Updating HTTPRoute docs for v1alpha2
* [#889](https://github.com/kubernetes-sigs/gateway-api/pull/889) : remove outdated version label
* [#888](https://github.com/kubernetes-sigs/gateway-api/pull/888) : Corrected broken getting started
* [#885](https://github.com/kubernetes-sigs/gateway-api/pull/885) : Fix incorrect urls
* [#883](https://github.com/kubernetes-sigs/gateway-api/pull/883) : v1alpha2 validation fix/update
* [#880](https://github.com/kubernetes-sigs/gateway-api/pull/880) : Reworking Policy vs. Filter Documentation
* [#878](https://github.com/kubernetes-sigs/gateway-api/pull/878) : Clarifying the fields that all Route types must include
* [#875](https://github.com/kubernetes-sigs/gateway-api/pull/875) : Fix HTTP path match documentation.
* [#872](https://github.com/kubernetes-sigs/gateway-api/pull/872) : Remove duplicate validation from CRD & Webhook
* [#870](https://github.com/kubernetes-sigs/gateway-api/pull/870) : Adding guidance on Kind vs. Resource in implementation guidelines
* [#865](https://github.com/kubernetes-sigs/gateway-api/pull/865) : Route cleanup for v1alpha2 sig-network review
* [#864](https://github.com/kubernetes-sigs/gateway-api/pull/864) : Merging v1alpha2 concepts docs into unversioned docs
* [#858](https://github.com/kubernetes-sigs/gateway-api/pull/858) : Fixing broken link to spec page
* [#857](https://github.com/kubernetes-sigs/gateway-api/pull/857) : Adding missing references pages to docs navigation
* [#853](https://github.com/kubernetes-sigs/gateway-api/pull/853) : docs: Use v0.4.0-rc1 in "Getting started with Gateway APIs" for v1alpha2
* [#845](https://github.com/kubernetes-sigs/gateway-api/pull/845) : Fix markdown list formatting.
* [#844](https://github.com/kubernetes-sigs/gateway-api/pull/844) : docs: add ssl passthrough note in FAQ
* [#843](https://github.com/kubernetes-sigs/gateway-api/pull/843) : Add APISIX implementation
* [#834](https://github.com/kubernetes-sigs/gateway-api/pull/834) : Fixes some broken links
* [#807](https://github.com/kubernetes-sigs/gateway-api/pull/807) : docs: update multiple-ns guide for v1alpha2


## v0.4.0-rc1

API version: v1alpha2

The working group expects that this release candidate is quite close to the final
v1alpha2 API. However, breaking API changes are still possible.

This release candidate is suitable for implementors, but the working group does
not recommend shipping products based on a release candidate API due to the
possibility of incompatible changes prior to the final release.

### Major Changes

* The Gateway API APIGroup has moved from `networking.x-k8s.io` to
`gateway.networking.k8s.io`. This means that, as far as the apiserver is
concerned, this version is wholly distinct from v1alpha1, and automatic conversion
is not possible. As part of this process, Gateway API is now subject to Kubernetes
API review, the same as changes made to core API resources. More details in
[#780](https://github.com/kubernetes-sigs/gateway-api/pull/780) and [#716](https://github.com/kubernetes-sigs/gateway-api/issues/716).

* Gateway-Route binding changes:
[GEP-724](https://gateway-api.sigs.k8s.io/geps/gep-724/). Currently, Gateways
choose which Routes are attached using a combination of object and namespace
selectors, with the option of also specifying object names. This has made a very
complex config, that's easy to misinterpret. As part of v1alpha2, we're changing to:
  * Gateways *may* specify what kind of Routes they support (defaults to same
  protocol if not specified), and where those Routes can be (defaults to same
  namespace).
  * Routes *must* directly reference the Gateways the want to attach to, this is
  a list, so a Route can attach to more than one Gateway.
  * The Route becomes attached only when the specifications intersect.

  We believe this is quite a bit easier to understand, and still gives good
  flexibility for most use cases.
  GEP added in [#725](https://github.com/kubernetes-sigs/gateway-api/pull/725).
  Implemented in [#754](https://github.com/kubernetes-sigs/gateway-api/pull/754).
  Further documentation was added in [#762](https://github.com/kubernetes-sigs/gateway-api/pull/762).


* Safer cross-namespace references:
([GEP-709](https://gateway-api.sigs.k8s.io/geps/gep-709/)): This concerns
(currently), references from Routes to Backends, and Gateways to Secrets. The
new behavior is:
  * By default, references across namespaces are not permitted; creating a
  reference across a namespace (like a Route referencing a Service in another
  namespace) must be rejected by implementations.
  * These references can be accepted by creating a ReferencePolicy in the
  referent (target) namespace, that specifies what Kind is allowed to accept
  incoming references, and from what namespace and Kind the references may be.

  The intent here is that the owner of the referent namespace must explicitly
  accept incoming references, otherwise we can run into all sorts of bad things
  from breaking the namespace security model.
  Implemented in [#741](https://github.com/kubernetes-sigs/gateway-api/pull/741).

* Attaching Policy to objects:
[GEP-713](https://gateway-api.sigs.k8s.io/geps/gep-713/): This has been added
so that we have an extensible mechanism for adding a cascading set of policy to
Gateway API objects.

  What policy? Well, it's kind of up to the implementations, but the best example
  to begin with is timeout policy.

  Timeout policy for HTTP connections is highly depedent on how the underlying
  implementation handles policy - it's very difficult to extract commonalities.

  This is intended to allow things like:
  * Attach a policy that specifies the default connection timeout for backends
  to a GatewayClass. All Gateways that are part of that Class will have Routes
  get that default connection timeout unless they specify differently.
  * If a Gateway that's a member of the GatewayClass has a different default
  attached, then that will beat the GatewayClass (for defaults, more specific
  object beats less specific object).
  * Alternatively, a Policy that mandates that you can't set the client timeout
  to "no timeout" can be attached to a GatewayClass as an override. An override
  will always take effect, with less specific beating more specific.

  This one is a bit complex, but will allow implementations to solve some things
  that currently require tools like admission control.
  Implemented in [#736](https://github.com/kubernetes-sigs/gateway-api/pull/736).

* As part of GEP-713, `BackendPolicy` has been removed, as its functionality is
now better handled using that mechanism. [#732](https://github.com/kubernetes-sigs/gateway-api/pull/732).

* Removal of certificate references from HTTPRoutes:
[GEP-746](https://gateway-api.sigs.k8s.io/geps/gep-746/):
  In v1alpha1, HTTPRoute objects have a stanza that allows referencing a TLS
  keypair, intended to allow people to have a more self-service model, where an
  app owner can provision a TLS keypair inside their own namespace, attach it to
  a HTTPRoute they control, and then have that used to secure their app.
  When implementing this, however, there are a large number of edge cases that
  are complex, hard to handle, and poorly defined - about checking SNI, hostname,
  and overrides, that made even writing a spec on how to implement this very
  difficult, let alone actually implementing it.

  In removing certificate references from HTTPRoute, we're using the
  ReferencePolicy from GEP-709 to allow Gateways to securely create a
  cross-namespace reference to TLS keypairs in app namespaces.
  We're hopeful that this will hit most of the self-service use case, and even
  if not, provide a basis to build from to meet it eventually.
  GEP added in [#749](https://github.com/kubernetes-sigs/gateway-api/pull/749).
  Implemented in [#768](https://github.com/kubernetes-sigs/gateway-api/pull/768).

* The `RouteForwardTo` (YAML: `routeForwardTo`) struct/stanza has been reworked
into the `BackendRef` (YAML: `backendRef`) struct/stanza,
[GEP-718](https://gateway-api.sigs.k8s.io/geps/gep-718/). As part of this change,
the `ServiceName` (YAML: `serviceName`) field has been removed, and Service
references must instead now use the `BackendRef`/`backendRef` struct/stanza.

### Other changes
* HTTP Method matching is now added into HTTPRoute, with Extended support:
[#733](https://github.com/kubernetes-sigs/gateway-api/pull/733).

* GatewayClass now has a 'Description' field that is printed as a column in
`kubectl get` output. You can now end up with output that looks like this:
  ```shell
  $> kubectl get gatewayclass
  NAME       CONTROLLER                            DESCRIPTION
  internal   gateway-controller-internal   For non-internet-facing Gateways.
  external   gateway-controller-external   For internet-facing Gateways.
  ```
  See [#610](https://github.com/kubernetes-sigs/gateway-api/issues/610) and
  [#653](https://github.com/kubernetes-sigs/gateway-api/pull/653) for the details.

*  [#671](https://github.com/kubernetes-sigs/gateway-api/pull/671): Controller is
now a required field in Gateway references from Route status. Fixes
[#669](https://github.com/kubernetes-sigs/gateway-api/pull/671).

*  [#657](https://github.com/kubernetes-sigs/gateway-api/pull/657): and
[#681](https://github.com/kubernetes-sigs/gateway-api/pull/681) Header Matching,
Query Param Matching, and HTTPRequestHeaderFilter now use named subobjects
instead of maps.

* [#796](https://github.com/kubernetes-sigs/gateway-api/pull/796) API Review suggestions:
  * listener.routes has been renamed to listener.allowedRoutes
  * The `NoSuchGatewayClass` has been removed after it was deprecated in v1alpha1
  * `*` is no longer a valid hostname. Instead, leaving hostname unspecified is interpreted as `*`.

### Documentation Updates
* [#782](https://github.com/kubernetes-sigs/gateway-api/pull/782) : Restructure docs and split into versioned and unversioned
* [#777](https://github.com/kubernetes-sigs/gateway-api/pull/777) : Fix typo
* [#765](https://github.com/kubernetes-sigs/gateway-api/pull/765) : document multi-value headers as undefined
* [#761](https://github.com/kubernetes-sigs/gateway-api/pull/761) : minor improvements to navigation on docs site
* [#760](https://github.com/kubernetes-sigs/gateway-api/pull/760) : Remove references of vendor configurations in GatewayTLSConfig
* [#756](https://github.com/kubernetes-sigs/gateway-api/pull/756) : Clarify docs on invalid serviceName
* [#755](https://github.com/kubernetes-sigs/gateway-api/pull/755) : Document the supported kubernetes versions
* [#745](https://github.com/kubernetes-sigs/gateway-api/pull/745) : Remove RouteTLSConfig requirement for gateway TLS passthrough.
* [#744](https://github.com/kubernetes-sigs/gateway-api/pull/744) : automate nav for GEPs
* [#743](https://github.com/kubernetes-sigs/gateway-api/pull/743) : Add READY and ADDRESS to gateway printer columns
* [#742](https://github.com/kubernetes-sigs/gateway-api/pull/742) : Moving method match to v1alpha2 example
* [#729](https://github.com/kubernetes-sigs/gateway-api/pull/729) : Adding suggested reasons for when conditions are healthy
* [#728](https://github.com/kubernetes-sigs/gateway-api/pull/728) : Fixing wording in enhancement template
* [#723](https://github.com/kubernetes-sigs/gateway-api/pull/723) : Clarifying Redirect Support levels
* [#756](https://github.com/kubernetes-sigs/gateway-api/pull/756) : Clarify docs on invalid serviceName

### Tooling and infra updates
* [#766](https://github.com/kubernetes-sigs/gateway-api/pull/766) : comment out the GEP notice
* [#758](https://github.com/kubernetes-sigs/gateway-api/pull/758) : bump up mkdocs and deps
* [#751](https://github.com/kubernetes-sigs/gateway-api/pull/751) : bump up deps to k8s v1.22
* [#748](https://github.com/kubernetes-sigs/gateway-api/pull/748) : fix kustomize to install v1a2 crds
* [#747](https://github.com/kubernetes-sigs/gateway-api/pull/747) : Cleaning up GEP Template


## v0.3.0

API Version: v1alpha1

### API changes

#### Gateway
- The `NoSuchGatewayClass` status reason has been deprecated.
  [#635](https://github.com/kubernetes-sigs/gateway-api/pull/635)

#### HTTPRoute
- `.spec.rules.matches.path` now has a default `prefix` match on the `/` path.
  [#584](https://github.com/kubernetes-sigs/gateway-api/pull/584)
- Conflict resolution guidance has been added for rules within a route.
  [#620](https://github.com/kubernetes-sigs/gateway-api/pull/620)
- HTTPRoute now supports query param matching.
  [#631](https://github.com/kubernetes-sigs/gateway-api/pull/631)

#### All Route Types
- Route status now includes controller name for each Gateway.
  [#616](https://github.com/kubernetes-sigs/gateway-api/pull/616)
- Conflict resolution guidance has been added for non-HTTP routes.
  [#626](https://github.com/kubernetes-sigs/gateway-api/pull/626)

#### Misc
- Fields of type LocalObjectRef do not default to "secrets". All LocalObjectRef
  fields must be specified.
  [#570](https://github.com/kubernetes-sigs/gateway-api/pull/570)
- CRDs have been added to gateway-api category
  [#592](https://github.com/kubernetes-sigs/gateway-api/pull/592)
- New "Age" column has been added to all resources in `kubectl get` output.
  [#592](https://github.com/kubernetes-sigs/gateway-api/pull/592)
- A variety of Go types have been changed to pointers to better reflect their
  optional status.
  [#564](https://github.com/kubernetes-sigs/gateway-api/pull/564)
  [#572](https://github.com/kubernetes-sigs/gateway-api/pull/572)
  [#579](https://github.com/kubernetes-sigs/gateway-api/pull/579)

#### Validation
- A new experimental validation package and validating webhook have been added.
  [#597](https://github.com/kubernetes-sigs/gateway-api/pull/597)
  [#617](https://github.com/kubernetes-sigs/gateway-api/pull/617)


## v0.2.0

API Version: v1alpha1

### API changes

Service APIs has been renamed to Gateway API.
[#536](https://github.com/kubernetes-sigs/service-apis/issues/536).


#### GatewayClass
- The default status condition of GatewayClass resource is now `Admitted:false`
  instead of `InvalidParameters:Unknown`.
  [#471](https://github.com/kubernetes-sigs/service-apis/pull/471).
- `GatewayClass.spec.parametersRef` now has an optional `namespace` field to
  refer to a namespace-scoped resource in addition to cluster-scoped resource.
  [#543](https://github.com/kubernetes-sigs/service-apis/pull/543).

#### Gateway
- `spec.listeners[].tls.mode` now defaults to `Terminate`.
  [#518](https://github.com/kubernetes-sigs/service-apis/pull/518).
- Empty `hostname` in a listener matches all request.
  [#525](https://github.com/kubernetes-sigs/service-apis/pull/525).

#### HTTPRoute
- New `set` property has been introduced for `HTTPRequestHeader` Filter. Headers
  specified under `set` are overriden instead of added.
  [#475](https://github.com/kubernetes-sigs/service-apis/pull/475).

#### Misc
- Maximum limit for `forwardTo` has been increased from `4` to `16` for all
  route types.
  [#493](https://github.com/kubernetes-sigs/service-apis/pull/493).
- Various changes have been made in the Kubernetes and Go API to align with
  upstream Kubernetes API conventions. Some of the fields have been changed to
  pointers in the Go API for this reason.
  [#538](https://github.com/kubernetes-sigs/service-apis/pull/538).

### Documentation

There are minor improvements to docs all around.
New guides, clarifications and various typos have been fixed.

## v0.1.0

API Version: v1alpha1

### API changes since v0.1.0-rc2
#### GatewayClass
- CRD now includes `gc` short name.
- Change the standard condition for GatewayClass to `Admitted`, with
  `InvalidParameters` as a sample reason for it to be false.

#### Gateway
- CRD now includes `gtw` short name.
- The `DroppedRoutes` condition has been renamed to `DegradedRoutes`.
- `ListenerStatus` now includes `Protocol` and `Hostname` to uniquely link the
  status to each listener.

#### Routes
- HTTPRoute clarifications:
  - Header name matching must be case-insensitive.
  - Match tiebreaking semantics have been outlined in detail.
- TCPRoute, TLSRoute, and UDPRoute:
  - At least 1 ForwardTo must be specified in each rule.
  - Clarification that if no matches are specified, all requests should match a
    rule.
- TCPRoute and UDPRoute: Validation has been added to ensure that 1-16 rules are
  specified, matching other route types.
- TLSRoute: SNIs are now optional in matches. If no SNI or extensionRef are
  specified, all requests match.

#### BackendPolicy
- CRD now includes `bp` short name.
- A new `networking.x-k8s.io/app-protocol` annotation can be used to specify
  AppProtocol on Services when the AppProtocol field is unavailable.


## v0.1.0-rc2

API Version: v1alpha-rc2

### API changes since v0.1.0-rc1
#### GatewayClass
- A recommendation to set a `gateway-exists-finalizer.networking.x-k8s.io`
  finalizer on GatewayClass has been added.
- `allowedGatewayNamespaces` has been removed from GatewayClass in favor of
  implementations with policy agents like Gatekeeper.

#### Gateway
- Fields in `listeners.routes` have been renamed:
  - `routes.routeSelector` -> `routes.selector`
  - `routes.routeNamespaces`-> `routes.namespaces`
- `clientCertificateRef` has been removed from BackendPolicy.
- In Listeners, `routes.namespaces` now defaults to `{from: "Same"}`.
- In Listeners, support has been added for specifying custom, domain prefixed
  protocols.
- In Listeners, `hostname` now closely matches Route hostname matching with wildcard
  support.
- A new `UnsupportedAddress` condition has been added to Listeners to indicate
  that a requested address is not supported.
- Clarification has been added to note that listeners may be merged in certain
  instances.

#### Routes
- HeaderMatchType now includes a RegularExpression option.
- Minimum weight has been decreased from 1 to 0.
- Port is now required on all Routes.
- On HTTPRoute, filters have been renamed:
  - `ModifyRequestHeader` -> `RequestHeaderModifier`
  - `MirrorRequest` -> `RequestMirror`
  - `Custom` -> `ExtensionRef`
- TLSRoute can now specify as many as 16 SNIs instead of 10.
- Limiting the number of Gateways that may be stored in RouteGatewayStatus to
  100.
- Support level of filters defined in ForwardTo has been clarified.
- Max weight has been increased to 1 million.


## v0.1.0-rc1

API Version: v1alpha-rc1

- Initial release candidate for v1alpha1.
