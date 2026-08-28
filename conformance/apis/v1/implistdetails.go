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

package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ImplementationListDetail holds the details required to add an implementation
// to the implementations list page.
type ImplementationListDetail struct {
	metav1.TypeMeta `json:",inline"`
	// The details in the fields in the Implementation struct _must_ match
	// the details in the conformance reports.
	Implementation `json:",inline"`

	// FullName is the text name of the project, used for the section name in the
	// implementation list.
	//
	// +required
	FullName string `json:"fullName"`

	// Description is the Markdown description of the implementation.
	//
	// +required
	Description string `json:"description"`
}
