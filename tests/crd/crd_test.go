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
	"bytes"
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
	"time"

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
	t.Cleanup(func() {
		require.NoError(t, testEnv.Stop())
	})
	require.NoError(t, err, "Error initializing test environment")

	// Setup kubectl and kubeconfig
	kubectlLocation = testEnv.ControlPlane.KubectlPath
	require.NotEmpty(t, kubectlLocation, "Error initializing Kubectl")

	kubeconfigLocation = fmt.Sprintf("%s/kubeconfig", filepath.Dir(kubectlLocation))
	err = os.WriteFile(kubeconfigLocation, testEnv.KubeConfig, 0o600)
	require.NoError(t, err, "Error initializing kubeconfig")

	apiResources, err := executeKubectlCommand(t, kubectlLocation, kubeconfigLocation, []string{"api-resources"})
	require.NoError(t, err)
	require.Contains(t, apiResources, "gateway.networking.k8s.io/v1")

	t.Run("safeupgrades VAP should validate correctly", func(t *testing.T) {
		if crdChannel == "experimental" {
			t.Skipf("skipping safeupgrades VAP")
		}

		output, err := executeKubectlCommand(t, kubectlLocation, kubeconfigLocation,
			[]string{"apply", "--server-side", "--wait", "-f", filepath.Join("..", "..", "..", "config", "crd", "standard", "gateway.networking.k8s.io_vap_safeupgrades.yaml")})
		require.NoError(t, err)

		// Even though --wait is applied I noticed a race condition that causes tests to fail.
		time.Sleep(time.Second)

		t.Run("should be able to install standard CRDs", func(t *testing.T) {
			output, err = executeKubectlCommand(t, kubectlLocation, kubeconfigLocation,
				[]string{"apply", "--server-side", "--wait", "-f", filepath.Join("..", "..", "..", "config", "crd", "standard")})
			require.NoError(t, err)
		})

		t.Run("should not be able to install k8s.io experimental CRDs", func(t *testing.T) {
			t.Cleanup(func() {
				output, err = executeKubectlCommand(t, kubectlLocation, kubeconfigLocation,
					[]string{"delete", "--wait", "-f", filepath.Join("..", "..", "..", "config", "crd", "experimental", "*.k8s.*")})
			})

			output, err = executeKubectlCommand(t, kubectlLocation, kubeconfigLocation,
				[]string{"apply", "--server-side", "--wait", "-f", filepath.Join("..", "..", "..", "config", "crd", "experimental", "*.k8s.*")})
			require.Error(t, err)
			assert.Contains(t, output, "Error from server (Invalid)")
			assert.Contains(t, output, "ValidatingAdmissionPolicy 'safe-upgrades.gateway.networking.k8s.io' with binding 'safe-upgrades.gateway.networking.k8s.io' denied request")

			// Check that --api-group has no invalid crd's
			output, err = executeKubectlCommand(t, kubectlLocation, kubeconfigLocation, []string{"describe", "CustomResourceDefinition"})
			require.NoError(t, err)
			assert.NotContains(t, output, "gateway.networking.k8s.io/channel: experimental", "output contains 'gateway.networking.k8s.io/channel: experimental'")
		})

		t.Run("should be able to install x-k8s.io experimental CRDs", func(t *testing.T) {
			t.Cleanup(func() {
				output, err = executeKubectlCommand(t, kubectlLocation, kubeconfigLocation,
					[]string{"delete", "--wait", "-f", filepath.Join("..", "..", "..", "config", "crd", "experimental", "*.x-k8s.*")})
			})

			output, err = executeKubectlCommand(t, kubectlLocation, kubeconfigLocation,
				[]string{"apply", "--server-side", "--wait", "-f", filepath.Join("..", "..", "..", "config", "crd", "experimental", "*.x-k8s.*")})
			require.NoError(t, err)
		})

		t.Run("should not be able to install CRDs with an older version", func(t *testing.T) {
			t.Cleanup(func() {
				output, err = executeKubectlCommand(t, kubectlLocation, kubeconfigLocation,
					[]string{"delete", "--wait", "-f", filepath.Join("..", "..", "..", "config", "crd", "standard", "gateway.networking.k8s.io_httproutes.yaml")})
			})

			// Read test crd into []byte
			httpCrd, err := os.ReadFile(filepath.Join("..", "..", "..", "config", "crd", "standard", "gateway.networking.k8s.io_httproutes.yaml"))
			require.NoError(t, err)

			// do replace on gateway.networking.k8s.io/bundle-version: v1.4.0
			re := regexp.MustCompile(`gateway\.networking\.k8s\.io\/bundle-version: \S*`)
			sub := []byte("gateway.networking.k8s.io/bundle-version: v1.3.0")
			oldCrd := re.ReplaceAll(httpCrd, sub)

			// supply crd to stdin of cmd and kubectl apply -f -
			output, err = executeKubectlCommandStdin(t, kubectlLocation, kubeconfigLocation, bytes.NewReader(oldCrd), []string{"apply", "-f", "-"})

			require.Error(t, err)
			assert.Contains(t, output, "ValidatingAdmissionPolicy 'safe-upgrades.gateway.networking.k8s.io' with binding 'safe-upgrades.gateway.networking.k8s.io' denied request")
		})
	})

	t.Run("should be able to install standard examples", func(t *testing.T) {
		output, err := executeKubectlCommand(t, kubectlLocation, kubeconfigLocation, []string{"apply", "--recursive", "-f", filepath.Join("..", "..", "examples", "standard")})
		assert.NoError(t, err, "output", output)
	})

	t.Run("should be able to install experimental examples", func(t *testing.T) {
		if crdChannel != "experimental" {
			t.Skipf("experimental channel not being tested")
		}
		output, err := executeKubectlCommand(t, kubectlLocation, kubeconfigLocation, []string{"apply", "--recursive", "-f", filepath.Join("..", "..", "examples", "experimental")})
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

	cmd := exec.CommandContext(t.Context(), kubectl, args...)
	cmd.Env = []string{
		fmt.Sprintf("KUBECONFIG=%s", kubeconfig),
	}

	output, err := cmd.CombinedOutput()
	return string(output), err
}

func executeKubectlCommandStdin(t *testing.T, kubectl, kubeconfig string, stdin io.Reader, args []string) (string, error) {
	t.Helper()

	cacheDir := filepath.Dir(kubeconfig)
	args = append([]string{"--cache-dir", cacheDir}, args...)

	cmd := exec.Command(kubectl, args...)
	cmd.Env = []string{
		fmt.Sprintf("KUBECONFIG=%s", kubeconfig),
	}
	cmd.Stdin = stdin

	output, err := cmd.CombinedOutput()
	return string(output), err
}

func getInvalidExamplesFiles(t *testing.T, crdChannel string) ([]string, error) {
	t.Helper()

	var files []string
	err := filepath.WalkDir(filepath.Join("..", "..", "hack", "invalid-examples"), func(path string, d fs.DirEntry, err error) error {
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
