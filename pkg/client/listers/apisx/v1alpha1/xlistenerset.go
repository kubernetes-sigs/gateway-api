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

// Code generated by lister-gen. DO NOT EDIT.

package v1alpha1

import (
	labels "k8s.io/apimachinery/pkg/labels"
	listers "k8s.io/client-go/listers"
	cache "k8s.io/client-go/tools/cache"
	apisxv1alpha1 "sigs.k8s.io/gateway-api/apisx/v1alpha1"
)

// XListenerSetLister helps list XListenerSets.
// All objects returned here must be treated as read-only.
type XListenerSetLister interface {
	// List lists all XListenerSets in the indexer.
	// Objects returned here must be treated as read-only.
	List(selector labels.Selector) (ret []*apisxv1alpha1.XListenerSet, err error)
	// XListenerSets returns an object that can list and get XListenerSets.
	XListenerSets(namespace string) XListenerSetNamespaceLister
	XListenerSetListerExpansion
}

// xListenerSetLister implements the XListenerSetLister interface.
type xListenerSetLister struct {
	listers.ResourceIndexer[*apisxv1alpha1.XListenerSet]
}

// NewXListenerSetLister returns a new XListenerSetLister.
func NewXListenerSetLister(indexer cache.Indexer) XListenerSetLister {
	return &xListenerSetLister{listers.New[*apisxv1alpha1.XListenerSet](indexer, apisxv1alpha1.Resource("xlistenerset"))}
}

// XListenerSets returns an object that can list and get XListenerSets.
func (s *xListenerSetLister) XListenerSets(namespace string) XListenerSetNamespaceLister {
	return xListenerSetNamespaceLister{listers.NewNamespaced[*apisxv1alpha1.XListenerSet](s.ResourceIndexer, namespace)}
}

// XListenerSetNamespaceLister helps list and get XListenerSets.
// All objects returned here must be treated as read-only.
type XListenerSetNamespaceLister interface {
	// List lists all XListenerSets in the indexer for a given namespace.
	// Objects returned here must be treated as read-only.
	List(selector labels.Selector) (ret []*apisxv1alpha1.XListenerSet, err error)
	// Get retrieves the XListenerSet from the indexer for a given namespace and name.
	// Objects returned here must be treated as read-only.
	Get(name string) (*apisxv1alpha1.XListenerSet, error)
	XListenerSetNamespaceListerExpansion
}

// xListenerSetNamespaceLister implements the XListenerSetNamespaceLister
// interface.
type xListenerSetNamespaceLister struct {
	listers.ResourceIndexer[*apisxv1alpha1.XListenerSet]
}
