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

package admission

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"reflect"

	admission "k8s.io/api/admission/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/klog/v2"

	v1alpha1 "sigs.k8s.io/gateway-api/apis/v1alpha1"
)

var (
	scheme = runtime.NewScheme()
	codecs = serializer.NewCodecFactory(scheme)
)

func log500(w http.ResponseWriter, err error) {
	klog.Errorf("failed to read request from client: %v\n", err)
	http.Error(w, err.Error(), http.StatusInternalServerError)
	return
}

// ensureKindAdmissionReview check that our admission server is only getting requests
// for kind AdmissionReview and reject all others
func ensureKindAdmissionReview(req []byte) bool {
	type reqBody struct {
		Kind  string                 `json:"kind"`
		Extra map[string]interface{} `json:"-"`
	}
	var msg reqBody
	err := json.Unmarshal(req, &msg)
	if err != nil {
		return false
	}
	if msg.Kind != "AdmissionReview" {
		return false
	}
	return true
}

// ServeHTTP parses AdmissionReview requests and responds back
// with the validation result of the entity.
func ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		klog.Errorf("webhook cannot handle %s method", r.Method)
		http.Error(w, fmt.Sprintf("invalid method %s, only POST requests are allowed", r.Method), http.StatusMethodNotAllowed)
	}
	if r.Body == nil {
		klog.Errorf("received request with empty body")
		http.Error(w, "admission review object is missing",
			http.StatusBadRequest)
		return
	}
	data, readErr := ioutil.ReadAll(r.Body)
	if readErr != nil {
		log500(w, readErr)
	}
	review := admission.AdmissionReview{}
	if err := json.Unmarshal(data, &review); err != nil {
		klog.Errorf("failed to parse AdmissionReview object: %v\n", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if !ensureKindAdmissionReview(data) {
		invalidKind := "submitted object is not of kind AdmissionReview"
		klog.Errorf(invalidKind)
		http.Error(w, invalidKind, http.StatusBadRequest)
		return
	}
	response, err := handleValidation(*review.Request)
	if err != nil {
		log500(w, err)
		return
	}
	review.Response = response
	data, err = json.Marshal(review)
	if err != nil {
		log500(w, err)
		return
	}
	_, err = w.Write(data)
	if err != nil {
		klog.Errorf("failed to write response: %v\n", err)
	}
	return
}

var (
	httpRoute = meta.GroupVersionResource{
		Group:    v1alpha1.SchemeGroupVersion.Group,
		Version:  v1alpha1.SchemeGroupVersion.Version,
		Resource: "httproutes",
	}
)

func routesEqual(new v1alpha1.HTTPRoute, old v1alpha1.HTTPRoute) bool {
	return reflect.DeepEqual(new.Spec, old.Spec)
}

func handleValidation(request admission.AdmissionRequest) (
	*admission.AdmissionResponse, error) {
	var response admission.AdmissionResponse
	var ok bool
	var message string
	var err error

	switch request.Resource {
	case httpRoute:
		hRoute := v1alpha1.HTTPRoute{}
		deserializer := codecs.UniversalDeserializer()
		_, _, err = deserializer.Decode(request.Object.Raw, nil, &hRoute)
		if err != nil {
			return nil, err
		}
		// The admission hook is only configured for create & update, so we can
		// ignore explicit validation for Connect & Delete.
		// nolint:exhaustive
		switch request.Operation {
		case admission.Create:
			ok, message, err = ValidateHTTPRoute(hRoute)
			if err != nil {
				return nil, err
			}
		case admission.Update:
			oldRoute := v1alpha1.HTTPRoute{}
			_, _, err = deserializer.Decode(request.OldObject.Raw, nil, &oldRoute)
			if err != nil {
				return nil, err
			}
			// validate if routes are changed
			if !routesEqual(hRoute, oldRoute) {
				ok, message, err = ValidateHTTPRoute(hRoute)
				if err != nil {
					return nil, err
				}
			} else {
				ok = true
			}
		default:
			return nil, fmt.Errorf("unknown operation '%v'", string(request.Operation))
		}
	default:
		return nil, fmt.Errorf("unknown resource '%v'", request.Resource.Resource)
	}

	if err != nil {
		return nil, err
	}
	response.UID = request.UID
	response.Allowed = ok
	response.Result = &meta.Status{
		Message: message,
	}
	if !ok {
		response.Result.Code = 400
	}
	return &response, nil
}
