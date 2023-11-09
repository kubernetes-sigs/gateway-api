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
	"context"
	"flag"
	"os"
	"path"

	"github.com/spf13/cobra"
	cobraflag "github.com/spf13/pflag"
	"k8s.io/klog/v2"

	"sigs.k8s.io/gateway-api/gwctl/pkg/cmd/describe"
	"sigs.k8s.io/gateway-api/gwctl/pkg/cmd/get"
	cmdutils "sigs.k8s.io/gateway-api/gwctl/pkg/cmd/utils"
	"sigs.k8s.io/gateway-api/gwctl/pkg/common"
	"sigs.k8s.io/gateway-api/gwctl/pkg/policymanager"
)

func main() {
	klog.InitFlags(nil)
	flag.Parse()
	cobraflag.CommandLine.AddGoFlagSet(flag.CommandLine)

	kubeconfig := os.Getenv("KUBECONFIG")
	if kubeconfig == "" {
		kubeconfig = path.Join(os.Getenv("HOME"), ".kube/config")
	}

	k8sClients, err := common.NewK8sClients(kubeconfig)
	if err != nil {
		panic(err)
	}

	policyManager := policymanager.New(k8sClients.DC)
	if err := policyManager.Init(context.Background()); err != nil {
		panic(err)
	}

	params := &cmdutils.CmdParams{
		K8sClients:    k8sClients,
		PolicyManager: policyManager,
		Out:           os.Stdout,
	}

	rootCmd := &cobra.Command{
		Use: "gwctl",
	}
	rootCmd.AddCommand(get.NewGetCommand(params))
	rootCmd.AddCommand(describe.NewDescribeCommand(params))

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
