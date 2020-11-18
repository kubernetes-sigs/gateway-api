# Changelog

## Table of Contents

- [v0.1.0](#v010)
- [v0.1.0-rc2](#v010-rc2)
- [v0.1.0-rc1](#v010-rc1)

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
  - Header name matching must be case-insenstive.
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
