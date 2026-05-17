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
      endpoint: "otel-collector.monitoring.svc:4317"
    samplingRate: 
      numerator: 5 # Represents 5/100 (5%) because denominator defaults to 100
    parentBasedSampling:
      mode: "On"
      samplingRate:
        numerator: 50 # Represents 50/100 (50%)
    customAttributes:
      - name: "env"
        type: Literal
        literalValue: "production"
      - name: "mcp_task_name"
        type: Metadata
        metadataKey: "my.custom.filter.mcp_task_name"

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
          - name: "mcp_task_name"
            type: Metadata
            metadataKey: "my.custom.filter.mcp_task_name"
          - name: "environment"
            type: Literal
            literalValue: "production"

  # 3. Access Logs Configuration
  accessLogs:
    mode: "Off" # Explicitly disabled while keeping the configuration intact
    matches: "response.code >= 500" # Conditional logging, CEL filtering for errors
    fields: # Configure specific fields to include, indicating their source
      - name: "start_time"
        type: Standard
        standardValue: "RequestStartTime"
      - name: "response_code"
        type: Standard
        standardValue: "ResponseCode"
      - name: "x-token-usage"
        type: Header
        headerName: "X-Token-Usage"
      - name: "mcp_task_name"
        type: Metadata
        metadataKey: "my.custom.filter.mcp_task_name"
```

#### Detailed Resource Description

The following are the Go structs modeling the proposed specification.

```Go
// TelemetryPolicy defines a direct policy attachment to configure 
// observability signals for Gateways.
type TelemetryPolicy struct {
  metav1.TypeMeta   `json:",inline"`
  metav1.ObjectMeta `json:"metadata,omitempty"`

  Spec TelemetryPolicySpec `json:"spec"`

  // status defines the observed state of TelemetryPolicy.
  // +optional
  Status TelemetryPolicyStatus `json:"status,omitempty"`
}

