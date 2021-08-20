# HTTPRoute

[HTTPRoute][httproute] is a Gateway API type for specifying routing behavior
of HTTP requests from a Gateway listener to an API object, i.e. Service.

## Spec

The specification of an HTTPRoute consists of:

- [Gateways][gateways]- Define which Gateways can use this HTTPRoute.
- [Hostnames][hostname] (optional)- Define a list of hostnames to use for
  matching the Host header of HTTP requests.
- [TLS][tls-config] (optional)- Defines the TLS certificate to use for
  Hostnames defined in this Route.
- [Rules][httprouterule]- Define a list of rules to perform actions against
  matching HTTP requests. Each rule consists
  of [matches][matches], [filters][filters] (optional), and [forwardTo][forwardto]
  (optional) fields.

The following illustrates an HTTPRoute that sends all traffic to one Service:
![httproute-basic-example](/images/httproute-basic-example.svg)

### Gateways

Gateways define which Gateways can use the HTTPRoute. If unspecified, `gateways`
defaults to `allow: SameNamespace` which allows all Gateways in the HTTPRoute's
namespace.

The following example allows Gateways from namespace "httproute-ns-example":
```yaml
kind: HTTPRoute
apiVersion: networking.x-k8s.io/v1alpha1
metadata:
  name: httproute-example
  namespace: httproute-ns-example
spec:
  gateways:
    allow: SameNamespace
```

Possible values for `allow` are:

- `All`: Gateways in any namespace can use this route.
- `FromList`: Only Gateways specified in `gatewayRefs` may use this route.
- `SameNamespace` (default): Only Gateways in the same namespace may use this
  route.

If `allow` results in preventing the selection of an HTTPRoute by a Gateway, an
“Admitted: false” condition must be set on the Gateway for this Route.

### Hostnames

Hostnames define a list of hostnames to match against the Host header of the
HTTP request. When a match occurs, the HTTPRoute is selected to perform request
routing based on rules and filters (optional). A hostname is the fully qualified
domain name of a network host, as defined by [RFC 3986][rfc-3986]. Note the
following deviations from the “host” part of the URI as defined in the RFC:

- IPs are not allowed.
- The : delimiter is not respected because ports are not allowed.

Incoming requests are matched against hostnames before the HTTPRoute rules are
evaluated. If no hostname is specified, traffic is routed based on HTTPRoute
rules and filters (optional).

The following example defines hostname "my.example.com" and allows Gateways
from the same namespace as HTTPRoute "httproute-example":
```yaml
kind: HTTPRoute
apiVersion: networking.x-k8s.io/v1alpha1
metadata:
  name: httproute-example
spec:
  gateways:
    allow: SameNamespace
  hostnames:
  - my.example.com
```

### TLS

TLS defines the TLS certificate used for hostnames defined in this HTTPRoute.
This configuration only takes effect if `certificate: Allow` is set for
`routeOverride` in the associated Gateway. For example:
```yaml
{% include 'v1alpha1/tls-basic.yaml' %}
```

`CertificateRef` refers to a Kubernetes object that contains a TLS certificate
and private key. This certificate MUST be used for TLS handshakes for the domain
this `tls` is associated with. If an entry in this list omits or specifies the
empty string for both the `group` and `kind`, the resource defaults to “secrets”.

**Notes:**

- HTTPRoute selection takes place after the TLS Handshake (ClientHello). Due to
this, a TLS certificate in an HTTPRoute will take precedence even if the request
has the potential to match multiple HTTPRoutes (in case multiple HTTPRoutes
share the same hostname).
- Collisions can happen if multiple HTTPRoutes define a TLS certificate for the
same hostname. In such case, the certificate in the oldest HTTPRoute is selected.

### Rules

Rules define semantics for matching an HTTP request based on conditions,
optionally executing additional processing steps, and optionally forwarding
the request to an API object.

#### Matches

Matches define conditions used for matching an HTTP request. Each match is
independent, i.e. this rule will be matched if any single match is satisfied.

Take the following matches configuration as an example:
```yaml
kind: HTTPRoute
apiVersion: networking.x-k8s.io/v1alpha1
...
matches:
  - path:
      value: "/foo"
    headers:
      values:
        version: "2"
  - path:
      value: "/v2/foo"
```

For a request to match against this rule, it must satisfy EITHER of the
following conditions:

 - A path prefixed with /foo **AND** contains the header "version: 2"
 - A path prefix of /v2/foo

