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

package v1alpha2

import (
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/listers"
	"k8s.io/client-go/tools/cache"
	v1alpha2 "sigs.k8s.io/gateway-api/apis/v1alpha2"
)

// BackendTrafficPolicyLister helps list BackendTrafficPolicies.
// All objects returned here must be treated as read-only.
type BackendTrafficPolicyLister interface {
	// List lists all BackendTrafficPolicies in the indexer.
	// Objects returned here must be treated as read-only.
	List(selector labels.Selector) (ret []*v1alpha2.BackendTrafficPolicy, err error)
	// BackendTrafficPolicies returns an object that can list and get BackendTrafficPolicies.
	BackendTrafficPolicies(namespace string) BackendTrafficPolicyNamespaceLister
	BackendTrafficPolicyListerExpansion
}

// backendTrafficPolicyLister implements the BackendTrafficPolicyLister interface.
type backendTrafficPolicyLister struct {
	listers.ResourceIndexer[*v1alpha2.BackendTrafficPolicy]
}

// NewBackendTrafficPolicyLister returns a new BackendTrafficPolicyLister.
func NewBackendTrafficPolicyLister(indexer cache.Indexer) BackendTrafficPolicyLister {
	return &backendTrafficPolicyLister{listers.New[*v1alpha2.BackendTrafficPolicy](indexer, v1alpha2.Resource("backendtrafficpolicy"))}
}

// BackendTrafficPolicies returns an object that can list and get BackendTrafficPolicies.
func (s *backendTrafficPolicyLister) BackendTrafficPolicies(namespace string) BackendTrafficPolicyNamespaceLister {
	return backendTrafficPolicyNamespaceLister{listers.NewNamespaced[*v1alpha2.BackendTrafficPolicy](s.ResourceIndexer, namespace)}
}

// BackendTrafficPolicyNamespaceLister helps list and get BackendTrafficPolicies.
// All objects returned here must be treated as read-only.
type BackendTrafficPolicyNamespaceLister interface {
	// List lists all BackendTrafficPolicies in the indexer for a given namespace.
	// Objects returned here must be treated as read-only.
	List(selector labels.Selector) (ret []*v1alpha2.BackendTrafficPolicy, err error)
	// Get retrieves the BackendTrafficPolicy from the indexer for a given namespace and name.
	// Objects returned here must be treated as read-only.
	Get(name string) (*v1alpha2.BackendTrafficPolicy, error)
	BackendTrafficPolicyNamespaceListerExpansion
}

// backendTrafficPolicyNamespaceLister implements the BackendTrafficPolicyNamespaceLister
// interface.
type backendTrafficPolicyNamespaceLister struct {
	listers.ResourceIndexer[*v1alpha2.BackendTrafficPolicy]
}
