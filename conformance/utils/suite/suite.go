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
	"context"
	"embed"
	"errors"
	"fmt"
	"sort"
	"strings"
	"sync"
	"testing"
	"time"

	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/sets"
	clientset "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"sigs.k8s.io/gateway-api/apis/v1beta1"
	"sigs.k8s.io/gateway-api/conformance"
	confv1 "sigs.k8s.io/gateway-api/conformance/apis/v1"
	"sigs.k8s.io/gateway-api/conformance/utils/config"
	"sigs.k8s.io/gateway-api/conformance/utils/flags"
	"sigs.k8s.io/gateway-api/conformance/utils/kubernetes"
	"sigs.k8s.io/gateway-api/conformance/utils/roundtripper"
	"sigs.k8s.io/gateway-api/pkg/consts"
)

// -----------------------------------------------------------------------------
// Conformance Test Suite - Public Types
// -----------------------------------------------------------------------------

// ConformanceTestSuite defines the test suite used to run Gateway API
// conformance tests.
type ConformanceTestSuite struct {
	Client                   client.Client
	Clientset                clientset.Interface
	RESTClient               *rest.RESTClient
	RestConfig               *rest.Config
	RoundTripper             roundtripper.RoundTripper
	GatewayClassName         string
	ControllerName           string
	Debug                    bool
	Cleanup                  bool
	BaseManifests            string
	MeshManifests            string
	Applier                  kubernetes.Applier
	SupportedFeatures        sets.Set[SupportedFeature]
	TimeoutConfig            config.TimeoutConfig
	SkipTests                sets.Set[string]
	RunTest                  string
	FS                       embed.FS
	UsableNetworkAddresses   []v1beta1.GatewayAddress
	UnusableNetworkAddresses []v1beta1.GatewayAddress

	// mode is the operating mode of the implementation.
	// The default value for it is "default".
	mode string

	// implementation contains the details of the implementation, such as
	// organization, project, etc.
	implementation confv1.Implementation

	// apiVersion is the version of the Gateway API installed in the cluster
	// and is extracted by the annotation gateway.networking.k8s.io/bundle-version
	// in the Gateway API CRDs.
	apiVersion string

	// apiChannel is the channel of the Gateway API installed in the cluster
	// and is extracted by the annotation gateway.networking.k8s.io/channel
	// in the Gateway API CRDs.
	apiChannel string

	// conformanceProfiles is a compiled list of profiles to check
	// conformance against.
	conformanceProfiles sets.Set[ConformanceProfileName]

	// running indicates whether the test suite is currently running.
	// Through this flag we prevent a Run() execution to happen in case
	// another Run() is ongoing.
	running bool

	// results stores the pass or fail results of each test that was run by
	// the test suite, organized by the tests unique name.
	results map[string]testResult

	// extendedSupportedFeatures is a compiled list of named features that were
	// marked as supported, and is used for reporting the test results.
	extendedSupportedFeatures map[ConformanceProfileName]sets.Set[SupportedFeature]

	// extendedUnsupportedFeatures is a compiled list of named features that were
	// marked as not supported, and is used for reporting the test results.
	extendedUnsupportedFeatures map[ConformanceProfileName]sets.Set[SupportedFeature]

	// lock is a mutex to help ensure thread safety of the test suite object.
	lock sync.RWMutex
}

