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
	"fmt"
	"net"
	"strings"

	"k8s.io/apimachinery/pkg/util/validation"
	"k8s.io/apimachinery/pkg/util/validation/field"

	gatewayv1a2 "sigs.k8s.io/gateway-api/apis/v1alpha2"
)

// ValidateGateway validates gw according to the Gateway API specification.
// For additional details of the Gateway spec, refer to:
//  https://gateway-api.sigs.k8s.io/spec/#gateway.networking.k8s.io/v1alpha2.Gateway
func ValidateGateway(gw *gatewayv1a2.Gateway) field.ErrorList {
	return validateGatewaySpec(&gw.Spec, field.NewPath("spec"))
}

// validateGatewaySpec validates whether required fields of spec are set according to the
// Gateway API specification.
func validateGatewaySpec(spec *gatewayv1a2.GatewaySpec, path *field.Path) field.ErrorList {
	if errList := validateGatewayListeners(spec.Listeners, path.Child("listeners")); len(errList) > 0 {
		fmt.Printf("Failed validating gateway listeners.")
		return errList
	}

	if errList := validateGatewayClassName(spec.GatewayClassName, path.Child("gatewayClassName")); len(errList) > 0 {
		return errList

	}

	if errList := validateGatewayAddresses(spec.Addresses, path.Child("addresses")); len(errList) > 0 {
		return errList
	}

	return nil
}

func validateGatewayClassName(gwclsName string, path *field.Path) field.ErrorList {
	var errs field.ErrorList
	if len(gwclsName) == 0 || len(gwclsName) > 253 {
		errs = append(errs, field.Invalid(path, gwclsName, "must greater than 1 and less than 253"))
	}
	return errs
}

func validateGatewayAddresses(addresses []gatewayv1a2.GatewayAddress, path *field.Path) field.ErrorList {
	var errs field.ErrorList
	if len(addresses) > 16 {
		errs = append(errs, field.Invalid(path, "addresses length", " maximum is 16 items."))
	}
	return errs
}

// validateGatewayListeners validates whether required fields of listeners are set according
// to the Gateway API specification.
func validateGatewayListeners(listeners []gatewayv1a2.Listener, path *field.Path) field.ErrorList {
	// TODO [danehans]: Add additional validation of listener fields.
	return validateListenerHostname(listeners, path)
}

// validateListenerHostname validates each listener hostname is not an IP address and is one
// of the following:
//  - A fully qualified domain name of a network host, as defined by RFC 3986.
//  - A DNS subdomain as defined by RFC 1123.
//  - A wildcard DNS subdomain as defined by RFC 1034 (section 4.3.3).
func validateListenerHostname(listeners []gatewayv1a2.Listener, path *field.Path) field.ErrorList {
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
