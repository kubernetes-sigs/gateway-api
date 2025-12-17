# Frequently Asked Questions (FAQ)

#### How can I get involved with Gateway API?

The [community](contributing/index.md) page keeps track of how to get involved
with the project.

#### Will Gateway API replace the Ingress API?
No. The Ingress API is GA since Kubernetes 1.19. There are no plans to deprecate
this API and we expect most Ingress controllers to support it indefinitely.

#### What are the differences between Ingress and Gateway API?
Ingress primarily targets exposing HTTP applications with a simple, declarative
syntax. Gateway API exposes a more general API for proxying that can be used for
more protocols than just HTTP, and models more infrastructure components to
provide better deployment and management options for cluster operators.

For more information, see the [Migrating from
Ingress](/guides/getting-started/migrating-from-ingress/) guide.

#### Will there be a default controller implementation?
No. There are already many great [implementations](implementations.md) to choose
from. The scope of this project is to define the API, conformance tests, and
overall documentation.

#### How can I expose custom capabilities through Gateway API?
There are a few mechanisms available for extending the API with
implementation-specific capabilities:

* The [Policy Attachment](reference/policy-attachment.md) model allows you to
decorate Gateway API objects with implementation-specific CRDs. A policy or
configuration object could match the Gateway API object either by name or by
using an explicit object reference.

* Use implementation-specific values for string fields in Gateway API resources.

* As a last resort, use implementation-specific annotations on Gateway API
  objects.

* Use API-defined extension points. Some Gateway API objects have explicit
[extension points](concepts/api-overview.md#extension-points) for implementations
to use.

#### Where can I find Gateway API releases?
Gateway API releases are tags of the [GitHub
repository](https://github.com/kubernetes-sigs/gateway-api). The [GitHub
releases](https://github.com/kubernetes-sigs/gateway-api/releases) page shows
all the releases.

#### How should I think about alpha API versions?
Similar to upstream Kubernetes, alpha API versions indicate that resources are
still experimental in nature and may either be removed or changed in breaking
ways in future releases of Gateway API.

See the [Versioning](concepts/versioning.md) documentation for more info.

#### Which Kubernetes versions are supported?
See our policy on [supported
versions](concepts/versioning.md#supported-versions).

#### Is SSL Passthrough supported?
SSL Passthrough (wherein a Gateway routes traffic with the [Transport Layer
Security (TLS)](https://en.wikipedia.org/wiki/Transport_Layer_Security)
encryption _intact_ to a backend service instead of terminating it) is supported
by [TLSRoutes](concepts/api-overview.md#tlsroute). See the [TLS
Guide](guides/tls.md) for more details about passthrough and other TLS
configurations.

#### What's the difference between Gateway API and an API Gateway?
An [API gateway](https://glossary.cncf.io/api-gateway/) is a tool that
aggregates unique application APIs, making them all available in one place.
It allows organizations to move key functions, such as authentication and
authorization or limiting the number of requests between applications, to a
centrally managed location. An API gateway functions as a common interface to (often external) API consumers.

Gateway API is an interface, defined as a set of Kubernetes resources, that 
models service networking in Kubernetes. One of the main resources is a `Gateway`,
which declares the Gateway type (or class) to instantiate and its configuration.
As a Gateway provider, you can implement Gateway API to model Kubernetes service
networking in an expressive, extensible, and role-oriented way.

Some API gateways can be programmed using the Gateway API.

#### Is Gateway API a standard for API Management?
No. API Management is a much broader concept than what Gateway API aims to be,
or what an API Gateway is intended to provide. An API Gateway can be an
essential part of an API Management solution. Gateway API can be seen as a way
to standardize provisioning of API Gateways.
