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

package l4server

import (
	"context"
	"fmt"

	tcpserver "sigs.k8s.io/gateway-api/conformance/echo-basic/tcpserver"
	"sigs.k8s.io/gateway-api/conformance/echo-basic/udpechoserver"
)

func Main() {
	ctx := context.Background()
	errchan := make(chan error, 2)
	tcpserver.Start(ctx, errchan)
	udpechoserver.Start(ctx, errchan)
	if err := <-errchan; err != nil {
		panic(fmt.Sprintf("Failed to start l4 server: %s\n", err.Error()))
	}
}
