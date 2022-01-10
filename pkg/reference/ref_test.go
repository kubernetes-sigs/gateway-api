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

package reference

import (
	"testing"

	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	"sigs.k8s.io/gateway-api/apis/v1alpha2"
)

func TestNamespacedRefCreate(t *testing.T) {
	kind := v1alpha2.Kind("SomeKind")
	group := v1alpha2.Group("SomeGroup")
	name := v1alpha2.ObjectName("SomeName")
	namespace := v1alpha2.Namespace("SomeNamespace")

	check := func(t *testing.T, ref Ref) {
		kind := metav1.GroupKind{
			Group: string(group),
			Kind:  string(kind),
		}

		assert.Equal(t, string(name), ref.Name())
		assert.Equal(t, string(namespace), ref.Namespace())
		assert.Equal(t, kind, ref.Kind())
	}

	t.Run("secret", func(t *testing.T) {
		check(t, Secret(&v1alpha2.SecretObjectReference{
			Group:     &group,
			Kind:      &kind,
			Name:      name,
			Namespace: &namespace,
		}))
	})

	t.Run("parent", func(t *testing.T) {
		check(t, Parent(&v1alpha2.ParentReference{
			Group:     &group,
			Kind:      &kind,
			Name:      name,
			Namespace: &namespace,
		}))
	})

	t.Run("backend", func(t *testing.T) {
		check(t, Backend(&v1alpha2.BackendObjectReference{
			Group:     &group,
			Kind:      &kind,
			Name:      name,
			Namespace: &namespace,
		}))
	})
}

func TestRefCreate(t *testing.T) {
	kind := v1alpha2.Kind("SomeKind")
	group := v1alpha2.Group("SomeGroup")
	name := v1alpha2.ObjectName("SomeName")

	check := func(t *testing.T, ref Ref) {
		kind := metav1.GroupKind{
			Group: string(group),
			Kind:  string(kind),
		}

		assert.Equal(t, string(name), ref.Name())
		assert.Equal(t, string(""), ref.Namespace())
		assert.Equal(t, kind, ref.Kind())
	}

	t.Run("secret", func(t *testing.T) {
		check(t, Secret(&v1alpha2.SecretObjectReference{
			Group: &group,
			Kind:  &kind,
			Name:  name,
		}))
	})

	t.Run("parent", func(t *testing.T) {
		check(t, Parent(&v1alpha2.ParentReference{
			Group: &group,
			Kind:  &kind,
			Name:  name,
		}))
	})

	t.Run("backend", func(t *testing.T) {
		check(t, Backend(&v1alpha2.BackendObjectReference{
			Group: &group,
			Kind:  &kind,
			Name:  name,
		}))
	})

	t.Run("local", func(t *testing.T) {
		ref := Local(&v1alpha2.LocalObjectReference{
			Group: group,
			Kind:  kind,
			Name:  name,
		})

		kind := metav1.GroupKind{
			Group: string(group),
			Kind:  string(kind),
		}

		assert.Equal(t, string(name), ref.Name())
		assert.Equal(t, "", ref.Namespace())
		assert.Equal(t, kind, ref.Kind())
	})
}

func TestNamespacedRefNamespaceOf(t *testing.T) {
	kind := v1alpha2.Kind("SomeKind")
	group := v1alpha2.Group("SomeGroup")
	name := v1alpha2.ObjectName("SomeName")
	namespace := v1alpha2.Namespace("SomeNamespace")

	parent := &v1alpha2.Gateway{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "parent",
			Namespace: "gateway-test",
		},
	}

	check := func(t *testing.T, ref Ref) {
		assert.Equal(t, string(namespace), NamespaceOf(parent, ref))
	}

	t.Run("secret", func(t *testing.T) {
		check(t, Secret(&v1alpha2.SecretObjectReference{
			Group:     &group,
			Kind:      &kind,
			Name:      name,
			Namespace: &namespace,
		}))
	})

	t.Run("parent", func(t *testing.T) {
		check(t, Parent(&v1alpha2.ParentReference{
			Group:     &group,
			Kind:      &kind,
			Name:      name,
			Namespace: &namespace,
		}))
	})

	t.Run("backend", func(t *testing.T) {
		check(t, Backend(&v1alpha2.BackendObjectReference{
			Group:     &group,
			Kind:      &kind,
			Name:      name,
			Namespace: &namespace,
		}))
	})
}

