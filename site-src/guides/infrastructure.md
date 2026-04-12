# Gateway infrastructure labels and annotations

???+ info "Extended Support Feature: GatewayInfrastructurePropagation"
    This feature is part of extended support. For more information on support levels, refer to our [conformance guide](../concepts/conformance.md).

Gateway API implementations are responsible for creating the backing
infrastructure needed to make each Gateway work. For example, implementations
running in a Kubernetes cluster often create [Services][service] and
[Deployments][deployment], while cloud-based implementations may create cloud
load balancer resources. In many cases, it can be helpful to be able to
propagate labels or annotations to these generated resources.


The [`infrastructure` field][infrastructure] on a Gateway allows you to specify
labels and annotations for the infrastructure created by the Gateway API controller.
For example, if your Gateway infrastructure is running in-cluster, you can specify
both Linkerd and Istio injection using the following Gateway configuration, making
it simpler for the infrastructure to be incorporated into whichever service mesh
you've installed.

```yaml
apiVersion: gateway.networking.k8s.io/v1
kind: Gateway
metadata:
  name: meshed-gateway
  namespace: incoming
spec:
  gatewayClassName: meshed-gateway-class
  listeners:
  - name: http-listener
    protocol: HTTP
    port: 80
  infrastructure:
    labels:
      istio-injection: enabled
    annotations:
      linkerd.io/inject: enabled
```

[infrastructure]: ../reference/spec.md#gatewayinfrastructure
[service]: https://kubernetes.io/docs/concepts/services-networking/service/
[deployment]: https://kubernetes.io/docs/concepts/workloads/controllers/deployment/
