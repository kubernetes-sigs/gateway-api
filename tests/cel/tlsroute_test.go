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
	"strings"
	"testing"
	"time"

	gatewayv1 "sigs.k8s.io/gateway-api/apis/v1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestValidateTLSRouteHostnames(t *testing.T) {
	ctx := context.Background()
	baseTLSRoute := gatewayv1.TLSRoute{
		Name:      "foo",
		Namespace: metav1.NamespaceDefault,
		Spec: gatewayv1.TLSRouteSpec{
			Rules: []gatewayv1.TLSRouteRule{
				{
					BackendRefs: []gatewayv1.BackendRef{
						{
							Name: gatewayv1.ObjectName("backend"),
							Port: new(gatewayv1.PortNumber(443)),
						},
					},
				},
			},
		},
	}

	longestHostname := strings.Join([]string{
		strings.Repeat("a", 63),
		strings.Repeat("b", 63),
		strings.Repeat("c", 63),
		strings.Repeat("d", 61),
	}, ".")

	longLabel := strings.Repeat("a", 64)

	testCases := []struct {
		desc       string
		hostnames  []gatewayv1.Hostname
		wantErrors []string
	}{
		{
			desc:      "valid hostname",
			hostnames: []gatewayv1.Hostname{"foo.example.com"},
		},
		{
			desc:      "valid wildcard hostname",
			hostnames: []gatewayv1.Hostname{"*.example.com"},
		},
		{
			desc:      "valid hostname with 253 characters",
			hostnames: []gatewayv1.Hostname{gatewayv1.Hostname(longestHostname)},
		},
		{
			desc:      "valid hostname with a 64-character label",
			hostnames: []gatewayv1.Hostname{gatewayv1.Hostname(longLabel + ".example.com")},
		},
		{
			desc:      "valid wildcard hostname with a 64-character label",
			hostnames: []gatewayv1.Hostname{gatewayv1.Hostname("*." + longLabel + ".example.com")},
		},
		{
			desc:       "uppercase hostname",
			hostnames:  []gatewayv1.Hostname{"Foo.com"},
			wantErrors: []string{"Hostnames must be valid based on RFC-1123"},
		},
		{
			desc:       "hostname with a leading dash",
			hostnames:  []gatewayv1.Hostname{"-a.com"},
			wantErrors: []string{"Hostnames must be valid based on RFC-1123"},
		},
		{
			desc:       "hostname with an empty label",
			hostnames:  []gatewayv1.Hostname{"foo..bar"},
			wantErrors: []string{"Hostnames must be valid based on RFC-1123"},
		},
		{
			desc:       "wildcard not in the first label",
			hostnames:  []gatewayv1.Hostname{"foo.*.bar"},
			wantErrors: []string{"Wildcards on hostnames must be the first label, and the rest of hostname must be valid based on RFC-1123"},
		},
		{
			desc:       "wildcard as part of the first label",
			hostnames:  []gatewayv1.Hostname{"*foo.bar"},
			wantErrors: []string{"Wildcards on hostnames must be the first label, and the rest of hostname must be valid based on RFC-1123"},
		},
		{
			desc:       "wildcard alone",
			hostnames:  []gatewayv1.Hostname{"*"},
			wantErrors: []string{"Wildcards on hostnames must be the first label, and the rest of hostname must be valid based on RFC-1123"},
		},
		{
			desc:       "wildcard with nothing after it",
			hostnames:  []gatewayv1.Hostname{"*."},
			wantErrors: []string{"Wildcards on hostnames must be the first label, and the rest of hostname must be valid based on RFC-1123"},
		},
		{
			desc:       "IP address",
			hostnames:  []gatewayv1.Hostname{"1.2.3.4"},
			wantErrors: []string{"Hostnames cannot contain an IP"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			route := baseTLSRoute.DeepCopy()
			route.Name = fmt.Sprintf("foo-%v", time.Now().UnixNano())
			route.Spec.Hostnames = tc.hostnames

			err := k8sClient.Create(ctx, route)

			if (len(tc.wantErrors) != 0) != (err != nil) {
				t.Fatalf("Unexpected response while creating TLSRoute; got err=\n%v\n;want error=%v", err, tc.wantErrors != nil)
			}

			var missingErrorStrings []string
			for _, wantError := range tc.wantErrors {
				if !celErrorStringMatches(err.Error(), wantError) {
					missingErrorStrings = append(missingErrorStrings, wantError)
				}
			}
			if len(missingErrorStrings) != 0 {
				t.Errorf("Unexpected response while creating TLSRoute; got err=\n%v\n;missing strings within error=%q", err, missingErrorStrings)
			}
		})
	}
}
