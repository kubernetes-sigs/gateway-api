# Cloudflare Tunnel Gateway Controller

A Kubernetes controller that implements the Gateway API on top of Cloudflare Tunnel. It watches Gateway, HTTPRoute, and GRPCRoute resources and serves traffic through an in-process L7 reverse proxy that embeds the cloudflared tunnel transport, so all routing, matching, and filtering happen in-cluster.

## Table of Contents

| API channel | Implementation version | Mode | Report |
| --- | --- | --- | --- |
| standard | [v3.0.2](https://github.com/lexfrei/cloudflare-tunnel-gateway-controller/releases/tag/v3.0.2) | default | [standard-v3.0.2-default-report.yaml](./standard-v3.0.2-default-report.yaml) |

## Reproduce

The project ships a conformance harness (`hack/conformance-setup.sh`) that stands up a kind cluster, installs the Gateway API CRDs, builds and deploys the controller, and runs the suite end to end against a live Cloudflare Tunnel.

1. Clone the controller repository and check out the release under test:

   ```bash
   git clone https://github.com/lexfrei/cloudflare-tunnel-gateway-controller.git
   cd cloudflare-tunnel-gateway-controller
   git checkout v3.0.2
   ```

2. Provide a `.env` file in the repository root with the Cloudflare Tunnel credentials the harness reads: the account ID, an API token scoped to the tunnel, the tunnel ID, the tunnel token, and the edge hostname routing to that tunnel.

3. Run the harness against the standard channel. It creates a fresh kind cluster, installs the standard-channel Gateway API v1.5.1 CRDs, deploys the controller, runs the GATEWAY-HTTP and GATEWAY-GRPC profiles, and writes the report:

   ```bash
   CONTROLLER_VERSION=v3.0.2 \
   CONFORMANCE_REPORT_OUTPUT="$PWD/standard-v3.0.2-default-report.yaml" \
     ./hack/conformance-setup.sh --channel standard --test
   ```

4. Inspect the generated report:

   ```bash
   cat ./standard-v3.0.2-default-report.yaml
   ```

## Notes

- The tunnel transport runs over HTTP/2 (`proxy.tunnel.protocol=http2`, which the harness sets), because QUIC drops gRPC trailers; this is required for the GATEWAY-GRPC profile.
- The suite reaches the implementation through the Cloudflare edge over HTTPS. Test Host headers are carried in an `X-Original-Host` header so the edge forwards hostnames that are not registered on the account; this is a conformance-only shim, not a production pattern.
- Skipped tests reflect Cloudflare Tunnel semantics or upstream-suite constraints rather than controller gaps. TLS is terminated at the Cloudflare edge, the tunnel exposes a single port and flattens listeners (so HTTPS-listener, listener-port, and listener-isolation cases do not apply), and the gRPC-weight sampler and WebSocket-dialer cases are upstream-suite limitations with fixes filed upstream.
