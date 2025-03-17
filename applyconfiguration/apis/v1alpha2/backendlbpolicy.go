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

// Code generated by applyconfiguration-gen. DO NOT EDIT.

package v1alpha2

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	types "k8s.io/apimachinery/pkg/types"
	managedfields "k8s.io/apimachinery/pkg/util/managedfields"
	v1 "k8s.io/client-go/applyconfigurations/meta/v1"
	apisv1alpha2 "sigs.k8s.io/gateway-api/apis/v1alpha2"
	internal "sigs.k8s.io/gateway-api/applyconfiguration/internal"
)

// BackendLBPolicyApplyConfiguration represents a declarative configuration of the BackendLBPolicy type for use
// with apply.
type BackendLBPolicyApplyConfiguration struct {
	v1.TypeMetaApplyConfiguration    `json:",inline"`
	*v1.ObjectMetaApplyConfiguration `json:"metadata,omitempty"`
	Spec                             *BackendLBPolicySpecApplyConfiguration `json:"spec,omitempty"`
	Status                           *PolicyStatusApplyConfiguration        `json:"status,omitempty"`
}

// BackendLBPolicy constructs a declarative configuration of the BackendLBPolicy type for use with
// apply.
func BackendLBPolicy(name, namespace string) *BackendLBPolicyApplyConfiguration {
	b := &BackendLBPolicyApplyConfiguration{}
	b.WithName(name)
	b.WithNamespace(namespace)
	b.WithKind("BackendLBPolicy")
	b.WithAPIVersion("gateway.networking.k8s.io/v1alpha2")
	return b
}

// ExtractBackendLBPolicy extracts the applied configuration owned by fieldManager from
// backendLBPolicy. If no managedFields are found in backendLBPolicy for fieldManager, a
// BackendLBPolicyApplyConfiguration is returned with only the Name, Namespace (if applicable),
// APIVersion and Kind populated. It is possible that no managed fields were found for because other
// field managers have taken ownership of all the fields previously owned by fieldManager, or because
// the fieldManager never owned fields any fields.
// backendLBPolicy must be a unmodified BackendLBPolicy API object that was retrieved from the Kubernetes API.
// ExtractBackendLBPolicy provides a way to perform a extract/modify-in-place/apply workflow.
// Note that an extracted apply configuration will contain fewer fields than what the fieldManager previously
// applied if another fieldManager has updated or force applied any of the previously applied fields.
// Experimental!
func ExtractBackendLBPolicy(backendLBPolicy *apisv1alpha2.BackendLBPolicy, fieldManager string) (*BackendLBPolicyApplyConfiguration, error) {
	return extractBackendLBPolicy(backendLBPolicy, fieldManager, "")
}

// ExtractBackendLBPolicyStatus is the same as ExtractBackendLBPolicy except
// that it extracts the status subresource applied configuration.
// Experimental!
func ExtractBackendLBPolicyStatus(backendLBPolicy *apisv1alpha2.BackendLBPolicy, fieldManager string) (*BackendLBPolicyApplyConfiguration, error) {
	return extractBackendLBPolicy(backendLBPolicy, fieldManager, "status")
}

func extractBackendLBPolicy(backendLBPolicy *apisv1alpha2.BackendLBPolicy, fieldManager string, subresource string) (*BackendLBPolicyApplyConfiguration, error) {
	b := &BackendLBPolicyApplyConfiguration{}
	err := managedfields.ExtractInto(backendLBPolicy, internal.Parser().Type("io.k8s.sigs.gateway-api.apis.v1alpha2.BackendLBPolicy"), fieldManager, b, subresource)
	if err != nil {
		return nil, err
	}
	b.WithName(backendLBPolicy.Name)
	b.WithNamespace(backendLBPolicy.Namespace)

	b.WithKind("BackendLBPolicy")
	b.WithAPIVersion("gateway.networking.k8s.io/v1alpha2")
	return b, nil
}

