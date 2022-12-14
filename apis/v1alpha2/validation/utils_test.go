/*
Copyright 2022 The Kubernetes Authors.

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
	gatewayv1a2 "sigs.k8s.io/gateway-api/apis/v1alpha2"
	"sigs.k8s.io/gateway-api/apis/v1beta1"
)

type routeRule interface {
	gatewayv1a2.TLSRouteRule | gatewayv1a2.UDPRouteRule
}

func makeRouteRules[T routeRule](ports ...*int32) (rules []T) {
	for _, port := range ports {
		rules = append(rules, T{
			BackendRefs: []gatewayv1a2.BackendRef{{
				BackendObjectReference: gatewayv1a2.BackendObjectReference{
					Port: (*v1beta1.PortNumber)(port),
				},
			}},
		})
	}
	return
}
