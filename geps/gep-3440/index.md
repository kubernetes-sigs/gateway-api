# GEP-3440: Gateway API Support for gRPC Retries

* Issue: [#3440](https://github.com/kubernetes-sigs/gateway-api/issues/3440)
* Status: Provisional

## TLDR
This proposal introduces support for gRPC retries in the Gateway API,
allowing for configuration of retry attempts, backoff duration, and retryable status codes for gRPC routes.

## Goals

- To allow specification of gRPC status codes that should be retried.
- To allow specification of the maximum number of times to retry a gRPC request.
- To allow specification of the minimum backoff interval between retry attempts for gRPC requests.
- Retry configuration must be applicable to most known Gateway API implementations for gRPC.
- To define any interaction with configured gRPC timeouts and backoff.

## Non-Goals

- No standard APIs for advanced retry logic, such as integrating with rate-limiting headers.
- No default retry policies for all routes within a namespace or for routes tied to a specific Gateway.
- No support for detailed backoff adjustments, like fine-tuning intervals, adding jitter, or setting max duration caps.
- No retry support for streaming or bidirectional APIs (maybe considered in future proposals).

## Introduction

To keep services reliable and resilient, a Gateway API implementation should be able to retry failed gRPC requests to
backend services before giving up and returning an error to clients.

Retries are helpful for several key reasons:
1. **Network failures**: Network issues can often cause temporary errors. Retrying a request helps to mitigate these
intermittent problems.
2. **Server-side failures**: Servers may fail temporarily due to overload or other issues.
Retrying allows requests to succeed once these conditions are resolved.
3. **Recovery from Temporary Errors**: Certain errors, like "Unavailable" or "resource-exhausted" are often short-lived.
Retrying can allow the request to complete once these issues clear up.

This proposal aims to establish a streamlined, consistent API for retrying gRPC requests, covering essential
functionality in a way that is broadly applicable across implementations.

## Background on implementations

Researching how different Gateway API implementations handle retries for gRPC requests.

### Envoy
Envoy supports retries for gRPC requests using the `retry_policy` field in the `route` configuration of the HTTP filter.
`retry_on` specifies the gRPC status codes that should trigger a retry by using `x-envoy-retry-grpc-on`,
and it supports a few built-in status codes like:
- `cancelled`: Envoy will attempt a retry if the gRPC status code in the response headers is “cancelled”.
- `deadline-exceeded`: Envoy will attempt a retry if the gRPC status code in the response headers is “deadline-exceeded”.
- `internal`: Envoy will attempt a retry if the gRPC status code in the response headers is “internal”.
- `resource-exhausted`: Envoy will attempt a retry if the gRPC status code in the response headers is “resource-exhausted”.
- `unavailable`: Envoy will attempt a retry if the gRPC status code in the response headers is “unavailable”.

As with the `x-envoy-retry-grpc-on` header, the number of retries can be controlled via the `x-envoy-max-retries` header.

By default, Envoy uses a fully jittered exponential backoff algorithm for retries.
This means that after a failed attempt, Envoy waits a random amount of time (with jitter) based on
an exponential growth pattern before trying again.
- **Default Timing**: The base interval starts at 25ms, and each subsequent retry can increase
this interval exponentially. By default, the maximum interval is capped at 250ms (10 times the base interval).
- **Per-Attempt Timeout (`per_try_timeout`)**: Envoy allows you to set a specific timeout for each retry attempt,
known as `per_try_timeout`. This timeout includes the initial request and each retry attempt.
If you don’t specify a `per_try_timeout`, Envoy uses the global route timeout for the total duration of the request.

In the Gateway API, this `per_try_timeout` will be equivalent to the BackendRequest timeout in the GRPCRouteRule.
This ensures that each retry attempt, including the initial one, respects the overall timeout defined for the backend
request, preventing retries from extending beyond the desired duration.

### Nginx
`ngx_http_grpc_module` in Nginx supports retries for gRPC requests using the `grpc_pass` directive.

For gRPC requests, Nginx allows retries under certain conditions by forwarding requests to another server in
an upstream pool when the initial request fails.
The following configuration options are available to control when and how retries occur:
1. **Retry Conditions** (`grpc_next_upstream`):
    Nginx can retry a request if certain issues are encountered, such as:
    - Network errors (e.g., connection or read errors).
    - Timeouts when establishing a connection or reading a response.
    - Invalid headers if the server sends an empty or malformed response.
    - Specific HTTP error codes (e.g., 500, 502, 503, 504, 429) can be configured as retryable for gRPC responses.
    By default, Nginx only retries on network error and timeout,
    but you can specify other conditions (like HTTP status codes) to expand retry options.
2. **Retry Limit by Time** (`grpc_next_upstream_timeout`):
   You can set a total time limit for how long Nginx will attempt retries.
   This limits the retry process to a specified time window, after which Nginx will stop attempting further retries.
3. **Retry Limit by Number** (`grpc_next_upstream_tries`):
    You can set a maximum number of retry attempts for a request.
    Once this limit is reached, Nginx will stop attempting further retries.
4. **Non-Idempotent Requests** (`non_idempotent`):
    By default, Nginx does not retry non-idempotent requests (like POST or PUT) because they can cause side effects
    if sent multiple times. However, you can enable retries for non-idempotent requests if needed.

**Important Considerations**:
- **Partial Responses**: Nginx can only retry if no part of the response has been sent to the client.
If an error occurs mid-response, retries are not possible.
- **Unsuccessful Attempts**: Errors like `timeout` and `invalid_header` are always considered unsuccessful and will
trigger retries if specified, while errors like `403` and `404` are not retryable by default.

### HAProxy
1. **Retry Conditions**: HAProxy can retry requests based on various network conditions
(e.g., connection failures, timeouts) and some HTTP error codes. While HAProxy does support gRPC via HTTP/2, it does not
have built-in support for handling specific gRPC error codes (like `Cancelled`, `Deadline Exceeded`).
It relies on HTTP-level conditions for retries, so its gRPC support is less granular than the GEP requires.
2. **Retry Limits**: HAProxy allows you to set a maximum number of retries for a request using the `retries` directive.
It also supports setting a timeout for the entire retry process using the `timeout connect` and `timeout server` directives.

### Traefik
1. **Retry Conditions**: Traefik allows for retries based on HTTP-level conditions (e.g., connection errors and
certain HTTP status codes like 500, 502, 503, and 504), but it does not natively interpret specific gRPC error codes
like `UNAVAILABLE` or `DEADLINE_EXCEEDED`. This means that, while Traefik can retry requests on common HTTP errors
that might represent temporary issues, it lacks the ability to directly handle and retry based on
gRPC-specific error codes, limiting its alignment with the GEP’s requirement for granular gRPC error handling.
2. **Retry Limits**: Traefik provides configurable retry attempts and can set a maximum number of retries. However,
Traefik does not offer per-try timeout controls specific to each retry attempt. Instead, it typically relies on a
global request timeout, limiting the flexibility needed for more precise gRPC retry management (like Envoy’s `per_try_timeout`).

## API
Having a dedicated API for gRPC retry conditions is necessary because gRPC uses
unique error codes (e.g., `UNAVAILABLE`, `DEADLINE_EXCEEDED`) that represent transient issues specific to its protocol,
which are not adequately covered by general HTTP status codes. gRPC also supports streaming and real-time communications,
making retry strategies more complex than those used for standard HTTP requests. Existing proxies like Envoy handle
gRPC retries with specialized logic, while other proxies rely on HTTP error codes, lacking the precision needed
for gRPC.

### Go

```go
type GRPCRouteRule struct {
    // Retry defines the configuration for when to retry a gRPC request.
    //
    // Support: Extended
    //
    // +optional
    // <gateway:experimental>
    Retry *GRPCRouteRetry `json:"retry,omitempty"`

    // ...
}

// GRPCRouteRetry defines retry configuration for a GRPCRoute.
//
// Implementations SHOULD retry on common transient gRPC errors
// if a retry configuration is specified.
//
type GRPCRouteRetry struct {
    // Reasons defines the gRPC error conditions for which a backend request
    // should be retried.
    //
    // Supported gRPC error conditions:
    // * "cancelled"
    // * "deadline-exceeded"
    // * "internal"
    // * "resource-exhausted"
    // * "unavailable"
    //
    // Implementations MUST support retrying requests for these conditions
    // when specified.
    //
    // Support: Extended
    //
    // +optional
    // <gateway:experimental>
    Reasons []GRPCRouteRetryCondition `json:"reasons,omitempty"`

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
    Attempts *int `json:"attempts,omitempty"`

    // Backoff specifies the minimum duration a Gateway should wait between
    // retry attempts, represented in Gateway API Duration formatting.
    //
    // For example, setting the `rules[].retry.backoff` field to `100ms`
    // will cause a backend request to be retried approximately 100 milliseconds
    // after timing out or receiving a specified retryable condition.
    //
    // Implementations MAY use an exponential or alternative backoff strategy,
    // MAY cap the maximum backoff duration, and MAY add jitter to stagger requests,
    // as long as unsuccessful backend requests are not retried before the configured
    // minimum duration.
    //
    // If a Request timeout (`rules[].timeouts.request`) is configured, the entire
    // duration of the initial request and any retry attempts MUST not exceed the
    // Request timeout. Ongoing retry attempts should be cancelled if this duration
    // is reached, and the Gateway MUST return a timeout error.
    //
    // Support: Extended
    //
    // +optional
    Backoff *Duration `json:"backoff,omitempty"`
}

// GRPCRouteRetryCondition defines a gRPC error condition for which a backend
// request should be retried.
//
// The following conditions are considered retryable:
//
// * "cancelled"
// * "deadline-exceeded"
// * "internal"
// * "resource-exhausted"
// * "unavailable"
//
// Implementations MAY support additional gRPC error codes if applicable.
//
// +kubebuilder:validation:Enum=cancelled;deadline-exceeded;internal;resource-exhausted;unavailable
type GRPCRouteRetryCondition string

// Duration is a string value representing a duration in time.
// Format follows GEP-2257, which is a subset of Golang's time.ParseDuration syntax.
//
// +kubebuilder:validation:Pattern=`^([0-9]{1,5}(h|m|s|ms)){1,4}$`
type Duration string
```

### YAML
```yaml
apiVersion: gateway.networking.k8s.io/v1
kind: GRPCRoute
metadata:
  name: foo-route
spec:
  parentRefs:
  - name: example-gateway
  hostnames:
  - "foo.example.com"
  rules:
  - matches:
    - method:
        service: com.example
        method: Login
    retry:
      reasons:
      - cancelled
      - deadline-exceeded
      - internal
      - resource-exhausted
      - unavailable
      attempts: 3
      backoff: 100ms
    backendRefs:
    - name: foo-svc
      port: 50051
```

## Conformance Details
To ensure correct gRPC retry functionality, the following tests must be implemented across Gateway API implementations:
1. `SupportGRPCRouteRetryBackendTimeout`
    - **Test**: Verify retries respect the BackendRequestTimeout. Requests should fail if the timeout is reached, even with retries.
    - **Expected**: Retries occur within the configured timeout, and fail if exceeded.
2. `SupportGRPCRouteRetry`
    - **Test**: Ensure retries are triggered for retryable gRPC errors (cancelled, deadline-exceeded, internal, resource-exhausted, unavailable).
    - **Expected**: Retries for retryable errors; no retries for non-retryable errors.
3. `SupportGRPCRouteRetryBackoff`
    - **Test**: Confirm retries use the configured backoff strategy.
    - **Expected**: Retries happen with increasing delay as per backoff configuration.

## Alternatives

### GRPCRoute filter
An alternative approach could be to introduce a new filter for GRPCRoute that handles retries. However, as we have already
established a `retry` field in the HTTPRouteRule, it makes sense to extend this to GRPCRoute for consistency.

## References

- [gRPC Retry Design](https://grpc.io/blog/guides/retry/)
- [gRPC Status Codes](https://grpc.io/docs/guides/error/)
- [Envoy Retry Policy](https://www.envoyproxy.io/docs/envoy/latest/api-v3/config/route/v3/route_components.proto#envoy-v3-api-msg-config-route-v3-retry-policy)
- [Nginx gRPC Module](https://nginx.org/en/docs/http/ngx_http_grpc_module.html)
- [HAProxy Retries](https://cbonte.github.io/haproxy-dconv/2.4/configuration.html#4.2-retries)
```
