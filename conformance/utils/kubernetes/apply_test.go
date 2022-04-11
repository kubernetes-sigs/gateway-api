/*
Copyright 2022 The Kubernetes Authors.

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

package kubernetes

import (
	"testing"

	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	"sigs.k8s.io/gateway-api/apis/v1alpha2"
)

func Test_prepareNamespace_empty(t *testing.T) {
	ns := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "v1",
			"kind":       "Namespace",
			"metadata": map[string]interface{}{
				"name": "test",
			},
		},
	}

	prepareNamespace(t, ns, nil)

	labels, _, err := unstructured.NestedMap(ns.Object, "metadata", "labels")
	require.NoError(t, err, "unexpected error getting labels")

	require.EqualValues(
		t,
		labels,
		map[string]interface{}{},
	)
}

func Test_prepareNamespace_simple(t *testing.T) {
	ns := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "v1",
			"kind":       "Namespace",
			"metadata": map[string]interface{}{
				"name": "test",
			},
		},
	}

	prepareNamespace(t, ns, map[string]string{
		"test": "true",
	})

	labels, _, err := unstructured.NestedMap(ns.Object, "metadata", "labels")
	require.NoError(t, err, "unexpected error getting labels")

	require.EqualValues(
		t,
		labels,
		map[string]interface{}{
			"test": "true",
		}, "unexpected Namespace labels",
	)
}

func Test_prepareNamespace_overwrite(t *testing.T) {
	ns := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "v1",
			"kind":       "Namespace",
			"metadata": map[string]interface{}{
				"name": "test",
				"labels": map[string]interface{}{
					"test": "false",
				},
			},
		},
	}

	prepareNamespace(t, ns, map[string]string{
		"test": "true",
	})

	labels, _, err := unstructured.NestedMap(ns.Object, "metadata", "labels")
	require.NoError(t, err, "unexpected error getting labels")

	require.EqualValues(
		t,
		labels,
		map[string]interface{}{
			"test": "true",
		}, "unexpected Namespace labels",
	)
}

func Test_prepareGateway_noPorts(t *testing.T) {
	gateway := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "gateway.networking.k8s.io/v1alpha2",
			"kind":       "Gateway",
			"metadata": map[string]interface{}{
				"name": "test",
			},
			"spec": map[string]interface{}{
				"gatewayClassName": "{GATEWAY_CLASS_NAME}",
				"listeners": []interface{}{
					map[string]interface{}{
						"name":     "http",
						"port":     80,
						"protocol": "HTTP",
						"allowedRoutes": map[string]interface{}{
							"namespaces": map[string]interface{}{
								"from": "Same",
							},
						},
					},
				},
			},
		},
	}
	unchanged := *gateway

	nextPort := prepareGateway(t, gateway, "test", nil, 0)
	require.Equal(t, 0, nextPort, "unexpected next valid port index")

	require.EqualValues(
		t,
		gateway,
		&unchanged,
		"expected Gateway to be unchanged",
	)
}

func Test_prepareGateway_ports(t *testing.T) {
	gateway := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "gateway.networking.k8s.io/v1alpha2",
			"kind":       "Gateway",
			"metadata": map[string]interface{}{
				"name": "test",
			},
			"spec": map[string]interface{}{
				"gatewayClassName": "{GATEWAY_CLASS_NAME}",
				"listeners": []interface{}{
					map[string]interface{}{
						"name":     "http",
						"port":     float64(80),
						"protocol": "HTTP",
						"allowedRoutes": map[string]interface{}{
							"namespaces": map[string]interface{}{
								"from": "Same",
							},
						},
					},
					map[string]interface{}{
						"name":     "https",
						"port":     float64(443),
						"protocol": "HTTPS",
						"allowedRoutes": map[string]interface{}{
							"namespaces": map[string]interface{}{
								"from": "Same",
							},
						},
					},
				},
			},
		},
	}

	nextPort := prepareGateway(t, gateway, "test", []v1alpha2.PortNumber{30080, 30081}, 0)
	require.Equal(t, 2, nextPort, "unexpected next valid port index")

	listeners, _, err := unstructured.NestedSlice(gateway.Object, "spec", "listeners")
	require.NoError(t, err, "unexpected error getting listeners")
	port, _, err := unstructured.NestedFieldCopy(listeners[0].(map[string]interface{}), "port")
	require.NoError(t, err, "unexpected error getting first listener port")
	require.EqualValues(
		t,
		30080,
		port,
		"unexpected first Gateway listener port",
	)

	port, _, err = unstructured.NestedFieldCopy(listeners[1].(map[string]interface{}), "port")
	require.NoError(t, err, "unexpected error getting second listener port")
	require.EqualValues(
		t,
		30081,
		port,
		"unexpected second Gateway listener port",
	)

}
