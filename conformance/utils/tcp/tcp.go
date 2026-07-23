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
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"net"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/util/wait"

	tcpserver "sigs.k8s.io/gateway-api-conformance-images/echo-basic/tcpserver"
	"sigs.k8s.io/gateway-api/conformance/utils/config"
	"sigs.k8s.io/gateway-api/conformance/utils/tlog"
)

type Dialer interface {
	DialContext(ctx context.Context, network, address string) (net.Conn, error)
	Dial(network, address string) (net.Conn, error)
}

type ExpectedResponse struct {
	BackendIsTLS bool
	Backend      string // Backend will be asserted from podname.
	Namespace    string
	TLSProtocol  string // Optional, if set will validate and value should match the definitions from tls.VersionName()
	Hostname     string // Optional, if set will verify if the SNI hostname captured by TLS matches this value
}

// MakeTCPRequestAndExpectEventuallyValidResponse makes a TCP request with the given parameters,
// understanding that the request may fail for some amount of time.
//
// Once the request succeeds consistently with the response having the answer for IS_TLS and desired information
// of the backend, it will pass, otherwise fails
func MakeTCPRequestAndExpectEventuallyValidResponse(t *testing.T, timeoutConfig config.TimeoutConfig, gwAddr string, serverCertificate []byte, serverName string, useTLS bool, expected ExpectedResponse) {
	t.Helper()

	var tlsConfig *tls.Config

	if useTLS {
		tlsConfig = &tls.Config{
			ServerName: serverName,
			MinVersion: tls.VersionTLS12,
		}
		if len(serverCertificate) > 0 {
			certPool := x509.NewCertPool()
			require.True(t, certPool.AppendCertsFromPEM(serverCertificate))
			tlsConfig.RootCAs = certPool
		}
	}
	dialer := makeClient(tlsConfig)
	WaitForValidTCPResponse(t, dialer, gwAddr, expected, timeoutConfig.MaxTimeToConsistency)
}

// WaitForConsistentTCPResponse - repeats the provided request until it completes with a response having
// the expected response consistently. The provided threshold determines how many times in
// a row this must occur to be considered "consistent".
// For every request, a new dial will happen so the following process will be verified:
// - WelcomeMessage matches
// - IS_TLS message matches
// - TEST message matches assertions
func WaitForValidTCPResponse(t *testing.T, dialer Dialer, gwAddr string, expected ExpectedResponse, maxTimeToConsistency time.Duration) {
	t.Helper()
	retry := func(err error) bool {
		if err != nil {
			tlog.Logf(t, "an error occurred during assertion, will retry: %s", err)
		}
		return false
	}

	assert.Eventually(t, func() bool {
		client, err := dialer.DialContext(t.Context(), "tcp", gwAddr)
		if err != nil {
			tlog.Logf(t, "client could not connect: %s; retrying", err)
			return false
		}
		defer func() {
			if closeErr := client.Close(); closeErr != nil {
				tlog.Logf(t, "error closing TCP connection: %s", closeErr)
			}
		}()
		tlog.Logf(t, "tcp client connected")
		message, err := bufio.NewReader(client).ReadString('\n')
		if err != nil {
			return retry(err)
		}
		if message != tcpserver.WelcomeMessage {
			return retry(fmt.Errorf("TCP server welcome message does not match: got %q", message))
		}

		if _, err = fmt.Fprint(client, "PING\n"); err != nil {
			return retry(err)
		}
		message, err = bufio.NewReader(client).ReadString('\n')
		if err != nil {
			return retry(err)
		}
		if message != "PONG\n" {
			return retry(fmt.Errorf("TCP server PING response does not match: got %q", message))
		}

		if _, err = fmt.Fprint(client, "IS_TLS\n"); err != nil {
			return retry(err)
		}
		message, err = bufio.NewReader(client).ReadString('\n')
		if err != nil {
			return retry(err)
		}
		if actual := strings.TrimSuffix(message, "\n"); actual != fmt.Sprintf("%t", expected.BackendIsTLS) {
			return retry(fmt.Errorf("TCP server TLS response does not match: got %q", actual))
		}

		if _, err = fmt.Fprint(client, "TEST\n"); err != nil {
			return retry(err)
		}
		message, err = bufio.NewReader(client).ReadString('\n')
		if err != nil {
			return retry(err)
		}

		payload := &tcpserver.TCPAssertions{}
		if err := json.Unmarshal([]byte(message), payload); err != nil {
			return retry(err)
		}

		if err := validateTestMessage(payload, expected); err != nil {
			return retry(err)
		}
		return true
	}, maxTimeToConsistency, time.Second)

	tlog.Logf(t, "Request passed")
}

