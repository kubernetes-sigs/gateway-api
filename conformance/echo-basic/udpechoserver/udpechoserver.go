/*
Copyright The Kubernetes Authors.

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

// Package udpechoserver provides a basic echo server used by Gateway API
// conformance tests. It listens on the configured port for both UDP and TCP
// (so a Gateway with mixed UDP/TCP listeners on the same port can target a
// single backend Service) and replies with a JSON envelope of the form:
//
//	{
//	  "request":   "<original request body>",
//	  "namespace": "<NAMESPACE env var>",
//	  "ingress":   "<INGRESS_NAME env var>",
//	  "service":   "<SERVICE_NAME env var>",
//	  "pod":       "<POD_NAME env var>"
//	}
//
// The pod context lets tests with multiple weighted backends distinguish
// replicas. Any field is returned as an empty string when its env var is
// unset.
package udpechoserver

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"net"
	"os"
)

// Context contains information about the pod where the udpechoserver is
// running. It mirrors the fields populated by echo-basic for HTTP responses
// so tests can identify the responding replica.
type Context struct {
	Namespace string `json:"namespace"`
	Ingress   string `json:"ingress"`
	Service   string `json:"service"`
	Pod       string `json:"pod"`
}

// EchoResponse is the JSON envelope this server returns for every datagram or
// TCP message it receives.
type EchoResponse struct {
	Request string `json:"request"`
	Context `json:",inline"`
}

// Main starts the UDP and TCP echo servers on the configured port (UDP_PORT
// env var, default "8080") and runs until a fatal error occurs. It is
// intended to be invoked from echo-basic when UDP_ECHO_SERVER is set.
func Main() {
	port := os.Getenv("UDP_PORT")
	if port == "" {
		port = "8080"
	}
	podContext := Context{
		Namespace: os.Getenv("NAMESPACE"),
		Ingress:   os.Getenv("INGRESS_NAME"),
		Service:   os.Getenv("SERVICE_NAME"),
		Pod:       os.Getenv("POD_NAME"),
	}

	errCh := make(chan error, 2)
	go func() { errCh <- serveUDP(port, podContext) }()
	go func() { errCh <- serveTCP(port, podContext) }()

	if err := <-errCh; err != nil {
		fmt.Println("echo server error:", err)
		os.Exit(1)
	}
}

// serveUDP listens on the given UDP port and replies to each datagram with a
// JSON-encoded EchoResponse.
func serveUDP(port string, podContext Context) error {
	addr, err := net.ResolveUDPAddr("udp", ":"+port)
	if err != nil {
		return fmt.Errorf("resolving UDP address: %w", err)
	}

	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		return fmt.Errorf("listening UDP: %w", err)
	}
	defer conn.Close()

	fmt.Printf("UDP server listening on :%s with context: %+v\n", port, podContext)

	buffer := make([]byte, 1024)
	for {
		n, remoteAddr, err := conn.ReadFromUDP(buffer)
		if err != nil {
			fmt.Println("Error reading UDP:", err)
			continue
		}
		fmt.Printf("Received UDP %s from %s\n", string(buffer[:n]), remoteAddr)

		payload, err := json.Marshal(EchoResponse{
			Request: string(buffer[:n]),
			Context: podContext,
		})
		if err != nil {
			fmt.Println("Error marshaling UDP response:", err)
			continue
		}

		if _, err := conn.WriteToUDP(payload, remoteAddr); err != nil {
			fmt.Println("Error writing UDP:", err)
		}
	}
}

// serveTCP listens on the given TCP port and replies to each connection's
// first line of input with a JSON-encoded EchoResponse.
func serveTCP(port string, podContext Context) error {
	var lc net.ListenConfig
	listener, err := lc.Listen(context.Background(), "tcp", ":"+port)
	if err != nil {
		return fmt.Errorf("listening TCP: %w", err)
	}
	defer listener.Close()

	fmt.Printf("TCP server listening on :%s with context: %+v\n", port, podContext)

	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("Error accepting TCP:", err)
			continue
		}
		go handleTCP(conn, podContext)
	}
}

// handleTCP reads the first line of input from the connection and writes back
// a JSON-encoded EchoResponse. The connection is closed when the response has
// been sent or an error occurs.
func handleTCP(conn net.Conn, podContext Context) {
	defer conn.Close()

	reader := bufio.NewReader(conn)
	line, err := reader.ReadString('\n')
	// ReadString returns whatever it managed to read alongside io.EOF when the
	// peer closes the connection without a trailing newline. Echo what we got
	// instead of failing in that case.
	if err != nil && len(line) == 0 {
		fmt.Println("Error reading TCP:", err)
		return
	}
	fmt.Printf("Received TCP %q from %s\n", line, conn.RemoteAddr())

	payload, err := json.Marshal(EchoResponse{
		Request: line,
		Context: podContext,
	})
	if err != nil {
		fmt.Println("Error marshaling TCP response:", err)
		return
	}

	if _, err := conn.Write(payload); err != nil {
		fmt.Println("Error writing TCP:", err)
	}
}
