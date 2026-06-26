---
title: "GEP: Standardized Telemetry API"
---

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

A key area of discussion for this GEP is whether this should be a standalone Policy Attachment (e.g., `TelemetryPolicy`) or inline configuration within `Gateway` or `HTTPRoute` resources.

This proposal argues that the Policy Attachment model is the most effective approach to meet the stated goals, primarily for two reasons:

1. **Separation of Concerns**: It allows different personas to manage Gateway infrastructure independently from the configuration of telemetry signals. Telemetry is typically configured by platform, observability, or security engineers rather than application developers. This also implies that HTTPRoute is not the ideal resource to target for the initial API implementation.
2. **Uniformity**: It enables a single policy to be applied uniformly across a set of Gateways, eliminating the need to duplicate complex telemetry configurations across individual resources.

To mitigate the challenge of complex merging semantics, this GEP restricts configuration such that only a single `TelemetryPolicy` can target a specific `Gateway` at any given time. If multiple `TelemetryPolicy` resources target the same object, precedence is determined based on the creation timestamp. This will allow us to start with simple config and iterate based on feedback whether multiple TelemetryPolicies on the same target are needed.

### High-level Considerations:

- **Tracing**: Configuration for OTLP endpoints, sampling rates (probabilistic and parent-based), and custom resource/span attributes.
- **Metrics**: Ability to enable/disable specific metric families and customize dimensions (labels/attributes).
- **Access Logs**: Filtering for smart logging (e.g., only log 5xx errors or high latency), multi-protocol support, and log format customization (including field selection).
- **Export Configuration**: Supporting TLS connections to telemetry collectors and the ability to inject custom headers (e.g., `Authorization`) into telemetry requests.

### Request Flow

* A platform operator creates a `TelemetryPolicy` resource targeting a `Gateway`.
* The Gateway API implementation reconciles this resource and configures the underlying data plane.
* The data plane extracts the specified signals and exports them to the telemetry infrastructure.

### The `TelemetryPolicy` Specification

