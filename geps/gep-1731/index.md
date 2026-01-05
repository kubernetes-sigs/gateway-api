# GEP-1731: HTTPRoute Retries

* Issue: [#1731](https://github.com/kubernetes-sigs/gateway-api/issues/1731)
* Status: Experimental

(See [status definitions](../overview.md#gep-states).)

## TLDR

To allow configuration of a Gateway to retry unsuccessful requests to backends before sending a response to a client request.

## Goals

* To allow specification of [HTTP status codes](https://www.rfc-editor.org/rfc/rfc9110#name-overview-of-status-codes) for which a request should be retried.
* To allow specification of the maximum number of times to retry a request.
* To allow specification of the minimum backoff interval between retry attempts.
* To define any interaction with configured HTTPRoute [timeouts](../gep-1742/index.md).
* Retry configuration must be applicable to most known Gateway API implementations.

## Future Goals

* To allow specification of a retry ["budget"](https://finagle.github.io/blog/2016/02/08/retry-budgets/) to determine whether a request should be retried, and any shared configuration or interaction with max count retry configuration.
* Define more precise semantics for retry configuration on "consumer" vs "producer" routes for mesh implementations.

## Non-Goals

* To allow more granular control of the backoff strategy than many dataplanes allow customizing, such as whether to use an exponential backoff interval between retry attempts, add jitter, or cap the backoff interval to a maximum duration.
* To allow specification of a default retry policy for all routes in a given namespace or attached to a particular Gateway.
* A standard API for approaches for retry logic other than max count or "budget", such as interaction with rate limiting headers.
* To allow specification of gRPC status codes for which a request should be retried (this should be covered in a separate GEP).
* Support for streaming or bidirectional APIs (these could be covered by a future GEP).

## Introduction

A Gateway API implementation should be able to retry failed HTTP requests to backends before delivering a response to a client for several reasons:

* Network Reliability: Networks can be unreliable, and connections might drop or time out. Retrying requests helps ensure that temporary issues don’t prevent the request from being completed.
* Load Balancing: If a server is temporarily overloaded or down, an implementation can retry the request to another server in a load-balanced environment.
* Error Handling: Some HTTP errors, like 500 Internal Server Error or 503 Service Unavailable, might be transient. Retrying the request can help bypass these temporary issues.

A primary audience for retries in Gateway API configuration are application developers (Ana) who want to ensure their applications are highly available and resilient. These users are best equipped to write sensible configuration, knowing which responses from their application should be retried and tolerances for timeouts and retry attempts to avoid overwhelming their applications with a ["retry storm"](https://learn.microsoft.com/en-us/azure/architecture/antipatterns/retry-storm/) which can cause performance issues, instability or outages.

Several Gateway API dataplanes support configuring retry semantics using their own bespoke configuration, but the details of these implementations and their user-facing configuration lack consistency between vendors. This proposal is an attempt to reconcile a minimal commonly implementable (and sufficiently useful) API from these divergent existing implementations.

### Background on implementations

Most implementations that handle HTTPRoute objects use a proxy as the data plane
implementation, which forwards traffic flows as directed by Gateway API configuration.

The following table is a review of all the listed implementations of Gateway API
at the time of writing, with the data plane they use for Layer 7, based on what information
could be found online. If there are errors here, or if the implementation doesn't
support Layer 7, please feel free to correct them.

| Implementation                           | Data Plane       |
|------------------------------------------|------------------|
| Acnodal EPIC                             | Envoy            |
| Apache APISIX                            | NGINX            |
| Azure Application Gateway for Containers | Envoy            |
| BIG-IP Kubernetes Gateway                | F5 BIG-IP        |
| Cilium                                   | Envoy            |
| Contour                                  | Envoy            |
| Emissary Ingress                         | Envoy            |
| Envoy Gateway                            | Envoy            |
| Flomesh Service Mesh                     | Pipy             |
| Gloo Edge                                | Envoy            |
| Google Kubernetes Engine (GKE)           | Similar to Envoy |
| HAProxy Ingress                          | HAProxy          |
| Hashicorp Consul                         | Envoy            |
| Istio                                    | Envoy            |
| Kong                                     | NGINX            |
| Kuma                                     | Envoy            |
| Linkerd                                  | linkerd2-proxy   |
| Litespeed                                | Litespeed WebADC |
| NGINX Gateway Fabric                     | NGINX            |
| Traefik                                  | Traefik          |
| VMWare Avi Load Balancer                 | ?                |

Implementation retry configuration details below have been summarized or copied verbatim from corresponding project documentation, with links to the original source where applicable.

#### Envoy

Retries are configurable in a [`RetryPolicy`](https://www.envoyproxy.io/docs/envoy/latest/api-v3/config/route/v3/route_components.proto#config-route-v3-retrypolicy) set on a [virtual host](https://www.envoyproxy.io/docs/envoy/latest/api-v3/config/route/v3/route_components.proto#envoy-v3-api-field-config-route-v3-virtualhost-retry-policy) (roughly equivalent to Listener), the [route](https://www.envoyproxy.io/docs/envoy/latest/api-v3/config/route/v3/route_components.proto#envoy-v3-api-field-config-route-v3-routeaction-retry-policy) (equivalent to HTTPRoute), or by setting [Router filter headers](https://www.envoyproxy.io/docs/envoy/latest/configuration/http/http_filters/router_filter). Policies are not merged - the most internal (specific) one becomes the enforced policy.

By default, Envoy uses a fully jittered exponential back-off algorithm for retries with a default base interval of 25ms, capped at a maximum interval, which defaults to 10 times the base interval (250ms). Explained in more depth in [`x-envoy-max-retries`](https://www.envoyproxy.io/docs/envoy/latest/configuration/http/http_filters/router_filter#x-envoy-max-retries).

* `retry_on` Allows specifying conditions for which a request will be retried, shared types with the header values documented in [x-envoy-retry-on](https://www.envoyproxy.io/docs/envoy/latest/configuration/http/http_filters/router_filter#config-http-filters-router-x-envoy-retry-on).

  * [**HTTP**](https://www.envoyproxy.io/docs/envoy/latest/configuration/http/http_filters/router_filter#x-envoy-retry-on)

    * `5xx` Retry when the backend returns any 5xx response code. Will not retry if a configured total outer request timeout has been exceeded.

    * `gateway-error` Will retry on 502, 503 and 504 response codes.

    * `reset` Retry if the server does not respond at all (disconnect, reset, timeout).

    * `reset-before-request` Retry if the server does not respond, but only if headers have not been sent yet.

    * `connect-failure` Retry on TCP connection failure, included in `5xx`.

    * `envoy-ratelimited` Retry even if the Envoy is trying to rate limit the connection.

    * `retriable-4xx` Currently includes only 409, with a warning that 409 status codes may indicate a condition which will consistently fail if retried.

    * `refused-stream` Retry if the server resets a stream with a `REFUSED_STREAM` error code, included in `5xx`.

    * `retriable-status-codes` Allows specifying any response code.

    * `retriable-headers` Allows specifying headers in the response for which a request should be retried.

    * `http3-post-connect-failure` Retry if a request is sent over HTTP/3 to the upstream server and fails after connecting.

* `num_retries` The allowed number of retries, defaults to 1. Further notes on specific behavior can be found under [`x-envoy-max-retries`](https://www.envoyproxy.io/docs/envoy/latest/configuration/http/http_filters/router_filter#x-envoy-max-retries).

* `per_try_timeout` Specifies a non-zero upstream timeout per retry attempt (including the initial attempt). If left unspecified, Envoy will use the global route timeout for the request. Equivalent to the Gateway API HTTPRouteRule BackendRequest timeout.

* `per_try_idle_timeout` This timeout is useful in cases in which total request timeout is bounded by a number of retries and a per_try_timeout, but there is a desire to ensure each try is making incremental progress.

* ...several options for retry priority and host selection, likely not portable

* `retriable_status_codes` Allows specifying any response code for which a request should be retried.

* `retry_back_off` Allows specifying a default base interval and max interval for retry attempts.

* `rate_limited_retry_back_off` Advanced configuration for adjusting retry attempt timing based on headers sent in the upstream server response. Alternative to the default exponential back off strategy.

* `retriable_headers` HTTP response headers for which a backend request should be retried.

* `retriable_request_headers` HTTP headers which must be present in the _request_ for retries to be attempted.

Supports configuring a [RetryBudget](https://www.envoyproxy.io/docs/envoy/latest/api-v3/config/cluster/v3/circuit_breaker.proto#envoy-v3-api-msg-config-cluster-v3-circuitbreakers-thresholds-retrybudget) with a following parameters in cluster CircuitBreaker thresholds.

* `budget_percent` Specifies the limit on concurrent retries as a percentage of the sum of active requests and active pending requests. For example, if there are 100 active requests and the budget_percent is set to 25, there may be 25 active retries. This parameter is optional. Defaults to 20%.

* `min_retry_concurrency` Specifies the minimum retry concurrency allowed for the retry budget. The limit on the number of active retries may never go below this number. This parameter is optional. Defaults to 3.

#### NGINX

The [`proxy_next_upstream`](http://nginx.org/en/docs/http/ngx_http_proxy_module.html#proxy_next_upstream) directive specifies in which cases a request should be passed to the next server:

* `error` An error occurred while establishing a connection with the server, passing a request to it, or reading the response header.
* `timeout` A timeout has occurred while establishing a connection with the server, passing a request to it, or reading the response header.
* `invalid_header` A server returned an empty or invalid response.
* `http_500`
* `http_502`
* `http_503`
* `http_504`
* `http_403`
* `http_404`
* `http_429`
* `non_idempotent` Normally, requests with a non-idempotent method (POST, LOCK, PATCH) are not passed to the next server if a request has been sent to an upstream server; enabling this option explicitly allows retrying such requests.
* `off` Disables passing a request to the next server.

Passing a request to the next server is only possible if nothing has been sent to a client yet. That is, if an error or timeout occurs in the middle of the transferring of a response, fixing this is impossible.

The directive also defines what is considered an unsuccessful attempt of communication with a server. The cases of error, timeout and invalid_header are always considered unsuccessful attempts, even if they are not specified in the directive. The cases of http_500, http_502, http_503, http_504, and http_429 are considered unsuccessful attempts only if they are specified in the directive. The cases of http_403 and http_404 are never considered unsuccessful attempts.

Passing a request to the next server can be limited by the number of tries and by time.

* `proxy_next_upstream_timeout time` Limits the time during which a request can be passed to the next server. The 0 value turns off this limitation. Equivalent to the Gateway API BackendRequest timeout.
* `proxy_next_upstream_tries number` Limits the number of possible tries for passing a request to the next server. The 0 value turns off this limitation.

The `max_fails` and `fail_timeout` parameters in the [`server`](http://nginx.org/en/docs/http/ngx_http_upstream_module.html#server) block of an [`upstream`](<http://nginx.org/en/docs/http/ngx_http_upstream_module.html>) module may also interact with retry logic.

* `max_fails=number` Sets the number of unsuccessful attempts to communicate with the server that should happen in the duration set by the `fail_timeout` parameter to consider the server unavailable for a duration also set by the `fail_timeout` parameter. By default, the number of unsuccessful attempts is set to 1. The zero value disables the accounting of attempts. What is considered an unsuccessful attempt is defined by the proxy_next_upstream, fastcgi_next_upstream, uwsgi_next_upstream, scgi_next_upstream, memcached_next_upstream, and grpc_next_upstream directives.

* `fail_timeout=time` Sets the time during which the specified number of unsuccessful attempts to communicate with the server should happen to consider the server unavailable; and the period of time the server will be considered unavailable. By default, the parameter is set to 10 seconds. Equivalent to Gateway API Request timeout.

May be possible to implement more advanced logic through Lua scripting or NGINX JavaScript.

#### HAProxy

Retry logic can be configured through the [`retry_on`](https://docs.haproxy.org/3.0/configuration.html#4.2-retry-on) and [`retries`](https://docs.haproxy.org/3.0/configuration.html#4.2-retries) proxy settings.

Using the `retry_on` directive replaces any previous settings with the new ones; it is not cumulative.

Using anything other than "none" and "conn-failure" allocates a buffer and copies the whole request into it, so it has memory and performance impacts. Requests not fitting in a single buffer will never be
retried.bufsize setting).

Warns that only requests known to be safe or idempotent requests, such as those with a unique transaction ID header to protect against accidental replays, should be retried.

* `none` never retry

* `conn-failure` Retry when the connection or the SSL handshake failed and the request could not be sent. This is the default.

* `empty-response` Retry when the server connection was closed after part of the request was sent, and nothing was received from the server. This type of failure may be caused by the request timeout on the server side, poor network condition, or a server crash or restart while processing the request.

* `junk-response` Retry when the server returned something not looking
like a complete HTTP response. This includes partial responses headers as well as non-HTTP contents. It usually is a bad idea to retry on such events, which may be caused a configuration issue (wrong server port)
or by the request being harmful to the server (buffer overflow attack for example).

* `response-timeout` The server timeout stroke while waiting for the server to respond to the request. This may be caused by poor network condition, the reuse of an idle connection which has expired on the path, or by the request being extremely expensive to process. It generally is a bad idea to retry on such events on servers dealing with heavy database processing (full scans, etc) as it may amplify denial of service attacks.

* `0rtt-rejected` Retry requests which were sent over early data and were
rejected by the server. These requests are generally considered to be safe to retry.

* <status> Retry on select HTTP status codes among "401" (Unauthorized), "403" (Forbidden), "404" (Not Found), "408" (Request Timeout), "425" (Too Early), "500" (Server Error), "501" (Not Implemented), "502" (Bad Gateway), "503" (Service Unavailable), "504" (Gateway Timeout).

* `all-retriable-errors` Retry request for any error that are considered
retriable. This currently activates "conn-failure", "empty-response", "junk-response", "response-timeout", "0rtt-rejected", "500", "502", "503", and "504".

The [`option redispatch`](https://docs.haproxy.org/3.0/configuration.html#4.2-option%20redispatch) configuration can be used to distribute retries across multiple backend servers, allowing the proxy to break cookie or consistent hash based persistence and redistribute them to a working server.

#### Traefik

Offers a [Retry middleware](https://doc.traefik.io/traefik/middlewares/http/retry/) allowing configuration of number of attempts and initial interval.

Reissues requests a given number of times to a backend server if that server does not reply. As soon as the server answers, the middleware stops retrying, _regardless of the response status_. Has an optional configuration to enable an exponential backoff.

* `attempts` Defines how many times the request should be retried.

* `initialInterval` Defines the first wait time in the backoff series. The maximum interval is calculated as twice the `initialInterval`. If unspecified, requests will be retried immediately. The value should be provided in seconds or as a valid duration format.

Supports configuration of a [Circuit Breaker](https://doc.traefik.io/traefik/middlewares/http/circuitbreaker/) which could possibly be used to implement budgeted retries. Each router gets its own instance of a given circuit breaker. One circuit breaker instance can be open while the other remains closed: their state is not shared. This is the expected behavior, we want you to be able to define what makes a service healthy without having to declare a circuit breaker for each route.

#### linkerd2-proxy

Linkerd supports [budgeted retries](https://linkerd.io/2.15/features/retries-and-timeouts/) and - as of [edge-24.7.5](https://github.com/linkerd/linkerd2/releases/tag/edge-24.7.5) - counted retries. In all cases, retries are implemented by the `linkerd2-proxy` making the request on behalf on an application workload.

Linkerd's budgeted retries allow retrying an indefinite number of times, as long as the fraction of retries remains within the budget. Budgeted retries are supported only using Linkerd's native ServiceProfile CRD, which allows enabling retries, setting the retry budget (by default, 20% plus 10 "extra" retries per second), and configuring the window over which the fraction of retries to non-retries is calculated.

Linkerd's counted retries work in much the same way as other retry implementations, permitted a fixed maximum number of retries for a given request. Counted retries can be configured with annotations on Service, HTTPRoute, and GRPCRoute resources, all of which allow setting the maximum number of retries and the maximum time a request will be permitted to linger before cancelling it and retrying.

For Service and HTTPRoute, Linkerd permits setting which HTTP statuses will be retried. For Service and GRPCRoute, Linkerd permits setting which gRPC statuses will be retried. Retry configurations on HTTPRoute and GRPCRoute resources take precedence over retry configurations on Service resources.

Neither type of Linkerd retries supports configuring retries based on connection status (e.g. connection timed out). The `linkerd2-proxy` maintains long-lived connections to destinations in use, and manages connection state independently of the application making requests. In this world, individual requests don't correlate well with connections being made or broken, which means that retries on connection state changes don't really make sense.

#### F5 BIG-IP

TODO

#### Pipy

TODO

#### Litespeed WebADC

TODO

## API

!!! warning
    Expectations for how implementations should handle connection errors are currently UNRESOLVED due to inconsistency between data planes, including how connections are established or maintained, granularity of how different types of errors are bucketed, default behavior and expected user needs. Please see comment thread at <https://github.com/kubernetes-sigs/gateway-api/pull/3199#discussion_r1697201266> for more detail.

### Go

```golang
type HTTPRouteRule struct {
    // Retry defines the configuration for when to retry an HTTP request.
    //
    // Support: Extended
    //
    // +optional
    // <gateway:experimental>
    Retry *HTTPRouteRetry `json:"retry,omitempty"`

    // ...
}

// HTTPRouteRetry defines retry configuration for an HTTPRoute.
//
// Implementations SHOULD retry on connection errors (disconnect, reset, timeout,
// TCP failure) if a retry stanza is configured.
//
type HTTPRouteRetry struct {
    // Codes defines the HTTP response status codes for which a backend request
    // should be retried.
    //
    // Support: Extended
    //
    // +optional
    // <gateway:experimental>
    Codes []HTTPRouteRetryStatusCode `json:"codes,omitempty"`

    // Attempts specifies the maximum number of times an individual request
    // from the gateway to a backend should be retried.
    //
    // If the maximum number of retries has been attempted without a successful
    // response from the backend, the Gateway MUST return an error.
    //
    // When this field is unspecified, the number of times to attempt to retry
    // a backend request is implementation-specific.
    //
    // Support: Extended
    //
    // +optional
    Attempts *Int `json:"attempts,omitempty"`

    // Backoff specifies the minimum duration a Gateway should wait between
    // retry attempts and is represented in Gateway API Duration formatting.
    //
    // For example, setting the `rules[].retry.backoff` field to the value
    // `100ms` will cause a backend request to first be retried approximately
    // 100 milliseconds after timing out or receiving a response code configured
    // to be retriable.
    //
    // An implementation MAY use an exponential or alternative backoff strategy
    // for subsequent retry attempts, MAY cap the maximum backoff duration to
    // some amount greater than the specified minimum, and MAY add arbitrary
    // jitter to stagger requests, as long as unsuccessful backend requests are
    // not retried before the configured minimum duration.
    //
    // If a Request timeout (`rules[].timeouts.request`) is configured on the
    // route, the entire duration of the initial request and any retry attempts
    // MUST not exceed the Request timeout duration. If any retry attempts are
    // still in progress when the Request timeout duration has been reached,
    // these SHOULD be canceled if possible and the Gateway MUST immediately
    // return a timeout error.
    //
    // If a BackendRequest timeout (`rules[].timeouts.backendRequest`) is
    // configured on the route, any retry attempts which reach the configured
    // BackendRequest timeout duration without a response SHOULD be canceled if
    // possible and the Gateway should wait for at least the specified backoff
    // duration before attempting to retry the backend request again.
    //
    // If a BackendRequest timeout is _not_ configured on the route, retry
    // attempts MAY time out after an implementation default duration, or MAY
    // remain pending until a configured Request timeout or implementation
    // default duration for total request time is reached.
    //
    // When this field is unspecified, the time to wait between retry attempts
    // is implementation-specific.
    //
    // Support: Extended
    //
    // +optional
    Backoff *Duration `json:"backoff,omitempty"`
}

// HTTPRouteRetryStatusCode defines an HTTP response status code for
// which a backend request should be retried.
//
// Implementations MUST support the following status codes as retriable:
//
// * 500
// * 502
// * 503
// * 504
//
// Implementations MAY support specifying additional discrete values in the
// 500-599 range.
//
// Implementations SHOULD NOT support retrying status codes in the 100-399
// range, as these responses are generally not appropriate to retry.
//
// Implementations MAY support specifying discrete values in the 400-499 range,
// which are often inadvisable to retry.
//
// Implementations MAY support discrete values in the 600-999 (inclusive)
// range, which are not valid for HTTP clients, but are sometimes used for
// communicating application-specific errors.
//
// +kubebuilder:validation:Minimum:=100
// +kubebuilder:validation:Maximum:=999
type HTTPRouteRetryStatusCode int

// Duration is a string value representing a duration in time. The format is
// as specified in GEP-2257, a strict subset of the syntax parsed by Golang
// time.ParseDuration.
//
// +kubebuilder:validation:Pattern=`^([0-9]{1,5}(h|m|s|ms)){1,4}$`
type Duration string
```

### YAML

```yaml
apiVersion: gateway.networking.k8s.io/v1
kind: HTTPRoute
metadata:
  name: retry-example
spec:
  ...
  rules:
  - backendRefs:
    - name: some-service
      port: 8080
    retry:
      codes:
      - 500
      - 502
      - 503
      - 504
      attempts: 2
      backoff: 100ms
```

## Conformance Details

Basic support for configuring retries in HTTPRoute up to a specified maximum count and with a minimum duration between attempts will be gated on the `SupportHTTPRouteRetry` feature. Retries will be tested in combination with request and backend timeouts if supported by the implementation.

Retrying requests based on HTTP status codes will be gated under the following features:

* `SupportHTTPRouteRetryBackendTimeout`

  * Will test that backend requests that exceed a BackendRequest timeout duration are retried if a `retry` stanza is configured.

* `SupportHTTPRouteRetryBackoff`

  * Backoff will only be tested that a retry does not start before the duration specified for conformance, not that the backoff duration is precise.
  * Not currently supportable by NGINX or HAProxy.

* `SupportHTTPRouteRetryCodes`

  * Only 500, 502, 503 and 504 will be tested for conformance.
  * Traefik does not seem to support specifying error codes, and will only retry on backend timeouts.

* `SupportHTTPRouteRetryConnectionError`

  * Will test that connections interrupted by a TCP failure, disconnect or reset are retried if a `retry` stanza is configured.

## Alternatives

### Standard RetryPolicy metaresource

This may be a reasonable approach for configuring broad default retry policies, but the UX of using a separate resource could feel overly burdensome for basic use cases or granular configuration, and the structure of the policy attachment model is [currently being revised](https://github.com/kubernetes-sigs/gateway-api/discussions/2927). Additionally, multiple data plane proxies warn in their documentation about the danger of configuring broad retry policies which may cause logical application errors from replaying non-idempotent requests or overwhelm backends. This approach may still be considered to enable future goals.

### HTTPRoute filter

Implementing a `requestRetryPolicy` [HTTPRouteFilter](../../reference/spec.md#httproutefilter) type is likely a reasonable alternative implementation (with the downside of slightly deeping nesting and more complex structural configuration) that was not fully explored.

Adding a new field to HTTPRouteRule instead is proposed for parity with the similar and intersecting configuration of [HTTPRouteTimeouts](../../reference/spec.md#httproutetimeouts).

## Other considerations

### What accommodations are needed for future retry budget support?

Changing the retry stanza to a Kubernetes "tagged union" pattern with something like `mode: "budget"` to support mutually-exclusive distinct sibling fields is possible as a non-breaking change if omitting the `mode` field defaults to the currently proposed behavior (which could retroactively become something like `mode: count`).

### Should whether to retry on connection errors be configurable?

Retrying on connection errors (disconnect, reset, timeout, TCP failure) is typically default behavior for most dataplanes, even those that don't support configurable retries based on HTTP status codes. Some proxies allow this to be configurable, but retrying attempts on these sort of errors is typically desirable and excluding it from the Gateway API spec both allows a simpler UX with less boilerplate, and avoids nuances between how these are defined and configured in different proxies.

### Should whether to retry on a backend timeout be configurable?

On Kubernetes, retrying should _typically_ route a backend request to a different pod if the original destination has become unhealthy and therefore should generally be safe. Even if a [BackendLBPolicy](../gep-1619/index.md) is configured, most dataplane implementations implement "soft" affinity rather than strict session routing. The warnings against this practice in NGINX and HAProxy documentation seem to reference risks with legacy deployment models using a small number of statically-defined servers. We could consider adding something like a `excludeRetryOnTimeout` boolean field (implementable by NGINX and HAProxy, not by Envoy) in the future if this behavior is desirable, while still retaining the retry-on-timeout behavior as a default.

## References

* <https://www.rfc-editor.org/rfc/rfc9110>
* <https://www.iana.org/assignments/http-status-codes/http-status-codes.xhtml>
* <https://gateway-api.sigs.k8s.io/geps/gep-2257/>
* <https://gateway-api.sigs.k8s.io/geps/gep-1742/>
* <https://github.com/kubernetes-sigs/gateway-api/issues/3139>
