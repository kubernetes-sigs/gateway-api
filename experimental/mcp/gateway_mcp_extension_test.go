package mcp

import (
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

func TestDeepCopyObject(t *testing.T) {
	route := &LLMRoute{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-route",
		},
		Spec: LLMRouteSpec{
			Rules: []LLMRouteRule{
				{
					Filters: []AgenticDataMeshFilter{
						{
							TargetModel: "vllm-cluster",
							EnableMCP:   true,
						},
					},
				},
			},
		},
	}

	copied := route.DeepCopyObject()
	copiedRoute, ok := copied.(*LLMRoute)
	if !ok {
		t.Fatalf("Expected copied object to be of type *LLMRoute")
	}

	if copiedRoute.Name != "test-route" {
		t.Errorf("Expected name to be 'test-route', got %s", copiedRoute.Name)
	}

	if len(copiedRoute.Spec.Rules) != 1 {
		t.Fatalf("Expected 1 rule, got %d", len(copiedRoute.Spec.Rules))
	}

	if copiedRoute.Spec.Rules[0].Filters[0].TargetModel != "vllm-cluster" {
		t.Errorf("Expected TargetModel 'vllm-cluster', got %s", copiedRoute.Spec.Rules[0].Filters[0].TargetModel)
	}
}

func TestRegisterLLMExtension(t *testing.T) {
	scheme := runtime.NewScheme()
	err := RegisterLLMExtension(scheme)
	if err != nil {
		t.Fatalf("Failed to register LLM extension: %v", err)
	}

	gvk := schema.GroupVersionKind{
		Group:   "gateway.networking.k8s.io",
		Version: "v1alpha1",
		Kind:    "LLMRoute",
	}

	obj, err := scheme.New(gvk)
	if err != nil {
		t.Fatalf("Failed to create LLMRoute from scheme: %v", err)
	}

	if _, ok := obj.(*LLMRoute); !ok {
		t.Errorf("Expected created object to be *LLMRoute")
	}
}
