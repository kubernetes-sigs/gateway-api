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

	"github.com/spf13/cobra"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/cli-runtime/pkg/genericiooptions"
	"k8s.io/klog/v2"

	cmdanalyze "sigs.k8s.io/gateway-api/gwctl/cmd/analyze"
	cmdapply "sigs.k8s.io/gateway-api/gwctl/cmd/apply"
	cmddelete "sigs.k8s.io/gateway-api/gwctl/cmd/delete"
	cmdget "sigs.k8s.io/gateway-api/gwctl/cmd/get"
	"sigs.k8s.io/gateway-api/gwctl/pkg/common"
	"sigs.k8s.io/gateway-api/gwctl/pkg/version"
)

func newRootCmd() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "gwctl",
		Short: "gwctl is a command-line tool for exploring Gateway API resources.",
		Long:  `gwctl provides a familiar kubectl-like interface for navigating the Kubernetes Gateway API's multi-resource model, offering visibility into resource relationships and the policies that affect them.`,
	}

	globalConfig := genericclioptions.NewConfigFlags(true)
	globalConfig.AddFlags(rootCmd.PersistentFlags())

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
		if err := klogFlags.Set("v", fmt.Sprintf("%v", verbosity)); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to configure verbosity for logging")
		}
	})

	ioStreams := genericiooptions.IOStreams{In: os.Stdin, Out: os.Stdout, ErrOut: os.Stderr}
	factory := common.NewFactory(globalConfig)
	rootCmd.AddCommand(cmdapply.NewCmd(factory, ioStreams))
	rootCmd.AddCommand(cmdget.NewCmd(factory, ioStreams, false))
	rootCmd.AddCommand(cmdget.NewCmd(factory, ioStreams, true))
	rootCmd.AddCommand(cmddelete.NewCmd(factory, ioStreams))
	rootCmd.AddCommand(cmdanalyze.NewCmd(factory, ioStreams))
	rootCmd.AddCommand(newVersionCommand())

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

func newVersionCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "version",
		Short: "Print the version information of gwctl",
		Long:  `Print the version information of gwctl, including version, git commit and build date.`,
		Run: func(*cobra.Command, []string) {
			fmt.Println(version.GetVersionInfo())
		},
	}
	return cmd
}
