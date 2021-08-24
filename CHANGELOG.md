# Changelog

## Table of Contents

- [v0.4.0-rc1](#v040-rc1)
- [v0.3.0](#v030)
- [v0.2.0](#v020)
- [v0.1.0](#v010)
- [v0.1.0-rc2](#v010-rc2)
- [v0.1.0-rc1](#v010-rc1)


## v0.4.0-rc1

API version: v1alpha2

This release is intended to be last alpha release, so there are a lot of breaking
API changes. Please read these release notes carefully.

### Major Changes

* The Gateway API APIGroup has moved from `networking.x-k8s.io` to
`gateway.networking.k8s.io`. This means that, as far as the apiserver is
concerned, this version is wholly distinct from v1alpha1, and automatic conversion
is not possible. As part of this process, Gateway API is now subject to Kubernetes
API review, the same as changes made to core API resources. More details in
[#780](https://github.com/kubernetes-sigs/gateway-api/pull/780) and [#716](https://github.com/kubernetes-sigs/gateway-api/issues/716).

* Gateway-Route binding changes,
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


* Safer cross-namespace references
([GEP-709](https://gateway-api.sigs.k8s.io/geps/gep-709/)): This one concerns
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

* Attaching Policy to objects,
[GEP-713](https://gateway-api.sigs.k8s.io/geps/gep-713/): This one has been added
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

* Removal of certificate references from HTTPRoutes,
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
  
*  [#671](https://github.com/kubernetes-sigs/gateway-api/pull/671) Controller is
now a required field in Gateway references from Route status. Fixes
[#669](https://github.com/kubernetes-sigs/gateway-api/pull/671).

*  [#657](https://github.com/kubernetes-sigs/gateway-api/pull/657) and
[#681](https://github.com/kubernetes-sigs/gateway-api/pull/681) Header Matching,
Query Param Matching, and HTTPRequestHeaderFilter now use named subobjects
instead of maps. 


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
