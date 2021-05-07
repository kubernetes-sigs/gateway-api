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
	"bytes"
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/lithammer/dedent"
	"github.com/stretchr/testify/assert"
	admission "k8s.io/api/admission/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var decoder = codecs.UniversalDeserializer()

func TestServeHTTPInvalidBody(t *testing.T) {
	assert := assert.New(t)
	res := httptest.NewRecorder()
	handler := http.HandlerFunc(ServeHTTP)
	req, err := http.NewRequest("POST", "", nil)
	req = req.WithContext(context.Background())
	assert.Nil(err)
	handler.ServeHTTP(res, req)
	assert.Equal(400, res.Code)
	assert.Equal("admission review object is missing\n",
		res.Body.String())
}

func TestServeHTTPInvalidMethod(t *testing.T) {
	assert := assert.New(t)
	res := httptest.NewRecorder()
	handler := http.HandlerFunc(ServeHTTP)
	req, err := http.NewRequest("GET", "", nil)
	req = req.WithContext(context.Background())
	assert.Nil(err)
	handler.ServeHTTP(res, req)
	assert.Equal(http.StatusMethodNotAllowed, res.Code)
	assert.Equal("invalid method GET, only POST requests are allowed\n",
		res.Body.String())
}

func TestServeHTTPSubmissions(t *testing.T) {

	const invalidHostname = "!@!.foo.com"
	const lowercaseRFC1123ErrorMessage = "a lowercase RFC 1123 subdomain must consist of lower case alphanumeric characters, '-' or '.', and must start and end with an alphanumeric character (e.g. 'example.com', regex used for validation is '[a-z0-9]([-a-z0-9]*[a-z0-9])?(\\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*')"

	for _, apiVersion := range []string{
		"admission.k8s.io/v1",
		"admission.k8s.io/v1",
	} {
		for _, tt := range []struct {
			name    string
			reqBody string

			wantRespCode        int
			wantSuccessResponse admission.AdmissionResponse
			wantFailureMessage  string
		}{
			{
				name: "malformed json missing colon at resource",
				reqBody: dedent.Dedent(`{
						"kind": "AdmissionReview",
						"apiVersion": "` + apiVersion + `",
						"request": {
							"uid": "7313cd05-eddc-4150-b88c-971a0d53b2ab",
							"resource": {
								"group": "networking.x-k8s.io",
								"version": "v1alpha1",
								"resource" "httproutes"
							},
							"object": {
								"apiVersion": "networking.x-k8s.io/v1alpha1",
								"kind": "HTTPRoute"
							},
						"operation": "CREATE"
						}
					}`),
				wantRespCode:       http.StatusBadRequest,
				wantFailureMessage: "invalid character '\"' after object key\n",
			},
			{
				name:               "request with empty body",
				wantRespCode:       http.StatusBadRequest,
				wantFailureMessage: "unexpected end of JSON input\n",
			},
			{
				name: "valid json but not of kind AdmissionReview",
				reqBody: dedent.Dedent(`{
						"kind": "NotReviewYouAreLookingFor",
						"apiVersion": "` + apiVersion + `",
						"request": {
							"uid": "7313cd05-eddc-4150-b88c-971a0d53b2ab",
							"resource": {
								"group": "networking.x-k8s.io",
								"version": "v1alpha1",
								"resource": "httproutes"
							},
							"object": {
								"apiVersion": "networking.x-k8s.io/v1alpha1",
								"kind": "HTTPRoute"
							},
						"operation": "CREATE"
						}
					}`),
				wantRespCode:       http.StatusBadRequest,
				wantFailureMessage: "submitted object is not of kind AdmissionReview\n",
			},
			{
				name: "valid v1alpha1 Gateway resource",
				reqBody: dedent.Dedent(`{
						"kind": "AdmissionReview",
						"apiVersion": "` + apiVersion + `",
						"request": {
							"uid": "7313cd05-eddc-4150-b88c-971a0d53b2ab",
							"resource": {
								"group": "networking.x-k8s.io",
								"version": "v1alpha1",
								"resource": "gateways"
							},
							"object": {
   								"kind": "Gateway",
   								"apiVersion": "networking.x-k8s.io/v1alpha1",
   								"metadata": {
   								   "name": "gateway-1",
   								   "labels": {
   								      "app": "foo"
   								   }
   								},
   								"spec": {
									"gatewayClassName": "contour-class",
									"listeners": [
										{
											"port": 80,
											"protocol": "HTTP",
											"hostname": "foo.com",
											"routes": {
												"group": "networking.x-k8s.io",
												"kind": "HTTPRoute",
												"namespaces": {
													"from": "All"
												}
											}
										}
									]
   								}
							},
						"operation": "CREATE"
						}
					}`),
				wantRespCode: http.StatusOK,
				wantSuccessResponse: admission.AdmissionResponse{
					UID:     "7313cd05-eddc-4150-b88c-971a0d53b2ab",
					Allowed: true,
					Result:  &metav1.Status{},
				},
			},
			{
				name: "invalid v1alpha1 Gateway resource with bad listener hostname",
				reqBody: dedent.Dedent(`{
						"kind": "AdmissionReview",
						"apiVersion": "` + apiVersion + `",
						"request": {
							"uid": "7313cd05-eddc-4150-b88c-971a0d53b2ab",
							"resource": {
								"group": "networking.x-k8s.io",
								"version": "v1alpha1",
								"resource": "gateways"
							},
							"object": {
   								"kind": "Gateway",
   								"apiVersion": "networking.x-k8s.io/v1alpha1",
   								"metadata": {
   								   "name": "gateway-1",
   								   "labels": {
   								      "app": "foo"
   								   }
   								},
   								"spec": {
									"gatewayClassName": "contour-class",
									"listeners": [
										{
											"port": 80,
											"protocol": "HTTP",
											"hostname": "` + invalidHostname + `",
											"routes": {
												"group": "networking.x-k8s.io",
												"kind": "HTTPRoute",
												"namespaces": {
													"from": "All"
												}
											}
										}
									]
   								}
							},
						"operation": "CREATE"
						}
					}`),
				wantRespCode: http.StatusOK,
				wantSuccessResponse: admission.AdmissionResponse{
					UID:     "7313cd05-eddc-4150-b88c-971a0d53b2ab",
					Allowed: false,
					Result: &metav1.Status{
						Code:    400,
						Message: fmt.Sprintf("spec.listeners[0].hostname: Invalid value: %q: %s", invalidHostname, lowercaseRFC1123ErrorMessage),
					},
				},
			},
			{
				name: "valid v1alpha2 Gateway resource",
				reqBody: dedent.Dedent(`{
						"kind": "AdmissionReview",
						"apiVersion": "` + apiVersion + `",
						"request": {
							"uid": "7313cd05-eddc-4150-b88c-971a0d53b2ab",
							"resource": {
								"group": "networking.x-k8s.io",
								"version": "v1alpha2",
								"resource": "gateways"
							},
							"object": {
   								"kind": "Gateway",
   								"apiVersion": "networking.x-k8s.io/v1alpha2",
   								"metadata": {
   								   "name": "gateway-1",
   								   "labels": {
   								      "app": "foo"
   								   }
   								},
   								"spec": {
									"gatewayClassName": "contour-class",
									"listeners": [
										{
											"port": 80,
											"protocol": "HTTP",
											"hostname": "foo.com",
											"routes": {
												"group": "networking.x-k8s.io",
												"kind": "HTTPRoute",
												"namespaces": {
													"from": "All"
												}
											}
										}
									]
   								}
							},
						"operation": "CREATE"
						}
					}`),
				wantRespCode: http.StatusOK,
				wantSuccessResponse: admission.AdmissionResponse{
					UID:     "7313cd05-eddc-4150-b88c-971a0d53b2ab",
					Allowed: true,
					Result:  &metav1.Status{},
				},
			},
			{
				name: "invalid v1alpha2 Gateway resource with bad listener hostname",
				reqBody: dedent.Dedent(`{
						"kind": "AdmissionReview",
						"apiVersion": "` + apiVersion + `",
						"request": {
							"uid": "7313cd05-eddc-4150-b88c-971a0d53b2ab",
							"resource": {
								"group": "networking.x-k8s.io",
								"version": "v1alpha2",
								"resource": "gateways"
							},
							"object": {
   								"kind": "Gateway",
   								"apiVersion": "networking.x-k8s.io/v1alpha2",
   								"metadata": {
   								   "name": "gateway-1",
   								   "labels": {
   								      "app": "foo"
   								   }
   								},
   								"spec": {
									"gatewayClassName": "contour-class",
									"listeners": [
										{
											"port": 80,
											"protocol": "HTTP",
											"hostname": "` + invalidHostname + `",
											"routes": {
												"group": "networking.x-k8s.io",
												"kind": "HTTPRoute",
												"namespaces": {
													"from": "All"
												}
											}
										}
									]
   								}
							},
						"operation": "CREATE"
						}
					}`),
				wantRespCode: http.StatusOK,
				wantSuccessResponse: admission.AdmissionResponse{
					UID:     "7313cd05-eddc-4150-b88c-971a0d53b2ab",
					Allowed: false,
					Result: &metav1.Status{
						Code:    400,
						Message: fmt.Sprintf("spec.listeners[0].hostname: Invalid value: %q: %s", invalidHostname, lowercaseRFC1123ErrorMessage),
					},
				},
			},
			{
				name: "valid HTTPRoute resource",
				reqBody: dedent.Dedent(`{
						"kind": "AdmissionReview",
						"apiVersion": "` + apiVersion + `",
						"request": {
							"uid": "7313cd05-eddc-4150-b88c-971a0d53b2ab",
							"resource": {
								"group": "networking.x-k8s.io",
								"version": "v1alpha1",
								"resource": "httproutes"
							},
							"object": {
   								"kind": "HTTPRoute",
   								"apiVersion": "networking.x-k8s.io/v1alpha1",
   								"metadata": {
   								   "name": "http-app-1",
   								   "labels": {
   								      "app": "foo"
   								   }
   								},
   								"spec": {
   								   "hostnames": [
   								      "foo.com"
   								   ],
   								   "rules": [
   								      {
   								         "matches": [
   								            {
   								               "path": {
   								                  "type": "Prefix",
   								                  "value": "/bar"
   								               }
   								            }
   								         ],
   								         "filters": [
   								            {
   								               "type": "RequestMirror",
   								               "requestMirror": {
   								                  "serviceName": "my-service1-staging",
   								                  "port": 8080
   								               }
   								            }
   								         ],
   								         "forwardTo": [
   								            {
   								               "serviceName": "my-service1",
   								               "port": 8080
   								            }
   								         ]
   								      }
   								   ]
   								}
							},
						"operation": "CREATE"
						}
					}`),
				wantRespCode: http.StatusOK,
				wantSuccessResponse: admission.AdmissionResponse{
					UID:     "7313cd05-eddc-4150-b88c-971a0d53b2ab",
					Allowed: true,
					Result:  &metav1.Status{},
				},
			},
			{
				name: "valid v1alpha2 HTTPRoute resource",
				reqBody: dedent.Dedent(`{
						"kind": "AdmissionReview",
						"apiVersion": "` + apiVersion + `",
						"request": {
							"uid": "7313cd05-eddc-4150-b88c-971a0d53b2ab",
							"resource": {
								"group": "networking.x-k8s.io",
								"version": "v1alpha2",
								"resource": "httproutes"
							},
							"object": {
   								"kind": "HTTPRoute",
   								"apiVersion": "networking.x-k8s.io/v1alpha2",
   								"metadata": {
   								   "name": "http-app-1",
   								   "labels": {
   								      "app": "foo"
   								   }
   								},
   								"spec": {
   								   "hostnames": [
   								      "foo.com"
   								   ],
   								   "rules": [
   								      {
   								         "matches": [
   								            {
   								               "path": {
   								                  "type": "Prefix",
   								                  "value": "/bar"
   								               }
   								            }
   								         ],
   								         "filters": [
   								            {
   								               "type": "RequestMirror",
   								               "requestMirror": {
   								                  "serviceName": "my-service1-staging",
   								                  "port": 8080
   								               }
   								            }
   								         ],
   								         "forwardTo": [
   								            {
   								               "serviceName": "my-service1",
   								               "port": 8080
   								            }
   								         ]
   								      }
   								   ]
   								}
							},
						"operation": "CREATE"
						}
					}`),
				wantRespCode: http.StatusOK,
				wantSuccessResponse: admission.AdmissionResponse{
					UID:     "7313cd05-eddc-4150-b88c-971a0d53b2ab",
					Allowed: true,
					Result:  &metav1.Status{},
				},
			},
			{
				name: "invalid HTTPRoute resource with two request mirror filters",
				reqBody: dedent.Dedent(`{
						"kind": "AdmissionReview",
						"apiVersion": "` + apiVersion + `",
						"request": {
							"uid": "7313cd05-eddc-4150-b88c-971a0d53b2ab",
							"resource": {
								"group": "networking.x-k8s.io",
								"version": "v1alpha1",
								"resource": "httproutes"
							},
							"object": {
   								"kind": "HTTPRoute",
   								"apiVersion": "networking.x-k8s.io/v1alpha1",
   								"metadata": {
   								   "name": "http-app-1",
   								   "labels": {
   								      "app": "foo"
   								   }
   								},
   								"spec": {
   								   "hostnames": [
   								      "foo.com"
   								   ],
   								   "rules": [
   								      {
   								         "matches": [
   								            {
   								               "path": {
   								                  "type": "Prefix",
   								                  "value": "/bar"
   								               }
   								            }
   								         ],
   								         "filters": [
   								            {
   								               "type": "RequestMirror",
   								               "requestMirror": {
   								                  "serviceName": "my-service1-staging",
   								                  "port": 8080
   								               }
   								            },
   								            {
   								               "type": "RequestMirror",
   								               "requestMirror": {
   								                  "serviceName": "my-service2-staging",
   								                  "port": 8080
   								               }
   								            }
   								         ],
   								         "forwardTo": [
   								            {
   								               "serviceName": "my-service1",
   								               "port": 8080
   								            }
   								         ]
   								      }
   								   ]
   								}
							},
						"operation": "CREATE"
						}
					}`),
				wantRespCode: http.StatusOK,
				wantSuccessResponse: admission.AdmissionResponse{
					UID:     "7313cd05-eddc-4150-b88c-971a0d53b2ab",
					Allowed: false,
					Result: &metav1.Status{
						Code:    400,
						Message: "spec.rules[0].filters: Invalid value: \"RequestMirror\": cannot be used multiple times in the same rule",
					},
				},
			},
			{
				name: "invalid v1alpha2 HTTPRoute resource with two request mirror filters",
				reqBody: dedent.Dedent(`{
						"kind": "AdmissionReview",
						"apiVersion": "` + apiVersion + `",
						"request": {
							"uid": "7313cd05-eddc-4150-b88c-971a0d53b2ab",
							"resource": {
								"group": "networking.x-k8s.io",
								"version": "v1alpha2",
								"resource": "httproutes"
							},
							"object": {
   								"kind": "HTTPRoute",
   								"apiVersion": "networking.x-k8s.io/v1alpha2",
   								"metadata": {
   								   "name": "http-app-1",
   								   "labels": {
   								      "app": "foo"
   								   }
   								},
   								"spec": {
   								   "hostnames": [
   								      "foo.com"
   								   ],
   								   "rules": [
   								      {
   								         "matches": [
   								            {
   								               "path": {
   								                  "type": "Prefix",
   								                  "value": "/bar"
   								               }
   								            }
   								         ],
   								         "filters": [
   								            {
   								               "type": "RequestMirror",
   								               "requestMirror": {
   								                  "serviceName": "my-service1-staging",
   								                  "port": 8080
   								               }
   								            },
   								            {
   								               "type": "RequestMirror",
   								               "requestMirror": {
   								                  "serviceName": "my-service2-staging",
   								                  "port": 8080
   								               }
   								            }
   								         ],
   								         "forwardTo": [
   								            {
   								               "serviceName": "my-service1",
   								               "port": 8080
   								            }
   								         ]
   								      }
   								   ]
   								}
							},
						"operation": "CREATE"
						}
					}`),
				wantRespCode: http.StatusOK,
				wantSuccessResponse: admission.AdmissionResponse{
					UID:     "7313cd05-eddc-4150-b88c-971a0d53b2ab",
					Allowed: false,
					Result: &metav1.Status{
						Code:    400,
						Message: "spec.rules[0].filters: Invalid value: \"RequestMirror\": cannot be used multiple times in the same rule",
					},
				},
			},
			{
				name: "unknown resource under networking.x-k8s.io",
				reqBody: dedent.Dedent(`{
						"kind": "AdmissionReview",
						"apiVersion": "` + apiVersion + `",
						"request": {
							"uid": "7313cd05-eddc-4150-b88c-971a0d53b2ab",
							"resource": {
								"group": "networking.x-k8s.io",
								"version": "v1alpha1",
								"resource": "brokenroutes"
							},
							"object": {
								"apiVersion": "networking.x-k8s.io/v1alpha1",
								"kind": "HTTPRoute"
							},
						"operation": "CREATE"
						}
					}`),
				wantRespCode:       http.StatusInternalServerError,
				wantFailureMessage: "unknown resource 'brokenroutes'\n",
			},
		} {
			tt := tt
			t.Run(fmt.Sprintf("%s/%s", apiVersion, tt.name), func(t *testing.T) {
				assert := assert.New(t)
				res := httptest.NewRecorder()
				handler := http.HandlerFunc(ServeHTTP)

				// send request
				req, err := http.NewRequest("POST", "", bytes.NewBuffer([]byte(tt.reqBody)))
				req = req.WithContext(context.Background())
				assert.Nil(err)
				handler.ServeHTTP(res, req)

				// check response assertions
				assert.Equal(tt.wantRespCode, res.Code)
				if tt.wantRespCode == http.StatusOK {
					var review admission.AdmissionReview
					_, _, err = decoder.Decode(res.Body.Bytes(), nil, &review)
					assert.Nil(err)
					assert.EqualValues(&tt.wantSuccessResponse, review.Response)
				} else {
					assert.Equal(res.Body.String(), tt.wantFailureMessage)
				}
			})
		}
	}
}