func makeClient(tlsConfig *tls.Config) Dialer {
	if tlsConfig == nil {
		return &net.Dialer{}
	}

	return &tls.Dialer{
		Config: tlsConfig,
	}
}

func validateTestMessage(payload *tcpserver.TCPAssertions, expected ExpectedResponse) error {
	if payload == nil {
		return fmt.Errorf("TCP response payload is nil")
	}
	if payload.Namespace != expected.Namespace {
		return fmt.Errorf("namespace does not match: got %q, want %q", payload.Namespace, expected.Namespace)
	}
	if !strings.HasPrefix(payload.Pod, fmt.Sprintf("%s-", expected.Backend)) {
		return fmt.Errorf("backend name does not match with pod prefix: pod=%q backend=%q", payload.Pod, expected.Backend)
	}
	if expected.Hostname != "" {
		if payload.TLSAssertion == nil {
			return fmt.Errorf("TLS server name does not match: no TLS assertions received, want %q", expected.Hostname)
		}
		if payload.TLSAssertion.ServerName != expected.Hostname {
			return fmt.Errorf("TLS server name does not match: got %q, want %q", payload.TLSAssertion.ServerName, expected.Hostname)
		}
	}
	if expected.TLSProtocol != "" {
		if payload.TLSAssertion == nil {
			return fmt.Errorf("TLS protocol does not match: no TLS assertions received, want %q", expected.TLSProtocol)
		}
		if payload.TLSAssertion.NegotiatedProtocol != expected.TLSProtocol {
			return fmt.Errorf("TLS protocol does not match: got %q, want %q", payload.TLSAssertion.NegotiatedProtocol, expected.TLSProtocol)
		}
	}
	return nil
}

// EchoSendOnce opens a single TCP connection to gwAddr, performs the
// tcpserver TEST handshake, and returns the pod name from the JSON envelope.
// It is intended for tests that need to attribute a single response to a
// specific backend Pod (for example, weighted routing).
func EchoSendOnce(ctx context.Context, gwAddr string, timeout time.Duration) (string, error) {
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

// ExpectAddressBeAvailable polls until a TCP connection to the provided address
// can be established, or fails the test if the timeout expires. It only verifies
// that the address accepts TCP connections; it does not validate an echo response
// or backend selection.
func ExpectAddressBeAvailable(t *testing.T, interval, timeout time.Duration, address string) {
	t.Helper()

	tlog.Logf(t, "performing TCP connection probe on %s", address)
	err := wait.PollUntilContextTimeout(t.Context(), interval, timeout, true,
		func(ctx context.Context) (bool, error) {
			var dialer net.Dialer
			conn, err := dialer.DialContext(ctx, "tcp", address)
			if err != nil {
				tlog.Logf(t, "failed to establish TCP connection to %s; retrying: %v", address, err)
				return false, nil
			}
			tlog.Logf(t, "established TCP connection to %s", address)
			if err := conn.Close(); err != nil {
				tlog.Logf(t, "failed to close TCP probe connection to %s: %v", address, err)
			}

			return true, nil
		})
	require.NoError(t, err, "failed waiting for TCP connection to %s after %v", address, timeout)
}
