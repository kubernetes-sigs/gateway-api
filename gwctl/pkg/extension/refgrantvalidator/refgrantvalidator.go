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

package refgrantvalidator

import (
	"fmt"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/klog/v2"

	gatewayv1 "sigs.k8s.io/gateway-api/apis/v1"
	gatewayv1beta1 "sigs.k8s.io/gateway-api/apis/v1beta1"
	"sigs.k8s.io/gateway-api/gwctl/pkg/common"
	"sigs.k8s.io/gateway-api/gwctl/pkg/topology"
	topologygw "sigs.k8s.io/gateway-api/gwctl/pkg/topology/gateway"
)

const (
	extensionName = "ReferenceGrantsForBackend"
)

type Extension struct {
	fetcher referenceGrantFetcher
}

func NewExtension(fetcher referenceGrantFetcher) *Extension {
	return &Extension{fetcher: fetcher}
}

// Extension calculates the effective policies for all Gateways, HTTPRoutes, and
// Backends in the Graph.
func (a *Extension) Execute(graph *topology.Graph) error {
	graph.RemoveMetadata(extensionName)
	if err := a.discoverReferenceGrantsForBackends(graph); err != nil {
		return err
	}
	return a.validateHTTPRoutes(graph)
}

func (a *Extension) discoverReferenceGrantsForBackends(graph *topology.Graph) error {
	referenceGrantsByNamespace := make(map[string][]*gatewayv1beta1.ReferenceGrant)
	for _, backendNode := range graph.Nodes[common.ServiceGK] {
		backendNS := backendNode.Object.GetNamespace()

		referenceGrants, ok := referenceGrantsByNamespace[backendNS]
		if !ok {
			var err error
			referenceGrants, err = a.fetcher.FetchReferenceGrantsForNamespace(backendNS)
			if err != nil {
				return err
			}
			referenceGrantsByNamespace[backendNS] = referenceGrants
		}

		for _, referenceGrant := range referenceGrants {
			backendRef := backendNode.GKNN()
			if ReferenceGrantExposes(referenceGrant, backendRef) {
				klog.V(1).InfoS("ReferenceGrant exposes Backend",
					"referenceGrant", referenceGrant.GetNamespace()+"/"+referenceGrant.GetName(),
					"backendRef", backendRef.Namespace+"/"+backendRef.Name,
				)
				if err := a.putReferenceGrantInNode(backendNode, referenceGrant); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func (a *Extension) validateHTTPRoutes(graph *topology.Graph) error {
	for _, httpRouteNode := range graph.Nodes[common.HTTPRouteGK] {
		if httpRouteNode.Depth > graph.MaxDepth {
			klog.V(3).InfoS("Not validating HTTPRoute since it's depth is greater than the max depth",
				"extension", extensionName, "httpRouteNode.Depth", httpRouteNode.Depth, "MaxDepth", graph.MaxDepth,
			)
			continue
		}

		for backendGKNN, backendNode := range topologygw.HTTPRouteNode(httpRouteNode).Backends() {
			// Ensure that if this is a cross namespace reference, then it is accepted
			// through some ReferenceGrant.
			if httpRouteNode.GKNN().Namespace != backendGKNN.Namespace {
				backendNodeMetadata, err := Access(backendNode)
				if err != nil {
					return err
				}

				var referenceAccepted bool
				if backendNodeMetadata != nil {
					for _, referenceGrant := range backendNodeMetadata.ReferenceGrants {
						if ReferenceGrantAccepts(referenceGrant, httpRouteNode.GKNN()) {
							referenceAccepted = true
							break
						}
					}
				}
				if !referenceAccepted {
					err := common.ReferenceNotPermittedError{ReferenceFromTo: common.ReferenceFromTo{
						ReferringObject: httpRouteNode.GKNN(),
						ReferredObject:  backendGKNN,
					}}
					if err := a.putReferenceGrantErrorInNode(httpRouteNode, err); err != nil {
						return err
					}
					klog.V(1).InfoS("Reference not permitted", "from", httpRouteNode.GKNN(), "to", backendGKNN)
					continue
				}
			}
		}
	}
	return nil
}

func (a *Extension) putReferenceGrantInNode(node *topology.Node, referenceGrant *gatewayv1beta1.ReferenceGrant) error {
	if node.Metadata == nil {
		node.Metadata = map[string]any{}
	}
	if node.Metadata[extensionName] == nil {
		node.Metadata[extensionName] = &NodeMetadata{
			ReferenceGrants: make(map[common.GKNN]*gatewayv1beta1.ReferenceGrant),
			Errors:          make([]error, 0),
		}
	}

	data, err := Access(node)
	if err != nil {
		return err
	}
	gknn := common.GKNN{
		Group:     common.ReferenceGrantGK.Group,
		Kind:      common.ReferenceGrantGK.Kind,
		Namespace: referenceGrant.GetNamespace(),
		Name:      referenceGrant.GetName(),
	}
	data.ReferenceGrants[gknn] = referenceGrant
	return nil
}

func (a *Extension) putReferenceGrantErrorInNode(node *topology.Node, refGrantErr error) error {
	if node.Metadata == nil {
		node.Metadata = map[string]any{}
	}
	if node.Metadata[extensionName] == nil {
		node.Metadata[extensionName] = &NodeMetadata{
			ReferenceGrants: make(map[common.GKNN]*gatewayv1beta1.ReferenceGrant),
			Errors:          make([]error, 0),
		}
	}

	data, err := Access(node)
	if err != nil {
		return err
	}
	data.Errors = append(data.Errors, refGrantErr)
	return nil
}

type NodeMetadata struct {
	ReferenceGrants map[common.GKNN]*gatewayv1beta1.ReferenceGrant
	Errors          []error
}

func Access(node *topology.Node) (*NodeMetadata, error) {
	rawData, ok := node.Metadata[extensionName]
	if !ok || rawData == nil {
		klog.V(3).InfoS(fmt.Sprintf("no data found in node for %v", extensionName), "node", node.GKNN())
		return nil, nil
	}
	data, ok := rawData.(*NodeMetadata)
	if !ok {
		return nil, fmt.Errorf("unable to perform type assertion for %v in node %v", extensionName, node.GKNN())
	}
	return data, nil
}

// ReferenceGrantExposes returns true if the provided reference grant "exposes"
// the given resource. "Exposes" means that the resource is part of the "To"
// fields within the ReferenceGrant.
func ReferenceGrantExposes(referenceGrant *gatewayv1beta1.ReferenceGrant, resource common.GKNN) bool {
	if referenceGrant.GetNamespace() != resource.Namespace {
		return false
	}
	for _, to := range referenceGrant.Spec.To {
		if to.Group != gatewayv1.Group(resource.Group) {
			continue
		}
		if to.Kind != gatewayv1.Kind(resource.Kind) {
			continue
		}
		if to.Name == nil || len(*to.Name) == 0 || *to.Name == gatewayv1.ObjectName(resource.Name) {
			return true
		}
	}
	return false
}

// ReferenceGrantAccepts returns true if the provided reference grant "accepts"
// references from the given resource. "Accepts" means that the resource is part
// of the "From" fields within the ReferenceGrant.
func ReferenceGrantAccepts(referenceGrant *gatewayv1beta1.ReferenceGrant, resource common.GKNN) bool {
	resource.Name = ""
	for _, from := range referenceGrant.Spec.From {
		fromRef := common.GKNN{
			Group:     string(from.Group),
			Kind:      string(from.Kind),
			Namespace: string(from.Namespace),
		}
		if fromRef == resource {
			return true
		}
	}
	return false
}

type referenceGrantFetcher interface {
	FetchReferenceGrantsForNamespace(string) ([]*gatewayv1beta1.ReferenceGrant, error)
}

var _ referenceGrantFetcher = (*defaultReferenceGrantFetcher)(nil)

type defaultReferenceGrantFetcher struct {
	factory                        common.Factory
	additionalResourcesByNamespace map[string][]*unstructured.Unstructured
}

type referenceGrantFetcherOption func(*defaultReferenceGrantFetcher)

func WithAdditionalResources(resources []*unstructured.Unstructured) referenceGrantFetcherOption { //nolint:revive
	return func(f *defaultReferenceGrantFetcher) {
		for _, resource := range resources {
			if resource.GroupVersionKind().GroupKind() == common.ReferenceGrantGK {
				f.additionalResourcesByNamespace[resource.GetNamespace()] = append(f.additionalResourcesByNamespace[resource.GetNamespace()], resource)
			}
		}
	}
}

func NewDefaultReferenceGrantFetcher(factory common.Factory, options ...referenceGrantFetcherOption) *defaultReferenceGrantFetcher { //nolint:revive
	f := &defaultReferenceGrantFetcher{
		factory:                        factory,
		additionalResourcesByNamespace: make(map[string][]*unstructured.Unstructured),
	}
	for _, option := range options {
		option(f)
	}
	return f
}

func (f *defaultReferenceGrantFetcher) FetchReferenceGrantsForNamespace(namespace string) ([]*gatewayv1beta1.ReferenceGrant, error) {
	infos, err := f.factory.NewBuilder().
		Unstructured().
		Flatten().
		NamespaceParam(namespace).RequireNamespace().
		AllNamespaces(false).
		ResourceTypeOrNameArgs(true, []string{fmt.Sprintf("%v.%v", common.ReferenceGrantGK.Kind, common.ReferenceGrantGK.Group)}...).
		ContinueOnError().
		Do().
		Infos()
	if err != nil {
		return nil, err
	}

	var result []*gatewayv1beta1.ReferenceGrant
	for _, info := range infos {
		u, err := runtime.DefaultUnstructuredConverter.ToUnstructured(info.Object)
		if err != nil {
			return nil, err
		}
		refGrant := &gatewayv1beta1.ReferenceGrant{}
		if err := runtime.DefaultUnstructuredConverter.FromUnstructured(u, refGrant); err != nil {
			return nil, err
		}
		result = append(result, refGrant)
	}

	// Return any additional ReferenceGrants if they have been provided.
	for _, u := range f.additionalResourcesByNamespace[namespace] {
		refGrant := &gatewayv1beta1.ReferenceGrant{}
		if err := runtime.DefaultUnstructuredConverter.FromUnstructured(u.UnstructuredContent(), refGrant); err != nil {
			return nil, fmt.Errorf("converting local ReferenceGrant from Unstructurued to typed: %v", err)
		}
		result = append(result, refGrant)
	}

	return result, nil
}