type TelemetryPolicySpec struct {
  // Identifies the target gateways to which this policy attaches (GEP-713).
  TargetRefs []NamespacedPolicyTargetReference `json:"targetRefs"`

  // Configuration for distributed tracing options.
  Tracing *TracingConfig `json:"tracing,omitempty"`

  // Configuration for metric generation and exports.
  Metrics *MetricsConfig `json:"metrics,omitempty"`

  // Configuration for access log generation.
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

// --- Tracing Types ---

type TracingConfig struct {
  // Mode explicitly controls if tracing is enabled. Valid values are "On" or "Off".
  // +kubebuilder:validation:Enum=On;Off
  // +kubebuilder:default=On
  Mode TelemetryMode `json:"mode,omitempty"`

  // Specifies the tracing backend. Includes type (e.g., "OTLP") and endpoint.
  Provider *TracingProvider `json:"provider,omitempty"`

  // The base sampling probability for traces.
  SamplingRate *Fraction `json:"samplingRate,omitempty"`

  // Configures whether to respect the sampling decision of the parent span.
  ParentBasedSampling *ParentBasedSampling `json:"parentBasedSampling,omitempty"`

  // Allows appending custom tags/attributes to spans.
  CustomAttributes []CustomAttribute `json:"customAttributes,omitempty"`
}

type TracingProvider struct {
  Endpoint string `json:"endpoint,omitempty"`
}

type Fraction struct {
  Numerator int32 `json:"numerator"`
  
  // +kubebuilder:default=100
  // +kubebuilder:validation:Minimum=1
  Denominator int32 `json:"denominator,omitempty"` // Allows e.g., 1 / 10000 for 0.01%
}

type ParentBasedSampling struct {
  // Mode explicitly controls if parent-based sampling is enabled. Valid values are "On" or "Off".
  // +kubebuilder:validation:Enum=On;Off
  // +kubebuilder:default=On
  Mode TelemetryMode `json:"mode,omitempty"`
  
  // The sampling rate to apply when the parent span decision is used.
  SamplingRate *Fraction `json:"samplingRate,omitempty"`
}

// CustomAttributeType defines the source of a trace attribute's value.
type CustomAttributeType string

const (
  // CustomAttributeTypeHeader extracts the value from an HTTP header.
  CustomAttributeTypeHeader CustomAttributeType = "Header"
  // CustomAttributeTypeMetadata extracts the value from proxy metadata or context.
  CustomAttributeTypeMetadata CustomAttributeType = "Metadata"
  // CustomAttributeTypeLiteral provides a static, user-defined string value.
  CustomAttributeTypeLiteral CustomAttributeType = "Literal"
)

type CustomAttribute struct {
  // Name is the key of the attribute as it will appear in the trace span.
  Name string `json:"name"`

  // Type specifies where the attribute value comes from.
  // Valid values are "Header", "Metadata", or "Literal".
  // +kubebuilder:validation:Enum=Header;Metadata;Literal
  Type CustomAttributeType `json:"type"`

  // HeaderName specifies the HTTP header to extract the value from.
  // This is required if Type is "Header".
  HeaderName *string `json:"headerName,omitempty"`

  // MetadataKey specifies the proxy/context metadata key to extract the value from.
  // This is required if Type is "Metadata".
  MetadataKey *string `json:"metadataKey,omitempty"`

  // LiteralValue specifies a static string value to attach.
  // This is required if Type is "Literal".
  LiteralValue *string `json:"literalValue,omitempty"`
}

// --- Metrics Types ---

type MetricsConfig struct {
  // Mode explicitly controls if metric generation is enabled. Valid values are "On" or "Off".
  // +kubebuilder:validation:Enum=On;Off
  // +kubebuilder:default=On
  Mode TelemetryMode `json:"mode,omitempty"`

  // List of configurations to customize specific metric families.
  Overrides []MetricOverride `json:"overrides,omitempty"`
}

type MetricOverride struct {
  // The metric name to override (e.g., "http_requests_total" or "gateway.networking.k8s.io/http/request_count").
  Name string `json:"name"`

  // Type of the metric (e.g., "Counter", "Histogram").
  Type string `json:"type,omitempty"`

  // Defines custom attributes to attach to the metric. 
  // These are appended to the standard labels emitted by the proxy.
  Attributes []MetricAttribute `json:"attributes,omitempty"`
}

// MetricAttributeType defines the source of a metric attribute's value.
type MetricAttributeType string

const (
  // MetricAttributeTypeHeader extracts the value from an HTTP header.
  MetricAttributeTypeHeader MetricAttributeType = "Header"
  // MetricAttributeTypeMetadata extracts the value from proxy metadata or context.
  MetricAttributeTypeMetadata MetricAttributeType = "Metadata"
  // MetricAttributeTypeLiteral provides a static, user-defined string value.
  MetricAttributeTypeLiteral MetricAttributeType = "Literal"
)

type MetricAttribute struct { 
  // Name is the key of the attribute as it will appear in the metric. 
  Name string `json:"name"` 
  
  // Type specifies where the attribute value comes from.
  // Valid values are "Header", "Metadata", or "Literal".
  // +kubebuilder:validation:Enum=Header;Metadata;Literal
  Type MetricAttributeType `json:"type"`

  // HeaderName specifies the HTTP header to extract the value from.
  // This is required if Type is "Header".
  HeaderName *string `json:"headerName,omitempty"`

  // MetadataKey specifies the proxy/context metadata key to extract the value from.
  // This is required if Type is "Metadata".
  MetadataKey *string `json:"metadataKey,omitempty"`

  // LiteralValue specifies a static string value to attach.
  // This is required if Type is "Literal".
  LiteralValue *string `json:"literalValue,omitempty"`
}

// --- Access Logs Types ---

type AccessLogsConfig struct {
  // Mode explicitly controls if access logging is enabled. Valid values are "On" or "Off".
  // +kubebuilder:validation:Enum=On;Off
  // +kubebuilder:default=On
  Mode TelemetryMode `json:"mode,omitempty"`

  // CEL expression for advanced filtering (e.g., matching response codes, headers).
  Matches string `json:"matches,omitempty"`

  // A list of specific fields or headers to include in the logs.
  Fields []string `json:"fields,omitempty"`

  // A list of specific fields to include in the logs, specifying their source.
  Fields []LogField `json:"fields,omitempty"`
}

// LogFieldType defines the source of a log field's value.
type LogFieldType string

const (
  // LogFieldTypeHeader extracts the value from an HTTP header.
  LogFieldTypeHeader LogFieldType = "Header"
  // LogFieldTypeMetadata extracts the value from proxy metadata or context.
  LogFieldTypeMetadata LogFieldType = "Metadata"
  // LogFieldTypeLiteral provides a static, user-defined string value.
  LogFieldTypeLiteral LogFieldType = "Literal"
  // LogFieldTypeStandard extracts a standard proxy log value (e.g., duration, start time).
  LogFieldTypeStandard LogFieldType = "Standard"
)

type LogField struct {
  // Name is the key/name of the field as it will appear in the access log output.
  Name string `json:"name"`

  // Type specifies where the field value comes from.
  // Valid values are "Header", "Metadata", "Literal", or "Standard".
  // +kubebuilder:validation:Enum=Header;Metadata;Literal;Standard
  Type LogFieldType `json:"type"`

  // HeaderName specifies the HTTP header to extract the value from.
  // This is required if Type is "Header".
  HeaderName *string `json:"headerName,omitempty"`

  // MetadataKey specifies the proxy/context metadata key to extract the value from.
  // This is required if Type is "Metadata".
  MetadataKey *string `json:"metadataKey,omitempty"`

  // LiteralValue specifies a static string value to attach.
  // This is required if Type is "Literal".
  LiteralValue *string `json:"literalValue,omitempty"`

  // StandardValue specifies a standard log property (e.g., "RequestStartTime", "Duration").
  // This is required if Type is "Standard".
  StandardValue *string `json:"standardValue,omitempty"`
}

// --- Policy Status ---

// TelemetryPolicyStatus defines the observed state of TelemetryPolicy.
type TelemetryPolicyStatus struct {
  // For Policy Status API conventions, see:
  // https://gateway-api.sigs.k8s.io/geps/gep-713/#the-status-stanza-of-policy-objects
  //
  // Ancestors is a list of ancestor resources (usually Backend) that are
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
 