// Options can be used to initialize a ConformanceTestSuite.
type ConformanceOptions struct {
	Client               client.Client
	Clientset            clientset.Interface
	RestConfig           *rest.Config
	GatewayClassName     string
	Debug                bool
	RoundTripper         roundtripper.RoundTripper
	BaseManifests        string
	MeshManifests        string
	NamespaceLabels      map[string]string
	NamespaceAnnotations map[string]string

	// CleanupBaseResources indicates whether or not the base test
	// resources such as Gateways should be cleaned up after the run.
	CleanupBaseResources       bool
	SupportedFeatures          sets.Set[SupportedFeature]
	ExemptFeatures             sets.Set[SupportedFeature]
	EnableAllSupportedFeatures bool
	TimeoutConfig              config.TimeoutConfig
	// SkipTests contains all the tests not to be run and can be used to opt out
	// of specific tests
	SkipTests []string
	// RunTest is a single test to run, mostly for development/debugging convenience.
	RunTest string

	FS *embed.FS

	// UsableNetworkAddresses is an optional pool of usable addresses for
	// Gateways for tests which need to test manual address assignments.
	UsableNetworkAddresses []v1beta1.GatewayAddress

	// UnusableNetworkAddresses is an optional pool of unusable addresses for
	// Gateways for tests which need to test failures with manual Gateway
	// address assignment.
	UnusableNetworkAddresses []v1beta1.GatewayAddress

	Mode                string
	AllowCRDsMismatch   bool
	Implementation      confv1.Implementation
	ConformanceProfiles sets.Set[ConformanceProfileName]
}

const (
	// undefinedKeyword is set in the ConformanceReport "GatewayAPIVersion" and
	// "GatewayAPIChannel" fields in case it's not possible to figure out the actual
	// values in the cluster, due to multiple versions of CRDs installed.
	undefinedKeyword = "UNDEFINED"
)

