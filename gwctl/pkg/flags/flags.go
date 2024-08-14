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

package flags

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/pflag"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	gatewayv1 "sigs.k8s.io/gateway-api/apis/v1"
	"sigs.k8s.io/gateway-api/gwctl/pkg/common"
)

type ForFlag string

func NewForFlag() *ForFlag {
	f := ForFlag("")
	return &f
}

func (f *ForFlag) AddFlag(flagSet *pflag.FlagSet) {
	flagSet.StringVar((*string)(f), "for", "", `Filter results to only those related to the specified resource. Format: TYPE[/NAMESPACE]/NAME. Not specifying a NAMESPACE assumes the 'default' value. Examples: gateway/ns2/foo-gateway, httproute/bar-httproute, service/ns1/my-svc`)
}

func (f *ForFlag) ToOption() (common.GKNN, error) {
	objRef := common.GKNN{}

	if *f != "" {
		parts := strings.Split(string(*f), "/")
		if len(parts) < 2 || len(parts) > 3 {
			fmt.Fprintf(os.Stderr, "invalid value used in --for flag; value must be in the format TYPE[/NAMESPACE]/NAME\n")
			os.Exit(1)
		}
		if len(parts) == 2 {
			objRef = common.GKNN{Kind: parts[0], Namespace: metav1.NamespaceDefault, Name: parts[1]}
		} else {
			objRef = common.GKNN{Kind: parts[0], Namespace: parts[1], Name: parts[2]}
		}
		switch strings.ToLower(objRef.Kind) {
		case "gatewayclass", "gateawyclasses":
			objRef.Group = gatewayv1.GroupVersion.Group
			objRef.Kind = "GatewayClass"
			objRef.Namespace = ""
		case "gateway", "gateways":
			objRef.Group = gatewayv1.GroupVersion.Group
			objRef.Kind = "Gateway"
		case "httproute", "httproutes":
			objRef.Group = gatewayv1.GroupVersion.Group
			objRef.Kind = "HTTPRoute"
		case "service", "services":
			objRef.Kind = "Service"
		default:
			fmt.Fprintf(os.Stderr, "invalid type provided in --for flag; type must be one of [gatewayclass, gateway, httproute, service]\n")
			os.Exit(1)
		}
	}

	return objRef, nil
}
