//go:build experimental
// +build experimental

/*
Copyright The Kubernetes Authors.

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

package main

import (
	"context"
	"fmt"
	"testing"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	xgatewayv1alpha1 "sigs.k8s.io/gateway-api/apisx/v1alpha1"
)

func TestXBackendSpec(t *testing.T) {
	tests := []struct {
		name       string
		spec       xgatewayv1alpha1.BackendSpec
		wantErrors []string
	}{
		{
			name: "protocol H2C without tls",
			spec: xgatewayv1alpha1.BackendSpec{
				Type: xgatewayv1alpha1.BackendTypeExternalHostname,
				Port: xgatewayv1alpha1.BackendPort{Port: 8080},
				ExternalHostname: &xgatewayv1alpha1.ExternalHostnameBackend{
					Hostname: "example.com",
				},
				Protocol: new(xgatewayv1alpha1.BackendProtocolH2C),
				TLS:      nil,
			},
			wantErrors: []string{},
		},
		{
			name: "protocol H2C with tls",
			spec: xgatewayv1alpha1.BackendSpec{
				Type: xgatewayv1alpha1.BackendTypeExternalHostname,
				Port: xgatewayv1alpha1.BackendPort{Port: 8080},
				ExternalHostname: &xgatewayv1alpha1.ExternalHostnameBackend{
					Hostname: "example.com",
				},
				Protocol: new(xgatewayv1alpha1.BackendProtocolH2C),
				TLS:      &xgatewayv1alpha1.BackendTLS{Mode: xgatewayv1alpha1.BackendTLSModeNone},
			},
			wantErrors: []string{"tls must be unset when protocol is H2C, use protocol HTTP2 for HTTP/2 with tls"},
		},
		{
			name: "protocol HTTP2 without tls",
			spec: xgatewayv1alpha1.BackendSpec{
				Type: xgatewayv1alpha1.BackendTypeExternalHostname,
				Port: xgatewayv1alpha1.BackendPort{Port: 8080},
				ExternalHostname: &xgatewayv1alpha1.ExternalHostnameBackend{
					Hostname: "example.com",
				},
				Protocol: new(xgatewayv1alpha1.BackendProtocolHTTP2),
				TLS:      nil,
			},
			wantErrors: []string{"tls must be set when protocol is HTTP2, use protocol H2C for HTTP/2 without tls"},
		},
		{
			name: "protocol HTTP2 with tls",
			spec: xgatewayv1alpha1.BackendSpec{
				Type: xgatewayv1alpha1.BackendTypeExternalHostname,
				Port: xgatewayv1alpha1.BackendPort{Port: 8080},
				ExternalHostname: &xgatewayv1alpha1.ExternalHostnameBackend{
					Hostname: "example.com",
				},
				Protocol: new(xgatewayv1alpha1.BackendProtocolHTTP2),
				TLS:      &xgatewayv1alpha1.BackendTLS{Mode: xgatewayv1alpha1.BackendTLSModeNone},
			},
			wantErrors: []string{},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			backend := &xgatewayv1alpha1.XBackend{
				ObjectMeta: metav1.ObjectMeta{
					Name:      fmt.Sprintf("foo-%v", time.Now().UnixNano()),
					Namespace: metav1.NamespaceDefault,
				},
				Spec: tc.spec,
			}
			validateXBackend(t, backend, tc.wantErrors)
		})
	}
}

func validateXBackend(t *testing.T, backend *xgatewayv1alpha1.XBackend, wantErrors []string) {
	t.Helper()

	ctx := context.Background()
	err := k8sClient.Create(ctx, backend)

	if (len(wantErrors) != 0) != (err != nil) {
		t.Fatalf("Unexpected response while creating XBackend %q; got err=\n%v\n;want error=%v", fmt.Sprintf("%v/%v", backend.Namespace, backend.Name), err, wantErrors)
	}

	var missingErrorStrings []string
	for _, wantError := range wantErrors {
		if !celErrorStringMatches(err.Error(), wantError) {
			missingErrorStrings = append(missingErrorStrings, wantError)
		}
	}
	if len(missingErrorStrings) != 0 {
		t.Errorf("Unexpected response while creating XBackend %q; got err=\n%v\n;missing strings within error=%q", fmt.Sprintf("%v/%v", backend.Namespace, backend.Name), err, missingErrorStrings)
	}
}
