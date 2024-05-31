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

package suite

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"sigs.k8s.io/gateway-api/pkg/consts"
)

func TestGetAPIVersionAndChannel(t *testing.T) {
	testCases := []struct {
		name            string
		crds            []apiextensionsv1.CustomResourceDefinition
		expectedVersion string
		expectedChannel string
		err             error
	}{
		{
			name: "no Gateway API CRDs",
			err:  errors.New("no Gateway API CRDs with the proper annotations found in the cluster"),
		},
		{
			name: "properly installed Gateway API CRDs",
			crds: []apiextensionsv1.CustomResourceDefinition{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "gateways.gateway.networking.k8s.io",
						Annotations: map[string]string{
							consts.BundleVersionAnnotation: "v1.2.0-dev",
							consts.ChannelAnnotation:       "standard",
						},
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "httproutes.gateway.networking.k8s.io",
						Annotations: map[string]string{
							consts.BundleVersionAnnotation: "v1.2.0-dev",
							consts.ChannelAnnotation:       "standard",
						},
					},
				},
			},
			expectedVersion: "v1.2.0-dev",
			expectedChannel: "standard",
		},
		{
			name: "properly installed Gateway API CRDs, with additional CRDs",
			crds: []apiextensionsv1.CustomResourceDefinition{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "gateways.gateway.networking.k8s.io",
						Annotations: map[string]string{
							consts.BundleVersionAnnotation: "v1.2.0-dev",
							consts.ChannelAnnotation:       "standard",
						},
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "httproutes.gateway.networking.k8s.io",
						Annotations: map[string]string{
							consts.BundleVersionAnnotation: "v1.2.0-dev",
							consts.ChannelAnnotation:       "standard",
						},
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "crd.fake.group.k8s.io",
					},
				},
			},
			expectedVersion: "v1.2.0-dev",
			expectedChannel: "standard",
		},
		{
			name: "installed Gateway API CRDs having multiple versions",
			crds: []apiextensionsv1.CustomResourceDefinition{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "gateways.gateway.networking.k8s.io",
						Annotations: map[string]string{
							consts.BundleVersionAnnotation: "v1.2.0-dev",
							consts.ChannelAnnotation:       "standard",
						},
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "httproutes.gateway.networking.k8s.io",
						Annotations: map[string]string{
							consts.BundleVersionAnnotation: "v2.0.0",
							consts.ChannelAnnotation:       "standard",
						},
					},
				},
			},
			err: errors.New("multiple gateway API CRDs versions detected"),
		},
		{
			name: "installed Gateway API  CRDs having multiple channels",
			crds: []apiextensionsv1.CustomResourceDefinition{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "gateways.gateway.networking.k8s.io",
						Annotations: map[string]string{
							consts.BundleVersionAnnotation: "v1.2.0-dev",
							consts.ChannelAnnotation:       "standard",
						},
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "httproutes.gateway.networking.k8s.io",
						Annotations: map[string]string{
							consts.BundleVersionAnnotation: "v1.2.0-dev",
							consts.ChannelAnnotation:       "experimental",
						},
					},
				},
			},
			err: errors.New("multiple gateway API CRDs channels detected"),
		},
		{
			name: "installed Gateway API CRDs having partial annotations",
			crds: []apiextensionsv1.CustomResourceDefinition{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "gateways.gateway.networking.k8s.io",
						Annotations: map[string]string{
							consts.BundleVersionAnnotation: "v1.2.0-dev",
							consts.ChannelAnnotation:       "standard",
						},
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "httproutes.gateway.networking.k8s.io",
						Annotations: map[string]string{
							consts.BundleVersionAnnotation: "v1.2.0-dev",
						},
					},
				},
			},
			err: errors.New("detected CRDs with partial version and channel annotations"),
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			version, channel, err := getAPIVersionAndChannel(tc.crds)
			assert.Equal(t, tc.expectedVersion, version)
			assert.Equal(t, tc.expectedChannel, channel)
			assert.Equal(t, tc.err, err)
		})
	}
}