We propose the `TelemetryPolicy` as a direct policy attachment in the `gateway.networking.k8s.io` API group. See [GEP-713](https://gateway-api.sigs.k8s.io/geps/gep-713/#classes-of-policies) for more information on direct attachment.

The following is an example that demonstrates the structure of the `TelemetryPolicy`.

```yaml
apiVersion: agentic.networking.x-k8s.io/v1alpha1
kind: TelemetryPolicy
metadata:
  name: standard-telemetry
  namespace: prod-ns
spec:
  # GEP-713 Attachment
  targetRefs:
  - group: gateway.networking.k8s.io
    kind: Gateway
    name: my-gateway
  
  # 1. Tracing Configuration
  tracing:
    mode: "On"
    provider:
      backendRef:
        group: ""
        kind: Service
        name: otel-collector
        namespace: monitoring
        port: 4317
    samplingRate: 
      numerator: 5 # Represents 5/100 (5%) because denominator defaults to 100
    parentBasedSampling:
      mode: "On"
      samplingRate:
        numerator: 50 # Represents 50/100 (50%)
    attributes:
      - name: "env"
        type: Literal
        literalValue: "production"
      - name: "mcp_tool_name"
        type: Reference
        attributeRef: "gen_ai.tool.name"

  # 2. Metrics Configuration
  metrics:
    mode: "On"
    overrides:
      - name: "example.com/http/request_count"
        type: Counter
        attributes: # Inject custom attributes/labels
          - name: "x-model-id"
            type: Header
            headerName: "X-Model-Id"
          - name: "mcp_tool_name"
            type: Reference
            attributeRef: "gen_ai.tool.name"
          - name: "environment"
            type: Literal
            literalValue: "production"

  # 3. Access Logs Configuration
  accessLogs:
    mode: "Off" # Explicitly disabled while keeping the configuration intact
    matches: "response.code >= 500" # Conditional logging, CEL filtering for errors
    fields: # Configure specific fields to include, indicating their source
      - path: ["http", "request", "start_time"] # Standard nested JSON structure
        type: Reference
        attributeRef: "timestamp"
      - path: ["http", "response", "status_code"]
        type: Reference
        attributeRef: "http.response.status_code"
      - path: ["token-usage"]
        type: Header
        headerName: "X-Token-Usage"
      - path: ["mcp.method"] # Segment with dots (Preserved verbatim as flat key)
        type: Reference
        attributeRef: "mcp.method.name"
```

#### Detailed Resource Description

The following are the Go structs modeling the proposed specification.

```golang
// TelemetryPolicy defines a Direct Attached Policy to configure 
// telemetry/observability signals for Gateways.
//
// By applying a TelemetryPolicy, platform operators and developers can ensure
// consistent collection, formatting, and export of observability signals.
//
// <gateway:util:excludeFromCRD>
// Notes for implementors:
//
// TelemetryPolicy is a Direct Attached Policy. Implementing controllers MUST
// adhere to the Policy Attachment guidelines (GEP-713).
//
// Precedence and Conflict Resolution:
// * To prevent complex merging semantics, only a single TelemetryPolicy is
//   permitted to target a specific Gateway resource at any given time.
// * If multiple TelemetryPolicy resources target the same Gateway, precedence
//   MUST be determined using the following criteria, continuing on ties:
//   1. The older policy by creation timestamp takes precedence.
//   2. The policy appearing first in alphabetical order by {namespace}/{name}.
// * For any TelemetryPolicy that does not take precedence, the controller
//   MUST set the `Accepted` condition on the policy status to `status: False` with
//   Reason `Conflicted`.
//
// Conformance:
// Implementations MUST support the core resource structure and `targetRefs`.
// Support for tracing, metrics, and accessLogs blocks is Extended, but if supported,
// their respective conformance profiles must be met.
// </gateway:util:excludeFromCRD>
//
// Support: Core (Resource shell and targetRefs), Extended (Signals)
type TelemetryPolicy struct {
  metav1.TypeMeta   `json:",inline"`
  metav1.ObjectMeta `json:"metadata,omitempty"`

  // Spec defines the desired state of TelemetryPolicy.
  //
  // +required
  Spec TelemetryPolicySpec `json:"spec"`

  // Status defines the observed state of TelemetryPolicy.
  //
  // +optional
  Status TelemetryPolicyStatus `json:"status,omitempty"`
}

// TelemetryPolicySpec defines the desired state and target of TelemetryPolicy.
//
// Specifying at least one target resource in `targetRefs` is required.
// Signals (tracing, metrics, and accessLogs) can be individually configured.
//
// Support: Core
type TelemetryPolicySpec struct {
  // TargetRefs identifies the gateways to which this policy applies (GEP-713).
  //
  // When configured, the telemetry settings defined in this policy are applied
  // uniformly to the referenced resources. In the absence of targetRefs, the policy is
  // invalid and will not be accepted.
  //
  // TargetRefs must be distinct.
  //
  // Support: Core
  //
  // +required
  // +kubebuilder:validation:MinItems=1
  TargetRefs []NamespacedPolicyTargetReference `json:"targetRefs"`

  // Tracing defines the configuration for distributed tracing.
  //
  // When configured, distributed tracing spans are generated and exported. In the
  // absence of this configuration, tracing behavior is determined by implementation
  // defaults (typically disabled).
  //
  // Support: Extended
  //
  // +optional
  Tracing *TracingConfig `json:"tracing,omitempty"`

  // Metrics defines the configuration for metric generation and custom attributes.
  //
  // When configured, custom metric attributes are applied. In the absence of this
  // configuration, metrics are generated according to implementation-default definitions.
  //
  // Support: Extended
  //
  // +optional
  Metrics *MetricsConfig `json:"metrics,omitempty"`

  // AccessLogs defines the configuration for access log generation and filters.
  //
  // When configured, access log generation, filtering, and attribute customisation are
  // applied. In the absence of this configuration, access logging is determined by
  // implementation defaults.
  //
  // Support: Extended
  //
  // +optional
  AccessLogs *AccessLogsConfig `json:"accessLogs,omitempty"`
}

// TelemetryMode defines the enablement state of a telemetry signal.
type TelemetryMode string

const (
  // TelemetryModeOn explicitly enables the telemetry signal.
  TelemetryModeOn  TelemetryMode = "On"
  // TelemetryModeOff explicitly disables the telemetry signal.
  TelemetryModeOff TelemetryMode = "Off"
)

// AttributeSourceType defines the source from which a telemetry attribute
// value is retrieved.
//
// Support: Core
type AttributeSourceType string
const (
  // AttributeSourceHeader indicates that the attribute value should be 
  // extracted from a specific HTTP header in the request or response.
  //
  // Support: Core
  AttributeSourceHeader = "Header"

  // AttributeSourceLiteral indicates that the attribute value is a static 
  // string provided directly in the policy configuration.
  //
  // Support: Core
  AttributeSourceLiteral = "Literal"

  // AttributeSourceReference extracts the value from a proxy-builtin reference variable
  // mapped to OpenTelemetry Semantic Conventions (e.g., "http.request.method").
  // See: https://opentelemetry.io/docs/specs/semconv/
  //
  // Support: Extended
  AttributeSourceReference = "Reference"
)

// Attribute defines a single flat key-value pair to attach to metrics and traces.
//
// This allows users to enrich spans and metrics with context like HTTP headers
// (e.g., "X-User-ID"), static tags, or built-in variables.
//
// Support: Core
type Attribute struct {
  // Name is the key of the attribute as it will appear in the output.
  // (e.g., as a span tag or metric label).
  //
  // +required
  Name string `json:"name"`

  // Type specifies where the attribute value comes from.
  // Valid values are "Header", "Literal", or "Reference".
  //
  // +required
  // +kubebuilder:validation:Enum=Header;Literal;Reference
  Type AttributeSourceType `json:"type"`

  // HeaderName specifies the HTTP header to extract the value from.
  // This is required if Type is "Header".
  //
  // +optional
  HeaderName *string `json:"headerName,omitempty"`

  // LiteralValue specifies a static string value to attach.
  // This is required if Type is "Literal".
  //
  // +optional
  LiteralValue *string `json:"literalValue,omitempty"`

  // AttributeRef refers to a standard OpenTelemetry attribute.
  // For example: "http.response.status_code" or "http.request.method".
  // This is required if Type is "Reference".
  // See: https://opentelemetry.io/docs/specs/semconv/
  //
  // +optional
  AttributeRef *string `json:"attributeRef,omitempty"`
}

// LogField defines a structured, potentially nested, field to include in JSON access logs.
//
// Support: Core
type LogField struct {
  // Path defines the nested key path under which the field value will be stored in the JSON payload.
  // Each element of the slice represents a nesting level. Any individual segment can contain dots,
  // which are preserved verbatim at that specific level of nesting.
  // For example, ["user.metadata", "id"] will serialize to {"user.metadata": {"id": "<value>"}}.
  //
  // +required
  // +kubebuilder:validation:MinItems=1
  Path []string `json:"path"`

  // Type specifies where the attribute value comes from.
  // Valid values are "Header", "Literal", or "Reference".
  //
  // +required
  // +kubebuilder:validation:Enum=Header;Literal;Reference
  Type AttributeSourceType `json:"type"`

  // HeaderName specifies the HTTP header to extract the value from.
  // This is required if Type is "Header".
  //
  // +optional
  HeaderName *string `json:"headerName,omitempty"`

  // LiteralValue specifies a static string value to attach.
  // This is required if Type is "Literal".
  //
  // +optional
  LiteralValue *string `json:"literalValue,omitempty"`

  // AttributeRef refers to a standard OpenTelemetry attribute.
  // For example: "http.response.status_code" or "http.request.method".
  // This is required if Type is "Reference".
  //
  // +optional
  AttributeRef *string `json:"attributeRef,omitempty"`
}

// --- Tracing Types ---

// TracingConfig defines the configuration for distributed tracing.
//
// Distributed tracing tracks the lifecycle of an individual request as it propagates through
// the Gateway and downstream services. Each service records a segment of the request's path
// as a "span". This configuration allows platform operators to enable tracing, select the
// destination backend, control the portion of traffic sampled, and inject custom values as
// span attributes.
//
// Users get granular visibility into request latency, system bottlenecks, and execution flows
// across complex distributed systems.
//
// Support: Extended
type TracingConfig struct {
  // Mode explicitly controls if tracing is enabled. Valid values are "On" or "Off".
  //
  // In the absence of this field, it defaults to "On".
  //
  // Support: Core (within Tracing feature)
  //
  // +kubebuilder:validation:Enum=On;Off
  // +kubebuilder:default=On
  Mode TelemetryMode `json:"mode,omitempty"`

  // Provider specifies the tracing collector or backend endpoint receiving OTLP spans.
  //
  // When configured, spans generated by the Gateway proxy are exported to this destination.
  // In the absence of this field, spans are exported to an implementation-defined default sink.
  //
  // Support: Core (within Tracing feature)
  //
  // +optional
  Provider *TracingProvider `json:"provider,omitempty"`

  // SamplingRate specifies the base probability of sampling new traces.
  //
  // Represented as a fraction. The probability of trace sampling is calculated as:
  //
  // $$ \text{Sampling Probability} = \frac{\text{Numerator}}{\text{Denominator}} $$
  //
  // For example, a Numerator of 5 and Denominator of 100 represents a 5% sampling rate.
  // * If configured, only the specified percentage of new traces will be initiated.
  // * In the absence of this field, an implementation-defined default is used.
  //
  // <gateway:util:excludeFromCRD>
  // Notes for implementors:
  //
  // Permutations of numerator > denominator are invalid and MUST be rejected via validation.
  // </gateway:util:excludeFromCRD>
  //
  // Support: Extended
  //
  // +optional
  SamplingRate *Fraction `json:"samplingRate,omitempty"`

  // ParentBasedSampling configures whether to respect the sampling decision of the parent span.
  //
  // * When Mode is "On", the proxy will respect the upstream trace parent's sampling decision.
  // * When Mode is "Off" or absent, the proxy applies its own local sampling rate decision.
  //
  // Support: Extended
  //
  // +optional
  ParentBasedSampling *ParentBasedSampling `json:"parentBasedSampling,omitempty"`

  // Attributes is a list of custom key-value pairs (or variables) attached to every span.
  //
  // When configured, these attributes are injected into every generated tracing span.
  // In the absence of attributes, only standard proxy-defined attributes are emitted.
  //
  // Support: Extended
  //
  // +optional
  Attributes []Attribute `json:"attributes,omitempty"`
}

// TracingProvider identifies the tracing backend that receives generated spans.
//
// Support: Core (within Tracing feature)
type TracingProvider struct {
  // BackendRef is a reference to a Kubernetes Service or other supported 
  // backend that receives OTLP traces.
  //
  // When configured, tracing data is exported to the referenced backend. If the reference
  // is invalid (e.g., the Service does not exist), the implementation should update the
  // policy's status conditions to indicate an unresolved reference.
  //
  // TLS configuration for the connection to the backend is managed by the referenced
  // object. For example, if the BackendRef points to a Service, a BackendTLSPolicy
  // can be attached to configure TLS. Alternatively, the referenced backend could be a
  // custom resource (e.g., XBackend) that natively manages TLS.
  //
  // Support: Core
  //
  // +required
  BackendRef BackendObjectReference `json:"backendRef"`
}

// Fraction represents a ratio used for probabilistic sampling rates.
//
// The probability is calculated as numerator / denominator (e.g. 5 / 100 = 5%).
//
// Support: Core (within Tracing feature)
type Fraction struct {
  // Numerator specifies the top of the fraction.
  //
  // +required
  Numerator int32 `json:"numerator"`

  // Denominator specifies the bottom of the fraction.
  //
  // Defaults to 100 if unspecified.
  //
  // +kubebuilder:default=100
  // +kubebuilder:validation:Minimum=1
  // +optional
  Denominator int32 `json:"denominator,omitempty"` // Allows e.g., 1 / 10000 for 0.01%
}

// ParentBasedSampling defines the sampling behavior when a request has a pre-existing upstream
// trace parent.
//
// Support: Extended
type ParentBasedSampling struct {
  // Mode explicitly controls if parent-based sampling is enabled. Valid values are "On" or "Off".
  //
  // Defaults to "On" if parent-based sampling is configured.
  //
  // Support: Extended
  //
  // +kubebuilder:validation:Enum=On;Off
  // +kubebuilder:default=On
  Mode TelemetryMode `json:"mode,omitempty"`
  
  // SamplingRate is the sampling rate to apply when parent-based sampling is active.
  //
  // This acts as a downsampling governor. It allows an operator to say: "I want to
  // respect the parent's decision, but only for 50% of those requests". Even if a
  // parent is already marked as "Sampled", this allows the Gateway to apply a secondary
  // filter so that it can respect the parent's intent while still controlling the volume
  // of spans reported.
  //
  // Support: Extended
  //
  // +optional
  SamplingRate *Fraction `json:"samplingRate,omitempty"`
}

// --- Metrics Types ---

// MetricsConfig defines configuration options for proxy metric generation.
//
// Metrics provide numeric measurements (counters, gauges, histograms) representing traffic
// volume, error rates, latency, proxy performance, etc. Gateway users get high-level
// observability into the behavior of their API traffic.
//
// Support: Extended
type MetricsConfig struct {
  // Mode explicitly controls if metric generation is enabled. Valid values are "On" or "Off".
  //
  // Defaults to "On" if the metrics block is configured.
  //
  // Support: Core (within Metrics feature)
  //
  // +kubebuilder:validation:Enum=On;Off
  // +kubebuilder:default=On
  Mode TelemetryMode `json:"mode,omitempty"`

  // Overrides defines a list of customizations to specific metric families.
  //
  // When configured, these overrides alter standard metrics by adding custom attributes
  // (labels) or changing metric-specific settings. In the absence of overrides, only
  // standard out-of-the-box proxy metrics are generated.
  //
  // Support: Extended
  //
  // +optional
  Overrides []MetricOverride `json:"overrides,omitempty"`
}

// MetricOverride configures customization for a specific named metric family.
//
// At present it allows the inclusion of additional attributes for existing metrics.
//
// Support: Extended
type MetricOverride struct {
  // Name specifies the exact name of the metric to override (e.g., "http_requests_total"
  // or "example.com/http/request_count").
  //
  // Support: Core (within Metrics override feature)
  //
  // +required
  Name string `json:"name"`

  // Type specifies the metric instrument type (e.g., "Counter", "Histogram").
  //
  // Support: Extended
  //
  // +optional
  Type string `json:"type,omitempty"`

  // Attributes defines custom labels/dimensions to append to the overridden metric.
  //
  // These allow tracking business or environment-specific tags on standard metrics,
  // such as classifying metrics by incoming API token metadata or model identifiers.
  //
  // Support: Extended
  //
  // +optional
  Attributes []Attribute `json:"attributes,omitempty"`
}

// --- Access Logs Types ---

// AccessLogsConfig defines the configuration for access log generation.
//
// Access logs record metadata for every individual request/response transaction
// (e.g., start time, status code, request path, size). Users get a persistent, readable audit
// trail of all traffic passing through the Gateway, which is critical for security audits,
// compliance, and troubleshooting of failures.
//
// Support: Extended
type AccessLogsConfig struct {
  // Mode explicitly controls if access logging is enabled. Valid values are "On" or "Off".
  //
  // Defaults to "On" if the accessLogs block is configured.
  //
  // Support: Core (within AccessLogs feature)
  //
  // +kubebuilder:validation:Enum=On;Off
  // +kubebuilder:default=On
  Mode TelemetryMode `json:"mode,omitempty"`

  // Matches specifies a Common Expression Language (CEL) expression used to filter
  // which requests are logged (e.g., "response.code >= 500").
  //
  // When configured, only requests matching the condition are written to access logs.
  // This enables "smart logging" (logging only errors or slow requests), reducing log volume
  // and storage costs. In the absence of a match filter, all requests are logged.
  //
  // <gateway:util:excludeFromCRD>
  // Notes for implementors:
  //
  // Not all data plane engines natively support CEL evaluation. If an implementation
  // cannot support CEL-based log filtering, it MUST raise an `Accepted` condition of
  // `False` with reason `UnsupportedField` or `Invalid` on the policy status when this field is used.
  // </gateway:util:excludeFromCRD>
  //
  // Support: Implementation-specific
  //
  // +optional
  Matches string `json:"matches,omitempty"`

  // Fields specifies the structured JSON fields and variables included in the generated access logs.
  //
  // When configured, the generated JSON log records will include these fields at the paths
  // defined by each LogField. In the absence of fields, the proxy uses its default structured
  // JSON format.
  //
  // Support: Extended
  //
  // +optional
  Fields []LogField `json:"fields,omitempty"`
}

// --- Policy Status ---

// TelemetryPolicyStatus defines the observed state of TelemetryPolicy.
type TelemetryPolicyStatus struct {
  // For Policy Status API conventions, see:
  // https://gateway-api.sigs.k8s.io/geps/gep-713/#the-status-stanza-of-policy-objects
  //
  // Ancestors is a list of ancestor resources (usually Gateway) that are
  // associated with the policy, and the status of the policy with respect to
  // each ancestor. When this policy attaches to a parent, the controller that
  // manages the parent and the ancestors MUST add an entry to this list when
  // the controller first sees the policy and SHOULD update the entry as
  // appropriate when the relevant ancestor is modified.
  //
  // Note that choosing the relevant ancestor is left to the Policy designers;
  // an important part of Policy design is designing the right object level at
  // which to namespace this status.
  //
  // Note also that implementations MUST ONLY populate ancestor status for
  // the Ancestor resources they are responsible for. Implementations MUST
  // use the ControllerName field to uniquely identify the entries in this list
  // that they are responsible for.
  //
  // Note that to achieve this, the list of PolicyAncestorStatus structs
  // MUST be treated as a map with a composite key, made up of the AncestorRef
  // and ControllerName fields combined.
  //
  // A maximum of 16 ancestors will be represented in this list. An empty list
  // means the Policy is not relevant for any ancestors.
  //
  // If this slice is full, implementations MUST NOT add further entries.
  // Instead they MUST consider the policy unimplementable and signal that
  // on any related resources such as the ancestor that would be referenced
  // here.
  //
  // +required
  // +listType=atomic
  // +kubebuilder:validation:MaxItems=16
  Ancestors []PolicyAncestorStatus `json:"ancestors"`
}
```

#### Attribute References and Portability

To ensure portability and avoid implementation-specific lock-in, the `Reference` attribute source type relies exclusively on standard [OpenTelemetry Semantic Conventions](https://opentelemetry.io/docs/specs/semconv/).

When users specify an `attributeRef`, they must use these standardized keys (e.g., `http.request.method`, `http.response.status_code`). The underlying Gateway API implementations are responsible for mapping these standard OpenTelemetry keys to their proxy-specific internal variables.

Implementations MUST NOT expose internal, proxy-specific variables through the `Reference` type. If an implementation does not support mapping a specific standard attribute, it SHOULD gracefully omit it or signal the limitation via policy status conditions.

## Alternatives Considered

### Implementation-Specific OpenTelemetry Enablement

During the initial proposal in the [kube-agentic-networking](https://github.com/kubernetes-sigs/kube-agentic-networking) subproject, an alternative was suggested to avoid defining a new API standard. The idea was that implementations should natively implement the OpenTelemetry specification for traces, metrics, and logs, and simply provide their own implementation-specific mechanisms to enable or disable these features.

**Reason for Rejection:** 
While this works as a baseline, it falls short when users need to customize their telemetry (which is fairly common). Customizations like adding specific attributes or conditional log filtering would require users to rely on vendor-specific APIs increasing the risk of lock-in.

### Inline Gateway Configuration

Another alternative considered was adding the telemetry configuration directly as a top-level struct on the `Gateway` resource instead of introducing a new Policy object.

**Reason for Rejection:** 
While inline configuration works well for a 1:1 mapping on a single Gateway, a separate Policy attachment model provides a decoupled, reusable configuration. A single `TelemetryPolicy` can be applied uniformly to multiple gateways, meaning platform operators and developers can ensure consistent telemetry signals across their infrastructure. This approach prevents configuration drift and avoids bloating the core `Gateway` API specification.

## Comparison with Prior Art

### Istio

[Istio](https://istio.io/)'s `Telemetry` API is the most direct prior art that inspired this proposal. It allows configuring observability at the mesh, namespace, and workload level.

* **Metrics**: Istio allows users to enable/disable specific metrics, add custom dimensions, and configure providers.
* **Logs**: Istio supports access logging configurations with CEL-like expressions for advanced filtering.
* **Traces**: Istio supports probabilistic sampling, context propagation, and custom span tags.
* **Customization**: For advanced telemetry use-cases not natively covered by the `Telemetry` API, Istio users can fall back to using `EnvoyFilter` resources. While highly flexible, `EnvoyFilter` requires deep knowledge of Envoy's internal xDS API. This is tightly coupled to the data plane implementation and can be brittle across version upgrades.
* **Comparison**: The proposed `TelemetryPolicy` adapts Istio's powerful intent-based capabilities to the standardized Gateway API attachment model.

### Envoy Gateway

[Envoy Gateway](https://gateway.envoyproxy.io/) configures observability through two distinct custom resources: `EnvoyGateway` for the control plane and `EnvoyProxy` for the underlying data plane proxies.

* **Metrics**: Envoy Gateway allows configuring Prometheus and OpenTelemetry sinks for both the control plane (using `EnvoyGateway` CRD) and the data plane proxies (using the `EnvoyProxy` CRD).
* **Logs**: Proxy access logs are configured via the `EnvoyProxy` resource. It supports exporting to file, OTLP, or gRPC Access Log Service (ALS) sinks. It uses CEL expressions for smart filtering (e.g., matching specific headers), and allows applying log configurations at the Route or Listener level.
* **Tracing**: Tracing is configured in the `EnvoyProxy` resource. It allows configuring sampling and supports appending custom tags derived from literals, environment variables, or request headers.
* **Customization**: For advanced telemetry use-cases not covered natively, users can fall back to the `EnvoyPatchPolicy` API to mutate the underlying xDS configuration using JSON Patch semantics. This is similar to Istio's `EnvoyFilter`.
* **Comparison**: While Envoy Gateway provides a robust, native telemetry configuration, it is tightly coupled to infrastructure-oriented CRDs. The proposed `TelemetryPolicy` allows users to configure telemetry behaviors using a portable `targetRef` model, without binding their observability intent to an Envoy-specific schema.

### Kuadrant

[Kuadrant](https://kuadrant.io/) provides observability for API management features like rate limiting and authentication. It is configured through a mix of its own custom resources and the underlying gateway's APIs.

* **Metrics**: Kuadrant enables metrics via the `Kuadrant` CR. It also introduces its own `TelemetryPolicy` API (extensions.kuadrant.io/v1alpha1) to add custom dimensions to metrics.
* **Logs**: For proxy access logging, Kuadrant relies on the underlying gateway provider (e.g., Istio's Telemetry API). However, it configures request correlation across its own components (Authorino, Limitador, and Wasm-shim) by specifying HTTP header identifiers in the `Kuadrant` CR.
* **Tracing**: Tracing is configured centrally via the `Kuadrant` CR. It exports OpenTelemetry spans for both the control plane and data plane components. It supports global trace filtering levels to control the verbosity of exported spans.
* **Customization**: To make low-level, custom modifications to the data plane configuration that are not supported by Kuadrant's native APIs, users can bypass Kuadrant and directly use the underlying gateway's mechanisms.
* **Comparison**: While Kuadrant provides powerful, identity-aware telemetry (like token tracking per user), its configuration is fragmented across the `Kuadrant` CR, components specific CRDs, its custom extension `TelemetryPolicy`, and the underlying gateway's native APIs. The proposed `TelemetryPolicy` aims to unify these intent-based capabilities into a single, provider-agnostic resource.
 
### Airlock Microgateway

[Airlock Microgateway](https://docs.airlock.com/microgateway/5.0/index/api/crds/telemetry/v1alpha1/) defines a `Telemetry` CRD to configure logging, metrics, and tracing.

* **Metrics**: While the `Telemetry` CRD broadly targets telemetry, metric generation is largely handled by default configurations rather than via customization within the CRD itself.
* **Logs**: Configures access logs with customizable JSON and ECS formats, relying on Envoy-specific log variables and dynamic metadata extraction. 
* **Tracing**: Supports configuring an OpenTelemetry provider with deep exporter settings (e.g., gRPC/HTTP endpoints and custom TLS certificate pinning) and sampling strategies (ratio or parent-based).
* **Customization**: Explicitly supports defining mechanisms to extract and propagate correlation identifiers from request headers directly within the telemetry configuration.
* **Comparison**: While Airlock utilizes a unified `Telemetry` custom resource, its specification includes implementation-specific details (like TLS pinning strategies and Envoy string formatting). The proposed `TelemetryPolicy` abstracts these into a more portable, generalized resource.

### NGINX Gateway Fabric

[NGINX Gateway Fabric](https://docs.nginx.com/nginx-gateway-fabric/reference/api/) splits its telemetry configuration across its `NginxProxy` and `ObservabilityPolicy` custom resources.

* **Metrics**: Global data plane observability, such as Prometheus metrics scraping, is managed via the `NginxProxy` resource which can be referenced from a `GatewayClass` or `Gateway`.
* **Logs**: Access log formatting and enablement are also managed centrally via the `NginxProxy` resource.
* **Tracing**: Distributed tracing is configured using the `ObservabilityPolicy`, which is a Direct Attached Policy that specifically targets `HTTPRoute` or `GRPCRoute`. It supports configuring OpenTelemetry sampling strategies (ratio or parent-based), context propagation, custom span names, and span attributes.
* **Customization**: For advanced proxy configurations not natively covered by the standard policies, users can inject raw NGINX configuration using the `SnippetsPolicy` at the Gateway level or the `SnippetsFilter` at the Route level.
* **Comparison**: NGINX Gateway Fabric separates its telemetry intents across multiple layers, splitting infrastructure-level metrics and logs from route-level tracing configurations. The proposed `TelemetryPolicy` consolidates these observability signals into a single Direct Attached Policy targeting the `Gateway`.
