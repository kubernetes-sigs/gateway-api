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
	"bytes"
	"context"
	"errors"
	"io"
	"io/ioutil"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/yaml"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// MustApplyWithCleanup creates or updates Kubernetes resources defined with the
// provided YAML file and registers a cleanup function for resources it created.
// Note that this does not remove resources that already existed in the cluster.
func MustApplyWithCleanup(t *testing.T, c client.Client, path string, gcName string, cleanup bool) {
	b, err := ioutil.ReadFile(path)
	require.NoErrorf(t, err, "error reading %s file", path)

	decoder := yaml.NewYAMLOrJSONDecoder(bytes.NewReader(b), 4096)
	for {
		uObj := unstructured.Unstructured{}
		if decodeErr := decoder.Decode(&uObj); decodeErr != nil {
			if errors.Is(decodeErr, io.EOF) {
				break
			}
			t.Logf("manifest: %s", string(b))
			require.NoErrorf(t, decodeErr, "error parsing manifest")
		}
		if len(uObj.Object) == 0 {
			continue
		}

		if uObj.GetKind() == "Gateway" {
			err = unstructured.SetNestedField(uObj.Object, gcName, "spec", "gatewayClassName")
			require.NoErrorf(t, err, "error setting `spec.gatewayClassName` on %s Gateway resource", uObj.GetName())
		}

		ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
		defer cancel()

		namespacedName := types.NamespacedName{Namespace: uObj.GetNamespace(), Name: uObj.GetName()}
		fetchedObj := uObj.DeepCopy()
		err := c.Get(ctx, namespacedName, fetchedObj)
		if err != nil {
			if !apierrors.IsNotFound(err) {
				require.NoErrorf(t, err, "error getting resource")
			}
			t.Logf("Creating %s %s", uObj.GetName(), uObj.GetKind())
			err = c.Create(ctx, &uObj)
			require.NoErrorf(t, err, "error creating resource")

			if cleanup {
				t.Cleanup(func() {
					err = c.Delete(ctx, &uObj)
					require.NoErrorf(t, err, "error deleting resource")
				})
			}
			continue
		}

		uObj.SetResourceVersion(fetchedObj.GetResourceVersion())
		t.Logf("Updating %s %s", uObj.GetName(), uObj.GetKind())
		err = c.Update(ctx, &uObj)

		if cleanup {
			t.Cleanup(func() {
				err = c.Delete(ctx, &uObj)
				require.NoErrorf(t, err, "error deleting resource")
			})
		}
		require.NoErrorf(t, err, "error updating resource")
	}
}
