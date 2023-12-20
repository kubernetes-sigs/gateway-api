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

// V1alpha1 includes alpha maturity API types and utilities for creating and
// handling GEP metadata YAML files. These types are _only_
// intended for use by tools to read and manipulate the set of GEP metadata YAML
// files that are written in Golang.
//
// Please note that everything here is considered experimental and subject to
// change. Expect breaking changes and/or complete removals if you start using
// them.

// +groupName=internal.gateway.networking.k8s.io
package v1alpha1
