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

package resourcediscovery

import (
	"fmt"

	"sigs.k8s.io/gateway-api/gwctl/pkg/common"
)

type ReferenceToNonExistentResourceError struct {
	ReferenceFromTo
}

func (r ReferenceToNonExistentResourceError) Error() string {
	return fmt.Sprintf("%v %q references a non-existent %v %q",
		r.referringObjectKind(), r.referringObjectName(),
		r.referredObjectKind(), r.referredObjectName())
}

type ReferenceNotPermittedError struct {
	ReferenceFromTo
}

func (r ReferenceNotPermittedError) Error() string {
	return fmt.Sprintf("%v %q is not permitted to reference %v %q",
		r.referringObjectKind(), r.referringObjectName(),
		r.referredObjectKind(), r.referredObjectName())
}

type ReferenceFromTo struct {
	// ReferringObject is the "from" object which is referring "to" some other
	// object.
	ReferringObject common.ObjRef
	// ReferredObject is the actual object which is being referenced by another
	// object.
	ReferredObject common.ObjRef
}

// referringObjectKind returns a human readable Kind.
func (r ReferenceFromTo) referringObjectKind() string {
	if r.ReferringObject.Group != "" {
		return fmt.Sprintf("%v(.%v)", r.ReferringObject.Kind, r.ReferringObject.Group)
	}
	return r.ReferringObject.Kind
}

// referredObjectKind returns a human readable Kind.
func (r ReferenceFromTo) referredObjectKind() string {
	if r.ReferredObject.Group != "" {
		return fmt.Sprintf("%v(.%v)", r.ReferredObject.Kind, r.ReferredObject.Group)
	}
	return r.ReferredObject.Kind
}

// referringObjectName returns a human readable Name.
func (r ReferenceFromTo) referringObjectName() string {
	if r.ReferringObject.Namespace != "" {
		return fmt.Sprintf("%v/%v", r.ReferringObject.Namespace, r.ReferringObject.Name)
	}
	return r.ReferringObject.Name
}

// referredObjectName returns a human readable Name.
func (r ReferenceFromTo) referredObjectName() string {
	if r.ReferredObject.Namespace != "" {
		return fmt.Sprintf("%v/%v", r.ReferredObject.Namespace, r.ReferredObject.Name)
	}
	return r.ReferredObject.Name
}