If no matches are specified, the default is a prefix path match on “/”,
which has the effect of matching every HTTP request.

#### Filters (optional)

Filters define processing steps that must be completed during the request or
response lifecycle. Filters act as an extension point to express additional
processing that may be performed in Gateway implementations. Some examples
include request or response modification, implementing authentication
strategies, rate-limiting, and traffic shaping.

The following example adds header "my-header: foo" to HTTP requests with Host
header "my.filter.com".
```yaml
{% include 'v1alpha1/http-filter.yaml' %}
```

API conformance is defined based on the filter type. The effects of ordering
multiple behaviors is currently unspecified. This may change in the future
based on feedback during the alpha stage.

Conformance levels are defined by the filter type:

 - All "core" filters MUST be supported by implementations.
 - Implementers are encouraged to support "extended" filters.
 - "Custom" filters have no API guarantees across implementations.

Specifying a core filter multiple times has unspecified or custom conformance.

#### ForwardTo (optional)

ForwardTo defines API objects where matching requests should be sent. If
unspecified, the rule performs no forwarding. If unspecified and no filters
are specified that would result in a response being sent, a 503 error code
is returned.

The following example forwards HTTP requests for prefiex `/bar` to service
"my-service1" on port `8080` and HTTP requests for prefex `/some/thing` with
header `magic: foo` to service "my-service2" on port `8080`:
```yaml
{% include 'v1alpha1/basic-http.yaml' %}
```

**Note:** Forwarding to a custom resource instead of a service can be
accomplished by specifying `backendRef` instead of `serviceName`. A
`backendRef` follows the standard Kubernetes `group`, `kind` and `name`
schema.

The following example uses the `weight` field to forward HTTP requests for
prefix `/bar` equally across service "my-trafficsplit-svc1" and service
"my-trafficsplit-svc2", i.e. traffic splitting:
```yaml
{% include 'v1alpha1/http-trafficsplit.yaml' %}
```

Reference the [forwardTo][forwardto] API documentation for additional details
of `weight` and other fields.

## Status

Status defines the observed state of HTTPRoute.

### RouteStatus

RouteStatus defines the observed state that is required across all route types.

#### Gateways

Gateways define a list of the Gateways that are associated with the HTTPRoute,
and the status of the HTTPRoute with respect to each of these Gateways. When a
Gateway selects this HTTPRoute, the controller that manages the Gateway should
add an entry to this list when the controller first sees the route and should
update the entry as appropriate when the route is modified.

The following example indicates HTTPRoute "http-example" has been admitted by
Gateway "gw-example" in namespace "gw-example-ns":
```yaml
kind: HTTPRoute
apiVersion: networking.x-k8s.io/v1alpha1
metadata:
  name: http-example
...
status:
  gateways:
  - gatewayRef:
      name: gw-example
      namespace: gw-example-ns
    conditions:
    - type: Admitted
      status: "True"
```

A maximum of 100 Gateways can be represented in this list. If this list is full,
there may be additional Gateways using this Route that are not included in the
list.

### Merging
Multiple HTTPRoutes can be attached to a single Gateway resource. Importantly,
only one Route rule may match each request. For more information on how conflict
resolution applies to merging, refer to the [API specification](httprouterule).


[httproute]: https://gateway-api.sigs.k8s.io/references/spec/#networking.x-k8s.io/v1alpha1.HTTPRoute
[gateways]: https://gateway-api.sigs.k8s.io/references/spec/#networking.x-k8s.io/v1alpha1.RouteGateways
[httprouterule]: https://gateway-api.sigs.k8s.io/references/spec/#networking.x-k8s.io/v1alpha1.HTTPRouteRule
[hostname]: https://gateway-api.sigs.k8s.io/references/spec/#networking.x-k8s.io/v1alpha1.Hostname
[tls-config]: https://gateway-api.sigs.k8s.io/references/spec/#networking.x-k8s.io/v1alpha1.RouteTLSConfig
[rfc-3986]: https://tools.ietf.org/html/rfc3986
[matches]: https://gateway-api.sigs.k8s.io/references/spec/#networking.x-k8s.io/v1alpha1.HTTPRouteMatch
[filters]: https://gateway-api.sigs.k8s.io/references/spec/#networking.x-k8s.io/v1alpha1.HTTPRouteFilter
[forwardto]: https://gateway-api.sigs.k8s.io/references/spec/#networking.x-k8s.io/v1alpha1.HTTPRouteForwardTo
