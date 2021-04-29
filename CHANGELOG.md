# Changelog

## Table of Contents

- [v0.3.0](#v030)
- [v0.2.0](#v020)
- [v0.1.0](#v010)
- [v0.1.0-rc2](#v010-rc2)
- [v0.1.0-rc1](#v010-rc1)

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
