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
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// RequestAssertions contains information about the request and the Ingress
type RequestAssertions struct {
	Path    string              `json:"path"`
	Host    string              `json:"host"`
	Method  string              `json:"method"`
	Proto   string              `json:"proto"`
	Headers map[string][]string `json:"headers"`

	Context `json:",inline"`

	TLS *TLSAssertions `json:"tls,omitempty"`
}

// TLSAssertions contains information about the TLS connection.
type TLSAssertions struct {
	Version            string   `json:"version"`
	PeerCertificates   []string `json:"peerCertificates,omitempty"`
	ServerName         string   `json:"serverName"`
	NegotiatedProtocol string   `json:"negotiatedProtocol,omitempty"`
	CipherSuite        string   `json:"cipherSuite"`
}

type preserveSlashes struct {
	mux http.Handler
}

func (s *preserveSlashes) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	r.URL.Path = strings.Replace(r.URL.Path, "//", "/", -1)
	s.mux.ServeHTTP(w, r)
}

// Context contains information about the context where the echoserver is running
type Context struct {
	Namespace string `json:"namespace"`
	Ingress   string `json:"ingress"`
	Service   string `json:"service"`
	Pod       string `json:"pod"`
}

var context Context

func main() {
	httpPort := os.Getenv("HTTP_PORT")
	if httpPort == "" {
		httpPort = "3000"
	}

	httpsPort := os.Getenv("HTTPS_PORT")
	if httpsPort == "" {
		httpsPort = "8443"
	}

	context = Context{
		Namespace: os.Getenv("NAMESPACE"),
		Ingress:   os.Getenv("INGRESS_NAME"),
		Service:   os.Getenv("SERVICE_NAME"),
		Pod:       os.Getenv("POD_NAME"),
	}

	httpMux := http.NewServeMux()
	httpMux.HandleFunc("/health", healthHandler)
	httpMux.HandleFunc("/status/", statusHandler)
	httpMux.HandleFunc("/", echoHandler)
	httpHandler := &preserveSlashes{httpMux}

	errchan := make(chan error)

	go func() {
		fmt.Printf("Starting server, listening on port %s (http)\n", httpPort)
		err := http.ListenAndServe(fmt.Sprintf(":%s", httpPort), httpHandler)
		if err != nil {
			errchan <- err
		}
	}()

	// Enable HTTPS if certificate and private key are given.
	if os.Getenv("TLS_SERVER_CERT") != "" && os.Getenv("TLS_SERVER_PRIVKEY") != "" {
		go func() {
			fmt.Printf("Starting server, listening on port %s (https)\n", httpsPort)
			err := listenAndServeTLS(fmt.Sprintf(":%s", httpsPort), os.Getenv("TLS_SERVER_CERT"), os.Getenv("TLS_SERVER_PRIVKEY"), os.Getenv("TLS_CLIENT_CACERTS"), httpHandler)
			if err != nil {
				errchan <- err
			}
		}()
	}

	if err := <-errchan; err != nil {
		panic(fmt.Sprintf("Failed to start listening: %s\n", err.Error()))
	}
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(200)
	w.Write([]byte(`OK`))
}

func statusHandler(w http.ResponseWriter, r *http.Request) {
	code := http.StatusBadRequest

	re := regexp.MustCompile(`^/status/(\d\d\d)$`)
	match := re.FindStringSubmatch(r.RequestURI)
	if match != nil {
		code, _ = strconv.Atoi(match[1])
	}

	w.WriteHeader(code)
}

func delayResponse(request *http.Request) error {
	d := request.FormValue("delay")
	if len(d) == 0 {
		return nil
	}

	t, err := time.ParseDuration(d)
	if err != nil {
		return err
	}
	time.Sleep(t)
	return nil
}

func echoHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("Echoing back request made to %s to client (%s)\n", r.RequestURI, r.RemoteAddr)

	// If the request has form ?delay=[:duration] wait for duration
	// For example, ?delay=10s will cause the response to wait 10s before responding
	if err := delayResponse(r); err != nil {
		processError(w, err, http.StatusInternalServerError)
		return
	}

	requestAssertions := RequestAssertions{
		r.RequestURI,
		r.Host,
		r.Method,
		r.Proto,
		r.Header,

		context,

		tlsStateToAssertions(r.TLS),
	}

	js, err := json.MarshalIndent(requestAssertions, "", " ")
	if err != nil {
		processError(w, err, http.StatusInternalServerError)
		return
	}

	writeEchoResponseHeaders(w, r.Header)
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Write(js)
}

func writeEchoResponseHeaders(w http.ResponseWriter, headers http.Header) {
	for _, headerKVList := range headers["X-Echo-Set-Header"] {
		headerKVs := strings.Split(headerKVList, ",")
		for _, headerKV := range headerKVs {
			name, value, _ := strings.Cut(strings.TrimSpace(headerKV), ":")
			// Add directly to the map to preserve casing.
			if len(w.Header()[name]) == 0 {
				w.Header()[name] = []string{value}
			} else {
				w.Header()[name][0] += "," + strings.TrimSpace(value)
			}
		}
	}
}

func processError(w http.ResponseWriter, err error, code int) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	body, err := json.Marshal(struct {
		Message string `json:"message"`
	}{
		err.Error(),
	})
	if err != nil {
		w.WriteHeader(code)
		fmt.Fprintln(w, err)
		return
	}

	w.WriteHeader(code)
	w.Write(body)
}

func listenAndServeTLS(addr string, serverCert string, serverPrivKey string, clientCA string, handler http.Handler) error {
	var config tls.Config

	// Optionally enable client certificate validation when client CA certificates are given.
	if clientCA != "" {
		ca, err := os.ReadFile(clientCA)
		if err != nil {
			return err
		}

		certPool := x509.NewCertPool()
		if ok := certPool.AppendCertsFromPEM(ca); !ok {
			return fmt.Errorf("unable to append certificate in %q to CA pool", clientCA)
		}

		// Verify certificate against given CA but also allow unauthenticated connections.
		config.ClientAuth = tls.VerifyClientCertIfGiven
		config.ClientCAs = certPool
	}

	srv := &http.Server{
		Addr:      addr,
		Handler:   handler,
		TLSConfig: &config,
	}

	return srv.ListenAndServeTLS(serverCert, serverPrivKey)
}

func tlsStateToAssertions(connectionState *tls.ConnectionState) *TLSAssertions {
	if connectionState != nil {
		var state TLSAssertions

		switch connectionState.Version {
		case tls.VersionTLS13:
			state.Version = "TLSv1.3"
		case tls.VersionTLS12:
			state.Version = "TLSv1.2"
		case tls.VersionTLS11:
			state.Version = "TLSv1.1"
		case tls.VersionTLS10:
			state.Version = "TLSv1.0"
		}

		state.NegotiatedProtocol = connectionState.NegotiatedProtocol
		state.ServerName = connectionState.ServerName
		state.CipherSuite = tls.CipherSuiteName(connectionState.CipherSuite)

		// Convert peer certificates to PEM blocks.
		for _, c := range connectionState.PeerCertificates {
			var out strings.Builder
			pem.Encode(&out, &pem.Block{
				Type:  "CERTIFICATE",
				Bytes: c.Raw,
			})
			state.PeerCertificates = append(state.PeerCertificates, out.String())
		}

		return &state
	}

	return nil
}
