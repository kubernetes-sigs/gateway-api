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

package grpc

import (
	"context"
	"net"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"

	pb "sigs.k8s.io/gateway-api/conformance/echo-basic/grpcechoserver"
)

// testEchoServer is a minimal in-process GrpcEcho server whose Echo always
// succeeds, used to exercise DefaultClient against a real connection.
type testEchoServer struct {
	pb.UnimplementedGrpcEchoServer
}

func (testEchoServer) Echo(_ context.Context, _ *pb.EchoRequest) (*pb.EchoResponse, error) {
	return &pb.EchoResponse{}, nil
}

// TestDefaultClientConcurrentSendRPC drives a single shared DefaultClient from many
// goroutines. The exported client documents no single-goroutine restriction, so
// ensureConnection must serialize access to the shared connection: without the lock,
// `go test -race` flags a data race on DefaultClient.Conn. Every concurrent request
// against the in-process server must also succeed.
func TestDefaultClientConcurrentSendRPC(t *testing.T) {
	t.Parallel()

	lis, err := (&net.ListenConfig{}).Listen(context.Background(), "tcp", "127.0.0.1:0")
	require.NoError(t, err)
	srv := grpc.NewServer()
	pb.RegisterGrpcEchoServer(srv, testEchoServer{})
	go func() { _ = srv.Serve(lis) }()
	defer srv.Stop()
	addr := lis.Addr().String()

	c := &DefaultClient{}
	expected := ExpectedResponse{EchoRequest: &pb.EchoRequest{}}

	const n = 50
	start := make(chan struct{})
	results := make(chan codes.Code, n)
	var wg sync.WaitGroup
	wg.Add(n)
	for range n {
		go func() {
			defer wg.Done()
			<-start
			resp, err := c.SendRPC(t, addr, expected, 5*time.Second)
			if err != nil {
				results <- codes.Unknown
				return
			}
			results <- resp.Code
		}()
	}
	close(start)
	wg.Wait()
	close(results)

	for code := range results {
		require.Equal(t, codes.OK, code, "every concurrent SendRPC on a shared client must succeed")
	}
}

// TestDefaultClientReusableAfterClose verifies a DefaultClient can be reused after
// Close: Close must drop the underlying connection so the next SendRPC redials,
// instead of short-circuiting ensureConnection on an already-closed connection.
func TestDefaultClientReusableAfterClose(t *testing.T) {
	t.Parallel()

	lis, err := (&net.ListenConfig{}).Listen(context.Background(), "tcp", "127.0.0.1:0")
	require.NoError(t, err)
	srv := grpc.NewServer()
	pb.RegisterGrpcEchoServer(srv, testEchoServer{})
	go func() { _ = srv.Serve(lis) }()
	defer srv.Stop()
	addr := lis.Addr().String()

	expected := ExpectedResponse{EchoRequest: &pb.EchoRequest{}}

	c := &DefaultClient{}
	resp, err := c.SendRPC(t, addr, expected, 5*time.Second)
	require.NoError(t, err)
	require.Equal(t, codes.OK, resp.Code, "first request on a fresh client must succeed")

	c.Close()

	resp2, err2 := c.SendRPC(t, addr, expected, 5*time.Second)
	require.NoError(t, err2)
	require.Equal(t, codes.OK, resp2.Code, "a DefaultClient must be reusable after Close")
}
