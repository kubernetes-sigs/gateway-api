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
	"context"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/yaml"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/client/interceptor"

	"sigs.k8s.io/gateway-api/conformance/utils/config"
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
			Object: map[string]any{
				"apiVersion": "v1",
				"kind":       "Namespace",
				"metadata": map[string]any{
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
			Object: map[string]any{
				"apiVersion": "v1",
				"kind":       "Namespace",
				"metadata": map[string]any{
					"name": "test",
					"labels": map[string]any{
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
			Object: map[string]any{
				"apiVersion": "v1",
				"kind":       "Namespace",
				"metadata": map[string]any{
					"name": "test",
					"labels": map[string]any{
						"test": "true",
					},
				},
			},
		}},
	}, {
		name:    "setting the gatewayClassName",
		applier: Applier{},
		given: `
apiVersion: gateway.networking.k8s.io/v1
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
			Object: map[string]any{
				"apiVersion": "gateway.networking.k8s.io/v1",
				"kind":       "Gateway",
				"metadata": map[string]any{
					"name": "test",
				},
				"spec": map[string]any{
					"gatewayClassName": "test-class",
					"listeners": []any{
						map[string]any{
							"name":     "http",
							"port":     int64(80),
							"protocol": "HTTP",
							"allowedRoutes": map[string]any{
								"namespaces": map[string]any{
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
apiVersion: gateway.networking.k8s.io/v1
kind:       GatewayClass
metadata:
  name: test
spec:
  controllerName: {GATEWAY_CONTROLLER_NAME}
`,
		expected: []unstructured.Unstructured{{
			Object: map[string]any{
				"apiVersion": "gateway.networking.k8s.io/v1",
				"kind":       "GatewayClass",
				"metadata": map[string]any{
					"name": "test",
				},
				"spec": map[string]any{
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
			require.Equal(t, tc.expected, resources)
		})
	}
}

func TestDeleteAndWaitBlocksUntilGone(t *testing.T) {
	scheme := runtime.NewScheme()
	require.NoError(t, corev1.AddToScheme(scheme))

	tests := []struct {
		name string
		obj  client.Object
	}{
		{
			name: "namespace held by a finalizer",
			obj: &corev1.Namespace{ObjectMeta: metav1.ObjectMeta{
				Name: "conformance-ns", Finalizers: []string{"conformance.gateway-api/test"},
			}},
		},
		{
			name: "namespaced resource held by a finalizer",
			obj: &corev1.ConfigMap{ObjectMeta: metav1.ObjectMeta{
				Name: "held", Namespace: "default", Finalizers: []string{"conformance.gateway-api/test"},
			}},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			key := client.ObjectKeyFromObject(tc.obj)

			// Object present on the first poll, gone on the next: drives the wait loop with no timing dependence.
			var polls int
			c := fake.NewClientBuilder().WithScheme(scheme).WithObjects(tc.obj).
				WithInterceptorFuncs(interceptor.Funcs{
					Get: func(ctx context.Context, cl client.WithWatch, key client.ObjectKey, obj client.Object, opts ...client.GetOption) error {
						polls++
						if polls >= 2 {
							fresh := tc.obj.DeepCopyObject().(client.Object)
							if err := cl.Get(ctx, key, fresh); err == nil {
								fresh.SetFinalizers(nil)
								_ = cl.Update(ctx, fresh)
							}
						}
						return cl.Get(ctx, key, obj, opts...)
					},
				}).Build()

			deleteAndWait(t, c, tc.obj, config.TimeoutConfig{DeleteTimeout: 5 * time.Second, DefaultPollInterval: time.Millisecond})

			probe := tc.obj.DeepCopyObject().(client.Object)
			err := c.Get(context.Background(), key, probe)
			require.True(t, apierrors.IsNotFound(err),
				"%s must be fully deleted after deleteAndWait returns, got err=%v", tc.name, err)
		})
	}
}
