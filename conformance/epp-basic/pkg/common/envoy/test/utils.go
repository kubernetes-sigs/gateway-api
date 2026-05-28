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

package test

import (
	"sort"

	extProcPb "github.com/envoyproxy/go-control-plane/envoy/service/ext_proc/v3"
)

// SortSetHeadersInResponses sorts HeaderMutation.SetHeaders by key in each response for deterministic comparison,
// since map iteration order in GenerateHeadersMutation is undefined. Use on both want and got before cmp.Diff.
func SortSetHeadersInResponses(responses []*extProcPb.ProcessingResponse) {
	for _, r := range responses {
		if r == nil || r.Response == nil {
			continue
		}
		var common *extProcPb.CommonResponse
		switch rr := r.Response.(type) {
		case *extProcPb.ProcessingResponse_RequestHeaders:
			if rr.RequestHeaders != nil && rr.RequestHeaders.Response != nil {
				common = rr.RequestHeaders.Response
			}
		case *extProcPb.ProcessingResponse_RequestBody:
			if rr.RequestBody != nil && rr.RequestBody.Response != nil {
				common = rr.RequestBody.Response
			}
		case *extProcPb.ProcessingResponse_ResponseHeaders:
			if rr.ResponseHeaders != nil && rr.ResponseHeaders.Response != nil {
				common = rr.ResponseHeaders.Response
			}
		case *extProcPb.ProcessingResponse_ResponseBody:
			if rr.ResponseBody != nil && rr.ResponseBody.Response != nil {
				common = rr.ResponseBody.Response
			}
		default:
			continue
		}
		if common != nil && common.HeaderMutation != nil && len(common.HeaderMutation.SetHeaders) > 1 {
			sort.Slice(common.HeaderMutation.SetHeaders, func(i, j int) bool {
				return common.HeaderMutation.SetHeaders[i].GetHeader().GetKey() <
					common.HeaderMutation.SetHeaders[j].GetHeader().GetKey()
			})
		}
	}
}
