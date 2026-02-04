/*
Copyright 2026 The Kubernetes Authors.

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

package tcpserver

import (
	"bufio"
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"math/big"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	namespace = "test-ns"
	ingress   = "test-ingress"
	service   = "test-service"
	pod       = "test-pod"
)

func TestTCPEchoServer(t *testing.T) {
	t.Setenv("NAMESPACE", namespace)
	t.Setenv("INGRESS_NAME", ingress)
	t.Setenv("SERVICE_NAME", service)
	t.Setenv("POD_NAME", pod)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	t.Run("with TCP Server only when there is no TLS certificate", func(t *testing.T) {
		tcpPort, err := getFreePort(ctx)
		require.NoError(t, err, "error allocating TCP Port for test")

		t.Logf("Using %d as TCP Port", tcpPort)
		t.Setenv("TCP_PORT", strconv.Itoa(tcpPort))

		tlsPort, err := getFreePort(ctx)
		require.NoError(t, err, "error allocating TLS Port for test")

		t.Logf("Using %d as TLS Port", tlsPort)
		t.Setenv("TLS_PORT", strconv.Itoa(tlsPort))
		errCh := make(chan error)
		go runServer(ctx, errCh)
		t.Cleanup(func() { close(errCh) })
		waitForListener(ctx, t, strconv.Itoa(tcpPort))

		dialer := &net.Dialer{}
		_, err = dialer.DialContext(ctx, "tcp", fmt.Sprintf("localhost:%s", strconv.Itoa(tlsPort)))
		require.Error(t, err, "trying to connect to an echo server without TLS enabled should fail")

		tcpClient, err := dialer.DialContext(ctx, "tcp", fmt.Sprintf("localhost:%s", strconv.Itoa(tcpPort)))
		require.NoError(t, err, "got an error while trying to connect to tcpserver")

		t.Cleanup(func() {
			tcpClient.Close() //nolint: errcheck
		})
		t.Run("check against the TCP server", func(t *testing.T) {
			message, err := bufio.NewReader(tcpClient).ReadString('\n')
			require.NoError(t, err, "error reading message response")
			assert.Equal(t, "Gateway API Test TCP Server\n", message)

			fmt.Fprintf(tcpClient, "PING\n")
			message, err = bufio.NewReader(tcpClient).ReadString('\n')
			require.NoError(t, err, "error reading message response")
			assert.Equal(t, "PONG\n", message)

			fmt.Fprintf(tcpClient, "IS_TLS\n")
			message, err = bufio.NewReader(tcpClient).ReadString('\n')
			require.NoError(t, err, "error reading message response")
			assert.Equal(t, "false\n", message)

			fmt.Fprintf(tcpClient, "TEST\n")
			message, err = bufio.NewReader(tcpClient).ReadString('\n')
			require.NoError(t, err, "error reading message response")
			payload := &TCPAssertions{}
			require.NoError(t, json.Unmarshal([]byte(message), payload))
			assertTestMessage(t, payload, nil)
		})
	})

	t.Run("with tls listener enabled", func(t *testing.T) {
		tcpPort, err := getFreePort(ctx)
		require.NoError(t, err, "error getting a port fot tcp server")

		t.Logf("Using %d as TCP Port", tcpPort)
		t.Setenv("TCP_PORT", strconv.Itoa(tcpPort))

		tlsPort, err := getFreePort(ctx)
		require.NoError(t, err, "error getting a random port fot tls server")
		t.Logf("Using %d as TLS Port", tlsPort)
		t.Setenv("TLS_PORT", strconv.Itoa(tlsPort))

		key, cert := generateSelfSignedKeypairForTests(t)
		t.Setenv("TLS_SERVER_PRIV_KEY", key)
		t.Setenv("TLS_SERVER_CERT", cert)

		errCh := make(chan error)
		go runServer(ctx, errCh)
		t.Cleanup(func() { close(errCh) })
		waitForListener(ctx, t, strconv.Itoa(tcpPort))
		waitForListener(ctx, t, strconv.Itoa(tlsPort))
		dialer := &net.Dialer{}

		t.Run("check against the TCP server", func(t *testing.T) {
			tcpClient, err := dialer.DialContext(ctx, "tcp", fmt.Sprintf("localhost:%s", strconv.Itoa(tcpPort)))
			require.NoError(t, err)

			t.Cleanup(func() {
				tcpClient.Close() //nolint: errcheck
			})

			t.Run("check the TCP server", func(t *testing.T) {
				message, err := bufio.NewReader(tcpClient).ReadString('\n')
				require.NoError(t, err, "error reading message response")
				assert.Equal(t, "Gateway API Test TCP Server\n", message)

				fmt.Fprintf(tcpClient, "PING\n")
				message, err = bufio.NewReader(tcpClient).ReadString('\n')
				require.NoError(t, err, "error reading message response")
				assert.Equal(t, "PONG\n", message)

				fmt.Fprintf(tcpClient, "IS_TLS\n")
				message, err = bufio.NewReader(tcpClient).ReadString('\n')
				require.NoError(t, err, "error reading message response")
				assert.Equal(t, "false\n", message)

				fmt.Fprintf(tcpClient, "TEST\n")
				message, err = bufio.NewReader(tcpClient).ReadString('\n')
				require.NoError(t, err, "error reading message response")
				payload := &TCPAssertions{}
				require.NoError(t, json.Unmarshal([]byte(message), payload))
				assertTestMessage(t, payload, nil)
			})
		})

		t.Run("check against a TLS server", func(t *testing.T) {
			cert, err := os.ReadFile(cert)
			require.NoError(t, err)
			certPool := x509.NewCertPool()
			require.True(t, certPool.AppendCertsFromPEM(cert))
			conf := &tls.Config{
				ServerName: "test.example.com",
				RootCAs:    certPool,
				MinVersion: tls.VersionTLS12,
			}

			tlsClient, err := (&tls.Dialer{
				Config: conf,
			}).DialContext(ctx, "tcp", fmt.Sprintf("localhost:%s", strconv.Itoa(tlsPort)))
			require.NoError(t, err)
			t.Cleanup(func() {
				tlsClient.Close() //nolint: errcheck
			})

			message, err := bufio.NewReader(tlsClient).ReadString('\n')
			require.NoError(t, err, "error reading message response")
			assert.Equal(t, "Gateway API Test TCP Server\n", message)

			fmt.Fprintf(tlsClient, "PING\n")
			message, err = bufio.NewReader(tlsClient).ReadString('\n')
			require.NoError(t, err, "error reading message response")
			assert.Equal(t, "PONG\n", message)

			fmt.Fprintf(tlsClient, "IS_TLS\n")
			message, err = bufio.NewReader(tlsClient).ReadString('\n')
			require.NoError(t, err, "error reading message response")
			assert.Equal(t, "true\n", message)

			fmt.Fprintf(tlsClient, "TEST\n")
			message, err = bufio.NewReader(tlsClient).ReadString('\n')
			require.NoError(t, err, "error reading message response")
			payload := &TCPAssertions{}
			require.NoError(t, json.Unmarshal([]byte(message), payload))
			expectedTLS := &TLSAssertions{
				ServerName:  "test.example.com",
				Version:     "TLS 1.3",
				Curves:      "X25519MLKEM768",
				CipherSuite: "TLS_AES_128_GCM_SHA256",
			}
			assertTestMessage(t, payload, expectedTLS)
		})
	})
}

func waitForListener(ctx context.Context, t *testing.T, port string) {
	t.Helper()
	require.Eventually(t, func() bool {
		tcpClient, err := (&net.Dialer{}).DialContext(ctx, "tcp", fmt.Sprintf("localhost:%s", port))
		if err != nil {
			t.Logf("got an error while trying to connect to tcpserver: %s, will retry", err)
			return false
		}
		require.NoError(t, tcpClient.Close())
		return true
	}, 5*time.Second, time.Second, "error waiting the server listener to be ready")
}

func assertTestMessage(t *testing.T, payload *TCPAssertions, expectedTLS *TLSAssertions) {
	t.Helper()
	require.NotNil(t, payload)
	assert.Equal(t, namespace, payload.Namespace, "namespace does not match")
	assert.Equal(t, service, payload.Service, "service does not match")
	assert.Equal(t, ingress, payload.Ingress, "ingress does not match")
	assert.Equal(t, pod, payload.Pod, "pod name does not match")

	assert.Equal(t, expectedTLS != nil, payload.IsTLS)
	assert.Equal(t, expectedTLS, payload.TLSAssertion)
}

func getFreePort(ctx context.Context) (int, error) {
	l, err := (&net.ListenConfig{}).Listen(ctx, "tcp", "localhost:0")
	if err != nil {
		return 0, err
	}
	defer l.Close()
	return l.Addr().(*net.TCPAddr).Port, nil
}

// This function generates a keypair on a temporary directory, and returns the key and cert path to be used on tests
func generateSelfSignedKeypairForTests(t *testing.T) (string, string) {
	t.Helper()
	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	require.NoError(t, err, "error generating private key")

	template := x509.Certificate{
		SerialNumber: big.NewInt(1),
		NotBefore:    time.Now(),
		NotAfter:     time.Now().Add(time.Minute),
		KeyUsage:     x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		DNSNames:     []string{"localhost", "test.example.com"},
	}

	derBytes, err := x509.CreateCertificate(rand.Reader, &template, &template, pub, priv)
	require.NoError(t, err, "error generating self signed certificate keypair for test")

	certPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "CERTIFICATE",
		Bytes: derBytes,
	})

	privBytes, err := x509.MarshalPKCS8PrivateKey(priv)
	require.NoError(t, err, "error encoding privatekey")

	keyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "PRIVATE KEY",
		Bytes: privBytes,
	})

	tmpDir, err := os.MkdirTemp("", "")
	require.NoError(t, err, "error creating temporary directory")

	keyPath := filepath.Join(tmpDir, "tls.key")
	crtPath := filepath.Join(tmpDir, "tls.crt")

	require.NoError(t, os.WriteFile(keyPath, keyPEM, 0o600), "error creating temporary private key file")
	require.NoError(t, os.WriteFile(crtPath, certPEM, 0o600), "error creating temporary certificate file")

	return keyPath, crtPath
}
