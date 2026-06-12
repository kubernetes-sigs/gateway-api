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

package echo

import (
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseResponse(t *testing.T) {
	tests := []struct {
		name   string
		output string
		check  func(t *testing.T, r Response)
	}{
		{
			name:   "empty input returns zero-value response with initialized maps",
			output: "",
			check: func(t *testing.T, r Response) {
				assert.Empty(t, r.ID)
				assert.Empty(t, r.Host)
				assert.Empty(t, r.Code)
				assert.NotNil(t, r.RequestHeaders)
				assert.NotNil(t, r.ResponseHeaders)
			},
		},
		{
			name: "parses all scalar fields from echo output",
			output: strings.Join([]string{
				"X-Request-Id=test-id-123",
				"Method=GET",
				"Proto=HTTP/1.1",
				"Alpn=h2",
				"ServiceVersion=v1",
				"ServicePort=8080",
				"StatusCode=200",
				"Host=example.com",
				"Hostname=echo-pod-abc",
				"URL=/status/200",
				"Cluster=kind-cluster",
				"IP=10.244.0.5",
			}, "\n"),
			check: func(t *testing.T, r Response) {
				assert.Equal(t, "test-id-123", r.ID)
				assert.Equal(t, "GET", r.Method)
				assert.Equal(t, "HTTP/1.1", r.Protocol)
				assert.Equal(t, "h2", r.Alpn)
				assert.Equal(t, "v1", r.Version)
				assert.Equal(t, "8080", r.Port)
				assert.Equal(t, "200", r.Code)
				assert.Equal(t, "example.com", r.Host)
				assert.Equal(t, "echo-pod-abc", r.Hostname)
				assert.Equal(t, "/status/200", r.URL)
				assert.Equal(t, "kind-cluster", r.Cluster)
				assert.Equal(t, "10.244.0.5", r.IP)
			},
		},
		{
			name:   "request ID match is case-insensitive",
			output: "x-request-id=lowercase-id",
			check: func(t *testing.T, r Response) {
				assert.Equal(t, "lowercase-id", r.ID)
			},
		},
		{
			name: "request headers accumulate multiple values for the same key",
			output: strings.Join([]string{
				"RequestHeader=X-Forwarded-For:10.0.0.1",
				"RequestHeader=X-Forwarded-For:10.0.0.2",
			}, "\n"),
			check: func(t *testing.T, r Response) {
				vals := r.RequestHeaders.Values("X-Forwarded-For")
				assert.Equal(t, []string{"10.0.0.1", "10.0.0.2"}, vals)
			},
		},
		{
			name: "response headers keep only the last value for a key",
			output: strings.Join([]string{
				"ResponseHeader=X-Custom:first",
				"ResponseHeader=X-Custom:second",
			}, "\n"),
			check: func(t *testing.T, r Response) {
				assert.Equal(t, "second", r.ResponseHeaders.Get("X-Custom"))
			},
		},
		{
			name:   "header line without colon separator is skipped",
			output: "RequestHeader=MalformedNoColon",
			check: func(t *testing.T, r Response) {
				assert.Empty(t, r.RequestHeaders)
			},
		},
		{
			name: "header value containing colons preserves everything after first colon",
			output: strings.Join([]string{
				"RequestHeader=Authorization:Bearer token:with:colons",
			}, "\n"),
			check: func(t *testing.T, r Response) {
				assert.Equal(t, "Bearer token:with:colons", r.RequestHeaders.Get("Authorization"))
			},
		},
		{
			name: "body lines with body prefix are parsed as key-value pairs",
			output: strings.Join([]string{
				"[1 body] color=blue",
				"[1 body] size=large",
			}, "\n"),
			check: func(t *testing.T, r Response) {
				body := r.Body()
				assert.Equal(t, []string{"blue", "large"}, body)
			},
		},
		{
			name:   "lines without body prefix are ignored for body parsing",
			output: "NotABodyLine=value",
			check: func(t *testing.T, r Response) {
				assert.Empty(t, r.Body())
			},
		},
		{
			name:   "missing fields remain empty strings",
			output: "StatusCode=503",
			check: func(t *testing.T, r Response) {
				assert.Equal(t, "503", r.Code)
				assert.Empty(t, r.Host)
				assert.Empty(t, r.Hostname)
				assert.Empty(t, r.Method)
				assert.Empty(t, r.ID)
			},
		},
		{
			name:   "RawContent preserves original input",
			output: "StatusCode=200\nHost=example.com",
			check: func(t *testing.T, r Response) {
				assert.Equal(t, "StatusCode=200\nHost=example.com", r.RawContent)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := ParseResponse(tt.output)
			tt.check(t, r)
		})
	}
}

func Test_parseMultipleResponses(t *testing.T) {
	tests := []struct {
		name      string
		output    string
		wantCount int
		check     func(t *testing.T, responses []Response)
	}{
		{
			name:      "empty input returns no responses",
			output:    "",
			wantCount: 0,
		},
		{
			name:      "single response is parsed",
			output:    "Hostname=pod-1\nStatusCode=200",
			wantCount: 1,
			check: func(t *testing.T, responses []Response) {
				assert.Equal(t, "pod-1", responses[0].Hostname)
				assert.Equal(t, "200", responses[0].Code)
			},
		},
		{
			name:      "responses are split by double newline",
			output:    "Hostname=pod-1\nStatusCode=200\n\nHostname=pod-2\nStatusCode=201",
			wantCount: 2,
			check: func(t *testing.T, responses []Response) {
				assert.Equal(t, "pod-1", responses[0].Hostname)
				assert.Equal(t, "pod-2", responses[1].Hostname)
			},
		},
		{
			name:      "sections without hostname or code are filtered out",
			output:    "Hostname=pod-1\n\nMethod=GET\n\nStatusCode=200",
			wantCount: 2,
			check: func(t *testing.T, responses []Response) {
				assert.Equal(t, "pod-1", responses[0].Hostname)
				assert.Equal(t, "200", responses[1].Code)
			},
		},
		{
			name:      "blank-only sections are skipped",
			output:    "Hostname=pod-1\n\n   \n\nHostname=pod-2",
			wantCount: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			responses := parseMultipleResponses(tt.output)
			assert.Len(t, responses, tt.wantCount)
			if tt.check != nil {
				tt.check(t, responses)
			}
		})
	}
}

func TestResponse_GetHeaders(t *testing.T) {
	reqHeaders := http.Header{"X-Req": {"val1"}}
	respHeaders := http.Header{"X-Resp": {"val2"}}
	r := Response{
		RequestHeaders:  reqHeaders,
		ResponseHeaders: respHeaders,
	}

	assert.Equal(t, reqHeaders, r.GetHeaders(RequestHeader))
	assert.Equal(t, respHeaders, r.GetHeaders(ResponseHeader))
	assert.Panics(t, func() {
		r.GetHeaders(HeaderType("bogus"))
	})
}

func TestResponse_Body(t *testing.T) {
	tests := []struct {
		name string
		body map[string]string
		want []string
	}{
		{
			name: "empty body returns nil",
			body: map[string]string{},
			want: nil,
		},
		{
			name: "single entry",
			body: map[string]string{"0": "hello"},
			want: []string{"hello"},
		},
		{
			name: "entries are sorted by key",
			body: map[string]string{"2": "c", "0": "a", "1": "b"},
			want: []string{"a", "b", "c"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := Response{rawBody: tt.body}
			assert.Equal(t, tt.want, r.Body())
		})
	}
}
