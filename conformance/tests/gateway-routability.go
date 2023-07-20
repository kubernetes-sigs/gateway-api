/*
Copyright 2023 The Kubernetes Authors.

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
	"regexp"
	"testing"

	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/sets"

	"sigs.k8s.io/gateway-api/apis/v1beta1"
	"sigs.k8s.io/gateway-api/conformance/utils/kubernetes"
	"sigs.k8s.io/gateway-api/conformance/utils/suite"
)

func init() {
	ConformanceTests = append(ConformanceTests, GatewayClassRoutability)
	ConformanceTests = append(ConformanceTests, GatewayPublicRoutability)
	ConformanceTests = append(ConformanceTests, GatewayPrivateRoutability)
	ConformanceTests = append(ConformanceTests, GatewayClusterRoutability)
	ConformanceTests = append(ConformanceTests, GatewayUnsupportedRoutability)
}

var GatewayClassRoutability = suite.ConformanceTest{
	ShortName: "GatewayClassRoutability",
	Features: []suite.SupportedFeature{
		suite.SupportGateway,
		suite.SupportGatewayClassRoutability,
	},
	Manifests: []string{
		"tests/gateway-routability-default.yaml",
	},
	Description: "A GatewayClass MUST list routabilities in its status. The first entry should be the default value for Gateways",
	Test: func(t *testing.T, s *suite.ConformanceTestSuite) {
		// At this point the suite setup checks for GatewayClass conditions
		// are all True using the kubernetes.NamespacesMustBeReady helper.
		ctx, cancel := context.WithTimeout(context.Background(), s.TimeoutConfig.GetTimeout)
		defer cancel()

		className := types.NamespacedName{Name: s.GatewayClassName}
		gwc := &v1beta1.GatewayClass{}
		err := s.Client.Get(ctx, className, gwc)
		require.NoErrorf(t, err, "error getting GatewayClass: %v", err)

		require.NotEmpty(t, gwc.Status.Routabilities, "A GatewayClass that supports routability MUST list them in Status.Routabilities")

		gwn := types.NamespacedName{Name: "gateway-default-routability", Namespace: "gateway-conformance-infra"}
		//nolint:errcheck // the helper throws an error if it fails
		kubernetes.WaitForGatewayAddress(t, s.Client, s.TimeoutConfig, gwn)

		ctx, cancel = context.WithTimeout(context.Background(), s.TimeoutConfig.GetTimeout)
		defer cancel()

		gw := &v1beta1.Gateway{}
		err = s.Client.Get(ctx, gwn, gw)
		require.NoErrorf(t, err, "error getting Gateway: %v", err)

		expectedRoutability := gwc.Status.Routabilities[0]
		actualRoutability := gw.Status.Addresses[0].Routability
		require.NotNil(t, actualRoutability, "expected gateway address to have set the routability")
		require.Equal(t, expectedRoutability, *actualRoutability, "the first entry in the GatewayClass.Status.Routabilities should be the default routability")
	},
}

var GatewayUnsupportedRoutability = suite.ConformanceTest{
	ShortName: "GatewayUnsupportedRoutability",
	Features: []suite.SupportedFeature{
		suite.SupportGateway,
		suite.SupportGatewayClassRoutability,
	},
	Manifests: []string{
		"tests/gateway-routability-broken.yaml",
	},
	Description: "A Gateway should set Accepted condition to False when it doesn't support a routability",
	Test: func(t *testing.T, s *suite.ConformanceTestSuite) {
		gwn := types.NamespacedName{Name: "gateway-broken-routability", Namespace: "gateway-conformance-infra"}
		kubernetes.GatewayMustHaveLatestConditions(t, s.Client, s.TimeoutConfig, gwn)

		ctx, cancel := context.WithTimeout(context.Background(), s.TimeoutConfig.GetTimeout)
		defer cancel()

		gw := &v1beta1.Gateway{}
		err := s.Client.Get(ctx, gwn, gw)
		require.NoErrorf(t, err, "error getting Gateway: %v", err)

		for _, cond := range gw.Status.Conditions {
			if cond.Type == string(v1beta1.GatewayConditionAccepted) {
				if cond.Status != metav1.ConditionFalse {
					t.Errorf("expected Accepted condition to be 'False': was %q", cond.Status)
				} else if cond.Reason != string(v1beta1.GatewayUnsupportedRoutability) {
					t.Errorf("expected Accepted condition reason to be %q: was %q", v1beta1.GatewayUnsupportedRoutability, cond.Status)
				}
			}
		}
	},
}

var GatewayPublicRoutability = suite.ConformanceTest{
	ShortName: "GatewayPublicRoutability",
	Features: []suite.SupportedFeature{
		suite.SupportGateway,
		suite.SupportGatewayClassRoutability,
		suite.SupportGatewayPublicRoutability,
	},
	Description: "A Gateway supports Public routability",
	Manifests: []string{
		"tests/gateway-routability-public.yaml",
	},
	Test: func(t *testing.T, s *suite.ConformanceTestSuite) {
		namespaces := []string{"gateway-conformance-infra"}
		kubernetes.NamespacesMustBeReady(t, s.Client, s.TimeoutConfig, namespaces)

		ctx, cancel := context.WithTimeout(context.Background(), s.TimeoutConfig.GetTimeout)
		defer cancel()

		gwn := types.NamespacedName{Name: "gateway-public-routability", Namespace: "gateway-conformance-infra"}
		gw := &v1beta1.Gateway{}
		err := s.Client.Get(ctx, gwn, gw)
		require.NoErrorf(t, err, "error getting Gateway: %v", err)

		allowedEntries := sets.New(
			v1beta1.GatewayRoutabilityPublic,
			v1beta1.GatewayRoutabilityPrivate,
			v1beta1.GatewayRoutabilityCluster,
		)

		validateAddresses(t, gw.Status.Addresses, allowedEntries)
	},
}

var GatewayPrivateRoutability = suite.ConformanceTest{
	ShortName: "GatewayPrivateRoutability",
	Features: []suite.SupportedFeature{
		suite.SupportGateway,
		suite.SupportGatewayClassRoutability,
		suite.SupportGatewayPrivateRoutability,
	},
	Description: "A Gateway supports Private routability",
	Manifests: []string{
		"tests/gateway-routability-private.yaml",
	},
	Test: func(t *testing.T, s *suite.ConformanceTestSuite) {
		namespaces := []string{"gateway-conformance-infra"}
		kubernetes.NamespacesMustBeReady(t, s.Client, s.TimeoutConfig, namespaces)

		ctx, cancel := context.WithTimeout(context.Background(), s.TimeoutConfig.GetTimeout)
		defer cancel()

		gwn := types.NamespacedName{Name: "gateway-private-routability", Namespace: "gateway-conformance-infra"}
		gw := &v1beta1.Gateway{}
		err := s.Client.Get(ctx, gwn, gw)
		require.NoErrorf(t, err, "error getting Gateway: %v", err)

		allowedEntries := sets.New(
			v1beta1.GatewayRoutabilityPrivate,
			v1beta1.GatewayRoutabilityCluster,
		)

		validateAddresses(t, gw.Status.Addresses, allowedEntries)
	},
}

var GatewayClusterRoutability = suite.ConformanceTest{
	ShortName: "GatewayClusterRoutability",
	Features: []suite.SupportedFeature{
		suite.SupportGateway,
		suite.SupportGatewayClassRoutability,
		suite.SupportGatewayClusterRoutability,
	},
	Description: "A Gateway supports Cluster routability",
	Manifests: []string{
		"tests/gateway-routability-cluster.yaml",
	},
	Test: func(t *testing.T, s *suite.ConformanceTestSuite) {
		namespaces := []string{"gateway-conformance-infra"}
		kubernetes.NamespacesMustBeReady(t, s.Client, s.TimeoutConfig, namespaces)

		ctx, cancel := context.WithTimeout(context.Background(), s.TimeoutConfig.GetTimeout)
		defer cancel()

		gwn := types.NamespacedName{Name: "gateway-cluster-routability", Namespace: "gateway-conformance-infra"}
		gw := &v1beta1.Gateway{}
		err := s.Client.Get(ctx, gwn, gw)
		require.NoErrorf(t, err, "error getting Gateway: %v", err)

		allowedEntries := sets.New(
			v1beta1.GatewayRoutabilityCluster,
		)

		validateAddresses(t, gw.Status.Addresses, allowedEntries)
	},
}

var vendorPrefixedRoutability = regexp.MustCompile(`^[a-z0-9]([-a-z0-9]*[a-z0-9])?(\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*\/[A-Za-z0-9\/\-_]+$`)

func validateAddresses(t *testing.T, addrs []v1beta1.GatewayStatusAddress, allowedEntries sets.Set[v1beta1.GatewayRoutability]) {
	t.Helper()

	require.NotEmpty(t, addrs, "A Gateway that supports routability must have addresses")

	for _, addr := range addrs {
		require.NotNilf(t, addr.Routability, "expected GatewayStatusAddress '%s/%s' to have a non-nil routability", *addr.Type, addr.Value)

		addressRoutability := *addr.Routability

		// Vendor prefix values are allowed to be present
		if vendorPrefixedRoutability.MatchString(string(addressRoutability)) {
			continue
		}

		require.Truef(t, allowedEntries.Has(addressRoutability), "Unexpected routability value: %q", addressRoutability)
	}
}
