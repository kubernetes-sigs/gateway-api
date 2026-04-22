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

package config

import (
	"strconv"
	"strings"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type TimeoutConfig struct {
	// TestIsolation represents the time block between test cases to enhance test isolation.
	// Max value for conformant implementation: None
	TestIsolation metav1.Duration `json:"testIsolation"`

	// CreateTimeout represents the maximum time for a Kubernetes object to be created.
	// Max value for conformant implementation: None
	CreateTimeout metav1.Duration `json:"createTimeout"`

	// DeleteTimeout represents the maximum time for a Kubernetes object to be deleted.
	// Max value for conformant implementation: None
	DeleteTimeout metav1.Duration `json:"deleteTimeout"`

	// GetTimeout represents the maximum time to get a Kubernetes object.
	// Max value for conformant implementation: None
	GetTimeout metav1.Duration `json:"getTimeout"`

	// GatewayMustHaveAddress represents the maximum time for at least one IP Address has been set in the status of a Gateway.
	// Max value for conformant implementation: None
	GatewayMustHaveAddress metav1.Duration `json:"gatewayMustHaveAddress"`

	// GatewayMustHaveCondition represents the maximum amount of time for a
	// Gateway to have the supplied Condition.
	// Max value for conformant implementation: None
	GatewayMustHaveCondition metav1.Duration `json:"gatewayMustHaveCondition"`

	// GatewayStatusMustHaveListeners represents the maximum time for a Gateway to have listeners in status that match the expected listeners.
	// Max value for conformant implementation: None
	GatewayStatusMustHaveListeners metav1.Duration `json:"gatewayStatusMustHaveListeners"`

	// GatewayListenersMustHaveConditions represents the maximum time for a Gateway to have all listeners with a specific condition.
	// Max value for conformant implementation: None
	GatewayListenersMustHaveConditions metav1.Duration `json:"gatewayListenersMustHaveConditions"`

	// ListenerSetMustHaveCondition represents the maximum amount of time for a
	// ListenerSet to have the supplied Condition.
	// Max value for conformant implementation: None
	ListenerSetMustHaveCondition metav1.Duration `json:"listenerSetMustHaveCondition"`

	// ListenerSetListenersMustHaveConditions represents the maximum time for a ListenerSet to have all listeners with a specific condition.
	// Max value for conformant implementation: None
	ListenerSetListenersMustHaveConditions metav1.Duration `json:"listenerSetListenersMustHaveConditions"`

	// GWCMustBeAccepted represents the maximum time for a GatewayClass to have an Accepted condition set to true.
	// Max value for conformant implementation: None
	GWCMustBeAccepted metav1.Duration `json:"gwcMustBeAccepted"`

	// HTTPRouteMustNotHaveParents represents the maximum time for an HTTPRoute to have either no parents or a single parent that is not accepted.
	// Max value for conformant implementation: None
	HTTPRouteMustNotHaveParents metav1.Duration `json:"httpRouteMustNotHaveParents"`

	// HTTPRouteMustHaveCondition represents the maximum time for an HTTPRoute to have the supplied Condition.
	// Max value for conformant implementation: None
	HTTPRouteMustHaveCondition metav1.Duration `json:"httpRouteMustHaveCondition"`

	// TLSRouteMustHaveCondition represents the maximum time for a TLSRoute to have the supplied Condition.
	// Max value for conformant implementation: None
	TLSRouteMustHaveCondition metav1.Duration `json:"tlsRouteMustHaveCondition"`

	// RouteMustHaveParents represents the maximum time for an xRoute to have parents in status that match the expected parents.
	// Max value for conformant implementation: None
	RouteMustHaveParents metav1.Duration `json:"routeMustHaveParents"`

	// ManifestFetchTimeout represents the maximum time for getting content from a https:// URL.
	// Max value for conformant implementation: None
	ManifestFetchTimeout metav1.Duration `json:"manifestFetchTimeout"`

	// MaxTimeToConsistency is the maximum time for requiredConsecutiveSuccesses (default 3) requests to succeed in a row before failing the test.
	// Max value for conformant implementation: 30 seconds
	MaxTimeToConsistency metav1.Duration `json:"maxTimeToConsistency"`

	// NamespacesMustBeReady represents the maximum time for the following to happen within
	// specified namespace(s):
	// * All Pods to be marked as "Ready"
	// * All Gateways to be marked as "Accepted" and "Programmed"
	// Max value for conformant implementation: None
	NamespacesMustBeReady metav1.Duration `json:"namespacesMustBeReady"`

	// RequestTimeout represents the maximum time for making an HTTP Request with the roundtripper.
	// Max value for conformant implementation: None
	RequestTimeout metav1.Duration `json:"requestTimeout"`

	// LatestObservedGenerationSet represents the maximum time for an ObservedGeneration to bump.
	// Max value for conformant implementation: None
	LatestObservedGenerationSet metav1.Duration `json:"latestObservedGenerationSet"`

	// DefaultTestTimeout is the default amount of time to wait for a test to complete
	DefaultTestTimeout metav1.Duration `json:"defaultTestTimeout"`

	// DefaultPollInterval is the default amount of time to poll for status checks.
	DefaultPollInterval metav1.Duration `json:"defaultPollInterval"`

	// RequiredConsecutiveSuccesses is the number of requests that must succeed in a row
	// to consider a response "consistent" before making additional assertions on the response body.
	// If this number is not reached within MaxTimeToConsistency, the test will fail.
	RequiredConsecutiveSuccesses int `json:"requiredConsecutiveSuccesses"`
}

// DefaultTimeoutConfig populates a TimeoutConfig with the default values.
func DefaultTimeoutConfig() TimeoutConfig {
	return TimeoutConfig{
		CreateTimeout:                          metav1.Duration{Duration: 60 * time.Second},
		DeleteTimeout:                          metav1.Duration{Duration: 10 * time.Second},
		GetTimeout:                             metav1.Duration{Duration: 10 * time.Second},
		GatewayMustHaveAddress:                 metav1.Duration{Duration: 180 * time.Second},
		GatewayMustHaveCondition:               metav1.Duration{Duration: 180 * time.Second},
		GatewayStatusMustHaveListeners:         metav1.Duration{Duration: 60 * time.Second},
		GatewayListenersMustHaveConditions:     metav1.Duration{Duration: 60 * time.Second},
		ListenerSetMustHaveCondition:           metav1.Duration{Duration: 180 * time.Second},
		ListenerSetListenersMustHaveConditions: metav1.Duration{Duration: 60 * time.Second},
		GWCMustBeAccepted:                      metav1.Duration{Duration: 180 * time.Second},
		HTTPRouteMustNotHaveParents:            metav1.Duration{Duration: 60 * time.Second},
		HTTPRouteMustHaveCondition:             metav1.Duration{Duration: 60 * time.Second},
		TLSRouteMustHaveCondition:              metav1.Duration{Duration: 60 * time.Second},
		RouteMustHaveParents:                   metav1.Duration{Duration: 60 * time.Second},
		ManifestFetchTimeout:                   metav1.Duration{Duration: 10 * time.Second},
		MaxTimeToConsistency:                   metav1.Duration{Duration: 30 * time.Second},
		NamespacesMustBeReady:                  metav1.Duration{Duration: 300 * time.Second},
		RequestTimeout:                         metav1.Duration{Duration: 10 * time.Second},
		LatestObservedGenerationSet:            metav1.Duration{Duration: 60 * time.Second},
		DefaultTestTimeout:                     metav1.Duration{Duration: 60 * time.Second},
		DefaultPollInterval:                    metav1.Duration{Duration: 100 * time.Millisecond},
		RequiredConsecutiveSuccesses:           3,
	}
}

// ParseTimeoutOverrides parses the timeout overrides string and updates the TimeoutConfig.
func ParseTimeoutOverrides(timeoutConfig *TimeoutConfig, overrides string) {
	if overrides == "" {
		return
	}
	pairs := strings.Split(overrides, ";")
	for _, pair := range pairs {
		parts := strings.Split(pair, ":")
		if len(parts) != 2 {
			continue
		}
		param := strings.TrimSpace(parts[0])
		val := strings.TrimSpace(parts[1])
		valInt, err := strconv.Atoi(val)
		if err != nil {
			continue
		}
		overrideDuration := metav1.Duration{Duration: time.Duration(valInt) * time.Second}

		switch param {
		case "CreateTimeout":
			timeoutConfig.CreateTimeout = overrideDuration
		case "DeleteTimeout":
			timeoutConfig.DeleteTimeout = overrideDuration
		case "GetTimeout":
			timeoutConfig.GetTimeout = overrideDuration
		case "GatewayMustHaveAddress":
			timeoutConfig.GatewayMustHaveAddress = overrideDuration
		case "GatewayMustHaveCondition":
			timeoutConfig.GatewayMustHaveCondition = overrideDuration
		case "GatewayStatusMustHaveListeners":
			timeoutConfig.GatewayStatusMustHaveListeners = overrideDuration
		case "GatewayListenersMustHaveConditions":
			timeoutConfig.GatewayListenersMustHaveConditions = overrideDuration
		case "ListenerSetMustHaveCondition":
			timeoutConfig.ListenerSetMustHaveCondition = overrideDuration
		case "ListenerSetListenersMustHaveConditions":
			timeoutConfig.ListenerSetListenersMustHaveConditions = overrideDuration
		case "GWCMustBeAccepted":
			timeoutConfig.GWCMustBeAccepted = overrideDuration
		case "HTTPRouteMustNotHaveParents":
			timeoutConfig.HTTPRouteMustNotHaveParents = overrideDuration
		case "HTTPRouteMustHaveCondition":
			timeoutConfig.HTTPRouteMustHaveCondition = overrideDuration
		case "TLSRouteMustHaveCondition":
			timeoutConfig.TLSRouteMustHaveCondition = overrideDuration
		case "RouteMustHaveParents":
			timeoutConfig.RouteMustHaveParents = overrideDuration
		case "ManifestFetchTimeout":
			timeoutConfig.ManifestFetchTimeout = overrideDuration
		case "MaxTimeToConsistency":
			timeoutConfig.MaxTimeToConsistency = overrideDuration
		case "NamespacesMustBeReady":
			timeoutConfig.NamespacesMustBeReady = overrideDuration
		case "RequestTimeout":
			timeoutConfig.RequestTimeout = overrideDuration
		case "LatestObservedGenerationSet":
			timeoutConfig.LatestObservedGenerationSet = overrideDuration
		case "DefaultTestTimeout":
			timeoutConfig.DefaultTestTimeout = overrideDuration
		case "DefaultPollInterval":
			timeoutConfig.DefaultPollInterval = overrideDuration
		case "RequiredConsecutiveSuccesses":
			timeoutConfig.RequiredConsecutiveSuccesses = valInt
		case "TestIsolation":
			timeoutConfig.TestIsolation = overrideDuration
		}
	}
}

func SetupTimeoutConfig(timeoutConfig *TimeoutConfig) {
	defaultTimeoutConfig := DefaultTimeoutConfig()
	if timeoutConfig.CreateTimeout.Duration == 0 {
		timeoutConfig.CreateTimeout = defaultTimeoutConfig.CreateTimeout
	}
	if timeoutConfig.DeleteTimeout.Duration == 0 {
		timeoutConfig.DeleteTimeout = defaultTimeoutConfig.DeleteTimeout
	}
	if timeoutConfig.GetTimeout.Duration == 0 {
		timeoutConfig.GetTimeout = defaultTimeoutConfig.GetTimeout
	}
	if timeoutConfig.GatewayMustHaveAddress.Duration == 0 {
		timeoutConfig.GatewayMustHaveAddress = defaultTimeoutConfig.GatewayMustHaveAddress
	}
	if timeoutConfig.GatewayMustHaveCondition.Duration == 0 {
		timeoutConfig.GatewayMustHaveCondition = defaultTimeoutConfig.GatewayMustHaveCondition
	}
	if timeoutConfig.GatewayStatusMustHaveListeners.Duration == 0 {
		timeoutConfig.GatewayStatusMustHaveListeners = defaultTimeoutConfig.GatewayStatusMustHaveListeners
	}
	if timeoutConfig.GatewayListenersMustHaveConditions.Duration == 0 {
		timeoutConfig.GatewayListenersMustHaveConditions = defaultTimeoutConfig.GatewayListenersMustHaveConditions
	}
	if timeoutConfig.GWCMustBeAccepted.Duration == 0 {
		timeoutConfig.GWCMustBeAccepted = defaultTimeoutConfig.GWCMustBeAccepted
	}
	if timeoutConfig.HTTPRouteMustNotHaveParents.Duration == 0 {
		timeoutConfig.HTTPRouteMustNotHaveParents = defaultTimeoutConfig.HTTPRouteMustNotHaveParents
	}
	if timeoutConfig.HTTPRouteMustHaveCondition.Duration == 0 {
		timeoutConfig.HTTPRouteMustHaveCondition = defaultTimeoutConfig.HTTPRouteMustHaveCondition
	}
	if timeoutConfig.RouteMustHaveParents.Duration == 0 {
		timeoutConfig.RouteMustHaveParents = defaultTimeoutConfig.RouteMustHaveParents
	}
	if timeoutConfig.ManifestFetchTimeout.Duration == 0 {
		timeoutConfig.ManifestFetchTimeout = defaultTimeoutConfig.ManifestFetchTimeout
	}
	if timeoutConfig.MaxTimeToConsistency.Duration == 0 {
		timeoutConfig.MaxTimeToConsistency = defaultTimeoutConfig.MaxTimeToConsistency
	}
	if timeoutConfig.NamespacesMustBeReady.Duration == 0 {
		timeoutConfig.NamespacesMustBeReady = defaultTimeoutConfig.NamespacesMustBeReady
	}
	if timeoutConfig.RequestTimeout.Duration == 0 {
		timeoutConfig.RequestTimeout = defaultTimeoutConfig.RequestTimeout
	}
	if timeoutConfig.LatestObservedGenerationSet.Duration == 0 {
		timeoutConfig.LatestObservedGenerationSet = defaultTimeoutConfig.LatestObservedGenerationSet
	}
	if timeoutConfig.TLSRouteMustHaveCondition.Duration == 0 {
		timeoutConfig.TLSRouteMustHaveCondition = defaultTimeoutConfig.TLSRouteMustHaveCondition
	}
	if timeoutConfig.DefaultTestTimeout.Duration == 0 {
		timeoutConfig.DefaultTestTimeout = defaultTimeoutConfig.DefaultTestTimeout
	}
	if timeoutConfig.DefaultPollInterval.Duration == 0 {
		timeoutConfig.DefaultPollInterval = defaultTimeoutConfig.DefaultPollInterval
	}
}
