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

package v1alpha2

import (
	context "context"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	gentype "k8s.io/client-go/gentype"
	apisv1alpha2 "sigs.k8s.io/gateway-api/apis/v1alpha2"
	applyconfigurationapisv1alpha2 "sigs.k8s.io/gateway-api/applyconfiguration/apis/v1alpha2"
	scheme "sigs.k8s.io/gateway-api/pkg/client/clientset/versioned/scheme"
)

// GRPCRoutesGetter has a method to return a GRPCRouteInterface.
// A group's client should implement this interface.
type GRPCRoutesGetter interface {
	GRPCRoutes(namespace string) GRPCRouteInterface
}

// GRPCRouteInterface has methods to work with GRPCRoute resources.
type GRPCRouteInterface interface {
	Create(ctx context.Context, gRPCRoute *apisv1alpha2.GRPCRoute, opts v1.CreateOptions) (*apisv1alpha2.GRPCRoute, error)
	Update(ctx context.Context, gRPCRoute *apisv1alpha2.GRPCRoute, opts v1.UpdateOptions) (*apisv1alpha2.GRPCRoute, error)
	// Add a +genclient:noStatus comment above the type to avoid generating UpdateStatus().
	UpdateStatus(ctx context.Context, gRPCRoute *apisv1alpha2.GRPCRoute, opts v1.UpdateOptions) (*apisv1alpha2.GRPCRoute, error)
	Delete(ctx context.Context, name string, opts v1.DeleteOptions) error
	DeleteCollection(ctx context.Context, opts v1.DeleteOptions, listOpts v1.ListOptions) error
	Get(ctx context.Context, name string, opts v1.GetOptions) (*apisv1alpha2.GRPCRoute, error)
	List(ctx context.Context, opts v1.ListOptions) (*apisv1alpha2.GRPCRouteList, error)
	Watch(ctx context.Context, opts v1.ListOptions) (watch.Interface, error)
	Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts v1.PatchOptions, subresources ...string) (result *apisv1alpha2.GRPCRoute, err error)
	Apply(ctx context.Context, gRPCRoute *applyconfigurationapisv1alpha2.GRPCRouteApplyConfiguration, opts v1.ApplyOptions) (result *apisv1alpha2.GRPCRoute, err error)
	// Add a +genclient:noStatus comment above the type to avoid generating ApplyStatus().
	ApplyStatus(ctx context.Context, gRPCRoute *applyconfigurationapisv1alpha2.GRPCRouteApplyConfiguration, opts v1.ApplyOptions) (result *apisv1alpha2.GRPCRoute, err error)
	GRPCRouteExpansion
}

// gRPCRoutes implements GRPCRouteInterface
type gRPCRoutes struct {
	*gentype.ClientWithListAndApply[*apisv1alpha2.GRPCRoute, *apisv1alpha2.GRPCRouteList, *applyconfigurationapisv1alpha2.GRPCRouteApplyConfiguration]
}

// newGRPCRoutes returns a GRPCRoutes
func newGRPCRoutes(c *GatewayV1alpha2Client, namespace string) *gRPCRoutes {
	return &gRPCRoutes{
		gentype.NewClientWithListAndApply[*apisv1alpha2.GRPCRoute, *apisv1alpha2.GRPCRouteList, *applyconfigurationapisv1alpha2.GRPCRouteApplyConfiguration](
			"grpcroutes",
			c.RESTClient(),
			scheme.ParameterCodec,
			namespace,
			func() *apisv1alpha2.GRPCRoute { return &apisv1alpha2.GRPCRoute{} },
			func() *apisv1alpha2.GRPCRouteList { return &apisv1alpha2.GRPCRouteList{} },
		),
	}
}
