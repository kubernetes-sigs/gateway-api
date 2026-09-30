/*
Copyright The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	v1 "sigs.k8s.io/gateway-api/apis/v1"
)

// +genclient
// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:storageversion
// +kubebuilder:resource:categories=gateway-api,shortName=xtelemetrypolicy
// +kubebuilder:printcolumn:name="Age",type=date,JSONPath=`.metadata.creationTimestamp`
//
// TelemetryPolicy is a Direct Attached Policy.
// +kubebuilder:metadata:labels="gateway.networking.k8s.io/policy=Direct"

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
//   - To prevent complex merging semantics, only a single TelemetryPolicy is
//     permitted to apply to a specific Gateway resource at any given time.
//   - If multiple TelemetryPolicy resources target the same Gateway, precedence
//     MUST be determined using the following criteria, continuing on ties:
//     1. The older policy by creation timestamp takes precedence.
//     2. The policy appearing first in alphabetical order by {namespace}/{name}.
//   - For any TelemetryPolicy that does not take precedence, the controller
//     MUST set the `Accepted` condition on the policy status to `status: False` with
//     Reason `Conflicted`.
//
// Conformance:
// Implementations MUST support the core resource structure and `targetRefs`.
// Support for the tracing block is Extended, but if supported,
// its respective conformance profile must be met.
// </gateway:util:excludeFromCRD>
//
// Support: Extended
type XTelemetryPolicy struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// Spec defines the desired state of TelemetryPolicy.
	// +required
	Spec TelemetryPolicySpec `json:"spec"`

	// Status defines the observed state of TelemetryPolicy.
	// +optional
	Status TelemetryPolicyStatus `json:"status,omitempty"`
}

// XTelemetryPolicyList contains a list of XTelemetryPolicy.
// +kubebuilder:object:root=true
type XTelemetryPolicyList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []XTelemetryPolicy `json:"items"`
}

// TelemetryPolicySpec defines the desired state and target of TelemetryPolicy.
//
// Specifying at least one target resource in `targetRefs` is required.
// Tracing behavior can be configured via the `tracing` field.
//
// +kubebuilder:validation:AtLeastOneOf=tracing
type TelemetryPolicySpec struct {
	// TargetRefs identifies the gateways to which this policy applies (GEP-713).
	//
	// When configured, the telemetry settings defined in this policy are applied
	// uniformly to the referenced resources. In the absence of targetRefs, the policy is
	// invalid and will not be accepted.
	//
	// TargetRefs must be distinct.
	//
	// Support: Core for Gateway
	//
	// +required
	// +listType=atomic
	// +kubebuilder:validation:MinItems=1
	// +kubebuilder:validation:MaxItems=16
	TargetRefs []v1.LocalObjectReference `json:"targetRefs"`

	// Tracing defines the configuration for distributed tracing.
	//
	// When configured, distributed tracing spans are generated and exported. In the
	// absence of this configuration, tracing behavior is determined by implementation
	// defaults.
	//
	// Support: Extended
	//
	// Feature Name: TelemetryPolicyTracing
	//
	// +optional
	Tracing *TracingConfig `json:"tracing,omitempty"`
}

// TracingMode defines the enablement state of tracing.
type TracingMode string

const (
	// TracingModeEnabled explicitly enables tracing.
	TracingModeEnabled TracingMode = "Enabled"

	// TracingModeDisabled explicitly disables tracing.
	TracingModeDisabled TracingMode = "Disabled"

	// TracingModeImplementationDefault means that the code should
	// use the implementation's default behavior for tracing.
	TracingModeImplementationDefault TracingMode = "ImplementationDefault"
)

// AttributeName defines the key of a span attribute or tag.
//
// +kubebuilder:validation:MinLength=1
// +kubebuilder:validation:MaxLength=256
// +kubebuilder:validation:Pattern=`^[a-zA-Z0-9_.:/-]+$`
type AttributeName string

// AttributeSourceType defines the source from which a telemetry attribute
// value is retrieved.
type AttributeSourceType string

const (
	// AttributeSourceHeader indicates that the attribute value should be
	// extracted from a specific HTTP header in the request or response.
	//
	// Support: Core
	AttributeSourceHeader AttributeSourceType = "Header"

	// AttributeSourceLiteral indicates that the attribute value is a static
	// string provided directly in the policy configuration.
	//
	// Support: Core
	AttributeSourceLiteral AttributeSourceType = "Literal"

	// AttributeSourceAttribute extracts the value from a proxy-builtin reference variable
	// mapped to OpenTelemetry Semantic Conventions (e.g., "http.request.method").
	// See: https://opentelemetry.io/docs/specs/semconv/
	//
	// Support: Extended
	//
	// Feature Name: TelemetryPolicyAttribute
	AttributeSourceAttribute AttributeSourceType = "Attribute"
)

// Attribute defines a single flat key-value pair to attach to traces.
//
// This allows users to enrich spans with context like HTTP headers
// (e.g., "X-User-ID"), static tags, or built-in variables.
//
// Support: Core
//
// +kubebuilder:validation:XValidation:rule="self.sourceType == 'Header' ? has(self.headerName) : !has(self.headerName)",message="headerName is required when sourceType is Header, and must be empty otherwise"
// +kubebuilder:validation:XValidation:rule="self.sourceType == 'Literal' ? has(self.literalValue) : !has(self.literalValue)",message="literalValue is required when sourceType is Literal, and must be empty otherwise"
// +kubebuilder:validation:XValidation:rule="self.sourceType == 'Attribute' ? has(self.attributeKey) : !has(self.attributeKey)",message="attributeKey is required when sourceType is Attribute, and must be empty otherwise"
type Attribute struct {
	// Name is the key of the attribute as it will appear in the output
	// (i.e., as a span tag).
	//
	// +required
	Name AttributeName `json:"name"`

	// SourceType specifies where the attribute value comes from.
	// Valid values are "Header", "Literal", or "Attribute".
	//
	// +unionDiscriminator
	// +required
	// +kubebuilder:validation:Enum=Header;Literal;Attribute
	SourceType AttributeSourceType `json:"sourceType"`

	// HeaderName specifies the HTTP header to extract the value from.
	// This is required if SourceType is "Header".
	//
	// +optional
	HeaderName v1.HTTPHeaderName `json:"headerName,omitempty"`

	// LiteralValue specifies a static string value to attach.
	// This is required if SourceType is "Literal".
	//
	// +optional
	// +kubebuilder:validation:MinLength=1
	// +kubebuilder:validation:MaxLength=1024
	LiteralValue string `json:"literalValue,omitempty"`

	// AttributeKey refers to a standard OpenTelemetry attribute.
	// For example: "http.response.status_code" or "http.request.method".
	// This is required if SourceType is "Attribute".
	// See: https://opentelemetry.io/docs/specs/semconv/
	//
	// +optional
	// +kubebuilder:validation:MinLength=1
	// +kubebuilder:validation:MaxLength=256
	// +kubebuilder:validation:Pattern=`^[a-z0-9_.-]+$`
	AttributeKey string `json:"attributeKey,omitempty"`
}

// TracingConfig defines the configuration for distributed tracing.
//
// Support: Extended
// +kubebuilder:validation:XValidation:rule="self.mode == 'Enabled' ? has(self.provider) : true",message="provider must be specified when mode is Enabled"
// +kubebuilder:validation:XValidation:rule="self.mode == 'Disabled' ? !has(self.provider) : true",message="provider must be empty when mode is Disabled"
type TracingConfig struct {
	// Mode explicitly controls if tracing is enabled. Valid values are "Enabled", "Disabled",
	// "ImplementationDefault".
	//
	// In the absence of this field, it defaults to "ImplementationDefault".
	//
	// Support: Core (within TelemetryPolicy feature)
	//
	// +kubebuilder:validation:Enum=Enabled;Disabled;ImplementationDefault
	// +kubebuilder:default=ImplementationDefault
	Mode TracingMode `json:"mode,omitempty"`

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
	// The sampling probability is represented as a fraction.
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
	SamplingRate *v1.Fraction `json:"samplingRate,omitempty"`

	// ParentBasedSampling configures whether to respect the sampling decision of the parent span.
	//
	// * When Mode is "Enabled", the proxy will respect the upstream trace parent's sampling
	//   decision.
	// * When Mode is "Disabled" or absent, the proxy applies its own local sampling rate
	//   decision.
	//
	// Support: Extended
	//
	// Feature Name: TelemetryPolicyParentBasedSampling
	//
	// +optional
	ParentBasedSampling *ParentBasedSampling `json:"parentBasedSampling,omitempty"`

	// ServiceName is the "service.name" attribute of the OpenTelemetry resource.
	// If absent, the implementation's default service name will be used.
	//
	// Support: Extended
	//
	// +optional
	// +kubebuilder:validation:MinLength=1
	// +kubebuilder:validation:MaxLength=253
	// +kubebuilder:validation:Pattern=`^(\*\.)?[a-z0-9]([-a-z0-9]*[a-z0-9])?(\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*$`
	ServiceName *string `json:"serviceName,omitempty"`

	// SpanName defines a custom name for the OTel span. By default, the name
	// is implementation-specific.
	//
	// Support: Extended
	//
	// +optional
	// +kubebuilder:validation:MinLength=1
	// +kubebuilder:validation:MaxLength=256
	SpanName *string `json:"spanName,omitempty"`

	// Attributes is a list of custom key-value pairs (or variables) attached to every span.
	//
	// When configured, these attributes are injected into every generated tracing span.
	// In the absence of attributes, only standard proxy-defined attributes are emitted.
	//
	// Support: Extended
	//
	// +listType=map
	// +listMapKey=name
	// +optional
	Attributes []Attribute `json:"attributes,omitempty"`
}

// TracingProvider identifies the tracing backend that receives generated spans.
//
// Support: Core for Service
//
// Support: Implementation-specific for any other resource
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
	BackendRef v1.BackendObjectReference `json:"backendRef"`

	// Headers specifies a list of custom headers to be added to the telemetry
	// export requests (e.g., for authentication).
	//
	// Support: Extended
	//
	// +optional
	// +kubebuilder:validation:MaxItems=16
	Headers []v1.HTTPHeader `json:"headers,omitempty"`
}

// ParentBasedSamplingMode defines the enablement mode for parent-based sampling.
type ParentBasedSamplingMode string

const (
	// ParentBasedSamplingModeEnabled explicitly enables parent-based sampling.
	ParentBasedSamplingModeEnabled ParentBasedSamplingMode = "Enabled"

	// ParentBasedSamplingModeDisabled explicitly disables parent-based sampling.
	ParentBasedSamplingModeDisabled ParentBasedSamplingMode = "Disabled"

	// ParentBasedSamplingModeImplementationDefault means that the code should
	// use the implementation's default behavior for parent-based sampling.
	ParentBasedSamplingModeImplementationDefault ParentBasedSamplingMode = "ImplementationDefault"
)

// ParentBasedSampling defines the sampling behavior when a request has a pre-existing upstream
// trace parent.
//
// Support: Extended
type ParentBasedSampling struct {
	// Mode explicitly controls if parent-based sampling is enabled. Valid values are "Enabled",
	// "Disabled", "ImplementationDefault".
	//
	// In the absence of this field, it defaults to "ImplementationDefault".
	//
	// Support: Extended
	//
	// +kubebuilder:validation:Enum=Enabled;Disabled;ImplementationDefault
	// +kubebuilder:default=ImplementationDefault
	Mode ParentBasedSamplingMode `json:"mode,omitempty"`

	// SamplingRate is the sampling rate to apply when parent-based sampling is active.
	//
	// This acts as a downsampling governor. It allows an operator to say: "I want to
	// respect the parent's decision, but only for 50% of those requests". Even if a
	// parent is already marked as "Sampled", this allows the Gateway to apply a secondary
	// filter so that it can respect the parent's intent while still controlling the volume
	// of spans reported.
	//
	// In the absence of this field, it defaults to 100% ({numerator: 100}).
	//
	// Support: Extended
	//
	// +optional
	// +kubebuilder:default={numerator: 100, denominator: 100}
	SamplingRate *v1.Fraction `json:"samplingRate,omitempty"`
}

// TelemetryPolicyStatus defines the observed state of TelemetryPolicy.
type TelemetryPolicyStatus struct {
	// For Policy Status API conventions, see:
	// https://gateway-api.sigs.k8s.io/geps/gep-713/#the-status-stanza-of-policy-objects
	//
	// Ancestors is a list of ancestor resources (specifically Gateway resources)
	// that are associated with the policy, and the status of the policy with
	// respect to each ancestor. When this policy attaches to a parent, the
	// controller that manages the parent and the ancestors MUST add an entry
	// to this list when the controller first sees the policy and SHOULD update
	// the entry as appropriate when the relevant ancestor is modified.
	//
	// For TelemetryPolicy, the ancestor MUST be the Gateway resource
	// referenced in spec.targetRefs.
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
	Ancestors []v1.PolicyAncestorStatus `json:"ancestors"`
}
