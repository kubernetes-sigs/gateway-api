/*
Copyright 2024 The Kubernetes Authors.

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
	"fmt"
	"math/rand"
	"testing"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	"google.golang.org/protobuf/encoding/prototext"
	"google.golang.org/protobuf/proto"

	pb "sigs.k8s.io/gateway-api/conformance/echo-basic/grpcechoserver"
)

const ServerAddress = "127.0.0.1"

// Let the kernel resolve an open port so multiple test instances can run concurrently.
const (
	ServerHTTPPort  = 0
	ServerHTTPSPort = 0
	RPCTimeout      = 10 * time.Second
)

const letters = "abcdefghijklmnopqrstuvwxyz0123456789"

func randStr(length int) string { //nolint:unparam
	s := ""
	for i := 0; i < length; i++ {
		letter := letters[rand.Int()%len(letters)] //nolint:gosec // This is test code.
		s += string([]byte{letter})
	}
	return s
}

type methodFunc = func(context.Context, pb.GrpcEchoClient, *pb.EchoRequest) (*pb.EchoResponse, error)

func clientAndServer(t *testing.T) (pb.GrpcEchoClient, serverConfig, string) {
	t.Helper()
	podContext := &pb.Context{
		Namespace:   randStr(12),
		Ingress:     randStr(12),
		ServiceName: randStr(12),
		Pod:         randStr(12),
	}
	config := serverConfig{
		HTTPPort:   ServerHTTPPort,
		PodContext: podContext,
	}
	httpPort, _ := runServer(config)

	dialOpts := []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}

	serverTarget := fmt.Sprintf("%s:%d", ServerAddress, httpPort)
	conn, err := grpc.NewClient(serverTarget, dialOpts...)
	if err != nil {
		t.Fatal(err)
	}

	stub := pb.NewGrpcEchoClient(conn)
	return stub, config, serverTarget
}

func testEchoMethod(t *testing.T, methodName string, f methodFunc) {
	t.Helper()
	stub, config, serverTarget := clientAndServer(t)

	const testHeaderKey = "foo"
	testHeaderValue := randStr(12)
	ctx, cancel := context.WithTimeout(context.Background(), RPCTimeout)
	defer cancel()
	ctx = metadata.AppendToOutgoingContext(ctx, testHeaderKey, testHeaderValue)

	req := pb.EchoRequest{}
	resp, err := f(ctx, stub, &req)
	if err != nil {
		t.Fatal(err)
	}

	echoedReq := pb.EchoRequest{}

	// nil is equivalent to default value.
	if resp.GetRequest() != nil {
		echoedReq = *resp.GetRequest()
	}

	if !proto.Equal(&echoedReq, &req) {
		t.Fatalf("echoed request did not equal sent request. expected: %s\ngot: %s\n", prototext.Format(&req), prototext.Format(&echoedReq))
	}

	if resp.GetAssertions() == nil {
		t.Fatalf("no assertions populated in response: %v", resp.GetAssertions())
	}

	const fullyQualifiedService = "/gateway_api_conformance.echo_basic.grpcecho.GrpcEcho/"
	expectedFullyQualifiedMethod := fullyQualifiedService + methodName
	if resp.GetAssertions().GetFullyQualifiedMethod() != expectedFullyQualifiedMethod {
		t.Fatalf("fully_qualified_method wrong. expected: %s, got: %s", resp.GetAssertions().GetFullyQualifiedMethod(), expectedFullyQualifiedMethod)
	}

	if resp.GetAssertions().GetAuthority() != serverTarget {
		t.Fatalf("serverTarget wrong. expected: %s, got: %s", resp.GetAssertions().GetAuthority(), serverTarget)
	}

	if resp.GetAssertions().GetContext() == nil || !proto.Equal(resp.GetAssertions().GetContext(), config.PodContext) {
		t.Fatalf("podContext wrong. expected %s\ngot: %s", prototext.Format(config.PodContext), prototext.Format(resp.GetAssertions().GetContext()))
	}

	echoedTestHeaderValues := []string{}
	for _, header := range resp.GetAssertions().GetHeaders() {
		if header.GetKey() == testHeaderKey {
			echoedTestHeaderValues = append(echoedTestHeaderValues, header.GetValue())
		}
	}

	if len(echoedTestHeaderValues) != 1 {
		t.Fatalf("echoed header value had unexpected size %d: %v", len(echoedTestHeaderValues), echoedTestHeaderValues)
	}

	echoedTestHeaderValue := echoedTestHeaderValues[0]

	if echoedTestHeaderValue != testHeaderValue {
		t.Fatalf("echoed header value was wrong. expected: %s, got: %s", testHeaderValue, echoedTestHeaderValue)
	}
}

func TestEchoMethod(t *testing.T) {
	testEchoMethod(t, "Echo", func(ctx context.Context, stub pb.GrpcEchoClient, req *pb.EchoRequest) (*pb.EchoResponse, error) {
		return stub.Echo(ctx, req)
	})
}

func TestEchoTwoMethod(t *testing.T) {
	testEchoMethod(t, "EchoTwo", func(ctx context.Context, stub pb.GrpcEchoClient, req *pb.EchoRequest) (*pb.EchoResponse, error) {
		return stub.EchoTwo(ctx, req)
	})
}

func TestEchoThreeMethod(t *testing.T) {
	stub, _, _ := clientAndServer(t)
	ctx, cancel := context.WithTimeout(context.Background(), RPCTimeout)
	defer cancel()
	req := pb.EchoRequest{}
	resp, err := stub.EchoThree(ctx, &req)
	if err == nil {
		t.Fatalf("Expected RPC to fail but got success: %v", resp)
	}

	code := status.Code(err)
	if code != codes.Unimplemented {
		t.Fatalf("Expected code Unimplemented but found %v: %v", code, err)
	}
}
