# GEP: Standardized Telemetry API

* Issue: #4768
* Status: Provisional

## TLDR

This proposal introduces a standardized, provider-agnostic Telemetry API to configure observability signals (metrics, access logs, and traces) for both North/South (Gateway) and East/West (Mesh) traffic, addressing the fragmentation caused by vendor-specific CRDs.

## Goals

1. Decouple the intent of observability from the underlying vendor-specific implementation.
2. Provide uniform configurability for both Gateway and Mesh use cases.
3. Account for the fact that the persona responsible for the Gateway/Mesh infrastructure can be different from the persona responsible for dictating the structure and behavior of telemetry signals.
4. Empower platform and observability teams to enforce uniform telemetry standards across large-scale heterogeneous environments.
5. Ensure the API is suitable for the broader ecosystem beyond specialized networking.
6. Accommodate emerging standards and new protocols through a flexible and extensible API design.

## Non-Goals

1. Defining how the telemetry is exported (sinks/shippers) beyond specifying the provider endpoint.
2. Replacing the underlying telemetry infrastructure (OTLP collectors, Prometheus, etc.).
3. Standardizing metrics; this proposal exclusively focuses on the telemetry configuration API.

## Introduction / Overview

This GEP proposes the addition of a standardized, provider-agnostic Telemetry API to the Gateway API project. The proposal aims to define a unified configuration model for the generation and propagation of telemetry signals (i.e., metrics, access logs, distributed traces) for both North/South (Gateway) and East/West (Mesh) traffic.

The API focuses on providing a consistent way to express observability intent, such as sampling rates for tracing, metric customization, and log filtering, regardless of the underlying data plane implementation.

## Purpose (Why and Who)

### The Fragmentation of Observability

In the current Kubernetes landscape, the "Who, What, Where, and How Long" of network traffic is answered differently depending on the underlying proxy technology. While the Gateway API specification has unified how traffic is routed via `HTTPRoute` and `Gateway`, it has deferred the standardization of how that traffic is observed. This deferral has led to "Observability Lock-in". Platform Engineering teams are forced to learn and manage distinct APIs for each environment. A standardized telemetry API is necessary to decouple the intent of observability from the implementation. Without such standardization it is difficult for platform owners to:

1. Enforce consistent auditing standards across different infrastructure providers.
2. Manage "Mesh" and "Gateway" observability with a single unified API.
3. Support emerging workloads like AI Agents, which require specialized metrics (e.g., token usage, model latency) and detailed audit logs for tool-use verification.

### The Emergence of Agentic Networking

The most recent pressing driver for this proposal is the shift in traffic patterns introduced by agentic workloads. We are moving from a deterministic Service-to-Service paradigm to a non-deterministic Agent-to-Tool and Agent-to-Agent paradigm.

In an Agentic Mesh:

* Entities are Autonomous: Unlike traditional workloads, the runtime behavior and resulting network traffic of AI Agents are driven dynamically by a Large Language Model (LLM) rather than being statically defined within the application code.
* Cost is Volatile: Usage is measured in tokens, not just requests. A single HTTP 200 OK could cost $0.01 or $10.00 depending on the prompt and model used.
* Context is King: Debugging requires knowing the semantic context: Which Model? Which Prompt? Which tool?

Telemetry configuration must be flexible enough to account for emerging standards, such as Generative AI semantic conventions. AI traffic should not be treated as opaque TCP streams or standard HTTP requests. Without a standardized API to enable the extraction and export of bespoke attributes the "Agentic Mesh" risks remaining an observability blind spot.

### Who
- **Platform Operators**: Need to ensure uniform observability across all networking infrastructure.
- **Observability Teams**: Responsible for the governance of telemetry data. They need to define and enforce standardized schemas and collection policies across the entire organization.
- **Security/Auditing Teams**: Require a standardized audit trail for all traffic, especially for autonomous agent actions.
- **Application Developers**: Benefit from consistent metrics and traces for debugging without worrying about the underlying mesh or gateway technology.

## API

### Policy Attachment vs. Inline Configuration

A key area of discussion for this GEP is whether this should be a standalone Policy Attachment (e.g., `TelemetryPolicy`) or inline configuration within `Gateway` and `Mesh` resources. 

This proposal argues that the Policy Attachment model is the most effective approach to meet the stated goals, primarily for two reasons:

1. **Separation of Concerns**: It allows different personas to manage Gateway infrastructure (the Platform Team) independently from the configuration of telemetry signals (the Observability team).
2. **Fleet-Wide Uniformity**: It enables a single policy to be applied uniformly across a fleet of Gateways and Meshes, eliminating the need to duplicate complex telemetry configurations across individual resources.

To mitigate the challenge of defining merging semantics, this GEP restricts configuration such that only a single `TelemetryPolicy` can target a `Gateway` or `Mesh` at any given time. Attaching multiple `TelemetryPolicy` resources to the same target is considered out of scope for this specification and constitutes undefined behavior.

### High-level Considerations:

- **Tracing**: Configuration for OTLP endpoints, sampling rates (probabilistic and parent-based), and custom span attributes.
- **Metrics**: Ability to enable/disable specific metric families and customize dimensions (labels/attributes).
- **Access Logs**: Filtering for smart logging (e.g., only log 5xx errors or high latency) and field selection.

## Request Flow

* A platform operator creates a `TelemetryPolicy` resource targeting a `Gateway` or `Mesh`.
* The Gateway API implementation reconciles this resource and configures the underlying data plane.
* The data plane extracts the specified signals and exports them to the telemetry infrastructure.

