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

package handlers

import (
	"context"
	"net"
	"strings"

	extProcPb "github.com/envoyproxy/go-control-plane/envoy/service/ext_proc/v3"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"sigs.k8s.io/controller-runtime/pkg/log"

	envoy "sigs.k8s.io/gateway-api/conformance/epp-basic/pkg/common/envoy"
	"sigs.k8s.io/gateway-api/conformance/epp-basic/pkg/datastore"

	"sigs.k8s.io/gateway-api/conformance/epp-basic/pkg/metadata"
)

func (s *StreamingServer) handleRequestHeaders(ctx context.Context, reqCtx *RequestContext, fullReq *extProcPb.ProcessingRequest,
	req *extProcPb.ProcessingRequest_RequestHeaders) error {
	logger := log.FromContext(ctx)

	var metadataEndpoints []string

	// Read endpoint subset from filter metadata per the EPP protocol.
	// The data plane sets envoy.lb.subset_hint / x-gateway-destination-endpoint-subset
	// to constrain which endpoints the EPP may pick from.
	requestMetadata := envoy.ExtractMetadataValues(fullReq)
	if subsetMap, ok := requestMetadata[metadata.SubsetFilterNamespace].(map[string]any); ok {
		if endpointList, ok := subsetMap[metadata.SubsetFilterKey].([]any); ok {
			for _, ep := range endpointList {
				if epStr, ok := ep.(string); ok {
					metadataEndpoints = append(metadataEndpoints, epStr)
				}
			}
		}
	}

	// Test path: read endpoint selection from the test-only request header.
	// This is used by conformance tests to steer routing to a specific pod, analogous
	// to the HeaderBasedTestingFilter plugin in the main EPP. The header always takes
	// priority over metadata, matching the full EPP where HeaderBasedTestingFilter
	// runs as a separate filter stage after metadata-based candidate selection.
	var filterEndpoints []string
	for _, header := range req.RequestHeaders.Headers.Headers {
		if header.Key == "test-epp-endpoint-selection" {
			val := envoy.GetHeaderValue(header)
			if val != "" {
				filterEndpoints = strings.Split(val, ",")
				logger.Info("Found test endpoint selection header", "value", val)
			}
			break
		}
	}
	if len(filterEndpoints) == 0 && len(metadataEndpoints) > 0 {
		filterEndpoints = metadataEndpoints
	}

	allPods := s.datastore.PodList(datastore.AllPodsPredicate)
	if len(allPods) == 0 {
		return status.Errorf(codes.Unavailable, "no pods available")
	}

	var candidates []*datastore.Endpoint
	if len(filterEndpoints) > 0 {
		// Build a set of IP addresses from the filter list. Filter entries may be
		// "ip" or "ip:port"; we match only on the IP portion.
		allowedIPs := make(map[string]struct{}, len(filterEndpoints))
		for _, ep := range filterEndpoints {
			ep = strings.TrimSpace(ep)
			if host, _, err := net.SplitHostPort(ep); err == nil {
				allowedIPs[host] = struct{}{}
			} else {
				allowedIPs[ep] = struct{}{}
			}
		}
		for _, pod := range allPods {
			if _, ok := allowedIPs[pod.Address]; ok {
				candidates = append(candidates, pod)
			}
		}
	}

	// If no matches or header not present, use all pods.
	if len(candidates) == 0 {
		candidates = allPods
	}

	// For the lightweight implementation (Round Robin), no request context is required.
	res, err := s.picker.Pick(ctx, &PickRequest{}, candidates)
	if err != nil {
		return err
	}

	reqCtx.TargetEndpoint = res.Endpoint
	if host, _, err := net.SplitHostPort(res.Endpoint); err == nil {
		reqCtx.SelectedPodIP = host
	} else {
		reqCtx.SelectedPodIP = res.Endpoint
	}

	logger.V(4).Info("Selected endpoint", "podIP", reqCtx.SelectedPodIP, "endpoint", reqCtx.TargetEndpoint)

	return nil
}
