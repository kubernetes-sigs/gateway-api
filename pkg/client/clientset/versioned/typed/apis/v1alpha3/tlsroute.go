/*
Copyright The Kubernetes Authors.

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

// Code generated by client-gen. DO NOT EDIT.

package v1alpha3

import (
	context "context"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	gentype "k8s.io/client-go/gentype"
	apisv1alpha3 "sigs.k8s.io/gateway-api/apis/v1alpha3"
	applyconfigurationapisv1alpha3 "sigs.k8s.io/gateway-api/applyconfiguration/apis/v1alpha3"
	scheme "sigs.k8s.io/gateway-api/pkg/client/clientset/versioned/scheme"
)

// TLSRoutesGetter has a method to return a TLSRouteInterface.
// A group's client should implement this interface.
type TLSRoutesGetter interface {
	TLSRoutes(namespace string) TLSRouteInterface
}

// TLSRouteInterface has methods to work with TLSRoute resources.
type TLSRouteInterface interface {
	Create(ctx context.Context, tLSRoute *apisv1alpha3.TLSRoute, opts v1.CreateOptions) (*apisv1alpha3.TLSRoute, error)
	Update(ctx context.Context, tLSRoute *apisv1alpha3.TLSRoute, opts v1.UpdateOptions) (*apisv1alpha3.TLSRoute, error)
	// Add a +genclient:noStatus comment above the type to avoid generating UpdateStatus().
	UpdateStatus(ctx context.Context, tLSRoute *apisv1alpha3.TLSRoute, opts v1.UpdateOptions) (*apisv1alpha3.TLSRoute, error)
	Delete(ctx context.Context, name string, opts v1.DeleteOptions) error
	DeleteCollection(ctx context.Context, opts v1.DeleteOptions, listOpts v1.ListOptions) error
	Get(ctx context.Context, name string, opts v1.GetOptions) (*apisv1alpha3.TLSRoute, error)
	List(ctx context.Context, opts v1.ListOptions) (*apisv1alpha3.TLSRouteList, error)
	Watch(ctx context.Context, opts v1.ListOptions) (watch.Interface, error)
	Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts v1.PatchOptions, subresources ...string) (result *apisv1alpha3.TLSRoute, err error)
	Apply(ctx context.Context, tLSRoute *applyconfigurationapisv1alpha3.TLSRouteApplyConfiguration, opts v1.ApplyOptions) (result *apisv1alpha3.TLSRoute, err error)
	// Add a +genclient:noStatus comment above the type to avoid generating ApplyStatus().
	ApplyStatus(ctx context.Context, tLSRoute *applyconfigurationapisv1alpha3.TLSRouteApplyConfiguration, opts v1.ApplyOptions) (result *apisv1alpha3.TLSRoute, err error)
	TLSRouteExpansion
}

// tLSRoutes implements TLSRouteInterface
type tLSRoutes struct {
	*gentype.ClientWithListAndApply[*apisv1alpha3.TLSRoute, *apisv1alpha3.TLSRouteList, *applyconfigurationapisv1alpha3.TLSRouteApplyConfiguration]
}

// newTLSRoutes returns a TLSRoutes
func newTLSRoutes(c *GatewayV1alpha3Client, namespace string) *tLSRoutes {
	return &tLSRoutes{
		gentype.NewClientWithListAndApply[*apisv1alpha3.TLSRoute, *apisv1alpha3.TLSRouteList, *applyconfigurationapisv1alpha3.TLSRouteApplyConfiguration](
			"tlsroutes",
			c.RESTClient(),
			scheme.ParameterCodec,
			namespace,
			func() *apisv1alpha3.TLSRoute { return &apisv1alpha3.TLSRoute{} },
			func() *apisv1alpha3.TLSRouteList { return &apisv1alpha3.TLSRouteList{} },
		),
	}
}
