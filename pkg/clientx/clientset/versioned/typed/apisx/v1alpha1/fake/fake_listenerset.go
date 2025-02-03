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

// Code generated by client-gen. DO NOT EDIT.

package fake

import (
	"context"
	json "encoding/json"
	"fmt"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	labels "k8s.io/apimachinery/pkg/labels"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	testing "k8s.io/client-go/testing"
	apisxv1alpha1 "sigs.k8s.io/gateway-api/apisx/applyconfiguration/apisx/v1alpha1"
	v1alpha1 "sigs.k8s.io/gateway-api/apisx/v1alpha1"
)

// FakeListenerSets implements ListenerSetInterface
type FakeListenerSets struct {
	Fake *FakeGatewayV1alpha1
	ns   string
}

var listenersetsResource = v1alpha1.SchemeGroupVersion.WithResource("listenersets")

var listenersetsKind = v1alpha1.SchemeGroupVersion.WithKind("ListenerSet")

// Get takes name of the listenerSet, and returns the corresponding listenerSet object, and an error if there is any.
func (c *FakeListenerSets) Get(ctx context.Context, name string, options v1.GetOptions) (result *v1alpha1.ListenerSet, err error) {
	emptyResult := &v1alpha1.ListenerSet{}
	obj, err := c.Fake.
		Invokes(testing.NewGetActionWithOptions(listenersetsResource, c.ns, name, options), emptyResult)

	if obj == nil {
		return emptyResult, err
	}
	return obj.(*v1alpha1.ListenerSet), err
}

// List takes label and field selectors, and returns the list of ListenerSets that match those selectors.
func (c *FakeListenerSets) List(ctx context.Context, opts v1.ListOptions) (result *v1alpha1.ListenerSetList, err error) {
	emptyResult := &v1alpha1.ListenerSetList{}
	obj, err := c.Fake.
		Invokes(testing.NewListActionWithOptions(listenersetsResource, listenersetsKind, c.ns, opts), emptyResult)

	if obj == nil {
		return emptyResult, err
	}

	label, _, _ := testing.ExtractFromListOptions(opts)
	if label == nil {
		label = labels.Everything()
	}
	list := &v1alpha1.ListenerSetList{ListMeta: obj.(*v1alpha1.ListenerSetList).ListMeta}
	for _, item := range obj.(*v1alpha1.ListenerSetList).Items {
		if label.Matches(labels.Set(item.Labels)) {
			list.Items = append(list.Items, item)
		}
	}
	return list, err
}

// Watch returns a watch.Interface that watches the requested listenerSets.
func (c *FakeListenerSets) Watch(ctx context.Context, opts v1.ListOptions) (watch.Interface, error) {
	return c.Fake.
		InvokesWatch(testing.NewWatchActionWithOptions(listenersetsResource, c.ns, opts))

}

// Create takes the representation of a listenerSet and creates it.  Returns the server's representation of the listenerSet, and an error, if there is any.
func (c *FakeListenerSets) Create(ctx context.Context, listenerSet *v1alpha1.ListenerSet, opts v1.CreateOptions) (result *v1alpha1.ListenerSet, err error) {
	emptyResult := &v1alpha1.ListenerSet{}
	obj, err := c.Fake.
		Invokes(testing.NewCreateActionWithOptions(listenersetsResource, c.ns, listenerSet, opts), emptyResult)

	if obj == nil {
		return emptyResult, err
	}
	return obj.(*v1alpha1.ListenerSet), err
}

// Update takes the representation of a listenerSet and updates it. Returns the server's representation of the listenerSet, and an error, if there is any.
func (c *FakeListenerSets) Update(ctx context.Context, listenerSet *v1alpha1.ListenerSet, opts v1.UpdateOptions) (result *v1alpha1.ListenerSet, err error) {
	emptyResult := &v1alpha1.ListenerSet{}
	obj, err := c.Fake.
		Invokes(testing.NewUpdateActionWithOptions(listenersetsResource, c.ns, listenerSet, opts), emptyResult)

	if obj == nil {
		return emptyResult, err
	}
	return obj.(*v1alpha1.ListenerSet), err
}

