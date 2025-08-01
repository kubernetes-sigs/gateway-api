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

package v1alpha3

import (
	apisv1 "sigs.k8s.io/gateway-api/apis/v1"
	v1 "sigs.k8s.io/gateway-api/applyconfiguration/apis/v1"
	v1alpha2 "sigs.k8s.io/gateway-api/applyconfiguration/apis/v1alpha2"
)

// TLSRouteSpecApplyConfiguration represents a declarative configuration of the TLSRouteSpec type for use
// with apply.
type TLSRouteSpecApplyConfiguration struct {
	v1.CommonRouteSpecApplyConfiguration `json:",inline"`
	Hostnames                            []apisv1.Hostname                         `json:"hostnames,omitempty"`
	Rules                                []v1alpha2.TLSRouteRuleApplyConfiguration `json:"rules,omitempty"`
}

// TLSRouteSpecApplyConfiguration constructs a declarative configuration of the TLSRouteSpec type for use with
// apply.
func TLSRouteSpec() *TLSRouteSpecApplyConfiguration {
	return &TLSRouteSpecApplyConfiguration{}
}

// WithParentRefs adds the given value to the ParentRefs field in the declarative configuration
// and returns the receiver, so that objects can be build by chaining "With" function invocations.
// If called multiple times, values provided by each call will be appended to the ParentRefs field.
func (b *TLSRouteSpecApplyConfiguration) WithParentRefs(values ...*v1.ParentReferenceApplyConfiguration) *TLSRouteSpecApplyConfiguration {
	for i := range values {
		if values[i] == nil {
			panic("nil value passed to WithParentRefs")
		}
		b.CommonRouteSpecApplyConfiguration.ParentRefs = append(b.CommonRouteSpecApplyConfiguration.ParentRefs, *values[i])
	}
	return b
}

// WithHostnames adds the given value to the Hostnames field in the declarative configuration
// and returns the receiver, so that objects can be build by chaining "With" function invocations.
// If called multiple times, values provided by each call will be appended to the Hostnames field.
func (b *TLSRouteSpecApplyConfiguration) WithHostnames(values ...apisv1.Hostname) *TLSRouteSpecApplyConfiguration {
	for i := range values {
		b.Hostnames = append(b.Hostnames, values[i])
	}
	return b
}

// WithRules adds the given value to the Rules field in the declarative configuration
// and returns the receiver, so that objects can be build by chaining "With" function invocations.
// If called multiple times, values provided by each call will be appended to the Rules field.
func (b *TLSRouteSpecApplyConfiguration) WithRules(values ...*v1alpha2.TLSRouteRuleApplyConfiguration) *TLSRouteSpecApplyConfiguration {
	for i := range values {
		if values[i] == nil {
			panic("nil value passed to WithRules")
		}
		b.Rules = append(b.Rules, *values[i])
	}
	return b
}
