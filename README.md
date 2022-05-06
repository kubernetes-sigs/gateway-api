# Kubernetes Gateway API

The Gateway API is a part of the [SIG Network][sn], and this repository contains
the specification and Custom Resource Definitions (CRDs).

## Status

The latest supported version is v1alpha2 as released by the [v0.4.3
release](https://github.com/kubernetes-sigs/gateway-api/releases/tag/v0.4.3) of
this project. This version of the API is expected to graduate to beta in the
future with relatively minimal changes.

## Documentation

### Website

The API specification and detailed documentation is available on the project
website: [https://gateway-api.sigs.k8s.io][ghp].

### Concepts

To get started, please read through [API concepts][concepts] and
[Security model][security-model]. These documents give the necessary background
to understand the API and the use-cases it targets.

### Getting started

Once you have a good understanding of the API at a higher-level, check out 
[getting started][getting-started] to install your first Gateway controller and try out 
one of the guides. 

### References

A complete API reference, please refer to:

- [API reference][spec]
- [Go docs for the package](https://pkg.go.dev/sigs.k8s.io/gateway-api/apis/v1alpha1)

## Contributing

Community meeting schedule, notes and developer guide can be found on the
[community page][cm].
Our Kubernetes Slack channel is [#sig-network-gateway-api][slack].

## Technical Leads

- @bowei
- @thockin

### Code of conduct

Participation in the Kubernetes community is governed by the
[Kubernetes Code of Conduct](code-of-conduct.md).

[ghp]: https://gateway-api.sigs.k8s.io/
[sn]: https://github.com/kubernetes/community/tree/master/sig-network
[cm]: https://gateway-api.sigs.k8s.io/contributing/community
[slack]: https://kubernetes.slack.com/messages/sig-network-gateway-api
[getting-started]: https://gateway-api.sigs.k8s.io/v1alpha2/guides/getting-started/
[spec]: https://gateway-api.sigs.k8s.io/v1alpha2/references/spec
[concepts]: https://gateway-api.sigs.k8s.io/concepts/api-overview
[security-model]: https://gateway-api.sigs.k8s.io/concepts/security-model

