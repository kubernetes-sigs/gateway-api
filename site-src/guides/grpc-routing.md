# gRPC routing

!!! info "Experimental Channel"

    The `GRPCRoute` resource described below is currently only included in the
    "Experimental" channel of Gateway API. For more information on release
    channels, refer to the [related documentation](https://gateway-api.sigs.k8s.io/concepts/versioning).

The [GRPCRoute resource](/api-types/grpcroute) allows you to match on gRPC traffic and
direct it to Kubernetes backends. This guide shows how the GRPCRoute matches
traffic on host, header, and service, and method fields and forwards it to different
Kubernetes Services.

The following diagram describes a required traffic flow across three different
Services:

- Traffic to `foo.example.com` for the `com.Example.Login` method is forwarded to `foo-svc`
- Traffic to `bar.example.com` with an `env: canary` header is forwarded
to `bar-svc-canary` for all services and methods
- Traffic to `bar.example.com` without the header is forwarded to `bar-svc` for
  all services and methods

<!--- Editable source available at site-src/images/grpc-routing.png -->
![gRPC Routing](/images/grpc-routing.png)

The dotted lines show the `Gateway` resources deployed to configure this routing
behavior. There are two `GRPCRoute` resources that create routing rules on the
same `prod` Gateway. This illustrates how more than one Route can bind to a
Gateway which allows Routes to merge on a `Gateway` as long as they don't
conflict. `GRPCRoute` follows the same Route merging semantics. For more
information on that, refer to the [documentation](/api-types/httproute#merging).

In order to receive traffic from a [Gateway][gateway], a `GRPCRoute` resource
must be configured with `ParentRefs` which reference the parent gateway(s) that it
should be attached to. The following example shows how the combination
of `Gateway` and `GRPCRoute` would be configured to serve gRPC traffic:

```yaml
{% include 'experimental/v1alpha2/grpc-routing/gateway.yaml' %}
```

A `GRPCRoute` can match against a [single set of hostnames][spec].
These hostnames are matched before any other matching within the GRPCRoute takes
place. Since `foo.example.com` and `bar.example.com` are separate hosts with
different routing requirements, each is deployed as its own GRPCRoute -
`foo-route` and `bar-route`.

The following `foo-route` will match any traffic for `foo.example.com` and apply
its routing rules to forward the traffic to the correct backend. Since there is
only one match specified, only requests for the `com.example.User.Login` method to
`foo.example.com` will be forwarded. RPCs of any other method` will not be matched
by this Route.

```yaml
{% include 'experimental/v1alpha2/grpc-routing/foo-grpcroute.yaml' %}
```

Similarly, the `bar-route` GRPCRoute matches RPCs for `bar.example.com`. All
traffic for this hostname will be evaluated against the routing rules. The most
specific match will take precedence which means that any traffic with the `env:
canary` header will be forwarded to `bar-svc-canary` and if the header is
missing or does not have the value `canary` then it will be forwarded to `bar-svc`.

```yaml
{% include 'experimental/v1alpha2/grpc-routing/bar-grpcroute.yaml' %}
```

[gRPC
Reflection](https://github.com/grpc/grpc/blob/v1.49.1/doc/server-reflection.md)
is required to use interactive clients such as
[`grpcurl`](https://github.com/fullstorydev/grpcurl) without having a local copy
of the target service's protocol buffers present on your local filesysem. To
enable this, first ensure that you have a gRPC reflection server listening on
your application pods, then add the reflection method to your `GRPCRoute`. This
is likely to be useful in development and staging environments, but this should
be enabled in production environments only after the security implications have
been considered.

```yaml
{% include 'experimental/v1alpha2/grpc-routing/reflection-grpcroute.yaml' %}
```

[gateway]: /references/spec/#gateway.networking.k8s.io/v1beta1.Gateway
[spec]: /references/spec/#gateway.networking.k8s.io%2fv1alpha2.GRPCRouteSpec
[svc]:https://kubernetes.io/docs/concepts/services-networking/service/
