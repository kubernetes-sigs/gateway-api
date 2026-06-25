package mcp

import (
	"context"
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	gatewayv1 "sigs.k8s.io/gateway-api/apis/v1"
)

// GroupVersion is group version used to register these objects
var GroupVersion = schema.GroupVersion{Group: "gateway.networking.k8s.io", Version: "v1alpha1"}

// AgenticDataMeshFilter configures the Model Context Protocol (MCP) and LLM
// routing layer for the Kubernetes Gateway API. This allows gateways to route
// requests dynamically based on LLM parameters, token budgets, or A2A state.
type AgenticDataMeshFilter struct {
	// TargetModel defines the requested LLM architecture (e.g., "vllm-cluster", "sllm-local").
	TargetModel string `json:"targetModel,omitempty"`

	// AutonomyMode, if enabled, allows the Gateway to automatically failover
	// from vLLM to sLLM when Prometheus detects high inference latency.
	AutonomyMode bool `json:"autonomyMode,omitempty"`

	// EnableMCP injects the Model Context Protocol headers and negotiates
	// A2A (Agent-to-Agent) context automatically at the edge.
	EnableMCP bool `json:"enableMCP,omitempty"`

	// TokenBudget sets a maximum token throughput limit for this route.
	TokenBudget int32 `json:"tokenBudget,omitempty"`
}

// LLMRoute is an extension of HTTPRoute tailored for Agentic workloads.
type LLMRoute struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   LLMRouteSpec   `json:"spec,omitempty"`
	Status gatewayv1.RouteStatus `json:"status,omitempty"`
}

// LLMRouteSpec defines the desired state of LLMRoute
type LLMRouteSpec struct {
	// ParentRefs references the Gateway(s) that this Route wants to be attached to.
	ParentRefs []gatewayv1.ParentReference `json:"parentRefs,omitempty"`

	// Rules are a list of MCP routing rules.
	Rules []LLMRouteRule `json:"rules,omitempty"`
}

// LLMRouteRule defines semantics for matching an A2A request.
type LLMRouteRule struct {
	// Matches define conditions used for matching the MCP intent.
	Matches []gatewayv1.HTTPRouteMatch `json:"matches,omitempty"`

	// Filters define the Agentic Data Mesh filters applied to the request.
	Filters []AgenticDataMeshFilter `json:"filters,omitempty"`

	// BackendRefs defines the backend(s) where matching requests should be sent.
	BackendRefs []gatewayv1.HTTPBackendRef `json:"backendRefs,omitempty"`
}

// DeepCopyObject is required by the Kubernetes runtime.Scheme interface.
func (in *LLMRoute) DeepCopyObject() runtime.Object {
	if in == nil {
		return nil
	}
	out := new(LLMRoute)
	*out = *in
	return out
}

// RegisterLLMExtension registers the MCP Agent Layer with the Gateway API scheme.
func RegisterLLMExtension(scheme *runtime.Scheme) error {
	scheme.AddKnownTypes(GroupVersion,
		&LLMRoute{},
	)
	metav1.AddToGroupVersion(scheme, GroupVersion)
	fmt.Println("Successfully registered Agentic Data Mesh and MCP Layer into Kubernetes Gateway API.")
	return nil
}
