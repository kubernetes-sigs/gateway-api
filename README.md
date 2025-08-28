# Kubernetes Gateway API

The Gateway API is a part of [SIG Network][sn], and this repository contains
the specification and Custom Resource Definitions (CRDs).

## Status

The latest supported version is `v1` as released by
the [v1.3.0 release][gh_release] of this project.

This version of the API is has GA level support for the following resources:

- `v1.GatewayClass`
- `v1.Gateway`
- `v1.HTTPRoute`
- `v1.GRPCRoute`

For all the other APIs and their support levels please consult [the spec][spec].

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

For a complete API reference, please refer to:

- [API reference][spec]
- [Go docs for the package][godoc]

## Gateway API conformance

If you are developing a Gateway API implementation and want to run conformance tests
against your project and eventually submit the proof of conformance, visit the [conformance
documentation][conformance-docs] for the test suite documentation, and the conformance
reports [readme][reports-readme] to see the reports submission rules. If you
are a user who wants to explore the features supported by the various implementations,
navigate the [conformance reports][conformance-reports]

## Contributing

Community meeting schedule, notes and developer guide can be found on the
[community page][cm].
Our Kubernetes Slack channel is [#sig-network-gateway-api][slack].

### Code of conduct

Participation in the Kubernetes community is governed by the
[Kubernetes Code of Conduct](code-of-conduct.md).

[ghp]: https://gateway-api.sigs.k8s.io/
[sn]: https://github.com/kubernetes/community/tree/master/sig-network
[cm]: https://gateway-api.sigs.k8s.io/contributing/community
[slack]: https://kubernetes.slack.com/messages/sig-network-gateway-api
[getting-started]: https://gateway-api.sigs.k8s.io/guides/
[spec]: https://gateway-api.sigs.k8s.io/reference/spec/
[concepts]: https://gateway-api.sigs.k8s.io/concepts/api-overview
[security-model]: https://gateway-api.sigs.k8s.io/concepts/security-model
[gh_release]: https://github.com/kubernetes-sigs/gateway-api/releases/tag/v1.3.0
[godoc]: https://pkg.go.dev/sigs.k8s.io/gateway-api
[conformance-docs]: https://gateway-api.sigs.k8s.io/concepts/conformance/
[reports-readme]: ./conformance/reports/README.md
[conformance-reports]: ./conformance/reports/
