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

package grpc

import (
	"fmt"
	"regexp"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	clientset "k8s.io/client-go/kubernetes"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"sigs.k8s.io/gateway-api/conformance/utils/config"
	"sigs.k8s.io/gateway-api/conformance/utils/http"
	"sigs.k8s.io/gateway-api/conformance/utils/kubernetes"
	"sigs.k8s.io/gateway-api/conformance/utils/tlog"
)

// ExpectMirroredRequest verifies that gRPC requests for the given ExpectedResponse
// were mirrored to each backend in mirrorPods, by scanning pod logs for the echo
// server's log line emitted on every received request.
func ExpectMirroredRequest(t *testing.T, c client.Client, cs clientset.Interface, mirrorPods []http.MirroredBackend, expected ExpectedResponse, timeoutConfig config.TimeoutConfig) {
	t.Helper()
	for i, mirrorPod := range mirrorPods {
		if mirrorPod.Name == "" {
			tlog.Fatalf(t, "Mirrored BackendRef[%d].Name wasn't provided in the testcase, this test should only check gRPC request mirror.", i)
		}
	}

	var wg sync.WaitGroup
	wg.Add(len(mirrorPods))

	assertionStart := time.Now()
	method := getFullyQualifiedMethod(&expected)

	for _, mirrorPod := range mirrorPods {
		go func(mirrorPod http.MirroredBackend) {
			defer wg.Done()

			require.Eventually(t, func() bool {
				mirrorLogRegexp := regexp.MustCompile(fmt.Sprintf("Echoing back gRPC request made to \\%s to client", method))

				tlog.Log(t, "Searching for the mirrored gRPC request log")
				tlog.Logf(t, `Reading "%s/%s" logs`, mirrorPod.Namespace, mirrorPod.Name)
				logs, err := kubernetes.DumpEchoLogs(t.Context(), mirrorPod.Namespace, mirrorPod.Name, c, cs, assertionStart)
				if err != nil {
					tlog.Logf(t, `Couldn't read "%s/%s" logs: %v`, mirrorPod.Namespace, mirrorPod.Name, err)
					return false
				}

				for _, log := range logs {
					if mirrorLogRegexp.MatchString(log) {
						return true
					}
				}
				return false
			}, timeoutConfig.RequestTimeout, time.Second, `Couldn't find mirrored gRPC request in "%s/%s" logs`, mirrorPod.Namespace, mirrorPod.Name)
		}(mirrorPod)
	}

	wg.Wait()

	tlog.Log(t, "Found mirrored gRPC request log in all desired backends")
}
