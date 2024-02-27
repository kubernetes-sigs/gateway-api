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

package resourcediscovery

import (
	"context"

	"sigs.k8s.io/controller-runtime/pkg/client"
	gatewayv1 "sigs.k8s.io/gateway-api/apis/v1"
	"sigs.k8s.io/gateway-api/gwctl/pkg/common"
	"sigs.k8s.io/gateway-api/gwctl/pkg/policymanager"
	"sigs.k8s.io/gateway-api/gwctl/pkg/relations"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime/schema"
	apimachinerytypes "k8s.io/apimachinery/pkg/types"
	"k8s.io/klog/v2"
)

// Filter struct defines parameters for filtering resources
type Filter struct {
	Namespace string
	Name      string
	Labels    map[string]string
}

// Discoverer orchestrates the discovery of resources and their associated
// policies, building a model of interconnected resources.
//
// TODO: Optimization Task: Implement a heuristic within each discovery function
// to intelligently choose between:
//   - Single API calls for efficient bulk fetching when appropriate.
//   - Multiple API calls for targeted retrieval when necessary.
type Discoverer struct {
	K8sClients    *common.K8sClients
	PolicyManager *policymanager.PolicyManager
}

// DiscoverResourcesForGatewayClass discovers resources related to a
// GatewayClass.
func (d Discoverer) DiscoverResourcesForGatewayClass(filter Filter) (*ResourceModel, error) {
	ctx := context.Background()
	resourceModel := &ResourceModel{}

	gatewayClasses, err := fetchGatewayClasses(ctx, d.K8sClients, filter)
	if err != nil {
		return resourceModel, err
	}
	resourceModel.addGatewayClasses(gatewayClasses...)

	d.discoverPolicies(ctx, resourceModel)

	return resourceModel, nil
}

// DiscoverResourcesForGateway discovers resources related to a Gateway.
func (d Discoverer) DiscoverResourcesForGateway(filter Filter) (*ResourceModel, error) {
	ctx := context.Background()
	resourceModel := &ResourceModel{}

	gateways, err := fetchGateways(ctx, d.K8sClients, filter)
	if err != nil {
		return resourceModel, err
	}
	resourceModel.addGateways(gateways...)

	d.discoverGatewayClassesFromGateways(ctx, resourceModel)
	d.discoverNamespaces(ctx, resourceModel)
	d.discoverPolicies(ctx, resourceModel)

	resourceModel.calculateEffectivePolicies()

	return resourceModel, nil
}

// DiscoverResourcesForHTTPRoute discovers resources related to an HTTPRoute.
func (d Discoverer) DiscoverResourcesForHTTPRoute(filter Filter) (*ResourceModel, error) {
	ctx := context.Background()
	resourceModel := &ResourceModel{}

	httpRoutes, err := fetchHTTPRoutes(ctx, d.K8sClients, filter)
	if err != nil {
		return resourceModel, err
	}
	resourceModel.addHTTPRoutes(httpRoutes...)

	d.discoverGatewaysFromHTTPRoutes(ctx, resourceModel)
	d.discoverGatewayClassesFromGateways(ctx, resourceModel)
	d.discoverNamespaces(ctx, resourceModel)
	d.discoverPolicies(ctx, resourceModel)

	resourceModel.calculateEffectivePolicies()

	return resourceModel, nil
}

// DiscoverResourcesForBackend discovers resources related to a Backend.
func (d Discoverer) DiscoverResourcesForBackend(filter Filter) (*ResourceModel, error) {
	ctx := context.Background()
	resourceModel := &ResourceModel{}

	backends, err := fetchBackends(ctx, d.K8sClients, filter)
	if err != nil {
		return resourceModel, err
	}
	resourceModel.addBackends(backends...)

	d.discoverHTTPRoutesFromBackends(ctx, resourceModel)
	d.discoverGatewaysFromHTTPRoutes(ctx, resourceModel)
	d.discoverGatewayClassesFromGateways(ctx, resourceModel)
	d.discoverNamespaces(ctx, resourceModel)
	d.discoverPolicies(ctx, resourceModel)

	resourceModel.calculateEffectivePolicies()

	return resourceModel, nil
}

