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
	"fmt"
	"os"

	"sigs.k8s.io/controller-runtime/pkg/client"

	gatewayv1 "sigs.k8s.io/gateway-api/apis/v1"
	"sigs.k8s.io/gateway-api/gwctl/pkg/common"
	"sigs.k8s.io/gateway-api/gwctl/pkg/policymanager"
	"sigs.k8s.io/gateway-api/gwctl/pkg/relations"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	apimachinerytypes "k8s.io/apimachinery/pkg/types"
	"k8s.io/klog/v2"
)

const (
	// Maximum number of events to be fetched for each resource when constructing
	// the resourceModel.
	maxEventsPerResource = 10
)

var (
	defaultGatewayClassGroupVersion = gatewayv1.GroupVersion
	defaultGatewayGroupVersion      = gatewayv1.GroupVersion
	defaultHTTPRouteGroupVersion    = gatewayv1.GroupVersion
)

// Filter struct defines parameters for filtering resources
type Filter struct {
	Namespace string
	Name      string
	Labels    labels.Selector
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

	// The API versions to be used when fetching Gateway API related resources. An
	// attempt will be made to discover this information from the discovery APIs.
	// Failure to do so will mean we use the "default" versions defined in this
	// file.
	PreferredGatewayClassGroupVersion metav1.GroupVersion
	PreferredGatewayGroupVersion      metav1.GroupVersion
	PreferredHTTPRouteGroupVersion    metav1.GroupVersion
}

func NewDiscoverer(k8sClients *common.K8sClients, policyManager *policymanager.PolicyManager) Discoverer {
	d := &Discoverer{
		K8sClients:                        k8sClients,
		PolicyManager:                     policyManager,
		PreferredGatewayClassGroupVersion: defaultGatewayClassGroupVersion,
		PreferredGatewayGroupVersion:      defaultGatewayGroupVersion,
		PreferredHTTPRouteGroupVersion:    defaultHTTPRouteGroupVersion,
	}

	// Find preferred versions of types.
	if err := d.initPreferredResourceVersions(); err != nil {
		klog.ErrorS(err, "Failed to find preferred version for Gateway API types. Will use the default versions.")
	}
	return *d
}

func (d *Discoverer) initPreferredResourceVersions() error {
	serverPreferredResources, err := d.K8sClients.DiscoveryClient.ServerPreferredResources()
	if err != nil {
		return err
	}
	for _, resourceList := range serverPreferredResources {
		if len(resourceList.APIResources) == 0 {
			continue
		}
		gv, err := schema.ParseGroupVersion(resourceList.GroupVersion)
		if err != nil {
			klog.ErrorS(err, "Failed to parse GroupVersion", "groupVersion", resourceList.GroupVersion)
			continue
		}
		if gv.Group != gatewayv1.GroupVersion.Group {
			continue
		}
		for _, resource := range resourceList.APIResources {
			switch resource.Kind {
			case "GatewayClass":
				d.PreferredGatewayClassGroupVersion.Version = gv.Version
			case "Gateway":
				d.PreferredGatewayGroupVersion.Version = gv.Version
			case "HTTPRoute":
				d.PreferredHTTPRouteGroupVersion.Version = gv.Version
			}
		}
	}
	return nil
}

// DiscoverResourcesForGatewayClass discovers resources related to a
// GatewayClass.
func (d Discoverer) DiscoverResourcesForGatewayClass(filter Filter) (*ResourceModel, error) {
	ctx := context.Background()
	resourceModel := &ResourceModel{}

	gatewayClasses, err := d.fetchGatewayClasses(ctx, filter)
	if err != nil {
		return resourceModel, err
	}
	resourceModel.addGatewayClasses(gatewayClasses...)

	d.discoverPolicies(resourceModel)

	return resourceModel, nil
}

// DiscoverResourcesForGateway discovers resources related to a Gateway.
func (d Discoverer) DiscoverResourcesForGateway(filter Filter) (*ResourceModel, error) {
	ctx := context.Background()
	resourceModel := &ResourceModel{}

	gateways, err := d.fetchGateways(ctx, filter)
	if err != nil {
		return resourceModel, err
	}
	resourceModel.addGateways(gateways...)

	d.discoverEventsForGateways(ctx, resourceModel)

	d.discoverHTTPRoutesFromGateways(ctx, resourceModel)
	d.discoverGatewayClassesFromGateways(ctx, resourceModel)
	d.discoverNamespaces(ctx, resourceModel)
	d.discoverPolicies(resourceModel)

	if err := resourceModel.calculateEffectivePolicies(); err != nil {
		return resourceModel, err
	}

	return resourceModel, nil
}

