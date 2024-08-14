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

package integration

import (
	_ "embed"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"k8s.io/cli-runtime/pkg/genericiooptions"

	cmdget "sigs.k8s.io/gateway-api/gwctl/cmd/get"
	"sigs.k8s.io/gateway-api/gwctl/pkg/common"
)

//go:embed testdata/sample1.yaml
var testdataSample1 string

func TestGet(t *testing.T) {
	factory := NewTestFactory(t, testdataSample1)

	testCases := []struct {
		name       string
		inputArgs  []string
		namespace  string // Controls the '-n' flag. Empty value means all-namespaces (-A)
		describe   bool
		wantOut    string
		wantErrOut string
	}{
		{
			name:      "get gateways -n test",
			inputArgs: []string{"gateways"},
			namespace: "test",
			wantOut: `
NAMESPACE  NAME       CLASS                           ADDRESSES  PORTS  PROGRAMMED  AGE
test       gateway-1  foo-com-external-gateway-class             80     Unknown     <unknown>
test       gateway-2  bar-com-internal-gateway-class             443    Unknown     <unknown>
`,
		},
		{
			name:      "get gateways",
			inputArgs: []string{"gateways"},
			namespace: "default",
			wantOut: `
NAMESPACE  NAME       CLASS                           ADDRESSES  PORTS  PROGRAMMED  AGE
default    gateway-3  foo-com-external-gateway-class             80     Unknown     <unknown>
`,
		},
		{
			name:      "get gatewayclasses",
			inputArgs: []string{"gatewayclasses"},
			wantOut: `
NAME                            CONTROLLER                      ACCEPTED  AGE
bar-com-internal-gateway-class  bar.baz/internal-gateway-class  Unknown   <unknown>
foo-com-external-gateway-class  foo.com/external-gateway-class  Unknown   <unknown>
`,
		},
		{
			name:      "get httproutes -A",
			inputArgs: []string{"httproutes"},
			namespace: "", // All namespaces
			wantOut: `
NAMESPACE  NAME         HOSTNAMES                          PARENT REFS  AGE
default    httproute-3  example4.com                       1            <unknown>
test       httproute-1  demo.com                           1            <unknown>
test       httproute-2  example.com,example2.com + 1 more  2            <unknown>
`,
		},
		{
			name:      "get services",
			inputArgs: []string{"services"},
			namespace: "default",
			wantOut: `
NAMESPACE  NAME   TYPE     AGE
default    svc-3  Service  <unknown>
`,
		},
		{
			name:      "describe gateways -n test",
			inputArgs: []string{"gateways"},
			namespace: "test",
			describe:  true,
			wantOut: `
Name: gateway-1
Namespace: test
Labels: null
Annotations: null
APIVersion: gateway.networking.k8s.io/v1
Kind: Gateway
Metadata:
  creationTimestamp: null
  uid: uid-for-test-gateway-1
Spec:
  gatewayClassName: foo-com-external-gateway-class
  listeners:
  - name: http
    port: 80
    protocol: HTTP
Status: {}
AttachedRoutes:
  Kind       Name
  ----       ----
  HTTPRoute  test/httproute-1
  HTTPRoute  test/httproute-2
Backends:
  Kind     Name
  ----     ----
  Service  test/svc-1
  Service  test/svc-2
DirectlyAttachedPolicies: <none>
InheritedPolicies: <none>
Events:
  Type     Reason  Age      From                   Message
  ----     ------  ---      ----                   -------
  Warning  SYNC    Unknown  my-gateway-controller  test message


Name: gateway-2
Namespace: test
Labels: null
Annotations: null
APIVersion: gateway.networking.k8s.io/v1
Kind: Gateway
Metadata:
  creationTimestamp: null
Spec:
  gatewayClassName: bar-com-internal-gateway-class
  listeners:
  - name: https
    port: 443
    protocol: HTTPS
Status: {}
AttachedRoutes:
  Kind       Name
  ----       ----
  HTTPRoute  test/httproute-2
Backends:
  Kind     Name
  ----     ----
  Service  test/svc-2
DirectlyAttachedPolicies: <none>
InheritedPolicies: <none>
Events: <none>
`,
		},
		{
			name:      "describe gateways gateway-1 -n test",
			inputArgs: []string{"gateways", "gateway-1"},
			namespace: "test",
			describe:  true,
			wantOut: `
Name: gateway-1
Namespace: test
Labels: null
Annotations: null
APIVersion: gateway.networking.k8s.io/v1
Kind: Gateway
Metadata:
  creationTimestamp: null
  uid: uid-for-test-gateway-1
Spec:
  gatewayClassName: foo-com-external-gateway-class
  listeners:
  - name: http
    port: 80
    protocol: HTTP
Status: {}
AttachedRoutes:
  Kind       Name
  ----       ----
  HTTPRoute  test/httproute-1
  HTTPRoute  test/httproute-2
Backends:
  Kind     Name
  ----     ----
  Service  test/svc-1
  Service  test/svc-2
DirectlyAttachedPolicies: <none>
InheritedPolicies: <none>
Events:
  Type     Reason  Age      From                   Message
  ----     ------  ---      ----                   -------
  Warning  SYNC    Unknown  my-gateway-controller  test message
`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			factory.namespace = tc.namespace

			iostreams, _, out, errOut := genericiooptions.NewTestIOStreams()
			cmd := cmdget.NewCmd(factory, iostreams, tc.describe)
			cmd.SetOut(out)
			cmd.SetErr(out)
			cmd.SetArgs(tc.inputArgs)

			err := cmd.Execute()
			if err != nil {
				t.Logf("Failed to execute command: %v", err)
				t.Logf("Debug: out=\n%v\n", out.String())
				t.Logf("Debug: errOut=\n%v\n", errOut.String())
				t.FailNow()
			}

			got := common.MultiLine(out.String())
			want := common.MultiLine(strings.TrimPrefix(tc.wantOut, "\n"))

			if diff := cmp.Diff(want, got, common.MultiLineTransformer); diff != "" {
				t.Fatalf("Unexpected diff:\n\ngot =\n\n%v\n\nwant =\n\n%v\n\ndiff (-want, +got) =\n\n%v", got, want, common.MultiLine(diff))
			}
		})
	}
}
