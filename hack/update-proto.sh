#!/bin/bash

# Copyright 2020 The Kubernetes Authors.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

set -o errexit
set -o nounset
set -o pipefail

go install k8s.io/code-generator/cmd/go-to-protobuf/protoc-gen-gogo

# Generate the protos
# TODO(https://github.com/kubernetes-sigs/service-apis/issues/79): replace package name (k8s.io) with sigs.k8s.io
go run k8s.io/code-generator/cmd/go-to-protobuf \
    --proto-import=$PWD/third_party/protobuf \
    --packages sigs.k8s.io/service-apis/apis/v1alpha1=k8s.io.service_apis.api.v1alpha1,-sigs.k8s.io/controller-runtime/pkg/scheme \
    --drop-embedded-fields k8s.io/apimachinery/pkg/runtime.SchemeBuilder,sigs.k8s.io/controller-runtime/pkg/scheme \
    --apimachinery-packages -k8s.io/apimachinery/pkg/runtime/schema,-k8s.io/apimachinery/pkg/runtime,-k8s.io/apimachinery/pkg/apis/meta/v1,-k8s.io/api/core/v1 \
    --go-header-file hack/boilerplate.go.txt
