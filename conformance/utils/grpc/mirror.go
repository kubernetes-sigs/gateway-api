package grpc

import (
	"fmt"
	"github.com/stretchr/testify/require"
	clientset "k8s.io/client-go/kubernetes"
	"regexp"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/gateway-api/conformance/utils/kubernetes"
	"sigs.k8s.io/gateway-api/conformance/utils/tlog"
	"sync"
	"testing"
	"time"
)

func ExpectMirroredRequest(t *testing.T, client client.Client, clientset clientset.Interface, mirrorPods []MirroredBackend) {
	for i, mirrorPod := range mirrorPods {
		if mirrorPod.Name == "" {
			tlog.Fatalf(t, "Mirrored BackendRef[%d].Name wasn't provided in the testcase, this test should only check http request mirror.", i)
		}
	}

	var wg sync.WaitGroup
	wg.Add(len(mirrorPods))

	assertionStart := time.Now()

	for _, mirrorPod := range mirrorPods {
		go func(mirrorPod MirroredBackend) {
			defer wg.Done()

			require.Eventually(t, func() bool {
				mirrorLogRegexp := regexp.MustCompile(fmt.Sprintf("Received over plaintext"))

				tlog.Log(t, "Searching for the mirrored request log")
				tlog.Logf(t, `Reading "%s/%s" logs`, mirrorPod.Namespace, mirrorPod.Name)
				logs, err := kubernetes.DumpEchoLogs(mirrorPod.Namespace, mirrorPod.Name, client, clientset, assertionStart)
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
			}, 60*time.Second, time.Millisecond*100, `Couldn't find mirrored request in "%s/%s" logs`, mirrorPod.Namespace, mirrorPod.Name)
		}(mirrorPod)
	}

	wg.Wait()

	tlog.Log(t, "Found mirrored request log in all desired backends")
}
