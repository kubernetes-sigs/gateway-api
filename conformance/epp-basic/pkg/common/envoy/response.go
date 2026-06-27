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
	eppb "github.com/envoyproxy/go-control-plane/envoy/service/ext_proc/v3"
)

// AddStreamedResponseBody splits responseBodyBytes into chunked body responses
// and appends them as ResponseBody ProcessingResponses, mirroring
// GenerateRequestBodyResponses for the request path.
func AddStreamedResponseBody(responses []*eppb.ProcessingResponse, responseBodyBytes []byte) []*eppb.ProcessingResponse {
	commonResponses := BuildChunkedBodyResponses(responseBodyBytes, true)
	for _, commonResp := range commonResponses {
		responses = append(responses, &eppb.ProcessingResponse{
			Response: &eppb.ProcessingResponse_ResponseBody{
				ResponseBody: &eppb.BodyResponse{
					Response: commonResp,
				},
			},
		})
	}
	return responses
}
