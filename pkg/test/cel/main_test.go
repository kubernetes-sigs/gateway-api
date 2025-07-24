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

package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	v1 "sigs.k8s.io/gateway-api/apis/v1"
	"sigs.k8s.io/gateway-api/apis/v1alpha2"
	"sigs.k8s.io/gateway-api/apis/v1alpha3"
	"sigs.k8s.io/gateway-api/apis/v1beta1"
	apisxv1alpha1 "sigs.k8s.io/gateway-api/apisx/v1alpha1"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
)

var k8sClient client.Client

func TestMain(m *testing.M) {
	scheme := runtime.NewScheme()
	var restConfig *rest.Config
	var testEnv *envtest.Environment
	var err error

	v1alpha3.Install(scheme)
	v1alpha2.Install(scheme)
	v1beta1.Install(scheme)
	v1.Install(scheme)
	apisxv1alpha1.Install(scheme)

	// Add core APIs in case we refer secrets, services and configmaps
	corev1.AddToScheme(scheme)

	// If one wants to use a local cluster, a KUBECONFIG envvar should be passed,
	// otherwise testenv will be used
	kubeconfig := os.Getenv("KUBECONFIG")
	if kubeconfig != "" {
		restConfig, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
		if err != nil {
			panic(fmt.Sprintf("Failed to get restConfig from BuildConfigFromFlags: %v", err))
		}
	} else {
		// The version used here MUST reflect the available versions at
		// controller-runtime repo: https://raw.githubusercontent.com/kubernetes-sigs/controller-tools/HEAD/envtest-releases.yaml
		// If the envvar is not passed, the latest GA will be used
		k8sVersion := os.Getenv("K8S_VERSION")

		crdChannel := "standard"
		if requestedCRDChannel, ok := os.LookupEnv("CRD_CHANNEL"); ok {
			crdChannel = requestedCRDChannel
		}

		testEnv = &envtest.Environment{
			Scheme:                      scheme,
			ErrorIfCRDPathMissing:       true,
			DownloadBinaryAssets:        true,
			DownloadBinaryAssetsVersion: k8sVersion,
			CRDInstallOptions: envtest.CRDInstallOptions{
				Paths: []string{
					filepath.Join("..", "..", "..", "config", "crd", crdChannel),
				},
				CleanUpAfterUse: true,
			},
		}

		restConfig, err = testEnv.Start()
		if err != nil {
			panic(fmt.Sprintf("Error initializing test environment: %v", err))
		}
	}

	k8sClient, err = client.New(restConfig, client.Options{
		Scheme: scheme,
	})
	if err != nil {
		panic(fmt.Sprintf("Error initializing Kubernetes client: %v", err))
	}

	rc := m.Run()
	if testEnv != nil {
		if err := testEnv.Stop(); err != nil {
			panic(fmt.Sprintf("error stopping test environment: %v", err))
		}
	}

	os.Exit(rc)
}

func ptrTo[T any](a T) *T {
	return &a
}

func celErrorStringMatches(got, want string) bool {
	gotL := strings.ToLower(got)
	wantL := strings.ToLower(want)

	// Starting in k8s v1.32, some CEL error messages changed to use "more" instead of "longer"
	alternativeWantL := strings.ReplaceAll(wantL, "longer", "more")

	// Starting in k8s v1.28, CEL error messages stopped adding spec and status prefixes to path names
	wantLAdjusted := strings.ReplaceAll(wantL, "spec.", "")
	wantLAdjusted = strings.ReplaceAll(wantLAdjusted, "status.", "")
	alternativeWantL = strings.ReplaceAll(alternativeWantL, "spec.", "")
	alternativeWantL = strings.ReplaceAll(alternativeWantL, "status.", "")

	// Enum validation messages changed in k8s v1.28:
	// Before: must be one of ['Exact', 'PathPrefix', 'RegularExpression']
	// After: supported values: "Exact", "PathPrefix", "RegularExpression"
	if strings.Contains(wantLAdjusted, "must be one of") {
		r := strings.NewReplacer(
			"must be one of", "supported values:",
			"[", "",
			"]", "",
			"'", "\"",
		)
		wantLAdjusted = r.Replace(wantLAdjusted)
	}
	return strings.Contains(gotL, wantL) || strings.Contains(gotL, wantLAdjusted) || strings.Contains(gotL, alternativeWantL)
}
