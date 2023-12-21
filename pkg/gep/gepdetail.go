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

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type GEPStatus string

const (

	// Common GEP statuses
	GEPStatusDeclined GEPStatus = "Declined"
	GEPStatusDeferred GEPStatus = "Deferred"

	// Memorandum Status
	GEPStatusMemorandum GEPStatus = "Memorandum"

	// Standard GEP statuses
	GEPStatusProvisional   GEPStatus = "Provisional"
	GEPStatusPrototyping   GEPStatus = "Prototyping"
	GEPStatusImplementable GEPStatus = "Implementable"
	GEPStatusExperimental  GEPStatus = "Experimental"
	GEPStatusStandard      GEPStatus = "Standard"
	GEPStatusCompleted     GEPStatus = "Completed"
)

// GEPDetail holds the metadata used to describe a Gateway API GEP (Gateway
// Enhancement Proposal)
type GEPDetail struct {
	metav1.TypeMeta `json:",inline"`

	// The GEP's number, as per the issue number representing the GEP.
	Number uint `json:"number"`

	// The GEP's name, usually the name of the issue without the "GEP:" prefix.
	Name string `json:"name"`

	// The GEP's status.
	//
	// Valid values are provided in the constants for the GEPStatus type.
	Status GEPStatus `json:"status"`

	// The GEP's authors, listed as their Github handles.
	Authors []string `json:"authors"`

	// Relationships describes the possible relationships between this GEP and
	// other GEPs.
	Relationships GEPRelationships `json:"relationships,omitempty"`

	// References provides a list of hyperlinks to other references used by the GEP.
	References []string `json:"references,omitempty"`

	// FeatureNames provides a list of feature names (used in conformance tests
	// and GatewayClass supported features lists)
	// TODO(youngnick): Move the canonical feature names list from
	// `conformance/utils/features.go` to its own package in `pkg`, and
	// then move this to SupportedFeatures type instead.
	FeatureNames []string

	// Changelog provides a list of hyperlinks to PRs that affected this GEP.
	Changelog []string
}

// GEPRelationships describes the possible relationships GEPs may have.
type GEPRelationships struct {
	// The GEP Obsoletes the listed GEPs.
	Obsoletes []GEPRelationship `json:"obsoletes,omitempty"`
	// The GEP is Obsoleted by the listed GEPs.
	ObsoletedBy []GEPRelationship `json:"obsoletedBy,omitempty"`
	// The GEP Updates the listed GEPs.
	Extends []GEPRelationship `json:"updates,omitempty"`
	// The GEP is Updated by the listed GEPs.
	ExtendedBy []GEPRelationship `json:"updatedBy,omitempty"`
	// The listed GEPs are relevant for some other reason.
	SeeAlso []GEPRelationship `json:"seeAlso,omitempty"`
}

type GEPRelationship struct {
	Number      uint   `json:"number"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
}
