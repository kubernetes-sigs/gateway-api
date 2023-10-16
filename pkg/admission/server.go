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
	"io"
	"net/http"

	admission "k8s.io/api/admission/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/apimachinery/pkg/util/validation/field"
	"k8s.io/klog/v2"

	v1 "sigs.k8s.io/gateway-api/apis/v1"
	v1Validation "sigs.k8s.io/gateway-api/apis/v1/validation"
	v1alpha2 "sigs.k8s.io/gateway-api/apis/v1alpha2"
	v1a2Validation "sigs.k8s.io/gateway-api/apis/v1alpha2/validation"
	v1beta1 "sigs.k8s.io/gateway-api/apis/v1beta1"
	v1b1Validation "sigs.k8s.io/gateway-api/apis/v1beta1/validation"
)

const admissionReview = "AdmissionReview"

var (
	scheme = runtime.NewScheme()
	codecs = serializer.NewCodecFactory(scheme)
)

var (
	v1a2TCPRouteGVP = meta.GroupVersionResource{
		Group:    v1alpha2.GroupVersion.Group,
		Version:  v1alpha2.GroupVersion.Version,
		Resource: "tcproutes",
	}
	v1a2UDPRouteGVP = meta.GroupVersionResource{
		Group:    v1alpha2.GroupVersion.Group,
		Version:  v1alpha2.GroupVersion.Version,
		Resource: "udproutes",
	}
	v1a2TLSRouteGVP = meta.GroupVersionResource{
		Group:    v1alpha2.GroupVersion.Group,
		Version:  v1alpha2.GroupVersion.Version,
		Resource: "tlsroutes",
	}
	v1a2GRPCRouteGVR = meta.GroupVersionResource{
		Group:    v1alpha2.SchemeGroupVersion.Group,
		Version:  v1alpha2.SchemeGroupVersion.Version,
		Resource: "grpcroutes",
	}
	v1b1HTTPRouteGVR = meta.GroupVersionResource{
		Group:    v1beta1.SchemeGroupVersion.Group,
		Version:  v1beta1.SchemeGroupVersion.Version,
		Resource: "httproutes",
	}
	v1b1GatewayGVR = meta.GroupVersionResource{
		Group:    v1beta1.SchemeGroupVersion.Group,
		Version:  v1beta1.SchemeGroupVersion.Version,
		Resource: "gateways",
	}
	v1b1GatewayClassGVR = meta.GroupVersionResource{
		Group:    v1beta1.SchemeGroupVersion.Group,
		Version:  v1beta1.SchemeGroupVersion.Version,
		Resource: "gatewayclasses",
	}
	v1HTTPRouteGVR = meta.GroupVersionResource{
		Group:    v1.SchemeGroupVersion.Group,
		Version:  v1.SchemeGroupVersion.Version,
		Resource: "httproutes",
	}
	v1GatewayGVR = meta.GroupVersionResource{
		Group:    v1.SchemeGroupVersion.Group,
		Version:  v1.SchemeGroupVersion.Version,
		Resource: "gateways",
	}
	v1GatewayClassGVR = meta.GroupVersionResource{
		Group:    v1.SchemeGroupVersion.Group,
		Version:  v1.SchemeGroupVersion.Version,
		Resource: "gatewayclasses",
	}
)

