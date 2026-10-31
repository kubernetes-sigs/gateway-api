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

package main

import "testing"

func TestProcessFeatureName(t *testing.T) {
	tests := []struct{ in, want string }{
		{"", ""},
		{"lowercase", ""},
		{"BackendTLSPolicy", "Backend TLS Policy"},
		{"BackendTLSPolicySANValidation", "Backend TLS Policy SAN Validation"},
		{"GatewayAddressEmpty", "Gateway Address Empty"},
		{"GatewayBackendClientCertificate", "Gateway Backend Client Certificate"},
		{"GatewayFrontendClientCertificateValidation", "Gateway Frontend Client Certificate Validation"},
		{"GatewayFrontendClientCertificateValidationInsecureFallback", "Gateway Frontend Client Certificate Validation Insecure Fallback"},
		{"GatewayHTTPListenerIsolation", "Gateway HTTP Listener Isolation"},
		{"GatewayHTTPSListenerDetectMisdirectedRequests", "Gateway HTTPS Listener Detect Misdirected Requests"},
		{"GatewayInfrastructure", "Gateway Infrastructure"},
		{"GatewayInfrastructurePropagation", "Gateway Infrastructure Propagation"},
		{"GatewayPort8080", "Gateway Port 8080"},
		{"GatewayStaticAddresses", "Gateway Static Addresses"},
		{"HTTPResponseHeaderModification", "HTTP Response Header Modification"},
		{"HTTPRoute303RedirectStatusCode", "HTTPRoute 303 Redirect Status Code"},
		{"HTTPRoute307RedirectStatusCode", "HTTPRoute 307 Redirect Status Code"},
		{"HTTPRoute308RedirectStatusCode", "HTTPRoute 308 Redirect Status Code"},
		{"HTTPRouteBackendProtocolH2C", "HTTPRoute Backend Protocol H2C"},
		{"HTTPRouteBackendProtocolWebSocket", "HTTPRoute Backend Protocol Web Socket"},
		{"HTTPRouteBackendRequestHeaderModification", "HTTPRoute Backend Request Header Modification"},
		{"HTTPRouteBackendTimeout", "HTTPRoute Backend Timeout"},
		{"HTTPRouteCORS", "HTTPRoute CORS"},
		{"HTTPRouteDestinationPortMatching", "HTTPRoute Destination Port Matching"},
		{"HTTPRouteHostRewrite", "HTTPRoute Host Rewrite"},
		{"HTTPRouteMethodMatching", "HTTPRoute Method Matching"},
		{"HTTPRouteNamedRouteRule", "HTTPRoute Named Route Rule"},
		{"HTTPRouteParentRefPort", "HTTPRoute Parent Ref Port"},
		{"HTTPRoutePathRedirect", "HTTPRoute Path Redirect"},
		{"HTTPRoutePathRewrite", "HTTPRoute Path Rewrite"},
		{"HTTPRoutePortRedirect", "HTTPRoute Port Redirect"},
		{"HTTPRouteQueryParamMatching", "HTTPRoute Query Param Matching"},
		{"HTTPRouteRequestMirror", "HTTPRoute Request Mirror"},
		{"HTTPRouteRequestMultipleMirrors", "HTTPRoute Request Multiple Mirrors"},
		{"HTTPRouteRequestPercentageMirror", "HTTPRoute Request Percentage Mirror"},
		{"HTTPRouteRequestTimeout", "HTTPRoute Request Timeout"},
		{"HTTPRouteResponseHeaderModification", "HTTPRoute Response Header Modification"},
		{"HTTPRouteRetry", "HTTPRoute Retry"},
		{"HTTPRouteRetryBackendTimeout", "HTTPRoute Retry Backend Timeout"},
		{"HTTPRouteRetryConnectionError", "HTTPRoute Retry Connection Error"},
		{"HTTPRouteSchemeRedirect", "HTTPRoute Scheme Redirect"},
		{"ListenerSet", "Listener Set"},
		{"MeshClusterIPMatching", "Mesh Cluster IP Matching"},
		{"MeshConsumerRoute", "Mesh Consumer Route"},
		{"MeshHTTPRouteBackendRequestHeaderModification", "Mesh HTTPRoute Backend Request Header Modification"},
		{"MeshHTTPRouteNamedRouteRule", "Mesh HTTPRoute Named Route Rule"},
		{"MeshHTTPRouteQueryParamMatching", "Mesh HTTPRoute Query Param Matching"},
		{"MeshHTTPRouteRedirectPath", "Mesh HTTPRoute Redirect Path"},
		{"MeshHTTPRouteRedirectPort", "Mesh HTTPRoute Redirect Port"},
		{"MeshHTTPRouteRewritePath", "Mesh HTTPRoute Rewrite Path"},
		{"MeshHTTPRouteSchemeRedirect", "Mesh HTTPRoute Scheme Redirect"},
		{"TCPRoute", "TCP Route"},
		{"TLSRouteModeMixed", "TLS Route Mode Mixed"},
		{"TLSRouteModeTerminate", "TLS Route Mode Terminate"},
		{"UDPRoute", "UDP Route"},
	}
	for _, tt := range tests {
		t.Run(tt.in, func(t *testing.T) {
			if got := ProcessFeatureName(tt.in); got != tt.want {
				t.Errorf("ProcessFeatureName(%q) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}
