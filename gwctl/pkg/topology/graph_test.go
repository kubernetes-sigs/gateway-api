/*
Copyright 2024 The Kubernetes Authors.

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

package topology

import (
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"

	"sigs.k8s.io/gateway-api/gwctl/pkg/common"
)

func TestGraph_AddNode(t *testing.T) {
	graph := &Graph{}

	node1 := &Node{Object: buildUnstructured(common.GKNN{Group: "1", Kind: "2", Namespace: "3", Name: "4"})}
	node2 := &Node{Object: buildUnstructured(common.GKNN{Group: "1", Kind: "2", Namespace: "3", Name: "5"})}
	node3 := &Node{Object: buildUnstructured(common.GKNN{Group: "1", Kind: "2", Namespace: "6", Name: "7"})}
	node4 := &Node{Object: buildUnstructured(common.GKNN{Group: "1", Kind: "8", Namespace: "3", Name: "4"})}

	graph.AddNode(node1)
	graph.AddNode(node2)
	graph.AddNode(node3)
	graph.AddNode(node4)

	wantGraph := &Graph{
		Nodes: map[schema.GroupKind]map[types.NamespacedName]*Node{
			{Group: "1", Kind: "2"}: {
				{Namespace: "3", Name: "4"}: node1,
				{Namespace: "3", Name: "5"}: node2,
				{Namespace: "6", Name: "7"}: node3,
			},
			{Group: "1", Kind: "8"}: {
				{Namespace: "3", Name: "4"}: node4,
			},
		},
	}

	if diff := cmp.Diff(wantGraph, graph); diff != "" {
		t.Fatalf("Unexpected diff in graph after AddNode operations: (-want, +got)\n%v", diff)
	}
}

func TestGraph_AddEdge(t *testing.T) {
	graph := &Graph{}

	gknn1 := common.GKNN{Group: "1", Kind: "2", Namespace: "3", Name: "4"}
	node1 := &Node{Object: buildUnstructured(gknn1)}

	gknn2 := common.GKNN{Group: "1", Kind: "2", Namespace: "3", Name: "5"}
	node2 := &Node{Object: buildUnstructured(gknn2)}

	gknn3 := common.GKNN{Group: "1", Kind: "2", Namespace: "6", Name: "7"}
	node3 := &Node{Object: buildUnstructured(gknn3)}

	gknn4 := common.GKNN{Group: "1", Kind: "8", Namespace: "3", Name: "4"}
	node4 := &Node{Object: buildUnstructured(gknn4)}

	childRelation := &Relation{Name: "child"}
	parentRelation := &Relation{Name: "parent"}

	graph.AddEdge(node1, node2, childRelation)
	graph.AddEdge(node1, node3, childRelation)
	graph.AddEdge(node1, node4, parentRelation)

	wantNode1 := &Node{
		Object: buildUnstructured(gknn1),
		OutNeighbors: map[*Relation]map[common.GKNN]*Node{
			childRelation: {
				gknn2: node2,
				gknn3: node3,
			},
			parentRelation: {
				gknn4: node4,
			},
		},
	}
	wantNode2 := &Node{
		Object: buildUnstructured(gknn2),
		InNeighbors: map[*Relation]map[common.GKNN]*Node{
			childRelation: {
				gknn1: node1,
			},
		},
	}
	cmpopts := []cmp.Option{cmp.Transformer("NeighborsTransformer", NeighborsTransformer)}
	if diff := cmp.Diff(wantNode1, node1, cmpopts...); diff != "" {
		t.Errorf("Unexpected diff in node1 after AddEdge operations: (-want, +got)\n%v", diff)
	}
	if diff := cmp.Diff(wantNode2, node2, cmpopts...); diff != "" {
		t.Errorf("Unexpected diff in node2 after AddEdge operations: (-want, +got)\n%v", diff)
	}
}

func NeighborsTransformer(neighbors map[*Relation]map[common.GKNN]*Node) map[*Relation]map[common.GKNN]bool {
	result := make(map[*Relation]map[common.GKNN]bool)
	for relation, nodeMap := range neighbors {
		result[relation] = make(map[common.GKNN]bool)
		for nodeGKNN := range nodeMap {
			result[relation][nodeGKNN] = true
		}
	}
	return result
}

func TestBuilder(t *testing.T) {
	gknn1 := common.GKNN{Group: "1", Kind: "1", Namespace: "1", Name: "1"}
	gknn2 := common.GKNN{Group: "2", Kind: "2", Namespace: "2", Name: "2"}
	gknn3 := common.GKNN{Group: "3", Kind: "3", Namespace: "3", Name: "3"}
	gknn4 := common.GKNN{Group: "4", Kind: "4", Namespace: "4", Name: "4"}
	gknn5 := common.GKNN{Group: "5", Kind: "5", Namespace: "5", Name: "5"} // Unreachable
	gknn6 := common.GKNN{Group: "6", Kind: "6", Namespace: "6", Name: "6"} // Unreachable

	relation2To1 := &Relation{
		From: gknn2.GroupKind(),
		To:   gknn1.GroupKind(),
		Name: "gk2_to_gk1",
		NeighborFunc: func(*unstructured.Unstructured) []common.GKNN {
			return []common.GKNN{gknn1}
		},
	}
	relation2To3 := &Relation{
		From: gknn2.GroupKind(),
		To:   gknn3.GroupKind(),
		Name: "gk2_to_gk3",
		NeighborFunc: func(*unstructured.Unstructured) []common.GKNN {
			return []common.GKNN{gknn3}
		},
	}
	relation4To3 := &Relation{
		From: gknn4.GroupKind(),
		To:   gknn3.GroupKind(),
		Name: "gk4_to_gk3",
		NeighborFunc: func(*unstructured.Unstructured) []common.GKNN {
			return []common.GKNN{gknn3}
		},
	}
	relation4To5 := &Relation{
		From:         gknn4.GroupKind(),
		To:           gknn5.GroupKind(),
		Name:         "gk4_to_gk5",
		NeighborFunc: func(*unstructured.Unstructured) []common.GKNN { return nil },
	}
	relation6To4 := &Relation{
		From:         gknn6.GroupKind(),
		To:           gknn4.GroupKind(),
		Name:         "gk6_to_gk4",
		NeighborFunc: func(*unstructured.Unstructured) []common.GKNN { return nil },
	}

	u1 := buildUnstructured(gknn1)
	u2 := buildUnstructured(gknn2)
	u3 := buildUnstructured(gknn3)
	u4 := buildUnstructured(gknn4)
	fakeFetcher := &fakeGroupKindFetcher{
		data: map[schema.GroupKind][]*unstructured.Unstructured{
			gknn1.GroupKind(): {u1},
			gknn2.GroupKind(): {u2},
			gknn3.GroupKind(): {u3},
			gknn4.GroupKind(): {u4},
		},
	}

	sources := []*unstructured.Unstructured{buildUnstructured(gknn3)}
	graph, err := NewBuilder(fakeFetcher).
		StartFrom(sources).
		UseRelationship(relation2To1).
		UseRelationship(relation2To3).
		UseRelationship(relation4To3).
		UseRelationship(relation4To5).
		UseRelationship(relation6To4).
		Build()
	if err != nil {
		t.Fatalf("Builder...Build() failed with error %v; want no errors", err)
	}

	wantGraph := &Graph{}
	node1 := &Node{Object: u1, Depth: 2}
	node2 := &Node{Object: u2, Depth: 1}
	node3 := &Node{Object: u3, Depth: 0}
	node4 := &Node{Object: u4, Depth: 1}
	wantGraph.AddNode(node1)
	wantGraph.AddNode(node2)
	wantGraph.AddNode(node3)
	wantGraph.AddNode(node4)
	wantGraph.AddEdge(node2, node1, relation2To1)
	wantGraph.AddEdge(node2, node3, relation2To3)
	wantGraph.AddEdge(node4, node3, relation4To3)

	if diff := cmp.Diff(wantGraph.Nodes, graph.Nodes); diff != "" {
		t.Fatalf("Builder...Build(): Unexpected diff in graph: (-want, +got)\n%v", diff)
	}
}

func buildUnstructured(gknn common.GKNN) *unstructured.Unstructured {
	return &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": fmt.Sprintf("%v/v1", gknn.Group),
			"kind":       gknn.Kind,
			"metadata": map[string]interface{}{
				"name":      gknn.Name,
				"namespace": gknn.Namespace,
			},
			"spec": map[string]interface{}{
				"key": "value",
			},
		},
	}
}

type fakeGroupKindFetcher struct {
	data map[schema.GroupKind][]*unstructured.Unstructured
}

func (f *fakeGroupKindFetcher) Fetch(gk schema.GroupKind) ([]*unstructured.Unstructured, error) {
	return f.data[gk], nil
}
