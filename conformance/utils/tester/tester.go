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

package tester

import (
	"fmt"
	"testing"
	"time"
)

// Tester allows users to add custom logging shims to conformance tests.
type Tester interface {
	testing.TB

	// Run calls a subtest and wraps the testing.T.Run call
	Run(name string, f func(Tester)) bool

	// Parallel should pass through to testing.T.Parallel
	Parallel()
}

// New returns a new impl.
func New(t *testing.T) Tester {
	return &loggingTester{T: t}
}

// loggingTester performs basic wrapping of testing.T
type loggingTester struct {
	*testing.T
}

// Run implements Tester
func (lt *loggingTester) Run(name string, f func(Tester)) bool {
	return lt.T.Run(name, func(t *testing.T) {
		f(&loggingTester{T: t})
	})
}

// Parallel implements Tester
func (lt *loggingTester) Parallel() {
	lt.T.Parallel()
}

func (lt *loggingTester) formatf(level, format string, args ...any) string {
	return fmt.Sprintf("%s%s] %s", level, time.Now().Format(time.RFC3339), fmt.Sprintf(format, args...))
}

func (lt *loggingTester) format(level string, args ...any) string {
	return fmt.Sprintf("%s%s] %s", level, time.Now().Format(time.RFC3339), fmt.Sprint(args...))
}

// Log implements testing.TB
func (lt *loggingTester) Log(args ...any) {
	lt.T.Helper()
	lt.T.Log(lt.format("I", args...))
}

// Logf implements testing.TB
func (lt *loggingTester) Logf(format string, args ...any) {
	lt.T.Helper()
	lt.T.Logf(lt.formatf("I", format, args...))
}

// Error implements testing.TB
func (lt *loggingTester) Error(args ...any) {
	lt.T.Helper()
	lt.T.Error(lt.format("E", args...))
}

// Errorf implements testing.TB
func (lt *loggingTester) Errorf(format string, args ...any) {
	lt.T.Helper()
	lt.T.Errorf(lt.formatf("E", format, args...))
}

// Fatal implements testing.TB
func (lt *loggingTester) Fatal(args ...any) {
	lt.T.Helper()
	lt.T.Fatal(lt.format("F", args...))
}

// Fatalf implements testing.TB
func (lt *loggingTester) Fatalf(format string, args ...any) {
	lt.T.Helper()
	lt.T.Fatalf(lt.formatf("F", format, args...))
}

// Skip implements testing.TB
func (lt *loggingTester) Skip(args ...any) {
	lt.T.Helper()
	lt.T.Skip(lt.format("S", args...))
}

// Skipf implements testing.TB
func (lt *loggingTester) Skipf(format string, args ...any) {
	lt.T.Helper()
	lt.T.Skipf(lt.formatf("S", format, args...))
}
