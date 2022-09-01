!!! info "Experimental Channel"

    The `TCPRoute` resource described below is currently only included in the
    "Experimental" channel of Gateway API. For more information on release
    channels, refer to the [related documentation](https://gateway-api.sigs.k8s.io/concepts/versioning).

Gateway API is designed to work with multiple protocols and [TCPRoute][tcproute]
is one such route which allows for managing [TCP][tcp] traffic.

In this example, we have one Gateway resource and two TCPRoute resources that
distribute the traffic with the following rules:

- All TCP streams on port 8080 of the Gateway are forwarded to port 6000 of
  `my-foo-service` Kubernetes Service.
- All TCP streams on port 8090 of the Gateway are forwarded to port 6000 of
  `my-bar-service` Kubernetes Service.

In this example two `TCP` listeners will be applied to the [Gateway][gateway]
in order to route them to two separate backend `TCPRoutes`, note that the
`protocol` set for the `listeners` on the `Gateway` is `TCP`:

```yaml
{% include 'experimental/basic-tcp.yaml' %}
```

In the above example we separate the traffic for the two separate backend TCP
[Services][svc] by using the `sectionName` field in the `parentRefs`:

```yaml
spec:
  parentRefs:
  - name: my-tcp-gateway
    sectionName: foo
```

This corresponds directly with the `name` in the `listeners` in the `Gateway`:

```yaml
  listeners:
  - name: foo
    protocol: TCP
    port: 8080
  - name: bar
    protocol: TCP
    port: 8090
```

In this way each `TCPRoute` "attaches" itself to a different port on the
`Gateway` so that the service `my-foo-service` is taking traffic for port `8080`
from outside the cluster and `my-bar-service` takes the port `8090` traffic.

[tcproute]:/references/spec/#gateway.networking.k8s.io/v1alpha2.TCPRoute
[tcp]:https://datatracker.ietf.org/doc/html/rfc793
[httproute]:/references/spec/#gateway.networking.k8s.io/v1alpha2.HTTPRoute
[gateway]:/references/spec/#gateway.networking.k8s.io/v1alpha2.Gateway
[svc]:https://kubernetes.io/docs/concepts/services-networking/service/
