/*
Copyright 2025 The Kubernetes Authors.

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

package handlers

import (
	"context"
	"io"
	"net"
	"sync/atomic"

	configPb "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	extProcPb "github.com/envoyproxy/go-control-plane/envoy/service/ext_proc/v3"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/structpb"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log"

	"sigs.k8s.io/gateway-api/conformance/epp-basic/pkg/datastore"
	"sigs.k8s.io/gateway-api/conformance/epp-basic/pkg/metadata"
)

// Datastore defines the interface required by the server.
type Datastore interface {
	PoolGet() (*datastore.EndpointPool, error)
	PodList(predicate func(*datastore.Endpoint) bool) []*datastore.Endpoint
}

func NewStreamingServer(datastore Datastore) *StreamingServer {
	return &StreamingServer{
		datastore: datastore,
		picker:    &RoundRobinPicker{},
	}
}

// StreamingServer implements the Envoy external processing server.
type StreamingServer struct {
	datastore Datastore
	picker    EndpointPicker
}

// RequestContext stores context information during the life time of an HTTP request.
type RequestContext struct {
	TargetEndpoint string
	SelectedPodIP  string
}

// PickRequest contains metadata from the incoming request to guide endpoint selection.
type PickRequest struct {
	Headers map[string]string
	Body    []byte
	Model   string
}

// PickResult holds the selected backend endpoint information.
type PickResult struct {
	Endpoint     string            // Primary endpoint (ip:port)
	Fallbacks    []string          // Optional fallback endpoints
	MutatedBody  []byte            // If non-nil, replaces the request body forwarded to Envoy
	ExtraHeaders map[string]string // Additional headers to set on the request
}

// EndpointPicker defines a strategy for selecting backend inference endpoints.
type EndpointPicker interface {
	Pick(ctx context.Context, req *PickRequest, endpoints []*datastore.Endpoint) (*PickResult, error)
}

// RoundRobinPicker implements a basic round-robin endpoint selection strategy.
type RoundRobinPicker struct {
	rrIndex uint64
}

// Pick selects an endpoint based on round-robin ordering.
func (p *RoundRobinPicker) Pick(ctx context.Context, req *PickRequest, endpoints []*datastore.Endpoint) (*PickResult, error) {
	if len(endpoints) == 0 {
		return nil, status.Errorf(codes.Unavailable, "no endpoints available")
	}

	index := atomic.AddUint64(&p.rrIndex, 1)
	selectedPod := endpoints[index%uint64(len(endpoints))]

	return &PickResult{
		Endpoint: net.JoinHostPort(selectedPod.Address, selectedPod.Port),
	}, nil
}

func (s *StreamingServer) Process(srv extProcPb.ExternalProcessor_ProcessServer) error {
	ctx := srv.Context()
	logger := ctrl.Log.WithName("ext-proc")
	ctx = log.IntoContext(ctx, logger)
	logger.Info("Processing new stream")

	reqCtx := &RequestContext{}

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		req, recvErr := srv.Recv()
		if recvErr == io.EOF || status.Code(recvErr) == codes.Canceled {
			return nil
		}
		if recvErr != nil {
			return status.Errorf(codes.Unknown, "cannot receive stream request: %v", recvErr)
		}

		switch v := req.Request.(type) {
		case *extProcPb.ProcessingRequest_RequestHeaders:
			logger.Info("Received request headers")
			err := s.handleRequestHeaders(ctx, reqCtx, req, v)
			if err != nil {
				logger.Error(err, "Failed to handle request headers")
				return status.Errorf(codes.Internal, "internal error: %v", err)
			}

			resp := &extProcPb.ProcessingResponse{
				Response: &extProcPb.ProcessingResponse_RequestHeaders{
					RequestHeaders: &extProcPb.HeadersResponse{
						Response: &extProcPb.CommonResponse{
							ClearRouteCache: true,
							HeaderMutation: &extProcPb.HeaderMutation{
								SetHeaders: []*configPb.HeaderValueOption{
									{
										Header: &configPb.HeaderValue{
											Key:      metadata.DestinationEndpointKey,
											RawValue: []byte(reqCtx.TargetEndpoint),
										},
									},
									{
										Header: &configPb.HeaderValue{
											Key:      "X-Echo-Set-Header",
											RawValue: []byte(metadata.ConformanceTestResultHeader + ":" + reqCtx.TargetEndpoint),
										},
									},
								},
							},
						},
					},
				},
				DynamicMetadata: &structpb.Struct{
					Fields: map[string]*structpb.Value{
						metadata.DestinationEndpointNamespace: structpb.NewStructValue(&structpb.Struct{
							Fields: map[string]*structpb.Value{
								metadata.DestinationEndpointKey: structpb.NewStringValue(reqCtx.TargetEndpoint),
							},
						}),
					},
				},
			}
			if err := srv.Send(resp); err != nil {
				return status.Errorf(codes.Unknown, "failed to send response back to Envoy: %v", err)
			}

		case *extProcPb.ProcessingRequest_RequestBody:
			logger.V(1).Info("Received request body", "endOfStream", v.RequestBody.EndOfStream)
			resp := &extProcPb.ProcessingResponse{
				Response: &extProcPb.ProcessingResponse_RequestBody{
					RequestBody: &extProcPb.BodyResponse{
						Response: &extProcPb.CommonResponse{},
					},
				},
			}
			if err := srv.Send(resp); err != nil {
				return status.Errorf(codes.Unknown, "failed to send body response back to Envoy: %v", err)
			}

		case *extProcPb.ProcessingRequest_ResponseHeaders:
			logger.Info("Received response headers")
			resp := s.handleResponseHeaders(ctx, req, v)
			if err := srv.Send(resp); err != nil {
				return status.Errorf(codes.Unknown, "failed to send response back to Envoy: %v", err)
			}

		case *extProcPb.ProcessingRequest_ResponseBody:
			logger.V(1).Info("Received response body", "endOfStream", v.ResponseBody.EndOfStream)
			resp := &extProcPb.ProcessingResponse{
				Response: &extProcPb.ProcessingResponse_ResponseBody{
					ResponseBody: &extProcPb.BodyResponse{
						Response: &extProcPb.CommonResponse{},
					},
				},
			}

			if err := srv.Send(resp); err != nil {
				return status.Errorf(codes.Unknown, "failed to send response body back to Envoy: %v", err)
			}

		default:
			// Ignore other request types (Trailers)
			logger.V(1).Info("Ignoring request type", "type", v)
		}
	}
}
