/*
Copyright 2021 The Kubernetes Authors.

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
	"bufio"
	"flag"
	"fmt"
	"io"
	"os"
	"os/signal"
	"path"
	"path/filepath"
	"syscall"
	"testing"

	"github.com/cucumber/godog"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/klog/v2"

	"sigs.k8s.io/gateway-api/conformance/tests/httpgateway"
	"sigs.k8s.io/gateway-api/conformance/utils"
)

var (
	godogFormat        string
	godogTags          string
	godogStopOnFailure bool
	godogNoColors      bool
	godogOutput        string
)

func TestMain(m *testing.M) {
	// register flags from klog (client-go verbose logging)
	klog.InitFlags(nil)

	flag.StringVar(&godogFormat, "format", "pretty", "Set godog format to use. Valid values are pretty and cucumber")
	flag.StringVar(&godogTags, "tags", "", "Tags for conformance test")
	flag.BoolVar(&godogStopOnFailure, "stop-on-failure ", false, "Stop when failure is found")
	flag.BoolVar(&godogNoColors, "no-colors", false, "Disable colors in godog output")
	flag.StringVar(&godogOutput, "output-directory", ".", "Output directory for test reports")

	flag.Parse()

	validFormats := sets.NewString("cucumber", "pretty")
	if !validFormats.Has(godogFormat) {
		klog.Fatalf("the godog format '%v' is not supported", godogFormat)
	}

	var err error
	utils.KubeClient, utils.GWClient, err = utils.LoadClientset()
	if err != nil {
		klog.Fatalf("error loading Kubernetes clientsets: %v", err)
	}

	if err = utils.CleanupNamespaces(); err != nil {
		klog.Fatalf("error deleting temporal namespaces: %v", err)
	}

	go handleSignals()

	os.Exit(m.Run())
}

var (
	features = map[string]func(*godog.ScenarioContext){
		"features/httpgateway/httpgateway.feature": httpgateway.InitializeScenario,
	}
)

func TestSuite(t *testing.T) {
	var failed bool
	for feature, scenarioContext := range features {
		err := testFeature(feature, scenarioContext)
		if err != nil {
			if godogStopOnFailure {
				t.Fatal(err)
			}

			failed = true
		}
	}

	if failed {
		t.Fatal("at least one step/scenario failed")
	}
}

func testFeature(feature string, scenarioInitializer func(*godog.ScenarioContext)) error {
	var testOutput io.Writer
	// default output is stdout
	testOutput = os.Stdout

	if godogFormat == "cucumber" {
		rf := path.Join(godogOutput, fmt.Sprintf("%v-report.json", filepath.Base(feature)))
		file, err := os.Create(rf)
		if err != nil {
			return fmt.Errorf("error creating report file %v: %w", rf, err)
		}

		defer file.Close()

		writer := bufio.NewWriter(file)
		defer writer.Flush()

		testOutput = writer
	}

	opts := godog.Options{
		Format:        godogFormat,
		Paths:         []string{feature},
		Tags:          godogTags,
		StopOnFailure: godogStopOnFailure,
		NoColors:      godogNoColors,
		Output:        testOutput,
		Concurrency:   1, // do not run tests concurrently
	}

	exitCode := godog.TestSuite{
		Name:                "conformance",
		ScenarioInitializer: scenarioInitializer,
		Options:             &opts,
	}.Run()
	if exitCode > 0 {
		return fmt.Errorf("unexpected exit code testing %v: %v", feature, exitCode)
	}

	return nil
}

func handleSignals() {
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)
	<-signals

	if err := utils.CleanupNamespaces(); err != nil {
		klog.Fatalf("error deleting temporal namespaces: %v", err)
	}

	os.Exit(1)
}
