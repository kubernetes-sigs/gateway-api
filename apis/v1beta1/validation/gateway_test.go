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

	gatewayv1b1 "sigs.k8s.io/gateway-api/apis/v1beta1"
)

func TestValidateGateway(t *testing.T) {
	listeners := []gatewayv1b1.Listener{
		{
			Hostname: nil,
		},
	}
	addresses := []gatewayv1b1.GatewayAddress{
		{
			Type: nil,
		},
	}
	baseGateway := gatewayv1b1.Gateway{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "foo",
			Namespace: metav1.NamespaceDefault,
		},
		Spec: gatewayv1b1.GatewaySpec{
			GatewayClassName: "foo",
			Listeners:        listeners,
			Addresses:        addresses,
		},
	}
	tlsConfig := gatewayv1b1.GatewayTLSConfig{}

	testCases := map[string]struct {
		mutate             func(gw *gatewayv1b1.Gateway)
		expectErrsOnFields []string
	}{
		"tls config present with http protocol": {
			mutate: func(gw *gatewayv1b1.Gateway) {
				gw.Spec.Listeners[0].Protocol = gatewayv1b1.HTTPProtocolType
				gw.Spec.Listeners[0].TLS = &tlsConfig
			},
			expectErrsOnFields: []string{"spec.listeners[0].tls"},
		},
		"tls config present with tcp protocol": {
			mutate: func(gw *gatewayv1b1.Gateway) {
				gw.Spec.Listeners[0].Protocol = gatewayv1b1.TCPProtocolType
				gw.Spec.Listeners[0].TLS = &tlsConfig
			},
			expectErrsOnFields: []string{"spec.listeners[0].tls"},
		},
		"tls config not set with https protocol": {
			mutate: func(gw *gatewayv1b1.Gateway) {
				gw.Spec.Listeners[0].Protocol = gatewayv1b1.HTTPSProtocolType
			},
			expectErrsOnFields: []string{"spec.listeners[0].tls"},
		},
		"tls config not set with tls protocol": {
			mutate: func(gw *gatewayv1b1.Gateway) {
				gw.Spec.Listeners[0].Protocol = gatewayv1b1.TLSProtocolType
			},
			expectErrsOnFields: []string{"spec.listeners[0].tls"},
		},
		"tls config not set with http protocol": {
			mutate: func(gw *gatewayv1b1.Gateway) {
				gw.Spec.Listeners[0].Protocol = gatewayv1b1.HTTPProtocolType
			},
			expectErrsOnFields: nil,
		},
		"tls config not set with tcp protocol": {
			mutate: func(gw *gatewayv1b1.Gateway) {
				gw.Spec.Listeners[0].Protocol = gatewayv1b1.TCPProtocolType
			},
			expectErrsOnFields: nil,
		},
		"tls config not set with udp protocol": {
			mutate: func(gw *gatewayv1b1.Gateway) {
				gw.Spec.Listeners[0].Protocol = gatewayv1b1.UDPProtocolType
			},
			expectErrsOnFields: nil,
		},
		"hostname present with tcp protocol": {
			mutate: func(gw *gatewayv1b1.Gateway) {
				hostname := gatewayv1b1.Hostname("foo.bar.com")
				gw.Spec.Listeners[0].Hostname = &hostname
				gw.Spec.Listeners[0].Protocol = gatewayv1b1.TCPProtocolType
			},
			expectErrsOnFields: []string{"spec.listeners[0].hostname"},
		},
		"hostname present with udp protocol": {
			mutate: func(gw *gatewayv1b1.Gateway) {
				hostname := gatewayv1b1.Hostname("foo.bar.com")
				gw.Spec.Listeners[0].Hostname = &hostname
				gw.Spec.Listeners[0].Protocol = gatewayv1b1.UDPProtocolType
			},
			expectErrsOnFields: []string{"spec.listeners[0].hostname"},
		},
		"certificatedRefs not set with https protocol and TLS terminate mode": {
			mutate: func(gw *gatewayv1b1.Gateway) {
				hostname := gatewayv1b1.Hostname("foo.bar.com")
				tlsMode := gatewayv1b1.TLSModeType("Terminate")
				gw.Spec.Listeners[0].Protocol = gatewayv1b1.HTTPSProtocolType
				gw.Spec.Listeners[0].Hostname = &hostname
				gw.Spec.Listeners[0].TLS = &tlsConfig
				gw.Spec.Listeners[0].TLS.Mode = &tlsMode
			},
			expectErrsOnFields: []string{"spec.listeners[0].tls.certificateRefs"},
		},
		"certificatedRefs not set with tls protocol and TLS terminate mode": {
			mutate: func(gw *gatewayv1b1.Gateway) {
				hostname := gatewayv1b1.Hostname("foo.bar.com")
				tlsMode := gatewayv1b1.TLSModeType("Terminate")
				gw.Spec.Listeners[0].Protocol = gatewayv1b1.TLSProtocolType
				gw.Spec.Listeners[0].Hostname = &hostname
				gw.Spec.Listeners[0].TLS = &tlsConfig
				gw.Spec.Listeners[0].TLS.Mode = &tlsMode
			},
			expectErrsOnFields: []string{"spec.listeners[0].tls.certificateRefs"},
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
