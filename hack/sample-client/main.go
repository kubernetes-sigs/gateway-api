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

package main

import (
	"context"
	"fmt"
	"log"

	gatewayxv1alpha2 "sigs.k8s.io/gateway-api/apis/experimental/v1alpha2"
	gatewayxv1beta1 "sigs.k8s.io/gateway-api/apis/experimental/v1beta1"
	gatewayv1 "sigs.k8s.io/gateway-api/apis/v1"
	gatewayv1alpha2 "sigs.k8s.io/gateway-api/apis/v1alpha2"
	gatewayv1beta1 "sigs.k8s.io/gateway-api/apis/v1beta1"

	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
)

func main() {
	cfg, err := config.GetConfig()
	if err != nil {
		log.Fatalf("Error loading Kubernetes config: %v", err)
	}

	c, err := client.NewWithWatch(cfg, client.Options{})
	if err != nil {
		log.Fatalf("Error setting up Kubernetes client: %v", err)
	}

	gatewayxv1alpha2.AddToScheme(c.Scheme())
	gatewayxv1beta1.AddToScheme(c.Scheme())
	gatewayv1alpha2.AddToScheme(c.Scheme())
	gatewayv1beta1.AddToScheme(c.Scheme())
	gatewayv1.AddToScheme(c.Scheme())

	xgwList := gatewayxv1beta1.XGatewayList{}
	err = c.List(context.TODO(), &xgwList)
	if err != nil {
		log.Fatalf("Error listing experimental Gateways: %v", err)
	}
	printGateways("experimental", gatewayv1.GatewayList(xgwList))

	gwList := gatewayv1.GatewayList{}
	err = c.List(context.TODO(), &gwList)
	if err != nil {
		log.Fatalf("Error listing standard Gateways: %v", err)
	}
	printGateways("standard", gwList)
}

func printGateways(channel string, gwList gatewayv1.GatewayList) {
	fmt.Printf("Printing %s channel Gateways:\n", channel)
	for _, gw := range gwList.Items {
		fmt.Printf("- %s\n", gw.Name)
	}
}
