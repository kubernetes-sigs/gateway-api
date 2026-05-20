---
title: "HTTP retries"
weight: 9
---
{{< details title="Experimental Channel Feature: HTTPRouteRetry" open="true" >}}
HTTPRoute retries are part of the Experimental Channel. For more information on release channels, refer to our [versioning guide](/docs/concepts/versioning/).
{{< /details >}}

The [HTTPRoute resource](/reference/api-types/httproute/) can be configured to
automatically retry unsuccessful requests to a backend before returning a
response to the client. This guide shows how to use this functionality.

## Retrying on HTTP status codes

The following `HTTPRoute` retries any request to `backend-svc` that
returns a 500, 502, 503, or 504, up to two additional attempts:

{{< readfile file="/examples/experimental/http-retries.yaml" code="true" lang="yaml" >}}

- `codes` lists the HTTP response status codes that trigger a retry.
  Implementations at least support `500`, `502`, `503`, and `504`.
- `attempts` is the maximum number of retries beyond the initial request.
- `backoff` is the minimum duration to wait between attempts, expressed using
  [Gateway API Duration formatting](/geps/gep-2257/).

## Retrying on connection errors

When the `retry` field is configured, implementations should also retry on
connection errors (disconnect, reset, timeout, TCP failure), even if no
`codes` are listed. The set of connection errors considered retriable is
implementation-specific, therefore check the implementation's documentation
for further details. The following rule retries connection failures up to two
times without retrying on any HTTP status code:

```yaml
  - backendRefs:
    - name: backend-svc
      port: 8080
    retry:
      attempts: 2
      backoff: 100ms
```

## Combining with timeouts

Retries interact with the [`timeouts`](/guides/http-timeouts/) on the same rule:

- If `timeouts.request` is set, the initial request *and* all retry attempts
  must complete within that duration. Retries still in progress when the
  request timeout expires should be canceled.
- If `timeouts.backendRequest` is set, any individual attempt that exceeds
  that duration without a response should be canceled, and the next retry
  should wait at least the configured `backoff` before being issued.

The following rule bounds each backend attempt to 500ms, the overall request
to 2s, and retries up to three times:

```yaml
  - backendRefs:
    - name: backend-svc
      port: 8080
    timeouts:
      request: 2s
      backendRequest: 500ms
    retry:
      codes:
      - 500
      - 502
      - 503
      - 504
      attempts: 3
      backoff: 100ms
```