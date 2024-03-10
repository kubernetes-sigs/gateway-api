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

package cmd

import (
	"context"
	"flag"
	"fmt"
	"os"
	"path"

	"github.com/spf13/cobra"
	"k8s.io/klog/v2"

	"sigs.k8s.io/gateway-api/gwctl/pkg/common"
	"sigs.k8s.io/gateway-api/gwctl/pkg/policymanager"
	cmdutils "sigs.k8s.io/gateway-api/gwctl/pkg/utils"
)

var (
	kubeConfigPath string
)

func newRootCmd() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "gwctl",
		Short: "gwctl is a command-line tool for exploring Gateway API resources.",
		Long:  `gwctl provides a familiar kubectl-like interface for navigating the Kubernetes Gateway API's multi-resource model, offering visibility into resource relationships and the policies that affect them.`,
	}
	cobra.OnInitialize(initConfig)
	rootCmd.PersistentFlags().StringVar(&kubeConfigPath, "kubeconfig", "", "path to kubeconfig file (default is the KUBECONFIG environment variable and if it isn't set, falls back to $HOME/.kube/config)")

	// initialize logging flags in a new flag set
	// otherwise it conflicts with cobra's flags
	klogFlags := flag.NewFlagSet("klog", flag.ExitOnError)
	klog.InitFlags(klogFlags)

	rootCmd.PersistentFlags().AddGoFlagSet(klogFlags)

	rootCmd.AddCommand(NewGetCommand())
	rootCmd.AddCommand(NewDescribeCommand())

	return rootCmd
}

func Execute() {
	rootCmd := newRootCmd()
	err := rootCmd.Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to execute command: %v\n", err)
		os.Exit(1)
	}
}

func initConfig() {
	if kubeConfigPath == "" {
		kubeConfigPath = os.Getenv("KUBECONFIG")
		if kubeConfigPath == "" {
			kubeConfigPath = path.Join(os.Getenv("HOME"), ".kube/config")
		}
	}
}

func getParams(path string) *cmdutils.CmdParams {
	k8sClients, err := common.NewK8sClients(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to create k8s clients: %v\n", err)
		os.Exit(1)
	}

	policyManager := policymanager.New(k8sClients.DC)
	if err := policyManager.Init(context.Background()); err != nil {
		fmt.Fprintf(os.Stderr, "failed to initialize policy manager: %v\n", err)
		os.Exit(1)
	}

	params := &cmdutils.CmdParams{
		K8sClients:    k8sClients,
		PolicyManager: policyManager,
		Out:           os.Stdout,
	}

	return params
}
