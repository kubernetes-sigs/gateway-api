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

func TestGatewayClusterTrustBundleReferenceNamespace(t *testing.T) {
	gateway := &gatewayv1.Gateway{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("foo-%v", time.Now().UnixNano()),
			Namespace: metav1.NamespaceDefault,
		},
		Spec: gatewayv1.GatewaySpec{
			GatewayClassName: "foo",
			Listeners: []gatewayv1.Listener{
				{
					Name:     gatewayv1.SectionName("https"),
					Protocol: gatewayv1.HTTPSProtocolType,
					Port:     gatewayv1.PortNumber(443),
					TLS: &gatewayv1.ListenerTLSConfig{
						Mode: new(gatewayv1.TLSModeType("Terminate")),
					},
				},
			},
			TLS: &gatewayv1.GatewayTLSConfig{
				Frontend: &gatewayv1.FrontendTLSConfig{
					Default: gatewayv1.TLSConfig{
						Validation: &gatewayv1.FrontendTLSValidation{
							CACertificateRefs: []gatewayv1.ObjectReference{
								{
									Group:     "certificates.k8s.io",
									Kind:      "ClusterTrustBundle",
									Name:      "example.com:internal-signer:v1",
									Namespace: new(gatewayv1.Namespace("unexpected")),
								},
							},
						},
					},
				},
			},
		},
	}

	err := k8sClient.Create(context.Background(), gateway)
	if err == nil {
		t.Fatalf("expected Gateway with a namespaced ClusterTrustBundle reference to be rejected")
	}
	wantError := "ClusterTrustBundle references must not specify namespace"
	if !celErrorStringMatches(err.Error(), wantError) {
		t.Errorf("Unexpected response while creating Gateway; got err=\n%v\n;missing string within error=%q", err, wantError)
	}
}

