# HTTPRoute

??? success "Standard Channel since v0.5.0"

    The `HTTPRoute` resource is GA and has been part of the Standard Channel since
    `v0.5.0`. For more information on release channels, refer to our [versioning
    guide](/concepts/versioning).

[HTTPRoute][httproute] is a Gateway API type for specifying routing behavior
of HTTP requests from a Gateway listener to an API object, i.e. Service.

## Spec

The specification of an HTTPRoute consists of:

- [ParentRefs][parentRef]- Define which Gateways this Route wants to be attached
  to.
- [Hostnames][hostname] (optional)- Define a list of hostnames to use for
  matching the Host header of HTTP requests.
- [Rules][httprouterule]- Define a list of rules to perform actions against
  matching HTTP requests. Each rule consists of [matches][matches],
  [filters][filters] (optional), [backendRefs][backendRef] (optional) and [timeouts][timeouts] (optional)
  fields.

The following illustrates an HTTPRoute that sends all traffic to one Service:
![httproute-basic-example](/images/httproute-basic-example.svg)

### Attaching to Gateways

Each Route includes a way to reference the parent resources it wants to attach
to. In most cases, that's going to be Gateways, but there is some flexibility
here for implementations to support other types of parent resources.

The following example shows how a Route would attach to the `acme-lb` Gateway:
```yaml
apiVersion: gateway.networking.k8s.io/v1beta1
kind: HTTPRoute
metadata:
  name: httproute-example
spec:
  parentRefs:
  - name: acme-lb
```

Note that the target Gateway needs to allow HTTPRoutes from the route's
namespace to be attached for the attachment to be successful.

You can also attach routes to specific sections of the parent resource.
For example, let's say that the `acme-lb` Gateway includes the following
listeners:

```yaml
  listeners:
  - name: foo
    protocol: HTTP
    port: 8080
    ...
  - name: bar
    protocol: HTTP
    port: 8090
    ...
  - name: baz
    protocol: HTTP
    port: 8090
    ...
```

You can bind a route to listener `foo` only, using the `sectionName` field
in `parentRefs`:

```yaml
spec:
  parentRefs:
  - name: acme-lb
    sectionName: foo
```

Alternatively, you can achieve the same effect by using the `port` field,
instead of `sectionName`, in the `parentRefs`:

```yaml
spec:
  parentRefs:
  - name: acme-lb
    port: 8080
```

Binding to a port also allows you to attach to multiple listeners at once.
For example, binding to port `8090` of the `acme-lb` Gateway would be more
convenient than binding to the corresponding listeners by name:

```yaml
spec:
  parentRefs:
  - name: acme-lb
    sectionName: bar
  - name: acme-lb
    sectionName: baz
```

However, when binding Routes by port number, Gateway admins will no longer have
the flexibility to switch ports on the Gateway without also updating the Routes.
The approach should only be used when a Route should apply to a specific port
number as opposed to listeners whose ports may be changed.

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

The following example defines hostname "my.example.com":
```yaml
apiVersion: gateway.networking.k8s.io/v1beta1
kind: HTTPRoute
metadata:
  name: httproute-example
spec:
  hostnames:
  - my.example.com
```

### Rules

Rules define semantics for matching an HTTP request based on conditions,
optionally executing additional processing steps, and optionally forwarding
the request to an API object.

#### Matches

Matches define conditions used for matching an HTTP request. Each match is
independent, i.e. this rule will be matched if any single match is satisfied.

Take the following matches configuration as an example:
```yaml
apiVersion: gateway.networking.k8s.io/v1beta1
kind: HTTPRoute
...
spec:
  rules:
  - matches:
    - path:
        value: "/foo"
      headers:
      - name: "version"
        value: "2"
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
{% include 'standard/http-filter.yaml' %}
```

API conformance is defined based on the filter type. The effects of ordering
multiple behaviors is currently unspecified. This may change in the future
based on feedback during the alpha stage.

Conformance levels are defined by the filter type:

 - All "core" filters MUST be supported by implementations.
 - Implementers are encouraged to support "extended" filters.
 - "Implementation-specific" filters have no API guarantees across implementations.

Specifying a core filter multiple times has unspecified or 
implementation-specific conformance.

All filters are expected to be compatible with each other except for the
URLRewrite and RequestRedirect filters, which may not be combined. If an
implementation can not support other combinations of filters, they must clearly
document that limitation. In cases where incompatible or unsupported
filters are specified and cause the `Accepted` condition to be set to status
`False`, implementations may use the `IncompatibleFilters` reason to specify
this configuration error.

#### BackendRefs (optional)

BackendRefs defines API objects where matching requests should be sent. If
unspecified, the rule performs no forwarding. If unspecified and no filters
are specified that would result in a response being sent, a 404 error code
is returned.

