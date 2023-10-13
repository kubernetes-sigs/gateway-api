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

package resourcehelpers

import (
	"context"
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/discovery"
	"k8s.io/utils/strings/slices"
	"sigs.k8s.io/gateway-api/gwctl/pkg/common"
)

func ListBackends(ctx context.Context, k8sClients *common.K8sClients, resourceType, namespace string) ([]unstructured.Unstructured, error) {
	return listOrGetBackends(ctx, k8sClients, resourceType, namespace, "")
}

func GetBackend(ctx context.Context, k8sClients *common.K8sClients, resourceType, namespace, name string) (unstructured.Unstructured, error) {
	backendsList, err := listOrGetBackends(ctx, k8sClients, resourceType, namespace, name)
	if err != nil {
		return unstructured.Unstructured{}, err
	}
	if len(backendsList) == 0 {
		return unstructured.Unstructured{}, nil
	}
	return backendsList[0], nil
}

func listOrGetBackends(ctx context.Context, k8sClients *common.K8sClients, resourceType, namespace, name string) ([]unstructured.Unstructured, error) {
	apiResource, err := apiResourceFromResourceType(resourceType, k8sClients.DiscoveryClient)
	if err != nil {
		return nil, err
	}
	gvr := schema.GroupVersionResource{
		Group:    apiResource.Group,
		Version:  apiResource.Version,
		Resource: apiResource.Name,
	}

	listOptions := metav1.ListOptions{}
	if name != "" {
		listOptions.FieldSelector = fields.OneTermEqualSelector("metadata.name", name).String()
	}

	var backendsList *unstructured.UnstructuredList
	if apiResource.Namespaced {
		backendsList, err = k8sClients.DC.Resource(gvr).Namespace(namespace).List(ctx, listOptions)
	} else {
		backendsList, err = k8sClients.DC.Resource(gvr).List(ctx, listOptions)
	}
	if err != nil {
		return nil, err
	}

	return backendsList.Items, nil
}

func apiResourceFromResourceType(resourceType string, discoveryClient discovery.DiscoveryInterface) (metav1.APIResource, error) {
	resourceGroups, err := discoveryClient.ServerPreferredResources()
	if err != nil {
		return metav1.APIResource{}, err
	}
	for _, resourceGroup := range resourceGroups {
		gv, err := schema.ParseGroupVersion(resourceGroup.GroupVersion)
		if err != nil {
			return metav1.APIResource{}, err
		}
		for _, resource := range resourceGroup.APIResources {
			var choices []string
			choices = append(choices, resource.Kind)
			choices = append(choices, resource.Name)
			choices = append(choices, resource.ShortNames...)
			choices = append(choices, resource.SingularName)
			if slices.Contains(choices, resourceType) {
				resource.Version = gv.Version
				return resource, nil
			}
		}
	}
	return metav1.APIResource{}, fmt.Errorf("GVR for %v not found in discovery", resourceType)
}
