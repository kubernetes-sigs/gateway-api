# Getting started with Gateway APIs

**1.**  **[Install a Gateway controller](#installing-a-gateway-controller)**
 _OR_  **[install the Gateway API CRDs manually](#installing-a-gateway-api-bundle)**

_THEN_

**2.**   **Try out one of the available guides:**

- [Simple Gateway](/v1alpha2/guides/simple-gateway) (a good one to start out with)
- [HTTP routing](/v1alpha2/guides/http-routing)
- [HTTP traffic splitting](/v1alpha2/guides/traffic-splitting)
- [Routing across Namespaces](/v1alpha2/guides/multiple-ns)
- [Configuring TLS](/v1alpha2/guides/tls)
- [TCP routing](/v1alpha2/guides/tcp)

## Installing a Gateway controller

There are [multiple projects](/implementations) that support the
Gateway API. By installing a Gateway controller in your Kubernetes cluster,
you can try out the guides above. This will demonstrate that the desired routing
configuration is actually being implemented by your Gateway resources (and the
network infrastructure that your Gateway resources represent). Note that many
of the Gateway controller setups will install and remove the Gateway API bundle
for you.

## Installing a Gateway API Bundle

A Gateway API bundle represents the set of CRDs and validating webhook
associated with a version of Gateway API.

### Install the CRDs

The following command will install the Gateway API CRDs. This includes
GatewayClass, Gateway, HTTPRoute, TCPRoute, and more. Note that a running
Gateway controller in your Kubernetes cluster is required to actually act on
these resources. Installing the CRDs will just allow you to see and apply the
resources, but they won't do anything without a controller implementing them.

```
kubectl apply -k "github.com/kubernetes-sigs/gateway-api/config/crd?ref=v0.4.3"
```

### Install the Webhook

The validating webhook included with Gateway API is still in active development
and not as stable as other components included in the API. We expect this
webhook to reach a greater level of stability in an upcoming v0.4 release.
Until that point, the webhook can be installed with the following kubectl
commands:

```
kubectl apply -f https://raw.githubusercontent.com/kubernetes-sigs/gateway-api/v0.4.3/deploy/admission_webhook.yaml
kubectl apply -f https://raw.githubusercontent.com/kubernetes-sigs/gateway-api/v0.4.3/deploy/certificate_config.yaml
```

### Cleanup
After you're done, you can clean up after yourself by uninstalling the Gateway
API CRDs and webhook by replacing "apply" with "delete" in the commands above.
If these resources are in-use or if they were installed by a Gateway controller,
then do not uninstall them. This will uninstall the Gateway API resources for
the entire cluster. Do not do this if they might be in-use by someone else as
this will break anything using these resources.
