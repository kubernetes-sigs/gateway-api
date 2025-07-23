/*
Copyright 2025 The Kubernetes Authors.

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

package crd_test

import (
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/envtest"

	v1 "sigs.k8s.io/gateway-api/apis/v1"
	"sigs.k8s.io/gateway-api/apis/v1alpha2"
	"sigs.k8s.io/gateway-api/apis/v1alpha3"
	"sigs.k8s.io/gateway-api/apis/v1beta1"
	apisxv1alpha1 "sigs.k8s.io/gateway-api/apisx/v1alpha1"
)

// This test is a replacement for the previously existing tests inside
// tests-crds-validation.sh.
// Instead of relying on a KinD cluster, we can rely on an `envtest` which
// usually is faster to create and teardown the environment
func TestCRDValidation(t *testing.T) {
	scheme := runtime.NewScheme()
	var testEnv *envtest.Environment
	var err error

	// We will extract these from envTest and create our own exec.Cmd to reproduce
	// a user doing "kubectl" commands
	var kubectlLocation, kubeconfigLocation string

	crdChannel := "standard"

	v1alpha3.Install(scheme)
	v1alpha2.Install(scheme)
	v1beta1.Install(scheme)
	v1.Install(scheme)
	apisxv1alpha1.Install(scheme)

	// Add core APIs in case we refer secrets, services and configmaps
	corev1.AddToScheme(scheme)

	// The version used here MUST reflect the available versions at
	// controller-runtime repo: https://raw.githubusercontent.com/kubernetes-sigs/controller-tools/HEAD/envtest-releases.yaml
	// If the envvar is not passed, the latest GA will be used
	k8sVersion := os.Getenv("K8S_VERSION")
	if requestedCRDChannel, ok := os.LookupEnv("CRD_CHANNEL"); ok {
		crdChannel = requestedCRDChannel
	}

	t.Run("should be able to start test environment", func(_ *testing.T) {
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

		_, err = testEnv.Start()
		if err != nil {
			panic(fmt.Sprintf("Error initializing test environment: %v", err))
		}
	})

	t.Cleanup(func() {
		require.NoError(t, testEnv.Stop())
	})

	t.Run("should be able to set kubectl and kubeconfig and connect to the cluster", func(t *testing.T) {
		kubectlLocation = testEnv.ControlPlane.KubectlPath
		require.NotEmpty(t, kubectlLocation)

		kubeconfigLocation = fmt.Sprintf("%s/kubeconfig", filepath.Dir(kubectlLocation))
		require.NoError(t, os.WriteFile(kubeconfigLocation, testEnv.KubeConfig, 0o600))

		apiResources, err := executeKubectlCommand(t, kubectlLocation, kubeconfigLocation, []string{"api-resources"})
		require.NoError(t, err)
		require.Contains(t, apiResources, "gateway.networking.k8s.io/v1")
	})

	t.Run("should be able to install standard examples", func(t *testing.T) {
		output, err := executeKubectlCommand(t, kubectlLocation, kubeconfigLocation, []string{"apply", "--recursive", "-f", filepath.Join("..", "..", "..", "examples", "standard")})
		assert.NoError(t, err, "output", output)
	})

	t.Run("should be able to install experimental examples", func(t *testing.T) {
		if crdChannel != "experimental" {
			t.Skipf("experimental channel not being tested")
		}
		output, err := executeKubectlCommand(t, kubectlLocation, kubeconfigLocation, []string{"apply", "--recursive", "-f", filepath.Join("..", "..", "..", "examples", "experimental")})
		assert.NoError(t, err, "output", output)
	})

	t.Run("should expect an error in case of validation failure", func(t *testing.T) {
		files, err := getInvalidExamplesFiles(t, crdChannel)
		require.NoError(t, err)

		for _, example := range files {
			t.Run(fmt.Sprintf("validate example %s", example), func(t *testing.T) {
				output, err := executeKubectlCommand(t, kubectlLocation, kubeconfigLocation, []string{"apply", "-f", example})
				require.Error(t, err)
				assert.True(t, expectedValidationError(output), "output does not contain the expected error", output)
			})
		}
	})
}

func expectedValidationError(cmdoutput string) bool {
	return strings.Contains(cmdoutput, "is invalid") ||
		strings.Contains(cmdoutput, "missing required field") ||
		strings.Contains(cmdoutput, "denied request") ||
		strings.Contains(cmdoutput, "Invalid value")
}

func executeKubectlCommand(t *testing.T, kubectl, kubeconfig string, args []string) (string, error) {
	t.Helper()

	cacheDir := filepath.Dir(kubeconfig)
	args = append([]string{"--cache-dir", cacheDir}, args...)

	cmd := exec.Command(kubectl, args...)
	cmd.Env = []string{
		fmt.Sprintf("KUBECONFIG=%s", kubeconfig),
	}

	output, err := cmd.CombinedOutput()
	return string(output), err
}

func getInvalidExamplesFiles(t *testing.T, crdChannel string) ([]string, error) {
	t.Helper()

	var files []string
	err := filepath.WalkDir(filepath.Join("..", "..", "..", "hack", "invalid-examples"), func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if crdChannel == "standard" && strings.Contains(path, "experimental") {
			return nil
		}

		if !d.IsDir() && filepath.Ext(path) == ".yaml" {
			files = append(files, path)
		}
		return nil
	})
	return files, err
}
