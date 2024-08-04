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

package policymanager

import "sigs.k8s.io/gateway-api/gwctl/pkg/common"

// ToPolicyRefs returns the Object references of all given policies. Note that
// these are not the value of targetRef within the Policies but rather the
// reference to the Policy object itself.
func ToPolicyRefs(policies []Policy) []common.GKNN {
	var result []common.GKNN
	for _, policy := range policies {
		result = append(result, policy.GKNN())
	}
	return result
}
