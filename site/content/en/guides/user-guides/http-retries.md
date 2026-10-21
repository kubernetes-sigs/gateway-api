---
title: "HTTP retries"
weight: 9
---
The [HTTPRoute resource](/reference/api-types/httproute/) can be configured to
automatically retry unsuccessful requests to a backend before returning a
response to the client. This is useful for hiding transient backend failures
(such as an HTTP `503` error during a rolling update, or a dropped connection)
from end users without requiring changes to client applications.

The `retry` field in an HTTPRouteRule can be used to specify which failures
should be retried and how many times.

## Retrying on HTTP status

{{< details title="Extended Support Feature: HTTPRouteRetryCodes" open="true" >}}
This feature is part of extended support. For more information on support levels, refer to our [conformance guide](docs/concepts/conformance/).
{{< /details >}}

The following `HTTPRoute` retries any request to `backend-svc` that
returns a `500`, `503`, or `504`, up to two additional attempts:

{{< readfile file="/examples/standard/http-retry/codes.yaml" code="true" lang="yaml" >}}

- `codes` lists the HTTP response status codes that trigger a retry.
  The codes `500`, `502`, `503`, and `504` are always supported, support
  for other codes varies by gateway, so consult the implementation's
  documentation before relying on them.
- `attempts` is the maximum number of retries beyond the initial request.
  With `attempts: 2`, the gateway may send up to three requests to the
  backend in total. Setting `attempts: 0` disables retries. If `attempts`
  is unspecified, it defaults to `1`.

## Retrying on connection errors

{{< details title="Extended Support Feature: HTTPRouteRetryConnectionError" open="true" >}}
This feature is part of extended support. For more information on support levels, refer to our [conformance guide](docs/concepts/conformance/).
{{< /details >}}

When the `retry` field is configured, connection errors (disconnect, reset,
timeout, TCP failure) are also retried, even if no `codes` are listed. The
exact set of connection errors that count as retriable varies by gateway,
so consult the implementation's documentation for the precise behaviour.

The following rule retries connection failures up to two times without
retrying on any HTTP status code. This is useful when the backend's status
codes should be returned to the client as-is, and only network-level
failures should be masked:

{{< readfile file="/examples/standard/http-retry/connection-error.yaml" code="true" lang="yaml" >}}

## Combining with timeouts

{{< details title="Extended Support Feature: HTTPRouteRequestTimeout, HTTPRouteRetryBackendTimeout" open="true" >}}
This feature is part of extended support. For more information on support levels, refer to our [conformance guide](docs/concepts/conformance/).
{{< /details >}}

Retries interact with the [`timeouts`](/guides/http-timeouts/) on the same rule:

- If `timeouts.request` is set, the initial request *and* all retry attempts
  must complete within that duration. Retries still in progress when the
  request timeout expires should be canceled.
- If `timeouts.backendRequest` is set, any individual attempt that exceeds
  that duration without a response should be canceled.

The following rule bounds each backend attempt to 500ms, the overall request
to 2s, and retries up to three times:

{{< readfile file="/examples/standard/http-retry/timeouts.yaml" code="true" lang="yaml" >}}

## Configuring backoff between retries

{{< details title="Implementation-specific Feature" open="true" >}}
This feature is part of implementation-specific support. For more information on support levels, refer to our [conformance guide](docs/concepts/conformance/).
{{< /details >}}

The `backoff` field sets the base interval the gateway waits before each
retry attempt (for example `100ms` or `1s`). The actual delay is derived from
this base interval by the implementation's retry strategy, typically exponential
backoff with random jitter to avoid many clients retrying at the same moment.
The delay may also be capped at an implementation-defined maximum. If `backoff`
is unspecified, the strategy and base interval are implementation-specific.

The following rule retries `503` responses up to three times, with the delay
between retries derived from a base backoff interval of `100ms`:

{{< readfile file="/examples/standard/http-retry/backoff.yaml" code="true" lang="yaml" >}}
