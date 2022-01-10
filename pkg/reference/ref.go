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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	"sigs.k8s.io/gateway-api/apis/v1alpha2"
)

var (
	// DefaultParentKind is the default metav1.GroupKind for ParentReferences.
	DefaultParentKind = metav1.GroupKind{
		Group: v1alpha2.GroupName,
		Kind:  "Gateway",
	}

	// DefaultSecretKind is the default metav1.GroupKind for SecretObjectReferences.
	DefaultSecretKind = metav1.GroupKind{
		Group: "",
		Kind:  "Secret",
	}

	// DefaultBackendObjectKind is the default metav1.GroupKind for BackendObjectReferences.
	DefaultBackendObjectKind = metav1.GroupKind{
		Group: "",
		Kind:  "Service",
	}
)

func referenceKind(gk metav1.GroupKind, g *v1alpha2.Group, k *v1alpha2.Kind) metav1.GroupKind {
	if g != nil {
		gk.Group = string(*g)
	}

	if k != nil {
		gk.Kind = string(*k)
	}

	return gk
}

func referenceNamespace(ns *v1alpha2.Namespace) string {
	if ns != nil {
		return string(*ns)
	}

	return ""
}

// Ref is an interface that represents a reference to a resource.
type Ref interface {
	Name() string
	Namespace() string
	Kind() metav1.GroupKind
}

// Local returns a Ref for the LocalObjectReference ref.
func Local(ref *v1alpha2.LocalObjectReference) Ref {
	return &localRef{ref}
}

type localRef struct {
	ref *v1alpha2.LocalObjectReference
}

func (l *localRef) Name() string {
	return string(l.ref.Name)
}

func (l *localRef) Namespace() string {
	return ""
}

func (l *localRef) Kind() metav1.GroupKind {
	return metav1.GroupKind{Group: string(l.ref.Group), Kind: string(l.ref.Kind)}
}

// Parent returns a Ref for the ParentReference ref.
func Parent(ref *v1alpha2.ParentReference) Ref {
	return &parentRef{ref}
}

type parentRef struct {
	ref *v1alpha2.ParentReference
}

func (p *parentRef) Name() string {
	return string(p.ref.Name)
}

func (p *parentRef) Namespace() string {
	return referenceNamespace(p.ref.Namespace)
}

func (p *parentRef) Kind() metav1.GroupKind {
	return referenceKind(DefaultParentKind, p.ref.Group, p.ref.Kind)
}

// Secret returns a Ref for the SecretObjectReference ref.
func Secret(ref *v1alpha2.SecretObjectReference) Ref {
	return &secretRef{ref}
}

type secretRef struct {
	ref *v1alpha2.SecretObjectReference
}

func (s *secretRef) Name() string {
	return string(s.ref.Name)
}

func (s *secretRef) Namespace() string {
	return referenceNamespace(s.ref.Namespace)
}

func (s *secretRef) Kind() metav1.GroupKind {
	return referenceKind(DefaultSecretKind, s.ref.Group, s.ref.Kind)
}

// Backend returns a Ref for the BackendObjectReference ref.
func Backend(ref *v1alpha2.BackendObjectReference) Ref {
	return &backendRef{ref}
}

type backendRef struct {
	ref *v1alpha2.BackendObjectReference
}

func (s *backendRef) Name() string {
	return string(s.ref.Name)
}

func (s *backendRef) Namespace() string {
	return referenceNamespace(s.ref.Namespace)
}

func (s *backendRef) Kind() metav1.GroupKind {
	return referenceKind(DefaultBackendObjectKind, s.ref.Group, s.ref.Kind)
}

// NamespaceOf returns the namespace of the Gateway API object reference
// given by ref when it is referenced from parent. If ref does not specify
// a namespace, then the namespace of parent is returned.
func NamespaceOf(parent metav1.Object, ref Ref) string {
	if ns := ref.Namespace(); ns != "" {
		return ns
	}

	return parent.GetNamespace()
}

// NamespacedName returns the fully qualified object name of ref when
// it is referenced by the object parent. If ref does not specify
// a namespace, then the namespace of parent is used.
func NamespacedName(parent metav1.Object, ref Ref) types.NamespacedName {
	return types.NamespacedName{
		Namespace: NamespaceOf(parent, ref),
		Name:      ref.Name(),
	}
}
