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

func validTracingProvider() xgatewayv1alpha1.TracingProvider {
	return xgatewayv1alpha1.TracingProvider{
		BackendRef: gatewayv1.BackendObjectReference{
			Name: "otel-collector",
			Port: new(gatewayv1.PortNumber(4317)),
		},
	}
}

func TestXTelemetryPolicyTracingMode(t *testing.T) {
	tests := []struct {
		name       string
		tracing    xgatewayv1alpha1.TracingConfig
		wantErrors []string
	}{
		{
			name: "mode Enabled with provider and sibling fields is accepted",
			tracing: xgatewayv1alpha1.TracingConfig{
				Mode:     xgatewayv1alpha1.TracingModeEnabled,
				Provider: validTracingProvider(),
				SamplingRate: gatewayv1.Fraction{
					Numerator: 50,
				},
				ParentBasedSampling: xgatewayv1alpha1.ParentBasedSampling{
					Mode: xgatewayv1alpha1.ParentBasedSamplingModeEnabled,
					SamplingRate: gatewayv1.Fraction{
						Numerator: 50,
					},
				},
				ServiceName: "my-gateway-service",
				SpanName:    "HTTP GET",
				Attributes: []xgatewayv1alpha1.Attribute{
					{
						Name:         "env",
						SourceType:   xgatewayv1alpha1.AttributeSourceLiteral,
						LiteralValue: "prod",
					},
				},
			},
			wantErrors: []string{},
		},
		{
			name: "mode Enabled without provider is rejected",
			tracing: xgatewayv1alpha1.TracingConfig{
				Mode: xgatewayv1alpha1.TracingModeEnabled,
			},
			wantErrors: []string{"provider is required when mode is Enabled, and must be empty otherwise"},
		},
		{
			name: "mode Disabled without sibling fields is accepted",
			tracing: xgatewayv1alpha1.TracingConfig{
				Mode: xgatewayv1alpha1.TracingModeDisabled,
			},
			wantErrors: []string{},
		},
		{
			name: "mode Disabled with provider is rejected",
			tracing: xgatewayv1alpha1.TracingConfig{
				Mode:     xgatewayv1alpha1.TracingModeDisabled,
				Provider: validTracingProvider(),
			},
			wantErrors: []string{"provider is required when mode is Enabled, and must be empty otherwise"},
		},
		{
			name: "mode Disabled with samplingRate is rejected",
			tracing: xgatewayv1alpha1.TracingConfig{
				Mode: xgatewayv1alpha1.TracingModeDisabled,
				SamplingRate: gatewayv1.Fraction{
					Numerator: 10,
				},
			},
			wantErrors: []string{"samplingRate can only be set when mode is Enabled"},
		},
		{
			name: "mode Disabled with parentBasedSampling is rejected",
			tracing: xgatewayv1alpha1.TracingConfig{
				Mode: xgatewayv1alpha1.TracingModeDisabled,
				ParentBasedSampling: xgatewayv1alpha1.ParentBasedSampling{
					Mode: xgatewayv1alpha1.ParentBasedSamplingModeEnabled,
				},
			},
			wantErrors: []string{"parentBasedSampling can only be set when mode is Enabled"},
		},
		{
			name: "mode Disabled with serviceName is rejected",
			tracing: xgatewayv1alpha1.TracingConfig{
				Mode:        xgatewayv1alpha1.TracingModeDisabled,
				ServiceName: "my-service",
			},
			wantErrors: []string{"serviceName can only be set when mode is Enabled"},
		},
		{
			name: "mode Disabled with spanName is rejected",
			tracing: xgatewayv1alpha1.TracingConfig{
				Mode:     xgatewayv1alpha1.TracingModeDisabled,
				SpanName: "my-span",
			},
			wantErrors: []string{"spanName can only be set when mode is Enabled"},
		},
		{
			name: "mode Disabled with attributes is rejected",
			tracing: xgatewayv1alpha1.TracingConfig{
				Mode: xgatewayv1alpha1.TracingModeDisabled,
				Attributes: []xgatewayv1alpha1.Attribute{
					{
						Name:         "env",
						SourceType:   xgatewayv1alpha1.AttributeSourceLiteral,
						LiteralValue: "prod",
					},
				},
			},
			wantErrors: []string{"attributes can only be set when mode is Enabled"},
		},
		{
			name: "mode ImplementationDefault without sibling fields is accepted",
			tracing: xgatewayv1alpha1.TracingConfig{
				Mode: xgatewayv1alpha1.TracingModeImplementationDefault,
			},
			wantErrors: []string{},
		},
		{
			name: "mode ImplementationDefault with provider is rejected",
			tracing: xgatewayv1alpha1.TracingConfig{
				Mode:     xgatewayv1alpha1.TracingModeImplementationDefault,
				Provider: validTracingProvider(),
			},
			wantErrors: []string{"provider is required when mode is Enabled, and must be empty otherwise"},
		},
		{
			name: "mode ImplementationDefault with samplingRate is rejected",
			tracing: xgatewayv1alpha1.TracingConfig{
				Mode: xgatewayv1alpha1.TracingModeImplementationDefault,
				SamplingRate: gatewayv1.Fraction{
					Numerator: 10,
				},
			},
			wantErrors: []string{"samplingRate can only be set when mode is Enabled"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			policy := &xgatewayv1alpha1.XTelemetryPolicy{
				ObjectMeta: metav1.ObjectMeta{
					Name:      fmt.Sprintf("foo-%v", time.Now().UnixNano()),
					Namespace: metav1.NamespaceDefault,
				},
				Spec: xgatewayv1alpha1.TelemetryPolicySpec{
					TargetRefs: []gatewayv1.LocalObjectReference{
						{
							Group: gatewayv1.GroupName,
							Kind:  "Gateway",
							Name:  "my-gateway",
						},
					},
					Tracing: tc.tracing,
				},
			}
			validateXTelemetryPolicy(t, policy, tc.wantErrors)
		})
	}
}

