---
title: "UDP routing"
weight: 14
---

Gateway API is designed to work with multiple protocols and [UDPRoute][udproute]
is one such route which allows for managing [UDP][udp] traffic.

In this example, we have one Gateway resource and two UDPRoute resources that
distribute the traffic with the following rules:

- All UDP datagrams on port 8080 of the Gateway are forwarded to port 6000 of
  `my-foo-service` Kubernetes Service.
- All UDP datagrams on port 8090 of the Gateway are forwarded to port 6000 of
  `my-bar-service` Kubernetes Service.

In this example two `UDP` listeners will be applied to the [Gateway][gateway]
in order to route them to two separate backend `UDPRoutes`, note that the
`protocol` set for the `listeners` on the `Gateway` is `UDP`:

{{< readfile file="/examples/standard/basic-udp.yaml" code="true" lang="yaml" >}}

In the above example we separate the traffic for the two separate backend UDP
[Services][svc] by using the `sectionName` field in the `parentRefs`:

```yaml
spec:
  parentRefs:
    - name: my-udp-gateway
      sectionName: foo
```

This corresponds directly with the `name` in the `listeners` in the `Gateway`:

```yaml
listeners:
  - name: foo
    protocol: UDP
    port: 8080
  - name: bar
    protocol: UDP
    port: 8090
```

In this way each `UDPRoute` "attaches" itself to a different port on the
`Gateway` so that the service `my-foo-service` is taking traffic for port `8080`
from outside the cluster and `my-bar-service` takes the port `8090` traffic.

Note that you can achieve this same result by binding the Routes to the Gateway
listeners using the `port` field in the `parentRefs`:

```yaml
spec:
  parentRefs:
    - name: my-udp-gateway
      port: 8080
```

Using the `port` field instead of `sectionName` for the attachment has the
downside of more tightly coupling the relationship between the Gateway and
its associated Routes. Refer to [Attaching to Gateways][attaching] for more
details.

{{< details title="Note" color="info" >}}
You cannot attach a `UDPRoute` to an `HTTP` or `HTTPS` listener. A `UDPRoute`
can only attach to listeners using the `UDP` protocol. Attempting to attach a
`UDPRoute` to an `HTTP` or `HTTPS` listener will result in the route not being
accepted.
{{< /details >}}

[udproute]: /reference/api-spec/main/spec/#udproute
[udp]: https://datatracker.ietf.org/doc/html/rfc768
[gateway]: /reference/api-spec/main/spec/#gateway
[svc]: https://kubernetes.io/docs/concepts/services-networking/service/
[attaching]: /reference/api-types/udproute/#attaching-to-gateways
