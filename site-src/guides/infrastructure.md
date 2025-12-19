# Gateway infrastructure labels and annotations

??? success "Standard Channel since v1.2.0"

    The `infrastructure` field is GA and has been part of the Standard Channel since
    `v1.2.0`. For more information on release channels, refer to our [versioning
    guide](../concepts/versioning.md).

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
