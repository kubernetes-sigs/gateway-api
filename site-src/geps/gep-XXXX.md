# GEP-XXXX: GRPCRoute

* Issue: [TODO](https://github.com/kubernetes-sigs/gateway-api/issues/696)
* Status: Implementable

(See definitions in [Kubernetes KEP][kep-status].

[kep-status]: https://github.com/kubernetes/enhancements/blob/master/keps/NNNN-kep-template/kep.yaml#L9

## Goal

Add an idiomatic GRPCRoute for routing gRPC traffic.

## Non-Goals

While certain gRPC implementations support multiple transports and multiple
interface definition languages (IDLs), this proposal limits itself to
[HTTP/2](https://developers.google.com/web/fundamentals/performance/http2) as
the transport and [Protocol Buffers](https://developers.google.com/protocol-buffers)
as the IDL, which makes up the vast majority of gRPC traffic in the wild.

## Introduction

At the time of writing, the only official Route resource within the Gateway APIs
is HTTPRoute. It _is_ possible to support other protocols via CRDs and
controllers taking advantage of this have started to pop up. However, in the
long run, this leads to a fragmented ecosystem.

gRPC is a [popular RPC framework adopted widely across the industry](https://grpc.io/about/#whos-using-grpc-and-why).
The protocol is used pervasively within the Kubernetes project itself as the basis for
many interfaces, including:

- [the CSI](https://github.com/container-storage-interface/spec/blob/5b0d4540158a260cb3347ef1c87ede8600afb9bf/spec.md),
- [the CRI](https://github.com/kubernetes/cri-api/blob/49fe8b135f4556ea603b1b49470f8365b62f808e/README.md),
- [the device plugin framework](https://kubernetes.io/docs/concepts/extend-kubernetes/compute-storage-net/device-plugins/)

Given gRPC's importance in the application-layer networking space and to
the Kubernetes project in particular, we must ensure that the gRPC control plane
configuration landscape does not Balkanize.

## API

The API deviates from `HTTPRoute` where it results in a better UX for gRPC
users, while mirroring it in all other cases.

### Example `GRPCRoute`

```yaml
kind: GRPCRoute
apiVersion: gateway.networking.k8s.io/v1alpha2
metadata:
  name: foo-grpcroute
spec:
  parentRefs:
  - name: my-gateway
  hostnames:
  - foo.com
  - bar.com
  rules:
  - matches:
      method:
        grpcService: helloworld.Greeter
        grpcMethod:  SayHello
      headers:
      - type: Exact
        name: magic
        value: foo

    filters:
    - type: GRPCRequestHeaderModifierFilter
      add:
        - name: my-header
          value: foo

    - type: GRPCRequestMirrorPolicyFilter
      destination:
        backendRef:
          name: mirror-svc

    - type: GRPCTimeoutPolicyFilter
      timeout: "30s"

    - type: GRPCRouteRetryPolicyFilter
      numRetries: 3
      retryConditions:
      - "refused-stream"
      - "cancelled"

    backendRefs:
    - name: foo-v1
      weight: 90
    - name: foo-v2
      weight: 10
```
