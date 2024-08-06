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

package printer

import (
	"bytes"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"

	"sigs.k8s.io/gateway-api/gwctl/pkg/common"
)

func TestTablePrinter_printHTTPRoute(t *testing.T) {
	options := PrinterOptions{}
	p := &TablePrinter{PrinterOptions: options}
	out := &bytes.Buffer{}

	for _, ns := range testData(t)[common.HTTPRouteGK] {
		p.printHTTPRoute(ns, out)
		p.Flush(out)
	}

	wantOut := `
NAMESPACE  NAME          HOSTNAMES                 PARENT REFS  AGE
ns-1       http-route-1  foo.com,bar.com + 5 more  1            <unknown>
`

	got := common.MultiLine(out.String())
	want := common.MultiLine(strings.TrimPrefix(wantOut, "\n"))

	if diff := cmp.Diff(want, got, common.MultiLineTransformer); diff != "" {
		t.Fatalf("Unexpected diff:\n\ngot =\n\n%v\n\nwant =\n\n%v\n\ndiff (-want, +got) =\n\n%v", got, want, common.MultiLine(diff))
	}
}
