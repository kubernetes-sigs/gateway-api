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

	admission "k8s.io/api/admission/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/klog/v2"

	v1alpha1 "sigs.k8s.io/gateway-api/apis/v1alpha1"
	v1a1Validation "sigs.k8s.io/gateway-api/apis/v1alpha1/validation"
	v1alpha2 "sigs.k8s.io/gateway-api/apis/v1alpha2"
	v1a2Validation "sigs.k8s.io/gateway-api/apis/v1alpha2/validation"
)

var (
	scheme = runtime.NewScheme()
	codecs = serializer.NewCodecFactory(scheme)
)

var (
	v1a1httpRouteGVR = meta.GroupVersionResource{
		Group:    v1alpha1.SchemeGroupVersion.Group,
		Version:  v1alpha1.SchemeGroupVersion.Version,
		Resource: "httproutes",
	}
	v1a2httpRouteGVR = meta.GroupVersionResource{
		Group:    v1alpha2.SchemeGroupVersion.Group,
		Version:  v1alpha2.SchemeGroupVersion.Version,
		Resource: "httproutes",
	}
)

func log500(w http.ResponseWriter, err error) {
	klog.Errorf("failed to process request: %v\n", err)
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
		http.Error(w, fmt.Sprintf("invalid method %s, only POST requests are allowed", r.Method), http.StatusMethodNotAllowed)
		return
	}

	if r.Body == nil {
		http.Error(w, "admission review object is missing",
			http.StatusBadRequest)
		return
	}
	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log500(w, err)
		return
	}

	review := admission.AdmissionReview{}
	err = json.Unmarshal(data, &review)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if !ensureKindAdmissionReview(data) {
		invalidKind := "submitted object is not of kind AdmissionReview"
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
		klog.Errorf("failed to write HTTP response: %v\n", err)
		return
	}
	return
}

func handleValidation(request admission.AdmissionRequest) (*admission.AdmissionResponse, error) {

	var (
		response admission.AdmissionResponse
		message  string
		ok       bool
	)

	if request.Operation == admission.Delete ||
		request.Operation == admission.Connect {
		response.UID = request.UID
		response.Allowed = true
		return &response, nil
	}

	switch request.Resource {
	case v1a1httpRouteGVR:
		var hRoute v1alpha1.HTTPRoute
		deserializer := codecs.UniversalDeserializer()
		_, _, err := deserializer.Decode(request.Object.Raw, nil, &hRoute)
		if err != nil {
			return nil, err
		}

		fieldErr := v1a1Validation.ValidateHTTPRoute(&hRoute)
		if fieldErr != nil {
			message = fmt.Sprintf("%s", fieldErr.ToAggregate())
			ok = false
		} else {
			ok = true
		}
	case v1a2httpRouteGVR:
		var hRoute v1alpha2.HTTPRoute
		deserializer := codecs.UniversalDeserializer()
		_, _, err := deserializer.Decode(request.Object.Raw, nil, &hRoute)
		if err != nil {
			return nil, err
		}

		fieldErr := v1a2Validation.ValidateHTTPRoute(&hRoute)
		if fieldErr != nil {
			message = fmt.Sprintf("%s", fieldErr.ToAggregate())
			ok = false
		} else {
			ok = true
		}
	default:
		return nil, fmt.Errorf("unknown resource '%v'", request.Resource.Resource)
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
