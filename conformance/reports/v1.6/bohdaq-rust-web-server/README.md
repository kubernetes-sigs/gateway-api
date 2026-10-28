# rust-web-server

[rust-web-server](https://github.com/bohdaq/rust-web-server) (rws) is a web
server and application framework in Rust that runs as a Kubernetes Gateway API
controller when built with its `gateway-api` feature: one binary watches the
cluster, writes status and serves the traffic of HTTPRoutes, GRPCRoutes,
TLSRoutes, TCPRoutes and UDPRoutes. Documentation:
<https://rws8.tech/deployment/kubernetes-gateway-api/>.

## Table of contents

| API channel | Implementation version | Mode | Report |
|-------------|------------------------|------|--------|
| standard | [18.8.2](https://github.com/bohdaq/rust-web-server/releases/tag/18.8.2) | shared | [18.8.2 report](./standard-18.8.2-shared-report.yaml) |
| standard | [18.8.2](https://github.com/bohdaq/rust-web-server/releases/tag/18.8.2) | provisioned | [18.8.2 report](./standard-18.8.2-provisioned-report.yaml) |

## Modes

- **shared**: the rws Deployment that runs the controller (Helm
  `gatewayAPI.mode: shared`, the default) serves every Gateway of its
  GatewayClass. Each Gateway gets its own `LoadBalancer` Service and an
  EndpointSlice of those pods.
- **provisioned**: (Helm `gatewayAPI.mode: provisioned`) every Gateway gets its
  own Deployment of rws pods, created by the controller in its namespace; the
  controller's pods serve nothing.

## Reproduce

The reports come from the `Gateway API conformance` workflow of the rws
repository (`.github/workflows/gateway-api.yml`), run on the release tag; the
steps below are what it does.

1. Build rws at the release tag:

   ```sh
   git clone https://github.com/bohdaq/rust-web-server.git && cd rust-web-server
   git checkout tags/18.8.2
   cargo build --bin rws --features "acme auth mimalloc gateway-api"
   ```

   (The workflow builds the static `x86_64-unknown-linux-musl` binary the Linux
   release and container image ship; a glibc build behaves the same.)

2. Fetch the Gateway API suite and its standard-channel CRDs:

   ```sh
   git clone --depth 1 --branch v1.6.3 https://github.com/kubernetes-sigs/gateway-api /tmp/gateway-api
   gh release download v1.6.3 --repo kubernetes-sigs/gateway-api --pattern standard-install.yaml --dir /tmp/gateway-api
   ```

3. Run the suite (needs docker, [kind](https://kind.sigs.k8s.io) v0.33, kubectl,
   helm and Go 1.26): the script creates a kind cluster, installs the CRDs,
   MetalLB (on kind's Docker network) and rws with its Helm chart in the given
   mode, then runs the suite with every profile and feature rws claims:

   ```sh
   FEATURES=$(grep -oP '^  FEATURES: \K.*' .github/workflows/gateway-api.yml)
   for mode in shared provisioned; do
     ci/e2e/gateway_api.sh target/debug/rws /tmp/gateway-api "/tmp/report-$mode.yaml" \
       GATEWAY-HTTP,GATEWAY-GRPC,GATEWAY-TLS,GATEWAY-TCP,GATEWAY-UDP "$FEATURES" "$mode"
     kind delete cluster --name rws-gateway-api
   done
   ```

4. The reports are `/tmp/report-shared.yaml` and `/tmp/report-provisioned.yaml`.
