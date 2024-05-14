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

There are [multiple projects](/implementations) that support Gateway API. By
installing a Gateway controller in your Kubernetes cluster, you can try out the
guides above. This will demonstrate that the desired routing configuration is
actually being implemented by your Gateway resources (and the network
infrastructure that your Gateway resources represent). Note that many of the
Gateway controller setups will install and remove the Gateway API bundle for
you.

## Installing Gateway API

!!! danger "Upgrades from earlier Experimental Channel releases"

    If you've previously installed an earlier version of experimental channel,
    refer to the [v1.1 upgrade notes](#v11-upgrade-notes).

A Gateway API bundle represents the set of CRDs associated with a version of
Gateway API. Each release includes two channels with different levels of
stability:

### Install Standard Channel

The standard release channel includes all resources that have graduated to GA or
beta, including GatewayClass, Gateway, HTTPRoute, and ReferenceGrant. To install
this channel, run the following kubectl command:

```bash
kubectl apply -f https://github.com/kubernetes-sigs/gateway-api/releases/download/v1.1.0/standard-install.yaml
```

### Install Experimental Channel

The experimental release channel includes everything in the standard release
channel plus some experimental resources and fields. This includes
TCPRoute, TLSRoute, UDPRoute and GRPCRoute.

Note that future releases of the API could include breaking changes to
experimental resources and fields. For example, any experimental resource or
field could be removed in a future release. For more information on the
experimental channel, refer to our [versioning
documentation](/concepts/versioning/).

To install the experimental channel, run the following kubectl command:

```bash
kubectl apply -f https://github.com/kubernetes-sigs/gateway-api/releases/download/v1.1.0/experimental-install.yaml
```

### v1.1 Upgrade Notes
If you are already using previous versions of GRPCRoute or BackendTLSPolicy
experimental channel CRDs from previous Gateway API releases, you'll need to be
careful with this upgrade. If you haven't installed Gateway API before, or have
exclusively used the standard channel of the API, you can skip the rest of this
section.

#### GRPCRoute
**Summary:** If you're already using v1alpha2 GRPCRoute, stick with the
experimental channel of GRPCRoute in v1.1 until the implementation(s) you're
relying on have been updated to support GRPCRoute v1.

**Explanation:** With the graduation of GRPCRoute to GA, it is now included in
standard channel. Unfortunately, that can be problematic for anyone that was
already using the experimental channel version of GRPCRoute. As a rule, CRDs in
standard channel do not expose alpha API version to avoid any version
deprecations in that channel. That means that the standard channel version of
GRPCRoute excludes v1alpha2. All implementations of GRPCRoute built before the
v1.1 release of Gateway API would have exclusively relied on v1alpha2 and will
need to be updated to support GRPCRoute v1. Until implementations have been
updated to support v1, you can safely upgrade to the experimental channel
version of GRPCRoute included in v1.1 that exposes both v1 and v1alpha2.

**Upgrade Sequence:** If you're already using v1alpha2 GRPCRoute, we'd recommend
the following upgrade sequence:

1. Install *experimental* v1.1 GRPCRoute CRD
2. Update all your manifests to use `v1` instead of `v1alpha2`
3. Upgrade to an implementation that supports GRPCRoute `v1` API Version
4. Install *standard* channel v1.1 GRPCRoute CRD

#### BackendTLSPolicy
**Summary:** If you've previously installed BackendTLSPolicy, wait until the
implementation(s) you're relying on have been updated to support `v1alpha3` of
the API. When upgrading to an implementation that supports `v1alpha3`, you'll
also need to uninstall the old BackendTLSPolicy CRD before installing the new
one.

**Explanation:** BackendTLSPolicy had several significant fields renamed in
v1.1, resulting in a version bump to v1alpha3. As this is experimental channel,
we are not providing an in-place upgrade path for this change, instead you'll
need to coordinate the CRD upgrade with the implementation(s) of
BackendTLSPolicy that you're relying on.

**Upgrade Sequence:** If you're already using v1alpha2 BackendTLSPolicy, we'd
recommend the following upgrade sequence:

1. Wait for your implementation of choice to release support for v1alpha3
2. Delete the older pre-v1.1 BackendTLSPolicy CRD (this will also delete all
   instances of BackendTLSPolicy in your cluster)
3. Install the new v1.1 BackendTLSPolicy CRD
4. Deploy the version of your implementation that supports BackendTLSPolicy v1alpha3

Note that some implementations may prefer switching the order of steps 3 and 4,
it's worth checking with any relevant release notes for your implementation of
choice.


### Cleanup

After you're done, you can clean up after yourself by uninstalling the Gateway
API CRDs by replacing "apply" with "delete" in the commands above. If these
resources are in-use or if they were installed by a Gateway controller, then do
not uninstall them. This will uninstall the Gateway API resources for the entire
cluster. Do not do this if they might be in-use by someone else as this will
break anything using these resources.

### More on CRD Management
This guide only provides a high level overview of how to get started with
Gateway API. For more on the topic of managing Gateway API CRDs, refer to our
[CRD Management Guide](/guides/crd-management).
