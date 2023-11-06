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

	apimachinerytypes "k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	gatewayv1beta1 "sigs.k8s.io/gateway-api/apis/v1beta1"
	"sigs.k8s.io/gateway-api/gwctl/pkg/common"
)

func ListGateways(ctx context.Context, k8sClients *common.K8sClients, namespace string) ([]gatewayv1beta1.Gateway, error) {
	gwList := &gatewayv1beta1.GatewayList{}
	if err := k8sClients.Client.List(ctx, gwList, client.InNamespace(namespace)); err != nil {
		return []gatewayv1beta1.Gateway{}, err
	}

	return gwList.Items, nil
}

func GetGateways(ctx context.Context, k8sClients *common.K8sClients, namespace, name string) (gatewayv1beta1.Gateway, error) {
	gw := &gatewayv1beta1.Gateway{}
	nn := apimachinerytypes.NamespacedName{Namespace: namespace, Name: name}
	if err := k8sClients.Client.Get(ctx, nn, gw); err != nil {
		return gatewayv1beta1.Gateway{}, err
	}

	return *gw, nil
}
