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

	configPb "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	extProcPb "github.com/envoyproxy/go-control-plane/envoy/service/ext_proc/v3"
	"sigs.k8s.io/controller-runtime/pkg/log"

	envoy "sigs.k8s.io/gateway-api/conformance/epp-basic/pkg/common/envoy"
	"sigs.k8s.io/gateway-api/conformance/epp-basic/pkg/metadata"
)

func (s *StreamingServer) handleResponseHeaders(ctx context.Context, fullReq *extProcPb.ProcessingRequest, respHeaders *extProcPb.ProcessingRequest_ResponseHeaders) *extProcPb.ProcessingResponse {
	logger := log.FromContext(ctx)
	logger.Info("Handling response headers")

	// Read the endpoint that actually served the request from Envoy's response metadata.
	// GKE Gateway sets envoy.lb.x-gateway-destination-endpoint-served after routing to the backend.
	var servedEndpoint string
	respMetadata := envoy.ExtractMetadataValues(fullReq)
	logger.Info("Extracted metadata in response headers", "metadata", respMetadata)
	lbMetadata, ok := respMetadata[metadata.DestinationEndpointNamespace].(map[string]any)
	if !ok {
		servedEndpoint = "fail: missing envoy lb metadata"
	} else if served, ok := lbMetadata[metadata.DestinationEndpointServedKey].(string); !ok {
		servedEndpoint = "fail: missing destination endpoint served metadata"
	} else {
		servedEndpoint = served
	}

	logger.Info("Setting conformance test result header", "header", metadata.ConformanceTestResultHeader, "value", servedEndpoint)

	headers := []*configPb.HeaderValueOption{
		{
			Header: &configPb.HeaderValue{
				Key:      metadata.ConformanceTestResultHeader,
				RawValue: []byte(servedEndpoint),
			},
		},
		{
			Header: &configPb.HeaderValue{
				// This is for debugging purpose only.
				Key:      "x-went-into-resp-headers",
				RawValue: []byte("true"),
			},
		},
	}

	// Include any non-system-owned headers from the original response.
	if respHeaders != nil && respHeaders.ResponseHeaders != nil && respHeaders.ResponseHeaders.Headers != nil {
		for _, header := range respHeaders.ResponseHeaders.Headers.Headers {
			key := header.Key
			headers = append(headers, &configPb.HeaderValueOption{
				Header: &configPb.HeaderValue{
					Key:      key,
					RawValue: []byte(envoy.GetHeaderValue(header)),
				},
			})
		}
	}

	resp := &extProcPb.ProcessingResponse{
		Response: &extProcPb.ProcessingResponse_ResponseHeaders{
			ResponseHeaders: &extProcPb.HeadersResponse{
				Response: &extProcPb.CommonResponse{
					HeaderMutation: &extProcPb.HeaderMutation{
						SetHeaders: headers,
					},
				},
			},
		},
	}

	return resp
}
