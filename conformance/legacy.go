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

package conformance

import (
	"testing"

	v1 "sigs.k8s.io/gateway-api/apis/v1"
	"sigs.k8s.io/gateway-api/apis/v1alpha2"
	"sigs.k8s.io/gateway-api/apis/v1beta1"
	"sigs.k8s.io/gateway-api/conformance/tests"
	"sigs.k8s.io/gateway-api/conformance/utils/flags"
	"sigs.k8s.io/gateway-api/conformance/utils/suite"

	"github.com/stretchr/testify/require"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	clientset "k8s.io/client-go/kubernetes"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
)

func DefaultLegacyOptions(t *testing.T) *suite.Options {
	cfg, err := config.GetConfig()
	require.NoError(t, err, "Error loading Kubernetes config")
	client, err := client.New(cfg, client.Options{})
	require.NoError(t, err, "Error initializing Kubernetes client")
	clientset, err := clientset.NewForConfig(cfg)
	require.NoError(t, err, "Error initializing Kubernetes clientset")

	require.NoError(t, v1alpha2.AddToScheme(client.Scheme()))
	require.NoError(t, v1beta1.AddToScheme(client.Scheme()))
	require.NoError(t, v1.AddToScheme(client.Scheme()))
	require.NoError(t, apiextensionsv1.AddToScheme(client.Scheme()))

	supportedFeatures := suite.ParseSupportedFeatures(*flags.SupportedFeatures)
	exemptFeatures := suite.ParseSupportedFeatures(*flags.ExemptFeatures)
	skipTests := suite.ParseSkipTests(*flags.SkipTests)
	namespaceLabels := suite.ParseKeyValuePairs(*flags.NamespaceLabels)
	namespaceAnnotations := suite.ParseKeyValuePairs(*flags.NamespaceAnnotations)
	return &suite.Options{
		Client:     client,
		RestConfig: cfg,
		FS:         &Manifests,
		// This clientset is needed in addition to the client only because
		// controller-runtime client doesn't support non CRUD sub-resources yet (https://github.com/kubernetes-sigs/controller-runtime/issues/452).
		Clientset:                  clientset,
		GatewayClassName:           *flags.GatewayClassName,
		Debug:                      *flags.ShowDebug,
		CleanupBaseResources:       *flags.CleanupBaseResources,
		SupportedFeatures:          supportedFeatures,
		ExemptFeatures:             exemptFeatures,
		EnableAllSupportedFeatures: *flags.EnableAllSupportedFeatures,
		NamespaceLabels:            namespaceLabels,
		NamespaceAnnotations:       namespaceAnnotations,
		SkipTests:                  skipTests,
		RunTest:                    *flags.RunTest,
	}
}

func RunLegacyConformance(t *testing.T, opts *suite.Options) {
	if opts == nil {
		opts = DefaultLegacyOptions(t)
	}

	t.Log("Running conformance tests with:")
	logOptions(t, opts)

	cSuite := suite.New(*opts)
	cSuite.Setup(t)
	cSuite.Run(t, tests.ConformanceTests)
}

func logOptions(t *testing.T, opts *suite.Options) {
	t.Logf("  GatewayClass: %s", opts.GatewayClassName)
	t.Logf("  Cleanup Resources: %t", opts.CleanupBaseResources)
	t.Logf("  Debug: %t", opts.Debug)
	t.Logf("  Enable All Features: %t", opts.EnableAllSupportedFeatures)
	t.Logf("  Supported Features: %v", opts.SupportedFeatures.UnsortedList())
	t.Logf("  ExemptFeatures: %v", opts.ExemptFeatures.UnsortedList())
}
