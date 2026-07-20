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
	"testing"

	envoyCorev3 "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	extProcPb "github.com/envoyproxy/go-control-plane/envoy/service/ext_proc/v3"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/types/known/structpb"

	"sigs.k8s.io/gateway-api/conformance/epp-basic/pkg/metadata"
)

func TestHandleResponseHeaders_MissingEnvoyLBMetadata(t *testing.T) {
	server := &StreamingServer{}

	resp := server.handleResponseHeaders(context.Background(), nil, &extProcPb.ProcessingRequest_ResponseHeaders{})

	setHeaders := resp.GetResponseHeaders().GetResponse().GetHeaderMutation().GetSetHeaders()
	assert.Len(t, setHeaders, 2)
	assert.Equal(t, metadata.ConformanceTestResultHeader, setHeaders[0].GetHeader().GetKey())
	assert.Equal(t, "fail: missing envoy lb metadata", string(setHeaders[0].GetHeader().GetRawValue()))
	assert.Equal(t, "x-went-into-resp-headers", setHeaders[1].GetHeader().GetKey())
	assert.Equal(t, "true", string(setHeaders[1].GetHeader().GetRawValue()))
}

func TestHandleResponseHeaders_MissingDestinationEndpointServedKey(t *testing.T) {
	server := &StreamingServer{}

	fullReq := &extProcPb.ProcessingRequest{
		MetadataContext: &envoyCorev3.Metadata{
			FilterMetadata: map[string]*structpb.Struct{
				metadata.DestinationEndpointNamespace: {
					Fields: map[string]*structpb.Value{
						"some-other-key": structpb.NewStringValue("value"),
					},
				},
			},
		},
	}

	resp := server.handleResponseHeaders(context.Background(), fullReq, &extProcPb.ProcessingRequest_ResponseHeaders{})

	setHeaders := resp.GetResponseHeaders().GetResponse().GetHeaderMutation().GetSetHeaders()
	assert.Len(t, setHeaders, 2)
	assert.Equal(t, metadata.ConformanceTestResultHeader, setHeaders[0].GetHeader().GetKey())
	assert.Equal(t, "fail: missing destination endpoint served metadata", string(setHeaders[0].GetHeader().GetRawValue()))
	assert.Equal(t, "x-went-into-resp-headers", setHeaders[1].GetHeader().GetKey())
	assert.Equal(t, "true", string(setHeaders[1].GetHeader().GetRawValue()))
}

func TestHandleResponseHeaders_UsesServedEndpointFromMetadata(t *testing.T) {
	server := &StreamingServer{}

	fullReq := &extProcPb.ProcessingRequest{
		MetadataContext: &envoyCorev3.Metadata{
			FilterMetadata: map[string]*structpb.Struct{
				metadata.DestinationEndpointNamespace: {
					Fields: map[string]*structpb.Value{
						metadata.DestinationEndpointServedKey: structpb.NewStringValue("10.0.0.2:3000"),
					},
				},
			},
		},
	}

	resp := server.handleResponseHeaders(context.Background(), fullReq, &extProcPb.ProcessingRequest_ResponseHeaders{})

	setHeaders := resp.GetResponseHeaders().GetResponse().GetHeaderMutation().GetSetHeaders()
	assert.Len(t, setHeaders, 2)
	assert.Equal(t, metadata.ConformanceTestResultHeader, setHeaders[0].GetHeader().GetKey())
	assert.Equal(t, "10.0.0.2:3000", string(setHeaders[0].GetHeader().GetRawValue()))
	assert.Equal(t, "x-went-into-resp-headers", setHeaders[1].GetHeader().GetKey())
	assert.Equal(t, "true", string(setHeaders[1].GetHeader().GetRawValue()))
}

func TestHandleResponseHeaders_ForwardsOriginalHeaders(t *testing.T) {
	server := &StreamingServer{}

	originalHeaders := &extProcPb.ProcessingRequest_ResponseHeaders{
		ResponseHeaders: &extProcPb.HttpHeaders{
			Headers: &envoyCorev3.HeaderMap{
				Headers: []*envoyCorev3.HeaderValue{
					{
						Key:   "x-custom-header",
						Value: "custom-value",
					},
				},
			},
		},
	}

	resp := server.handleResponseHeaders(context.Background(), nil, originalHeaders)

	setHeaders := resp.GetResponseHeaders().GetResponse().GetHeaderMutation().GetSetHeaders()
	assert.Len(t, setHeaders, 3)
	assert.Equal(t, metadata.ConformanceTestResultHeader, setHeaders[0].GetHeader().GetKey())
	assert.Equal(t, "x-went-into-resp-headers", setHeaders[1].GetHeader().GetKey())
	assert.Equal(t, "x-custom-header", setHeaders[2].GetHeader().GetKey())
	assert.Equal(t, "custom-value", string(setHeaders[2].GetHeader().GetRawValue()))
}