func TestRefNamespaceOf(t *testing.T) {
	kind := v1alpha2.Kind("SomeKind")
	group := v1alpha2.Group("SomeGroup")
	name := v1alpha2.ObjectName("SomeName")

	parent := &v1alpha2.Gateway{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "parent",
			Namespace: "gateway-test",
		},
	}

	check := func(t *testing.T, ref Ref) {
		assert.Equal(t, "gateway-test", NamespaceOf(parent, ref))
	}

	t.Run("secret", func(t *testing.T) {
		check(t, Secret(&v1alpha2.SecretObjectReference{
			Group: &group,
			Kind:  &kind,
			Name:  name,
		}))
	})

	t.Run("parent", func(t *testing.T) {
		check(t, Parent(&v1alpha2.ParentReference{
			Group: &group,
			Kind:  &kind,
			Name:  name,
		}))
	})

	t.Run("backend", func(t *testing.T) {
		check(t, Backend(&v1alpha2.BackendObjectReference{
			Group: &group,
			Kind:  &kind,
			Name:  name,
		}))
	})
}

func TestNamespacedRefNamespacedName(t *testing.T) {
	kind := v1alpha2.Kind("SomeKind")
	group := v1alpha2.Group("SomeGroup")
	name := v1alpha2.ObjectName("SomeName")
	namespace := v1alpha2.Namespace("SomeNamespace")

	parent := &v1alpha2.Gateway{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "parent",
			Namespace: "gateway-test",
		},
	}

	check := func(t *testing.T, ref Ref) {
		n := types.NamespacedName{
			Namespace: string(namespace),
			Name:      string(name),
		}
		assert.Equal(t, n, NamespacedName(parent, ref))
	}

	t.Run("secret", func(t *testing.T) {
		check(t, Secret(&v1alpha2.SecretObjectReference{
			Group:     &group,
			Kind:      &kind,
			Name:      name,
			Namespace: &namespace,
		}))
	})

	t.Run("parent", func(t *testing.T) {
		check(t, Parent(&v1alpha2.ParentReference{
			Group:     &group,
			Kind:      &kind,
			Name:      name,
			Namespace: &namespace,
		}))
	})

	t.Run("backend", func(t *testing.T) {
		check(t, Backend(&v1alpha2.BackendObjectReference{
			Group:     &group,
			Kind:      &kind,
			Name:      name,
			Namespace: &namespace,
		}))
	})
}

func TestRefNamespacedName(t *testing.T) {
	kind := v1alpha2.Kind("SomeKind")
	group := v1alpha2.Group("SomeGroup")
	name := v1alpha2.ObjectName("SomeName")

	parent := &v1alpha2.Gateway{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "parent",
			Namespace: "gateway-test",
		},
	}

	check := func(t *testing.T, ref Ref) {
		n := types.NamespacedName{
			Namespace: "gateway-test",
			Name:      string(name),
		}
		assert.Equal(t, n, NamespacedName(parent, ref))
	}

	t.Run("secret", func(t *testing.T) {
		check(t, Secret(&v1alpha2.SecretObjectReference{
			Group: &group,
			Kind:  &kind,
			Name:  name,
		}))
	})

	t.Run("parent", func(t *testing.T) {
		check(t, Parent(&v1alpha2.ParentReference{
			Group: &group,
			Kind:  &kind,
			Name:  name,
		}))
	})

	t.Run("backend", func(t *testing.T) {
		check(t, Backend(&v1alpha2.BackendObjectReference{
			Group: &group,
			Kind:  &kind,
			Name:  name,
		}))
	})
}

func TestRefDefaultKind(t *testing.T) {
	check := func(group string, kind string, ref Ref) {
		gk := ref.Kind()
		assert.Equal(t, group, gk.Group)
		assert.Equal(t, kind, gk.Kind)
	}

	check("", "", Local(&v1alpha2.LocalObjectReference{}))
	check(string(v1alpha2.GroupName), "Gateway", Parent(&v1alpha2.ParentReference{}))
	check("", "Secret", Secret(&v1alpha2.SecretObjectReference{}))
	check("", "Service", Backend(&v1alpha2.BackendObjectReference{}))

	grp := v1alpha2.Group("G")
	check("G", "", Local(&v1alpha2.LocalObjectReference{Group: grp}))
	check("G", "Gateway", Parent(&v1alpha2.ParentReference{Group: &grp}))
	check("G", "Secret", Secret(&v1alpha2.SecretObjectReference{Group: &grp}))
	check("G", "Service", Backend(&v1alpha2.BackendObjectReference{Group: &grp}))

	knd := v1alpha2.Kind("K")
	check("", "K", Local(&v1alpha2.LocalObjectReference{Kind: knd}))
	check(string(v1alpha2.GroupName), "K", Parent(&v1alpha2.ParentReference{Kind: &knd}))
	check("", "K", Secret(&v1alpha2.SecretObjectReference{Kind: &knd}))
	check("", "K", Backend(&v1alpha2.BackendObjectReference{Kind: &knd}))

}
