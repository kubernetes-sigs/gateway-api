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

// Package resourcediscovery discovers and maps relationships between Gateway
// API resources. It constructs a graph-like model representing resource
// dependencies and their interactions.
//
// Key features:
//
// # Resource discovery:
//   - Fetches GatewayClasses, Gateways, HTTPRoutes, Backends, and Policies
//     based on filters.
//   - Discovers related resources by following references between them.
//
// # Resource model construction:
//   - Builds a graph-like model representing resources and their connections.
//   - Tracks relationships between GatewayClasses, Gateways, HTTPRoutes,
//     Backends, Namespaces, and Policies.
//
// # Policy evaluation:
//   - Identifies effective policies applicable to each resource, considering
//     inheritance and hierarchy.
package resourcediscovery
