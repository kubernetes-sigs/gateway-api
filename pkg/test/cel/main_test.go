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
	"path"
	"strings"
	"testing"

	v1 "sigs.k8s.io/gateway-api/apis/v1"
	"sigs.k8s.io/gateway-api/apis/v1alpha2"
	"sigs.k8s.io/gateway-api/apis/v1alpha3"
	"sigs.k8s.io/gateway-api/apis/v1beta1"

	"k8s.io/client-go/tools/clientcmd"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var k8sClient client.Client

func TestMain(m *testing.M) {
	kubeconfig := os.Getenv("KUBECONFIG")
	if kubeconfig == "" {
		kubeconfig = path.Join(os.Getenv("HOME"), ".kube/config")
	}

	restConfig, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		panic(fmt.Sprintf("Failed to get restConfig from BuildConfigFromFlags: %v", err))
	}

	k8sClient, err = client.New(restConfig, client.Options{})
	if err != nil {
		panic(fmt.Sprintf("Error initializing Kubernetes client: %v", err))
	}
	v1alpha3.Install(k8sClient.Scheme())
	v1alpha2.Install(k8sClient.Scheme())
	v1beta1.Install(k8sClient.Scheme())
	v1.Install(k8sClient.Scheme())

	os.Exit(m.Run())
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