// NewConformanceTestSuite is a helper to use for creating a new ConformanceTestSuite.
func NewConformanceTestSuite(options ConformanceOptions) (*ConformanceTestSuite, error) {
	if err := apiextensionsv1.AddToScheme(options.Client.Scheme()); err != nil {
		return nil, err
	}

	config.SetupTimeoutConfig(&options.TimeoutConfig)

	roundTripper := options.RoundTripper
	if roundTripper == nil {
		roundTripper = &roundtripper.DefaultRoundTripper{Debug: options.Debug, TimeoutConfig: options.TimeoutConfig}
	}

	installedCRDs := &apiextensionsv1.CustomResourceDefinitionList{}
	err := options.Client.List(context.TODO(), installedCRDs)
	if err != nil {
		return nil, err
	}
	apiVersion, apiChannel, err := getAPIVersionAndChannel(installedCRDs.Items)
	if err != nil {
		// in case an error is returned and the AllowCRDsMismatch flag is false, the suite fails.
		// This is the default behavior but can be customized in case one wants to experiment
		// with mixed versions/channels of the API.
		if !options.AllowCRDsMismatch {
			return nil, err
		}
		apiVersion = undefinedKeyword
		apiChannel = undefinedKeyword
	}

	mode := flags.DefaultMode
	if options.Mode != "" {
		mode = options.Mode
	}

	if options.FS == nil {
		options.FS = &conformance.Manifests
	}

	// test suite callers are required to provide a conformance profile OR at
	// minimum a list of features which they support.
	if options.SupportedFeatures == nil && options.ConformanceProfiles.Len() == 0 && !options.EnableAllSupportedFeatures {
		return nil, fmt.Errorf("no conformance profile was selected for test run, and no supported features were provided so no tests could be selected")
	}

	// test suite callers can potentially just run all tests by saying they
	// cover all features, if they don't they'll need to have provided a
	// conformance profile or at least some specific features they support.
	if options.EnableAllSupportedFeatures {
		options.SupportedFeatures = AllFeatures
	} else if options.SupportedFeatures == nil {
		options.SupportedFeatures = sets.New[SupportedFeature]()
	}

	for feature := range options.ExemptFeatures {
		options.SupportedFeatures.Delete(feature)
	}

	suite := &ConformanceTestSuite{
		Client:           options.Client,
		Clientset:        options.Clientset,
		RestConfig:       options.RestConfig,
		RoundTripper:     roundTripper,
		GatewayClassName: options.GatewayClassName,
		Debug:            options.Debug,
		Cleanup:          options.CleanupBaseResources,
		BaseManifests:    options.BaseManifests,
		MeshManifests:    options.MeshManifests,
		Applier: kubernetes.Applier{
			NamespaceLabels:      options.NamespaceLabels,
			NamespaceAnnotations: options.NamespaceAnnotations,
		},
		SupportedFeatures:           options.SupportedFeatures,
		TimeoutConfig:               options.TimeoutConfig,
		SkipTests:                   sets.New(options.SkipTests...),
		RunTest:                     options.RunTest,
		FS:                          *options.FS,
		UsableNetworkAddresses:      options.UsableNetworkAddresses,
		UnusableNetworkAddresses:    options.UnusableNetworkAddresses,
		results:                     make(map[string]testResult),
		extendedUnsupportedFeatures: make(map[ConformanceProfileName]sets.Set[SupportedFeature]),
		extendedSupportedFeatures:   make(map[ConformanceProfileName]sets.Set[SupportedFeature]),
		conformanceProfiles:         options.ConformanceProfiles,
		implementation:              options.Implementation,
		mode:                        mode,
		apiVersion:                  apiVersion,
		apiChannel:                  apiChannel,
	}

	for _, conformanceProfileName := range options.ConformanceProfiles.UnsortedList() {
		conformanceProfile, err := getConformanceProfileForName(conformanceProfileName)
		if err != nil {
			return nil, fmt.Errorf("failed to retrieve conformance profile: %w", err)
		}
		// the use of a conformance profile implicitly enables any features of
		// that profile which are supported at a Core level of support.
		for _, f := range conformanceProfile.CoreFeatures.UnsortedList() {
			if !options.SupportedFeatures.Has(f) {
				suite.SupportedFeatures.Insert(f)
			}
		}
		for _, f := range conformanceProfile.ExtendedFeatures.UnsortedList() {
			if options.SupportedFeatures.Has(f) {
				if suite.extendedSupportedFeatures[conformanceProfileName] == nil {
					suite.extendedSupportedFeatures[conformanceProfileName] = sets.New[SupportedFeature]()
				}
				suite.extendedSupportedFeatures[conformanceProfileName].Insert(f)
			} else {
				if suite.extendedUnsupportedFeatures[conformanceProfileName] == nil {
					suite.extendedUnsupportedFeatures[conformanceProfileName] = sets.New[SupportedFeature]()
				}
				suite.extendedUnsupportedFeatures[conformanceProfileName].Insert(f)
			}
			// Add Exempt Features into unsupported features list
			if options.ExemptFeatures.Has(f) {
				suite.extendedUnsupportedFeatures[conformanceProfileName].Insert(f)
			}
		}
	}

	// apply defaults
	if suite.BaseManifests == "" {
		suite.BaseManifests = "base/manifests.yaml"
	}
	if suite.MeshManifests == "" {
		suite.MeshManifests = "mesh/manifests.yaml"
	}

	return suite, nil
}

// -----------------------------------------------------------------------------
// Conformance Test Suite - Public Methods
// -----------------------------------------------------------------------------

