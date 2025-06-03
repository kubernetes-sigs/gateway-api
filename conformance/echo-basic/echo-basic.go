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
	"encoding/json"
	"encoding/pem"
	"fmt"
	"io"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
	"golang.org/x/net/websocket"

	g "sigs.k8s.io/gateway-api/conformance/echo-basic/grpc"
)

type preserveSlashes struct {
	mux http.Handler
}

func (s *preserveSlashes) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	r.URL.Path = strings.ReplaceAll(r.URL.Path, "//", "/")
	s.mux.ServeHTTP(w, r)
}

var context Context

// RequestAssertions contains information about the request and the Ingress.
type RequestAssertions struct {
	Path    string              `json:"path"`
	Host    string              `json:"host"`
	Method  string              `json:"method"`
	Proto   string              `json:"proto"`
	Headers map[string][]string `json:"headers"`

	Context `json:",inline"`

	TLS *TLSAssertions `json:"tls,omitempty"`
	SNI string         `json:"sni"`
}

// TLSAssertions contains information about the TLS connection.
type TLSAssertions struct {
	Version          string   `json:"version"`
	PeerCertificates []string `json:"peerCertificates,omitempty"`
	// ServerName is the name sent from the peer using SNI.
	ServerName         string `json:"serverName"`
	NegotiatedProtocol string `json:"negotiatedProtocol,omitempty"`
	CipherSuite        string `json:"cipherSuite"`
}

// Context contains information about the context where the echoserver is running.
type Context struct {
	Namespace string `json:"namespace"`
	Ingress   string `json:"ingress"`
	Service   string `json:"service"`
	Pod       string `json:"pod"`
}

func main() {
	if os.Getenv("GRPC_ECHO_SERVER") != "" {
		g.Main()
		return
	}

	httpPort := os.Getenv("HTTP_PORT")
	if httpPort == "" {
		httpPort = "3000"
	}

	h2cPort := os.Getenv("H2C_PORT")
	if h2cPort == "" {
		h2cPort = "3001"
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
	httpMux.Handle("/ws", websocket.Handler(wsHandler))
	httpHandler := &preserveSlashes{httpMux}

	errchan := make(chan error)

	go func() {
		fmt.Printf("Starting server, listening on port %s (http)\n", httpPort)
		err := http.ListenAndServe(fmt.Sprintf(":%s", httpPort), httpHandler) //nolint:gosec
		if err != nil {
			errchan <- err
		}
	}()

	go runH2CServer(h2cPort, errchan)

	// Enable HTTPS if server certificate and private key are given. (TLS_SERVER_CERT, TLS_SERVER_PRIVKEY)
	if os.Getenv("TLS_SERVER_CERT") != "" && os.Getenv("TLS_SERVER_PRIVKEY") != "" {
		go func() {
			fmt.Printf("Starting server, listening on port %s (https)\n", httpsPort)
			err := listenAndServeTLS(fmt.Sprintf(":%s", httpsPort), os.Getenv("TLS_SERVER_CERT"), os.Getenv("TLS_SERVER_PRIVKEY"), httpHandler)
			if err != nil {
				errchan <- err
			}
		}()
	}

	// Enable secure backend if CA certificate is given. (CA_CERT)
	if os.Getenv("CA_CERT") != "" {
		// Start the backend server and listen on port 9443.
		go runBackendTLSServer("9443", errchan)
	}

	if err := <-errchan; err != nil {
		panic(fmt.Sprintf("Failed to start listening: %s\n", err.Error()))
	}
}

func wsHandler(ws *websocket.Conn) {
	fmt.Println("established websocket connection", ws.RemoteAddr())
	// Echo websocket frames from the connection back to the client
	// until io.EOF
	_, _ = io.Copy(ws, ws)
}

func healthHandler(w http.ResponseWriter, r *http.Request) { //nolint:revive
	w.WriteHeader(200)
	_, _ = w.Write([]byte(`OK`))
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

func runH2CServer(h2cPort string, errchan chan<- error) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.ProtoMajor != 2 && r.Header.Get("Upgrade") != "h2c" {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprint(w, "Expected h2c request")
			return
		}

		echoHandler(w, r)
	})
	h2c := &http.Server{
		ReadHeaderTimeout: time.Second,
		Addr:              fmt.Sprintf(":%s", h2cPort),
		Handler:           h2c.NewHandler(handler, &http2.Server{}),
	}
	fmt.Printf("Starting server, listening on port %s (h2c)\n", h2cPort)
	err := h2c.ListenAndServe()
	if err != nil {
		errchan <- err
	}
}

