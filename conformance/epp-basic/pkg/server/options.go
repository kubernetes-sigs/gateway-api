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
	"github.com/spf13/pflag"
	"k8s.io/apimachinery/pkg/util/sets"
)

const (
	DefaultGrpcPort       = 9002
	DefaultGrpcHealthPort = 9003
	DefaultPoolNamespace  = "default" // default when pool namespace is empty (CLI flag default is empty)
)

// Options contains configuration values necessary to create and run the lwepp.
type Options struct {
	GRPCPort            int    // gRPC port used for communicating with Envoy proxy.
	GRPCHealthPort      int    // Port for gRPC liveness and readiness probes.
	PoolGroup           string // Kubernetes resource group of the InferencePool this Endpoint Picker is associated with.
	PoolNamespace       string // Namespace of the InferencePool this Endpoint Picker is associated with.
	PoolName            string // Name of the InferencePool this Endpoint Picker is associated with.
	MetricsPort         int    // The metrics port exposed by lwepp.
	EndpointTargetPorts []int  // Target ports of model server pods.
	HealthChecking      bool   // Enables health checking.
	SecureServing       bool   // Enables TLS on the ext-proc gRPC server.
}

// NewOptions returns a new Options struct initialized with the default values.
func NewOptions() *Options {
	return &Options{
		GRPCPort:            DefaultGrpcPort,
		GRPCHealthPort:      DefaultGrpcHealthPort,
		PoolGroup:           "inference.networking.k8s.io",
		EndpointTargetPorts: []int{},
		MetricsPort:         9090,
		SecureServing:       true,
	}
}

func (opts *Options) AddFlags(fs *pflag.FlagSet) {
	if fs == nil {
		fs = pflag.CommandLine
	}

	fs.IntVar(&opts.GRPCPort, "grpc-port", opts.GRPCPort, "gRPC port used for communicating with Envoy proxy.")
	fs.StringVar(&opts.PoolGroup, "pool-group", opts.PoolGroup,
		"Kubernetes resource group of the InferencePool this Endpoint Picker is associated with. Only `inference.networking.k8s.io/v1` is currently supported.")
	fs.StringVar(&opts.PoolNamespace, "pool-namespace", opts.PoolNamespace,
		"Namespace of the InferencePool this Endpoint Picker is associated with.")
	fs.StringVar(&opts.PoolName, "pool-name", opts.PoolName, "Name of the InferencePool this Endpoint Picker is associated with.")
	fs.IntVar(&opts.MetricsPort, "metrics-port", opts.MetricsPort, "The metrics port exposed by lwepp.")
	fs.IntSliceVar(&opts.EndpointTargetPorts, "endpoint-target-ports", opts.EndpointTargetPorts, "Target ports of model server pods. "+
		"Format: a comma-separated list of numbers without whitespace (e.g., '3000,3001,3002').")
	fs.IntVar(&opts.GRPCHealthPort, "grpc-health-port", opts.GRPCHealthPort, "Port for gRPC liveness and readiness probes.")
	fs.BoolVar(&opts.HealthChecking, "health-checking", opts.HealthChecking, "Enables health checking.")
	fs.BoolVar(&opts.SecureServing, "secure-serving", opts.SecureServing, "Enables TLS on the ext-proc gRPC server.")
}

func (opts *Options) Complete() error {
	opts.EndpointTargetPorts = removeDuplicatePorts(opts.EndpointTargetPorts)
	return nil
}

func (opts *Options) Validate() error {
	return nil
}

func removeDuplicatePorts(ports []int) []int {
	seen := sets.NewInt()
	unique := make([]int, 0, len(ports))

	for _, val := range ports {
		if !seen.Has(val) {
			unique = append(unique, val)
			seen.Insert(val)
		}
	}
	return unique
}