func TestXTelemetryPolicyParentBasedSampling(t *testing.T) {
	tests := []struct {
		name                string
		parentBasedSampling xgatewayv1alpha1.ParentBasedSampling
		wantErrors          []string
	}{
		{
			name: "parentBasedSampling mode Enabled without samplingRate is accepted",
			parentBasedSampling: xgatewayv1alpha1.ParentBasedSampling{
				Mode: xgatewayv1alpha1.ParentBasedSamplingModeEnabled,
			},
			wantErrors: []string{},
		},
		{
			name: "parentBasedSampling mode Enabled with samplingRate is accepted",
			parentBasedSampling: xgatewayv1alpha1.ParentBasedSampling{
				Mode: xgatewayv1alpha1.ParentBasedSamplingModeEnabled,
				SamplingRate: gatewayv1.Fraction{
					Numerator: 50,
				},
			},
			wantErrors: []string{},
		},
		{
			name: "parentBasedSampling mode Disabled without samplingRate is accepted",
			parentBasedSampling: xgatewayv1alpha1.ParentBasedSampling{
				Mode: xgatewayv1alpha1.ParentBasedSamplingModeDisabled,
			},
			wantErrors: []string{},
		},
		{
			name: "parentBasedSampling mode Disabled with samplingRate is rejected",
			parentBasedSampling: xgatewayv1alpha1.ParentBasedSampling{
				Mode: xgatewayv1alpha1.ParentBasedSamplingModeDisabled,
				SamplingRate: gatewayv1.Fraction{
					Numerator: 50,
				},
			},
			wantErrors: []string{"samplingRate can only be set when mode is Enabled"},
		},
		{
			name: "parentBasedSampling mode ImplementationDefault without samplingRate is accepted",
			parentBasedSampling: xgatewayv1alpha1.ParentBasedSampling{
				Mode: xgatewayv1alpha1.ParentBasedSamplingModeImplementationDefault,
			},
			wantErrors: []string{},
		},
		{
			name: "parentBasedSampling mode ImplementationDefault with samplingRate is rejected",
			parentBasedSampling: xgatewayv1alpha1.ParentBasedSampling{
				Mode: xgatewayv1alpha1.ParentBasedSamplingModeImplementationDefault,
				SamplingRate: gatewayv1.Fraction{
					Numerator: 50,
				},
			},
			wantErrors: []string{"samplingRate can only be set when mode is Enabled"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			policy := &xgatewayv1alpha1.XTelemetryPolicy{
				ObjectMeta: metav1.ObjectMeta{
					Name:      fmt.Sprintf("foo-%v", time.Now().UnixNano()),
					Namespace: metav1.NamespaceDefault,
				},
				Spec: xgatewayv1alpha1.TelemetryPolicySpec{
					TargetRefs: []gatewayv1.LocalObjectReference{
						{
							Group: gatewayv1.GroupName,
							Kind:  "Gateway",
							Name:  "my-gateway",
						},
					},
					Tracing: xgatewayv1alpha1.TracingConfig{
						Mode:                xgatewayv1alpha1.TracingModeEnabled,
						Provider:            validTracingProvider(),
						ParentBasedSampling: tc.parentBasedSampling,
					},
				},
			}
			validateXTelemetryPolicy(t, policy, tc.wantErrors)
		})
	}
}

func validateXTelemetryPolicy(t *testing.T, policy *xgatewayv1alpha1.XTelemetryPolicy, wantErrors []string) {
	t.Helper()

	ctx := context.Background()
	err := k8sClient.Create(ctx, policy)

	if (len(wantErrors) != 0) != (err != nil) {
		t.Fatalf("Unexpected response while creating XTelemetryPolicy %q; got err=\n%v\n;want error=%v", fmt.Sprintf("%v/%v", policy.Namespace, policy.Name), err, wantErrors)
	}

	var missingErrorStrings []string
	for _, wantError := range wantErrors {
		if !celErrorStringMatches(err.Error(), wantError) {
			missingErrorStrings = append(missingErrorStrings, wantError)
		}
	}
	if len(missingErrorStrings) != 0 {
		t.Errorf("Unexpected response while creating XTelemetryPolicy %q; got err=\n%v\n;missing strings within error=%q", fmt.Sprintf("%v/%v", policy.Namespace, policy.Name), err, missingErrorStrings)
	}
}
