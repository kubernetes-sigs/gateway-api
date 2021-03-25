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
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/golang/glog"
	"k8s.io/klog/v2"

	"sigs.k8s.io/gateway-api/pkg/admission"
)

var (
	tlsCert, tlsKey   string
	showVersion, help bool
)

const appVersion = "0.0.1"

func main() {
	flag.StringVar(&tlsCert, "tlsCertFile", "/etc/certs/tls.crt", "File with x509 certificate")
	flag.StringVar(&tlsKey, "tlsKeyFile", "/etc/certs/tls.key", "File with private key to tlsCertFile")
	flag.BoolVar(&showVersion, "version", false, "Show release version and exit")
	flag.BoolVar(&help, "help", false, "Show flag defaults and exit")
	flag.Parse()

	if showVersion {
		fmt.Printf("gateway-api-admission version: %v\n", appVersion)
		os.Exit(0)
	}

	if help {
		fmt.Printf("gateway-api-admission version: %v\n", appVersion)
		flag.PrintDefaults()
		os.Exit(0)
	}

	klog.Infof("gateway-api-admission webhook version: %v", appVersion)

	certs, err := tls.LoadX509KeyPair(tlsCert, tlsKey)
	if err != nil {
		klog.Fatalf("failed to load admission webhook keypair with err: %v", err)
	}

	server := &http.Server{
		Addr: ":8443",
		// Require at least TLS12 to satisfy golint G402.
		TLSConfig: &tls.Config{MinVersion: tls.VersionTLS12, Certificates: []tls.Certificate{certs}},
	}
	mux := http.NewServeMux()
	mux.HandleFunc("/validate", admission.ServeHTTP)
	server.Handler = mux

	go func() {
		err := server.ListenAndServeTLS("", "")
		klog.Fatalf("admission webhook server stopped with err: %v", err)
	}()

	glog.Info("admission webhook server started and listening on :8443")

	// gracefully shutdown
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)
	<-signalChan

	glog.Info("admission webhook received kill signal, shutdown handled gracefully")
	if err := server.Shutdown(context.Background()); err != nil {
		klog.Fatalf("server shutdown failed:%+v", err)
	}
}