// UpdateStatus was generated because the type contains a Status member.
// Add a +genclient:noStatus comment above the type to avoid generating UpdateStatus().
func (c *FakeListenerSets) UpdateStatus(ctx context.Context, listenerSet *v1alpha1.ListenerSet, opts v1.UpdateOptions) (result *v1alpha1.ListenerSet, err error) {
	emptyResult := &v1alpha1.ListenerSet{}
	obj, err := c.Fake.
		Invokes(testing.NewUpdateSubresourceActionWithOptions(listenersetsResource, "status", c.ns, listenerSet, opts), emptyResult)

	if obj == nil {
		return emptyResult, err
	}
	return obj.(*v1alpha1.ListenerSet), err
}

// Delete takes name of the listenerSet and deletes it. Returns an error if one occurs.
func (c *FakeListenerSets) Delete(ctx context.Context, name string, opts v1.DeleteOptions) error {
	_, err := c.Fake.
		Invokes(testing.NewDeleteActionWithOptions(listenersetsResource, c.ns, name, opts), &v1alpha1.ListenerSet{})

	return err
}

// DeleteCollection deletes a collection of objects.
func (c *FakeListenerSets) DeleteCollection(ctx context.Context, opts v1.DeleteOptions, listOpts v1.ListOptions) error {
	action := testing.NewDeleteCollectionActionWithOptions(listenersetsResource, c.ns, opts, listOpts)

	_, err := c.Fake.Invokes(action, &v1alpha1.ListenerSetList{})
	return err
}

// Patch applies the patch and returns the patched listenerSet.
func (c *FakeListenerSets) Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts v1.PatchOptions, subresources ...string) (result *v1alpha1.ListenerSet, err error) {
	emptyResult := &v1alpha1.ListenerSet{}
	obj, err := c.Fake.
		Invokes(testing.NewPatchSubresourceActionWithOptions(listenersetsResource, c.ns, name, pt, data, opts, subresources...), emptyResult)

	if obj == nil {
		return emptyResult, err
	}
	return obj.(*v1alpha1.ListenerSet), err
}

// Apply takes the given apply declarative configuration, applies it and returns the applied listenerSet.
func (c *FakeListenerSets) Apply(ctx context.Context, listenerSet *apisxv1alpha1.ListenerSetApplyConfiguration, opts v1.ApplyOptions) (result *v1alpha1.ListenerSet, err error) {
	if listenerSet == nil {
		return nil, fmt.Errorf("listenerSet provided to Apply must not be nil")
	}
	data, err := json.Marshal(listenerSet)
	if err != nil {
		return nil, err
	}
	name := listenerSet.Name
	if name == nil {
		return nil, fmt.Errorf("listenerSet.Name must be provided to Apply")
	}
	emptyResult := &v1alpha1.ListenerSet{}
	obj, err := c.Fake.
		Invokes(testing.NewPatchSubresourceActionWithOptions(listenersetsResource, c.ns, *name, types.ApplyPatchType, data, opts.ToPatchOptions()), emptyResult)

	if obj == nil {
		return emptyResult, err
	}
	return obj.(*v1alpha1.ListenerSet), err
}

// ApplyStatus was generated because the type contains a Status member.
// Add a +genclient:noStatus comment above the type to avoid generating ApplyStatus().
func (c *FakeListenerSets) ApplyStatus(ctx context.Context, listenerSet *apisxv1alpha1.ListenerSetApplyConfiguration, opts v1.ApplyOptions) (result *v1alpha1.ListenerSet, err error) {
	if listenerSet == nil {
		return nil, fmt.Errorf("listenerSet provided to Apply must not be nil")
	}
	data, err := json.Marshal(listenerSet)
	if err != nil {
		return nil, err
	}
	name := listenerSet.Name
	if name == nil {
		return nil, fmt.Errorf("listenerSet.Name must be provided to Apply")
	}
	emptyResult := &v1alpha1.ListenerSet{}
	obj, err := c.Fake.
		Invokes(testing.NewPatchSubresourceActionWithOptions(listenersetsResource, c.ns, *name, types.ApplyPatchType, data, opts.ToPatchOptions(), "status"), emptyResult)

	if obj == nil {
		return emptyResult, err
	}
	return obj.(*v1alpha1.ListenerSet), err
}
