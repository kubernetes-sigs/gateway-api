/*
Copyright 2025 The Kubernetes Authors.

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

package server

import (
	"fmt"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"

	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/client"
	metricsserver "sigs.k8s.io/controller-runtime/pkg/metrics/server"
	v1 "sigs.k8s.io/gateway-api-inference-extension/api/v1"
	"sigs.k8s.io/gateway-api/conformance/epp-basic/pkg/common"
)

var scheme = runtime.NewScheme()

func init() {
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))
	utilruntime.Must(v1.Install(scheme))
}

// defaultManagerOptions returns the default options used to create the manager.
func defaultManagerOptions(cfg ControllerConfig, gknn common.GKNN, metricsServerOptions metricsserver.Options) ctrl.Options {
	opt := ctrl.Options{
		Scheme: scheme,
		Cache: cache.Options{
			ByObject: map[client.Object]cache.ByObject{
				&corev1.Pod{}: {
					Namespaces: map[string]cache.Config{
						gknn.Namespace: {},
					},
				},
			},
		},
		Metrics: metricsServerOptions,
	}
	if cfg.startCrdReconcilers {

		opt.Cache.ByObject[&v1.InferencePool{}] = cache.ByObject{
			Namespaces: map[string]cache.Config{gknn.Namespace: {FieldSelector: fields.SelectorFromSet(fields.Set{
				"metadata.name": gknn.Name,
			})}},
		}
	}
	return opt
}

// NewDefaultManager creates a new controller manager with default configuration.
// Optional override functions can be passed to customize the manager options (e.g., for testing).
func NewDefaultManager(controllerCfg ControllerConfig, gknn common.GKNN, restConfig *rest.Config, metricsServerOptions metricsserver.Options, overrides ...func(*ctrl.Options)) (ctrl.Manager, error) {
	opt := defaultManagerOptions(controllerCfg, gknn, metricsServerOptions)

	for _, override := range overrides {
		override(&opt)
	}

	manager, err := ctrl.NewManager(restConfig, opt)

	if err != nil {
		return nil, fmt.Errorf("failed to create controller manager: %v", err)
	}
	return manager, nil
}
