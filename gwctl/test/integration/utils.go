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

package integration

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"

	corev1 "k8s.io/api/core/v1"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/cli-runtime/pkg/resource"
	"k8s.io/client-go/discovery/cached/memory"
	fakediscovery "k8s.io/client-go/discovery/fake"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest/fake"
	"k8s.io/client-go/restmapper"
	clientgotesting "k8s.io/client-go/testing"
	cmdtesting "k8s.io/kubectl/pkg/cmd/testing"

	gatewayv1 "sigs.k8s.io/gateway-api/apis/v1"
	"sigs.k8s.io/gateway-api/gwctl/pkg/common"
)

type TestFactory struct {
	namespace          string
	unstructuredClient resource.RESTClient
	restMapper         meta.RESTMapper
}

func NewTestFactory(t *testing.T, yamls ...string) *TestFactory {
	yaml := strings.Join(yamls, "\n---\n")

	infos, err := resource.NewLocalBuilder().
		Unstructured().
		Stream(bytes.NewBufferString(yaml), "input").
		Flatten().
		ContinueOnError().
		Do().
		Infos()
	if err != nil {
		t.Fatal(err)
	}

	restMapper := mustRestMapper(t, infos)

	f := &TestFactory{
		unstructuredClient: mustFakeRestClient(t, infos, restMapper),
		restMapper:         restMapper,
	}
	return f
}

func (f *TestFactory) NewBuilder() *resource.Builder {
	return resource.NewFakeBuilder(
		func(_ schema.GroupVersion) (resource.RESTClient, error) {
			return f.unstructuredClient, nil
		},
		func() (meta.RESTMapper, error) {
			return f.restMapper, nil
		},
		func() (restmapper.CategoryExpander, error) {
			return resource.FakeCategoryExpander, nil
		},
	)
}

func (f *TestFactory) KubeConfigNamespace() (string, bool, error) {
	return f.namespace, false, nil
}

// mustRestMapper maintains a set of all resources recognized by the fake server.
func mustRestMapper(t *testing.T, infos []*resource.Info) meta.RESTMapper {
	resourceList := []*metav1.APIResourceList{
		{
			GroupVersion: gatewayv1.SchemeGroupVersion.String(),
			APIResources: []metav1.APIResource{
				{Name: "gatewayclasses", Namespaced: false, Kind: common.GatewayClassGK.Kind},
				{Name: "gateways", Namespaced: true, Kind: common.GatewayGK.Kind},
				{Name: "httproutes", Namespaced: true, Kind: common.HTTPRouteGK.Kind},
				{Name: "referencegrants", Namespaced: true, Kind: common.ReferenceGrantGK.Kind},
			},
		},
		{
			GroupVersion: corev1.SchemeGroupVersion.String(),
			APIResources: []metav1.APIResource{
				{Name: "pods", Namespaced: true, Kind: "Pod"},
				{Name: "services", Namespaced: true, Kind: "Service"},
				{Name: "secrets", Namespaced: true, Kind: "Secret"},
				{Name: "namespaces", Namespaced: false, Kind: "Namespace"},
				{Name: "events", Namespaced: true, Kind: "Event"},
			},
		},
		{
			GroupVersion: apiextensionsv1.SchemeGroupVersion.String(),
			APIResources: []metav1.APIResource{
				{Name: "customresourcedefinitions", Namespaced: false, Kind: "CustomResourceDefinition"},
			},
		},
	}

	// For each CRD, make the underlying API Kind available for discovery

	crdResourcesByGroupVersion := map[string][]metav1.APIResource{}
	for _, info := range infos {
		obj, err := runtime.DefaultUnstructuredConverter.ToUnstructured(info.Object)
		if err != nil {
			t.Fatal(err)
		}
		u := &unstructured.Unstructured{Object: obj}
		if u.GroupVersionKind().Kind != "CustomResourceDefinition" {
			continue
		}

		crd := &apiextensionsv1.CustomResourceDefinition{}
		if err := runtime.DefaultUnstructuredConverter.FromUnstructured(u.Object, crd); err != nil {
			t.Fatalf("failed to convert unstructured to CustomResourceDefinitino: %v", err)
		}

		// Build the APIResourceList
		group := crd.Spec.Group
		version := "v1"
		if len(crd.Spec.Versions) != 0 {
			version = crd.Spec.Versions[0].Name
		}
		groupVersion := group + "/" + version
		crdResourcesByGroupVersion[groupVersion] = append(crdResourcesByGroupVersion[groupVersion], metav1.APIResource{
			Name:       crd.Spec.Names.Plural,
			Namespaced: crd.Spec.Scope == apiextensionsv1.NamespaceScoped,
			Kind:       crd.Spec.Names.Kind,
		})
	}
	for groupVersion, apiResources := range crdResourcesByGroupVersion {
		resourceList = append(resourceList, &metav1.APIResourceList{
			GroupVersion: groupVersion,
			APIResources: apiResources,
		})
	}

	fakeDiscoveryClient := &fakediscovery.FakeDiscovery{
		Fake: &clientgotesting.Fake{
			Resources: resourceList,
		},
	}
	cachedDiscoveryClient := memory.NewMemCacheClient(fakeDiscoveryClient)
	return restmapper.NewDeferredDiscoveryRESTMapper(cachedDiscoveryClient)
}

