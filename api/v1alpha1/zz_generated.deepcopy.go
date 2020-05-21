// +build !ignore_autogenerated

/*

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

// Code generated by controller-gen. DO NOT EDIT.

package v1alpha1

import (
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	runtime "k8s.io/apimachinery/pkg/runtime"
)

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ConfigMapsDefaultLocalObjectReference) DeepCopyInto(out *ConfigMapsDefaultLocalObjectReference) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ConfigMapsDefaultLocalObjectReference.
func (in *ConfigMapsDefaultLocalObjectReference) DeepCopy() *ConfigMapsDefaultLocalObjectReference {
	if in == nil {
		return nil
	}
	out := new(ConfigMapsDefaultLocalObjectReference)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ForwardToTarget) DeepCopyInto(out *ForwardToTarget) {
	*out = *in
	out.TargetRef = in.TargetRef
	if in.TargetPort != nil {
		in, out := &in.TargetPort, &out.TargetPort
		*out = new(TargetPort)
		**out = **in
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ForwardToTarget.
func (in *ForwardToTarget) DeepCopy() *ForwardToTarget {
	if in == nil {
		return nil
	}
	out := new(ForwardToTarget)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Gateway) DeepCopyInto(out *Gateway) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	in.Status.DeepCopyInto(&out.Status)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Gateway.
func (in *Gateway) DeepCopy() *Gateway {
	if in == nil {
		return nil
	}
	out := new(Gateway)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *Gateway) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *GatewayClass) DeepCopyInto(out *GatewayClass) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	in.Status.DeepCopyInto(&out.Status)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new GatewayClass.
func (in *GatewayClass) DeepCopy() *GatewayClass {
	if in == nil {
		return nil
	}
	out := new(GatewayClass)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *GatewayClass) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *GatewayClassCondition) DeepCopyInto(out *GatewayClassCondition) {
	*out = *in
	in.LastTransitionTime.DeepCopyInto(&out.LastTransitionTime)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new GatewayClassCondition.
func (in *GatewayClassCondition) DeepCopy() *GatewayClassCondition {
	if in == nil {
		return nil
	}
	out := new(GatewayClassCondition)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *GatewayClassList) DeepCopyInto(out *GatewayClassList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]GatewayClass, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new GatewayClassList.
func (in *GatewayClassList) DeepCopy() *GatewayClassList {
	if in == nil {
		return nil
	}
	out := new(GatewayClassList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *GatewayClassList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *GatewayClassSpec) DeepCopyInto(out *GatewayClassSpec) {
	*out = *in
	in.AllowedGatewayNamespaces.DeepCopyInto(&out.AllowedGatewayNamespaces)
	if in.AllowedRouteNamespaces != nil {
		in, out := &in.AllowedRouteNamespaces, &out.AllowedRouteNamespaces
		*out = new(v1.LabelSelector)
		(*in).DeepCopyInto(*out)
	}
	if in.ParametersRef != nil {
		in, out := &in.ParametersRef, &out.ParametersRef
		*out = new(ConfigMapsDefaultLocalObjectReference)
		**out = **in
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new GatewayClassSpec.
func (in *GatewayClassSpec) DeepCopy() *GatewayClassSpec {
	if in == nil {
		return nil
	}
	out := new(GatewayClassSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *GatewayClassStatus) DeepCopyInto(out *GatewayClassStatus) {
	*out = *in
	if in.Conditions != nil {
		in, out := &in.Conditions, &out.Conditions
		*out = make([]GatewayClassCondition, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new GatewayClassStatus.
func (in *GatewayClassStatus) DeepCopy() *GatewayClassStatus {
	if in == nil {
		return nil
	}
	out := new(GatewayClassStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *GatewayCondition) DeepCopyInto(out *GatewayCondition) {
	*out = *in
	in.LastTransitionTime.DeepCopyInto(&out.LastTransitionTime)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new GatewayCondition.
func (in *GatewayCondition) DeepCopy() *GatewayCondition {
	if in == nil {
		return nil
	}
	out := new(GatewayCondition)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *GatewayList) DeepCopyInto(out *GatewayList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]Gateway, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new GatewayList.
func (in *GatewayList) DeepCopy() *GatewayList {
	if in == nil {
		return nil
	}
	out := new(GatewayList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *GatewayList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *GatewayObjectReference) DeepCopyInto(out *GatewayObjectReference) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new GatewayObjectReference.
func (in *GatewayObjectReference) DeepCopy() *GatewayObjectReference {
	if in == nil {
		return nil
	}
	out := new(GatewayObjectReference)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *GatewaySpec) DeepCopyInto(out *GatewaySpec) {
	*out = *in
	if in.Listeners != nil {
		in, out := &in.Listeners, &out.Listeners
		*out = make([]Listener, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	in.Routes.DeepCopyInto(&out.Routes)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new GatewaySpec.
func (in *GatewaySpec) DeepCopy() *GatewaySpec {
	if in == nil {
		return nil
	}
	out := new(GatewaySpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *GatewayStatus) DeepCopyInto(out *GatewayStatus) {
	*out = *in
	if in.Conditions != nil {
		in, out := &in.Conditions, &out.Conditions
		*out = make([]GatewayCondition, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	if in.Listeners != nil {
		in, out := &in.Listeners, &out.Listeners
		*out = make([]ListenerStatus, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new GatewayStatus.
func (in *GatewayStatus) DeepCopy() *GatewayStatus {
	if in == nil {
		return nil
	}
	out := new(GatewayStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *HTTPHeaderModifier) DeepCopyInto(out *HTTPHeaderModifier) {
	*out = *in
	if in.Add != nil {
		in, out := &in.Add, &out.Add
		*out = make(map[string]string, len(*in))
		for key, val := range *in {
			(*out)[key] = val
		}
	}
	if in.Remove != nil {
		in, out := &in.Remove, &out.Remove
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new HTTPHeaderModifier.
func (in *HTTPHeaderModifier) DeepCopy() *HTTPHeaderModifier {
	if in == nil {
		return nil
	}
	out := new(HTTPHeaderModifier)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *HTTPRequestAction) DeepCopyInto(out *HTTPRequestAction) {
	*out = *in
	if in.ForwardTo != nil {
		in, out := &in.ForwardTo, &out.ForwardTo
		*out = new(ForwardToTarget)
		(*in).DeepCopyInto(*out)
	}
	if in.Modify != nil {
		in, out := &in.Modify, &out.Modify
		*out = new(HTTPRequestModifier)
		(*in).DeepCopyInto(*out)
	}
	if in.ExtensionRef != nil {
		in, out := &in.ExtensionRef, &out.ExtensionRef
		*out = new(ConfigMapsDefaultLocalObjectReference)
		**out = **in
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new HTTPRequestAction.
func (in *HTTPRequestAction) DeepCopy() *HTTPRequestAction {
	if in == nil {
		return nil
	}
	out := new(HTTPRequestAction)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *HTTPRequestMatch) DeepCopyInto(out *HTTPRequestMatch) {
	*out = *in
	if in.Path != nil {
		in, out := &in.Path, &out.Path
		*out = new(string)
		**out = **in
	}
	if in.HeaderType != nil {
		in, out := &in.HeaderType, &out.HeaderType
		*out = new(string)
		**out = **in
	}
	if in.Header != nil {
		in, out := &in.Header, &out.Header
		*out = make(map[string]string, len(*in))
		for key, val := range *in {
			(*out)[key] = val
		}
	}
	if in.ExtensionRef != nil {
		in, out := &in.ExtensionRef, &out.ExtensionRef
		*out = new(ConfigMapsDefaultLocalObjectReference)
		**out = **in
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new HTTPRequestMatch.
func (in *HTTPRequestMatch) DeepCopy() *HTTPRequestMatch {
	if in == nil {
		return nil
	}
	out := new(HTTPRequestMatch)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *HTTPRequestModifier) DeepCopyInto(out *HTTPRequestModifier) {
	*out = *in
	if in.Headers != nil {
		in, out := &in.Headers, &out.Headers
		*out = new(HTTPHeaderModifier)
		(*in).DeepCopyInto(*out)
	}
	if in.ExtensionRef != nil {
		in, out := &in.ExtensionRef, &out.ExtensionRef
		*out = new(ConfigMapsDefaultLocalObjectReference)
		**out = **in
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new HTTPRequestModifier.
func (in *HTTPRequestModifier) DeepCopy() *HTTPRequestModifier {
	if in == nil {
		return nil
	}
	out := new(HTTPRequestModifier)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *HTTPRoute) DeepCopyInto(out *HTTPRoute) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	in.Status.DeepCopyInto(&out.Status)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new HTTPRoute.
func (in *HTTPRoute) DeepCopy() *HTTPRoute {
	if in == nil {
		return nil
	}
	out := new(HTTPRoute)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *HTTPRoute) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *HTTPRouteHost) DeepCopyInto(out *HTTPRouteHost) {
	*out = *in
	if in.Rules != nil {
		in, out := &in.Rules, &out.Rules
		*out = make([]HTTPRouteRule, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	if in.ExtensionRef != nil {
		in, out := &in.ExtensionRef, &out.ExtensionRef
		*out = new(ConfigMapsDefaultLocalObjectReference)
		**out = **in
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new HTTPRouteHost.
func (in *HTTPRouteHost) DeepCopy() *HTTPRouteHost {
	if in == nil {
		return nil
	}
	out := new(HTTPRouteHost)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *HTTPRouteList) DeepCopyInto(out *HTTPRouteList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]HTTPRoute, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new HTTPRouteList.
func (in *HTTPRouteList) DeepCopy() *HTTPRouteList {
	if in == nil {
		return nil
	}
	out := new(HTTPRouteList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *HTTPRouteList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *HTTPRouteRule) DeepCopyInto(out *HTTPRouteRule) {
	*out = *in
	if in.Match != nil {
		in, out := &in.Match, &out.Match
		*out = new(HTTPRequestMatch)
		(*in).DeepCopyInto(*out)
	}
	if in.Actions != nil {
		in, out := &in.Actions, &out.Actions
		*out = make([]HTTPRequestAction, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new HTTPRouteRule.
func (in *HTTPRouteRule) DeepCopy() *HTTPRouteRule {
	if in == nil {
		return nil
	}
	out := new(HTTPRouteRule)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *HTTPRouteSpec) DeepCopyInto(out *HTTPRouteSpec) {
	*out = *in
	if in.Hosts != nil {
		in, out := &in.Hosts, &out.Hosts
		*out = make([]HTTPRouteHost, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	if in.Default != nil {
		in, out := &in.Default, &out.Default
		*out = new(HTTPRouteHost)
		(*in).DeepCopyInto(*out)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new HTTPRouteSpec.
func (in *HTTPRouteSpec) DeepCopy() *HTTPRouteSpec {
	if in == nil {
		return nil
	}
	out := new(HTTPRouteSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *HTTPRouteStatus) DeepCopyInto(out *HTTPRouteStatus) {
	*out = *in
	if in.GatewayRefs != nil {
		in, out := &in.GatewayRefs, &out.GatewayRefs
		*out = make([]GatewayObjectReference, len(*in))
		copy(*out, *in)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new HTTPRouteStatus.
func (in *HTTPRouteStatus) DeepCopy() *HTTPRouteStatus {
	if in == nil {
		return nil
	}
	out := new(HTTPRouteStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Listener) DeepCopyInto(out *Listener) {
	*out = *in
	if in.Address != nil {
		in, out := &in.Address, &out.Address
		*out = new(ListenerAddress)
		**out = **in
	}
	if in.Port != nil {
		in, out := &in.Port, &out.Port
		*out = new(int32)
		**out = **in
	}
	if in.Protocol != nil {
		in, out := &in.Protocol, &out.Protocol
		*out = new(string)
		**out = **in
	}
	if in.TLS != nil {
		in, out := &in.TLS, &out.TLS
		*out = new(TLSConfig)
		(*in).DeepCopyInto(*out)
	}
	if in.ExtensionRef != nil {
		in, out := &in.ExtensionRef, &out.ExtensionRef
		*out = new(ConfigMapsDefaultLocalObjectReference)
		**out = **in
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Listener.
func (in *Listener) DeepCopy() *Listener {
	if in == nil {
		return nil
	}
	out := new(Listener)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ListenerAddress) DeepCopyInto(out *ListenerAddress) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ListenerAddress.
func (in *ListenerAddress) DeepCopy() *ListenerAddress {
	if in == nil {
		return nil
	}
	out := new(ListenerAddress)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ListenerCondition) DeepCopyInto(out *ListenerCondition) {
	*out = *in
	in.LastTransitionTime.DeepCopyInto(&out.LastTransitionTime)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ListenerCondition.
func (in *ListenerCondition) DeepCopy() *ListenerCondition {
	if in == nil {
		return nil
	}
	out := new(ListenerCondition)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ListenerStatus) DeepCopyInto(out *ListenerStatus) {
	*out = *in
	if in.Address != nil {
		in, out := &in.Address, &out.Address
		*out = new(ListenerAddress)
		**out = **in
	}
	if in.Conditions != nil {
		in, out := &in.Conditions, &out.Conditions
		*out = make([]ListenerCondition, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ListenerStatus.
func (in *ListenerStatus) DeepCopy() *ListenerStatus {
	if in == nil {
		return nil
	}
	out := new(ListenerStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *RouteBindingSelector) DeepCopyInto(out *RouteBindingSelector) {
	*out = *in
	in.NamespaceSelector.DeepCopyInto(&out.NamespaceSelector)
	if in.RouteSelector != nil {
		in, out := &in.RouteSelector, &out.RouteSelector
		*out = new(v1.LabelSelector)
		(*in).DeepCopyInto(*out)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new RouteBindingSelector.
func (in *RouteBindingSelector) DeepCopy() *RouteBindingSelector {
	if in == nil {
		return nil
	}
	out := new(RouteBindingSelector)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *SecretsDefaultLocalObjectReference) DeepCopyInto(out *SecretsDefaultLocalObjectReference) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new SecretsDefaultLocalObjectReference.
func (in *SecretsDefaultLocalObjectReference) DeepCopy() *SecretsDefaultLocalObjectReference {
	if in == nil {
		return nil
	}
	out := new(SecretsDefaultLocalObjectReference)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ServicesDefaultLocalObjectReference) DeepCopyInto(out *ServicesDefaultLocalObjectReference) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ServicesDefaultLocalObjectReference.
func (in *ServicesDefaultLocalObjectReference) DeepCopy() *ServicesDefaultLocalObjectReference {
	if in == nil {
		return nil
	}
	out := new(ServicesDefaultLocalObjectReference)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *TLSConfig) DeepCopyInto(out *TLSConfig) {
	*out = *in
	if in.CertificateRefs != nil {
		in, out := &in.CertificateRefs, &out.CertificateRefs
		*out = make([]SecretsDefaultLocalObjectReference, len(*in))
		copy(*out, *in)
	}
	if in.MinimumVersion != nil {
		in, out := &in.MinimumVersion, &out.MinimumVersion
		*out = new(string)
		**out = **in
	}
	if in.Options != nil {
		in, out := &in.Options, &out.Options
		*out = make(map[string]string, len(*in))
		for key, val := range *in {
			(*out)[key] = val
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new TLSConfig.
func (in *TLSConfig) DeepCopy() *TLSConfig {
	if in == nil {
		return nil
	}
	out := new(TLSConfig)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *TcpRoute) DeepCopyInto(out *TcpRoute) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	out.Spec = in.Spec
	out.Status = in.Status
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new TcpRoute.
func (in *TcpRoute) DeepCopy() *TcpRoute {
	if in == nil {
		return nil
	}
	out := new(TcpRoute)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *TcpRoute) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *TcpRouteList) DeepCopyInto(out *TcpRouteList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]TcpRoute, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new TcpRouteList.
func (in *TcpRouteList) DeepCopy() *TcpRouteList {
	if in == nil {
		return nil
	}
	out := new(TcpRouteList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *TcpRouteList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *TcpRouteSpec) DeepCopyInto(out *TcpRouteSpec) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new TcpRouteSpec.
func (in *TcpRouteSpec) DeepCopy() *TcpRouteSpec {
	if in == nil {
		return nil
	}
	out := new(TcpRouteSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *TcpRouteStatus) DeepCopyInto(out *TcpRouteStatus) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new TcpRouteStatus.
func (in *TcpRouteStatus) DeepCopy() *TcpRouteStatus {
	if in == nil {
		return nil
	}
	out := new(TcpRouteStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *TrafficSplit) DeepCopyInto(out *TrafficSplit) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	out.Spec = in.Spec
	out.Status = in.Status
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new TrafficSplit.
func (in *TrafficSplit) DeepCopy() *TrafficSplit {
	if in == nil {
		return nil
	}
	out := new(TrafficSplit)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *TrafficSplit) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *TrafficSplitList) DeepCopyInto(out *TrafficSplitList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]TrafficSplit, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new TrafficSplitList.
func (in *TrafficSplitList) DeepCopy() *TrafficSplitList {
	if in == nil {
		return nil
	}
	out := new(TrafficSplitList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *TrafficSplitList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *TrafficSplitSpec) DeepCopyInto(out *TrafficSplitSpec) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new TrafficSplitSpec.
func (in *TrafficSplitSpec) DeepCopy() *TrafficSplitSpec {
	if in == nil {
		return nil
	}
	out := new(TrafficSplitSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *TrafficSplitStatus) DeepCopyInto(out *TrafficSplitStatus) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new TrafficSplitStatus.
func (in *TrafficSplitStatus) DeepCopy() *TrafficSplitStatus {
	if in == nil {
		return nil
	}
	out := new(TrafficSplitStatus)
	in.DeepCopyInto(out)
	return out
}
