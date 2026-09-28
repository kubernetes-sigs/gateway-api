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
	gatewayv1 "sigs.k8s.io/gateway-api/apis/v1"
	xgatewayv1alpha1 "sigs.k8s.io/gateway-api/apisx/v1alpha1"
)

func TestXBackendSpec(t *testing.T) {
	tests := []struct {
		name       string
		spec       xgatewayv1alpha1.BackendSpec
		wantErrors []string
	}{
		{
			name: "port without number is rejected",
			spec: xgatewayv1alpha1.BackendSpec{
				Type: xgatewayv1alpha1.BackendTypeExternalHostname,
				ExternalHostname: &xgatewayv1alpha1.ExternalHostnameBackend{
					Hostname: "example.com",
				},
				Port: xgatewayv1alpha1.BackendPort{},
			},
			wantErrors: []string{"should have at least 1 properties"},
		},
		{
			name: "port with number is accepted",
			spec: xgatewayv1alpha1.BackendSpec{
				Type: xgatewayv1alpha1.BackendTypeExternalHostname,
				ExternalHostname: &xgatewayv1alpha1.ExternalHostnameBackend{
					Hostname: "example.com",
				},
				Port: xgatewayv1alpha1.BackendPort{Number: xgatewayv1alpha1.PortNumber(8080)},
			},
			wantErrors: []string{},
		},
		{
			name: "protocol H2C without tls is accepted",
			spec: xgatewayv1alpha1.BackendSpec{
				Type: xgatewayv1alpha1.BackendTypeExternalHostname,
				Port: xgatewayv1alpha1.BackendPort{Number: xgatewayv1alpha1.PortNumber(8080)},
				ExternalHostname: &xgatewayv1alpha1.ExternalHostnameBackend{
					Hostname: "example.com",
				},
				Protocol: new(xgatewayv1alpha1.BackendProtocolH2C),
				TLS:      nil,
			},
			wantErrors: []string{},
		},
		{
			name: "protocol H2C with tls mode None is accepted",
			spec: xgatewayv1alpha1.BackendSpec{
				Type: xgatewayv1alpha1.BackendTypeExternalHostname,
				Port: xgatewayv1alpha1.BackendPort{Number: xgatewayv1alpha1.PortNumber(8080)},
				ExternalHostname: &xgatewayv1alpha1.ExternalHostnameBackend{
					Hostname: "example.com",
				},
				Protocol: new(xgatewayv1alpha1.BackendProtocolH2C),
				TLS: &xgatewayv1alpha1.BackendTLS{
					Mode: xgatewayv1alpha1.BackendTLSModeNone,
				},
			},
			wantErrors: []string{},
		},
		{
			name: "protocol H2C with tls mode ServerOnly is rejected",
			spec: xgatewayv1alpha1.BackendSpec{
				Type: xgatewayv1alpha1.BackendTypeExternalHostname,
				Port: xgatewayv1alpha1.BackendPort{Number: xgatewayv1alpha1.PortNumber(8080)},
				ExternalHostname: &xgatewayv1alpha1.ExternalHostnameBackend{
					Hostname: "example.com",
				},
				Protocol: new(xgatewayv1alpha1.BackendProtocolH2C),
				TLS: &xgatewayv1alpha1.BackendTLS{
					Mode: xgatewayv1alpha1.BackendTLSModeServerOnly,
				},
			},
			wantErrors: []string{"tls must be disabled when protocol is H2C, use protocol HTTP2 for HTTP/2 with tls"},
		},
		{
			name: "protocol HTTP2 without tls is rejected",
			spec: xgatewayv1alpha1.BackendSpec{
				Type: xgatewayv1alpha1.BackendTypeExternalHostname,
				Port: xgatewayv1alpha1.BackendPort{Number: xgatewayv1alpha1.PortNumber(8080)},
				ExternalHostname: &xgatewayv1alpha1.ExternalHostnameBackend{
					Hostname: "example.com",
				},
				Protocol: new(xgatewayv1alpha1.BackendProtocolHTTP2),
				TLS:      nil,
			},
			wantErrors: []string{"tls must be enabled when protocol is HTTP2, use protocol H2C for HTTP/2 without tls"},
		},
		{
			name: "protocol HTTP2 with tls mode None is rejected",
			spec: xgatewayv1alpha1.BackendSpec{
				Type: xgatewayv1alpha1.BackendTypeExternalHostname,
				Port: xgatewayv1alpha1.BackendPort{Number: xgatewayv1alpha1.PortNumber(8080)},
				ExternalHostname: &xgatewayv1alpha1.ExternalHostnameBackend{
					Hostname: "example.com",
				},
				Protocol: new(xgatewayv1alpha1.BackendProtocolHTTP2),
				TLS: &xgatewayv1alpha1.BackendTLS{
					Mode: xgatewayv1alpha1.BackendTLSModeNone,
				},
			},
			wantErrors: []string{"tls must be enabled when protocol is HTTP2, use protocol H2C for HTTP/2 without tls"},
		},
		{
			name: "protocol HTTP2 with tls mode ServerOnly is accepted",
			spec: xgatewayv1alpha1.BackendSpec{
				Type: xgatewayv1alpha1.BackendTypeExternalHostname,
				Port: xgatewayv1alpha1.BackendPort{Number: xgatewayv1alpha1.PortNumber(8080)},
				ExternalHostname: &xgatewayv1alpha1.ExternalHostnameBackend{
					Hostname: "example.com",
				},
				Protocol: new(xgatewayv1alpha1.BackendProtocolHTTP2),
				TLS: &xgatewayv1alpha1.BackendTLS{
					Mode: xgatewayv1alpha1.BackendTLSModeServerOnly,
				},
			},
			wantErrors: []string{},
		},
		{
			name: "protocol WSS without tls is rejected",
			spec: xgatewayv1alpha1.BackendSpec{
				Type: xgatewayv1alpha1.BackendTypeExternalHostname,
				Port: xgatewayv1alpha1.BackendPort{Number: xgatewayv1alpha1.PortNumber(8080)},
				ExternalHostname: &xgatewayv1alpha1.ExternalHostnameBackend{
					Hostname: "example.com",
				},
				Protocol: new(xgatewayv1alpha1.BackendProtocolWSS),
				TLS:      nil,
			},
			wantErrors: []string{"tls must be enabled when protocol is WSS"},
		},
		{
			name: "protocol WSS with tls mode None is rejected",
			spec: xgatewayv1alpha1.BackendSpec{
				Type: xgatewayv1alpha1.BackendTypeExternalHostname,
				Port: xgatewayv1alpha1.BackendPort{Number: xgatewayv1alpha1.PortNumber(8080)},
				ExternalHostname: &xgatewayv1alpha1.ExternalHostnameBackend{
					Hostname: "example.com",
				},
				Protocol: new(xgatewayv1alpha1.BackendProtocolWSS),
				TLS: &xgatewayv1alpha1.BackendTLS{
					Mode: xgatewayv1alpha1.BackendTLSModeNone,
				},
			},
			wantErrors: []string{"tls must be enabled when protocol is WSS"},
		},
		{
			name: "protocol WSS with tls mode ServerOnly is accepted",
			spec: xgatewayv1alpha1.BackendSpec{
				Type: xgatewayv1alpha1.BackendTypeExternalHostname,
				Port: xgatewayv1alpha1.BackendPort{Number: xgatewayv1alpha1.PortNumber(8080)},
				ExternalHostname: &xgatewayv1alpha1.ExternalHostnameBackend{
					Hostname: "example.com",
				},
				Protocol: new(xgatewayv1alpha1.BackendProtocolWSS),
				TLS: &xgatewayv1alpha1.BackendTLS{
					Mode: xgatewayv1alpha1.BackendTLSModeServerOnly,
				},
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

func TestXBackendSessionPersistence(t *testing.T) {
	tests := []struct {
		name       string
		spec       xgatewayv1alpha1.BackendSpec
		wantErrors []string
	}{
		{
			name: "session persistence is supported for EndpointSelector",
			spec: xgatewayv1alpha1.BackendSpec{
				Type: xgatewayv1alpha1.BackendTypeEndpointSelector,
				EndpointSelector: &xgatewayv1alpha1.EndpointSelectorBackend{
					LabelSelector: xgatewayv1alpha1.LabelSelector{
						MatchLabels: map[gatewayv1.LabelKey]gatewayv1.LabelValue{
							"app": "foo",
						},
					},
				},
				Port: xgatewayv1alpha1.BackendPort{
					Number: xgatewayv1alpha1.PortNumber(80),
				},
				SessionPersistence: &xgatewayv1alpha1.SessionPersistence{
					Cookie: &gatewayv1.CookieConfig{},
				},
			},
			wantErrors: []string{},
		},
		{
			name: "session persistence is rejected for ExternalHostname",
			spec: xgatewayv1alpha1.BackendSpec{
				Type: xgatewayv1alpha1.BackendTypeExternalHostname,
				ExternalHostname: &xgatewayv1alpha1.ExternalHostnameBackend{
					Hostname: "example.com",
				},
				Port: xgatewayv1alpha1.BackendPort{
					Number: xgatewayv1alpha1.PortNumber(80),
				},
				SessionPersistence: &xgatewayv1alpha1.SessionPersistence{
					Cookie: &gatewayv1.CookieConfig{},
				},
			},
			wantErrors: []string{"sessionPersistence must be unset when type is not EndpointSelector"},
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
