/*
Copyright 2024 The Kubernetes Authors.

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

package utils

import (
	"sigs.k8s.io/gateway-api/gwctl/pkg/extension/notfoundrefvalidator"
	"sigs.k8s.io/gateway-api/gwctl/pkg/extension/refgrantvalidator"
	"sigs.k8s.io/gateway-api/gwctl/pkg/topology"
)

func AggregateAnalysisErrors(node *topology.Node) ([]error, error) {
	var analysisErrors []error
	refGrantValidationMetadata, err := refgrantvalidator.Access(node)
	if err != nil {
		return nil, err
	}
	if refGrantValidationMetadata != nil && len(refGrantValidationMetadata.Errors) != 0 {
		analysisErrors = append(analysisErrors, refGrantValidationMetadata.Errors...)
	}
	notFoundRefValidatorMetadata, err := notfoundrefvalidator.Access(node)
	if err != nil {
		return nil, err
	}
	if notFoundRefValidatorMetadata != nil && len(notFoundRefValidatorMetadata.Errors) != 0 {
		analysisErrors = append(analysisErrors, notFoundRefValidatorMetadata.Errors...)
	}
	return analysisErrors, nil
}