// WithKind sets the Kind field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the Kind field is set to the value of the last call.
func (b *BackendLBPolicyApplyConfiguration) WithKind(value string) *BackendLBPolicyApplyConfiguration {
	b.TypeMetaApplyConfiguration.Kind = &value
	return b
}

// WithAPIVersion sets the APIVersion field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the APIVersion field is set to the value of the last call.
func (b *BackendLBPolicyApplyConfiguration) WithAPIVersion(value string) *BackendLBPolicyApplyConfiguration {
	b.TypeMetaApplyConfiguration.APIVersion = &value
	return b
}

// WithName sets the Name field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the Name field is set to the value of the last call.
func (b *BackendLBPolicyApplyConfiguration) WithName(value string) *BackendLBPolicyApplyConfiguration {
	b.ensureObjectMetaApplyConfigurationExists()
	b.ObjectMetaApplyConfiguration.Name = &value
	return b
}

// WithGenerateName sets the GenerateName field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the GenerateName field is set to the value of the last call.
func (b *BackendLBPolicyApplyConfiguration) WithGenerateName(value string) *BackendLBPolicyApplyConfiguration {
	b.ensureObjectMetaApplyConfigurationExists()
	b.ObjectMetaApplyConfiguration.GenerateName = &value
	return b
}

// WithNamespace sets the Namespace field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the Namespace field is set to the value of the last call.
func (b *BackendLBPolicyApplyConfiguration) WithNamespace(value string) *BackendLBPolicyApplyConfiguration {
	b.ensureObjectMetaApplyConfigurationExists()
	b.ObjectMetaApplyConfiguration.Namespace = &value
	return b
}

// WithUID sets the UID field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the UID field is set to the value of the last call.
func (b *BackendLBPolicyApplyConfiguration) WithUID(value types.UID) *BackendLBPolicyApplyConfiguration {
	b.ensureObjectMetaApplyConfigurationExists()
	b.ObjectMetaApplyConfiguration.UID = &value
	return b
}

// WithResourceVersion sets the ResourceVersion field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the ResourceVersion field is set to the value of the last call.
func (b *BackendLBPolicyApplyConfiguration) WithResourceVersion(value string) *BackendLBPolicyApplyConfiguration {
	b.ensureObjectMetaApplyConfigurationExists()
	b.ObjectMetaApplyConfiguration.ResourceVersion = &value
	return b
}

// WithGeneration sets the Generation field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the Generation field is set to the value of the last call.
func (b *BackendLBPolicyApplyConfiguration) WithGeneration(value int64) *BackendLBPolicyApplyConfiguration {
	b.ensureObjectMetaApplyConfigurationExists()
	b.ObjectMetaApplyConfiguration.Generation = &value
	return b
}

// WithCreationTimestamp sets the CreationTimestamp field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the CreationTimestamp field is set to the value of the last call.
func (b *BackendLBPolicyApplyConfiguration) WithCreationTimestamp(value metav1.Time) *BackendLBPolicyApplyConfiguration {
	b.ensureObjectMetaApplyConfigurationExists()
	b.ObjectMetaApplyConfiguration.CreationTimestamp = &value
	return b
}

// WithDeletionTimestamp sets the DeletionTimestamp field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the DeletionTimestamp field is set to the value of the last call.
func (b *BackendLBPolicyApplyConfiguration) WithDeletionTimestamp(value metav1.Time) *BackendLBPolicyApplyConfiguration {
	b.ensureObjectMetaApplyConfigurationExists()
	b.ObjectMetaApplyConfiguration.DeletionTimestamp = &value
	return b
}

// WithDeletionGracePeriodSeconds sets the DeletionGracePeriodSeconds field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the DeletionGracePeriodSeconds field is set to the value of the last call.
func (b *BackendLBPolicyApplyConfiguration) WithDeletionGracePeriodSeconds(value int64) *BackendLBPolicyApplyConfiguration {
	b.ensureObjectMetaApplyConfigurationExists()
	b.ObjectMetaApplyConfiguration.DeletionGracePeriodSeconds = &value
	return b
}

