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
	"regexp"

	"k8s.io/apimachinery/pkg/util/validation/field"

	gatewayv1a2 "sigs.k8s.io/gateway-api/apis/v1beta1"
)

var (
	// set of protocols for which we need to validate that hostname is empty
	protocolsHostnameInvalid = map[gatewayv1a2.ProtocolType]struct{}{
		gatewayv1a2.TCPProtocolType: {},
		gatewayv1a2.UDPProtocolType: {},
	}
	// set of protocols for which TLSConfig shall not be present
	protocolsTLSInvalid = map[gatewayv1a2.ProtocolType]struct{}{
		gatewayv1a2.HTTPProtocolType: {},
		gatewayv1a2.UDPProtocolType:  {},
		gatewayv1a2.TCPProtocolType:  {},
	}

	addressTypesValid = map[gatewayv1a2.AddressType]struct{}{
		gatewayv1a2.HostnameAddressType: {},
		gatewayv1a2.IPAddressType:       {},
	}
)

// ValidateGateway validates gw according to the Gateway API specification.
// For additional details of the Gateway spec, refer to:
//  https://gateway-api.sigs.k8s.io/v1beta1/references/spec/#gateway.networking.k8s.io/v1beta1.Gateway
//
// Validation that is not possible with CRD annotations may be added here in the future.
// See https://github.com/kubernetes-sigs/gateway-api/issues/868 for more information.
func ValidateGateway(gw *gatewayv1a2.Gateway) field.ErrorList {
	return validateGatewaySpec(&gw.Spec, field.NewPath("spec"))
}

// validateGatewaySpec validates whether required fields of spec are set according to the
// Gateway API specification.
func validateGatewaySpec(spec *gatewayv1a2.GatewaySpec, path *field.Path) field.ErrorList {
	var errs field.ErrorList
	errs = append(errs, validateGatewayListeners(spec.Listeners, path.Child("listeners"))...)
	errs = append(errs, validateAddresses(spec.Addresses, path.Child("addresses"))...)
	return errs
}

// validateGatewayListeners validates whether required fields of listeners are set according
// to the Gateway API specification.
func validateGatewayListeners(listeners []gatewayv1a2.Listener, path *field.Path) field.ErrorList {
	var errs field.ErrorList
	errs = append(errs, validateListenerTLSConfig(listeners, path)...)
	errs = append(errs, validateListenerHostname(listeners, path)...)
	return errs
}

func validateListenerTLSConfig(listeners []gatewayv1a2.Listener, path *field.Path) field.ErrorList {
	var errs field.ErrorList
	for i, l := range listeners {
		if isProtocolInSubset(l.Protocol, protocolsTLSInvalid) && l.TLS != nil {
			errs = append(errs, field.Forbidden(path.Index(i).Child("tls"), fmt.Sprintf("should be empty for protocol %v", l.Protocol)))
		}
	}
	return errs
}

func isProtocolInSubset(protocol gatewayv1a2.ProtocolType, set map[gatewayv1a2.ProtocolType]struct{}) bool {
	_, ok := set[protocol]
	return ok
}

// validateListenerHostname validates each listener hostname
// should be empty in case protocol is TCP or UDP
func validateListenerHostname(listeners []gatewayv1a2.Listener, path *field.Path) field.ErrorList {
	var errs field.ErrorList
	for i, h := range listeners {
		if isProtocolInSubset(h.Protocol, protocolsHostnameInvalid) && h.Hostname != nil {
			errs = append(errs, field.Forbidden(path.Index(i).Child("hostname"), fmt.Sprintf("should be empty for protocol %v", h.Protocol)))
		}
	}
	return errs
}

// domainPrefixedStringRegex is a regex used in validation to determine whether
// a provided string is a domain-prefixed string. Domain-prefixed strings are used
// to indicate custom (implementation-specific) address types.
var domainPrefixedStringRegex = regexp.MustCompile(`^([a-zA-Z0-9][a-zA-Z0-9\-]*[a-zA-Z0-9]\.)*([A-Za-z0-9]|[A-Za-z0-9][A-Za-z0-9\-]*[A-Za-z0-9])\/[a-zA-Z0-9]+$`)

// validateAddresses validates each listener address
// if there are addresses set. Otherwise, returns no error.
func validateAddresses(addresses []gatewayv1a2.GatewayAddress, path *field.Path) field.ErrorList {
	var errs field.ErrorList

	for i, a := range addresses {
		if a.Type == nil {
			continue
		}
		_, ok := addressTypesValid[*a.Type]
		if !ok {
			// Found something that's not one of the upstream AddressTypes
			// Next, check for a domain-prefixed string
			match := domainPrefixedStringRegex.Match([]byte(*a.Type))
			if !match {
				errs = append(errs, field.Invalid(path.Index(i).Child("type"), a.Type, "should either be a defined constant or a domain-prefixed string (example.com/Type)"))
			}
		}

	}
	return errs
}
