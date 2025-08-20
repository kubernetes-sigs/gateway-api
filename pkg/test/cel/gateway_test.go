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
	"testing"
	"time"

	gatewayv1 "sigs.k8s.io/gateway-api/apis/v1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestValidateGateway(t *testing.T) {
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
		desc         string
		mutate       func(gw *gatewayv1.Gateway)
		mutateStatus func(gw *gatewayv1.Gateway)
		wantErrors   []string
	}{
		{
			desc: "tls config present with http protocol",
			mutate: func(gw *gatewayv1.Gateway) {
				gw.Spec.Listeners = []gatewayv1.Listener{
					{
						Name:     gatewayv1.SectionName("http"),
						Protocol: gatewayv1.HTTPProtocolType,
						Port:     gatewayv1.PortNumber(8080),
						TLS:      &gatewayv1.ListenerTLSConfig{},
					},
				}
			},
			wantErrors: []string{"tls must not be specified for protocols ['HTTP', 'TCP', 'UDP']"},
		},
		{
			desc: "https protocol with Passthrough tls mode",
			mutate: func(gw *gatewayv1.Gateway) {
				gw.Spec.Listeners = []gatewayv1.Listener{
					{
						Name:     gatewayv1.SectionName("https"),
						Protocol: gatewayv1.HTTPSProtocolType,
						Port:     gatewayv1.PortNumber(8080),
						TLS: &gatewayv1.ListenerTLSConfig{
							Mode: ptrTo(gatewayv1.TLSModeType("Passthrough")),
						},
					},
				}
			},
			wantErrors: []string{"tls mode must be Terminate for protocol HTTPS"},
		},
		{
			desc: "tls mode not set with https protocol and tls config present",
			mutate: func(gw *gatewayv1.Gateway) {
				gw.Spec.Listeners = []gatewayv1.Listener{
					{
						Name:     gatewayv1.SectionName("https"),
						Protocol: gatewayv1.HTTPSProtocolType,
						Port:     gatewayv1.PortNumber(8080),
						TLS: &gatewayv1.ListenerTLSConfig{
							CertificateRefs: []gatewayv1.SecretObjectReference{
								{Name: gatewayv1.ObjectName("foo")},
							},
						},
					},
				}
			},
		},
		{
			desc: "tls config present with tcp protocol",
			mutate: func(gw *gatewayv1.Gateway) {
				gw.Spec.Listeners = []gatewayv1.Listener{
					{
						Name:     gatewayv1.SectionName("tcp"),
						Protocol: gatewayv1.TCPProtocolType,
						Port:     gatewayv1.PortNumber(8080),
						TLS:      &gatewayv1.ListenerTLSConfig{},
					},
				}
			},
			wantErrors: []string{"tls must not be specified for protocols ['HTTP', 'TCP', 'UDP']"},
		},
		{
			desc: "tls config not set with https protocol",
			mutate: func(gw *gatewayv1.Gateway) {
				gw.Spec.Listeners = []gatewayv1.Listener{
					{
						Name:     gatewayv1.SectionName("https"),
						Protocol: gatewayv1.HTTPSProtocolType,
						Port:     gatewayv1.PortNumber(8443),
					},
				}
			},
		},
		{
			desc: "tls config not set with tls protocol",
			mutate: func(gw *gatewayv1.Gateway) {
				gw.Spec.Listeners = []gatewayv1.Listener{
					{
						Name:     gatewayv1.SectionName("tls"),
						Protocol: gatewayv1.TLSProtocolType,
						Port:     gatewayv1.PortNumber(8443),
					},
				}
			},
		},
		{
			desc: "tls config not set with http protocol",
			mutate: func(gw *gatewayv1.Gateway) {
				gw.Spec.Listeners = []gatewayv1.Listener{
					{
						Name:     gatewayv1.SectionName("http"),
						Protocol: gatewayv1.HTTPProtocolType,
						Port:     gatewayv1.PortNumber(8080),
					},
				}
			},
		},
		{
			desc: "tls config not set with tcp protocol",
			mutate: func(gw *gatewayv1.Gateway) {
				gw.Spec.Listeners = []gatewayv1.Listener{
					{
						Name:     gatewayv1.SectionName("tcp"),
						Protocol: gatewayv1.TCPProtocolType,
						Port:     gatewayv1.PortNumber(8080),
					},
				}
			},
		},
		{
			desc: "tls config not set with udp protocol",
			mutate: func(gw *gatewayv1.Gateway) {
				gw.Spec.Listeners = []gatewayv1.Listener{
					{
						Name:     gatewayv1.SectionName("udp"),
						Protocol: gatewayv1.UDPProtocolType,
						Port:     gatewayv1.PortNumber(8080),
					},
				}
			},
		},
		{
			desc: "hostname present with tcp protocol",
			mutate: func(gw *gatewayv1.Gateway) {
				hostname := gatewayv1.Hostname("foo")
				gw.Spec.Listeners = []gatewayv1.Listener{
					{
						Name:     gatewayv1.SectionName("tcp"),
						Protocol: gatewayv1.TCPProtocolType,
						Port:     gatewayv1.PortNumber(8080),
						Hostname: &hostname,
					},
				}
			},
			wantErrors: []string{"hostname must not be specified for protocols ['TCP', 'UDP']"},
		},
		{
			desc: "hostname present with udp protocol",
			mutate: func(gw *gatewayv1.Gateway) {
				hostname := gatewayv1.Hostname("foo")
				gw.Spec.Listeners = []gatewayv1.Listener{
					{
						Name:     gatewayv1.SectionName("udp"),
						Protocol: gatewayv1.UDPProtocolType,
						Port:     gatewayv1.PortNumber(8080),
						Hostname: &hostname,
					},
				}
			},
			wantErrors: []string{"hostname must not be specified for protocols ['TCP', 'UDP']"},
		},
		{
			desc: "certificateRefs not set with HTTPS protocol and TLS terminate mode",
			mutate: func(gw *gatewayv1.Gateway) {
				tlsMode := gatewayv1.TLSModeType("Terminate")
				gw.Spec.Listeners = []gatewayv1.Listener{
					{
						Name:     gatewayv1.SectionName("https"),
						Protocol: gatewayv1.HTTPSProtocolType,
						Port:     gatewayv1.PortNumber(8443),
						TLS: &gatewayv1.ListenerTLSConfig{
							Mode: &tlsMode,
						},
					},
				}
			},
			wantErrors: []string{"certificateRefs or options must be specified when mode is Terminate"},
		},
		{
			desc: "certificateRefs not set with TLS protocol and TLS terminate mode",
			mutate: func(gw *gatewayv1.Gateway) {
				tlsMode := gatewayv1.TLSModeType("Terminate")
				gw.Spec.Listeners = []gatewayv1.Listener{
					{
						Name:     gatewayv1.SectionName("tls"),
						Protocol: gatewayv1.TLSProtocolType,
						Port:     gatewayv1.PortNumber(8443),
						TLS: &gatewayv1.ListenerTLSConfig{
							Mode: &tlsMode,
						},
					},
				}
			},
			wantErrors: []string{"certificateRefs or options must be specified when mode is Terminate"},
		},
		{
			desc: "certificateRefs set with HTTPS protocol and TLS terminate mode",
			mutate: func(gw *gatewayv1.Gateway) {
				tlsMode := gatewayv1.TLSModeType("Terminate")
				gw.Spec.Listeners = []gatewayv1.Listener{
					{
						Name:     gatewayv1.SectionName("https"),
						Protocol: gatewayv1.HTTPSProtocolType,
						Port:     gatewayv1.PortNumber(8443),
						TLS: &gatewayv1.ListenerTLSConfig{
							Mode: &tlsMode,
							CertificateRefs: []gatewayv1.SecretObjectReference{
								{Name: gatewayv1.ObjectName("foo")},
							},
						},
					},
				}
			},
		},
		{
			desc: "certificateRefs set with TLS protocol and TLS terminate mode",
			mutate: func(gw *gatewayv1.Gateway) {
				tlsMode := gatewayv1.TLSModeType("Terminate")
				gw.Spec.Listeners = []gatewayv1.Listener{
					{
						Name:     gatewayv1.SectionName("tls"),
						Protocol: gatewayv1.TLSProtocolType,
						Port:     gatewayv1.PortNumber(8443),
						TLS: &gatewayv1.ListenerTLSConfig{
							Mode: &tlsMode,
							CertificateRefs: []gatewayv1.SecretObjectReference{
								{Name: gatewayv1.ObjectName("foo")},
							},
						},
					},
				}
			},
		},
		{
			desc: "options set with HTTPS protocol and TLS terminate mode",
			mutate: func(gw *gatewayv1.Gateway) {
				tlsMode := gatewayv1.TLSModeType("Terminate")
				gw.Spec.Listeners = []gatewayv1.Listener{
					{
						Name:     gatewayv1.SectionName("https"),
						Protocol: gatewayv1.HTTPSProtocolType,
						Port:     gatewayv1.PortNumber(8443),
						TLS: &gatewayv1.ListenerTLSConfig{
							Mode: &tlsMode,
							Options: map[gatewayv1.AnnotationKey]gatewayv1.AnnotationValue{
								"networking.example.com/tls-version": "1.2",
							},
						},
					},
				}
			},
		},
		{
			desc: "options set with tls protocol and TLS terminate mode",
			mutate: func(gw *gatewayv1.Gateway) {
				tlsMode := gatewayv1.TLSModeType("Terminate")
				gw.Spec.Listeners = []gatewayv1.Listener{
					{
						Name:     gatewayv1.SectionName("tls"),
						Protocol: gatewayv1.TLSProtocolType,
						Port:     gatewayv1.PortNumber(8443),
						TLS: &gatewayv1.ListenerTLSConfig{
							Mode: &tlsMode,
							Options: map[gatewayv1.AnnotationKey]gatewayv1.AnnotationValue{
								"networking.example.com/tls-version": "1.2",
							},
						},
					},
				}
			},
		},
		{
			desc: "names are not unique within the Gateway",
			mutate: func(gw *gatewayv1.Gateway) {
				gw.Spec.Listeners = []gatewayv1.Listener{
					{
						Name:     gatewayv1.SectionName("http"),
						Protocol: gatewayv1.HTTPProtocolType,
						Port:     gatewayv1.PortNumber(80),
					},
					{
						Name:     gatewayv1.SectionName("http"),
						Protocol: gatewayv1.HTTPProtocolType,
						Port:     gatewayv1.PortNumber(8000),
					},
					{
						Name:     gatewayv1.SectionName("http"),
						Protocol: gatewayv1.HTTPProtocolType,
						Port:     gatewayv1.PortNumber(8080),
					},
				}
			},
			wantErrors: []string{"Listener name must be unique within the Gateway"},
		},
		{
			desc: "names are unique within the Gateway",
			mutate: func(gw *gatewayv1.Gateway) {
				gw.Spec.Listeners = []gatewayv1.Listener{
					{
						Name:     gatewayv1.SectionName("http-1"),
						Protocol: gatewayv1.HTTPProtocolType,
						Port:     gatewayv1.PortNumber(80),
					},
					{
						Name:     gatewayv1.SectionName("http-2"),
						Protocol: gatewayv1.HTTPProtocolType,
						Port:     gatewayv1.PortNumber(8000),
					},
					{
						Name:     gatewayv1.SectionName("http-3"),
						Protocol: gatewayv1.HTTPProtocolType,
						Port:     gatewayv1.PortNumber(8080),
					},
				}
			},
		},
		{
			desc: "combination of port, protocol, and hostname are not unique for each listener",
			mutate: func(gw *gatewayv1.Gateway) {
				hostnameFoo := gatewayv1.Hostname("foo.com")
				gw.Spec.Listeners = []gatewayv1.Listener{
					{
						Name:     gatewayv1.SectionName("foo"),
						Protocol: gatewayv1.HTTPProtocolType,
						Port:     gatewayv1.PortNumber(80),
						Hostname: &hostnameFoo,
					},
					{
						Name:     gatewayv1.SectionName("bar"),
						Protocol: gatewayv1.HTTPProtocolType,
						Port:     gatewayv1.PortNumber(80),
						Hostname: &hostnameFoo,
					},
				}
			},
			wantErrors: []string{"Combination of port, protocol and hostname must be unique for each listener"},
		},
		{
			desc: "combination of port and protocol are not unique for each listener when hostnames not set",
			mutate: func(gw *gatewayv1.Gateway) {
				gw.Spec.Listeners = []gatewayv1.Listener{
					{
						Name:     gatewayv1.SectionName("foo"),
						Protocol: gatewayv1.HTTPProtocolType,
						Port:     gatewayv1.PortNumber(80),
					},
					{
						Name:     gatewayv1.SectionName("bar"),
						Protocol: gatewayv1.HTTPProtocolType,
						Port:     gatewayv1.PortNumber(80),
					},
				}
			},
			wantErrors: []string{"Combination of port, protocol and hostname must be unique for each listener"},
		},
		{
			desc: "port is unique when protocol and hostname are the same",
			mutate: func(gw *gatewayv1.Gateway) {
				hostnameFoo := gatewayv1.Hostname("foo.com")
				gw.Spec.Listeners = []gatewayv1.Listener{
					{
						Name:     gatewayv1.SectionName("foo"),
						Protocol: gatewayv1.HTTPProtocolType,
						Port:     gatewayv1.PortNumber(80),
						Hostname: &hostnameFoo,
					},
					{
						Name:     gatewayv1.SectionName("bar"),
						Protocol: gatewayv1.HTTPProtocolType,
						Port:     gatewayv1.PortNumber(8000),
						Hostname: &hostnameFoo,
					},
				}
			},
		},
		{
			desc: "hostname is unique when protocol and port are the same",
			mutate: func(gw *gatewayv1.Gateway) {
				hostnameFoo := gatewayv1.Hostname("foo.com")
				hostnameBar := gatewayv1.Hostname("bar.com")
				gw.Spec.Listeners = []gatewayv1.Listener{
					{
						Name:     gatewayv1.SectionName("foo"),
						Protocol: gatewayv1.HTTPProtocolType,
						Port:     gatewayv1.PortNumber(80),
						Hostname: &hostnameFoo,
					},
					{
						Name:     gatewayv1.SectionName("bar"),
						Protocol: gatewayv1.HTTPProtocolType,
						Port:     gatewayv1.PortNumber(80),
						Hostname: &hostnameBar,
					},
				}
			},
		},
		{
			desc: "one omitted hostname is unique when protocol and port are the same",
			mutate: func(gw *gatewayv1.Gateway) {
				hostnameFoo := gatewayv1.Hostname("foo.com")
				gw.Spec.Listeners = []gatewayv1.Listener{
					{
						Name:     gatewayv1.SectionName("foo"),
						Protocol: gatewayv1.HTTPProtocolType,
						Port:     gatewayv1.PortNumber(80),
						Hostname: &hostnameFoo,
					},
					{
						Name:     gatewayv1.SectionName("bar"),
						Protocol: gatewayv1.HTTPProtocolType,
						Port:     gatewayv1.PortNumber(80),
					},
				}
			},
		},
		{
			desc: "protocol is unique when port and hostname are the same",
			mutate: func(gw *gatewayv1.Gateway) {
				hostnameFoo := gatewayv1.Hostname("foo.com")
				gw.Spec.Listeners = []gatewayv1.Listener{
					{
						Name:     gatewayv1.SectionName("foo"),
						Protocol: gatewayv1.HTTPProtocolType,
						Port:     gatewayv1.PortNumber(8000),
						Hostname: &hostnameFoo,
					},
					{
						Name:     gatewayv1.SectionName("bar"),
						Protocol: gatewayv1.HTTPSProtocolType,
						Port:     gatewayv1.PortNumber(8000),
						Hostname: &hostnameFoo,
						TLS: &gatewayv1.ListenerTLSConfig{
							CertificateRefs: []gatewayv1.SecretObjectReference{
								{Name: gatewayv1.ObjectName("foo")},
							},
						},
					},
				}
			},
		},
		{
			desc: "ip address and hostname in addresses are valid",
			mutate: func(gw *gatewayv1.Gateway) {
				gw.Spec.Addresses = []gatewayv1.GatewaySpecAddress{
					{
						Type:  ptrTo(gatewayv1.IPAddressType),
						Value: "1.2.3.4",
					},
					{
						Type:  ptrTo(gatewayv1.IPAddressType),
						Value: "1111:2222:3333:4444::",
					},
					{
						Type:  ptrTo(gatewayv1.HostnameAddressType),
						Value: "foo.bar",
					},
				}
			},
		},
		{
			desc: "ip address and hostname in addresses are invalid",
			mutate: func(gw *gatewayv1.Gateway) {
				gw.Spec.Addresses = []gatewayv1.GatewaySpecAddress{
					{
						Type:  ptrTo(gatewayv1.IPAddressType),
						Value: "1.2.3.4:8080",
					},
					{
						Type:  ptrTo(gatewayv1.HostnameAddressType),
						Value: "*foo/bar",
					},
					{
						Type:  ptrTo(gatewayv1.HostnameAddressType),
						Value: "12:34:56::",
					},
				}
			},
			wantErrors: []string{"Invalid value: \"1.2.3.4:8080\": spec.addresses[0].value in body must be of type ipv4"},
		},
		{
			desc: "ip address and hostname in status addresses are valid",
			mutateStatus: func(gw *gatewayv1.Gateway) {
				gw.Status.Addresses = []gatewayv1.GatewayStatusAddress{
					{
						Type:  ptrTo(gatewayv1.IPAddressType),
						Value: "1.2.3.4",
					},
					{
						Type:  ptrTo(gatewayv1.IPAddressType),
						Value: "1111:2222:3333:4444::",
					},
					{
						Type:  ptrTo(gatewayv1.HostnameAddressType),
						Value: "foo.bar",
					},
				}
			},
		},
		{
			desc: "ip address and hostname in status addresses are invalid",
			mutateStatus: func(gw *gatewayv1.Gateway) {
				gw.Status.Addresses = []gatewayv1.GatewayStatusAddress{
					{
						Type:  ptrTo(gatewayv1.IPAddressType),
						Value: "1.2.3.4:8080",
					},
					{
						Type:  ptrTo(gatewayv1.HostnameAddressType),
						Value: "*foo/bar",
					},
					{
						Type:  ptrTo(gatewayv1.HostnameAddressType),
						Value: "12:34:56::",
					},
				}
			},
			wantErrors: []string{"Invalid value: \"1.2.3.4:8080\": status.addresses[0].value in body must be of type ipv4"},
		},
		{
			desc: "duplicate ip address or hostname",
			mutate: func(gw *gatewayv1.Gateway) {
				gw.Spec.Addresses = []gatewayv1.GatewaySpecAddress{
					{
						Type:  ptrTo(gatewayv1.IPAddressType),
						Value: "1.2.3.4",
					},
					{
						Type:  ptrTo(gatewayv1.IPAddressType),
						Value: "1.2.3.4",
					},
					{
						Type:  ptrTo(gatewayv1.HostnameAddressType),
						Value: "foo.bar",
					},
					{
						Type:  ptrTo(gatewayv1.HostnameAddressType),
						Value: "foo.bar",
					},
				}
			},
			wantErrors: []string{"IPAddress values must be unique", "Hostname values must be unique"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			gw := baseGateway.DeepCopy()
			gw.Name = fmt.Sprintf("foo-%v", time.Now().UnixNano())

			if tc.mutate != nil {
				tc.mutate(gw)
			}
			err := k8sClient.Create(ctx, gw)

			if tc.mutateStatus != nil {
				tc.mutateStatus(gw)
				err = k8sClient.Status().Update(ctx, gw)
			}

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
