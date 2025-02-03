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

// Code generated by applyconfiguration-gen. DO NOT EDIT.

package v1alpha1

import (
	v1 "sigs.k8s.io/gateway-api/apis/v1"
)

// ListenerEntryApplyConfiguration represents a declarative configuration of the ListenerEntry type for use
// with apply.
type ListenerEntryApplyConfiguration struct {
	Name          *v1.SectionName      `json:"name,omitempty"`
	Hostname      *v1.Hostname         `json:"hostname,omitempty"`
	Port          *v1.PortNumber       `json:"port,omitempty"`
	Protocol      *v1.ProtocolType     `json:"protocol,omitempty"`
	TLS           *v1.GatewayTLSConfig `json:"tls,omitempty"`
	AllowedRoutes *v1.AllowedRoutes    `json:"allowedRoutes,omitempty"`
}

// ListenerEntryApplyConfiguration constructs a declarative configuration of the ListenerEntry type for use with
// apply.
func ListenerEntry() *ListenerEntryApplyConfiguration {
	return &ListenerEntryApplyConfiguration{}
}

// WithName sets the Name field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the Name field is set to the value of the last call.
func (b *ListenerEntryApplyConfiguration) WithName(value v1.SectionName) *ListenerEntryApplyConfiguration {
	b.Name = &value
	return b
}

// WithHostname sets the Hostname field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the Hostname field is set to the value of the last call.
func (b *ListenerEntryApplyConfiguration) WithHostname(value v1.Hostname) *ListenerEntryApplyConfiguration {
	b.Hostname = &value
	return b
}

// WithPort sets the Port field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the Port field is set to the value of the last call.
func (b *ListenerEntryApplyConfiguration) WithPort(value v1.PortNumber) *ListenerEntryApplyConfiguration {
	b.Port = &value
	return b
}

// WithProtocol sets the Protocol field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the Protocol field is set to the value of the last call.
func (b *ListenerEntryApplyConfiguration) WithProtocol(value v1.ProtocolType) *ListenerEntryApplyConfiguration {
	b.Protocol = &value
	return b
}

// WithTLS sets the TLS field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the TLS field is set to the value of the last call.
func (b *ListenerEntryApplyConfiguration) WithTLS(value v1.GatewayTLSConfig) *ListenerEntryApplyConfiguration {
	b.TLS = &value
	return b
}

// WithAllowedRoutes sets the AllowedRoutes field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the AllowedRoutes field is set to the value of the last call.
func (b *ListenerEntryApplyConfiguration) WithAllowedRoutes(value v1.AllowedRoutes) *ListenerEntryApplyConfiguration {
	b.AllowedRoutes = &value
	return b
}
