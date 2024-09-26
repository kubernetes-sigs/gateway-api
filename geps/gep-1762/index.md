# GEP-1762: In Cluster Gateway Deployments

* Status: Standard

## Overview

Gateway API provides a common abstraction over different implementations, whether they are implemented by cloud load balancers, in-cluster deployments, or other mechanisms. However, many in-cluster implementations have solved some of the same problems in different ways.

Related discussions:

* [Support cluster-local Gateways](https://github.com/kubernetes-sigs/gateway-api/discussions/1247)
* [Scaling Gateway Resources](https://github.com/kubernetes-sigs/gateway-api/discussions/1355)
* [Manual deployments](https://github.com/kubernetes-sigs/gateway-api/issues/1687)
* [Merging Gateways](https://github.com/kubernetes-sigs/gateway-api/pull/1863/)
* [Per-Gateway Infrastructure](https://github.com/kubernetes-sigs/gateway-api/pull/1757)

## Goals

* Provide prescriptive guidance for how in-cluster implementations should behave.
* Provide requirements for how in-cluster implementations should behave.

Note that some changes will be suggestions, while others will be requirements.

## Non-Goals

* Provide guidance to how out-of-cluster implementations should behave. Rather, this document aims to bring consistency between these types.

## Terminology

This document uses a few terms throughout. To ensure consistency, they are defined below:

* In-cluster deployment: refers to an implementation that actuates a `Gateway` by running a data plane in the cluster.
  This is *often*, but not necessarily, by deploying a `Deployment`/`DaemonSet` and `Service`.
* Automated deployment: refers to an implementation that automatically deploys the data plane based on a `Gateway`.
  That is, the user simply creates a `Gateway` resource and the rest is handled behind the scenes by the implementation.

## Design

This GEP both introduces new API fields, and standardizes how implementations should behave when implementing the existing API.

### Automated Deployments

A simple `Gateway`, as is configured below is assumed to be an automated deployment:

```yaml
apiVersion: gateway.networking.k8s.io/v1beta1
kind: Gateway
metadata:
  name: my-gateway
spec:
  gatewayClassName: example
  listeners:
  - name: default
    port: 80
    protocol: HTTP
```

With this configuration, an implementation:

* MUST mark the Gateway as `Programmed` and provide an address in `Status.Addresses` where the Gateway can be reached on each configured port.
* MUST label all generated resources (Service, Deployment, etc) with `gateway.networking.k8s.io/gateway-name: my-gateway` (where `my-gateway` is the name of the Gateway resource).
* MUST provision generated resources in the same namespace as the Gateway if they are namespace scoped resources.
  * Cluster scoped resources are not recommended.
* SHOULD name all generated resources `my-gateway-example` (`<NAME>-<GATEWAY CLASS>`).
  This is not simply `NAME` to reduce the chance of conflicts with existing resources.
  Where required, this can also serve as the prefix for the object.

### Customizations

With any in-cluster deployment, customization requirements will arise. 

Some common requirements would be:

* `Service.spec.type`, to control whether a service is a `ClusterIP` or `LoadBalancer`.
* IP in the Service to assign to it.
* Arbitrary labels and annotations on generated resources.
* Any other arbitrary fields; the list is unbounded. Some examples would be:
  * CPU and memory requests
  * Service `externalTrafficPolicy`
  * Affinity rules

This GEP currently only aims to solve a subset of these concerns. Additional concerns may be addressed in future revisions or other GEPs.

#### Gateway Type

This is handled by [GEP-1651](https://github.com/kubernetes-sigs/gateway-api/pull/1653), so won't be described here.

#### Gateway IP

This section just clarifies an existing part of the spec, how to handle `.spec.addresses` for in-cluster implementations.
Like all other Gateway types, this should impact the address the `Gateway` is reachable at.

For implementations using a `Service`, this means the `clusterIP` or `loadBalancerIP` (depending on the `Service` type).

For example:

```yaml
apiVersion: gateway.networking.k8s.io/v1beta1
kind: Gateway
metadata:
  name: my-gateway
spec:
  addresses:
  - type: IPAddress
    value: 1.1.1.1
  gatewayClassName: example
  listeners:
  - name: default
    port: 80
    protocol: HTTP
```

This would generate a `Service` with `clusterIP` or `loadBalancerIP`, depending on the Service type.

#### Labels and Annotations

Labels and annotations for generated resources are specified in `infrastructure`:

```yaml
apiVersion: gateway.networking.k8s.io/v1beta1
kind: Gateway
metadata:
  name: my-gateway
spec:
  infrastructure:
    labels:
      foo: bar
    annotations:
      name: my-annotation
```

These are both `map[string]string` types, just like in `ObjectMeta`.

Any labels or annotations here are added to all generated resources.
Note this may mean an annotation intended for a `Service` may end up on a `Deployment` (for example).
This is typically not a concern; however, if an implementation is aware of specific meanings of certain labels or annotations, they MAY
exclude these from irrelevant resources.

This is intended to clearly identify resources associated with a specific application, environment, or Gateway.
Additionally, it can be used support integration with the kitchen-sink of Kubernetes extensions which rely on labels and annotations.

Validation will be added to prevent any usage with `gateway.networking.k8s.io/` prefix, to avoid conflicts with `gateway.networking.k8s.io/gateway-name` or other future additions.

#### Arbitrary Customization

GEP-1867 introduces a new `infrastructure` field, which allows customization of some common configurations (version, size, etc)
and allows a per-Gateway generic `parametersRef`.
This can be utilized for the remainder of customizations.

### Resource Attachment

Resources generated in response to the `Gateway` will have two attributes:

* A `gateway.networking.k8s.io/gateway-name: <NAME>` label.
* A name `<NAME>-<GATEWAY CLASS>`. This format is not strictly required for implementations, but strongly recommended for consistency in attachment.

The generated resources MUST be in the same namespaces as the `Gateway`.

Implementations MAY set `ownerReferences` to the `Gateway` in most cases, as well, but this is not required
as some implementations may have different cleanup mechanisms.

The `gateway.networking.k8s.io/gateway-name` label and standardize resource naming format can be relied on to attach resources to.
While "Policy attachment" in Gateway API would use attachment to the actual `Gateway` resource itself,
many existing resources attach only to resources like `Deployment` or `Service`.

An example using these:
```yaml
apiVersion: gateway.networking.k8s.io/v1beta1
kind: Gateway
metadata:
  name: gateway
spec:
  gatewayClassName: example
  listeners:
  - name: default
    hostname: "example.com"
    port: 80
    protocol: HTTP
---
apiVersion: autoscaling/v2
kind: HorizontalPodAutoscaler
metadata:
  name: gateway
spec:
  # Match the generated Deployment by reference
  # Note: Do not use `kind: Gateway`.
  scaleTargetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: gateway-example
  minReplicas: 2
  maxReplicas: 5
  metrics:
  - type: Resource
    resource:
      name: cpu
      target:
        type: Utilization
        averageUtilization: 50
---
apiVersion: policy/v1
kind: PodDisruptionBudget
metadata:
  name: gateway
spec:
  minAvailable: 1
  selector:
    # Match the generated Deployment by label
    matchLabels:
      gateway.networking.k8s.io/gateway-name: gateway
```

Note: there is [discussion](https://github.com/kubernetes-sigs/gateway-api/discussions/1355) around a way to attach a HPA to a Gateway directly.

## API

This GEP extends the `infrastructure` API introduced in [GEP-1867](https://gateway-api.sigs.k8s.io/geps/gep-1867).

```go
type GatewayInfrastructure struct {
	// Labels that should be applied to any resources created in response to this Gateway.
	//
	// For implementations creating other Kubernetes objects, this should be the `metadata.labels` field on resources.
	// For other implementations, this refers to any relevant (implementation specific) "labels" concepts.
	//
	// An implementation may chose to add additional implementation-specific labels as they see fit.
	//
	// Support: Extended
	// +kubebuilder:validation:MaxItems=8
	Labels map[string]string `json:"labels,omitempty"`
	// Annotations that should be applied to any resources created in response to this Gateway.
	//
	// For implementations creating other Kubernetes objects, this should be the `metadata.annotations` field on resources.
	// For other implementations, this refers to any relevant (implementation specific) "annotations" concepts.
	//
	// An implementation may chose to add additional implementation-specific annotations as they see fit.
	//
	// Support: Extended
	// +kubebuilder:validation:MaxItems=8
	Annotations map[string]string `json:"annotations,omitempty"`
	...
}
```

## Future Work

* Allow various policies, [such as HPA](https://github.com/kubernetes-sigs/gateway-api/discussions/1355), to attach directly to `Gateway` rather than just `Deployment`.
