# Zentinel Gateway API Conformance

Zentinel is a security-first reverse proxy built on Cloudflare's Pingora framework. The `zentinel-gateway` controller implements the Kubernetes Gateway API, translating Gateway, HTTPRoute, GRPCRoute, TLSRoute, and Ingress resources into Zentinel proxy configuration.

## Implementation Details

| Field | Value |
|-------|-------|
| Organization | zentinelproxy |
| Project | [zentinel](https://github.com/zentinelproxy/zentinel) |
| Version | 0.6.1 |
| Controller Name | `zentinelproxy.io/gateway-controller` |
| GatewayClass Name | `zentinel` |
| Conformance Profile | Gateway HTTP |
| Gateway API Version | v1.2.1 |
| Channel | standard |

## How to Run

```bash
# Prerequisites: kind, kubectl, helm, go, docker

# Run the full conformance suite
./scripts/conformance-test.sh

# Or with report generation
./scripts/conformance-test.sh --report
```

## Contact

- GitHub: [@zentinelproxy](https://github.com/zentinelproxy)
- Website: [zentinelproxy.io](https://zentinelproxy.io)