// Setup ensures the base resources required for conformance tests are installed
// in the cluster. It also ensures that all relevant resources are ready.
func (suite *ConformanceTestSuite) Setup(t *testing.T) {
	suite.Applier.FS = suite.FS
	suite.Applier.UsableNetworkAddresses = suite.UsableNetworkAddresses
	suite.Applier.UnusableNetworkAddresses = suite.UnusableNetworkAddresses

	if suite.SupportedFeatures.Has(SupportGateway) {
		t.Logf("Test Setup: Ensuring GatewayClass has been accepted")
		suite.ControllerName = kubernetes.GWCMustHaveAcceptedConditionTrue(t, suite.Client, suite.TimeoutConfig, suite.GatewayClassName)

		suite.Applier.GatewayClass = suite.GatewayClassName
		suite.Applier.ControllerName = suite.ControllerName

		t.Logf("Test Setup: Applying base manifests")
		suite.Applier.MustApplyWithCleanup(t, suite.Client, suite.TimeoutConfig, suite.BaseManifests, suite.Cleanup)

		t.Logf("Test Setup: Applying programmatic resources")
		secret := kubernetes.MustCreateSelfSignedCertSecret(t, "gateway-conformance-web-backend", "certificate", []string{"*"})
		suite.Applier.MustApplyObjectsWithCleanup(t, suite.Client, suite.TimeoutConfig, []client.Object{secret}, suite.Cleanup)
		secret = kubernetes.MustCreateSelfSignedCertSecret(t, "gateway-conformance-infra", "tls-validity-checks-certificate", []string{"*", "*.org"})
		suite.Applier.MustApplyObjectsWithCleanup(t, suite.Client, suite.TimeoutConfig, []client.Object{secret}, suite.Cleanup)
		secret = kubernetes.MustCreateSelfSignedCertSecret(t, "gateway-conformance-infra", "tls-passthrough-checks-certificate", []string{"abc.example.com"})
		suite.Applier.MustApplyObjectsWithCleanup(t, suite.Client, suite.TimeoutConfig, []client.Object{secret}, suite.Cleanup)
		secret = kubernetes.MustCreateSelfSignedCertSecret(t, "gateway-conformance-app-backend", "tls-passthrough-checks-certificate", []string{"abc.example.com"})
		suite.Applier.MustApplyObjectsWithCleanup(t, suite.Client, suite.TimeoutConfig, []client.Object{secret}, suite.Cleanup)

		t.Logf("Test Setup: Ensuring Gateways and Pods from base manifests are ready")
		namespaces := []string{
			"gateway-conformance-infra",
			"gateway-conformance-app-backend",
			"gateway-conformance-web-backend",
		}
		kubernetes.NamespacesMustBeReady(t, suite.Client, suite.TimeoutConfig, namespaces)
	}
	if suite.SupportedFeatures.Has(SupportMesh) {
		t.Logf("Test Setup: Applying base manifests")
		suite.Applier.MustApplyWithCleanup(t, suite.Client, suite.TimeoutConfig, suite.MeshManifests, suite.Cleanup)
		t.Logf("Test Setup: Ensuring Gateways and Pods from mesh manifests are ready")
		namespaces := []string{
			"gateway-conformance-mesh",
			"gateway-conformance-mesh-consumer",
			"gateway-conformance-app-backend",
			"gateway-conformance-web-backend",
		}
		kubernetes.MeshNamespacesMustBeReady(t, suite.Client, suite.TimeoutConfig, namespaces)
	}
}

// Run runs the provided set of conformance tests.
func (suite *ConformanceTestSuite) Run(t *testing.T, tests []ConformanceTest) error {
	// verify that the test suite isn't already running, don't start a new run
	// until the previous run finishes
	suite.lock.Lock()
	if suite.running {
		suite.lock.Unlock()
		return fmt.Errorf("can't run the test suite multiple times in parallel: the test suite is already running")
	}

	// if the test suite is not currently running, reset reporting and start a
	// new test run.
	suite.running = true
	suite.results = nil
	suite.lock.Unlock()

	// run all tests and collect the test results for conformance reporting
	results := make(map[string]testResult)
	for _, test := range tests {
		succeeded := t.Run(test.ShortName, func(t *testing.T) {
			test.Run(t, suite)
		})
		res := testSucceeded
		if suite.SkipTests.Has(test.ShortName) {
			res = testSkipped
		}
		if !suite.SupportedFeatures.HasAll(test.Features...) {
			res = testNotSupported
		}

		if !succeeded {
			res = testFailed
		}

		results[test.ShortName] = testResult{
			test:   test,
			result: res,
		}
	}

	// now that the tests have completed, mark the test suite as not running
	// and report the test results.
	suite.lock.Lock()
	suite.running = false
	suite.results = results
	suite.lock.Unlock()

	return nil
}

