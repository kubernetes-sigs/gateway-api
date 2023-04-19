# Getting started with Gateway API

**1.**  **[Install a Gateway controller](#installing-a-gateway-controller)**
 _OR_  **[install the Gateway API CRDs manually](#installing-gateway-api)**

_THEN_

**2.**   **Try out one of the available guides:**

- [Simple Gateway](/guides/simple-gateway) (a good one to start out with)
- [HTTP routing](/guides/http-routing)
- [HTTP redirects and rewrites](/guides/http-redirect-rewrite)
- [HTTP traffic splitting](/guides/traffic-splitting)
- [Routing across Namespaces](/guides/multiple-ns)
- [Configuring TLS](/guides/tls)
- [TCP routing](/guides/tcp)
- [gRPC routing](/guides/grpc-routing)
- [Migrating from Ingress](/guides/migrating-from-ingress)

## Installing a Gateway controller

There are [multiple projects](/implementations) that support the
Gateway API. By installing a Gateway controller in your Kubernetes cluster,
you can try out the guides above. This will demonstrate that the desired routing
configuration is actually being implemented by your Gateway resources (and the
network infrastructure that your Gateway resources represent). Note that many
of the Gateway controller setups will install and remove the Gateway API bundle
for you.

## Installing Gateway API

A Gateway API bundle represents the set of CRDs and validating webhook
associated with a version of Gateway API. Each release includes two
channels with different levels of stability:

### Install Standard Channel

The standard release channel includes all resources that have graduated to beta,
including GatewayClass, Gateway, ReferenceGrant, and HTTPRoute. To install this
channel, run the following kubectl command:

```
kubectl apply -f https://github.com/kubernetes-sigs/gateway-api/releases/download/v0.6.2/standard-install.yaml
```

### Install Experimental Channel

The experimental release channel includes everything in the standard release
channel plus some experimental resources and fields. This includes
TCPRoute, TLSRoute, UDPRoute and GRPCRoute. 

Note that future releases of the API could include breaking changes to
experimental resources and fields. For example, any experimental resource or
field could be removed in a future release. For more information on the
experimental channel, refer to our [versioning
documentation](https://gateway-api.sigs.k8s.io/concepts/versioning/).

To install the experimental channel, run the following kubectl command:

```
kubectl apply -f https://github.com/kubernetes-sigs/gateway-api/releases/download/v0.6.2/experimental-install.yaml
```

### Cleanup
After you're done, you can clean up after yourself by uninstalling the Gateway
API CRDs and webhook by replacing "apply" with "delete" in the commands above.
If these resources are in-use or if they were installed by a Gateway controller,
then do not uninstall them. This will uninstall the Gateway API resources for
the entire cluster. Do not do this if they might be in-use by someone else as
this will break anything using these resources.
