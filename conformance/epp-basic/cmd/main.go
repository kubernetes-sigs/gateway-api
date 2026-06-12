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

package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/spf13/pflag"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	metricsserver "sigs.k8s.io/controller-runtime/pkg/metrics/server"

	"sigs.k8s.io/gateway-api/conformance/epp-basic/pkg/common"
	"sigs.k8s.io/gateway-api/conformance/epp-basic/pkg/datastore"
	"sigs.k8s.io/gateway-api/conformance/epp-basic/pkg/server"
)

func main() {
	zapOpts := zap.Options{Development: true}
	zapOpts.BindFlags(flag.CommandLine)
	pflag.CommandLine.AddGoFlagSet(flag.CommandLine)

	opts := server.NewOptions()
	opts.AddFlags(pflag.CommandLine)
	pflag.Parse()

	ctrl.SetLogger(zap.New(zap.UseFlagOptions(&zapOpts)))
	setupLog := ctrl.Log.WithName("setup")

	if err := opts.Complete(); err != nil {
		setupLog.Error(err, "Failed to complete options")
		os.Exit(1)
	}
	if err := opts.Validate(); err != nil {
		setupLog.Error(err, "Failed to validate options")
		os.Exit(1)
	}

	ctx := ctrl.SetupSignalHandler()

	cfg, err := ctrl.GetConfig()
	if err != nil {
		setupLog.Error(err, "Failed to get Kubernetes rest config")
		os.Exit(1)
	}

	ds := datastore.NewDatastore(ctx)

	gknn := common.GKNN{
		NamespacedName: types.NamespacedName{Name: opts.PoolName, Namespace: opts.PoolNamespace},
		GroupKind: schema.GroupKind{
			Group: opts.PoolGroup,
			Kind:  "InferencePool",
		},
	}

	metricsServerOptions := metricsserver.Options{
		BindAddress: fmt.Sprintf(":%d", opts.MetricsPort),
	}

	mgr, err := server.NewDefaultManager(server.NewControllerConfig(true), gknn, cfg, metricsServerOptions)
	if err != nil {
		setupLog.Error(err, "Failed to create controller manager")
		os.Exit(1)
	}

	runner := &server.ExtProcServerRunner{
		GrpcPort:       opts.GRPCPort,
		GrpcHealthPort: opts.GRPCHealthPort,
		GKNN:           gknn,
		Datastore:      ds,
		HealthChecking: opts.HealthChecking,
		SecureServing:  opts.SecureServing,
	}

	if err := runner.SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "Failed to setup EPP controllers")
		os.Exit(1)
	}

	if err := mgr.Add(runner.AsRunnable(setupLog)); err != nil {
		setupLog.Error(err, "Failed to add runner to manager")
		os.Exit(1)
	}

	// Start health server directly so it is available immediately, before cache sync.
	go func() {
		if err := runner.HealthServerRunnable(setupLog).Start(ctx); err != nil {
			setupLog.Error(err, "Health server exited")
		}
	}()

	setupLog.Info("Starting manager")
	if err := mgr.Start(ctx); err != nil {
		setupLog.Error(err, "Problem running manager")
		os.Exit(1)
	}
}
