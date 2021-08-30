package utils

import (
	"testing"

	gatewayv1a2 "sigs.k8s.io/gateway-api/apis/v1alpha2"
)

func Test_Utils(t *testing.T) {

	var exportedPort1 gatewayv1a2.PortNumber = 65535
	var exportedPort2 gatewayv1a2.PortNumber = 65536
	var exportedPort3 gatewayv1a2.PortNumber = 1

	table := []struct {
		pathType     string
		expectedPath gatewayv1a2.PathMatchType
		port         int
		expectedPort *gatewayv1a2.PortNumber
	}{
		{
			pathType:     "Exact",
			expectedPath: gatewayv1a2.PathMatchExact,
			port:         0,
			expectedPort: nil,
		},
		{
			pathType:     "Exact",
			expectedPath: gatewayv1a2.PathMatchExact,
			port:         65535,
			expectedPort: &exportedPort1,
		},
		{
			pathType:     "Exact",
			expectedPath: gatewayv1a2.PathMatchExact,
			port:         65536,
			expectedPort: &exportedPort2,
		},
		{
			pathType:     "Prefix",
			expectedPath: gatewayv1a2.PathMatchPrefix,
			port:         0,
			expectedPort: nil,
		},
		{
			pathType:     "RegularExpression",
			expectedPath: gatewayv1a2.PathMatchRegularExpression,
			port:         65536,
			expectedPort: nil,
		},
		{
			pathType:     "ImplementationSpecific",
			expectedPath: gatewayv1a2.PathMatchImplementationSpecific,
			port:         1,
			expectedPort: &exportedPort3,
		},
	}

	for _, entry := range table {
		if path := PathMatchTypePtr(entry.pathType); path != &entry.expectedPath {
			t.Error("failed in path match type pointer.")
		}
		if port := PortNumberPtr(entry.port); port != entry.expectedPort {
			t.Error("failed in port number pointer.")
		}
	}
}
g