// discoverGatewayClassesFromGateways will add GatewayClasses associated with
// Gateways in the resourceModel.
func (d Discoverer) discoverGatewayClassesFromGateways(ctx context.Context, resourceModel *ResourceModel) {
	gatewayClasses, err := fetchGatewayClasses(ctx, d.K8sClients, Filter{ /* all GatewayClasses */ })
	if err != nil {
		klog.V(1).ErrorS(err, "Failed to list all GatewayClasses")
	}

	// Build temporary index for GatewayClasses
	gatewayClassesByID := make(map[gatewayClassID]gatewayv1.GatewayClass)
	for _, gatewayClass := range gatewayClasses {
		gatewayClassesByID[GatewayClassID(gatewayClass.GetName())] = gatewayClass
	}

	for gatewayID, gatewayNode := range resourceModel.Gateways {
		gatewayClassID := GatewayClassID(relations.FindGatewayClassNameForGateway(*gatewayNode.Gateway))
		gatewayClass, ok := gatewayClassesByID[gatewayClassID]
		if !ok {
			klog.V(1).ErrorS(nil, "GatewayClass referenced in Gateway does not exist",
				"gateway", gatewayNode.Gateway.GetNamespace()+"/"+gatewayNode.Gateway.GetName(),
			)
			continue
		}

		resourceModel.addGatewayClasses(gatewayClass)
		resourceModel.connectGatewayWithGatewayClass(gatewayID, gatewayClassID)
	}
}

// discoverGatewaysFromHTTPRoutes will add Gateways associated with HTTPRoutes
// in the resourceModel.
func (d Discoverer) discoverGatewaysFromHTTPRoutes(ctx context.Context, resourceModel *ResourceModel) {
	// Visit all gateways corresponding to the httpRoutes
	for _, httpRouteNode := range resourceModel.HTTPRoutes {
		for _, gatewayRef := range relations.FindGatewayRefsForHTTPRoute(*httpRouteNode.HTTPRoute) {
			// Check if Gateway already exists in the resourceModel.
			if _, ok := resourceModel.Gateways[GatewayID(gatewayRef.Namespace, gatewayRef.Name)]; ok {
				// Gateway already exists in the resourceModel, skip re-fetching.
				continue
			}

			// Gateway doesn't already exist so fetch and add it to the resourceModel.
			gateways, err := fetchGateways(ctx, d.K8sClients, Filter{Namespace: gatewayRef.Namespace, Name: gatewayRef.Name})
			if err != nil {
				klog.V(1).ErrorS(err, "Gateway referenced by HTTPRoute not found",
					"gateway", gatewayRef.String(),
					"httproute", httpRouteNode.HTTPRoute.GetNamespace()+"/"+httpRouteNode.HTTPRoute.GetName(),
				)
				continue
			}
			resourceModel.addGateways(gateways[0])
		}
	}

	// Connect gatewayd with httproutes.
	for httpRouteID, httpRouteNode := range resourceModel.HTTPRoutes {
		for _, gatewayRef := range relations.FindGatewayRefsForHTTPRoute(*httpRouteNode.HTTPRoute) {
			resourceModel.connectHTTPRouteWithGateway(httpRouteID, GatewayID(gatewayRef.Namespace, gatewayRef.Name))
		}
	}
}

// discoverHTTPRoutesFromBackends will add HTTPRoutes that reference any Backend
// present in resourceModel.
func (d Discoverer) discoverHTTPRoutesFromBackends(ctx context.Context, resourceModel *ResourceModel) {
	httpRoutes, err := fetchHTTPRoutes(ctx, d.K8sClients, Filter{ /* all HTTPRoutes */ })
	if err != nil {
		klog.V(1).ErrorS(err, "Failed to list all HTTPRoutes")
	}

	for _, httpRoute := range httpRoutes {
		var found bool
		for _, backendRef := range relations.FindBackendRefsForHTTPRoute(httpRoute) {
			backendID := BackendID(backendRef.Group, backendRef.Kind, backendRef.Namespace, backendRef.Name)
			_, ok := resourceModel.Backends[backendID]
			if !ok {
				continue
			}
			found = true

			resourceModel.addHTTPRoutes(httpRoute)
			resourceModel.connectHTTPRouteWithBackend(HTTPRouteID(httpRoute.GetNamespace(), httpRoute.GetName()), backendID)
		}
		if !found {
			klog.V(1).InfoS("Skipping HTTPRoute since it does not reference any required Backend",
				"httpRoute", httpRoute.GetNamespace()+"/"+httpRoute.GetName(),
			)
		}
	}
}

// discoverNamespaces adds Namespaces for resources that exist in the
// resourceModel.
func (d Discoverer) discoverNamespaces(ctx context.Context, resourceModel *ResourceModel) {
	for gatewayID, gatewayNode := range resourceModel.Gateways {
		resourceModel.addNamespace(gatewayNode.Gateway.GetNamespace())
		resourceModel.connectGatewayWithNamespace(gatewayID, NamespaceID(gatewayNode.Gateway.GetNamespace()))
	}
	for httpRouteID, httpRouteNode := range resourceModel.HTTPRoutes {
		resourceModel.addNamespace(httpRouteNode.HTTPRoute.GetNamespace())
		resourceModel.connectHTTPRouteWithNamespace(httpRouteID, NamespaceID(httpRouteNode.HTTPRoute.GetNamespace()))
	}
	for backendID, backendNode := range resourceModel.Backends {
		resourceModel.addNamespace(backendNode.Backend.GetNamespace())
		resourceModel.connectBackendWithNamespace(backendID, NamespaceID(backendNode.Backend.GetNamespace()))
	}
}