// WithLabels puts the entries into the Labels field in the declarative configuration
// and returns the receiver, so that objects can be build by chaining "With" function invocations.
// If called multiple times, the entries provided by each call will be put on the Labels field,
// overwriting an existing map entries in Labels field with the same key.
func (b *BackendLBPolicyApplyConfiguration) WithLabels(entries map[string]string) *BackendLBPolicyApplyConfiguration {
	b.ensureObjectMetaApplyConfigurationExists()
	if b.ObjectMetaApplyConfiguration.Labels == nil && len(entries) > 0 {
		b.ObjectMetaApplyConfiguration.Labels = make(map[string]string, len(entries))
	}
	for k, v := range entries {
		b.ObjectMetaApplyConfiguration.Labels[k] = v
	}
	return b
}

// WithAnnotations puts the entries into the Annotations field in the declarative configuration
// and returns the receiver, so that objects can be build by chaining "With" function invocations.
// If called multiple times, the entries provided by each call will be put on the Annotations field,
// overwriting an existing map entries in Annotations field with the same key.
func (b *BackendLBPolicyApplyConfiguration) WithAnnotations(entries map[string]string) *BackendLBPolicyApplyConfiguration {
	b.ensureObjectMetaApplyConfigurationExists()
	if b.ObjectMetaApplyConfiguration.Annotations == nil && len(entries) > 0 {
		b.ObjectMetaApplyConfiguration.Annotations = make(map[string]string, len(entries))
	}
	for k, v := range entries {
		b.ObjectMetaApplyConfiguration.Annotations[k] = v
	}
	return b
}

// WithOwnerReferences adds the given value to the OwnerReferences field in the declarative configuration
// and returns the receiver, so that objects can be build by chaining "With" function invocations.
// If called multiple times, values provided by each call will be appended to the OwnerReferences field.
func (b *BackendLBPolicyApplyConfiguration) WithOwnerReferences(values ...*v1.OwnerReferenceApplyConfiguration) *BackendLBPolicyApplyConfiguration {
	b.ensureObjectMetaApplyConfigurationExists()
	for i := range values {
		if values[i] == nil {
			panic("nil value passed to WithOwnerReferences")
		}
		b.ObjectMetaApplyConfiguration.OwnerReferences = append(b.ObjectMetaApplyConfiguration.OwnerReferences, *values[i])
	}
	return b
}

// WithFinalizers adds the given value to the Finalizers field in the declarative configuration
// and returns the receiver, so that objects can be build by chaining "With" function invocations.
// If called multiple times, values provided by each call will be appended to the Finalizers field.
func (b *BackendLBPolicyApplyConfiguration) WithFinalizers(values ...string) *BackendLBPolicyApplyConfiguration {
	b.ensureObjectMetaApplyConfigurationExists()
	for i := range values {
		b.ObjectMetaApplyConfiguration.Finalizers = append(b.ObjectMetaApplyConfiguration.Finalizers, values[i])
	}
	return b
}

func (b *BackendLBPolicyApplyConfiguration) ensureObjectMetaApplyConfigurationExists() {
	if b.ObjectMetaApplyConfiguration == nil {
		b.ObjectMetaApplyConfiguration = &v1.ObjectMetaApplyConfiguration{}
	}
}

// WithSpec sets the Spec field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the Spec field is set to the value of the last call.
func (b *BackendLBPolicyApplyConfiguration) WithSpec(value *BackendLBPolicySpecApplyConfiguration) *BackendLBPolicyApplyConfiguration {
	b.Spec = value
	return b
}

// WithStatus sets the Status field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the Status field is set to the value of the last call.
func (b *BackendLBPolicyApplyConfiguration) WithStatus(value *PolicyStatusApplyConfiguration) *BackendLBPolicyApplyConfiguration {
	b.Status = value
	return b
}

// GetName retrieves the value of the Name field in the declarative configuration.
func (b *BackendLBPolicyApplyConfiguration) GetName() *string {
	b.ensureObjectMetaApplyConfigurationExists()
	return b.ObjectMetaApplyConfiguration.Name
}
