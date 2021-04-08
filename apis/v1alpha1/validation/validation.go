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
	"net"
	"strings"

	gatewayv1a1 "sigs.k8s.io/gateway-api/apis/v1alpha1"

	"k8s.io/apimachinery/pkg/util/validation"
	"k8s.io/apimachinery/pkg/util/validation/field"
)

// ValidateGateway validates gw according to the Gateway API specification.
// For additional details of the Gateway spec, refer to:
//  https://gateway-api.sigs.k8s.io/spec/#networking.x-k8s.io/v1alpha1.Gateway
func ValidateGateway(gw *gatewayv1a1.Gateway) field.ErrorList {
	return validateGatewaySpec(&gw.Spec, field.NewPath("spec"))
}

// validateGatewaySpec validates whether required fields of spec are set according to the
// Gateway API specification.
func validateGatewaySpec(spec *gatewayv1a1.GatewaySpec, path *field.Path) field.ErrorList {
	// TODO [danehans]: Add additional validation of spec fields.
	return validateGatewayListeners(spec.Listeners, path.Child("listeners"))
}

// validateGatewayListeners validates whether required fields of listeners are set according
// to the Gateway API specification.
func validateGatewayListeners(listeners []gatewayv1a1.Listener, path *field.Path) field.ErrorList {
	// TODO [danehans]: Add additional validation of listener fields.
	return validateListenerHostname(listeners, path)
}

// validateListenerHostname validates each listener hostname is not an IP address and is one
// of the following:
//  - A fully qualified domain name of a network host, as defined by RFC 3986.
//  - A DNS subdomain as defined by RFC 1123.
//  - A wildcard DNS subdomain as defined by RFC 1034 (section 4.3.3).
func validateListenerHostname(listeners []gatewayv1a1.Listener, path *field.Path) field.ErrorList {
	var errs field.ErrorList
	for i, h := range listeners {
		// When unspecified, “”, or *, all hostnames are matched.
		if h.Hostname == nil || (*h.Hostname == "" || *h.Hostname == "*") {
			continue
		}
		hostname := string(*h.Hostname)
		if ip := net.ParseIP(hostname); ip != nil {
			errs = append(errs, field.Invalid(path.Index(i).Child("hostname"), hostname, "must be a DNS hostname, not an IP address"))
		}
		if strings.Contains(hostname, "*") {
			for _, msg := range validation.IsWildcardDNS1123Subdomain(hostname) {
				errs = append(errs, field.Invalid(path.Index(i).Child("hostname"), hostname, msg))
			}
		} else {
			for _, msg := range validation.IsDNS1123Subdomain(hostname) {
				errs = append(errs, field.Invalid(path.Index(i).Child("hostname"), hostname, msg))
			}
		}
	}
	return errs
}
