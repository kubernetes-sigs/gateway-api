//go:build experimental
// +build experimental

/*
Copyright 2023 The Kubernetes Authors.

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
	"strings"
	"testing"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	gatewayv1 "sigs.k8s.io/gateway-api/apis/v1"
)

func TestGatewayInfrastructureLabels(t *testing.T) {
	ctx := context.Background()
	baseGateway := gatewayv1.Gateway{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "foo",
			Namespace: metav1.NamespaceDefault,
		},
		Spec: gatewayv1.GatewaySpec{
			GatewayClassName: "foo",
			Listeners: []gatewayv1.Listener{
				{
					Name:     gatewayv1.SectionName("http"),
					Protocol: gatewayv1.HTTPProtocolType,
					Port:     gatewayv1.PortNumber(80),
				},
			},
		},
	}

	testCases := []struct {
		name       string
		wantErrors []string
		labels     map[gatewayv1.LabelKey]gatewayv1.LabelValue
	}{
		{
			name: "valid label keys and values",
			labels: map[gatewayv1.LabelKey]gatewayv1.LabelValue{
				"app":                   "gateway",
				"tier":                  "frontend",
				"example":               "MyValue",
				"example.com":           "my.name",
				"example.com/path":      "123-my-value",
				"example.com/path.html": "",
			},
		},
		{
			name: "invalid label key with invalid DNS prefix",
			labels: map[gatewayv1.LabelKey]gatewayv1.LabelValue{
				"Example.com/key": "value",
			},
			wantErrors: []string{"Label keys must be in the form of an optional DNS subdomain prefix followed by a required name segment of up to 63 characters"},
		},
		{
			name: "invalid label key with invalid name",
			labels: map[gatewayv1.LabelKey]gatewayv1.LabelValue{
				"key~@@@": "value",
			},
			wantErrors: []string{"Label keys must be in the form of an optional DNS subdomain prefix followed by a required name segment of up to 63 characters"},
		},
		{
			name: "invalid label key with DNS prefix too long",
			labels: map[gatewayv1.LabelKey]gatewayv1.LabelValue{
				gatewayv1.LabelKey(strings.Repeat("a", 254) + "/key"): "value",
			},
			wantErrors: []string{"If specified, the label key's prefix must be a DNS subdomain not longer than 253 characters in total."},
		},
		{
			name: "invalid label key with name too long",
			labels: map[gatewayv1.LabelKey]gatewayv1.LabelValue{
				gatewayv1.LabelKey(strings.Repeat("a", 64)): "value",
			},
			wantErrors: []string{"Label keys must be in the form of an optional DNS subdomain prefix followed by a required name segment of up to 63 characters."},
		},
		{
			name: "invalid label value with too many characters",
			labels: map[gatewayv1.LabelKey]gatewayv1.LabelValue{
				"key": gatewayv1.LabelValue(strings.Repeat("a", 64)),
			},
			wantErrors: []string{"Too long: may not be more than 63"},
		},
		{
			name: "invalid label value with invalid characters",
			labels: map[gatewayv1.LabelKey]gatewayv1.LabelValue{
				"key": "v a l u e",
			},
			wantErrors: []string{"spec.infrastructure.labels.key in body should match '^(([A-Za-z0-9][-A-Za-z0-9_.]*)?[A-Za-z0-9])?$'"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			gw := baseGateway.DeepCopy()
			gw.Name = fmt.Sprintf("foo-%v", time.Now().UnixNano())

			gw.Spec.Infrastructure = &gatewayv1.GatewayInfrastructure{Labels: tc.labels}
			err := k8sClient.Create(ctx, gw)

			if (len(tc.wantErrors) != 0) != (err != nil) {
				t.Fatalf("Unexpected response while creating Gateway; got err=\n%v\n;want error=%v", err, tc.wantErrors != nil)
			}

			var missingErrorStrings []string
			for _, wantError := range tc.wantErrors {
				if !celErrorStringMatches(err.Error(), wantError) {
					missingErrorStrings = append(missingErrorStrings, wantError)
				}
			}
			if len(missingErrorStrings) != 0 {
				t.Errorf("Unexpected response while creating Gateway; got err=\n%v\n;missing strings within error=%q", err, missingErrorStrings)
			}
		})
	}
}

func TestGatewayAddressRoutability(t *testing.T) {
	tests := []struct {
		name          string
		routability   any
		statusAddress bool
		wantError     bool
	}{
		{name: "empty spec", routability: ""},
		{name: "cluster spec", routability: "Cluster"},
		{name: "prefixed spec", routability: "example.com/scope"},
		{name: "sentinel spec", routability: "testing.gateway.networking.k8s.io/sentinel"},
		{name: "unknown spec", routability: "Unknown", wantError: true},
		{name: "empty prefix spec", routability: "/scope", wantError: true},
		{name: "empty path spec", routability: "example.com/", wantError: true},
		{name: "invalid prefix spec", routability: "example..com/scope", wantError: true},
		{name: "reserved spec", routability: "gateway.networking.k8s.io/sentinel", wantError: true},
		{name: "empty status", routability: "", statusAddress: true},
		{name: "cluster status", routability: "Cluster", statusAddress: true},
		{name: "prefixed status", routability: "example.com/scope", statusAddress: true},
		{name: "sentinel status", routability: "testing.gateway.networking.k8s.io/sentinel", statusAddress: true},
		{name: "unknown status", routability: "Unknown", statusAddress: true, wantError: true},
		{name: "empty prefix status", routability: "/scope", statusAddress: true, wantError: true},
		{name: "empty path status", routability: "example.com/", statusAddress: true, wantError: true},
		{name: "invalid prefix status", routability: "example..com/scope", statusAddress: true, wantError: true},
		{name: "reserved status", routability: "gateway.networking.k8s.io/sentinel", statusAddress: true, wantError: true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			name := fmt.Sprintf("routability-%d", time.Now().UnixNano())
			obj := &unstructured.Unstructured{Object: map[string]any{
				"apiVersion": "gateway.networking.k8s.io/v1",
				"kind":       "Gateway",
				"metadata": map[string]any{
					"name":      name,
					"namespace": metav1.NamespaceDefault,
				},
				"spec": map[string]any{
					"gatewayClassName": "foo",
					"addresses": []any{map[string]any{
						"type":        "IPAddress",
						"value":       "1.2.3.4",
						"routability": tc.routability,
					}},
					"listeners": []any{map[string]any{
						"name":     "http",
						"protocol": "HTTP",
						"port":     int64(80),
					}},
				},
			}}
			var status map[string]any
			if tc.statusAddress {
				obj.Object["spec"].(map[string]any)["addresses"] = []any{}
				status = map[string]any{
					"addresses": []any{map[string]any{
						"type":        "IPAddress",
						"value":       "1.2.3.4",
						"routability": tc.routability,
					}},
				}
			}

			ctx := context.Background()
			err := k8sClient.Create(ctx, obj)
			if err == nil && tc.statusAddress {
				obj.Object["status"] = status
				err = k8sClient.Status().Update(ctx, obj)
			}
			if (err != nil) != tc.wantError {
				t.Fatalf("unexpected validation result: got err=%v, want error=%v", err, tc.wantError)
			}
		})
	}
}
