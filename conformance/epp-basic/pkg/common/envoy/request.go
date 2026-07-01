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

package envoy

import (
	extProcPb "github.com/envoyproxy/go-control-plane/envoy/service/ext_proc/v3"
)

// GenerateRequestBodyResponses splits the request body bytes into chunked body
// responses and wraps each chunk in a ProcessingResponse_RequestBody envelope.
func GenerateRequestBodyResponses(requestBodyBytes []byte) []*extProcPb.ProcessingResponse {
	commonResponses := BuildChunkedBodyResponses(requestBodyBytes, true)
	responses := make([]*extProcPb.ProcessingResponse, 0, len(commonResponses))
	for _, commonResp := range commonResponses {
		resp := &extProcPb.ProcessingResponse{
			Response: &extProcPb.ProcessingResponse_RequestBody{
				RequestBody: &extProcPb.BodyResponse{
					Response: commonResp,
				},
			},
		}
		responses = append(responses, resp)
	}
	return responses
}
