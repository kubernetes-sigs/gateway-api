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

package effectivepolicy

import (
	"sigs.k8s.io/gateway-api/gwctl/pkg/common"
	"sigs.k8s.io/gateway-api/gwctl/pkg/policymanager"
)

type Calculator struct {
	k8sClients    *common.K8sClients
	policyManager *policymanager.PolicyManager

	GatewayClasses *gatewayClasses
	Namespaces     *namespaces
	Gateways       *gateways
	HTTPRoutes     *httpRoutes
	Backends       *backends
}

func NewCalculator(k8sClients *common.K8sClients, policyManager *policymanager.PolicyManager) *Calculator {
	epc := &Calculator{
		k8sClients:     k8sClients,
		policyManager:  policyManager,
		GatewayClasses: &gatewayClasses{},
		Namespaces:     &namespaces{},
		Gateways:       &gateways{},
		HTTPRoutes:     &httpRoutes{},
		Backends:       &backends{},
	}

	epc.Namespaces.epc = epc
	epc.GatewayClasses.epc = epc
	epc.Gateways.epc = epc
	epc.HTTPRoutes.epc = epc
	epc.Backends.epc = epc

	return epc
}
