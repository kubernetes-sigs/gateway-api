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

package translator

import (
	gatewayv1a2 "sigs.k8s.io/gateway-api/apis/v1alpha2"
)

// PortNumberPtr translates an int to a *PortNumber
func PortNumberPtr(p int) *gatewayv1a2.PortNumber {
	result := gatewayv1a2.PortNumber(p)
	return &result
}

// PortNumberInt32 translates reference value of ptr to Int32
func PortNumberInt32(name *gatewayv1a2.PortNumber) int32 {
	portNum := int32(*name)
	return portNum
}

// SectionNamePtr translates an int to a *SectionName
func SectionNamePtr(sectionName string) *gatewayv1a2.SectionName {
	gwSectionName := gatewayv1a2.SectionName(sectionName)
	return &gwSectionName
}

// SectionNameStr translates reference value of ptr to string
func SectionNameStr(name *gatewayv1a2.SectionName) string {
	sectionName := string(*name)
	return sectionName
}

// HostnamePtr translates an int to a *Hostname
func HostnamePtr(host string) *gatewayv1a2.Hostname {
	h := gatewayv1a2.Hostname(host)
	return &h
}

// HostnameStr translates reference value of ptr to string
func HostnameStr(name *gatewayv1a2.Hostname) string {
	hostName := string(*name)
	return hostName
}

// PreciseHostnamePtr translates an int to a *PreciseHostname
func PreciseHostnamePtr(host string) *gatewayv1a2.PreciseHostname {
	h := gatewayv1a2.PreciseHostname(host)
	return &h
}

// PreciseHostnameStr translates reference value of ptr to string
func PreciseHostnameStr(name *gatewayv1a2.PreciseHostname) string {
	prechostName := string(*name)
	return prechostName
}

// GroupPtr translates an int to a *Group
func GroupPtr(group string) *gatewayv1a2.Group {
	gwGroup := gatewayv1a2.Group(group)
	return &gwGroup
}

// GroupStr translates reference value of ptr to string
func GroupStr(name *gatewayv1a2.Group) string {
	groupStr := string(*name)
	return groupStr
}

// KindPtr translates an int to a *Kind
func KindPtr(kind string) *gatewayv1a2.Kind {
	gwKind := gatewayv1a2.Kind(kind)
	return &gwKind
}

// KindStr translates reference value of ptr to string
func KindStr(name *gatewayv1a2.Kind) string {
	kindStr := string(*name)
	return kindStr
}

// NamespacePtr translates an int to a *Namespace
func NamespacePtr(namespace string) *gatewayv1a2.Namespace {
	gwNamespace := gatewayv1a2.Namespace(namespace)
	return &gwNamespace
}

// NamespaceStr translates reference value of ptr to string
func NamespaceStr(name *gatewayv1a2.Namespace) string {
	namespace := string(*name)
	return namespace
}

// ObjectNamePtr translates an int to a *ObjectName
func ObjectNamePtr(name string) *gatewayv1a2.ObjectName {
	objectName := gatewayv1a2.ObjectName(name)
	return &objectName
}

// ObjectNameStr translates reference value of ptr to string
func ObjectNameStr(name *gatewayv1a2.ObjectName) string {
	objname := string(*name)
	return objname
}

// GatewayControllerPtr translates an int to a *GatewayController
func GatewayControllerPtr(name string) *gatewayv1a2.GatewayController {
	gwCtrl := gatewayv1a2.GatewayController(name)
	return &gwCtrl
}

// GatewayControllerStr translates reference value of ptr to string
func GatewayControllerStr(name *gatewayv1a2.GatewayController) string {
	gw := string(*name)
	return gw
}

// AnnotationKeyPtr translates an int to a *AnnotationKey
func AnnotationKeyPtr(name string) *gatewayv1a2.AnnotationKey {
	key := gatewayv1a2.AnnotationKey(name)
	return &key
}

// AnnotationKeyStr translates reference value of ptr to string
func AnnotationKeyStr(name *gatewayv1a2.AnnotationKey) string {
	key := string(*name)
	return key
}

// AnnotationValuePtr translates an int to a *AnnotationValue
func AnnotationValuePtr(name string) *gatewayv1a2.AnnotationValue {
	val := gatewayv1a2.AnnotationValue(name)
	return &val
}

// AnnotationValueStr translates reference value of ptr to string
func AnnotationValueStr(name *gatewayv1a2.AnnotationValue) string {
	val := string(*name)
	return val
}

// AddressTypePtr translates an int to a *AddressType
func AddressTypePtr(name string) *gatewayv1a2.AddressType {
	addr := gatewayv1a2.AddressType(name)
	return &addr
}

// AddressTypeStr translates reference value of ptr to string
func AddressTypeStr(name *gatewayv1a2.AddressType) string {
	val := string(*name)
	return val
}

// RouteConditionTypePtr translates an int to a *RouteConditionType
func RouteConditionTypePtr(name string) *gatewayv1a2.RouteConditionType {
	str := gatewayv1a2.RouteConditionType(name)
	return &str
}

// RouteConditionTypeStr translates reference value of ptr to string
func RouteConditionTypeStr(name *gatewayv1a2.RouteConditionType) string {
	val := string(*name)
	return val
}

// RouteConditionReasonPtr translates an int to a *RouteConditionReason
func RouteConditionReasonPtr(name string) *gatewayv1a2.RouteConditionReason {
	str := gatewayv1a2.RouteConditionReason(name)
	return &str
}

// RouteConditionReasonStr translates reference value of ptr to string
func RouteConditionReasonStr(name *gatewayv1a2.RouteConditionType) string {
	val := string(*name)
	return val
}

// ProtocolTypePtr translates an int to a *ProtocolType
func ProtocolTypePtr(name string) *gatewayv1a2.ProtocolType {
	proto := gatewayv1a2.ProtocolType(name)
	return &proto
}

// ProtocolTypeStr translates reference value of ptr to string
func ProtocolTypeStr(name *gatewayv1a2.ProtocolType) string {
	val := string(*name)
	return val
}

// TLSModeTypePtr translates an int to a *TLSModeType
func TLSModeTypePtr(name string) *gatewayv1a2.TLSModeType {
	tls := gatewayv1a2.TLSModeType(name)
	return &tls
}

// TLSModeTypeStr translates reference value of ptr to string
func TLSModeTypeStr(name *gatewayv1a2.TLSModeType) string {
	val := string(*name)
	return val
}
