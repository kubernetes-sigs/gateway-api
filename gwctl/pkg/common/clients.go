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

	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apimachinery/pkg/runtime"
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

	fakeClient := fakeclient.NewClientBuilder().WithScheme(scheme).WithRuntimeObjects(initRuntimeObjects...).Build()
	fakeDC := fakedynamicclient.NewSimpleDynamicClient(scheme, initRuntimeObjects...)
	fakeDiscoveryClient := fakeclientset.NewSimpleClientset().Discovery()

	return &K8sClients{
		Client:          fakeClient,
		DC:              fakeDC,
		DiscoveryClient: fakeDiscoveryClient,
	}
}

func PtrTo[T any](a T) *T {
	return &a
}
