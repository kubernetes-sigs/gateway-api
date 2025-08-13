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

package grpc

import (
	"testing"

	"sigs.k8s.io/gateway-api/conformance/utils/tlog"

	clientset "k8s.io/client-go/kubernetes"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func ExpectMirroredRequest(t *testing.T, client client.Client, clientset clientset.Interface, mirrorPods []MirroredBackend, methodName string) {
	tlog.Log(t, "TODO: Check for mirroring in grpc")
}
