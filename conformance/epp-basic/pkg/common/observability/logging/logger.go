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
	"context"

	"github.com/go-logr/logr"
	uberzap "go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
)

// atomicLevel is shared between InitSetupLogging and InitLogging so the log
// level can be adjusted after the controller-runtime delegation is fulfilled.
var atomicLevel = uberzap.NewAtomicLevelAt(zapcore.InfoLevel)

func customLevelEncoder(l zapcore.Level, enc zapcore.PrimitiveArrayEncoder) {
	if l >= 0 {
		zapcore.LowercaseLevelEncoder(l, enc)
		return
	}

	switch l {
	case zapcore.Level(-1 * DEBUG): // -4
		enc.AppendString("debug")
	case zapcore.Level(-1 * TRACE): // -5
		enc.AppendString("trace")
	default:
		if l >= zapcore.Level(-1*VERBOSE) { // >= -3 (i.e. V(1)-V(3))
			enc.AppendString("info")
		} else {
			enc.AppendString("trace")
		}
	}
}

func InitSetupLogging() {
	config := uberzap.NewProductionEncoderConfig()
	config.EncodeLevel = customLevelEncoder

	logger := zap.New(
		zap.Level(atomicLevel),
		zap.RawZapOpts(uberzap.AddCaller()),
		zap.Encoder(zapcore.NewJSONEncoder(config)),
	)
	ctrl.SetLogger(logger)
}

func InitLogging(opts *zap.Options) {
	// Update the shared atomic level so the logger created in InitSetupLogging
	// (and all loggers derived from it) pick up the new verbosity.
	// ctrl.SetLogger only fulfills the delegation once, so calling it again
	// after InitSetupLogging is a no-op. Instead we mutate the atomic level.
	if opts.Level != nil {
		switch lvl := opts.Level.(type) {
		case uberzap.AtomicLevel:
			atomicLevel.SetLevel(lvl.Level())
		case zapcore.Level:
			atomicLevel.SetLevel(lvl)
		}
	}
}

// NewTestLogger creates a new Zap logger using the dev mode.
func NewTestLogger() logr.Logger {
	return zap.New(
		zap.UseDevMode(true),
		zap.Level(uberzap.NewAtomicLevelAt(zapcore.Level(-1*TRACE))),
		zap.RawZapOpts(uberzap.AddCaller()),
	)
}

// NewTestLoggerIntoContext creates a new Zap logger using the dev mode and inserts it into the given context.
func NewTestLoggerIntoContext(ctx context.Context) context.Context {
	return log.IntoContext(ctx, NewTestLogger())
}
