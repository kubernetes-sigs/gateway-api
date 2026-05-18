/*
Copyright The Kubernetes Authors.

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

package udp

import (
	"context"
	"fmt"
	"net"
	"testing"
	"time"

	"k8s.io/apimachinery/pkg/util/wait"

	"sigs.k8s.io/gateway-api/conformance/utils/tlog"
)

// ExpectEchoResponse polls until a UDP echo round-trip against the given
// gateway address succeeds, or the timeout is exceeded.
func ExpectEchoResponse(t *testing.T, timeout time.Duration, gwAddr string) {
	t.Helper()

	const probe = "gateway-api-conformance-udp-echo"
	tlog.Logf(t, "performing UDP echo probe on %s", gwAddr)
	err := wait.PollUntilContextTimeout(context.TODO(), time.Second, timeout, true,
		func(ctx context.Context) (bool, error) {
			var dialer net.Dialer
			conn, err := dialer.DialContext(ctx, "udp", gwAddr)
			if err != nil {
				tlog.Logf(t, "failed to dial UDP %s: %v", gwAddr, err)
				return false, nil
			}
			defer conn.Close()

			if err = conn.SetDeadline(time.Now().Add(2 * time.Second)); err != nil {
				return false, fmt.Errorf("setting UDP deadline: %w", err)
			}
			if _, err = conn.Write([]byte(probe)); err != nil {
				tlog.Logf(t, "failed to write UDP probe: %v", err)
				return false, nil
			}

			buf := make([]byte, 1024)
			n, err := conn.Read(buf)
			if err != nil {
				tlog.Logf(t, "failed to read UDP echo response: %v", err)
				return false, nil
			}
			tlog.Logf(t, "got UDP echo response (%d bytes) from %s", n, gwAddr)
			return true, nil
		})
	if err != nil {
		t.Errorf("UDP echo probe never succeeded against %s: %v", gwAddr, err)
	}
}
