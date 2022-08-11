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

package utils

import (
	gatewayv1a2 "sigs.k8s.io/gateway-api/apis/v1alpha2"
)

// PathMatchTypePtr translates a string to *PathMatchType
func PathMatchTypePtr(s string) *gatewayv1a2.PathMatchType {
	result := gatewayv1a2.PathMatchType(s)
	return &result
}

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

func ListenerHostnameToPtr(host string) *gatewayv1a2.Hostname {
	h := *gatewayv1a2.Hostname(host)
	return &h
}

func ListenerHostnameFromPtr(name *Hostname) string {
	hostName := string(*name)
	return hostName
}

func PreciseHostnameToPtr(host string) *PreciseHostname {
	h := PreciseHostname(host)
	return &h
}

func PreciseHostnameFromPtr(name *PreciseHostname) string {
	prechostName := string(*name)
	return prechostName
}

func GroupToPtr(group string) *Group {
	gwGroup := Group(group)
	return &gwGroup
}

func GroupFromPtr(name *Group) string {
	groupStr := string(*name)
	return groupStr
}

func KindToPtr(kind string) *Kind {
	gwKind := Kind(kind)
	return &gwKind
}

func KindFromPtr(name *Kind) string {
	kindStr := string(*name)
	return kindStr
}

func NamespaceToPtr(namespace string) *Namespace {
	gwNamespace := Namespace(namespace)
	return &gwNamespace
}

func NamespaceFromPtr(name *Namespace) string {
	namespace := string(*name)
	return namespace
}

func ObjectNameToPtr(name string) *ObjectName {
	objectName := ObjectName(name)
	return &objectName
}

func ObjectNameFromPtr(name *ObjectName) string {
	objname := string(*name)
	return objname
}

func GatewayControllerToPtr(name string) *GatewayController {
	gwCtrl := GatewayController(name)
	return &gwCtrl
}

func GatewayControllerFromPtr(name *GatewayController) string {
	gw := string(*name)
	return gw
}

func AnnotationKeyToPtr(name string) *AnnotationKey {
	key := AnnotationKey(name)
	return &key
}

func AnnotationKeyFromPtr(name *AnnotationKey) string {
	key := string(*name)
	return key
}

func AnnotationValueToPtr(name string) *AnnotationValue {
	val := AnnotationValue(name)
	return &val
}

func AnnotationValueFromPtr(name *AnnotationValue) string {
	val := string(*name)
	return val
}

func AddressTypeToPtr(name string) *AddressType {
	addr := AddressType(name)
	return &addr
}

func AddressTypeFromPtr(name *AddressType) string {
	val := string(*name)
	return val
}

func RouteConditionTypeToPtr(name string) *RouteConditionType {
	str := RouteConditionType(name)
	return &str
}

func RouteConditionTypeFromPtr(name *RouteConditionType) string {
	val := string(*name)
	return val
}

func RouteConditionReasonToPtr(name string) *RouteConditionReason {
	str := RouteConditionReason(name)
	return &str
}

func RouteConditionReasonFromPtr(name *RouteConditionType) string {
	val := string(*name)
	return val
}

func ProtocolTypeToPtr(name string) *ProtocolType {
	proto := ProtocolType(name)
	return &proto
}

func ProtocolTypeFromPtr(name *ProtocolType) string {
	val := string(*name)
	return val
}

func TLSModeTypePtr(name string) *TLSModeType {
	tls := TLSModeType(name)
	return &tls
}

func TLSModeTypeFromPtr(name *TLSModeType) string {
	val := string(*name)
	return val
}