// DiscoverResourcesForHTTPRoute discovers resources related to an HTTPRoute.
func (d Discoverer) DiscoverResourcesForHTTPRoute(filter Filter) (*ResourceModel, error) {
	ctx := context.Background()
	resourceModel := &ResourceModel{}

	httpRoutes, err := d.fetchHTTPRoutes(ctx, filter)
	if err != nil {
		return resourceModel, err
	}
	resourceModel.addHTTPRoutes(httpRoutes...)

	d.discoverGatewaysFromHTTPRoutes(ctx, resourceModel)
	d.discoverGatewayClassesFromGateways(ctx, resourceModel)
	d.discoverNamespaces(ctx, resourceModel)
	d.discoverPolicies(resourceModel)

	if err := resourceModel.calculateEffectivePolicies(); err != nil {
		return resourceModel, err
	}

	return resourceModel, nil
}

// DiscoverResourcesForBackend discovers resources related to a Backend.
func (d Discoverer) DiscoverResourcesForBackend(filter Filter) (*ResourceModel, error) {
	ctx := context.Background()
	resourceModel := &ResourceModel{}

	backends, err := d.fetchBackends(ctx, filter)
	if err != nil {
		return resourceModel, err
	}
	resourceModel.addBackends(backends...)

	d.discoverHTTPRoutesFromBackends(ctx, resourceModel)
	d.discoverGatewaysFromHTTPRoutes(ctx, resourceModel)
	d.discoverGatewayClassesFromGateways(ctx, resourceModel)
	d.discoverNamespaces(ctx, resourceModel)
	d.discoverPolicies(resourceModel)

	if err := resourceModel.calculateEffectivePolicies(); err != nil {
		return resourceModel, err
	}

	return resourceModel, nil
}

// DiscoverResourcesForNamespace discovers resources related to a Namespace.
func (d Discoverer) DiscoverResourcesForNamespace(filter Filter) (*ResourceModel, error) {
	ctx := context.Background()
	resourceModel := &ResourceModel{}

	namespaces, err := d.fetchNamespace(ctx, filter)
	if err != nil {
		return resourceModel, err
	}

	resourceModel.addNamespace(namespaces...)

	d.discoverPolicies(resourceModel)

	return resourceModel, nil
}

