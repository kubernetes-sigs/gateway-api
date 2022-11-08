---
description: >
  We’re pleased to announce the release of v1alpha2, our second alpha API
  version. This release includes some major changes and improvements. This post
  will cover the highlights.
---

# Introducing Gateway API v1alpha2

<small style="position:relative; top:-30px;">
  :octicons-calendar-24: October 14, 2021 ·
  :octicons-clock-24: 5 min read
</small>

We’re pleased to announce the release of v1alpha2, our second alpha API version.
This release includes some major changes and improvements. This post will cover
the highlights.

## Highlights

### New API Group
To recognize our status as an official Kubernetes API, we've transitioned from
an experimental API group (`networking.x-k8s.io`) to the new
`gateway.networking.k8s.io` API group. This means that, as far as the apiserver
is concerned, this version is wholly distinct from v1alpha1, and automatic
conversion is not possible.

![New API group for v1alpha2](/images/v1alpha2-group.png)

### Simpler Route-Gateway Binding
In v1alpha1 we provided many ways to connect Gateways and Routes. This was a bit
more complicated to understand than we'd like. With v1alpha2, we've focused on
simpler attachment mechanism: 

* Routes directly reference the Gateway(s) they want to attach to. This is a
  list, so a Route can attach to more than one Gateway.
* Each Gateway listener can choose to specify the kinds of Routes they support
  and where they can be. This defaults to Routes that support the specified
  protocol in the same Namespace as the Gateway. 

For example, the following HTTPRoute uses the `parentRefs` field to attach
itself to the `prod-web-gw` Gateway.

```yaml
apiVersion: gateway.networking.k8s.io/v1alpha2
kind: HTTPRoute
metadata:
  name: foo
spec:
  parentRefs:
  - name: prod-web
  rules:
  - backendRefs:
    - name: foo-svc
      port: 8080
```

This is covered in more detail in [GEP 724](https://gateway-api.sigs.k8s.io/geps/gep-709/).

### Safe Cross Namespace References

!!! info "Experimental Channel"

    The `ReferenceGrant` resource described below is currently only included in the
    "Experimental" channel of Gateway API. For more information on release
    channels, refer to the [related documentation](https://gateway-api.sigs.k8s.io/concepts/versioning).

It is quite challenging to cross namespace boundaries in a safe manner. With
Gateway API, we had several key feature requests that required this capability.
Most notably, forwarding traffic to backends in other namespaces and referring
to TLS certificates in other namespaces.

To accomplish this, we've introduced a new ReferenceGrant resource that
provides a handshake mechanism. By default, references across namespaces are not
permitted; creating a reference across a namespace (like a Route referencing a
Service in another namespace) must be rejected by implementations. These
references can be accepted by creating a ReferenceGrant in the referent
(target) namespace, that specifies what Kind is allowed to accept incoming
references, and from what namespace and Kind the references may be.

For example, the following ReferenceGrant would allow HTTPRoutes in the prod
namespace to forward traffic to Services wherever this ReferenceGrant was
installed:

```yaml
apiVersion: gateway.networking.k8s.io/v1alpha2
kind: ReferenceGrant
metadata:
  name: allow-prod-traffic
spec:
  from:
  - group: gateway.networking.k8s.io
    kind: HTTPRoute
    namespace: prod
  to:
  - group: ""
    kind: Service
```

This is covered in more detail in [GEP 709](https://gateway-api.sigs.k8s.io/geps/gep-709/).

### Policy Attachment
One of the key goals of this API is to provide meaningful and consistent
extension points. In v1alpha2, we've introduced a new standard for attaching
policies to Gateway API resources.

What is a policy? Well, it's kind of up to the implementations, but the best
example to begin with is timeout policy.

Timeout policy for HTTP connections is highly dependent on how the underlying
implementation handles policy - it's very difficult to extract commonalities.

This is intended to allow things like:

* Attaching a policy that specifies the default connection timeout for backends
  to a GatewayClass. All Gateways that are part of that Class will have Routes
  get that default connection timeout unless they specify differently.
* If a Gateway that's a member of the GatewayClass has a different default
  attached, then that will beat the GatewayClass (for defaults, more specific
  object beats less specific object).
* Alternatively, a Policy that mandates that you can't set the client timeout to
  "no timeout" can be attached to a GatewayClass as an override. An override
  will always take effect, with less specific beating more specific.

As a simple example, a TimeoutPolicy may be attached to a Gateway. The effects
of that policy would cascade down to Routes attached to that policy:

![Simple Ingress Example](/images/policy/ingress-simple.png)

This is covered in more detail in [GEP 713](https://gateway-api.sigs.k8s.io/geps/gep-713/).

## Next Steps
There are a lot of changes in v1alpha2 that we haven't covered here. For the
full changelog, refer to our [v0.4.0 release
notes](https://github.com/kubernetes-sigs/gateway-api/releases/tag/v0.4.0). 

Many of our [implementations](/implementations) are planning to release support
for the v1alpha2 API in the coming weeks. We'll update our documentation as
v1alpha2 implementations become available.

We still have lots more to work on. Some of our next items to discuss include:

* Conformance testing
* Route delegation
* Rewrite support
* L4 Route matching

If these kinds of topics interest you, we'd love to have your input. Refer to
our [community page](/contributing/community) to see how you can get involved.
