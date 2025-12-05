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
			wantErrors: []string{"Too long: may not be longer than 63"},
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