func log500(w http.ResponseWriter, err error) {
	klog.Errorf("failed to process request: %v\n", err)
	http.Error(w, err.Error(), http.StatusInternalServerError)
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
	data, err := io.ReadAll(r.Body)
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

	if review.Kind != admissionReview {
		http.Error(w, "submitted object is not of kind AdmissionReview", http.StatusBadRequest)
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
}

func handleValidation(request admission.AdmissionRequest) (*admission.AdmissionResponse, error) {
	var (
		response     admission.AdmissionResponse
		deserializer = codecs.UniversalDeserializer()
		fieldErr     field.ErrorList
	)

	if request.Operation == admission.Delete ||
		request.Operation == admission.Connect {
		response.UID = request.UID
		response.Allowed = true
		return &response, nil
	}

	switch request.Resource {
	case v1a2TCPRouteGVP:
		var tRoute v1alpha2.TCPRoute
		_, _, err := deserializer.Decode(request.Object.Raw, nil, &tRoute)
		if err != nil {
			return nil, err
		}
		fieldErr = v1a2Validation.ValidateTCPRoute(&tRoute)
	case v1a2UDPRouteGVP:
		var uRoute v1alpha2.UDPRoute
		_, _, err := deserializer.Decode(request.Object.Raw, nil, &uRoute)
		if err != nil {
			return nil, err
		}
		fieldErr = v1a2Validation.ValidateUDPRoute(&uRoute)
	case v1a2TLSRouteGVP:
		var tRoute v1alpha2.TLSRoute
		_, _, err := deserializer.Decode(request.Object.Raw, nil, &tRoute)
		if err != nil {
			return nil, err
		}
		fieldErr = v1a2Validation.ValidateTLSRoute(&tRoute)
	case v1a2GRPCRouteGVR:
		var gRoute v1alpha2.GRPCRoute
		_, _, err := deserializer.Decode(request.Object.Raw, nil, &gRoute)
		if err != nil {
			return nil, err
		}

		fieldErr = v1a2Validation.ValidateGRPCRoute(&gRoute)
	case v1b1HTTPRouteGVR:
		var hRoute v1beta1.HTTPRoute
		_, _, err := deserializer.Decode(request.Object.Raw, nil, &hRoute)
		if err != nil {
			return nil, err
		}

		fieldErr = v1b1Validation.ValidateHTTPRoute(&hRoute)
	case v1b1GatewayGVR:
		var gateway v1beta1.Gateway
		_, _, err := deserializer.Decode(request.Object.Raw, nil, &gateway)
		if err != nil {
			return nil, err
		}
		fieldErr = v1b1Validation.ValidateGateway(&gateway)
	case v1b1GatewayClassGVR:
		// runs only for updates
		if request.Operation != admission.Update {
			break
		}
		var gatewayClass v1beta1.GatewayClass
		_, _, err := deserializer.Decode(request.Object.Raw, nil, &gatewayClass)
		if err != nil {
			return nil, err
		}
		var gatewayClassOld v1beta1.GatewayClass
		_, _, err = deserializer.Decode(request.OldObject.Raw, nil, &gatewayClassOld)
		if err != nil {
			return nil, err
		}
		fieldErr = v1b1Validation.ValidateGatewayClassUpdate(&gatewayClassOld, &gatewayClass)
	case v1HTTPRouteGVR:
		var hRoute v1.HTTPRoute
		_, _, err := deserializer.Decode(request.Object.Raw, nil, &hRoute)
		if err != nil {
			return nil, err
		}
		fieldErr = v1Validation.ValidateHTTPRoute(&hRoute)
	case v1GatewayGVR:
		var gateway v1.Gateway
		_, _, err := deserializer.Decode(request.Object.Raw, nil, &gateway)
		if err != nil {
			return nil, err
		}
		fieldErr = v1Validation.ValidateGateway(&gateway)
	case v1GatewayClassGVR:
		// runs only for updates
		if request.Operation != admission.Update {
			break
		}
		var gatewayClass v1.GatewayClass
		_, _, err := deserializer.Decode(request.Object.Raw, nil, &gatewayClass)
		if err != nil {
			return nil, err
		}
		var gatewayClassOld v1.GatewayClass
		_, _, err = deserializer.Decode(request.OldObject.Raw, nil, &gatewayClassOld)
		if err != nil {
			return nil, err
		}
		fieldErr = v1Validation.ValidateGatewayClassUpdate(&gatewayClassOld, &gatewayClass)
	default:
		return nil, fmt.Errorf("unknown resource '%v'", request.Resource.Resource)
	}

	if len(fieldErr) > 0 {
		return &admission.AdmissionResponse{
			UID:     request.UID,
			Allowed: false,
			Result: &meta.Status{
				Message: fmt.Sprintf("%s", fieldErr.ToAggregate()),
				Code:    400,
			},
		}, nil
	}

	return &admission.AdmissionResponse{
		UID:     request.UID,
		Allowed: true,
		Result:  &meta.Status{},
	}, nil
}
