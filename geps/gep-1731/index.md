# GEP-1731: Configurable Retries

* Issue: [#1731](https://github.com/kubernetes-sigs/gateway-api/issues/1731)
* Status: Implementable

(See status definitions [here](/geps/overview/#gep-states).)

## TLDR

To allow configuration of a Gateway to retry unsuccessful requests to backends before sending a response to a client request.

## Goals

* To allow specification of [HTTP status codes](https://www.rfc-editor.org/rfc/rfc9110#name-overview-of-status-codes) for which a request should be retried.
* To allow specification of the maximum number of times to retry a request.
* To allow specification of the minimum backoff interval between retry attempts.
* To define any interaction with configured HTTPRoute [timeouts](/geps/gep-1742/).
* Retry configuration must be applicable to most known Gateway API implementations.

## Future Goals

* To allow specification of gRPC status codes for which a request should be retried.
* To allow specification of a retry ["budget"](https://finagle.github.io/blog/2016/02/08/retry-budgets/) to determine whether a request should be retried, and any shared configuration or interaction with max count retry configuration.
* Define more precise semantics for retry configuration on "consumer" vs "producer" routes for mesh implementations.

## Non-Goals

* To allow more granular control of the backoff strategy than many dataplanes allow customizing, such as whether to use an exponential backoff interval between retry attempts, add jitter, or cap the backoff interval to a maximum duration.
* To allow specification of a default retry policy for all routes in a given namespace or attached to a particular Gateway.
* A standard API for approaches for retry logic other than max count or "budget", such as interaction with rate limiting headers.
* Support for unary or bidirectional streams, which may have different considerations for timeouts or request/response patterns within the stream after establishment.

## Introduction

TODO

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

  * [**gRPC**](https://www.envoyproxy.io/docs/envoy/latest/configuration/http/http_filters/router_filter#x-envoy-retry-grpc-on)

    * `cancelled`
    * `deadline-exceeded`
    * `internal`
    * `resource-exhausted`
    * `unavailable`

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

May be possible to implement more advanced logic through Lua scripting.

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

* `all-retryable-errors` Retry request for any error that are considered
retryable. This currently activates "conn-failure", "empty-response", "junk-response", "response-timeout", "0rtt-rejected", "500", "502", "503", and "504".

The [`option redispatch`](https://docs.haproxy.org/3.0/configuration.html#4.2-option%20redispatch) configuration can be used to distribute retries across multiple backend servers, allowing the proxy to break cookie or consistent hash based persistence and redistribute them to a working server.

#### Traefik

Offers a [Retry middleware](https://doc.traefik.io/traefik/middlewares/http/retry/) allowing configuration of number of attempts and initial interval.

Reissues requests a given number of times to a backend server if that server does not reply. As soon as the server answers, the middleware stops retrying, _regardless of the response status_. Has an optional configuration to enable an exponential backoff.

* `attempts` Defines how many times the request should be retried.

* `initialInterval` Defines the first wait time in the backoff series. The maximum interval is calculated as twice the `initialInterval`. If unspecified, requests will be retried immediately. The value should be provided in seconds or as a valid duration format.

Supports configuration of a [Circuit Breaker](https://doc.traefik.io/traefik/middlewares/http/circuitbreaker/) which could possibly be used to implement budgeted retries. Each router gets its own instance of a given circuit breaker. One circuit breaker instance can be open while the other remains closed: their state is not shared. This is the expected behavior, we want you to be able to define what makes a service healthy without having to declare a circuit breaker for each route.

#### linkerd2-proxy

Linkerd supports a configurable [retry budget](https://linkerd.io/2.15/features/retries-and-timeouts/) per-route using its [ServiceProfile](https://linkerd.io/2.15/reference/service-profiles/) CRD. The default budget is 20%, but Linkerd also supports a certain number of "extra" retries every second (10 by default), for a better user experience with low-traffic services.

Linkerd is unique in that it does not currently support counted retries, although this is an area of active development.

#### F5 BIG-IP

TODO

#### Pipy

TODO

#### Litespeed WebADC

TODO

## API

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
// Implementations MUST retry on connection errors (disconnect, reset, timeout,
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

    // Attempts specifies the maxmimum number of times an individual request
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
    // 100 milliseconds after timing out or reciveing a response code configured
    // to be retryable.
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
    // these SHOULD be cancelled if possible and the Gateway MUST immediately
    // return a timeout error.
    //
    // If a BackendRequest timeout (`rules[].timeouts.backendRequest`) is
    // configured on the route, any retry attempts which reach the configured
    // BackendRequest timeout duration without a response SHOULD be cancelled if
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
// Implementations MUST support the following status codes as retryable:
//
// * "500"
// * "502"
// * "503"
// * "504"
//
// Implementations SHOULD support the `"5xx"` shorthand for matching status
// codes in the 500-599 range. For implementations supporting `"5xx"`, this
// shorthand MUST match the following status codes:
//
// * "500"
// * "502"
// * "503"
// * "504"
//
// The `"5xx"` shorthand MAY additionally match a well-documented subset of
// arbitrary status codes in the 500-599 range, or match all status codes in
// this range.
//
// Implementations SHOULD NOT support retrying status codes in the 100-399
// range, as these responses are generally not appropriate to retry.
//
// Implementations MAY support specifying discrete values in the 400-499 range,
// which are often inadvisable to retry, and MAY support discrete values in the
// 600-999 (inclusive) range, which are not valid for HTTP clients, but are
// sometimes used for communicating application-specific errors.
//
// Implementations MAY support additional shorthand codes for any `[0-9]xx`
// range.
//
// +kubebuilder:validation:Pattern=`^[1-9](?:[0-9][0-9]|xx)$`
type HTTPRouteRetryStatusCode string

// Duration is a string value representing a duration in time. The foramat is
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
        - 5xx
        attempts: 2
        backoff: 100ms
```

## Conformance Details

Basic support for configuring retries in HTTPRoute up to a specified maximum count and with a minimum duration between attempts will be gated on the `SupportHTTPRouteRetry` feature. Retries will be tested in combination with request and backend timeouts if supported by the implementation.

Retrying requests based on HTTP status codes will be gated under the following features:

* `SupportHTTPRRouteRetryCodes`
  * Only 500, 502, 503 and 504 will be tested for conformance.
  * Traefik does not seem to support specifying error codes, and will only retry on backend timeouts.
* `SupportHTTPRRouteRetryCode5xx`
  * 500, 502, 503 and 504 will all be tested for conformance.
  * Arbitrary status codes in the 500-599 (inclusive) range will not be tested for conformance.

Implementations MAY support specifying additional individual error codes in the valid 100-599 (inclusive) range, invalid 600-999 (inclusive) range or prefix shorthands for any `[0-9]xx` range, but none of these will be tested in conformance.

## Alternatives

### Standard RetryPolicy metaresource

This may be a reasonable approach for configuring broad default retry policies, but the UX of using a separate resource could feel overly burdensome for basic use cases or granular configuration, and the structure of the policy attachment model is [currently being revised](https://github.com/kubernetes-sigs/gateway-api/discussions/2927). Additionally, multiple data plane proxies warn in their documentation about the danger of configuring broad retry policies which may cause logical application errors from replaying non-idempotent requests or overwhelm backends. This approach may still be considered to enable future goals.

### HTTPRoute filter

TODO

### What accommodations are needed for future retry budget support?

Should the `retry` stanza follow the Kubernetes "tagged union" pattern with something like a `mode: "count"` to allow future design space for `mode: "budget"` with distinct sibling fields?

### Should whether to retry on connection errors be configurable?

Retrying on connection errors (disconnect, reset, timeout, TCP failure) is typically default behavior for most dataplanes, even those that don't support configurable retries based on HTTP status codes. Some proxies allow this to be configurable, but retrying attempts on these sort of errors is typically desirable and excluding it from the Gateway API spec both allows a simpler UX with less boilerplate, and avoids nuances between how these are defined and configured in different proxies.

### Should whether to retry on a backend timeout be configurable?

On Kubernetes, retrying should _typically_ route a backend request to a different pod if the original destination has become unhealthy and therefore should generally be safe. Even if a [BackendLBPolicy](https://gateway-api.sigs.k8s.io/geps/gep-1619/) is configured, most dataplane implementations implement "soft" affinity rather than strict session routing. The warnings against this practice in NGINX and HAProxy documentation seem to reference risks with legacy deployment models using a small number of statically-defined servers.

## References

* <https://www.rfc-editor.org/rfc/rfc9110>
* <https://www.iana.org/assignments/http-status-codes/http-status-codes.xhtml>
* <https://gateway-api.sigs.k8s.io/geps/gep-2257/>
* <https://gateway-api.sigs.k8s.io/geps/gep-1742/>
* <https://github.com/kubernetes-sigs/gateway-api/issues/3139>
