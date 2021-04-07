# Kubernetes Gateway API

The Gateway API is a part of the [SIG Network][sn], and this repository contains
the specification and Custom Resource Definitions (CRDs).

*Note: This project was previously named "Service APIs" until being renamed to
"Gateway API" in February 2021.*

## Documentation

### Website

The API specification and detailed documentation is available on the project
website: [https://gateway-api.sigs.k8s.io][ghp].

### Get started

To get started, please read through [API concepts][concepts] and
[Security model][security-model]. These documents give the necessary background
to understand the API and the use-cases it targets.

### Guides

Once you have a good understanding of the API at a higher-level, please
follow one of our [guides][guides] to dive deeper into different parts of
the API.

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
[guides]: https://gateway-api.sigs.k8s.io/guides
[spec]: https://gateway-api.sigs.k8s.io/references/spec
[concepts]: https://gateway-api.sigs.k8s.io/concepts/api-overview
[security-model]: https://gateway-api.sigs.k8s.io/concepts/security-model

