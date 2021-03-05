Gateway API is designed to work with multiple protocols.
[TCPRoute](/spec/#networking.x-k8s.io/v1alpha1.TCPRoute) is one such route which
allows for managing TCP traffic.

In this example, we have one Gateway resource and two TCPRoute resources that
distribute the traffic with the following rules:

- All TCP streams on port 8080 of the Gateway are forwarded to port 6000 of
  `my-foo-service` Kubernetes Service.
- All TCP streams on port 8090 of the Gateway are forwarded to port 6000 of
  `my-bar-service` Kubernetes Service.

Please note the following:

- The `protocol` of listeners on the Gateway is `TCP`.
- Each listener selects exactly one TCPRoute. This is important since the routing
  decision is performed based on destination port only. If more metadata is used
  for routing decisions, then one may associate multiple TCPRoutes to a single
  Gateway listener. Implementations can support such use-cases by adding a custom
  resource to specify advanced routing properties and then referencing it in
  `spec.rules[].matches[].extensionRef`. Conflicts due to routing colisions should
  be resolved as per the [conflict resolution](guidelines.md#conflicts) guidelines.

```
{% include 'basic-tcp.yaml' %}
```
