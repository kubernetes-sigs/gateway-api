/*
Copyright 2026 The Kubernetes Authors.

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

package tests

import (
	"context"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	v1 "sigs.k8s.io/gateway-api/apis/v1"
	"sigs.k8s.io/gateway-api/conformance/utils/kubernetes"
	"sigs.k8s.io/gateway-api/conformance/utils/suite"
	"sigs.k8s.io/gateway-api/pkg/features"
)

func init() {
	ConformanceTests = append(ConformanceTests, GatewayMaxLengthName)
}

var GatewayMaxLengthName = suite.ConformanceTest{
	ShortName:   "GatewayMaxLengthName",
	Description: "A Gateway with a name at the maximum allowed Kubernetes resource name length (253 characters) should be handled gracefully without crashing or blocking",
	Features: []features.FeatureName{
		features.SupportGateway,
	},
	Provisional: true,
	Parallel:    true,
	Test: func(t *testing.T, s *suite.ConformanceTestSuite) {
		// Generate a name at the maximum DNS subdomain length (253 chars).
		// Must follow RFC 1123 subdomain: lowercase alphanumeric, start/end with alphanumeric.
		name := "g" + strings.Repeat("a", 252)

		gwNN := types.NamespacedName{
			Name:      name,
			Namespace: suite.InfrastructureNamespace,
		}

		ctx, cancel := context.WithTimeout(context.Background(), s.TimeoutConfig.DefaultTestTimeout)
		defer cancel()

		gw := &v1.Gateway{
			ObjectMeta: metav1.ObjectMeta{
				Name:      name,
				Namespace: suite.InfrastructureNamespace,
			},
			Spec: v1.GatewaySpec{
				GatewayClassName: v1.ObjectName(s.GatewayClassName),
				Listeners: []v1.Listener{
					{
						Name:     v1.SectionName("http"),
						Port:     v1.PortNumber(8080),
						Protocol: v1.HTTPProtocolType,
					},
				},
			},
		}

		t.Logf("Creating Gateway with name of length %d", len(name))
		err := s.Client.Create(ctx, gw)
		require.NoErrorf(t, err, "error creating Gateway with max-length name")

		t.Cleanup(func() {
			ctx, cancel := context.WithTimeout(context.Background(), s.TimeoutConfig.DeleteTimeout)
			defer cancel()
			err := s.Client.Delete(ctx, gw)
			if err != nil {
				t.Logf("error deleting Gateway: %v", err)
			}
		})

		t.Logf("Waiting for Gateway %s/%s to be processed by the controller", gwNN.Namespace, gwNN.Name)
		kubernetes.GatewayMustHaveLatestConditions(t, s.Client, s.TimeoutConfig, gwNN)

		t.Logf("Reading Gateway %s/%s to inspect conditions", gwNN.Namespace, gwNN.Name)
		currentGW := &v1.Gateway{}
		err = s.Client.Get(ctx, gwNN, currentGW)
		require.NoError(t, err, "error getting Gateway: %v", err)

		for _, cond := range currentGW.Status.Conditions {
			t.Logf("Gateway condition: type=%s status=%s reason=%s message=%s", cond.Type, cond.Status, cond.Reason, cond.Message)
		}
		if len(currentGW.Status.Conditions) == 0 {
			t.Logf("No conditions set on Gateway %s/%s", gwNN.Namespace, gwNN.Name)
		}

		t.Logf("Gateway %s/%s handled successfully (conditions logged above)", gwNN.Namespace, gwNN.Name)
	},
}
