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
	v1 "sigs.k8s.io/gateway-api/applyconfiguration/apis/v1"
)

// TLSRouteStatusApplyConfiguration represents a declarative configuration of the TLSRouteStatus type for use
// with apply.
type TLSRouteStatusApplyConfiguration struct {
	v1.RouteStatusApplyConfiguration `json:",inline"`
}

// TLSRouteStatusApplyConfiguration constructs a declarative configuration of the TLSRouteStatus type for use with
// apply.
func TLSRouteStatus() *TLSRouteStatusApplyConfiguration {
	return &TLSRouteStatusApplyConfiguration{}
}

// WithParents adds the given value to the Parents field in the declarative configuration
// and returns the receiver, so that objects can be build by chaining "With" function invocations.
// If called multiple times, values provided by each call will be appended to the Parents field.
func (b *TLSRouteStatusApplyConfiguration) WithParents(values ...*v1.RouteParentStatusApplyConfiguration) *TLSRouteStatusApplyConfiguration {
	for i := range values {
		if values[i] == nil {
			panic("nil value passed to WithParents")
		}
		b.RouteStatusApplyConfiguration.Parents = append(b.RouteStatusApplyConfiguration.Parents, *values[i])
	}
	return b
}
