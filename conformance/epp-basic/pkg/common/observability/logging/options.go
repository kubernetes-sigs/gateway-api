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

package logging

import (
	"flag"

	"github.com/spf13/pflag"
	uberzap "go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
)

const (
	ZapLogLevelFlagName = "zap-log-level"
)

// LoggingOptions contains logging configuration for command-line flags.
type LoggingOptions struct {
	LogVerbosity int         // Number for the log level verbosity.
	ZapOptions   zap.Options // Zap logging options.

	// internal
	loggingFS *pflag.FlagSet // FlagSet used in AddFlags() and consulted in Complete()
}

// NewOptions returns a new LoggingOptions struct initialized with default values.
func NewOptions() *LoggingOptions {
	return &LoggingOptions{
		LogVerbosity: DEFAULT,
		ZapOptions:   zap.Options{Development: true},
	}
}

// AddFlags binds the LoggingOptions fields to command-line flags on the given FlagSet.
func (opts *LoggingOptions) AddFlags(fs *pflag.FlagSet) {
	if fs == nil {
		fs = pflag.CommandLine
	}
	opts.loggingFS = fs

	fs.IntVarP(&opts.LogVerbosity, "v", "v", opts.LogVerbosity,
		"Number for the log level verbosity.")

	// Bind zap flags (zap expects a standard Go FlagSet; pflag.FlagSet is not compatible).
	gofs := flag.NewFlagSet("zap", flag.ExitOnError)
	opts.ZapOptions.BindFlags(gofs)
	fs.AddGoFlagSet(gofs)
}

// Complete performs post-processing of parsed command-line arguments.
// Derives the zap log level from the -v flag when --zap-log-level is not set explicitly.
func (opts *LoggingOptions) Complete() error {
	zapLogLevelFlag := opts.loggingFS.Lookup(ZapLogLevelFlagName)
	if zapLogLevelFlag != nil && !zapLogLevelFlag.Changed {
		// See https://pkg.go.dev/sigs.k8s.io/controller-runtime/pkg/log/zap#Options.Level
		lvl := -1 * (opts.LogVerbosity)
		opts.ZapOptions.Level = uberzap.NewAtomicLevelAt(zapcore.Level(int8(lvl)))
		zapLogLevelFlag.Changed = true
	}
	return nil
}

// Validate checks the LoggingOptions for invalid values.
func (opts *LoggingOptions) Validate() error {
	// Log verbosity must be non-negative; set to default if invalid.
	if opts.LogVerbosity < 0 {
		opts.LogVerbosity = DEFAULT
	}
	return nil
}
