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

package gateway

import (
	"bytes"
	"log"

	graphviz "github.com/goccy/go-graphviz"
	"github.com/goccy/go-graphviz/cgraph"

	"sigs.k8s.io/gateway-api/gwctl/pkg/common"
	"sigs.k8s.io/gateway-api/gwctl/pkg/topology"
)

// TODO:
//   - Show policy nodes. Attempt to group policy nodes along with their target
//     nodes in a single subgraph so they get rendered closer together.
func ToDot(gwctlGraph *topology.Graph) ([]byte, error) {
	g := graphviz.New()
	cGraph, err := g.Graph()
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := cGraph.Close(); err != nil {
			log.Fatal(err)
		}
		g.Close()
	}()
	cGraph.SetRankDir(cgraph.BTRank)

	cNodeMap := map[common.GKNN]*cgraph.Node{}

	// Create nodes.
	for _, nodeMap := range gwctlGraph.Nodes {
		for _, node := range nodeMap {
			cNode, err := cGraph.CreateNode(node.GKNN().String())
			if err != nil {
				return nil, err
			}
			cNodeMap[node.GKNN()] = cNode
			cNode.SetStyle(cgraph.FilledNodeStyle)
			cNode.SetFillColor(nodeColor(node))

			// Set the Node label
			gk := node.GKNN().GroupKind()
			if gk.Group == common.GatewayGK.Group {
				gk.Group = ""
			}
			name := node.GKNN().NamespacedName().String()
			if node.GKNN().Namespace == "" {
				name = node.GKNN().Name
			}
			cNode.SetLabel(gk.String() + "\n" + name)
		}
	}

	// Create edges.
	for fromNodeGKNN, cFromNode := range cNodeMap {
		fromNode := gwctlGraph.Nodes[fromNodeGKNN.GroupKind()][fromNodeGKNN.NamespacedName()]

		for relation, outNodeMap := range fromNode.OutNeighbors {
			for toNodeGKNN := range outNodeMap {
				cToNode := cNodeMap[toNodeGKNN]

				// If this is an edge from an HTTPRoute to a Service, then
				// reverse the direction of the edge (to affect the rank), and
				// then reverse the display again to show the correct direction.
				// The end result being that Services now get assigned the
				// correct rank.
				reverse := (fromNode.GKNN().GroupKind() == common.HTTPRouteGK && toNodeGKNN.GroupKind() == common.ServiceGK) ||
					(fromNode.GKNN().GroupKind() == common.GatewayGK && toNodeGKNN.GroupKind() == common.NamespaceGK)
				u, v := cFromNode, cToNode
				if reverse {
					u, v = v, u
				}

				e, err := cGraph.CreateEdge(relation.Name, u, v)
				if err != nil {
					return nil, err
				}
				e.SetLabel(relation.Name)
				if reverse {
					e.SetDir(cgraph.BackDir)
				}
				// Create a dotted line for the relation to the namespace.
				if toNodeGKNN.Kind == common.NamespaceGK.Kind {
					e.SetStyle(cgraph.DottedEdgeStyle)
				}
			}
		}
	}

	var buf bytes.Buffer
	if err := g.Render(cGraph, "dot", &buf); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func nodeColor(node *topology.Node) string {
	switch node.GKNN().GroupKind() {
	case common.NamespaceGK:
		return "#d08770"
	case common.GatewayClassGK:
		return "#e5e9f0"
	case common.GatewayGK:
		return "#ebcb8b"
	case common.HTTPRouteGK:
		return "#a3be8c"
	case common.ServiceGK:
		return "#88c0d0"
	}
	return "#d8dee9"
}
