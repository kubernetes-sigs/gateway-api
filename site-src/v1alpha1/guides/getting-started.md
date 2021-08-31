# Getting started with Gateway APIs


**1.**  **[Install a Gateway controller](#installing-a-gateway-controller)**
 _OR_  **[install the Gateway API CRDs manually](#installing-gateway-api-crds-manually)**

_THEN_

**2.**   **Try out one of the available guides:**

- [Simple Gateway](/v1alpha1/guides/simple-gateway) (a good one to start out with)
- [HTTP routing](/v1alpha1/guides/http-routing)
- [HTTP traffic splitting](/v1alpha1/guides/traffic-splitting)
- [Routing across Namespaces](/v1alpha1/guides/multiple-ns)
- [Configuring TLS](/v1alpha1/guides/tls)
- [TCP routing](/v1alpha1/guides/tcp)

## Installing a Gateway controller

There are [multiple projects](references/implementations) that support the Gateway
API. By installing a Gateway controller in your Kubernetes cluster, you can
try out the guides above. This will demonstrate that the desired routing
configuration is actually being implemented by your Gateway resources (and the
network infrastructure that your Gateway resources represent). Note that many
of the Gateway controller setups will install and remove the Gateway API CRDs
for you.

## Installing Gateway API CRDs manually

The following command will install the Gateway API CRDs. This includes the
GatewayClass, Gateway, HTTPRoute, TCPRoute, and more. Note that a running
Gateway controller in your Kubernetes cluster is required to actually act on
these resources. Installing the CRDs will just allow you to see and apply the
resources, though they won't do anything.

```
kubectl kustomize "github.com/kubernetes-sigs/gateway-api/config/crd?ref=v0.3.0" \
| kubectl apply -f -
```

After you're done, you can clean up after yourself by uninstalling the
Gateway API CRDs. The following command will remove all GatewayClass, Gateway,
and associated resources in your cluster. If these resources are in-use or
if they were installed by a Gateway controller, then do not uninstall them.
This will uninstall the Gateway API CRDs for the entire cluster. Do not do
this if they might be in-use by someone else as this will break anything using
these resources.


```
kubectl kustomize "github.com/kubernetes-sigs/gateway-api/config/crd?ref=v0.3.0" \
| kubectl delete -f -
```

