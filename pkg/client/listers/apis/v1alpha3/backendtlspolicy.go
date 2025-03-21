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

package v1alpha3

import (
	labels "k8s.io/apimachinery/pkg/labels"
	listers "k8s.io/client-go/listers"
	cache "k8s.io/client-go/tools/cache"
	apisv1alpha3 "sigs.k8s.io/gateway-api/apis/v1alpha3"
)

// BackendTLSPolicyLister helps list BackendTLSPolicies.
// All objects returned here must be treated as read-only.
type BackendTLSPolicyLister interface {
	// List lists all BackendTLSPolicies in the indexer.
	// Objects returned here must be treated as read-only.
	List(selector labels.Selector) (ret []*apisv1alpha3.BackendTLSPolicy, err error)
	// BackendTLSPolicies returns an object that can list and get BackendTLSPolicies.
	BackendTLSPolicies(namespace string) BackendTLSPolicyNamespaceLister
	BackendTLSPolicyListerExpansion
}

// backendTLSPolicyLister implements the BackendTLSPolicyLister interface.
type backendTLSPolicyLister struct {
	listers.ResourceIndexer[*apisv1alpha3.BackendTLSPolicy]
}

// NewBackendTLSPolicyLister returns a new BackendTLSPolicyLister.
func NewBackendTLSPolicyLister(indexer cache.Indexer) BackendTLSPolicyLister {
	return &backendTLSPolicyLister{listers.New[*apisv1alpha3.BackendTLSPolicy](indexer, apisv1alpha3.Resource("backendtlspolicy"))}
}

// BackendTLSPolicies returns an object that can list and get BackendTLSPolicies.
func (s *backendTLSPolicyLister) BackendTLSPolicies(namespace string) BackendTLSPolicyNamespaceLister {
	return backendTLSPolicyNamespaceLister{listers.NewNamespaced[*apisv1alpha3.BackendTLSPolicy](s.ResourceIndexer, namespace)}
}

// BackendTLSPolicyNamespaceLister helps list and get BackendTLSPolicies.
// All objects returned here must be treated as read-only.
type BackendTLSPolicyNamespaceLister interface {
	// List lists all BackendTLSPolicies in the indexer for a given namespace.
	// Objects returned here must be treated as read-only.
	List(selector labels.Selector) (ret []*apisv1alpha3.BackendTLSPolicy, err error)
	// Get retrieves the BackendTLSPolicy from the indexer for a given namespace and name.
	// Objects returned here must be treated as read-only.
	Get(name string) (*apisv1alpha3.BackendTLSPolicy, error)
	BackendTLSPolicyNamespaceListerExpansion
}

// backendTLSPolicyNamespaceLister implements the BackendTLSPolicyNamespaceLister
// interface.
type backendTLSPolicyNamespaceLister struct {
	listers.ResourceIndexer[*apisv1alpha3.BackendTLSPolicy]
}