// discoverPolicies adds Policies for resources that exist in the resourceModel.
func (d Discoverer) discoverPolicies(ctx context.Context, resourceModel *ResourceModel) {
	resourceModel.addPolicyIfTargetExists(d.PolicyManager.GetPolicies()...)
}

// fetchGatewayClasses fetches GatewayClasses based on a filter.
func fetchGatewayClasses(ctx context.Context, k8sClients *common.K8sClients, filter Filter) ([]gatewayv1.GatewayClass, error) {
	if filter.Name != "" {
		// Use Get call.
		gatewayClass := &gatewayv1.GatewayClass{}
		nn := apimachinerytypes.NamespacedName{Name: filter.Name}
		if err := k8sClients.Client.Get(ctx, nn, gatewayClass); err != nil {
			return []gatewayv1.GatewayClass{}, err
		}

		return []gatewayv1.GatewayClass{*gatewayClass}, nil
	}

	// Use List call.
	options := &client.ListOptions{
		Namespace:     filter.Namespace,
		LabelSelector: labels.SelectorFromSet(filter.Labels),
	}
	gatewayClassList := &gatewayv1.GatewayClassList{}
	if err := k8sClients.Client.List(ctx, gatewayClassList, options); err != nil {
		return []gatewayv1.GatewayClass{}, err
	}

	return gatewayClassList.Items, nil
}

// fetchGateways fetches Gateways based on a filter.
func fetchGateways(ctx context.Context, k8sClients *common.K8sClients, filter Filter) ([]gatewayv1.Gateway, error) {
	if filter.Name != "" {
		// Use Get call.
		gateway := &gatewayv1.Gateway{}
		nn := apimachinerytypes.NamespacedName{Namespace: filter.Namespace, Name: filter.Name}
		if err := k8sClients.Client.Get(ctx, nn, gateway); err != nil {
			return []gatewayv1.Gateway{}, err
		}

		return []gatewayv1.Gateway{*gateway}, nil
	}

	// Use List call.
	options := &client.ListOptions{
		Namespace:     filter.Namespace,
		LabelSelector: labels.SelectorFromSet(filter.Labels),
	}
	gatewayList := &gatewayv1.GatewayList{}
	if err := k8sClients.Client.List(ctx, gatewayList, options); err != nil {
		return []gatewayv1.Gateway{}, err
	}

	return gatewayList.Items, nil
}

// fetchHTTPRoutes fetches HTTPRoutes based on a filter.
func fetchHTTPRoutes(ctx context.Context, k8sClients *common.K8sClients, filter Filter) ([]gatewayv1.HTTPRoute, error) {
	if filter.Name != "" {
		// Use Get call.
		httpRoute := &gatewayv1.HTTPRoute{}
		nn := apimachinerytypes.NamespacedName{Namespace: filter.Namespace, Name: filter.Name}
		if err := k8sClients.Client.Get(ctx, nn, httpRoute); err != nil {
			return []gatewayv1.HTTPRoute{}, err
		}

		return []gatewayv1.HTTPRoute{*httpRoute}, nil
	}

	// Use List call.
	options := &client.ListOptions{
		Namespace:     filter.Namespace,
		LabelSelector: labels.SelectorFromSet(filter.Labels),
	}
	httpRouteList := &gatewayv1.HTTPRouteList{}
	if err := k8sClients.Client.List(ctx, httpRouteList, options); err != nil {
		return []gatewayv1.HTTPRoute{}, err
	}

	return httpRouteList.Items, nil
}

// fetchBackends fetches Backends based on a filter.
//
// At the moment, this is exclusively used for Backends of type Service, though
// it still returns a slice of unstructured.Unstructured for future extensions.
func fetchBackends(ctx context.Context, k8sClients *common.K8sClients, filter Filter) ([]unstructured.Unstructured, error) {
	gvr := schema.GroupVersionResource{
		Group:    "",
		Version:  "v1",
		Resource: "services",
	}

	if filter.Name != "" {
		// Use Get call.
		backend, err := k8sClients.DC.Resource(gvr).Namespace(filter.Namespace).Get(ctx, filter.Name, metav1.GetOptions{})
		if err != nil {
			return []unstructured.Unstructured{}, err
		}

		return []unstructured.Unstructured{*backend}, nil
	}

	// Use List call.
	listOptions := metav1.ListOptions{
		LabelSelector: labels.SelectorFromSet(filter.Labels).String(),
	}
	var backendsList *unstructured.UnstructuredList
	backendsList, err := k8sClients.DC.Resource(gvr).Namespace(filter.Namespace).List(ctx, listOptions)
	if err != nil {
		return nil, err
	}

	return backendsList.Items, nil
}
