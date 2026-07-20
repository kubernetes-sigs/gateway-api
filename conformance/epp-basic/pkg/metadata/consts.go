/*
Copyright 2025 The Kubernetes Authors.

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

package metadata

const (
	// SubsetFilterNamespace is the key for the outer namespace struct in the metadata field of the extproc request that is used to wrap the subset filter.
	SubsetFilterNamespace = "envoy.lb.subset_hint"
	// SubsetFilterKey is the metadata key used by Envoy to specify an array candidate pods for serving the request.
	// If not specified, all the pods that are associated with the pool are candidates.
	SubsetFilterKey = "x-gateway-destination-endpoint-subset"
	// DestinationEndpointNamespace is the key for the outer namespace struct in the metadata field of the extproc response that is used to wrap the target endpoint.
	DestinationEndpointNamespace = "envoy.lb"
	// DestinationEndpointKey is the header and response metadata key used by Envoy to route to the appropriate pod.
	DestinationEndpointKey = "x-gateway-destination-endpoint"
	// DestinationEndpointServedKey is the metadata key used by Envoy to specify the endpoint that served the request.
	DestinationEndpointServedKey = "x-gateway-destination-endpoint-served"
	// ConformanceTestResultHeader is the header used by the conformance test to specify the endpoint that served the request.
	ConformanceTestResultHeader = "x-conformance-test-served-endpoint"
	// FlowFairnessIDKey is the header key used to pass the fairness ID to be used in Flow Control.
	FlowFairnessIDKey = "x-gateway-inference-fairness-id"
	// ObjectiveKey is the header key used to specify the objective of an incoming request.
	ObjectiveKey = "x-gateway-inference-objective"
	// ModelNameRewriteKey is the header key used to specify the model name to be used when the request is forwarded to the model server.
	ModelNameRewriteKey = "x-gateway-model-name-rewrite"
)