func TestGatewayClusterTrustBundleReferenceEmptyNamespace(t *testing.T) {
	gateway := &gatewayv1.Gateway{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("foo-%v", time.Now().UnixNano()),
			Namespace: metav1.NamespaceDefault,
		},
		Spec: gatewayv1.GatewaySpec{
			GatewayClassName: "foo",
			Listeners: []gatewayv1.Listener{
				{
					Name:     gatewayv1.SectionName("https"),
					Protocol: gatewayv1.HTTPSProtocolType,
					Port:     gatewayv1.PortNumber(443),
					TLS: &gatewayv1.ListenerTLSConfig{
						Mode: new(gatewayv1.TLSModeType("Terminate")),
					},
				},
			},
			TLS: &gatewayv1.GatewayTLSConfig{
				Frontend: &gatewayv1.FrontendTLSConfig{
					Default: gatewayv1.TLSConfig{
						Validation: &gatewayv1.FrontendTLSValidation{
							CACertificateRefs: []gatewayv1.ObjectReference{
								{
									Group:     "certificates.k8s.io",
									Kind:      "ClusterTrustBundle",
									Name:      "example.com:internal-signer:v1",
									Namespace: new(gatewayv1.Namespace("")),
								},
							},
						},
					},
				},
			},
		},
	}

	err := k8sClient.Create(context.Background(), gateway)
	if err == nil {
		t.Fatal("expected Gateway with an empty ClusterTrustBundle reference namespace to be rejected")
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
		{name: "prefix containing k8s.io spec", routability: "k8s.io.example.com/scope"},
		{name: "middle prefix containing k8s.io spec", routability: "example.k8s.io.com/scope"},
		{name: "path containing k8s.io spec", routability: "example.com/scope/k8s.io"},
		{name: "sentinel spec", routability: "testing.x-k8s.io/sentinel"},
		{name: "other reserved testing prefix spec", routability: "testing.x-k8s.io/unsupported"},
		{name: "unknown spec", routability: "Unknown", wantError: true},
		{name: "empty prefix spec", routability: "/scope", wantError: true},
		{name: "empty path spec", routability: "example.com/", wantError: true},
		{name: "invalid prefix spec", routability: "example..com/scope", wantError: true},
		{name: "reserved k8s.io spec", routability: "k8s.io/scope", wantError: true},
		{name: "reserved subdomain spec", routability: "example.k8s.io/scope", wantError: true},
		{name: "reserved nested subdomain spec", routability: "foo.bar.k8s.io/scope", wantError: true},
		{name: "reserved gateway subdomain spec", routability: "gateway.networking.k8s.io/sentinel", wantError: true},
		{name: "empty status", routability: "", statusAddress: true},
		{name: "cluster status", routability: "Cluster", statusAddress: true},
		{name: "prefixed status", routability: "example.com/scope", statusAddress: true},
		{name: "prefix containing k8s.io status", routability: "k8s.io.example.com/scope", statusAddress: true},
		{name: "middle prefix containing k8s.io status", routability: "example.k8s.io.com/scope", statusAddress: true},
		{name: "path containing k8s.io status", routability: "example.com/scope/k8s.io", statusAddress: true},
		{name: "sentinel status", routability: "testing.x-k8s.io/sentinel", statusAddress: true},
		{name: "other reserved testing prefix status", routability: "testing.x-k8s.io/unsupported", statusAddress: true},
		{name: "unknown status", routability: "Unknown", statusAddress: true, wantError: true},
		{name: "empty prefix status", routability: "/scope", statusAddress: true, wantError: true},
		{name: "empty path status", routability: "example.com/", statusAddress: true, wantError: true},
		{name: "invalid prefix status", routability: "example..com/scope", statusAddress: true, wantError: true},
		{name: "reserved k8s.io status", routability: "k8s.io/scope", statusAddress: true, wantError: true},
		{name: "reserved subdomain status", routability: "example.k8s.io/scope", statusAddress: true, wantError: true},
		{name: "reserved nested subdomain status", routability: "foo.bar.k8s.io/scope", statusAddress: true, wantError: true},
		{name: "reserved gateway subdomain status", routability: "gateway.networking.k8s.io/sentinel", statusAddress: true, wantError: true},
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

func TestListenerFiltersProtocolRestriction(t *testing.T) {
	tests := []struct {
		name        string
		protocol    string
		withFilters bool
		wantError   bool
	}{
		{name: "HTTP without filters", protocol: "HTTP"},
		{name: "HTTPS without filters", protocol: "HTTPS"},
		{name: "TCP without filters", protocol: "TCP"},
		{name: "TLS without filters", protocol: "TLS"},
		{name: "UDP without filters", protocol: "UDP"},

		{name: "HTTP with filters", protocol: "HTTP", withFilters: true},
		{name: "HTTPS with filters", protocol: "HTTPS", withFilters: true},

		{name: "TCP with filters", protocol: "TCP", withFilters: true, wantError: true},
		{name: "TLS with filters", protocol: "TLS", withFilters: true, wantError: true},
		{name: "UDP with filters", protocol: "UDP", withFilters: true, wantError: true},
		{name: "custom protocol with filters", protocol: "example.com/foo", withFilters: true, wantError: true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			name := fmt.Sprintf("listener-filters-%d", time.Now().UnixNano())
			listener := map[string]any{
				"name":     "listener",
				"protocol": tc.protocol,
				"port":     int64(80),
			}
			// TLS listeners require a TLS block; HTTPS + Terminate is the
			// minimum that satisfies the existing per-protocol CEL rules,
			// which run before the new one and would otherwise mask it.
			switch tc.protocol {
			case "HTTPS":
				listener["tls"] = map[string]any{
					"mode": "Terminate",
					"certificateRefs": []any{map[string]any{
						"kind": "Secret",
						"name": "example-cert",
					}},
				}
			case "TLS":
				listener["tls"] = map[string]any{
					"mode": "Passthrough",
				}
			}
			if tc.withFilters {
				listener["filters"] = map[string]any{
					"requests": []any{map[string]any{
						"type": "ExtensionRef",
						"extensionRef": map[string]any{
							"group": "example.com",
							"kind":  "PreRoutingExtension",
							"name":  "example",
						},
					}},
				}
			}

			obj := &unstructured.Unstructured{Object: map[string]any{
				"apiVersion": "gateway.networking.k8s.io/v1",
				"kind":       "Gateway",
				"metadata": map[string]any{
					"name":      name,
					"namespace": metav1.NamespaceDefault,
				},
				"spec": map[string]any{
					"gatewayClassName": "foo",
					"listeners":        []any{listener},
				},
			}}

			ctx := context.Background()
			err := k8sClient.Create(ctx, obj)
			if (err != nil) != tc.wantError {
				t.Fatalf("unexpected validation result: got err=%v, want error=%v", err, tc.wantError)
			}
			if tc.wantError && err != nil {
				if !strings.Contains(err.Error(), "filters may only be set when protocol is HTTP or HTTPS") {
					t.Fatalf("expected rejection to cite the filters CEL rule; got: %v", err)
				}
			}
		})
	}
}

