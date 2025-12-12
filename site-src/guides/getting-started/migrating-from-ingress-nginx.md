# A Welcome Guide for Ingress-NGINX Users

Welcome! If you're an Ingress-NGINX user, you're in the right place. This page is a central hub of resources for those considering or actively migrating to Gateway API. We aim to continue to build out this page to be more comprehensive over time. If you're looking for a general overview of migrating from Ingress, please refer to our [Migrating from Ingress guide](./migrating-from-ingress.md).

We understand that migrations can be complex, and our goal is to provide you with the information and tools you need for a smooth transition.

## Common Questions

### How does Gateway API compare to Ingress?

Gateway API provides a role-oriented and more expressive API than Ingress. Where Ingress combines the concepts of a load balancer and routing rules into a single resource, Gateway API splits them apart:

* **Gateway:** Defines where and how traffic enters the cluster, a task for a cluster operator.
* **HTTPRoute:** Defines how traffic is routed to services, a task for an application developer.

This separation allows for safer, multi-tenant infrastructure. For a deeper dive, check the [Ingress Migration Guide](./migrating-from-ingress.md).

### How do I map Ingress-NGINX features to Gateway API?

Many annotation-based features in Ingress-NGINX have corresponding fields in Gateway API. For example, traffic splitting, header manipulation, and TLS configuration are all native to the API. For a detailed mapping, we recommend exploring the [Gateway API HTTPRoute documentation](../../../api-types/httproute/) and the documentation of your chosen [implementation](/implementations/).

### Can I try Gateway API without removing Ingress-NGINX?

Yes, and it's highly recommended. You can run a Gateway API controller alongside your existing Ingress-NGINX controller. They will each get a different external IP address, allowing you to test and validate your new configuration in isolation without affecting production traffic.

## Migration Resources

A successful migration requires careful planning. These resources are here to help.

### ingress2gateway

Manually translating complex Ingress rules and annotations is error-prone. The **[ingress2gateway](https://github.com/kubernetes-sigs/ingress2gateway)** tool is designed to automate this process. It reads your existing Ingress resources and converts them into the corresponding Gateway and HTTPRoute resources.

The tool is under active development, with ongoing work to support the most widely-used Ingress-NGINX annotations. We strongly recommend using it as the starting point for your migration.

### Choosing an Implementation

The first step is to select an implementation that fits your needs. Key factors to consider include:

* **Conformance:** Check the [conformance reports](https://gateway-api.sigs.k8s.io/implementations/) to ensure the implementation supports the Gateway API features you require.
* **Underlying Technology:** Your team's familiarity with a proxy like Envoy, NGINX, or others can influence your choice.
* **Integration:** Your cloud provider or CNI may already offer an integrated Gateway API implementation.

## Lots in Progress

We've got a lot of work in progress that we hope will help your migration to Gateway API even better in the future. In addition to our ongoing work to improve ingress2gateway and work towards a v1.0 release, we're also planning for the next release of Gateway API (currently targeting February). In that release, we're hoping to graduate the following features to GA:

* TLSRoute
* ListenerSet
* HTTPRoute CORS filter

If there are other features that we should be working on, please let us know.

## We're Here to Help

The Gateway API community is committed to making the migration experience as smooth as possible. The `ingress2gateway` tool is a key part of this, and we are actively working to improve its annotation support.

If you have questions, encounter issues, or are missing a feature, please get in touch:

* **File an issue** on the [Gateway API repository](https://github.com/kubernetes-sigs/gateway-api/issues).
* **Join a community meeting** to discuss your use case.
* **Provide feedback** on the `ingress2gateway` tool by opening an issue on its [repository](https://github.com/kubernetes-sigs/ingress2gateway/issues).

Your feedback is invaluable in helping us improve the API and the migration tooling for everyone.
