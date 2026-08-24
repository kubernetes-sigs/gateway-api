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

func TestXBackendBa(t *testing.T) {
	tests := []struct {
		name       string
		wantErrors []string
		portName   *string
	}{
		{
			name:       "no port name",
			portName:   nil,
			wantErrors: []string{},
		},
		{
			name:       "empty port name",
			portName:   new(""),
			wantErrors: []string{},
		},
		{
			name:       "simple lowercase port name",
			portName:   new("http"),
			wantErrors: []string{},
		},
		{
			name:       "port name with whitespace is rejected",
			portName:   new("my port"),
			wantErrors: []string{"Name must be a valid DNS label"},
		},
		{
			name:       "port name with uppercase characters are rejected",
			portName:   new("HTTP"),
			wantErrors: []string{"Name must be a valid DNS label"},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			backend := &xgatewayv1alpha1.XBackend{
				ObjectMeta: metav1.ObjectMeta{
					Name:      fmt.Sprintf("foo-%v", time.Now().UnixNano()),
					Namespace: metav1.NamespaceDefault,
				},
				Spec: xgatewayv1alpha1.BackendSpec{
					Type: xgatewayv1alpha1.BackendTypeExternalHostname,
					ExternalHostname: &xgatewayv1alpha1.ExternalHostnameBackend{
						Hostname: "example.com",
					},
					Port: xgatewayv1alpha1.BackendPort{
						Name: tc.portName,
						Port: 80,
					},
				},
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
