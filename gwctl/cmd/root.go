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
	"flag"
	"fmt"
	"os"
	"path"

	"github.com/spf13/cobra"
	"k8s.io/klog/v2"

	"sigs.k8s.io/gateway-api/gwctl/pkg/utils"
)

func newRootCmd() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "gwctl",
		Short: "gwctl is a command-line tool for exploring Gateway API resources.",
		Long:  `gwctl provides a familiar kubectl-like interface for navigating the Kubernetes Gateway API's multi-resource model, offering visibility into resource relationships and the policies that affect them.`,
	}

	var kubeConfigPath string
	rootCmd.PersistentFlags().StringVar(&kubeConfigPath, "kubeconfig", "", "path to kubeconfig file (default is the KUBECONFIG environment variable and if it isn't set, falls back to $HOME/.kube/config)")

	// Initialize flags for klog.
	//
	// These are not directly added to the rootCmd since we ony want to expose the
	// verbosity (-v) flag and not the rest. To achieve that, we'll define a
	// separate verbosity flag whose value we'll propagate to the klogFlags.
	var verbosity int
	rootCmd.PersistentFlags().IntVarP(&verbosity, "v", "v", 0, "number for the log level verbosity (defaults to 0)")
	klogFlags := flag.NewFlagSet("klog", flag.ExitOnError)
	klog.InitFlags(klogFlags)

	cobra.OnInitialize(func() {
		if kubeConfigPath == "" {
			kubeConfigPath = os.Getenv("KUBECONFIG")
			if kubeConfigPath == "" {
				kubeConfigPath = path.Join(os.Getenv("HOME"), ".kube/config")
			}
		}
		if err := klogFlags.Set("v", fmt.Sprintf("%v", verbosity)); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to configure verbosity for logging")
		}
	})

	factory := utils.NewFactory(&kubeConfigPath)

	rootCmd.AddCommand(NewSubCommand(factory, os.Stdout, commandNameGet))
	rootCmd.AddCommand(NewSubCommand(factory, os.Stdout, commandNameDescribe))

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

func addNamespaceFlag(p *string, cmd *cobra.Command) {
	cmd.Flags().StringVarP(p, "namespace", "n", "default", "")
}

func addAllNamespacesFlag(p *bool, cmd *cobra.Command) {
	cmd.Flags().BoolVarP(p, "all-namespaces", "A", false, "If present, list requested resources from all namespaces.")
}

func addLabelSelectorFlag(p *string, cmd *cobra.Command) {
	cmd.Flags().StringVarP(p, "selector", "l", "", "Selector (label query) to filter on, supports '=', '==', and '!='.(e.g. -l key1=value1,key2=value2). Matching objects must satisfy all of the specified label constraints.")
}

func addOutputFormatFlag(p *string, cmd *cobra.Command) {
	cmd.Flags().StringVarP(p, "output", "o", "", `Output format. Must be one of (yaml, json)`)
}

func addForFlag(p *string, cmd *cobra.Command) {
	cmd.Flags().StringVar(p, "for", "", `Filter results to only those related to the specified resource. Format: TYPE[/NAMESPACE]/NAME. Not specifying a NAMESPACE assumes the 'default' value. Examples: gateway/ns2/foo-gateway, httproute/bar-httproute, service/ns1/my-svc`)
}
