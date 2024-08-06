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

package common

import (
	"fmt"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"

	gatewayv1 "sigs.k8s.io/gateway-api/apis/v1"
	gatewayv1beta1 "sigs.k8s.io/gateway-api/apis/v1beta1"
)

const (
	gwctlPolicyGroup = "gwctl.gateway.networking.k8s.io"
)

var (
	GatewayClassGK   schema.GroupKind = schema.GroupKind{Group: gatewayv1.GroupName, Kind: "GatewayClass"}
	GatewayGK        schema.GroupKind = schema.GroupKind{Group: gatewayv1.GroupName, Kind: "Gateway"}
	HTTPRouteGK      schema.GroupKind = schema.GroupKind{Group: gatewayv1.GroupName, Kind: "HTTPRoute"}
	NamespaceGK      schema.GroupKind = schema.GroupKind{Group: corev1.GroupName, Kind: "Namespace"}
	ServiceGK        schema.GroupKind = schema.GroupKind{Group: corev1.GroupName, Kind: "Service"}
	ReferenceGrantGK schema.GroupKind = schema.GroupKind{Group: gatewayv1beta1.GroupName, Kind: "ReferenceGrant"}
	PolicyGK         schema.GroupKind = schema.GroupKind{Group: gwctlPolicyGroup, Kind: "Policy"}
	PolicyCRDGK      schema.GroupKind = schema.GroupKind{Group: gwctlPolicyGroup, Kind: "PolicyCRD"}
)

type GKNN struct {
	Group     string `json:",omitempty"`
	Kind      string `json:",omitempty"`
	Namespace string `json:",omitempty"`
	Name      string `json:",omitempty"`
}

func (g GKNN) GroupKind() schema.GroupKind {
	return schema.GroupKind{
		Group: g.Group,
		Kind:  g.Kind,
	}
}

func (g GKNN) NamespacedName() types.NamespacedName {
	return types.NamespacedName{
		Namespace: g.Namespace,
		Name:      g.Name,
	}
}

func (g GKNN) String() string {
	gk := g.Kind
	if g.Group != "" {
		gk = fmt.Sprintf("%v.%v", g.Kind, g.Group)
	}
	name := g.Name
	if g.Namespace != "" {
		name = fmt.Sprintf("%v/%v", g.Namespace, g.Name)
	}
	return gk + "/" + name
}

func (g GKNN) MarshalText() ([]byte, error) {
	return []byte(g.String()), nil
}

func GKNNFromUnstructured(u *unstructured.Unstructured) GKNN {
	return GKNN{
		Group:     u.GetObjectKind().GroupVersionKind().Group,
		Kind:      u.GetObjectKind().GroupVersionKind().Kind,
		Namespace: u.GetNamespace(),
		Name:      u.GetName(),
	}
}
