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

package meshtests

import (
	"fmt"
	"testing"

	"google.golang.org/grpc/codes"

	pb "sigs.k8s.io/gateway-api/conformance/echo-basic/grpcechoserver"
	"sigs.k8s.io/gateway-api/conformance/utils/echo"
	"sigs.k8s.io/gateway-api/conformance/utils/grpc"
	"sigs.k8s.io/gateway-api/conformance/utils/suite"
	"sigs.k8s.io/gateway-api/conformance/utils/weight"
	"sigs.k8s.io/gateway-api/pkg/features"
)

func init() {
	MeshConformanceTests = append(MeshConformanceTests, MeshGRPCRouteWeight)
}

var MeshGRPCRouteWeight = suite.ConformanceTest{
	ShortName:   "MeshGRPCRouteWeight",
	Description: "A GRPCRoute with weighted backends in mesh mode",
	Manifests:   []string{"tests/mesh/grpcroute-weight.yaml"},
	Features: []features.FeatureName{
		features.SupportMesh,
		features.SupportGRPCRoute,
	},
	Test: func(t *testing.T, s *suite.ConformanceTestSuite) {
		// Connect to mesh app to get the service address
		client := echo.ConnectToApp(t, s, echo.MeshAppEchoV1)
		
		t.Run("Requests should have a distribution that matches the weight", func(t *testing.T) {
			expected := grpc.ExpectedResponse{
				EchoRequest: &pb.EchoRequest{},
				Response:    grpc.Response{Code: codes.OK},
				Namespace:   "gateway-conformance-mesh",
			}

			// Assert request succeeds before doing our distribution check
			grpcClient := &grpc.DefaultClient{}
			defer grpcClient.Close()
			resp, err := grpcClient.SendRPC(t, client.Address+":9000", expected, s.TimeoutConfig.MaxTimeToConsistency)
			if err != nil {
				t.Skipf("gRPC mesh test requires gRPC support on mesh services: %v", err)
			}
			if resp.Code != codes.OK {
				t.Skipf("gRPC mesh test requires working gRPC endpoints: got %v", resp.Code)
			}

			expectedWeights := map[string]float64{
				"echo-v1": 0.7,
				"echo-v2": 0.3,
			}

			sender := weight.NewFunctionBasedSender(func() (string, error) {
				uniqueExpected := expected
				if err := grpc.AddEntropy(&uniqueExpected); err != nil {
					return "", fmt.Errorf("error adding entropy: %w", err)
				}
				
				grpcClient := &grpc.DefaultClient{}
				defer grpcClient.Close()
				resp, err := grpcClient.SendRPC(t, client.Address+":9000", uniqueExpected, s.TimeoutConfig.MaxTimeToConsistency)
				if err != nil {
					return "", fmt.Errorf("failed to send gRPC mesh request: %w", err)
				}
				if resp.Code != codes.OK {
					return "", fmt.Errorf("expected OK response, got %v", resp.Code)
				}
				return resp.Response.GetAssertions().GetContext().GetPod(), nil
			})

			for i := 0; i < 10; i++ {
				if err := weight.TestWeightedDistribution(sender, expectedWeights); err != nil {
					t.Logf("Traffic distribution test failed (%d/10): %s", i+1, err)
				} else {
					return
				}
			}
			t.Fatal("Weighted distribution tests failed")
		})
	},
}