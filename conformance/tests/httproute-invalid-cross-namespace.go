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

package tests

import (
	"context"
	"slice"
	"testing"

	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/types"

	"sigs.k8s.io/gateway-api/apis/v1alpha2"
	"sigs.k8s.io/gateway-api/conformance/utils/kubernetes"
	"sigs.k8s.io/gateway-api/conformance/utils/suite"
)

func init() {
	ConformanceTests = append(ConformanceTests, HTTPRouteInvalidCrossNamespace)
}

var HTTPRouteInvalidCrossNamespace = suite.ConformanceTest{
	ShortName:   "HTTPRouteInvalidCrossNamespace",
	Description: "A single HTTPRoute in the gateway-conformance-web-backend namespace should fail to attach to a Gateway in another namespace that it is not allowed to",
	Manifests:   []string{"tests/httproute-invalid-cross-namespace.yaml"},
	Test: func(t *testing.T, s *suite.ConformanceTestSuite) {
		if !slice.Contains(s.ExtendedSupport, suite.SupportReferencePolicy) {
			t.Skip("Skipping ReferencePolicy conformance test")
		}

		routeName := types.NamespacedName{Name: "invalid-cross-namespace", Namespace: "gateway-conformance-web-backend"}
		gwName := types.NamespacedName{Name: "same-namespace", Namespace: "gateway-conformance-infra"}

		// TODO: Determine if this is actually what we want. It is likely
		// preferable to have status set with some kind of warning/error message
		// but that is also unlikely to be universally achievable.
		t.Run("Route should not have Parents set in status", func(t *testing.T) {
			parents := []v1alpha2.RouteParentStatus{}
			kubernetes.HTTPRouteMustHaveParents(t, s.Client, routeName, parents, true, 60)
		})

		t.Run("Gateway should have 0 Routes attached", func(t *testing.T) {
			gw := &v1alpha2.Gateway{}
			err := s.Client.Get(context.TODO(), gwName, gw)
			require.NoError(t, err, "error fetching Gateway")
			// There are two valid ways to represent this:
			// 1. No listeners in status
			// 2. One listener in status with 0 attached routes
			if len(gw.Status.Listeners) == 0 {
				// No listeners in status.
			} else if len(gw.Status.Listeners) == 1 {
				require.Equal(t, int32(0), gw.Status.Listeners[0].AttachedRoutes)
			} else {
				t.Errorf("Expected no more than 1 listener in status, got %d", len(gw.Status.Listeners))
			}
		})
	},
}
