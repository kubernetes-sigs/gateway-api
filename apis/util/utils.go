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

package utils

import (
	gatewayv1a2 "sigs.k8s.io/gateway-api/apis/v1alpha2"
)

// PathMatchTypePtr generate path match type
func PathMatchTypePtr(s string) *gatewayv1a2.PathMatchType {
	if s != string(gatewayv1a2.PathMatchExact) && s != string(gatewayv1a2.PathMatchPrefix) && s != string(gatewayv1a2.PathMatchRegularExpression) &&
		s != string(gatewayv1a2.PathMatchImplementationSpecific) {
		return nil
	}
	result := gatewayv1a2.PathMatchType(s)
	return &result
}

// PortNumberPtr generate port number
func PortNumberPtr(p int) *gatewayv1a2.PortNumber {
	if p < 1 || p > 65535 {
		return nil
	}
	result := gatewayv1a2.PortNumber(p)
	return &result
}
