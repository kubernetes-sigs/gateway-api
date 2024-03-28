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

package conformance_test

import (
	"os"
	"testing"

	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
	"sigs.k8s.io/yaml"

	gatewayv1 "sigs.k8s.io/gateway-api/apis/v1"
	gatewayv1alpha2 "sigs.k8s.io/gateway-api/apis/v1alpha2"
	gatewayv1beta1 "sigs.k8s.io/gateway-api/apis/v1beta1"
	confv1 "sigs.k8s.io/gateway-api/conformance/apis/v1"
	"sigs.k8s.io/gateway-api/conformance/tests"
	"sigs.k8s.io/gateway-api/conformance/utils/flags"
	"sigs.k8s.io/gateway-api/conformance/utils/suite"
)

var (
	cfg                  *rest.Config
	k8sClientset         *kubernetes.Clientset
	mgrClient            client.Client
	supportedFeatures    sets.Set[suite.SupportedFeature]
	exemptFeatures       sets.Set[suite.SupportedFeature]
	namespaceLabels      map[string]string
	namespaceAnnotations map[string]string
	implementation       *confv1.Implementation
	mode                 string
	allowCRDsMismatch    bool
	conformanceProfiles  sets.Set[suite.ConformanceProfileName]
	skipTests            []string
)

func TestConformance(t *testing.T) {
	var err error
	cfg, err = config.GetConfig()
	if err != nil {
		t.Fatalf("Error loading Kubernetes config: %v", err)
	}
	mgrClient, err = client.New(cfg, client.Options{})
	if err != nil {
		t.Fatalf("Error initializing Kubernetes client: %v", err)
	}
	k8sClientset, err = kubernetes.NewForConfig(cfg)
	if err != nil {
		t.Fatalf("Error initializing Kubernetes REST client: %v", err)
	}

	gatewayv1alpha2.AddToScheme(mgrClient.Scheme())
	gatewayv1beta1.AddToScheme(mgrClient.Scheme())
	gatewayv1.AddToScheme(mgrClient.Scheme())
	apiextensionsv1.AddToScheme(mgrClient.Scheme())

	// conformance flags
	supportedFeatures = suite.ParseSupportedFeatures(*flags.SupportedFeatures)
	exemptFeatures = suite.ParseSupportedFeatures(*flags.ExemptFeatures)
	skipTests = suite.ParseSkipTests(*flags.SkipTests)
	namespaceLabels = suite.ParseKeyValuePairs(*flags.NamespaceLabels)
	namespaceAnnotations = suite.ParseKeyValuePairs(*flags.NamespaceAnnotations)
	conformanceProfiles = suite.ParseConformanceProfiles(*flags.ConformanceProfiles)
	if len(conformanceProfiles) == 0 {
		t.Fatal("conformance profiles need to be given")
	}
	mode = *flags.Mode
	allowCRDsMismatch = *flags.AllowCRDsMismatch

	implementation, err = suite.ParseImplementation(
		*flags.ImplementationOrganization,
		*flags.ImplementationProject,
		*flags.ImplementationURL,
		*flags.ImplementationVersion,
		*flags.ImplementationContact,
	)
	if err != nil {
		t.Fatalf("Error parsing implementation's details: %v", err)
	}
	testConformance(t)
}

func testConformance(t *testing.T) {
	t.Logf("Running conformance tests with %s GatewayClass\n cleanup: %t\n debug: %t\n enable all features: %t \n supported features: [%v]\n exempt features: [%v]",
		*flags.GatewayClassName, *flags.CleanupBaseResources, *flags.ShowDebug, *flags.EnableAllSupportedFeatures, *flags.SupportedFeatures, *flags.ExemptFeatures)

	cSuite, err := suite.NewConformanceTestSuite(
		suite.ConformanceOptions{
			Client:     mgrClient,
			RestConfig: cfg,
			// This clientset is needed in addition to the client only because
			// controller-runtime client doesn't support non CRUD sub-resources yet (https://github.com/kubernetes-sigs/controller-runtime/issues/452).
			Clientset:                  k8sClientset,
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
			Mode:                       mode,
			AllowCRDsMismatch:          allowCRDsMismatch,
			Implementation:             *implementation,
			ConformanceProfiles:        conformanceProfiles,
		})
	if err != nil {
		t.Fatalf("error creating conformance test suite: %v", err)
	}

	cSuite.Setup(t)
	cSuite.Run(t, tests.ConformanceTests)
	report, err := cSuite.Report()
	if err != nil {
		t.Fatalf("error generating conformance profile report: %v", err)
	}
	writeReport(t.Logf, *report, *flags.ReportOutput)
}

func writeReport(logf func(string, ...any), report confv1.ConformanceReport, output string) error {
	rawReport, err := yaml.Marshal(report)
	if err != nil {
		return err
	}

	if output != "" {
		if err = os.WriteFile(output, rawReport, 0o600); err != nil {
			return err
		}
	}
	logf("Conformance report:\n%s", string(rawReport))

	return nil
}
