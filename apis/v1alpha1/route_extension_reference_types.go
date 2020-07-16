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

package v1alpha1

// RouteMatchExtensionObjectReference identifies a route-match extension object
// within a known namespace.
//
// +k8s:deepcopy-gen=false
type RouteMatchExtensionObjectReference = ConfigMapsDefaultLocalObjectReference

// RouteFilterExtensionObjectReference identifies a route-filter extension
// object within a known namespace.
//
// +k8s:deepcopy-gen=false
type RouteFilterExtensionObjectReference = ConfigMapsDefaultLocalObjectReference

// RouteActionExtensionObjectReference identifies a route-action extension
// object within a known namespace.
//
// +k8s:deepcopy-gen=false
type RouteActionExtensionObjectReference = ConfigMapsDefaultLocalObjectReference
