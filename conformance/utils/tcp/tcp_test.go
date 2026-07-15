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

package tcp

import (
	"bufio"
	"context"
	"encoding/json"
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	tcpserver "sigs.k8s.io/gateway-api/conformance/echo-basic/tcpserver"
)

func TestWaitForValidTCPResponseRetriesStaleResponse(t *testing.T) {
	listener, err := (&net.ListenConfig{}).Listen(context.Background(), "tcp", "127.0.0.1:0")
	require.NoError(t, err)
	t.Cleanup(func() { listener.Close() }) //nolint:errcheck

	serverDone := make(chan struct{})
	go func() {
		defer close(serverDone)
		for i := range 2 {
			conn, err := listener.Accept()
			if err != nil {
				return
			}
			handleTCPTestConnection(conn, i == 0)
		}
	}()

	WaitForValidTCPResponse(t, &net.Dialer{}, listener.Addr().String(), ExpectedResponse{
		Backend:   "backend",
		Namespace: "test-ns",
	}, 3*time.Second)

	select {
	case <-serverDone:
	case <-time.After(time.Second):
		t.Fatal("TCP test server did not receive the retry")
	}
}

func handleTCPTestConnection(conn net.Conn, staleWelcome bool) {
	defer conn.Close() //nolint:errcheck

	welcome := tcpserver.WelcomeMessage
	if staleWelcome {
		welcome = "stale response\n"
	}
	if _, err := conn.Write([]byte(welcome)); err != nil {
		return
	}

	reader := bufio.NewReader(conn)
	if _, err := reader.ReadString('\n'); err != nil {
		return
	}
	if _, err := conn.Write([]byte("PONG\n")); err != nil {
		return
	}
	if _, err := reader.ReadString('\n'); err != nil {
		return
	}
	if _, err := conn.Write([]byte("false\n")); err != nil {
		return
	}
	if _, err := reader.ReadString('\n'); err != nil {
		return
	}
	payload, err := json.Marshal(tcpserver.TCPAssertions{Context: tcpserver.Context{Namespace: "test-ns", Pod: "backend-pod"}})
	if err != nil {
		return
	}
	_, _ = conn.Write(append(payload, '\n'))
}
