/*
Copyright 2022 The Kubernetes Authors.

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

// conformance_test contains code to run the conformance tests. This is in its own package to avoid circular imports.
package conformance_test

import (
	"strings"
	"testing"

	"sigs.k8s.io/gateway-api/apis/v1alpha2"
	"sigs.k8s.io/gateway-api/apis/v1beta1"
	"sigs.k8s.io/gateway-api/conformance/tests"
	"sigs.k8s.io/gateway-api/conformance/utils/flags"
	"sigs.k8s.io/gateway-api/conformance/utils/suite"

	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/client-go/kubernetes"
	_ "k8s.io/client-go/plugin/pkg/client/auth"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
)

func TestConformance(t *testing.T) {
	cfg, err := config.GetConfig()
	if err != nil {
		t.Fatalf("Error loading Kubernetes config: %v", err)
	}
	client, err := client.New(cfg, client.Options{})
	if err != nil {
		t.Fatalf("Error initializing Kubernetes client: %v", err)
	}
	clientset, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		t.Fatalf("Error initializing Kubernetes REST client: %v", err)
	}

	v1alpha2.AddToScheme(client.Scheme())
	v1beta1.AddToScheme(client.Scheme())

	supportedFeatures := parseSupportedFeatures(*flags.SupportedFeatures)
	exemptFeatures := parseSupportedFeatures(*flags.ExemptFeatures)

	t.Logf("Running conformance tests with %s GatewayClass\n cleanup: %t\n debug: %t\n enable all features: %t \n supported features: [%v]\n exempt features: [%v]",
		*flags.GatewayClassName, *flags.CleanupBaseResources, *flags.ShowDebug, *flags.EnableAllSupportedFeatures, *flags.SupportedFeatures, *flags.ExemptFeatures)

	cSuite := suite.New(suite.Options{
		Client:                     client,
		RESTClient:                 clientset.CoreV1().RESTClient().(*rest.RESTClient),
		RestConfig:                 cfg,
		GatewayClassName:           *flags.GatewayClassName,
		Debug:                      *flags.ShowDebug,
		CleanupBaseResources:       *flags.CleanupBaseResources,
		SupportedFeatures:          supportedFeatures,
		ExemptFeatures:             exemptFeatures,
		EnableAllSupportedFeatures: *flags.EnableAllSupportedFeatures,
	})
	cSuite.Setup(t)

	cSuite.Run(t, tests.ConformanceTests)
}

// parseSupportedFeatures parses flag arguments and converts the string to
// sets.Set[suite.SupportedFeature]
func parseSupportedFeatures(f string) sets.Set[suite.SupportedFeature] {
	if f == "" {
		return nil
	}
	res := sets.Set[suite.SupportedFeature]{}
	for _, value := range strings.Split(f, ",") {
		res.Insert(suite.SupportedFeature(value))
	}
	return res
}
