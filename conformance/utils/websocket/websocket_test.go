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

package websocket

import (
	"net/http/httptest"
	"strings"
	"testing"

	"golang.org/x/net/websocket"
)

// TestDefaultDialerDial verifies that DefaultDialer performs a real WebSocket
// handshake and round-trips a message — i.e. that the default injection point
// preserves the historical websocket.Dial behavior the suite relied on before
// the dialer was made overridable.
func TestDefaultDialerDial(t *testing.T) {
	server := httptest.NewServer(websocket.Handler(func(ws *websocket.Conn) {
		var msg string
		if err := websocket.Message.Receive(ws, &msg); err != nil {
			return
		}
		_ = websocket.Message.Send(ws, msg)
	}))
	defer server.Close()

	url := strings.Replace(server.URL, "http://", "ws://", 1) + "/"

	conn, err := (&DefaultDialer{}).Dial(url, "", "http://example.com/")
	if err != nil {
		t.Fatalf("DefaultDialer.Dial(%q) returned error: %v", url, err)
	}
	defer conn.Close()

	const want = "websocket round-trip"
	if err := websocket.Message.Send(conn, want); err != nil {
		t.Fatalf("failed to send message: %v", err)
	}

	var got string
	if err := websocket.Message.Receive(conn, &got); err != nil {
		t.Fatalf("failed to receive message: %v", err)
	}

	if got != want {
		t.Fatalf("unexpected echo: want %q, got %q", want, got)
	}
}
