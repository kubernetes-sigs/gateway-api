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

package main

import (
	"fmt"
	"sort"
	"strings"

	"sigs.k8s.io/gateway-api/conformance/utils/suite"
)

func main() {
	features := make([]string, suite.AllFeatures.Len())
	for i, feat := range suite.AllFeatures.UnsortedList() {
		features[i] = string(feat)
	}
	sort.Strings(features)
	fmt.Println(strings.Join(features, ";"))
}