// Report emits a ConformanceReport for the previously completed test run.
// If no run completed prior to running the report, and error is emitted.
func (suite *ConformanceTestSuite) Report() (*confv1.ConformanceReport, error) {
	suite.lock.RLock()
	if suite.running {
		suite.lock.RUnlock()
		return nil, fmt.Errorf("can't generate report: the test suite is currently running")
	}
	defer suite.lock.RUnlock()

	testNames := make([]string, 0, len(suite.results))
	for tN := range suite.results {
		testNames = append(testNames, tN)
	}
	sort.Strings(testNames)
	profileReports := newReports()
	for _, tN := range testNames {
		testResult := suite.results[tN]
		conformanceProfiles := getConformanceProfilesForTest(testResult.test, suite.conformanceProfiles).UnsortedList()
		sort.Slice(conformanceProfiles, func(i, j int) bool {
			return conformanceProfiles[i].Name < conformanceProfiles[j].Name
		})
		for _, profile := range conformanceProfiles {
			profileReports.addTestResults(*profile, testResult)
		}
	}

	profileReports.compileResults(suite.extendedSupportedFeatures, suite.extendedUnsupportedFeatures)

	return &confv1.ConformanceReport{
		TypeMeta: v1.TypeMeta{
			APIVersion: "gateway.networking.k8s.io/v1alpha1",
			Kind:       "ConformanceReport",
		},
		Date:              time.Now().Format(time.RFC3339),
		Mode:              suite.mode,
		Implementation:    suite.implementation,
		GatewayAPIVersion: suite.apiVersion,
		GatewayAPIChannel: suite.apiChannel,
		ProfileReports:    profileReports.list(),
	}, nil
}

// ParseImplementation parses implementation-specific flag arguments and
// creates a *confv1a1.Implementation.
func ParseImplementation(org, project, url, version, contact string) (*confv1.Implementation, error) {
	if org == "" {
		return nil, errors.New("implementation's organization can not be empty")
	}
	if project == "" {
		return nil, errors.New("implementation's project can not be empty")
	}
	if url == "" {
		return nil, errors.New("implementation's url can not be empty")
	}
	if version == "" {
		return nil, errors.New("implementation's version can not be empty")
	}
	contacts := strings.Split(contact, ",")
	if len(contacts) == 0 {
		return nil, errors.New("implementation's contact can not be empty")
	}

	// TODO: add data validation https://github.com/kubernetes-sigs/gateway-api/issues/2178

	return &confv1.Implementation{
		Organization: org,
		Project:      project,
		URL:          url,
		Version:      version,
		Contact:      contacts,
	}, nil
}

// ParseConformanceProfiles parses flag arguments and converts the string to
// sets.Set[ConformanceProfileName].
func ParseConformanceProfiles(p string) sets.Set[ConformanceProfileName] {
	res := sets.Set[ConformanceProfileName]{}
	if p == "" {
		return res
	}

	for _, value := range strings.Split(p, ",") {
		res.Insert(ConformanceProfileName(value))
	}
	return res
}

// getAPIVersionAndChannel iterates over all the crds installed in the cluster and check the version and channel annotations.
// In case the annotations are not found or there are crds with different versions or channels, an error is returned.
func getAPIVersionAndChannel(crds []apiextensionsv1.CustomResourceDefinition) (version string, channel string, err error) {
	for _, crd := range crds {
		v, okv := crd.Annotations[consts.BundleVersionAnnotation]
		c, okc := crd.Annotations[consts.ChannelAnnotation]
		if !okv && !okc {
			continue
		}
		if !okv || !okc {
			return "", "", errors.New("detected CRDs with partial version and channel annotations")
		}
		if version != "" && v != version {
			return "", "", errors.New("multiple gateway API CRDs versions detected")
		}
		if channel != "" && c != channel {
			return "", "", errors.New("multiple gateway API CRDs channels detected")
		}
		version = v
		channel = c
	}
	if version == "" || channel == "" {
		return "", "", errors.New("no Gateway API CRDs with the proper annotations found in the cluster")
	}
	if version != consts.BundleVersion {
		return "", "", errors.New("the installed CRDs version is different from the suite version")
	}

	return version, channel, nil
}
