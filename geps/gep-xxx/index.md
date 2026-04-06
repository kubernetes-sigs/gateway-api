# GEP XXX: EndpointSelector Resource

* Status: Provisional

## TLDR

The `EndpointSelector` resource provides a mechanism to target workloads for both routing and policy, usually through composition with higher-level resources like `Backend`. This GEP is designed to mirror an analgous proposal in upstream Kubernetes [INSERT KEP HERE], thus making the `gateway.networking.k8s.io` version of this resource an explicit stopgap that SHOULD NOT progress to the standard channel.

## Motivation

Historically, Kubernetes has tightly coupled the `Service` and `EndpointSlice` resources. The critital healthchecking and load balancing features of `EndpointSlice` have been inaccessible to users who want to target workloads without the overhead of a `Service`. The `EndpointSelector` resource decouples this relationship, allowing users to directly target workloads based on labels, without needing to create a `Service` as an intermediary.

## User Stories

TODO: Will likely just point to the KEP.

## Goals

## API

There are two main interfaces that will trigger the creation of an `EndpointSelector`:

1. When a user manually creates an `EndpointSelector` resource. In this case, the user opts-in to managing the lifecycle of the resource (e.g. nothing will set ownerReferences). Most of the time, users will choose this UX when wanting to decorate an existing `Service` with context or functionality described in a higher-level resource (e.g. `Backend` type `EndpointSelector` (TODO: confirm naming once `Backend` GEP is finalized)). These resources would reference the `EndpointSelector` via an object reference field in their spec. CRDs that utilize this pattern SHOULD add status to their APIs to report the validity of the reference.
2. A controller will create an `EndpointSelector` in response to a user creating a higher-level resource (e.g. `InferencePool`) with an inline selector field in its spec. In this case, the controller is responsible for managing the lifecycle of the `EndpointSelector` resource, including setting `ownerReferences` to ensure proper garbage collection when the owning resource is deleted. The controller MUST also ensure that the `EndpointSelector` is correctly configured based on the selector criteria specified in the higher-level resource. Furthermore, since multiple controllers may reconcile the same workload-selecting resource, they MUST add the `gateway.networking.k8s.io/managed-by` label and use `generateName` instead of `name` for the `EndpointSelector` to avoid naming conflicts and ensure that multiple controllers can create their own `EndpointSelector` resources for the same workload without interfering with each other. NOTE: all object references for workload-selecting resources MUST allow pluggable groups and MUST NOT default to the `gateway.networking.k8s.io` group as we expect the upstream Kubernetes API to become the canonical group in the future.

### Example (manual creation)

```yaml
apiVersion: gateway.networking.x-k8s.io/v1alpha1
kind: EndpointSelector
  metadata:
    name: myapp
    namespace: default
spec:
  selector:
    matchLabels:
      app: myapp
  ports:
    - name: http
      port: 80
      protocol: TCP
```

### Example (controller creation)

```yaml
apiVersion: inference.networking.k8s.io/v1
kind: InferencePool
metadata:
  name: vllm-qwen3-32b
spec:
  targetPorts:
    - number: 8000
  selector:
    app: vllm-qwen3-32b
  extensionRef:
    name: vllm-qwen3-32b-epp
    port: 9002
    failureMode: FailOpen
---
apiVersion: gateway.networking.x-k8s.io/v1alpha1
kind: EndpointSelector
metadata:
  generateName: vllm-qwen3-32b-
  namespace: default
  labels:
    gateway.networking.k8s.io/managed-by: inference-controller.io/gateway
spec:
  selector:
    matchLabels:
      app: vllm-qwen3-32b
  ports:
    - name: default
      port: 9002
      protocol: TCP
```

### Implementation Guide

**TODO (expand)**: Use the endpointslice-controller code in your controller implementation to create `EndpointSlices` directly since the native Kubernetes implementation will take some time to make its way to GA. This will allow us to iterate on the API and implementation in parallel with the upstream Kubernetes KEP. This feature should be configurable though a feature flag since it is not intended to be a long-term solution.
