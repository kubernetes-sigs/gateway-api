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

package suite

import (
	"fmt"

	"k8s.io/apimachinery/pkg/util/sets"

	"sigs.k8s.io/gateway-api/pkg/features"
)

// -----------------------------------------------------------------------------
// Conformance Profiles - Public Types
// -----------------------------------------------------------------------------

// ConformanceProfile is a group of features that have a related purpose, e.g.
// to cover specific protocol support or a specific feature present in Gateway
// API.
//
// For more details see the relevant GEP: https://gateway-api.sigs.k8s.io/geps/gep-1709/
type ConformanceProfile struct {
	Name             ConformanceProfileName
	CoreFeatures     sets.Set[features.SupportedFeature]
	ExtendedFeatures sets.Set[features.SupportedFeature]
}

type ConformanceProfileName string

const (
	// HTTPConformanceProfileName indicates the name of the conformance profile
	// which covers HTTP functionality, such as the HTTPRoute API.
	HTTPConformanceProfileName ConformanceProfileName = "HTTP"

	// TLSConformanceProfileName indicates the name of the conformance profile
	// which covers TLS stream functionality, such as the TLSRoute API.
	TLSConformanceProfileName ConformanceProfileName = "TLS"

	// GRPCConformanceProfileName indicates the name of the conformance profile
	// which covers GRPC functionality, such as the GRPCRoute API.
	GRPCConformanceProfileName ConformanceProfileName = "GRPC"

	// MeshConformanceProfileName indicates the name of the conformance profile
	// which covers service mesh functionality.
	MeshConformanceProfileName ConformanceProfileName = "MESH"
)

// -----------------------------------------------------------------------------
// Conformance Profiles - Public Vars
// -----------------------------------------------------------------------------

var (
	// HTTPConformanceProfile is a ConformanceProfile that covers testing HTTP
	// related functionality with Gateways.
	HTTPConformanceProfile = ConformanceProfile{
		Name: HTTPConformanceProfileName,
		CoreFeatures: sets.New(
			features.SupportGateway,
			features.SupportReferenceGrant,
			features.SupportHTTPRoute,
		),
		ExtendedFeatures: sets.New[features.SupportedFeature]().
			Insert(features.GatewayExtendedFeatures.UnsortedList()...).
			Insert(features.HTTPRouteExtendedFeatures.UnsortedList()...),
	}

	// TLSConformanceProfile is a ConformanceProfile that covers testing TLS
	// related functionality with Gateways.
	TLSConformanceProfile = ConformanceProfile{
		Name: TLSConformanceProfileName,
		CoreFeatures: sets.New(
			features.SupportGateway,
			features.SupportReferenceGrant,
			features.SupportTLSRoute,
		),
		ExtendedFeatures: features.GatewayExtendedFeatures,
	}

	// GRPCConformanceProfile is a ConformanceProfile that covers testing GRPC
	// related functionality with Gateways.
	GRPCConformanceProfile = ConformanceProfile{
		Name: GRPCConformanceProfileName,
		CoreFeatures: sets.New(
			features.SupportGateway,
			features.SupportReferenceGrant,
			features.SupportGRPCRoute,
		),
	}

	// MeshConformanceProfile is a ConformanceProfile that covers testing
	// service mesh related functionality.
	MeshConformanceProfile = ConformanceProfile{
		Name: MeshConformanceProfileName,
		CoreFeatures: sets.New(
			features.SupportMesh,
			features.SupportHTTPRoute,
		),
		ExtendedFeatures: features.HTTPRouteExtendedFeatures,
	}
)

// RegisterConformanceProfile allows downstream tests to register unique profiles that
// define their own set of features
func RegisterConformanceProfile(p ConformanceProfile) {
	_, ok := conformanceProfileMap[p.Name]
	if ok {
		panic(fmt.Sprintf("ConformanceProfile named %q is already registered", p.Name))
	}
	conformanceProfileMap[p.Name] = p
}

// -----------------------------------------------------------------------------
// Conformance Profiles - Private Profile Mapping Helpers
// -----------------------------------------------------------------------------

// conformanceProfileMap maps short human-readable names to their respective
// ConformanceProfiles.
var conformanceProfileMap = map[ConformanceProfileName]ConformanceProfile{
	HTTPConformanceProfileName: HTTPConformanceProfile,
	TLSConformanceProfileName:  TLSConformanceProfile,
	GRPCConformanceProfileName: GRPCConformanceProfile,
	MeshConformanceProfileName: MeshConformanceProfile,
}

// getConformanceProfileForName retrieves a known ConformanceProfile by its simple
// human readable ConformanceProfileName.
func getConformanceProfileForName(name ConformanceProfileName) (ConformanceProfile, error) {
	profile, ok := conformanceProfileMap[name]
	if !ok {
		return profile, fmt.Errorf("%s is not a valid conformance profile", name)
	}

	return profile, nil
}

// getConformanceProfilesForTest retrieves the ConformanceProfiles a test belongs to.
func getConformanceProfilesForTest(test ConformanceTest, conformanceProfiles sets.Set[ConformanceProfileName]) sets.Set[*ConformanceProfile] {
	matchingConformanceProfiles := sets.New[*ConformanceProfile]()
	for _, conformanceProfileName := range conformanceProfiles.UnsortedList() {
		cp := conformanceProfileMap[conformanceProfileName]
		hasAllFeatures := true
		for _, feature := range test.Features {
			if !cp.CoreFeatures.Has(feature) && !cp.ExtendedFeatures.Has(feature) {
				hasAllFeatures = false
				break
			}
		}
		if hasAllFeatures {
			matchingConformanceProfiles.Insert(&cp)
		}
	}

	return matchingConformanceProfiles
}
