# Lexfrei's Cloudflare Tunnel Gateway Controller

A Kubernetes controller that implements the Gateway API on top of Cloudflare Tunnel. It watches Gateway, HTTPRoute, and GRPCRoute resources and serves traffic through an in-process L7 reverse proxy that embeds the cloudflared tunnel transport, so all routing, matching, and filtering happen in-cluster.

## Table of Contents

| API channel | Implementation version | Mode | Report |
| --- | --- | --- | --- |
| standard | [v3.3.1](https://github.com/lexfrei/cloudflare-tunnel-gateway-controller/releases/tag/v3.3.1) | default | [standard-v3.3.1-default-report.yaml](./standard-v3.3.1-default-report.yaml) |

## Reproduce

The project ships a conformance harness (`hack/conformance-setup.sh`) that stands up a kind cluster, installs the Gateway API CRDs, builds and deploys the controller, and runs the suite end to end against a live Cloudflare Tunnel.

1. Clone the controller repository and check out the release under test:

   ```bash
   git clone https://github.com/lexfrei/cloudflare-tunnel-gateway-controller.git
   cd cloudflare-tunnel-gateway-controller
   git checkout v3.3.1
   ```

2. Provide a `.env` file in the repository root with the Cloudflare Tunnel credentials the harness reads: the account ID, an API token scoped to the tunnel, the tunnel ID, the tunnel token, and the edge hostname routing to that tunnel.

3. Run the harness against the standard channel. It creates a fresh kind cluster, installs the standard-channel Gateway API v1.6.1 CRDs, deploys the controller, runs the GATEWAY-HTTP and GATEWAY-GRPC profiles, and writes the report:

   ```bash
   CONTROLLER_VERSION=v3.3.1 \
   CONFORMANCE_REPORT_OUTPUT="$PWD/standard-v3.3.1-default-report.yaml" \
     ./hack/conformance-setup.sh --channel standard --test
   ```

4. Inspect the generated report:

   ```bash
   cat ./standard-v3.3.1-default-report.yaml
   ```

## Notes

- The tunnel transport runs over HTTP/2 (`proxy.tunnel.protocol=http2`, which the harness sets), because QUIC drops gRPC trailers; this is required for the GATEWAY-GRPC profile.
- The suite reaches the implementation through the Cloudflare edge over HTTPS. Test Host headers are carried in an `X-Original-Host` header so the edge forwards hostnames that are not registered on the account; this is a conformance-only shim, not a production pattern.
- The three skipped tests reflect Cloudflare Tunnel semantics, not controller gaps. TLS terminates at the Cloudflare edge, so the `HTTPRouteHTTPSListener` test cannot run: there is no in-cluster HTTPS listener. The tunnel exposes a single port (`HTTPRouteListenerPortMatching`), and its shared routing table flattens all listeners, so the multi-Gateway isolation case (`HTTPRouteMultipleGateways`) does not apply. Hard per-Gateway isolation requires the opt-in dedicated data plane, which this default-mode report does not exercise.
- `BackendTLSPolicy`, its SAN-validation companion, and `GatewayBackendClientCertificate` are implemented but listed unsupported here. Their tests are all gated on `SupportBackendTLSPolicy`, whose parent test needs an in-cluster HTTPS listener for the re-encrypt case that edge-terminated TLS cannot serve, so the feature is not claimed (see #5103).
