---
title: "Standardized Attribute Dictionary"
---

* Issue: [#5253](https://github.com/kubernetes-sigs/gateway-api/issues/5253)
* Status: Memorandum

## TLDR

Multiple emerging Gateway API proposals (e.g., `TelemetryPolicy` and `PayloadProcessor`) have a need to reference attributes. This memorandum establishes a standardized, vendor-neutral dictionary of attributes anchored on the [OpenTelemetry Semantic Conventions](https://opentelemetry.io/docs/specs/semconv/).

To guarantee portability, avoid semantic drift, and maintain backward compatibility, this dictionary defines canonical attribute keys and data types restricted exclusively to *stable* OpenTelemetry (OTel) conventions relevant to Gateway networking.

## Goals

* Establish a shared dictionary of attribute keys and data types for use across Gateway API specifications.
* Include only **stable** OTel attributes, so users inherit OTel's guarantees.
* Anchor the dictionary on stable OTel Semantic Conventions to eliminate fragmentation caused by vendor-specific proxy variables.

## Non-Goals

* Introducing any new Gateway API fields, CRDs, or resources.
* Defining experimental, provisional, or developmental OpenTelemetry attributes.
* Mandating that all Gateway API implementations support attribute extraction.

## Motivation

As the Gateway API evolves, configurations increasingly require referencing dynamic attributes for connections, requests, responses, etc. For example:

1. Telemetry: Enriching trace spans with request metadata.
2. Policy Evaluation & Expression Matching: Filtering, routing, or processing payloads in expressions based on runtime state.
3. Traffic Management: Header mutation, rewrites, and conditional rate limiting based on client or connection characteristics.

Without a common specification, each implementation exposes its own native syntax (e.g., `%REQ(...)%` in Envoy or `$remote_addr` in NGINX), leading to vendor lock-in and broken portability. By defining a canonical dictionary anchored on stable OpenTelemetry Semantic Conventions, Gateway API decouples user-facing intent from the underlying implementation.

## Design Principles

1. **Flat Key Namespace:** Each entry in this dictionary represents an individual fully-qualified attribute key (for example, `http.request.method` is valid, whereas an entire sub-tree object like `http.request` is not). This avoids complex object hierarchies and ensures predictable behavior in both key-value configurations and expressions evaluators.
2. **Strict Stability Boundary:** Only attributes formally categorized as **stable** in upstream OpenTelemetry are included. All other attributes are excluded from the dictionary to prevent breaking changes.
3. **Implementation Decoupling:** Implementations MUST translate these standard keys into their internal variables and MUST NOT expose vendor-specific command strings in portable APIs.

## The Attribute Dictionary

Gateway API resources and policies may reference the following attributes. For exact semantics, refer to the upstream [OpenTelemetry Semantic Conventions](https://opentelemetry.io/docs/specs/semconv/).

### HTTP

| Attribute | OTel Data Type |
| --- | --- |
| [http.request.method](https://opentelemetry.io/docs/specs/semconv/registry/attributes/http/#http-request-method) | string |
| [http.request.method_original](https://opentelemetry.io/docs/specs/semconv/registry/attributes/http/#http-request-method-original) | string |
| [http.response.status_code](https://opentelemetry.io/docs/specs/semconv/registry/attributes/http/#http-response-status-code) | int |
| [http.route](https://opentelemetry.io/docs/specs/semconv/registry/attributes/http/#http-route) | string |
| [http.request.header.<key>](https://opentelemetry.io/docs/specs/semconv/registry/attributes/http/#http-request-header) | string[] |
| [http.response.header.<key>](https://opentelemetry.io/docs/specs/semconv/registry/attributes/http/#http-response-header) | string[] |
| [user_agent.original](https://opentelemetry.io/docs/specs/semconv/registry/attributes/user-agent/) | string |
| [error.type](https://opentelemetry.io/docs/specs/semconv/registry/attributes/error/#error-type) | string |

### URL

| Attribute | OTel Data Type |
| --- | --- |
| [url.scheme](https://opentelemetry.io/docs/specs/semconv/registry/attributes/url/#url-scheme) | string |
| [url.path](https://opentelemetry.io/docs/specs/semconv/registry/attributes/url/#url-path) | string |
| [url.query](https://opentelemetry.io/docs/specs/semconv/registry/attributes/url/#url-query) | string |
| [url.full](https://opentelemetry.io/docs/specs/semconv/registry/attributes/url/#url-full) | string |
| [url.fragment](https://opentelemetry.io/docs/specs/semconv/registry/attributes/url/#url-fragment) | string |

### Connection

| Attribute | OTel Data Type |
| --- | --- |
| [client.address](https://opentelemetry.io/docs/specs/semconv/registry/attributes/client/#client-address) | string |
| [client.port](https://opentelemetry.io/docs/specs/semconv/registry/attributes/client/#client-port) | int |
| [server.address](https://opentelemetry.io/docs/specs/semconv/registry/attributes/server/#server-address) | string |
| [server.port](https://opentelemetry.io/docs/specs/semconv/registry/attributes/server/#server-port) | int |
| [network.peer.address](https://opentelemetry.io/docs/specs/semconv/registry/attributes/network/#network-peer-address) | string |
| [network.peer.port](https://opentelemetry.io/docs/specs/semconv/registry/attributes/network/#network-peer-port) | int |
| [network.protocol.name](https://opentelemetry.io/docs/specs/semconv/registry/attributes/network/#network-protocol-name) | string |
| [network.protocol.version](https://opentelemetry.io/docs/specs/semconv/registry/attributes/network/#network-protocol-version) | string |
| [network.transport](https://opentelemetry.io/docs/specs/semconv/registry/attributes/network/#network-transport) | string |
| [network.type](https://opentelemetry.io/docs/specs/semconv/registry/attributes/network/#network-type) | string |

### Infrastructure

| Attribute | OTel Data Type |
| --- | --- |
| [k8s.cluster.name](https://opentelemetry.io/docs/specs/semconv/registry/attributes/k8s/#k8s-cluster-name) | string |
| [k8s.cluster.uid](https://opentelemetry.io/docs/specs/semconv/registry/attributes/k8s/#k8s-cluster-uid) | string |

# Usage Guidelines Across Gateway API

1. **Cross-API Portability:** Any Gateway API feature that accepts attribute references SHOULD use the keys defined in this dictionary.
2. **Missing or Unavailable Attributes:** If an attribute is referenced in a context where it is not yet available (e.g., evaluating `http.response.status_code` during pre-routing request filtering), the evaluation engine SHOULD treat the attribute value as `null`, empty, or unset, rather than failing the transaction.
3. **Type Consistency:** Consumers of these attributes can rely on the data types defined above (e.g., `http.response.status_code` is always an integer; `http.request.method` is always a string).

## Attributes outside the dictionary

An API MAY permit keys outside the dictionary. Such keys carry no portability guarantee and SHOULD be documented as implementation-specific. Such keys MUST NOT collide with the OTel namespaces the dictionary has adopted. Vendor or organization keys MUST be namespaced, e.g. `example.com/tenant`.

Several **experimental** OTel attributes are directly relevant to Gateway API today, among them `tls.*` for Listener and BackendTLSPolicy observability, `rpc.*` for GRPCRoute, `gen_ai.*`, `mcp.*`, etc. Each becomes a candidate for admission when OTel stabilizes it.

# Lifecycle

* **Upstream Semantic Convention Changes:** If OpenTelemetry deprecates a stable attribute, Gateway API will retain support through a standard deprecation cycle before considering removal.
* **Handling Unsupported Attributes:** If a user configuration requests an attribute that an implementation does not support, the implementation SHOULD omit the attribute gracefully or surface a warning condition on the corresponding policy status. Implementations MUST NOT crash or fail unrelated routing operations.