The following example forwards HTTP requests for prefix `/bar` to service
"my-service1" on port `8080` and HTTP requests for prefix `/some/thing` with
header `magic: foo` to service "my-service2" on port `8080`:
```yaml
{% include 'standard/basic-http.yaml' %}
```

The following example uses the `weight` field to forward 90% of HTTP requests to
`foo.example.com` to the "foo-v1" Service and the other 10% to the "foo-v2"
Service:
```yaml
{% include 'standard/traffic-splitting/traffic-split-2.yaml' %}
```

Reference the [backendRef][backendRef] API documentation for additional details
on `weight` and other fields.

#### Timeouts (optional)

??? example "Experimental Channel since v1.0.0"

    HTTPRoute timeouts have been part of the Experimental Channel since `v1.0.0`.
    For more information on release channels, refer to our 
    [versioning guide](/concepts/versioning).

HTTPRoute Rules include a `Timeouts` field. If unspecified, timeout behavior is implementation-specific.

There are 2 kinds of timeouts that can be configured in an HTTPRoute Rule:

1. `request` is the timeout for the Gateway API implementation to send a response to a client HTTP request. This timeout is intended to cover as close to the whole request-response transaction as possible, although an implementation MAY choose to start the timeout after the entire request stream has been received instead of immediately after the transaction is initiated by the client.

2. `backendRequest` is a timeout for a single request from the Gateway to a backend. This timeout covers the time from when the request first starts being sent from the gateway to when the full response has been received from the backend. This can be particularly helpful if the Gateway retries connections to a backend.

Because the `request` timeout encompasses the `backendRequest` timeout, the value of `backendRequest` must not be greater than the value of `request` timeout.

Timeouts are optional, and their fields are of type [Duration](/geps/gep-2257/). A zero-valued timeout ("0s") MUST be interpreted as disabling the timeout. A valid non-zero-valued timeout MUST be >= 1ms.

The following example uses the `request` field which will cause a timeout if a client request is taking longer than 10 seconds to complete. The example also defines a 2s `backendRequest` which specifies a timeout for an individual request from the gateway to a backend service `timeout-svc`:
```yaml
{% include 'experimental/http-route-timeouts/timeout-example.yaml' %}
```

Reference the [timeouts][timeouts] API documentation for additional details.

##### Backend Protocol

??? example "Experimental Channel since v1.0.0"

    This concept has been part of the Experimental Channel since `v1.0.0`.
    For more information on release channels, refer to our 
    [versioning guide](/concepts/versioning).

Some implementations may require the [backendRef][backendRef] to be labeled 
explicitly in order to route traffic using a certain protocol. For Kubernetes 
Service backends this can be done by specifying the [`appProtocol`][appProtocol]
field.


## Status

Status defines the observed state of HTTPRoute.

### RouteStatus

RouteStatus defines the observed state that is required across all route types.

#### Parents

Parents define a list of the Gateways (or other parent resources) that are
associated with the HTTPRoute, and the status of the HTTPRoute with respect to
each of these Gateways. When a HTTPRoute adds a reference to a Gateway in
parentRefs, the controller that manages the Gateway should add an entry to this
list when the controller first sees the route and should update the entry as
appropriate when the route is modified.

The following example indicates HTTPRoute "http-example" has been accepted by
Gateway "gw-example" in namespace "gw-example-ns":
```yaml
apiVersion: gateway.networking.k8s.io/v1beta1
kind: HTTPRoute
metadata:
  name: http-example
...
status:
  parents:
  - parentRefs:
      name: gw-example
      namespace: gw-example-ns
    conditions:
    - type: Accepted
      status: "True"
```

## Merging
Multiple HTTPRoutes can be attached to a single Gateway resource. Importantly,
only one Route rule may match each request. For more information on how conflict
resolution applies to merging, refer to the [API specification][httprouterule].


[httproute]: /reference/spec/#gateway.networking.k8s.io/v1.HTTPRoute
[httprouterule]: /reference/spec/#gateway.networking.k8s.io/v1.HTTPRouteRule
[hostname]: /reference/spec/#gateway.networking.k8s.io/v1.Hostname
[rfc-3986]: https://tools.ietf.org/html/rfc3986
[matches]: /reference/spec/#gateway.networking.k8s.io/v1.HTTPRouteMatch
[filters]: /reference/spec/#gateway.networking.k8s.io/v1.HTTPRouteFilter
[backendRef]: /reference/spec/#gateway.networking.k8s.io/v1.HTTPBackendRef
[parentRef]: /reference/spec/#gateway.networking.k8s.io/v1.ParentRef
[timeouts]: /reference/spec/#gateway.networking.k8s.io/v1.HTTPRouteTimeouts
[appProtocol]: https://kubernetes.io/docs/concepts/services-networking/service/#application-protocol
