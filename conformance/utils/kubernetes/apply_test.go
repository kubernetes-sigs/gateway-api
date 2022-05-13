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

	"sigs.k8s.io/gateway-api/apis/v1beta1"
	_ "sigs.k8s.io/gateway-api/conformance/utils/flags"
)

type given struct {
	resources string
	expected  []unstructured.Unstructured
}

func TestPrepareResources(t *testing.T) {
	tests := []struct {
		name    string
		givens  []given
		applier *Applier
	}{{
		name:    "empty namespace labels",
		applier: NewApplier(nil, nil),
		givens: []given{{
			resources: `
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
		}},
	}, {
		name: "simple namespace labels",
		applier: NewApplier(
			map[string]string{
				"test": "false",
			},
			nil,
		),
		givens: []given{{
			resources: `
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
		}},
	}, {
		name: "overwrite namespace labels",
		applier: NewApplier(
			map[string]string{
				"test": "true",
			},
			nil,
		),
		givens: []given{{
			resources: `
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
		}},
	}, {
		name:    "no listener ports given",
		applier: NewApplier(nil, nil),
		givens: []given{{
			resources: `
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
		}},
	}, {
		name: "multiple gateways each with multiple listeners",
		applier: NewApplier(
			nil,
			[]v1beta1.PortNumber{8000, 8001, 8002, 8003},
		),
		givens: []given{{
			resources: `
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
    - name: https
      port: 443
      protocol: HTTPS
      allowedRoutes:
        namespaces:
          from: Same
---
apiVersion: gateway.networking.k8s.io/v1beta1
kind:       Gateway
metadata:
  name: test2
spec:
  gatewayClassName: {GATEWAY_CLASS_NAME}
  listeners:
    - name: http
      port: 80
      protocol: HTTP
      allowedRoutes:
        namespaces:
          from: Same
    - name: https
      port: 443
      protocol: HTTPS
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
								"port":     int64(8000),
								"protocol": "HTTP",
								"allowedRoutes": map[string]interface{}{
									"namespaces": map[string]interface{}{
										"from": "Same",
									},
								},
							},
							map[string]interface{}{
								"name":     "https",
								"port":     int64(8001),
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
			}, {
				Object: map[string]interface{}{
					"apiVersion": "gateway.networking.k8s.io/v1beta1",
					"kind":       "Gateway",
					"metadata": map[string]interface{}{
						"name": "test2",
					},
					"spec": map[string]interface{}{
						"gatewayClassName": "test-class",
						"listeners": []interface{}{
							map[string]interface{}{
								"name":     "http",
								"port":     int64(8002),
								"protocol": "HTTP",
								"allowedRoutes": map[string]interface{}{
									"namespaces": map[string]interface{}{
										"from": "Same",
									},
								},
							},
							map[string]interface{}{
								"name":     "https",
								"port":     int64(8003),
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
			}},
		}},
	}, {
		name:    "gateway with multiple listeners on the same port",
		applier: NewApplier(nil, []v1beta1.PortNumber{8000}),
		givens: []given{{
			resources: `
apiVersion: gateway.networking.k8s.io/v1beta1
kind:       Gateway
metadata:
  name: test
spec:
  gatewayClassName: {GATEWAY_CLASS_NAME}
  listeners:
    - name: http-a
      port: 80
      hostname: a.test.com
      protocol: HTTP
      allowedRoutes:
        namespaces:
          from: Same
    - name: http-b
      port: 80
      hostname: b.test.com
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
								"name":     "http-a",
								"port":     int64(8000),
								"hostname": "a.test.com",
								"protocol": "HTTP",
								"allowedRoutes": map[string]interface{}{
									"namespaces": map[string]interface{}{
										"from": "Same",
									},
								},
							},
							map[string]interface{}{
								"name":     "http-b",
								"port":     int64(8000),
								"hostname": "b.test.com",
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
		}},
	}, {
		name: "multiple calls to apply with gateways",
		applier: NewApplier(
			nil, []v1beta1.PortNumber{8000, 8001},
		),
		givens: []given{{
			resources: `
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
								"port":     int64(8000),
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
			resources: `
apiVersion: gateway.networking.k8s.io/v1beta1
kind:       Gateway
metadata:
  name: test2
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
						"name": "test2",
					},
					"spec": map[string]interface{}{
						"gatewayClassName": "test-class",
						"listeners": []interface{}{
							map[string]interface{}{
								"name":     "http",
								"port":     int64(8001),
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
		}},
	}, {
		name: "multiple calls to apply with gateways, free ports",
		applier: NewApplier(
			nil, []v1beta1.PortNumber{8000},
		),
		givens: []given{{
			resources: `
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
								"port":     int64(8000),
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
			resources: `
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
								"port":     int64(8000),
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
		}},
	}, {
		name:    "setting the controllerName for a GatewayClass",
		applier: &Applier{},
		givens: []given{{
			resources: `
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
		}},
	}}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			tc.applier.GatewayClass = "test-class"
			tc.applier.ControllerName = "test-controller"
			for i, given := range tc.givens {
				decoder := yaml.NewYAMLOrJSONDecoder(strings.NewReader(given.resources), 4096)

				resources, err := tc.applier.prepareResources(t, decoder)
				require.NoError(t, err, "unexpected error preparing resources")

				require.EqualValues(t, given.expected, resources, "values differ for given resources #%d", i)
			}
		})
	}
}
