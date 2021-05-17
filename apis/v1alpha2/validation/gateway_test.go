/*
Copyright 2021 The Kubernetes Authors.

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

package validation

import (
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	gatewayv1a2 "sigs.k8s.io/gateway-api/apis/v1alpha2"
)

func TestValidateGateway(t *testing.T) {
	listeners := []gatewayv1a2.Listener{
		{
			Hostname: nil,
		},
	}
	baseGateway := gatewayv1a2.Gateway{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "foo",
			Namespace: metav1.NamespaceDefault,
		},
		Spec: gatewayv1a2.GatewaySpec{
			GatewayClassName: "foo",
			Listeners:        listeners,
		},
	}

	testCases := map[string]struct {
		mutate             func(gw *gatewayv1a2.Gateway)
		expectErrsOnFields []string
	}{
		"nil hostname": {
			mutate:             func(gw *gatewayv1a2.Gateway) {},
			expectErrsOnFields: []string{},
		},
		"empty string hostname": {
			mutate: func(gw *gatewayv1a2.Gateway) {
				hostname := gatewayv1a2.Hostname("")
				gw.Spec.Listeners[0].Hostname = &hostname
			},
			expectErrsOnFields: []string{},
		},
		"wildcard hostname": {
			mutate: func(gw *gatewayv1a2.Gateway) {
				hostname := gatewayv1a2.Hostname("*")
				gw.Spec.Listeners[0].Hostname = &hostname
			},
			expectErrsOnFields: []string{},
		},
		"wildcard-prefixed hostname": {
			mutate: func(gw *gatewayv1a2.Gateway) {
				hostname := gatewayv1a2.Hostname("*.example.com")
				gw.Spec.Listeners[0].Hostname = &hostname
			},
			expectErrsOnFields: []string{},
		},
		"valid dns subdomain": {
			mutate: func(gw *gatewayv1a2.Gateway) {
				hostname := gatewayv1a2.Hostname("foo.example.com")
				gw.Spec.Listeners[0].Hostname = &hostname
			},
			expectErrsOnFields: []string{},
		},
		// Invalid use cases
		"IPv4 address hostname": {
			mutate: func(gw *gatewayv1a2.Gateway) {
				hostname := gatewayv1a2.Hostname("1.2.3.4")
				gw.Spec.Listeners[0].Hostname = &hostname
			},
			expectErrsOnFields: []string{"spec.listeners[0].hostname"},
		},
		"Invalid IPv4 address hostname": {
			mutate: func(gw *gatewayv1a2.Gateway) {
				hostname := gatewayv1a2.Hostname("1.2.3..4")
				gw.Spec.Listeners[0].Hostname = &hostname
			},
			expectErrsOnFields: []string{"spec.listeners[0].hostname"},
		},
		"IPv4 address with port hostname": {
			mutate: func(gw *gatewayv1a2.Gateway) {
				hostname := gatewayv1a2.Hostname("1.2.3.4:8080")
				gw.Spec.Listeners[0].Hostname = &hostname
			},
			expectErrsOnFields: []string{"spec.listeners[0].hostname"},
		},
		"IPv6 address hostname": {
			mutate: func(gw *gatewayv1a2.Gateway) {
				hostname := gatewayv1a2.Hostname("2001:db8::68")
				gw.Spec.Listeners[0].Hostname = &hostname
			},
			expectErrsOnFields: []string{"spec.listeners[0].hostname", "spec.listeners[0].hostname"},
		},
		"IPv6 link-local address hostname": {
			mutate: func(gw *gatewayv1a2.Gateway) {
				hostname := gatewayv1a2.Hostname("fe80::/10")
				gw.Spec.Listeners[0].Hostname = &hostname
			},
			expectErrsOnFields: []string{"spec.listeners[0].hostname"},
		},
		"dns subdomain with port": {
			mutate: func(gw *gatewayv1a2.Gateway) {
				hostname := gatewayv1a2.Hostname("foo.example.com:8080")
				gw.Spec.Listeners[0].Hostname = &hostname
			},
			expectErrsOnFields: []string{"spec.listeners[0].hostname"},
		},
		"dns subdomain with invalid wildcard label": {
			mutate: func(gw *gatewayv1a2.Gateway) {
				hostname := gatewayv1a2.Hostname("*.*.com")
				gw.Spec.Listeners[0].Hostname = &hostname
			},
			expectErrsOnFields: []string{"spec.listeners[0].hostname"},
		},
		"dns subdomain with multiple wildcards": {
			mutate: func(gw *gatewayv1a2.Gateway) {
				hostname := gatewayv1a2.Hostname("*.foo.*.com")
				gw.Spec.Listeners[0].Hostname = &hostname
			},
			expectErrsOnFields: []string{"spec.listeners[0].hostname"},
		},
		"dns subdomain with wildcard root label": {
			mutate: func(gw *gatewayv1a2.Gateway) {
				hostname := gatewayv1a2.Hostname("*.foo.*.com")
				gw.Spec.Listeners[0].Hostname = &hostname
			},
			expectErrsOnFields: []string{"spec.listeners[0].hostname"},
		},
	}

	for name, tc := range testCases {
		tc := tc
		t.Run(name, func(t *testing.T) {
			gw := baseGateway.DeepCopy()
			tc.mutate(gw)
			errs := ValidateGateway(gw)
			if len(tc.expectErrsOnFields) != len(errs) {
				t.Fatalf("Expected %d errors, got %d errors: %v", len(tc.expectErrsOnFields), len(errs), errs)
			}
			for i, err := range errs {
				if err.Field != tc.expectErrsOnFields[i] {
					t.Errorf("Expected error on field: %s, got: %s", tc.expectErrsOnFields[i], err.Error())
				}
			}
		})
	}
}
