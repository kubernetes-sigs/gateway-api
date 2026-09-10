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

package http

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsKnownRedirectScheme(t *testing.T) {
	testCases := []struct {
		name   string
		scheme string
		want   bool
	}{
		{name: "http is known", scheme: "http", want: true},
		{name: "https is known", scheme: "https", want: true},
		{name: "scheme match is case-insensitive", scheme: "HTTPS", want: true},
		{name: "grpc is not known", scheme: "grpc", want: false},
		{name: "ftp is not known", scheme: "ftp", want: false},
		{name: "empty scheme is not known", scheme: "", want: false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.want, isKnownRedirectScheme(tc.scheme))
		})
	}
}
