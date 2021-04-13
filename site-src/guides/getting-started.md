# Getting started with Gateway APIs

## Installing CRDs

This project provides a collection of Custom Resource Definitions (CRDs) that can
be installed into any Kubernetes (>= 1.16) cluster.

To install the CRDs, please execute:

```
kubectl kustomize "github.com/kubernetes-sigs/gateway-api/config/crd?ref=v0.2.0" \
| kubectl apply -f -
```

## Install an implementation

[Multiple projects](implementations.md) implement the APIs defined by this
project.  You will need to either install an implementation or verify that one
is already setup for your cluster.

## Sample Gateway

Once you have the CRDs and an implementation installed, you are ready to
use Gateway API.

In this example, we are installing three resources:

- An `acme-lb` GatewayClass which is being managed by a `acme.io/gateway-controller`
  controller running in the cluster. Typically, a GatewayClass is provided by
  the implementation and must be installed in the cluster.
- A Gateway which is of type `acme-lb`:
    - This gateway has a single HTTP listener on port 80 which selects HTTPRoutes
      from all namespaces which have the label `app: foo` on them.

- Finally, we have an HTTPRoute resource which is attached to the above Gateway
  and has two rules:
    - All requests with path beginning with `/bar` are forwarded to my-service1
      Kubernetes Service.
    - All requests with path beginning with `/some/thing` AND have an HTTP header
      `magic: foo` are forwarded to my-service2 Kubernetes Service.

With this configuration, you now have a Gateway resource which is forwarding
traffic to two Kubernetes Services based on HTTP request metadata.

```
{% include 'basic-http.yaml' %}
```

For more advanced examples, please read the other [guides](/guides/index.md).

## Uninstalling the CRDs

To uninstall the CRDs and all resources created with them, run the following
command. Note that this will remove all GatewayClass and Gateway resources in
your cluster. If you have been using these resources for any other purpose do
not uninstall these CRDs.

```
kubectl kustomize "github.com/kubernetes-sigs/gateway-api/config/crd?ref=v0.1.0" \
| kubectl delete -f -
```
