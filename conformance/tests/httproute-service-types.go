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

package tests

import (
	"context"
	"slices"
	"testing"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	discoveryv1 "k8s.io/api/discovery/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/utils/ptr"
	client "sigs.k8s.io/controller-runtime/pkg/client"

	"sigs.k8s.io/gateway-api/conformance/utils/http"
	"sigs.k8s.io/gateway-api/conformance/utils/kubernetes"
	"sigs.k8s.io/gateway-api/conformance/utils/suite"
)

func init() {
	ConformanceTests = append(ConformanceTests, HTTPRouteServiceTypes)
}

var HTTPRouteServiceTypes = suite.ConformanceTest{
	ShortName:   "HTTPRouteServiceTypes",
	Description: "A single HTTPRoute should be able to route traffic to various service type backends",
	Features: []suite.SupportedFeature{
		suite.SupportGateway,
		suite.SupportHTTPRoute,
	},
	Manifests: []string{"tests/httproute-service-types.yaml"},
	Test: func(t *testing.T, suite *suite.ConformanceTestSuite) {
		var (
			typeManualEndpoints = []string{
				"manual-endpoints",
				"headless-manual-endpoints",
			}

			typeManualEndpointSlices = []string{
				"manual-endpointslices",
				"headless-manual-endpointslices",
			}

			typeManaged = []string{
				"headless",
			}

			serviceTypes = slices.Concat(typeManaged, typeManualEndpoints, typeManualEndpointSlices)

			ctx     = context.TODO()
			ns      = "gateway-conformance-infra"
			routeNN = types.NamespacedName{Name: "service-types", Namespace: ns}
			gwNN    = types.NamespacedName{Name: "same-namespace", Namespace: ns}
		)

		gwAddr := kubernetes.GatewayAndHTTPRoutesMustBeAccepted(t, suite.Client, suite.TimeoutConfig, suite.ControllerName, kubernetes.NewGatewayRef(gwNN), routeNN)
		kubernetes.HTTPRouteMustHaveResolvedRefsConditionsTrue(t, suite.Client, suite.TimeoutConfig, routeNN, gwNN)

		deployment := &appsv1.Deployment{}
		if err := suite.Client.Get(ctx, client.ObjectKey{Namespace: ns, Name: "infra-backend-v1"}, deployment); err != nil {
			t.Fatal("Failed to list Deployment 'infra-backend-v1':", err)
		}

		selector, err := metav1.LabelSelectorAsSelector(deployment.Spec.Selector)
		if err != nil {
			t.Fatal("Failed to parse Deployment selector", err)
		}

		// Setup Manual Endpoints
		pods := &corev1.PodList{}
		if err := suite.Client.List(ctx, pods, client.MatchingLabelsSelector{Selector: selector}, client.InNamespace(ns)); err != nil {
			t.Fatal("Failed to list infra-backend-v1 Pods:", err)
		}

		if len(pods.Items) == 0 {
			t.Fatal("Expected infra-backend-v1 to have running Pods")
		}

		setupEndpoints(t, suite.Client, typeManualEndpoints, ns, pods)
		setupEndpointSlices(t, suite.Client, typeManualEndpointSlices, ns, pods)

		for i, path := range serviceTypes {
			expected := http.ExpectedResponse{
				Request:   http.Request{Path: "/" + path},
				Response:  http.Response{StatusCode: 200},
				Backend:   "infra-backend-v1",
				Namespace: "gateway-conformance-infra",
			}

			t.Run(expected.GetTestCaseName(i), func(t *testing.T) {
				t.Parallel()
				http.MakeRequestAndExpectEventuallyConsistentResponse(t, suite.RoundTripper, suite.TimeoutConfig, gwAddr, expected)
			})
		}
	},
}

func setupEndpoints(t *testing.T, klient client.Client, endpointNames []string, ns string, pods *corev1.PodList) {
	for _, endpointName := range endpointNames {
		endpoints := &corev1.Endpoints{}
		if err := klient.Get(context.TODO(), client.ObjectKey{Name: endpointName, Namespace: ns}, endpoints); err != nil {
			t.Fatalf("Unable to fetch Endpoint %q: %v", endpointName, err)
		}

		patch := client.MergeFrom(endpoints.DeepCopy())

		endpoints.Subsets = []corev1.EndpointSubset{{
			Addresses: make([]corev1.EndpointAddress, len(pods.Items)),
			Ports: []corev1.EndpointPort{{
				Name:     "first-port",
				Protocol: corev1.ProtocolTCP,
				Port:     3000,
			}},
		}}

		for i, pod := range pods.Items {
			endpoints.Subsets[0].Addresses[i] = corev1.EndpointAddress{
				IP:       pod.Status.PodIP,
				NodeName: ptr.To(pod.Spec.NodeName),
				TargetRef: &corev1.ObjectReference{
					Kind:      "Pod",
					Name:      pod.GetName(),
					Namespace: pod.GetNamespace(),
					UID:       pod.GetUID(),
				},
			}
		}
		if err := klient.Patch(context.TODO(), endpoints, patch); err != nil {
			t.Fatalf("Failed to patch Endpoint %q: %v", endpointName, err)
		}
	}
}

func setupEndpointSlices(t *testing.T, klient client.Client, endpointNames []string, ns string, pods *corev1.PodList) {
	for _, endpointName := range endpointNames {
		endpointSlice := &discoveryv1.EndpointSlice{}
		if err := klient.Get(context.TODO(), client.ObjectKey{Name: endpointName, Namespace: ns}, endpointSlice); err != nil {
			t.Fatalf("Unable to fetch EndpointSlice %q: %v", endpointName, err)
		}

		patch := client.MergeFrom(endpointSlice.DeepCopy())

		endpointSlice.Endpoints = make([]discoveryv1.Endpoint, len(pods.Items))

		for i, pod := range pods.Items {
			endpointSlice.Endpoints[i] = discoveryv1.Endpoint{
				Addresses: []string{pod.Status.PodIP},
				Conditions: discoveryv1.EndpointConditions{
					Ready:       ptr.To(true),
					Serving:     ptr.To(true),
					Terminating: ptr.To(false),
				},
				NodeName: ptr.To(pod.Spec.NodeName),
				TargetRef: &corev1.ObjectReference{
					Kind:      "Pod",
					Name:      pod.GetName(),
					Namespace: pod.GetNamespace(),
					UID:       pod.GetUID(),
				},
			}
		}
		if err := klient.Patch(context.TODO(), endpointSlice, patch); err != nil {
			t.Fatalf("Failed to patch EndpointSlice %q: %v", endpointName, err)
		}
	}
}
