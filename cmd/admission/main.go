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
	"context"
	"crypto/tls"
	"errors"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"k8s.io/klog/v2"

	"sigs.k8s.io/gateway-api/pkg/admission"
)

var (
	tlsCertFilePath, tlsKeyFilePath string
	showVersion, help               bool
)

var (
	VERSION = "dev"
	COMMIT  = "dev"
)

func main() {
	flag.StringVar(&tlsCertFilePath, "tlsCertFile", "/etc/certs/tls.crt", "File with x509 certificate")
	flag.StringVar(&tlsKeyFilePath, "tlsKeyFile", "/etc/certs/tls.key", "File with private key to tlsCertFile")
	flag.BoolVar(&showVersion, "version", false, "Show release version and exit")
	flag.BoolVar(&help, "help", false, "Show flag defaults and exit")
	klog.InitFlags(nil)
	flag.Parse()

	if showVersion {
		printVersion()
		os.Exit(0)
	}

	if help {
		printVersion()
		flag.PrintDefaults()
		os.Exit(0)
	}

	printVersion()

	certs, err := tls.LoadX509KeyPair(tlsCertFilePath, tlsKeyFilePath)
	if err != nil {
		klog.Fatalf("failed to load TLS cert-key for admission-webhook-server: %v", err)
	}

	server := &http.Server{
		Addr:              ":8443",
		ReadHeaderTimeout: 10 * time.Second, // for Potential Slowloris Attack (G112)
		// Require at least TLS12 to satisfy golint G402.
		TLSConfig: &tls.Config{MinVersion: tls.VersionTLS12, Certificates: []tls.Certificate{certs}},
	}
	mux := http.NewServeMux()
	mux.HandleFunc("/validate", admission.ServeHTTP)
	server.Handler = mux

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		err := server.ListenAndServeTLS("", "")
		if errors.Is(err, http.ErrServerClosed) {
			klog.Fatalf("admission-webhook-server stopped: %v", err)
		}
	}()
	klog.Info("admission webhook server started and listening on :8443")

	// gracefully shutdown
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)
	<-signalChan

	klog.Info("admission webhook received kill signal")
	if err := server.Shutdown(context.Background()); err != nil {
		klog.Fatalf("server shutdown failed:%+v", err)
	}
	wg.Wait()
}

func printVersion() {
	fmt.Printf("gateway-api-admission-webhook version: %v (%v)\n", VERSION, COMMIT)
}
