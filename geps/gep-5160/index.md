---
title: "GEP: Standardized Rate Limit API"
---

* Issue: #5160
* Status: Provisional

## TLDR

This proposal introduces a standardized, provider-agnostic Rate Limit API for defining rate limit buckets, expressions for assigning requests to buckets and optionally request costs computation, addressing the fragmentation caused by vendor-specific CRDs.

## Goals

* Establish a standardized model for configuring provider-agnostic rate limit policies for Gateways.

## Non-Goals

1. Defining the implementation of the global rate limit state shared by multiple Gateway replicas.
2. Defining API or metrics for human or machine observability of the current rate limit state.

## Introduction / Overview

This GEP proposes the addition of a standardized, provider-agnostic Rate Limit API to the Gateway API project. The proposal aims to define a unified configuration model for the definition of rate limit buckets, expressions for assigning requests to buckets and optionally request costs computation.

The API focuses on providing a consistent way to express intent of rate limit policies regardless of the underlying data plane implementation.

## Purpose (Why and Who)

### The Fragmentation of Rate Limit Policies

While the Gateway API specification has unified how traffic is routed via `HTTPRoute` and `Gateway`, there is no standard way of configuring rate limit policies for traffic flowing through Gateway. The absence of
vendor agnostic rate limit APIs have led to multiple vendor or project specific APIs. Platform Engineering teams are forced to learn and manage distinct APIs for each implementation. A standardized API is necessary to decouple the intent of observability from the implementation. Without such standardization it is difficult for platform owners to:

1. Enforce consistent rate limjti rules across different infrastructure providers.
2. Support emerging workloads like AI Agents, which elevate the criticality of rate limiting due to the higher costs of requests.

### Who

- **Platform Operators**: Need to ensure uniform enforcement of the rate limit policies across all networking infrastructure.
- **Platform Operators**: Need to ensure fair sharing of compute resources by different classes of traffic.
- **Platform Operators**: Need to control costs incurred by different classes of traffic per time unit.

## API

### Policy Attachment vs. Inline Configuration

Rate Limnit is proposed to use the Policy Attachment model as the most effective approach to meet the stated goals, primarily for two reasons:

1. **Separation of Concerns**: It allows different personas to manage Gateway infrastructure independently from the configuration of rate limits. Rate limits are typically configured by platform, but could also be configured by cluster operators. This also enables rate limit policy to target different traffic scopes, from Gateway to HTTPRoute or a Service.
2. **Uniformity**: It enables a single policy to be applied uniformly across a set of Gateways, eliminating the need to duplicate complex telemetry configurations across individual resources.

This approach can lead to multiple rate limit policies to be applied to the same scope. If a request matches multiple rate limit buckets, it is throttled if any of the assigned buckets is above its configured limit.

### High-level Considerations:

- **Statically Defined Buckets**: Policy specifies a static number of rate limit buckets with expressions for assigning requests to buckets.
- **Dynamic Buckets**: Ability for dataplane to create rate limit buckets dynamically based on attribute values of observed traffic. For example, this allows the definition of fair sharing of resources between workloads without the need to define buckets for each workload in configuration.
- **Shadow Mode**: For testing rate limit policies without enforcement.
- **Variable Cost Requests**: Supporting rate limiting of requests that have different costs such as inference requests. This includes supporting the case where request cost is known only at the response time.

## Example Request Flow

* A platform operator creates a `RateLimitPolicy` resource targeting a `HTTPRoute`.
* The Gateway API implementation reconciles this resource and configures the underlying data plane.
* The data plane enforces defined policy for the traffic matching targeted `HTTPRoute`.
