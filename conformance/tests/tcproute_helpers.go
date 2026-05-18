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

package tests

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"net"
	"strings"
	"testing"
	"time"

	"k8s.io/apimachinery/pkg/util/wait"

	"sigs.k8s.io/gateway-api/conformance/echo-basic/tcpserver"
	"sigs.k8s.io/gateway-api/conformance/utils/tlog"
)

// expectTCPEchoResponse polls until a TCP echo round-trip against the given
// gateway address succeeds, or the timeout is exceeded. It speaks the
// tcpserver protocol: read the welcome banner and validate a PING/PONG
// exchange to confirm end-to-end TCP routing through the Gateway.
func expectTCPEchoResponse(t *testing.T, timeout time.Duration, gwAddr string) {
	t.Helper()

	tlog.Logf(t, "performing TCP echo probe on %s", gwAddr)
	err := wait.PollUntilContextTimeout(context.TODO(), time.Second, timeout, true,
		func(ctx context.Context) (bool, error) {
			var dialer net.Dialer
			conn, err := dialer.DialContext(ctx, "tcp", gwAddr)
			if err != nil {
				tlog.Logf(t, "failed to dial TCP %s: %v", gwAddr, err)
				return false, nil
			}
			defer conn.Close()

			if err = conn.SetDeadline(time.Now().Add(2 * time.Second)); err != nil {
				return false, fmt.Errorf("setting TCP deadline: %w", err)
			}

			reader := bufio.NewReader(conn)
			welcome, err := reader.ReadString('\n')
			if err != nil {
				tlog.Logf(t, "failed to read TCP welcome message: %v", err)
				return false, nil
			}
			if welcome != tcpserver.WelcomeMessage {
				tlog.Logf(t, "unexpected TCP welcome message: %q", welcome)
				return false, nil
			}

			if _, err = fmt.Fprint(conn, "PING\n"); err != nil {
				tlog.Logf(t, "failed to write PING: %v", err)
				return false, nil
			}
			pong, err := reader.ReadString('\n')
			if err != nil {
				tlog.Logf(t, "failed to read PONG: %v", err)
				return false, nil
			}
			if strings.TrimSpace(pong) != "PONG" {
				tlog.Logf(t, "unexpected response to PING: %q", pong)
				return false, nil
			}
			tlog.Logf(t, "got TCP PONG response from %s", gwAddr)
			return true, nil
		})
	if err != nil {
		t.Errorf("TCP echo probe never succeeded against %s: %v", gwAddr, err)
	}
}

// tcpEchoSendOnce opens a single TCP connection to gwAddr, performs the
// tcpserver TEST handshake, and returns the pod name from the JSON envelope.
func tcpEchoSendOnce(ctx context.Context, gwAddr string, timeout time.Duration) (string, error) {
	var dialer net.Dialer
	conn, err := dialer.DialContext(ctx, "tcp", gwAddr)
	if err != nil {
		return "", fmt.Errorf("dialing TCP %s: %w", gwAddr, err)
	}
	defer conn.Close()

	if err = conn.SetDeadline(time.Now().Add(timeout)); err != nil {
		return "", fmt.Errorf("setting TCP deadline: %w", err)
	}

	reader := bufio.NewReader(conn)
	welcome, err := reader.ReadString('\n')
	if err != nil {
		return "", fmt.Errorf("reading TCP welcome message: %w", err)
	}
	if welcome != tcpserver.WelcomeMessage {
		return "", fmt.Errorf("unexpected TCP welcome message: %q", welcome)
	}

	if _, err = fmt.Fprint(conn, "TEST\n"); err != nil {
		return "", fmt.Errorf("writing TEST: %w", err)
	}

	line, err := reader.ReadString('\n')
	if err != nil {
		return "", fmt.Errorf("reading TEST response: %w", err)
	}

	var resp tcpserver.TCPAssertions
	if err := json.Unmarshal([]byte(line), &resp); err != nil {
		return "", fmt.Errorf("decoding TCP echo response %q: %w", line, err)
	}
	if resp.Pod == "" {
		return "", fmt.Errorf("TCP echo response missing pod name: %q", line)
	}
	return resp.Pod, nil
}
