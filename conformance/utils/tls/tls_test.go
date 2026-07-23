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

package tls

import (
	"context"
	"io"
	"net"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"sigs.k8s.io/gateway-api/conformance/utils/config"
)

type dialContextFunc func(ctx context.Context, network, address string) (net.Conn, error)

func (f dialContextFunc) DialContext(ctx context.Context, network, address string) (net.Conn, error) {
	return f(ctx, network, address)
}

type trackedConn struct {
	net.Conn
	closed atomic.Bool
}

func (c *trackedConn) Close() error {
	c.closed.Store(true)
	return c.Conn.Close()
}

func TestWaitForTLSConnectionRejectionRetriesAfterAttemptTimeout(t *testing.T) {
	var attempts atomic.Int32
	dialer := dialContextFunc(func(ctx context.Context, _, _ string) (net.Conn, error) {
		if attempts.Add(1) == 1 {
			<-ctx.Done()
			return nil, ctx.Err()
		}
		return nil, io.EOF
	})
	timeoutConfig := config.TimeoutConfig{
		MaxTimeToConsistency: 500 * time.Millisecond,
		RequestTimeout:       20 * time.Millisecond,
	}

	waitForTLSConnectionRejection(t, timeoutConfig, dialer, "example.com:443", time.Millisecond)

	assert.GreaterOrEqual(t, attempts.Load(), int32(2))
}

func TestWaitForTLSConnectionRejectionClosesUnexpectedConnection(t *testing.T) {
	clientConn, serverConn := net.Pipe()
	defer serverConn.Close()

	conn := &trackedConn{Conn: clientConn}
	var attempts atomic.Int32
	dialer := dialContextFunc(func(context.Context, string, string) (net.Conn, error) {
		if attempts.Add(1) == 1 {
			return conn, nil
		}
		return nil, io.EOF
	})
	timeoutConfig := config.TimeoutConfig{
		MaxTimeToConsistency: 500 * time.Millisecond,
		RequestTimeout:       20 * time.Millisecond,
	}

	waitForTLSConnectionRejection(t, timeoutConfig, dialer, "example.com:443", time.Millisecond)

	assert.True(t, conn.closed.Load())
}
