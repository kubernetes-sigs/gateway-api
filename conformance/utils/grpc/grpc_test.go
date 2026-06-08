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
	"testing"
	"time"

	"google.golang.org/grpc/codes"

	pb "sigs.k8s.io/gateway-api/conformance/echo-basic/grpcechoserver"
	"sigs.k8s.io/gateway-api/conformance/utils/config"
)

// recordingClient is a Client that records whether Close was called and returns
// a canned response, so MakeRequestAndExpectEventuallyConsistentResponse can be
// exercised without a real gRPC server.
type recordingClient struct {
	closed    bool
	sendCalls int
	response  *Response
}

var _ Client = (*recordingClient)(nil)

func (c *recordingClient) SendRPC(_ *testing.T, _ string, _ ExpectedResponse, _ time.Duration) (*Response, error) {
	c.sendCalls++

	return c.response, nil
}

func (c *recordingClient) Close() { c.closed = true }

// TestMakeRequestDoesNotCloseCallerSuppliedClient pins the ownership contract:
// the helper closes only the DefaultClient it constructs when passed nil, never
// a caller-supplied (injected) client, which the caller may reuse afterwards.
func TestMakeRequestDoesNotCloseCallerSuppliedClient(t *testing.T) {
	t.Parallel()

	client := &recordingClient{response: &Response{Code: codes.Unavailable}}

	timeoutConfig := config.TimeoutConfig{
		RequiredConsecutiveSuccesses: 1,
		MaxTimeToConsistency:         5 * time.Second,
	}
	// A non-OK code keeps compareResponse from inspecting the (absent) echo
	// payload; the helper only needs the request to "succeed" once to return.
	expected := ExpectedResponse{
		EchoRequest: &pb.EchoRequest{},
		Response:    Response{Code: codes.Unavailable},
	}

	MakeRequestAndExpectEventuallyConsistentResponse(t, client, timeoutConfig, "192.0.2.1:50051", expected)

	if client.sendCalls == 0 {
		t.Fatal("expected the helper to call SendRPC on the supplied client")
	}
	if client.closed {
		t.Fatal("helper must not Close a caller-supplied client it does not own")
	}
}
