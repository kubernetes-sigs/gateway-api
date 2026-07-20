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
	"testing"

	corev3 "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	extProcPb "github.com/envoyproxy/go-control-plane/envoy/service/ext_proc/v3"
	"github.com/google/go-cmp/cmp"
	"google.golang.org/protobuf/testing/protocmp"
)

func makeHeaderValueOption(key, value string) *corev3.HeaderValueOption {
	return &corev3.HeaderValueOption{
		Header: &corev3.HeaderValue{
			Key:      key,
			RawValue: []byte(value),
		},
	}
}

func TestSortSetHeadersInResponses(t *testing.T) {
	tests := []struct {
		name      string
		responses []*extProcPb.ProcessingResponse
		want      []*extProcPb.ProcessingResponse
	}{
		{
			name:      "nil slice",
			responses: nil,
			want:      nil,
		},
		{
			name:      "nil response element is skipped",
			responses: []*extProcPb.ProcessingResponse{nil},
			want:      []*extProcPb.ProcessingResponse{nil},
		},
		{
			name: "response with nil Response field is skipped",
			responses: []*extProcPb.ProcessingResponse{
				{Response: nil},
			},
			want: []*extProcPb.ProcessingResponse{
				{Response: nil},
			},
		},
		{
			name: "RequestHeaders response with unsorted headers gets sorted",
			responses: []*extProcPb.ProcessingResponse{
				{
					Response: &extProcPb.ProcessingResponse_RequestHeaders{
						RequestHeaders: &extProcPb.HeadersResponse{
							Response: &extProcPb.CommonResponse{
								HeaderMutation: &extProcPb.HeaderMutation{
									SetHeaders: []*corev3.HeaderValueOption{
										makeHeaderValueOption("z-header", "z-val"),
										makeHeaderValueOption("a-header", "a-val"),
										makeHeaderValueOption("m-header", "m-val"),
									},
								},
							},
						},
					},
				},
			},
			want: []*extProcPb.ProcessingResponse{
				{
					Response: &extProcPb.ProcessingResponse_RequestHeaders{
						RequestHeaders: &extProcPb.HeadersResponse{
							Response: &extProcPb.CommonResponse{
								HeaderMutation: &extProcPb.HeaderMutation{
									SetHeaders: []*corev3.HeaderValueOption{
										makeHeaderValueOption("a-header", "a-val"),
										makeHeaderValueOption("m-header", "m-val"),
										makeHeaderValueOption("z-header", "z-val"),
									},
								},
							},
						},
					},
				},
			},
		},
		{
			name: "ResponseHeaders with unsorted headers gets sorted",
			responses: []*extProcPb.ProcessingResponse{
				{
					Response: &extProcPb.ProcessingResponse_ResponseHeaders{
						ResponseHeaders: &extProcPb.HeadersResponse{
							Response: &extProcPb.CommonResponse{
								HeaderMutation: &extProcPb.HeaderMutation{
									SetHeaders: []*corev3.HeaderValueOption{
										makeHeaderValueOption("z-header", "z-val"),
										makeHeaderValueOption("a-header", "a-val"),
									},
								},
							},
						},
					},
				},
			},
			want: []*extProcPb.ProcessingResponse{
				{
					Response: &extProcPb.ProcessingResponse_ResponseHeaders{
						ResponseHeaders: &extProcPb.HeadersResponse{
							Response: &extProcPb.CommonResponse{
								HeaderMutation: &extProcPb.HeaderMutation{
									SetHeaders: []*corev3.HeaderValueOption{
										makeHeaderValueOption("a-header", "a-val"),
										makeHeaderValueOption("z-header", "z-val"),
									},
								},
							},
						},
					},
				},
			},
		},
		{
			name: "ResponseBody with unsorted headers gets sorted",
			responses: []*extProcPb.ProcessingResponse{
				{
					Response: &extProcPb.ProcessingResponse_ResponseBody{
						ResponseBody: &extProcPb.BodyResponse{
							Response: &extProcPb.CommonResponse{
								HeaderMutation: &extProcPb.HeaderMutation{
									SetHeaders: []*corev3.HeaderValueOption{
										makeHeaderValueOption("c-header", "c-val"),
										makeHeaderValueOption("a-header", "a-val"),
									},
								},
							},
						},
					},
				},
			},
			want: []*extProcPb.ProcessingResponse{
				{
					Response: &extProcPb.ProcessingResponse_ResponseBody{
						ResponseBody: &extProcPb.BodyResponse{
							Response: &extProcPb.CommonResponse{
								HeaderMutation: &extProcPb.HeaderMutation{
									SetHeaders: []*corev3.HeaderValueOption{
										makeHeaderValueOption("a-header", "a-val"),
										makeHeaderValueOption("c-header", "c-val"),
									},
								},
							},
						},
					},
				},
			},
		},
		{
			name: "unhandled response type is skipped",
			responses: []*extProcPb.ProcessingResponse{
				{
					Response: &extProcPb.ProcessingResponse_ResponseTrailers{},
				},
			},
			want: []*extProcPb.ProcessingResponse{
				{
					Response: &extProcPb.ProcessingResponse_ResponseTrailers{},
				},
			},
		},
		{
			name: "nil HeaderMutation is skipped",
			responses: []*extProcPb.ProcessingResponse{
				{
					Response: &extProcPb.ProcessingResponse_RequestHeaders{
						RequestHeaders: &extProcPb.HeadersResponse{
							Response: &extProcPb.CommonResponse{
								HeaderMutation: nil,
							},
						},
					},
				},
			},
			want: []*extProcPb.ProcessingResponse{
				{
					Response: &extProcPb.ProcessingResponse_RequestHeaders{
						RequestHeaders: &extProcPb.HeadersResponse{
							Response: &extProcPb.CommonResponse{
								HeaderMutation: nil,
							},
						},
					},
				},
			},
		},
		{
			name: "nil and valid responses mixed",
			responses: []*extProcPb.ProcessingResponse{
				nil,
				{
					Response: &extProcPb.ProcessingResponse_RequestHeaders{
						RequestHeaders: &extProcPb.HeadersResponse{
							Response: &extProcPb.CommonResponse{
								HeaderMutation: &extProcPb.HeaderMutation{
									SetHeaders: []*corev3.HeaderValueOption{
										makeHeaderValueOption("z-header", "z"),
										makeHeaderValueOption("a-header", "a"),
									},
								},
							},
						},
					},
				},
				{Response: nil},
				{
					Response: &extProcPb.ProcessingResponse_RequestBody{
						RequestBody: &extProcPb.BodyResponse{
							Response: &extProcPb.CommonResponse{
								HeaderMutation: &extProcPb.HeaderMutation{
									SetHeaders: []*corev3.HeaderValueOption{
										makeHeaderValueOption("c-header", "c"),
										makeHeaderValueOption("b-header", "b"),
									},
								},
							},
						},
					},
				},
			},
			want: []*extProcPb.ProcessingResponse{
				nil,
				{
					Response: &extProcPb.ProcessingResponse_RequestHeaders{
						RequestHeaders: &extProcPb.HeadersResponse{
							Response: &extProcPb.CommonResponse{
								HeaderMutation: &extProcPb.HeaderMutation{
									SetHeaders: []*corev3.HeaderValueOption{
										makeHeaderValueOption("a-header", "a"),
										makeHeaderValueOption("z-header", "z"),
									},
								},
							},
						},
					},
				},
				{Response: nil},
				{
					Response: &extProcPb.ProcessingResponse_RequestBody{
						RequestBody: &extProcPb.BodyResponse{
							Response: &extProcPb.CommonResponse{
								HeaderMutation: &extProcPb.HeaderMutation{
									SetHeaders: []*corev3.HeaderValueOption{
										makeHeaderValueOption("b-header", "b"),
										makeHeaderValueOption("c-header", "c"),
									},
								},
							},
						},
					},
				},
			},
		},
		{
			name: "multiple response objects are sorted independently",
			responses: []*extProcPb.ProcessingResponse{
				{
					Response: &extProcPb.ProcessingResponse_RequestHeaders{
						RequestHeaders: &extProcPb.HeadersResponse{
							Response: &extProcPb.CommonResponse{
								HeaderMutation: &extProcPb.HeaderMutation{
									SetHeaders: []*corev3.HeaderValueOption{
										makeHeaderValueOption("c-header", "c"),
										makeHeaderValueOption("a-header", "a"),
									},
								},
							},
						},
					},
				},
				{
					Response: &extProcPb.ProcessingResponse_RequestBody{
						RequestBody: &extProcPb.BodyResponse{
							Response: &extProcPb.CommonResponse{
								HeaderMutation: &extProcPb.HeaderMutation{
									SetHeaders: []*corev3.HeaderValueOption{
										makeHeaderValueOption("z-header", "z"),
										makeHeaderValueOption("b-header", "b"),
									},
								},
							},
						},
					},
				},
			},
			want: []*extProcPb.ProcessingResponse{
				{
					Response: &extProcPb.ProcessingResponse_RequestHeaders{
						RequestHeaders: &extProcPb.HeadersResponse{
							Response: &extProcPb.CommonResponse{
								HeaderMutation: &extProcPb.HeaderMutation{
									SetHeaders: []*corev3.HeaderValueOption{
										makeHeaderValueOption("a-header", "a"),
										makeHeaderValueOption("c-header", "c"),
									},
								},
							},
						},
					},
				},
				{
					Response: &extProcPb.ProcessingResponse_RequestBody{
						RequestBody: &extProcPb.BodyResponse{
							Response: &extProcPb.CommonResponse{
								HeaderMutation: &extProcPb.HeaderMutation{
									SetHeaders: []*corev3.HeaderValueOption{
										makeHeaderValueOption("b-header", "b"),
										makeHeaderValueOption("z-header", "z"),
									},
								},
							},
						},
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			SortSetHeadersInResponses(tt.responses)
			if diff := cmp.Diff(tt.want, tt.responses, protocmp.Transform()); diff != "" {
				t.Errorf("SortSetHeadersInResponses() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
