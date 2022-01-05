/*
Copyright 2021 The Kubernetes Authors.

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

package v1alpha2

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

func TestRefKindMatch(t *testing.T) {
	group := func(s string) *Group {
		g := Group(s)
		return &g
	}

	kind := func(s string) *Kind {
		k := Kind(s)
		return &k
	}

	t.Run("LocalObjectReference", func(t *testing.T) {
		ref := LocalObjectReference{Group: "foo", Kind: "bar"}
		assert.True(t, ref.HasKind(schema.GroupKind{Group: "foo", Kind: "bar"}))
	})

	t.Run("SecretObjectReference", func(t *testing.T) {
		ref := SecretObjectReference{Group: group("foo"), Kind: kind("bar")}
		assert.True(t, ref.HasKind(schema.GroupKind{Group: "foo", Kind: "bar"}))
	})

	t.Run("BackendObjectReference", func(t *testing.T) {
		ref := BackendObjectReference{Group: group("foo"), Kind: kind("bar")}
		assert.True(t, ref.HasKind(schema.GroupKind{Group: "foo", Kind: "bar"}))
	})

	t.Run("nil group matches empty", func(t *testing.T) {
		ref := BackendObjectReference{Kind: kind("bar")}
		assert.True(t, ref.HasKind(schema.GroupKind{Group: "", Kind: "bar"}))
	})

	t.Run("nil has an empty kind", func(t *testing.T) {
		ref := BackendObjectReference{}
		assert.True(t, ref.HasKind(schema.GroupKind{Group: "", Kind: ""}))
	})
}
