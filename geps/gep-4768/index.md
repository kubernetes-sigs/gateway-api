# GEP: Standardized Telemetry API

* Issue: #4768
* Status: Provisional

## TLDR

This proposal introduces a standardized, provider-agnostic Telemetry API to configure observability signals (metrics, access logs, and traces) for North/South (Gateway) traffic, addressing the fragmentation caused by vendor-specific CRDs.

## Goals

* Establish a standardized model to configure provider-agnostic telemetry (metrics, access logs, and traces) for Gateways.

## Non-Goals

1. Defining how the telemetry is exported (sinks/shippers) beyond specifying the provider endpoint and relevant connectivity parameters.
2. Replacing the underlying telemetry infrastructure (OTLP collectors, Prometheus, etc.).
3. Standardizing metrics; this proposal exclusively focuses on the telemetry configuration API.

## Introduction / Overview

This GEP proposes the addition of a standardized, provider-agnostic Telemetry API to the Gateway API project. The proposal aims to define a unified configuration model for the generation and propagation of telemetry signals (i.e., metrics, access logs, distributed traces) for North/South (Gateway) traffic.

The API focuses on providing a consistent way to express observability intent, such as sampling rates for tracing, metric customization, and log filtering, regardless of the underlying data plane implementation.

## Purpose (Why and Who)

### The Fragmentation of Observability

In the current Kubernetes landscape, the "Who, What, Where, and How Long" of network traffic is answered differently depending on the underlying proxy technology. While the Gateway API specification has unified how traffic is routed via `HTTPRoute` and `Gateway`, it has deferred the standardization of how that traffic is observed. This deferral has led to "Observability Lock-in". Platform Engineering teams are forced to learn and manage distinct APIs for each environment. A standardized telemetry API is necessary to decouple the intent of observability from the implementation. Without such standardization it is difficult for platform owners to:

1. Enforce consistent auditing and observability standards across different infrastructure providers.
2. Support emerging workloads like AI Agents, which elevate the criticality of observability due to their autonomous, non-deterministic nature and requirements for specialized signals.

### Who

- **Platform Operators**: Need to ensure uniform observability across all networking infrastructure.
- **Observability Teams**: Responsible for the governance of telemetry data. They need to define and enforce standardized schemas and collection policies across the entire organization.
- **Security/Auditing Teams**: Require a standardized audit trail for all traffic, an increasingly important need with the emergence of autonomous agent actions.
- **Application Developers**: Benefit from consistent metrics and traces for debugging without worrying about the underlying gateway technology.

## API

### Policy Attachment vs. Inline Configuration

A key area of discussion for this GEP is whether this should be a standalone Policy Attachment (e.g., `TelemetryPolicy`) or inline configuration within the `Gateway` resource. 

This proposal argues that the Policy Attachment model is the most effective approach to meet the stated goals, primarily for two reasons:

1. **Separation of Concerns**: It allows different personas to manage Gateway infrastructure independently from the configuration of telemetry signals.
2. **Uniformity**: It enables a single policy to be applied uniformly across a set of Gateways, eliminating the need to duplicate complex telemetry configurations across individual resources.

To mitigate the challenge of complex merging semantics, this GEP restricts configuration such that only a single `TelemetryPolicy` can target a specific `Gateway` at any given time. If multiple `TelemetryPolicy` resources target the same object, precedence is determined based on the creation timestamp.

### High-level Considerations:

- **Tracing**: Configuration for OTLP endpoints, sampling rates (probabilistic and parent-based), and custom resource/span attributes.
- **Metrics**: Ability to enable/disable specific metric families and customize dimensions (labels/attributes).
- **Access Logs**: Filtering for smart logging (e.g., only log 5xx errors or high latency), multi-protocol support, and log format customization (including field selection).
- **Export Configuration**: Supporting TLS connections to telemetry collectors and the ability to inject custom headers (e.g., `Authorization`) into telemetry requests.

## Request Flow

* A platform operator creates a `TelemetryPolicy` resource targeting a `Gateway`.
* The Gateway API implementation reconciles this resource and configures the underlying data plane.
* The data plane extracts the specified signals and exports them to the telemetry infrastructure.

