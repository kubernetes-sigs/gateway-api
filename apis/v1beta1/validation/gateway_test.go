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
		"names are not unique within the Gateway": {
			mutate: func(gw *gatewayv1b1.Gateway) {
				hostnameFoo := gatewayv1b1.Hostname("foo.com")
				hostnameBar := gatewayv1b1.Hostname("bar.com")
				gw.Spec.Listeners[0].Name = "foo"
				gw.Spec.Listeners[0].Hostname = &hostnameFoo
				gw.Spec.Listeners = append(gw.Spec.Listeners,
					gatewayv1b1.Listener{
						Name:     "foo",
						Hostname: &hostnameBar,
					},
				)
			},
			expectErrsOnFields: []string{"spec.listeners[1].name"},
		},
		"combination of port, protocol, and hostname are not unique for each listener": {
			mutate: func(gw *gatewayv1b1.Gateway) {
				hostnameFoo := gatewayv1b1.Hostname("foo.com")
				gw.Spec.Listeners[0].Name = "foo"
				gw.Spec.Listeners[0].Hostname = &hostnameFoo
				gw.Spec.Listeners[0].Protocol = gatewayv1b1.HTTPProtocolType
				gw.Spec.Listeners[0].Port = 80
				gw.Spec.Listeners = append(gw.Spec.Listeners,
					gatewayv1b1.Listener{
						Name:     "bar",
						Hostname: &hostnameFoo,
						Protocol: gatewayv1b1.HTTPProtocolType,
						Port:     80,
					},
				)
			},
			expectErrsOnFields: []string{"spec.listeners[1]"},
		},
		"combination of port and protocol are not unique for each listenr when hostnames not set": {
			mutate: func(gw *gatewayv1b1.Gateway) {
				gw.Spec.Listeners[0].Name = "foo"
				gw.Spec.Listeners[0].Protocol = gatewayv1b1.HTTPProtocolType
				gw.Spec.Listeners[0].Port = 80
				gw.Spec.Listeners = append(gw.Spec.Listeners,
					gatewayv1b1.Listener{
						Name:     "bar",
						Protocol: gatewayv1b1.HTTPProtocolType,
						Port:     80,
					},
				)
			},
			expectErrsOnFields: []string{"spec.listeners[1]"},
		},
		"port is unique when protocol and hostname are the same": {
			mutate: func(gw *gatewayv1b1.Gateway) {
				hostnameFoo := gatewayv1b1.Hostname("foo.com")
				gw.Spec.Listeners[0].Name = "foo"
				gw.Spec.Listeners[0].Hostname = &hostnameFoo
				gw.Spec.Listeners[0].Protocol = gatewayv1b1.HTTPProtocolType
				gw.Spec.Listeners[0].Port = 80
				gw.Spec.Listeners = append(gw.Spec.Listeners,
					gatewayv1b1.Listener{
						Name:     "bar",
						Hostname: &hostnameFoo,
						Protocol: gatewayv1b1.HTTPProtocolType,
						Port:     8080,
					},
				)
			},
			expectErrsOnFields: nil,
		},
		"hostname is unique when protocol and port are the same": {
			mutate: func(gw *gatewayv1b1.Gateway) {
				hostnameFoo := gatewayv1b1.Hostname("foo.com")
				hostnameBar := gatewayv1b1.Hostname("bar.com")
				gw.Spec.Listeners[0].Name = "foo"
				gw.Spec.Listeners[0].Hostname = &hostnameFoo
				gw.Spec.Listeners[0].Protocol = gatewayv1b1.HTTPProtocolType
				gw.Spec.Listeners[0].Port = 80
				gw.Spec.Listeners = append(gw.Spec.Listeners,
					gatewayv1b1.Listener{
						Name:     "bar",
						Hostname: &hostnameBar,
						Protocol: gatewayv1b1.HTTPProtocolType,
						Port:     80,
					},
				)
			},
			expectErrsOnFields: nil,
		},
		"protocol is unique when port and hostname are the same": {
			mutate: func(gw *gatewayv1b1.Gateway) {
				hostnameFoo := gatewayv1b1.Hostname("foo.com")
				tlsConfigFoo := tlsConfig
				tlsModeFoo := gatewayv1b1.TLSModeType("Terminate")
				tlsConfigFoo.Mode = &tlsModeFoo
				tlsConfigFoo.CertificateRefs = []gatewayv1b1.SecretObjectReference{
					{
						Name: "FooCertificateRefs",
					},
				}
				gw.Spec.Listeners[0].Name = "foo"
				gw.Spec.Listeners[0].Hostname = &hostnameFoo
				gw.Spec.Listeners[0].Protocol = gatewayv1b1.HTTPSProtocolType
				gw.Spec.Listeners[0].Port = 8000
				gw.Spec.Listeners[0].TLS = &tlsConfigFoo
				gw.Spec.Listeners = append(gw.Spec.Listeners,
					gatewayv1b1.Listener{
						Name:     "bar",
						Hostname: &hostnameFoo,
						Protocol: gatewayv1b1.TLSProtocolType,
						Port:     8000,
						TLS:      &tlsConfigFoo,
					},
				)
			},
			expectErrsOnFields: nil,
		},
		"ip address and hostname in addresses are valid": {
			mutate: func(gw *gatewayv1b1.Gateway) {
				gw.Spec.Addresses = []gatewayv1b1.GatewayAddress{
					{
						Type:  ptrTo(gatewayv1b1.IPAddressType),
						Value: "1.2.3.4",
					},
					{
						Type:  ptrTo(gatewayv1b1.HostnameAddressType),
						Value: "foo.bar",
					},
				}
			},
			expectErrsOnFields: nil,
		},
		"ip address and hostname in addresses are invalid": {
			mutate: func(gw *gatewayv1b1.Gateway) {
				gw.Spec.Addresses = []gatewayv1b1.GatewayAddress{
					{
						Type:  ptrTo(gatewayv1b1.IPAddressType),
						Value: "1.2.3.4:8080",
					},
					{
						Type:  ptrTo(gatewayv1b1.HostnameAddressType),
						Value: "*foo/bar",
					},
					{
						Type:  ptrTo(gatewayv1b1.HostnameAddressType),
						Value: "12:34:56::",
					},
				}
			},
			expectErrsOnFields: []string{"spec.addresses[0]", "spec.addresses[1]", "spec.addresses[2]"},
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
