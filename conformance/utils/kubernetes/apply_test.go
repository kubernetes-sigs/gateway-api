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
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/util/yaml"

	_ "sigs.k8s.io/gateway-api/conformance/utils/flags"
)

func TestPrepareResources(t *testing.T) {
	tests := []struct {
		name     string
		given    string
		expected []unstructured.Unstructured
		applier  Applier
	}{{
		name:    "empty namespace labels",
		applier: Applier{},
		given: `
apiVersion: v1
kind: Namespace
metadata:
  name: test
`,
		expected: []unstructured.Unstructured{{
			Object: map[string]interface{}{
				"apiVersion": "v1",
				"kind":       "Namespace",
				"metadata": map[string]interface{}{
					"name": "test",
				},
			},
		}},
	}, {
		name: "simple namespace labels",
		applier: Applier{
			NamespaceLabels: map[string]string{
				"test": "false",
			},
		},
		given: `
apiVersion: v1
kind: Namespace
metadata:
  name: test
`,
		expected: []unstructured.Unstructured{{
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
		}},
	}, {
		name: "overwrite namespace labels",
		applier: Applier{
			NamespaceLabels: map[string]string{
				"test": "true",
			},
		},
		given: `
apiVersion: v1
kind: Namespace
metadata:
  name: test
  labels:
    test: 'false'
`,
		expected: []unstructured.Unstructured{{
			Object: map[string]interface{}{
				"apiVersion": "v1",
				"kind":       "Namespace",
				"metadata": map[string]interface{}{
					"name": "test",
					"labels": map[string]interface{}{
						"test": "true",
					},
				},
			},
		}},
	}, {
		name:    "setting the gatewayClassName",
		applier: Applier{},
		given: `
apiVersion: gateway.networking.k8s.io/v1beta1
kind:       Gateway
metadata:
  name: test
spec:
  gatewayClassName: {GATEWAY_CLASS_NAME}
  listeners:
    - name: http
      port: 80
      protocol: HTTP
      allowedRoutes:
        namespaces:
          from: Same
`,
		expected: []unstructured.Unstructured{{
			Object: map[string]interface{}{
				"apiVersion": "gateway.networking.k8s.io/v1beta1",
				"kind":       "Gateway",
				"metadata": map[string]interface{}{
					"name": "test",
				},
				"spec": map[string]interface{}{
					"gatewayClassName": "test-class",
					"listeners": []interface{}{
						map[string]interface{}{
							"name":     "http",
							"port":     int64(80),
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
		}},
	}, {
		name:    "setting the controllerName for a GatewayClass",
		applier: Applier{},
		given: `
apiVersion: gateway.networking.k8s.io/v1beta1
kind:       GatewayClass
metadata:
  name: test
spec:
  controllerName: {GATEWAY_CONTROLLER_NAME}
`,
		expected: []unstructured.Unstructured{{
			Object: map[string]interface{}{
				"apiVersion": "gateway.networking.k8s.io/v1beta1",
				"kind":       "GatewayClass",
				"metadata": map[string]interface{}{
					"name": "test",
				},
				"spec": map[string]interface{}{
					"controllerName": "test-controller",
				},
			},
		}},
	}}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			decoder := yaml.NewYAMLOrJSONDecoder(strings.NewReader(tc.given), 4096)

			tc.applier.GatewayClass = "test-class"
			tc.applier.ControllerName = "test-controller"
			resources, err := tc.applier.prepareResources(t, decoder)

			require.NoError(t, err, "unexpected error preparing resources")
			require.EqualValues(t, tc.expected, resources)
		})
	}
}