func TestListenerFilterUnionDiscriminator(t *testing.T) {
	extRef := map[string]any{
		"group": "example.com",
		"kind":  "PreRoutingExtension",
		"name":  "example",
	}
	extAuth := map[string]any{
		"protocol": "HTTP",
		"backendRef": map[string]any{
			"group": "",
			"kind":  "Service",
			"name":  "auth",
			"port":  int64(9000),
		},
		"http": map[string]any{},
	}

	tests := []struct {
		name         string
		filterType   string
		externalAuth map[string]any
		extensionRef map[string]any
		wantError    bool
		errContains  string
	}{
		// Happy paths.
		{
			name:         "valid ExternalAuth filter",
			filterType:   "ExternalAuth",
			externalAuth: extAuth,
		},
		{
			name:         "valid ExtensionRef filter",
			filterType:   "ExtensionRef",
			extensionRef: extRef,
		},

		// Missing value fields.
		{
			name:        "ExternalAuth filter with empty value field",
			filterType:  "ExternalAuth",
			wantError:   true,
			errContains: "externalAuth must be specified for ExternalAuth filter.type",
		},
		{
			name:        "ExtensionRef filter with empty value field",
			filterType:  "ExtensionRef",
			wantError:   true,
			errContains: "extensionRef must be specified for ExtensionRef filter.type",
		},

		// Value/type mismatch.
		{
			name:         "ExternalAuth filter with non-matching field",
			filterType:   "ExternalAuth",
			extensionRef: extRef,
			wantError:    true,
			errContains:  "extensionRef must be nil if the filter.type is not ExtensionRef",
		},
		{
			name:         "ExtensionRef filter with non-matching field",
			filterType:   "ExtensionRef",
			externalAuth: extAuth,
			wantError:    true,
			errContains:  "externalAuth must be nil if the filter.type is not ExternalAuth",
		},

		// Both value fields set.
		{
			name:         "ExternalAuth filter with both fields set",
			filterType:   "ExternalAuth",
			externalAuth: extAuth,
			extensionRef: extRef,
			wantError:    true,
			errContains:  "extensionRef must be nil if the filter.type is not ExtensionRef",
		},
		{
			name:         "ExtensionRef filter with both fields set",
			filterType:   "ExtensionRef",
			externalAuth: extAuth,
			extensionRef: extRef,
			wantError:    true,
			errContains:  "externalAuth must be nil if the filter.type is not ExternalAuth",
		},

		// Enum guard: a type value outside the allowed set is rejected by the
		// kubebuilder Enum marker, independent of the union CEL rules above.
		{
			name:        "invalid type value",
			filterType:  "RequestHeaderModifier",
			wantError:   true,
			errContains: `supported values: "ExternalAuth", "ExtensionRef"`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			name := fmt.Sprintf("listener-filter-union-%d", time.Now().UnixNano())
			filter := map[string]any{
				"type": tc.filterType,
			}
			if tc.externalAuth != nil {
				filter["externalAuth"] = tc.externalAuth
			}
			if tc.extensionRef != nil {
				filter["extensionRef"] = tc.extensionRef
			}

			obj := &unstructured.Unstructured{Object: map[string]any{
				"apiVersion": "gateway.networking.k8s.io/v1",
				"kind":       "Gateway",
				"metadata": map[string]any{
					"name":      name,
					"namespace": metav1.NamespaceDefault,
				},
				"spec": map[string]any{
					"gatewayClassName": "foo",
					"listeners": []any{map[string]any{
						"name":     "listener",
						"protocol": "HTTP",
						"port":     int64(80),
						"filters": map[string]any{
							"requests": []any{filter},
						},
					}},
				},
			}}

			ctx := context.Background()
			err := k8sClient.Create(ctx, obj)
			if (err != nil) != tc.wantError {
				t.Fatalf("unexpected validation result: got err=%v, want error=%v", err, tc.wantError)
			}
			if tc.wantError && err != nil && !strings.Contains(err.Error(), tc.errContains) {
				t.Fatalf("expected error to contain %q, got: %v", tc.errContains, err)
			}
		})
	}
}
