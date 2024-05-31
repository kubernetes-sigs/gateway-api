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
	"testing"

	corev1 "k8s.io/api/core/v1"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/dynamic"
	fakedynamicclient "k8s.io/client-go/dynamic/fake"
	fakeclientset "k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/clientcmd"
	"sigs.k8s.io/controller-runtime/pkg/client"
	fakeclient "sigs.k8s.io/controller-runtime/pkg/client/fake"

	gatewayv1 "sigs.k8s.io/gateway-api/apis/v1"
	gatewayv1alpha2 "sigs.k8s.io/gateway-api/apis/v1alpha2"
	gatewayv1alpha3 "sigs.k8s.io/gateway-api/apis/v1alpha3"
	gatewayv1beta1 "sigs.k8s.io/gateway-api/apis/v1beta1"
)

type K8sClients struct {
	Client          client.Client
	DC              dynamic.Interface
	DiscoveryClient discovery.DiscoveryInterface
}

func NewK8sClients(kubeconfig string) (*K8sClients, error) {
	restConfig, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		return nil, fmt.Errorf("failed to get restConfig from BuildConfigFromFlags: %v", err)
	}

	client, err := client.New(restConfig, client.Options{})
	if err != nil {
		return nil, fmt.Errorf("failed to initialize Kubernetes client: %v", err)
	}
	if err := gatewayv1alpha3.Install(client.Scheme()); err != nil {
		return nil, err
	}
	if err := gatewayv1alpha2.Install(client.Scheme()); err != nil {
		return nil, err
	}
	if err := gatewayv1beta1.Install(client.Scheme()); err != nil {
		return nil, err
	}
	if err := gatewayv1.Install(client.Scheme()); err != nil {
		return nil, err
	}

	dc := dynamic.NewForConfigOrDie(restConfig)

	return &K8sClients{
		Client:          client,
		DC:              dc,
		DiscoveryClient: discovery.NewDiscoveryClientForConfigOrDie(restConfig),
	}, nil
}

func MustClientsForTest(t *testing.T, initRuntimeObjects ...runtime.Object) *K8sClients {
	scheme := scheme.Scheme
	if err := gatewayv1alpha3.Install(scheme); err != nil {
		t.Fatal(err)
	}
	if err := gatewayv1alpha2.Install(scheme); err != nil {
		t.Fatal(err)
	}
	if err := gatewayv1beta1.Install(scheme); err != nil {
		t.Fatal(err)
	}
	if err := gatewayv1.Install(scheme); err != nil {
		t.Fatal(err)
	}
	if err := apiextensionsv1.AddToScheme(scheme); err != nil {
		t.Fatal(err)
	}

	// These extractorFuncs are used to properly mock the kubernetes client
	// dependency in unit tests. They enable the ability to be able to list Events
	// associated with a specific resource.
	eventKindExtractorFunc := func(o client.Object) []string {
		return []string{o.(*corev1.Event).InvolvedObject.Kind}
	}
	eventNameExtractorFunc := func(o client.Object) []string {
		return []string{o.(*corev1.Event).InvolvedObject.Name}
	}
	eventNamespaceExtractorFunc := func(o client.Object) []string {
		return []string{o.(*corev1.Event).InvolvedObject.Namespace}
	}
	eventUIDExtractorFunc := func(o client.Object) []string {
		return []string{string(o.(*corev1.Event).InvolvedObject.UID)}
	}

	fakeClient := fakeclient.NewClientBuilder().
		WithScheme(scheme).
		WithRuntimeObjects(initRuntimeObjects...).
		WithIndex(&corev1.Event{}, "involvedObject.kind", eventKindExtractorFunc).
		WithIndex(&corev1.Event{}, "involvedObject.name", eventNameExtractorFunc).
		WithIndex(&corev1.Event{}, "involvedObject.namespace", eventNamespaceExtractorFunc).
		WithIndex(&corev1.Event{}, "involvedObject.uid", eventUIDExtractorFunc).
		Build()
	fakeDiscoveryClient := fakeclientset.NewSimpleClientset().Discovery()

	// Setup a fake DynamicClient, which requires some special handling.
	//
	// When objects are injected using `NewSimpleDynamicClient` or
	// `NewSimpleDynamicClientWithCustomListKinds` to the fake resource tracker,
	// internally it will use the `meta.UnsafeGuessKindToResource()` function.
	// This incorrectly guesses the GVR of Gateway with Resource being guessed as
	// "gatewaies" (instead of "gateways"). As a workaround, we will have to
	// create Gateway objects separately.
	//
	// Also, because of this incorrect guessing, we need to register the
	// GatewayList type separately.
	gatewayv1GVR := schema.GroupVersionResource{
		Group:    gatewayv1.GroupVersion.Group,
		Version:  gatewayv1.GroupVersion.Version,
		Resource: "gateways",
	}
	gvrToListKind := map[schema.GroupVersionResource]string{
		gatewayv1GVR: "GatewayList",
	}
	for _, obj := range initRuntimeObjects {
		if crd, ok := obj.(*apiextensionsv1.CustomResourceDefinition); ok {
			gvr := schema.GroupVersionResource{
				Group:    crd.Spec.Group,
				Version:  crd.Spec.Versions[0].Name,
				Resource: crd.Spec.Names.Plural, // CRD Kinds directly map to the Resource.
			}
			gvrToListKind[gvr] = crd.Spec.Names.Kind + "List"
		}
	}
	fakeDC := fakedynamicclient.NewSimpleDynamicClientWithCustomListKinds(scheme, gvrToListKind)
	for _, obj := range initRuntimeObjects {
		var err error
		if gateway, ok := obj.(*gatewayv1.Gateway); ok {
			// Register Gateway with correct GVR. This needs to be done explicitly for
			// Gateway since the automatically guess GVR is incorrect.
			//
			// Automatic guessing of GVR uses `meta.UnsafeGuessKindToResource()` which
			// pluralizes "gateway" to "gatewaies" (since the singular ends in a 'y')
			err = fakeDC.Tracker().Create(gatewayv1GVR, gateway, gateway.Namespace)
		} else {
			// Register non-Gateway resources automatically without GVR.
			err = fakeDC.Tracker().Add(obj)
		}
		if err != nil {
			t.Fatalf("Failed to add object to fake DynamicClient: %v", err)
		}
	}

	return &K8sClients{
		Client:          fakeClient,
		DC:              fakeDC,
		DiscoveryClient: fakeDiscoveryClient,
	}
}

func PtrTo[T any](a T) *T {
	return &a
}
