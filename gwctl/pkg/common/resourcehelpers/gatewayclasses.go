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

package resourcehelpers

import (
	"context"
	_ "embed"

	apimachinerytypes "k8s.io/apimachinery/pkg/types"
	gatewayv1beta1 "sigs.k8s.io/gateway-api/apis/v1beta1"
	"sigs.k8s.io/gateway-api/gwctl/pkg/common"
)

func ListGatewayClasses(ctx context.Context, k8sClients *common.K8sClients) ([]gatewayv1beta1.GatewayClass, error) {
	gwcList := &gatewayv1beta1.GatewayClassList{}
	if err := k8sClients.Client.List(ctx, gwcList); err != nil {
		return []gatewayv1beta1.GatewayClass{}, err
	}

	return gwcList.Items, nil
}

func GetGatewayClass(ctx context.Context, k8sClients *common.K8sClients, name string) (gatewayv1beta1.GatewayClass, error) {
	gwc := &gatewayv1beta1.GatewayClass{}
	nn := apimachinerytypes.NamespacedName{Name: name}
	if err := k8sClients.Client.Get(ctx, nn, gwc); err != nil {
		return gatewayv1beta1.GatewayClass{}, err
	}

	return *gwc, nil
}
