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

	"k8s.io/apimachinery/pkg/util/validation/field"

	gatewayv1b1 "sigs.k8s.io/gateway-api/apis/v1beta1"
)

var (
	// set of protocols for which we need to validate that hostname is empty
	protocolsHostnameInvalid = map[gatewayv1b1.ProtocolType]struct{}{
		gatewayv1b1.TCPProtocolType: {},
		gatewayv1b1.UDPProtocolType: {},
	}
	// set of protocols for which TLSConfig shall not be present
	protocolsTLSInvalid = map[gatewayv1b1.ProtocolType]struct{}{
		gatewayv1b1.HTTPProtocolType: {},
		gatewayv1b1.UDPProtocolType:  {},
		gatewayv1b1.TCPProtocolType:  {},
	}
	// set of protocols for which TLSConfig must be set
	protocolsTLSRequired = map[gatewayv1b1.ProtocolType]struct{}{
		gatewayv1b1.HTTPSProtocolType: {},
		gatewayv1b1.TLSProtocolType:   {},
	}
)

// ValidateGateway validates gw according to the Gateway API specification.
// For additional details of the Gateway spec, refer to:
//
//	https://gateway-api.sigs.k8s.io/v1beta1/references/spec/#gateway.networking.k8s.io/v1beta1.Gateway
//
// Validation that is not possible with CRD annotations may be added here in the future.
// See https://github.com/kubernetes-sigs/gateway-api/issues/868 for more information.
func ValidateGateway(gw *gatewayv1b1.Gateway) field.ErrorList {
	return ValidateGatewaySpec(&gw.Spec, field.NewPath("spec"))
}

// ValidateGatewaySpec validates whether required fields of spec are set according to the
// Gateway API specification.
func ValidateGatewaySpec(spec *gatewayv1b1.GatewaySpec, path *field.Path) field.ErrorList {
	var errs field.ErrorList
	errs = append(errs, validateGatewayListeners(spec.Listeners, path.Child("listeners"))...)
	return errs
}

// validateGatewayListeners validates whether required fields of listeners are set according
// to the Gateway API specification.
func validateGatewayListeners(listeners []gatewayv1b1.Listener, path *field.Path) field.ErrorList {
	var errs field.ErrorList
	errs = append(errs, ValidateListenerTLSConfig(listeners, path)...)
	errs = append(errs, validateListenerHostname(listeners, path)...)
	errs = append(errs, ValidateTLSCertificateRefs(listeners, path)...)
	return errs
}

// validateListenerTLSConfig validates TLS config must be set when protocol is HTTPS or TLS,
// and TLS config shall not be present when protocol is HTTP, TCP or UDP
func ValidateListenerTLSConfig(listeners []gatewayv1b1.Listener, path *field.Path) field.ErrorList {
	var errs field.ErrorList
	for i, l := range listeners {
		if isProtocolInSubset(l.Protocol, protocolsTLSRequired) && l.TLS == nil {
			errs = append(errs, field.Forbidden(path.Index(i).Child("tls"), fmt.Sprintf("must be set for protocol %v", l.Protocol)))
		}
		if isProtocolInSubset(l.Protocol, protocolsTLSInvalid) && l.TLS != nil {
			errs = append(errs, field.Forbidden(path.Index(i).Child("tls"), fmt.Sprintf("should be empty for protocol %v", l.Protocol)))
		}
	}
	return errs
}

func isProtocolInSubset(protocol gatewayv1b1.ProtocolType, set map[gatewayv1b1.ProtocolType]struct{}) bool {
	_, ok := set[protocol]
	return ok
}

// validateListenerHostname validates each listener hostname
// should be empty in case protocol is TCP or UDP
func validateListenerHostname(listeners []gatewayv1b1.Listener, path *field.Path) field.ErrorList {
	var errs field.ErrorList
	for i, h := range listeners {
		if isProtocolInSubset(h.Protocol, protocolsHostnameInvalid) && h.Hostname != nil {
			errs = append(errs, field.Forbidden(path.Index(i).Child("hostname"), fmt.Sprintf("should be empty for protocol %v", h.Protocol)))
		}
	}
	return errs
}

// ValidateTLSCertificateRefs validates the certificateRefs
// must be set and not empty when tls config is set and
// TLSModeType is terminate
func ValidateTLSCertificateRefs(listeners []gatewayv1b1.Listener, path *field.Path) field.ErrorList {
	var errs field.ErrorList
	for i, c := range listeners {
		if isProtocolInSubset(c.Protocol, protocolsTLSRequired) && c.TLS != nil {
			if *c.TLS.Mode == gatewayv1b1.TLSModeTerminate && len(c.TLS.CertificateRefs) == 0 {
				errs = append(errs, field.Forbidden(path.Index(i).Child("tls").Child("certificateRefs"), fmt.Sprintln("should be set and not empty when TLSModeType is Terminate")))
			}
		}
	}
	return errs
}
