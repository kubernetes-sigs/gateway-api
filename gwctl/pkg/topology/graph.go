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

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/klog/v2"

	"sigs.k8s.io/gateway-api/gwctl/pkg/common"
)

const (
	DefaultGraphMaxDepth = 3
)

type Graph struct {
	Nodes   map[schema.GroupKind]map[types.NamespacedName]*Node
	Sources []*Node
	// MaxDepth represents the value of the maximum depth parameter used while
	// performing the BFS from Source nodes. Note that terminal nodes can have a
	// depth equal to MaxDepth+1. Any validations that need to happen should be
	// done for nodes <= MaxDepth. This ensures that any external references
	// from nodes <= MaxDepth can be validated.
	MaxDepth  int
	Relations []*Relation
}

func (g *Graph) AddNode(node *Node) {
	klog.V(3).InfoS("AddNode", "node", node.GKNN())
	if g.Nodes == nil {
		g.Nodes = make(map[schema.GroupKind]map[types.NamespacedName]*Node)
	}
	if g.Nodes[node.GKNN().GroupKind()] == nil {
		g.Nodes[node.GKNN().GroupKind()] = make(map[types.NamespacedName]*Node)
	}
	g.Nodes[node.GKNN().GroupKind()][node.GKNN().NamespacedName()] = node
}

func (g *Graph) DeleteNode(node *Node) {
	klog.V(3).InfoS("DeleteNode", "node", node.GKNN())
	if g.Nodes == nil {
		return
	}
	if g.Nodes[node.GKNN().GroupKind()] == nil {
		return
	}
	delete(g.Nodes[node.GKNN().GroupKind()], node.GKNN().NamespacedName())
	if len(g.Nodes[node.GKNN().GroupKind()]) == 0 {
		delete(g.Nodes, node.GKNN().GroupKind())
	}
	return
}

func (g *Graph) DeleteNodeUsingGKNN(nodeGKNN common.GKNN) {
	klog.V(3).InfoS("DeleteNodeUsingGKNN", "nodeGKNN", nodeGKNN)
	if !g.HasNode(nodeGKNN) {
		return
	}
	g.DeleteNode(g.Nodes[nodeGKNN.GroupKind()][nodeGKNN.NamespacedName()])
}

func (g *Graph) HasNode(nodeGKNN common.GKNN) bool {
	if g.Nodes == nil {
		return false
	}
	if g.Nodes[nodeGKNN.GroupKind()] == nil {
		return false
	}
	return g.Nodes[nodeGKNN.GroupKind()][nodeGKNN.NamespacedName()] != nil
}

func (g *Graph) AddEdge(from *Node, to *Node, relation *Relation) {
	klog.V(3).InfoS("AddEdge", "from", from.GKNN(), "to", to.GKNN())
	if from.OutNeighbors == nil {
		from.OutNeighbors = make(map[*Relation]map[common.GKNN]*Node)
	}
	if from.OutNeighbors[relation] == nil {
		from.OutNeighbors[relation] = make(map[common.GKNN]*Node)
	}
	from.OutNeighbors[relation][to.GKNN()] = to

	if to.InNeighbors == nil {
		to.InNeighbors = make(map[*Relation]map[common.GKNN]*Node)
	}
	if to.InNeighbors[relation] == nil {
		to.InNeighbors[relation] = make(map[common.GKNN]*Node)
	}
	to.InNeighbors[relation][from.GKNN()] = from
}

func (g *Graph) RemoveEdge(from *Node, to *Node, relation *Relation) {
	klog.V(3).InfoS("RemoveEdge", "from", from.GKNN(), "to", to.GKNN())
	delete(from.OutNeighbors[relation], to.GKNN())
	if len(from.OutNeighbors[relation]) == 0 {
		delete(from.OutNeighbors, relation)
	}

	delete(to.InNeighbors[relation], from.GKNN())
	if len(to.InNeighbors[relation]) == 0 {
		delete(to.InNeighbors, relation)
	}
}

func (g *Graph) RemoveMetadata(category string) {
	for gk := range g.Nodes {
		for nn := range g.Nodes[gk] {
			node := g.Nodes[gk][nn]
			if node.Metadata != nil {
				delete(node.Metadata, category)
			}
		}
	}
}

type Node struct {
	Object       *unstructured.Unstructured
	InNeighbors  map[*Relation]map[common.GKNN]*Node
	OutNeighbors map[*Relation]map[common.GKNN]*Node
	Depth        int
	Metadata     map[string]any
}

func (n *Node) GKNN() common.GKNN {
	return common.GKNN{
		Group:     n.Object.GroupVersionKind().Group,
		Kind:      n.Object.GroupVersionKind().Kind,
		Namespace: n.Object.GetNamespace(),
		Name:      n.Object.GetName(),
	}
}

func MustAccessObject[T any](node *Node, concreteObj T) T {
	if err := runtime.DefaultUnstructuredConverter.FromUnstructured(node.Object.UnstructuredContent(), concreteObj); err != nil {
		panic(fmt.Sprintf("failed to convert unstructured %v to structured: %v", node.GKNN().GroupKind(), err))
	}
	return concreteObj
}

type NeighborFunc func(*unstructured.Unstructured) []common.GKNN

type Relation struct {
	From         schema.GroupKind
	To           schema.GroupKind
	Name         string
	NeighborFunc NeighborFunc
}

type Builder struct {
	Sources   []*unstructured.Unstructured
	Relations []*Relation
	Fetcher   common.GroupKindFetcher
	MaxDepth  int
}