// Channel variable to store the SNI retrieved in runBackendTLS handler func.
var sniChannel = make(chan string)

func runBackendTLSServer(port string, errchan chan<- error) {
	// This handler function runs within the backend server to find the SNI
	// and return it in the RequestAssertions.
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.RequestURI, "backendTLS") {
			// Find the sni stored in the channel.
			sni := <-sniChannel
			if sni == "" {
				err := fmt.Errorf("error finding SNI: SNI is empty")
				// If there are some test cases without SNI, then they must handle this error properly.
				processError(w, err, http.StatusBadRequest)
			}
			requestAssertions := RequestAssertions{
				r.RequestURI,
				r.Host,
				r.Method,
				r.Proto,
				r.Header,

				context,

				tlsStateToAssertions(r.TLS),
				sni,
			}
			processRequestAssertions(requestAssertions, w, r)
		} else {
			// This should never happen, but just in case.
			processError(w, fmt.Errorf("backend server called without correct uri"), http.StatusBadRequest)
		}
	})

	config, err := makeTLSConfig(os.Getenv("CA_CERT"))
	if err != nil {
		errchan <- err
	}
	btlsServer := &http.Server{
		Addr:              fmt.Sprintf(":%s", port),
		Handler:           handler,
		ReadHeaderTimeout: time.Second,
		TLSConfig:         config,
	}
	fmt.Printf("Starting server, listening on port %s (btls)\n", port)
	err = btlsServer.ListenAndServeTLS(os.Getenv("CA_CERT"), os.Getenv("CA_CERT_KEY"))
	if err != nil {
		fmt.Printf("Failed to start server: %v\n", err)
		errchan <- err
	}
}

func makeTLSConfig(cacert string) (*tls.Config, error) {
	var config tls.Config

	if cacert == "" {
		return &config, fmt.Errorf("empty CA cert specified")
	}
	cert, err := tls.LoadX509KeyPair(cacert, os.Getenv("CA_CERT_KEY"))
	if err != nil {
		return &config, fmt.Errorf("failed to load key pair: %v", err)
	}
	certs := []tls.Certificate{cert}

	// Verify certificate against given CA but also allow unauthenticated connections.
	config.ClientAuth = tls.VerifyClientCertIfGiven
	config.Certificates = certs
	config.GetConfigForClient = func(info *tls.ClientHelloInfo) (*tls.Config, error) {
		if info != nil {
			// Store the SNI from the ClientHello in the sniChannel.
			if info.ServerName == "" {
				return nil, fmt.Errorf("no SNI specified")
			}
			sniChannel <- info.ServerName
			return nil, nil
		}
		return nil, fmt.Errorf("no client hello available")
	}

	return &config, nil
}

func echoHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("Echoing back request made to %s to client (%s)\n", r.RequestURI, r.RemoteAddr)

	// If the request has form ?delay=[:duration] wait for duration
	// For example, ?delay=10s will cause the response to wait 10s before responding
	err := delayResponse(r)
	if err != nil {
		fmt.Printf("error : %v\n", err)
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
		"",
	}
	processRequestAssertions(requestAssertions, w, r)
}

func processRequestAssertions(requestAssertions RequestAssertions, w http.ResponseWriter, r *http.Request) {
	js, err := json.MarshalIndent(requestAssertions, "", " ")
	if err != nil {
		processError(w, err, http.StatusInternalServerError)
		return
	}

	writeEchoResponseHeaders(w, r.Header)
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	_, _ = w.Write(js)
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

func processError(w http.ResponseWriter, err error, code int) { //nolint:unparam
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
	_, _ = w.Write(body)
}

func listenAndServeTLS(addr string, serverCert, serverPrivKey string, handler http.Handler) error {
	srv := &http.Server{ //nolint:gosec
		Addr:    addr,
		Handler: handler,
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
			err := pem.Encode(&out, &pem.Block{
				Type:  "CERTIFICATE",
				Bytes: c.Raw,
			})
			if err != nil {
				fmt.Printf("failed to encode certificate: %v\n", err)
			} else {
				state.PeerCertificates = append(state.PeerCertificates, out.String())
			}
		}

		return &state
	}

	return nil
}
