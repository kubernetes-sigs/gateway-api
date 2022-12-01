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

package translate

import (
	gatewayv1b1 "sigs.k8s.io/gateway-api/apis/v1beta1"
)

// PortNumberPtr translates an int to a *PortNumber
func PortNumberPtr(p int) *gatewayv1b1.PortNumber {
	result := gatewayv1b1.PortNumber(p)
	return &result
}

// PortNumberInt32 translates reference value of ptr to Int32
func PortNumberInt32(name *gatewayv1b1.PortNumber) int32 {
	portNum := int32(*name)
	return portNum
}

// SectionNamePtr translates a string to a *SectionName
func SectionNamePtr(sectionName string) *gatewayv1b1.SectionName {
	gwSectionName := gatewayv1b1.SectionName(sectionName)
	return &gwSectionName
}

// SectionNameStr translates reference value of ptr to string
func SectionNameStr(name *gatewayv1b1.SectionName) string {
	sectionName := string(*name)
	return sectionName
}

// HostnamePtr translates a string to a *Hostname
func HostnamePtr(host string) *gatewayv1b1.Hostname {
	h := gatewayv1b1.Hostname(host)
	return &h
}

// HostnameStr translates reference value of ptr to string
func HostnameStr(name *gatewayv1b1.Hostname) string {
	hostName := string(*name)
	return hostName
}

// PreciseHostnamePtr translates a string to a *PreciseHostname
func PreciseHostnamePtr(host string) *gatewayv1b1.PreciseHostname {
	h := gatewayv1b1.PreciseHostname(host)
	return &h
}

// PreciseHostnameStr translates reference value of ptr to string
func PreciseHostnameStr(name *gatewayv1b1.PreciseHostname) string {
	prechostName := string(*name)
	return prechostName
}

// GroupPtr translates a string to a *Group
func GroupPtr(group string) *gatewayv1b1.Group {
	gwGroup := gatewayv1b1.Group(group)
	return &gwGroup
}

// GroupStr translates reference value of ptr to string
func GroupStr(name *gatewayv1b1.Group) string {
	groupStr := string(*name)
	return groupStr
}

// KindPtr translates a string to a *Kind
func KindPtr(kind string) *gatewayv1b1.Kind {
	gwKind := gatewayv1b1.Kind(kind)
	return &gwKind
}

// KindStr translates reference value of ptr to string
func KindStr(name *gatewayv1b1.Kind) string {
	kindStr := string(*name)
	return kindStr
}

// NamespacePtr translates a string to a *Namespace
func NamespacePtr(namespace string) *gatewayv1b1.Namespace {
	gwNamespace := gatewayv1b1.Namespace(namespace)
	return &gwNamespace
}

// NamespaceStr translates reference value of ptr to string
func NamespaceStr(name *gatewayv1b1.Namespace) string {
	namespace := string(*name)
	return namespace
}

// ObjectNamePtr translates a string to a *ObjectName
func ObjectNamePtr(name string) *gatewayv1b1.ObjectName {
	objectName := gatewayv1b1.ObjectName(name)
	return &objectName
}

// ObjectNameStr translates reference value of ptr to string
func ObjectNameStr(name *gatewayv1b1.ObjectName) string {
	objname := string(*name)
	return objname
}

// GatewayControllerPtr translates a string to a *GatewayController
func GatewayControllerPtr(name string) *gatewayv1b1.GatewayController {
	gwCtrl := gatewayv1b1.GatewayController(name)
	return &gwCtrl
}

// GatewayControllerStr translates reference value of ptr to string
func GatewayControllerStr(name *gatewayv1b1.GatewayController) string {
	gw := string(*name)
	return gw
}

// AnnotationKeyPtr translates a string to a *AnnotationKey
func AnnotationKeyPtr(name string) *gatewayv1b1.AnnotationKey {
	key := gatewayv1b1.AnnotationKey(name)
	return &key
}

// AnnotationKeyStr translates reference value of ptr to string
func AnnotationKeyStr(name *gatewayv1b1.AnnotationKey) string {
	key := string(*name)
	return key
}

// AnnotationValuePtr translates a string to a *AnnotationValue
func AnnotationValuePtr(name string) *gatewayv1b1.AnnotationValue {
	val := gatewayv1b1.AnnotationValue(name)
	return &val
}

// AnnotationValueStr translates reference value of ptr to string
func AnnotationValueStr(name *gatewayv1b1.AnnotationValue) string {
	val := string(*name)
	return val
}

// AddressTypePtr translates a string to a *AddressType
func AddressTypePtr(name string) *gatewayv1b1.AddressType {
	addr := gatewayv1b1.AddressType(name)
	return &addr
}

// AddressTypeStr translates reference value of ptr to string
func AddressTypeStr(name *gatewayv1b1.AddressType) string {
	val := string(*name)
	return val
}

// RouteConditionTypePtr translates a string to a *RouteConditionType
func RouteConditionTypePtr(name string) *gatewayv1b1.RouteConditionType {
	str := gatewayv1b1.RouteConditionType(name)
	return &str
}

// RouteConditionTypeStr translates reference value of ptr to string
func RouteConditionTypeStr(name *gatewayv1b1.RouteConditionType) string {
	val := string(*name)
	return val
}

// RouteConditionReasonPtr translates a string to a *RouteConditionReason
func RouteConditionReasonPtr(name string) *gatewayv1b1.RouteConditionReason {
	str := gatewayv1b1.RouteConditionReason(name)
	return &str
}

// RouteConditionReasonStr translates reference value of ptr to string
func RouteConditionReasonStr(name *gatewayv1b1.RouteConditionType) string {
	val := string(*name)
	return val
}

// ProtocolTypePtr translates a string to a *ProtocolType
func ProtocolTypePtr(name string) *gatewayv1b1.ProtocolType {
	proto := gatewayv1b1.ProtocolType(name)
	return &proto
}

// ProtocolTypeStr translates reference value of ptr to string
func ProtocolTypeStr(name *gatewayv1b1.ProtocolType) string {
	val := string(*name)
	return val
}

// TLSModeTypePtr translates a string to a *TLSModeType
func TLSModeTypePtr(name string) *gatewayv1b1.TLSModeType {
	tls := gatewayv1b1.TLSModeType(name)
	return &tls
}

// TLSModeTypeStr translates reference value of ptr to string
func TLSModeTypeStr(name *gatewayv1b1.TLSModeType) string {
	val := string(*name)
	return val
}