func NewBuilder(fetcher common.GroupKindFetcher) *Builder {
	return &Builder{
		Fetcher:  fetcher,
		MaxDepth: DefaultGraphMaxDepth,
	}
}

func (b *Builder) StartFrom(sources []*unstructured.Unstructured) *Builder {
	b.Sources = sources
	return b
}

func (b *Builder) UseRelationship(relation *Relation) *Builder {
	b.Relations = append(b.Relations, relation)
	return b
}

func (b *Builder) UseRelationships(relations []*Relation) *Builder {
	b.Relations = append(b.Relations, relations...)
	return b
}

func (b *Builder) WithMaxDepth(maxDepth int) *Builder {
	b.MaxDepth = maxDepth
	return b
}

func (b *Builder) Build() (*Graph, error) {
	graph := &Graph{
		MaxDepth:  b.MaxDepth,
		Relations: b.Relations,
	}

	for _, obj := range b.Sources {
		node := &Node{Object: obj}
		graph.Sources = append(graph.Sources, node)
		graph.AddNode(node)
	}

	// Perform BFS from source GroupKinds to figure out distinct other
	// GroupKinds that need to be fetched.
	allGroupKinds := b.determineUniqueGroupKinds()
	if len(allGroupKinds) == 1 {
		// This case means there's only one GroupKind which is the same as the
		// Sources, so nothing needs to be done.
		return graph, nil
	}
	// Fetch all relevant resources and add them to the graph as Nodes. At a
	// later point, we will remove the resources which are not relevant.
	const inf = 100000000
	for _, groupKind := range allGroupKinds {
		klog.V(3).InfoS("Fetching resources", "groupKind", groupKind)
		resources, err := b.Fetcher.Fetch(groupKind)
		if err != nil {
			return nil, err
		}
		for _, resource := range resources {
			node := &Node{Object: resource, Depth: inf}
			if !graph.HasNode(node.GKNN()) {
				graph.AddNode(node)
			}
		}
	}

	// Connect related resources.
	for _, relation := range b.Relations {
		for _, fromNode := range graph.Nodes[relation.From] {
			for _, toNodeGKNN := range relation.NeighborFunc(fromNode.Object) {
				if _, ok := graph.Nodes[toNodeGKNN.GroupKind()]; !ok {
					continue
				}
				toNode := graph.Nodes[toNodeGKNN.GroupKind()][toNodeGKNN.NamespacedName()]
				if toNode != nil {
					graph.AddEdge(fromNode, toNode, relation)
				}
			}
		}
	}

	// Perform BFS.

	q := []*Node{} // q is a Queue used in the BFS.

	// Initialize the sources for the BFS
	for _, source := range b.Sources {
		gknn := (&Node{Object: source}).GKNN()
		node := graph.Nodes[gknn.GroupKind()][gknn.NamespacedName()]
		node.Depth = 0
		q = append(q, node)
	}

	for len(q) != 0 {
		u := q[0]
		q = q[1:]

		if u.Depth+1 > b.MaxDepth+1 {
			break
		}

		// Don't expand from Namespaces to other resources.
		if u.GKNN().GroupKind() == common.NamespaceGK {
			continue
		}

		// Don't expand from GatewayClasses if it is not the source node.
		// TODO: Find appropriate ways to encode this with the
		//   topology/relations.
		if u.GKNN().GroupKind() == common.GatewayClassGK && u.Depth != 0 {
			continue
		}

		allNeighbors := []map[*Relation]map[common.GKNN]*Node{
			u.InNeighbors,
			u.OutNeighbors,
		}

		// For vertex u, find all adjacent vertices v.
		for _, neighbor := range allNeighbors {
			for _, nodes := range neighbor {
				for _, v := range nodes {
					visited := v.Depth < inf
					if visited {
						continue
					}
					v.Depth = u.Depth + 1
					q = append(q, v)
				}
			}
		}
	}

	// BFS is now complete. Delete all Nodes which still have infinite depth.
	for gk, nodes := range graph.Nodes {
		for _, u := range nodes {
			if u.Depth < inf {
				continue
			}

			// For each vertex u, find all vertices v which have an outgoing
			// edge to u (ie. u has an incoming edge v -> u)
			for relation, neighbors := range u.InNeighbors {
				for _, v := range neighbors {
					graph.RemoveEdge(v, u, relation)
				}
			}

			graph.DeleteNode(u)
		}

		if len(nodes) == 0 {
			delete(graph.Nodes, gk)
		}
	}

	return graph, nil
}

func (b *Builder) determineUniqueGroupKinds() []schema.GroupKind {
	result := []schema.GroupKind{} // result is the set of unique GroupKinds having depth <= b.MaxDepth

	q := []schema.GroupKind{} // q is a Queue used in the BFS.
	visited := map[schema.GroupKind]bool{}
	depth := map[schema.GroupKind]int{}

	// Initialize the sources for the BFS
	for _, source := range b.Sources {
		gk := source.GroupVersionKind().GroupKind()
		if !visited[gk] {
			result = append(result, gk)
			visited[gk] = true
			q = append(q, gk)
			depth[gk] = 0
		}
	}

	for len(q) != 0 {
		u := q[0]
		q = q[1:]

		// For vertex u, find all adjacent vertices v.
		for _, relation := range b.Relations {
			if relation.From != u && relation.To != u {
				continue
			}
			v := relation.To
			if u == v {
				v = relation.From
			}
			if visited[v] {
				continue
			}

			visited[v] = true
			depth[v] = depth[u] + 1
			q = append(q, v)
			if depth[v] <= b.MaxDepth {
				result = append(result, v)
			} else {
				return result
			}
		}
	}

	return result
}
