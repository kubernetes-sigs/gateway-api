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

const (
	// BodyByteLimit is the max limit of 62Kb per streamed chunk.
	// Certain envoy implementations set a max limit of 64Kb per streamed chunk, intentionally setting this lower for a safe margin.
	BodyByteLimit = 62000
)

// BuildChunkedBodyResponses splits the bodyBytes into chunks of size BodyByteLimit and returns a slice of CommonResponse.
// If setEos is true, the last chunk will have EndOfStream set to true.
func BuildChunkedBodyResponses(bodyBytes []byte, setEos bool) []*extProcPb.CommonResponse {
	responses := []*extProcPb.CommonResponse{}
	startingIndex := 0
	bodyLen := len(bodyBytes)

	if bodyLen == 0 {
		return []*extProcPb.CommonResponse{
			{
				BodyMutation: &extProcPb.BodyMutation{
					Mutation: &extProcPb.BodyMutation_StreamedResponse{
						StreamedResponse: &extProcPb.StreamedBodyResponse{
							Body:        bodyBytes,
							EndOfStream: setEos,
						},
					},
				},
			},
		}
	}

	for startingIndex < bodyLen {
		eos := false
		len := min(bodyLen-startingIndex, BodyByteLimit)
		chunk := bodyBytes[startingIndex : len+startingIndex]
		if setEos && len+startingIndex >= bodyLen {
			eos = true
		}

		commonResp := &extProcPb.CommonResponse{
			BodyMutation: &extProcPb.BodyMutation{
				Mutation: &extProcPb.BodyMutation_StreamedResponse{
					StreamedResponse: &extProcPb.StreamedBodyResponse{
						Body:        chunk,
						EndOfStream: eos,
					},
				},
			},
		}
		responses = append(responses, commonResp)
		startingIndex += len
	}

	return responses
}
