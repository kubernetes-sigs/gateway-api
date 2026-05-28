# Lightweight Endpoint Picker (LWEPP)

This package provides a minimal, lightweight reference implementation of the Endpoint Picker (EPP).

## Core Functions

- **Envoy Integration**: Implements the Envoy External Processing (ext_proc) protocol to:
  - Receive request headers and set the target endpoint header to guide Envoy's routing decision.
  - Receive response headers and add a header indicating which endpoint served the request.
- **Simple Load Balancing**: Performs basic round-robin load balancing across available pods in the target pool.

## Conformance Testing Support

The LWEPP includes two behaviors specifically to support the Gateway API Inference Extension conformance test suite.

### Header-Based Endpoint Filtering

Conformance tests need to steer individual requests to a specific backend pod in order to verify that routing works correctly. When the `test-epp-endpoint-selection` request header is present, the LWEPP restricts its candidate pool to only the pods whose IP addresses appear in the header (comma-separated). If none of the listed IPs match a known pod, or if the header is absent, the LWEPP falls back to round-robin across all available pods.

> **Note**: This header is only intended for use in test environments. It should not be present in production traffic.

### Served Endpoint Reporting

After a request is completed, conformance tests verify that the request was actually served by the pod the EPP selected. To support this, the LWEPP reads the `x-gateway-destination-endpoint-served` value from Envoy's response metadata (populated by a conforming gateway after routing to the backend) and writes it to the `x-conformance-test-served-endpoint` response header. The conformance test client reads this header to confirm correct routing.