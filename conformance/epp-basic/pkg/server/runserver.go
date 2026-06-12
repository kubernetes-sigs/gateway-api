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

package server

import (
	"context"
	"crypto/tls"
	"fmt"
	"time"

	extProcPb "github.com/envoyproxy/go-control-plane/envoy/service/ext_proc/v3"
	"github.com/go-logr/logr"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/health"
	healthgrpc "google.golang.org/grpc/health/grpc_health_v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/manager"

	"sigs.k8s.io/gateway-api/conformance/epp-basic/internal/runnable"
	tlsutil "sigs.k8s.io/gateway-api/conformance/epp-basic/internal/tls"
	"sigs.k8s.io/gateway-api/conformance/epp-basic/pkg/common"
	"sigs.k8s.io/gateway-api/conformance/epp-basic/pkg/controller"
	"sigs.k8s.io/gateway-api/conformance/epp-basic/pkg/datastore"
	"sigs.k8s.io/gateway-api/conformance/epp-basic/pkg/handlers"
)

// ExtProcServerRunner provides methods to manage an external process server.
type ExtProcServerRunner struct {
	GrpcPort       int
	GrpcHealthPort int
	GKNN           common.GKNN
	Datastore      datastore.Datastore
	HealthChecking bool
	SecureServing  bool
}

// NewDefaultExtProcServerRunner creates a runner with default values.
// Note: Dependencies like Datastore, Scheduler, SD need to be set separately.
func NewDefaultExtProcServerRunner() *ExtProcServerRunner {
	opts := NewOptions()
	if opts.PoolNamespace == "" {
		opts.PoolNamespace = DefaultPoolNamespace
	}

	gknn := common.GKNN{
		NamespacedName: types.NamespacedName{Name: opts.PoolName, Namespace: opts.PoolNamespace},
		GroupKind: schema.GroupKind{
			Group: opts.PoolGroup,
			Kind:  "InferencePool",
		},
	}
	return &ExtProcServerRunner{
		GrpcPort:       opts.GRPCPort,
		GKNN:           gknn,
		HealthChecking: opts.HealthChecking,
		// Datastore can be assigned later.
	}
}

// SetupWithManager sets up the runner with the given manager.
func (r *ExtProcServerRunner) SetupWithManager(mgr ctrl.Manager) error {
	if err := (&controller.InferencePoolReconciler{
		Datastore: r.Datastore,
		Reader:    mgr.GetClient(),
	}).SetupWithManager(mgr); err != nil {
		return fmt.Errorf("failed setting up InferencePoolReconciler - %w", err)
	}

	if err := (&controller.PodReconciler{
		Datastore: r.Datastore,
		Reader:    mgr.GetClient(),
	}).SetupWithManager(mgr); err != nil {
		return fmt.Errorf("failed setting up PodReconciler - %w", err)
	}
	return nil
}

// AsRunnable returns a Runnable that can be used to start the ext-proc gRPC server.
// The runnable implements LeaderElectionRunnable with leader election disabled.
func (r *ExtProcServerRunner) AsRunnable(logger logr.Logger) manager.Runnable {
	return runnable.NoLeaderElection(manager.RunnableFunc(func(ctx context.Context) error {
		var srv *grpc.Server
		if r.SecureServing {
			cert, err := tlsutil.CreateSelfSignedTLSCertificate(logger)
			if err != nil {
				return fmt.Errorf("failed to create self-signed certificate: %w", err)
			}
			creds := credentials.NewTLS(&tls.Config{
				Certificates: []tls.Certificate{cert},
				NextProtos:   []string{"h2"},
			})
			srv = grpc.NewServer(grpc.Creds(creds))
		} else {
			srv = grpc.NewServer()
		}

		extProcServer := handlers.NewStreamingServer(r.Datastore)
		extProcPb.RegisterExternalProcessorServer(srv, extProcServer)

		if r.HealthChecking {
			healthcheck := health.NewServer()
			healthgrpc.RegisterHealthServer(srv, healthcheck)
			svcName := extProcPb.ExternalProcessor_ServiceDesc.ServiceName
			logger.Info("Setting ExternalProcessor service status to SERVING", "serviceName", svcName)
			healthcheck.SetServingStatus(svcName, healthgrpc.HealthCheckResponse_SERVING)
		}

		// Forward to the gRPC runnable.
		return runnable.GRPCServer("ext-proc", srv, r.GrpcPort).Start(ctx)
	}))
}

// HealthServerRunnable returns a Runnable that starts a dedicated gRPC health server
// on GrpcHealthPort for liveness and readiness probes.
func (r *ExtProcServerRunner) HealthServerRunnable(logger logr.Logger) manager.Runnable {
	return runnable.NoLeaderElection(manager.RunnableFunc(func(ctx context.Context) error {
		srv := grpc.NewServer()
		healthcheck := health.NewServer()
		healthgrpc.RegisterHealthServer(srv, healthcheck)
		healthcheck.SetServingStatus("inference-extension", healthgrpc.HealthCheckResponse_NOT_SERVING)
		logger.Info("Starting gRPC health server", "port", r.GrpcHealthPort)

		// Flip to SERVING once the pool has been synced into the datastore.
		go func() {
			for {
				select {
				case <-ctx.Done():
					return
				case <-time.After(100 * time.Millisecond):
					if r.Datastore.PoolHasSynced() {
						healthcheck.SetServingStatus("inference-extension", healthgrpc.HealthCheckResponse_SERVING)
						return
					}
				}
			}
		}()

		return runnable.GRPCServer("health", srv, r.GrpcHealthPort).Start(ctx)
	}))
}