func mustFakeRestClient(t *testing.T, infos []*resource.Info, restMapper meta.RESTMapper) *fake.RESTClient {
	resourcesByPath := loadResourcesByPath(t, infos, restMapper)

	codec := unstructured.NewJSONFallbackEncoder(scheme.Codecs.LegacyCodec(scheme.Scheme.PrioritizedVersionsAllGroups()...))
	roundTripper := func(req *http.Request) (*http.Response, error) {
		path, method := req.URL.Path, req.Method
		pathAndQuery := path
		if req.URL.RawQuery != "" {
			pathAndQuery = path + "?" + req.URL.RawQuery
		}

		if method != "GET" {
			t.Fatalf("request url: %+v, and request: %+v", req.URL, req)
			return nil, nil
		}

		responseBody := resourcesByPath[pathAndQuery]
		if responseBody == nil {
			t.Logf("No resources found, request url: %+v, and request: %+v", req.URL, req)

			responseBody = &unstructured.UnstructuredList{
				Object: map[string]interface{}{
					"apiVersion": "v1",
					"kind":       "List",
				},
			}
		}

		return &http.Response{
			StatusCode: http.StatusOK,
			Header:     cmdtesting.DefaultHeader(),
			Body:       io.NopCloser(bytes.NewReader([]byte(runtime.EncodeOrDie(codec, responseBody)))),
		}, nil
	}

	result := &fake.RESTClient{
		NegotiatedSerializer: resource.UnstructuredPlusDefaultContentConfig().NegotiatedSerializer,
		Client:               fake.CreateHTTPClient(roundTripper),
	}

	return result
}

func loadResourcesByPath(t *testing.T, infos []*resource.Info, restMapper meta.RESTMapper) map[string]runtime.Object {
	lister := map[string][]*unstructured.Unstructured{}
	getter := map[string]*unstructured.Unstructured{}

	for _, info := range infos {
		t.Logf("Loading resource: %v", info)

		// Find REST Path
		restMapping, err := restMapper.RESTMapping(info.Object.GetObjectKind().GroupVersionKind().GroupKind())
		if err != nil {
			t.Fatal(err)
		}

		// Conver to unstructured
		obj, err := runtime.DefaultUnstructuredConverter.ToUnstructured(info.Object)
		if err != nil {
			t.Fatal(err)
		}
		u := &unstructured.Unstructured{Object: obj}

		// Add path for listing across all namespaces.
		listPath := "/" + restMapping.Resource.Resource
		lister[listPath] = append(lister[listPath], u)

		// Special handling for Events which are listed using a field selector:
		// 'involvedObject.uid'
		if restMapping.GroupVersionKind.Kind == "Event" {
			event := &corev1.Event{}
			if err := runtime.DefaultUnstructuredConverter.FromUnstructured(u.Object, event); err != nil {
				t.Fatal(err)
			}
			query := fmt.Sprintf("fieldSelector=involvedObject.uid%%3D%v", event.InvolvedObject.UID)
			path := listPath + "?" + query
			lister[path] = append(lister[path], u)
		}

		// In case of namespaced resources, add path for listing for the specific namespace
		if restMapping.Scope.Name() == meta.RESTScopeNameNamespace {
			listPath = "/namespaces/" + info.Namespace + listPath
			lister[listPath] = append(lister[listPath], u)
		}

		// Add path for getting the individual resource
		getter[listPath+"/"+info.Name] = u
	}

	result := map[string]runtime.Object{}
	for key, items := range lister {
		uList := &unstructured.UnstructuredList{
			Object: map[string]interface{}{
				"apiVersion": "v1",
				"kind":       "List",
			},
		}
		for _, item := range items {
			t.Logf("adding %v for path %v", item.GetNamespace()+"/"+item.GetName(), key)
			uList.Items = append(uList.Items, *item)
		}
		result[key] = uList
	}
	for key, value := range getter {
		result[key] = value
	}
	return result
}
