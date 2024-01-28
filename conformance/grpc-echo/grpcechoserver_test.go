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

package main

import (
	"testing"
	"context"
	"time"
	"fmt"
	"math/rand"

	"google.golang.org/grpc"
	// "google.golang.org/grpc/reflection"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/credentials/insecure"
	// "google.golang.org/grpc/credentials"
	// "google.golang.org/grpc/peer"

	"google.golang.org/protobuf/proto"

	pb "sigs.k8s.io/gateway-api/conformance/grpc-echo/grpcechoserver"
)

const ServerAddress = "127.0.0.1"

// Let the kernel resolve an open port so multiple test instances can run concurrently.
const ServerHTTPPort = 0
const ServerHTTPSPort = 0
const RPCTimeout = 10 * time.Second

const letters = "abcdefghijklmnopqrstuvwxyz0123456789"

func randStr(length int) string {
	s := ""
	for i := 0; i < length; i++ {
		letter := letters[rand.Int() % len(letters)]
		s = s + string([]byte{letter})
	}
	return s
}

func TestReflectionService(t *testing.T) {
	// TODO: Implement.
}

// TODO: Parameterize over Echo2 and Echo3 as well.
func TestEchoService(t *testing.T) {
	podContext := pb.Context{
		Namespace: 	randStr(12),
		Ingress:   	randStr(12),
		ServiceName:   	randStr(12),
		Pod:       	randStr(12),
	}
	config := serverConfig{
		HTTPPort: ServerHTTPPort,
		PodContext: podContext,
	}
	httpPort, _ := runServer(config)

	dialOpts := []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}

	serverTarget := fmt.Sprintf("%s:%d", ServerAddress, httpPort)
	conn, err := grpc.Dial(serverTarget, dialOpts...)
	if err != nil {
		t.Fatal(err)
	}

	headers := &metadata.MD{}
	trailers := &metadata.MD{}
	ctx, _ := context.WithTimeout(context.Background(), RPCTimeout)

	stub := pb.NewGrpcEchoClient(conn)
	req := pb.EchoRequest{}
	resp, err := stub.Echo(ctx, &req, grpc.Header(headers), grpc.Trailer(trailers))
	if err != nil {
		t.Fatal(err)
	}

	echoedReq := pb.EchoRequest{}

	// nil is equivalent to default value.
	if resp.GetRequest() != nil {
		echoedReq = *resp.GetRequest()
	}

	if !proto.Equal(&echoedReq, &req) {
		t.Fatalf("echoed request did not equal sent request. expected: %v\ngot: %v\n", req, echoedReq)
	}

	if resp.GetAssertions() == nil {
		t.Fatalf("no assertions populated in response: %v", resp.GetAssertions())
	}

	const expectedFullyQualifiedMethod = "/gateway_api_conformance.grpc_echo.grpcecho.GrpcEcho/Echo"
	if resp.GetAssertions().GetFullyQualifiedMethod() != expectedFullyQualifiedMethod {
		t.Fatalf("fully_qualified_method wrong. expected: %s, got: %s", resp.GetAssertions().GetFullyQualifiedMethod(), expectedFullyQualifiedMethod)
	}

	if resp.GetAssertions().GetAuthority() != serverTarget {
		t.Fatalf("serverTarget wrong. expected: %s, got: %s", resp.GetAssertions().GetAuthority(), serverTarget)
	}

	if resp.GetAssertions().GetContext() == nil || !proto.Equal(resp.GetAssertions().GetContext(), &podContext) {
		t.Fatalf("podContext wrong. expected %v\ngot: %v", podContext, resp.GetAssertions().GetContext())
	}

	t.Fatalf("%v", resp)
}