// discoverGatewayClassesFromGateways will add GatewayClasses associated with
// Gateways in the resourceModel.
func (d Discoverer) discoverGatewayClassesFromGateways(ctx context.Context, resourceModel *ResourceModel) {
	gatewayClasses, err := d.fetchGatewayClasses(ctx, Filter{ /* all GatewayClasses */ Labels: labels.Everything()})
	if err != nil {
		klog.V(1).ErrorS(err, "Failed to list all GatewayClasses")
	}

	// Build temporary index for GatewayClasses
	gatewayClassesByID := make(map[gatewayClassID]gatewayv1.GatewayClass)
	for _, gatewayClass := range gatewayClasses {
		gatewayClassesByID[GatewayClassID(gatewayClass.GetName())] = gatewayClass
	}

	for gatewayID, gatewayNode := range resourceModel.Gateways {
		gwcID := GatewayClassID(relations.FindGatewayClassNameForGateway(*gatewayNode.Gateway))
		gatewayClass, ok := gatewayClassesByID[gwcID]
		if !ok {
			klog.V(1).ErrorS(nil, "GatewayClass referenced in Gateway does not exist",
				"gateway", gatewayNode.Gateway.GetNamespace()+"/"+gatewayNode.Gateway.GetName(),
			)
			continue
		}

		resourceModel.addGatewayClasses(gatewayClass)
		resourceModel.connectGatewayWithGatewayClass(gatewayID, gwcID)
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
			gateways, err := d.fetchGateways(ctx, Filter{Namespace: gatewayRef.Namespace, Name: gatewayRef.Name, Labels: labels.Everything()})
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

// discoverHTTPRoutesFromGateways will add HTTPRoutes that are attached to any
// Gateway in the resourceModel.
func (d Discoverer) discoverHTTPRoutesFromGateways(ctx context.Context, resourceModel *ResourceModel) {
	httpRoutes, err := d.fetchHTTPRoutes(ctx, Filter{ /* all HTTPRoutes */ Labels: labels.Everything()})
	if err != nil {
		klog.V(1).ErrorS(err, "Failed to list all HTTPRoutes")
	}

	// Loop through all HTTPRoutes and figure out which are linked to a Gateway
	// that exists in the ResourceModel.
	for _, httpRoute := range httpRoutes {
		klog.V(1).InfoS("Evaluating whether HTTPRoute needs to be included in the resourceModel",
			"httpRoute", httpRoute.GetNamespace()+"/"+httpRoute.GetName(),
		)
		var isHTTPRouteAttachedToValidGateway bool

		for _, gatewayRef := range relations.FindGatewayRefsForHTTPRoute(httpRoute) {
			// Check if Gateway exists in the resourceModel.
			gatewayID := GatewayID(gatewayRef.Namespace, gatewayRef.Name)
			_, ok := resourceModel.Gateways[gatewayID]
			if !ok {
				continue
			}

			// At this point, we know that httpRoute is attached to a Gateway which
			// exists in the resourceModel.
			klog.V(1).InfoS("HTTPRoute included in the resource model because it is attached to a relevant Gateway",
				"httpRoute", httpRoute.GetNamespace()+"/"+httpRoute.GetName(),
				"gateway", gatewayRef.Namespace+"/"+gatewayRef.Name,
			)
			isHTTPRouteAttachedToValidGateway = true

			resourceModel.addHTTPRoutes(httpRoute)
			resourceModel.connectHTTPRouteWithGateway(HTTPRouteID(httpRoute.GetNamespace(), httpRoute.GetName()), gatewayID)
		}

		if !isHTTPRouteAttachedToValidGateway {
			klog.V(1).InfoS("Skipping HTTPRoute since it does not reference any relevant Gateway",
				"httpRoute", httpRoute.GetNamespace()+"/"+httpRoute.GetName(),
			)
		}
	}
}

// discoverHTTPRoutesFromBackends will add HTTPRoutes that reference any Backend
// present in resourceModel.
func (d Discoverer) discoverHTTPRoutesFromBackends(ctx context.Context, resourceModel *ResourceModel) {
	httpRoutes, err := d.fetchHTTPRoutes(ctx, Filter{ /* all HTTPRoutes */ Labels: labels.Everything()})
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
	namespacesList := &corev1.NamespaceList{}
	if err := d.K8sClients.Client.List(ctx, namespacesList, &client.ListOptions{}); err != nil {
		fmt.Fprintf(os.Stderr, "failed to fetch list of namespaces: %v\n", err)
		os.Exit(1)
	}

	namespaceMap := make(map[string]corev1.Namespace)
	for _, namespace := range namespacesList.Items {
		namespaceMap[namespace.Name] = namespace
	}

	for gatewayID, gatewayNode := range resourceModel.Gateways {
		resourceModel.addNamespace(namespaceMap[gatewayNode.Gateway.GetNamespace()])
		resourceModel.connectGatewayWithNamespace(gatewayID, NamespaceID(gatewayNode.Gateway.GetNamespace()))
	}
	for httpRouteID, httpRouteNode := range resourceModel.HTTPRoutes {
		resourceModel.addNamespace(namespaceMap[httpRouteNode.HTTPRoute.GetNamespace()])
		resourceModel.connectHTTPRouteWithNamespace(httpRouteID, NamespaceID(httpRouteNode.HTTPRoute.GetNamespace()))
	}
	for backendID, backendNode := range resourceModel.Backends {
		resourceModel.addNamespace(namespaceMap[backendNode.Backend.GetNamespace()])
		resourceModel.connectBackendWithNamespace(backendID, NamespaceID(backendNode.Backend.GetNamespace()))
	}
}

// discoverPolicies adds Policies for resources that exist in the resourceModel.
func (d Discoverer) discoverPolicies(resourceModel *ResourceModel) {
	resourceModel.addPolicyIfTargetExists(d.PolicyManager.GetPolicies()...)
}

// discoverEventsForGateways adds Events associated with Gateways that exist in
// the resourceModel.
func (d Discoverer) discoverEventsForGateways(ctx context.Context, resourceModel *ResourceModel) {
	for _, gatewayNode := range resourceModel.Gateways {
		eventList := &corev1.EventList{}
		options := &client.ListOptions{
			FieldSelector: fields.AndSelectors(
				fields.OneTermEqualSelector("involvedObject.kind", "Gateway"),
				fields.OneTermEqualSelector("involvedObject.name", gatewayNode.Gateway.Name),
				fields.OneTermEqualSelector("involvedObject.namespace", gatewayNode.Gateway.Namespace),
				fields.OneTermEqualSelector("involvedObject.uid", string(gatewayNode.Gateway.UID)),
			),
			Limit: maxEventsPerResource,
		}
		if err := d.K8sClients.Client.List(ctx, eventList, options); err != nil {
			klog.V(1).ErrorS(err, "Failed to list events associated with Gateway",
				"gateway", gatewayNode.Gateway.Namespace+"/"+gatewayNode.Gateway.Name)
			continue
		}

		gatewayNode.Events = append(gatewayNode.Events, eventList.Items...)
	}
}

// fetchGatewayClasses fetches GatewayClasses based on a filter.
func (d Discoverer) fetchGatewayClasses(ctx context.Context, filter Filter) ([]gatewayv1.GatewayClass, error) {
	gvr := schema.GroupVersionResource{
		Group:    defaultGatewayClassGroupVersion.Group,
		Version:  defaultGatewayClassGroupVersion.Version,
		Resource: "gatewayclasses",
	}
	if d.PreferredGatewayClassGroupVersion != (metav1.GroupVersion{}) {
		gvr.Version = d.PreferredGatewayClassGroupVersion.Version
	}

	if filter.Name != "" {
		// Use Get call.
		gatewayClassUnstructured, err := d.K8sClients.DC.Resource(gvr).Get(ctx, filter.Name, metav1.GetOptions{})
		if err != nil {
			return []gatewayv1.GatewayClass{}, err
		}
		gatewayClass := &gatewayv1.GatewayClass{}
		if err := runtime.DefaultUnstructuredConverter.FromUnstructured(gatewayClassUnstructured.UnstructuredContent(), gatewayClass); err != nil {
			return []gatewayv1.GatewayClass{}, fmt.Errorf("failed to convert unstructured GatewayClass to structured: %v", err)
		}
		return []gatewayv1.GatewayClass{*gatewayClass}, nil
	}

	// Use List call.
	labelSelector := ""
	if filter.Labels != nil {
		labelSelector = filter.Labels.String()
	}
	listOptions := metav1.ListOptions{
		LabelSelector: labelSelector,
	}
	gatewayClassListUnstructured, err := d.K8sClients.DC.Resource(gvr).List(ctx, listOptions)
	if err != nil {
		return []gatewayv1.GatewayClass{}, err
	}
	gatewayClassList := &gatewayv1.GatewayClassList{}
	if err := runtime.DefaultUnstructuredConverter.FromUnstructured(gatewayClassListUnstructured.UnstructuredContent(), gatewayClassList); err != nil {
		return []gatewayv1.GatewayClass{}, fmt.Errorf("failed to convert unstructured GatewayClassList to structured: %v", err)
	}
	return gatewayClassList.Items, nil
}

// fetchGateways fetches Gateways based on a filter.
func (d Discoverer) fetchGateways(ctx context.Context, filter Filter) ([]gatewayv1.Gateway, error) {
	gvr := schema.GroupVersionResource{
		Group:    defaultGatewayGroupVersion.Group,
		Version:  defaultGatewayGroupVersion.Version,
		Resource: "gateways",
	}
	if d.PreferredGatewayGroupVersion != (metav1.GroupVersion{}) {
		gvr.Version = d.PreferredGatewayGroupVersion.Version
	}

	if filter.Name != "" {
		// Use Get call.
		gatewayUnstructured, err := d.K8sClients.DC.Resource(gvr).Namespace(filter.Namespace).Get(ctx, filter.Name, metav1.GetOptions{})
		if err != nil {
			return []gatewayv1.Gateway{}, err
		}
		gateway := &gatewayv1.Gateway{}
		if err := runtime.DefaultUnstructuredConverter.FromUnstructured(gatewayUnstructured.UnstructuredContent(), gateway); err != nil {
			return []gatewayv1.Gateway{}, fmt.Errorf("failed to convert unstructured Gateway to structured: %v", err)
		}
		return []gatewayv1.Gateway{*gateway}, nil
	}

	// Use List call.
	labelSelector := ""
	if filter.Labels != nil {
		labelSelector = filter.Labels.String()
	}
	listOptions := metav1.ListOptions{
		LabelSelector: labelSelector,
	}
	gatewayListUnstructured, err := d.K8sClients.DC.Resource(gvr).Namespace(filter.Namespace).List(ctx, listOptions)
	if err != nil {
		return []gatewayv1.Gateway{}, err
	}
	gatewayList := &gatewayv1.GatewayList{}
	if err := runtime.DefaultUnstructuredConverter.FromUnstructured(gatewayListUnstructured.UnstructuredContent(), gatewayList); err != nil {
		return []gatewayv1.Gateway{}, fmt.Errorf("failed to convert unstructured GatewayList to structured: %v", err)
	}
	return gatewayList.Items, nil
}

// fetchHTTPRoutes fetches HTTPRoutes based on a filter.
func (d Discoverer) fetchHTTPRoutes(ctx context.Context, filter Filter) ([]gatewayv1.HTTPRoute, error) {
	gvr := schema.GroupVersionResource{
		Group:    defaultHTTPRouteGroupVersion.Group,
		Version:  defaultHTTPRouteGroupVersion.Version,
		Resource: "httproutes",
	}
	if d.PreferredHTTPRouteGroupVersion != (metav1.GroupVersion{}) {
		gvr.Version = d.PreferredHTTPRouteGroupVersion.Version
	}

	if filter.Name != "" {
		// Use Get call.
		httpRouteUnstructured, err := d.K8sClients.DC.Resource(gvr).Namespace(filter.Namespace).Get(ctx, filter.Name, metav1.GetOptions{})
		if err != nil {
			return []gatewayv1.HTTPRoute{}, err
		}
		httpRoute := &gatewayv1.HTTPRoute{}
		if err := runtime.DefaultUnstructuredConverter.FromUnstructured(httpRouteUnstructured.UnstructuredContent(), httpRoute); err != nil {
			return []gatewayv1.HTTPRoute{}, fmt.Errorf("failed to convert unstructured HTTPRoute to structured: %v", err)
		}
		return []gatewayv1.HTTPRoute{*httpRoute}, nil
	}

	// Use List call.
	labelSelector := ""
	if filter.Labels != nil {
		labelSelector = filter.Labels.String()
	}
	listOptions := metav1.ListOptions{
		LabelSelector: labelSelector,
	}
	httpRouteListUnstructured, err := d.K8sClients.DC.Resource(gvr).Namespace(filter.Namespace).List(ctx, listOptions)
	if err != nil {
		return []gatewayv1.HTTPRoute{}, err
	}
	httpRouteList := &gatewayv1.HTTPRouteList{}
	if err := runtime.DefaultUnstructuredConverter.FromUnstructured(httpRouteListUnstructured.UnstructuredContent(), httpRouteList); err != nil {
		return []gatewayv1.HTTPRoute{}, fmt.Errorf("failed to convert unstructured HTTPRouteList to structured: %v", err)
	}
	return httpRouteList.Items, nil
}

// fetchBackends fetches Backends based on a filter.
//
// At the moment, this is exclusively used for Backends of type Service, though
// it still returns a slice of unstructured.Unstructured for future extensions.
func (d Discoverer) fetchBackends(ctx context.Context, filter Filter) ([]unstructured.Unstructured, error) {
	gvr := schema.GroupVersionResource{
		Group:    "",
		Version:  "v1",
		Resource: "services",
	}

	if filter.Name != "" {
		// Use Get call.
		backend, err := d.K8sClients.DC.Resource(gvr).Namespace(filter.Namespace).Get(ctx, filter.Name, metav1.GetOptions{})
		if err != nil {
			return []unstructured.Unstructured{}, err
		}

		return []unstructured.Unstructured{*backend}, nil
	}

	// Use List call.
	labelSelector := ""
	if filter.Labels != nil {
		labelSelector = filter.Labels.String()
	}
	listOptions := metav1.ListOptions{
		LabelSelector: labelSelector,
	}
	var backendsList *unstructured.UnstructuredList
	backendsList, err := d.K8sClients.DC.Resource(gvr).Namespace(filter.Namespace).List(ctx, listOptions)
	if err != nil {
		return nil, err
	}

	return backendsList.Items, nil
}

// fetchNamespace fetches Namespaces based on a filter.
func (d Discoverer) fetchNamespace(ctx context.Context, filter Filter) ([]corev1.Namespace, error) {
	if filter.Name != "" {
		// Use Get call.
		namespace := &corev1.Namespace{}
		nn := apimachinerytypes.NamespacedName{Name: filter.Name}
		err := d.K8sClients.Client.Get(ctx, nn, namespace)
		if err != nil {
			return []corev1.Namespace{}, err
		}
		return []corev1.Namespace{*namespace}, nil
	}

	// Use List call.
	options := &client.ListOptions{
		Namespace:     filter.Namespace,
		LabelSelector: filter.Labels,
	}
	namespacesList := &corev1.NamespaceList{}
	if err := d.K8sClients.Client.List(ctx, namespacesList, options); err != nil {
		return []corev1.Namespace{}, err
	}

	return namespacesList.Items, nil
}
