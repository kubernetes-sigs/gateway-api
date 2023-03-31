//go:build experimental
// +build experimental

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

package suite

import (
	"fmt"
	"strings"

	"k8s.io/apimachinery/pkg/util/sets"
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
	CoreFeatures     sets.Set[SupportedFeature]
	ExtendedFeatures sets.Set[SupportedFeature]
}

type ConformanceProfileName string

const (
	// HTTPConformanceProfileName indicates the name of the conformance profile
	// which covers HTTP functionality, such as the HTTPRoute API.
	HTTPConformanceProfileName ConformanceProfileName = "HTTP"

	// TLSConformanceProfileName indicates the name of the conformance profile
	// which covers TLS stream functionality, such as the TLSRoute API.
	TLSConformanceProfileName ConformanceProfileName = "TLS"

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
			SupportGateway,
			SupportReferenceGrant,
			SupportHTTPRoute,
		),
		ExtendedFeatures: sets.New(
			SupportHTTPRouteQueryParamMatching,
			SupportHTTPRouteMethodMatching,
			SupportHTTPResponseHeaderModification,
			SupportHTTPRoutePortRedirect,
			SupportHTTPRouteSchemeRedirect,
			SupportHTTPRoutePathRedirect,
			SupportHTTPRouteHostRewrite,
			SupportHTTPRoutePathRewrite,
		),
	}

	// TLSConformanceProfile is a ConformanceProfile that covers testing TLS
	// related functionality with Gateways.
	TLSConformanceProfile = ConformanceProfile{
		Name: TLSConformanceProfileName,
		CoreFeatures: sets.New(
			SupportGateway,
			SupportReferenceGrant,
			SupportTLSRoute,
		),
	}

	// MeshConformanceProfile is a ConformanceProfile that covers testing
	// service mesh related functionality.
	MeshConformanceProfile = ConformanceProfile{
		Name: MeshConformanceProfileName,
		CoreFeatures: sets.New(
			SupportMesh,
		),
	}
)

// -----------------------------------------------------------------------------
// Conformance Profiles - Private Profile Mapping Helpers
// -----------------------------------------------------------------------------

// conformanceProfileMap maps short human-readable names to their respective
// ConformanceProfiles.
var conformanceProfileMap = map[ConformanceProfileName]ConformanceProfile{
	HTTPConformanceProfileName: HTTPConformanceProfile,
	TLSConformanceProfileName:  TLSConformanceProfile,
	MeshConformanceProfileName: MeshConformanceProfile,
}

// getConformanceProfileForName retrieves a known ConformanceProfile by it's simple
// human readable ConformanceProfileName.
func getConformanceProfileForName(name ConformanceProfileName) (ConformanceProfile, error) {
	profile, ok := conformanceProfileMap[name]
	if !ok {
		return profile, fmt.Errorf("%s is not a valid conformance profile", name)
	}

	return profile, nil
}

// getConformanceProfileForTest retrieves the ConformanceProfile a test belongs to
// given the name of the test.
//
// TODO: this is a hack right now using the name of the test itself to determine
// what profile a test belongs to. If we take this past the
// Prototyping/Provisional phase we should look into associating profiles more
// directly with the test (perhaps ON the tests like features).
func getConformanceProfileForTest(name string) (ConformanceProfile, error) {
	var conformanceProfileName ConformanceProfileName
	switch {
	case strings.HasPrefix(name, string(HTTPConformanceProfileName)):
		conformanceProfileName = HTTPConformanceProfileName
	case strings.HasPrefix(name, string(TLSConformanceProfileName)):
		conformanceProfileName = TLSConformanceProfileName
	case strings.HasPrefix(name, string(MeshConformanceProfileName)):
		conformanceProfileName = MeshConformanceProfileName
	}
	return getConformanceProfileForName(conformanceProfileName)
}
