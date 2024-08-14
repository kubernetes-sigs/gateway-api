/*
Copyright 2023 The Kubernetes Authors.

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

package common

import (
	"fmt"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

type GroupKindFetcher interface {
	Fetch(gk schema.GroupKind) ([]*unstructured.Unstructured, error)
}

var _ GroupKindFetcher = (*defaultGroupKindFetcher)(nil)

type defaultGroupKindFetcher struct {
	factory                 Factory
	additionalResourcesByGK map[schema.GroupKind][]*unstructured.Unstructured
}

type groupKindFetcherOption func(*defaultGroupKindFetcher)

func WithAdditionalResources(resources []*unstructured.Unstructured) groupKindFetcherOption { //nolint:revive
	return func(f *defaultGroupKindFetcher) {
		for _, resource := range resources {
			gk := resource.GetObjectKind().GroupVersionKind().GroupKind()
			f.additionalResourcesByGK[gk] = append(f.additionalResourcesByGK[gk], resource)
		}
	}
}

func NewDefaultGroupKindFetcher(factory Factory, options ...groupKindFetcherOption) *defaultGroupKindFetcher { //nolint:revive
	d := &defaultGroupKindFetcher{
		factory:                 factory,
		additionalResourcesByGK: make(map[schema.GroupKind][]*unstructured.Unstructured),
	}
	for _, option := range options {
		option(d)
	}
	return d
}

func (d defaultGroupKindFetcher) Fetch(gk schema.GroupKind) ([]*unstructured.Unstructured, error) {
	infos, err := d.factory.NewBuilder().
		Unstructured().
		Flatten().
		AllNamespaces(true).
		ResourceTypeOrNameArgs(true, []string{fmt.Sprintf("%v.%v", gk.Kind, gk.Group)}...).
		ContinueOnError().
		Do().
		Infos()
	if err != nil {
		return nil, err
	}

	var result []*unstructured.Unstructured
	for _, info := range infos {
		o, err := runtime.DefaultUnstructuredConverter.ToUnstructured(info.Object)
		if err != nil {
			return nil, err
		}
		result = append(result, &unstructured.Unstructured{Object: o})
	}

	// Return any additional Resources if they have been provided.
	for _, u := range d.additionalResourcesByGK[gk] {
		result = append(result, u)
	}

	return result, nil
}
