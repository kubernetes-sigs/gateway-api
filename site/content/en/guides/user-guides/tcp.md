---
title: "TCP routing"
weight: 14
---

{{% alert color="info" title="Experimental Channel" %}}

The `TCPRoute` resource described below is currently only included in the
"Experimental" channel of Gateway API. For more information on release
channels, refer to our [versioning guide](/docs/concepts/versioning/).

{{% /alert %}}
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
apiVersion: gateway.networking.k8s.io/v1
kind: Gateway
metadata:
  name: my-tcp-gateway
spec:
  gatewayClassName: my-tcp-gateway-class
  listeners:
  - name: foo
    protocol: TCP
    port: 8080
    allowedRoutes:
      kinds:
      - kind: TCPRoute
  - name: bar
    protocol: TCP
    port: 8090
    allowedRoutes:
      kinds:
      - kind: TCPRoute
---
apiVersion: gateway.networking.k8s.io/v1alpha2
kind: TCPRoute
metadata:
  name: tcp-app-1
spec:
  parentRefs:
  - name: my-tcp-gateway
    sectionName: foo
  rules:
  - backendRefs:
    - name: my-foo-service
      port: 6000
---
apiVersion: gateway.networking.k8s.io/v1alpha2
kind: TCPRoute
metadata:
  name: tcp-app-2
spec:
  parentRefs:
  - name: my-tcp-gateway
    sectionName: bar
  rules:
  - backendRefs:
    - name: my-bar-service
      port: 6000
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

Note that you can achieve this same result by binding the Routes to the Gateway
listeners using the `port` field in the `parentRefs`:

```yaml
spec:
  parentRefs:
  - name: my-tcp-gateway
    port: 8080
```

Using the `port` field instead of `sectionName` for the attachment has the
downside of more tightly coupling the relationship between the Gateway and
its associated Routes. Refer to [Attaching to Gateways][attaching] for more
details.

[tcproute]: /reference/api-spec/main/spec/#tcproute
[tcp]: https://datatracker.ietf.org/doc/html/rfc793
[httproute]: /reference/api-spec/main/spec/#httproute
[gateway]: /reference/api-spec/main/spec/#gateway
[svc]: https://kubernetes.io/docs/concepts/services-networking/service/
[attaching]: /reference/api-types/httproute/#attaching-to-gateways
