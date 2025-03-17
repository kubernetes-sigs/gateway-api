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

package v1

import (
	apisv1 "sigs.k8s.io/gateway-api/apis/v1"
)

// GRPCRouteFilterApplyConfiguration represents a declarative configuration of the GRPCRouteFilter type for use
// with apply.
type GRPCRouteFilterApplyConfiguration struct {
	Type                   *apisv1.GRPCRouteFilterType                `json:"type,omitempty"`
	RequestHeaderModifier  *HTTPHeaderFilterApplyConfiguration        `json:"requestHeaderModifier,omitempty"`
	ResponseHeaderModifier *HTTPHeaderFilterApplyConfiguration        `json:"responseHeaderModifier,omitempty"`
	RequestMirror          *HTTPRequestMirrorFilterApplyConfiguration `json:"requestMirror,omitempty"`
	ExtensionRef           *LocalObjectReferenceApplyConfiguration    `json:"extensionRef,omitempty"`
}

// GRPCRouteFilterApplyConfiguration constructs a declarative configuration of the GRPCRouteFilter type for use with
// apply.
func GRPCRouteFilter() *GRPCRouteFilterApplyConfiguration {
	return &GRPCRouteFilterApplyConfiguration{}
}

// WithType sets the Type field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the Type field is set to the value of the last call.
func (b *GRPCRouteFilterApplyConfiguration) WithType(value apisv1.GRPCRouteFilterType) *GRPCRouteFilterApplyConfiguration {
	b.Type = &value
	return b
}

// WithRequestHeaderModifier sets the RequestHeaderModifier field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the RequestHeaderModifier field is set to the value of the last call.
func (b *GRPCRouteFilterApplyConfiguration) WithRequestHeaderModifier(value *HTTPHeaderFilterApplyConfiguration) *GRPCRouteFilterApplyConfiguration {
	b.RequestHeaderModifier = value
	return b
}

// WithResponseHeaderModifier sets the ResponseHeaderModifier field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the ResponseHeaderModifier field is set to the value of the last call.
func (b *GRPCRouteFilterApplyConfiguration) WithResponseHeaderModifier(value *HTTPHeaderFilterApplyConfiguration) *GRPCRouteFilterApplyConfiguration {
	b.ResponseHeaderModifier = value
	return b
}

// WithRequestMirror sets the RequestMirror field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the RequestMirror field is set to the value of the last call.
func (b *GRPCRouteFilterApplyConfiguration) WithRequestMirror(value *HTTPRequestMirrorFilterApplyConfiguration) *GRPCRouteFilterApplyConfiguration {
	b.RequestMirror = value
	return b
}

// WithExtensionRef sets the ExtensionRef field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the ExtensionRef field is set to the value of the last call.
func (b *GRPCRouteFilterApplyConfiguration) WithExtensionRef(value *LocalObjectReferenceApplyConfiguration) *GRPCRouteFilterApplyConfiguration {
	b.ExtensionRef = value
	return b
}
