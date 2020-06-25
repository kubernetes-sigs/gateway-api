#!/usr/bin/env bash

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

SCRIPT_ROOT=$(dirname "${BASH_SOURCE}")/..

go install k8s.io/code-generator/cmd/{client-gen,lister-gen,informer-gen,deepcopy-gen,register-gen}

# Go installs the above commands to get installed in $GOBIN if defined, and $GOPATH/bin otherwise:
GOBIN="$(go env GOBIN)"
gobin="${GOBIN:-$(go env GOPATH)/bin}"

OUTPUT_PKG=sigs.k8s.io/service-apis/pkg/client
FQ_APIS=sigs.k8s.io/service-apis/api/v1alpha1
APIS_PKG=sigs.k8s.io/service-apis
CLIENTSET_NAME=versioned
CLIENTSET_PKG_NAME=clientset

if [[ "${VERIFY_CODEGEN:-}" == "true" ]]; then
  echo "Running in verification mode"
  VERIFY_FLAG="--verify-only"
fi
COMMON_FLAGS="${VERIFY_FLAG:-} --go-header-file ${SCRIPT_ROOT}/hack/boilerplate.go.txt"

echo "Generating deepcopy funcs"
"${gobin}/deepcopy-gen" --input-dirs "${FQ_APIS}" -O zz_generated.deepcopy --bounding-dirs "${APIS_PKG}" ${COMMON_FLAGS}

echo "Generating clientset at ${OUTPUT_PKG}/${CLIENTSET_PKG_NAME}"
"${gobin}/client-gen" --clientset-name "${CLIENTSET_NAME}" --input-base "" --input-dirs "${FQ_APIS}" --output-package "${OUTPUT_PKG}/${CLIENTSET_PKG_NAME}" ${COMMON_FLAGS}

echo "Generating listers at ${OUTPUT_PKG}/listers"
"${gobin}/lister-gen" --input-dirs "${FQ_APIS}" --output-package "${OUTPUT_PKG}/listers" ${COMMON_FLAGS}

echo "Generating informers at ${OUTPUT_PKG}/informers"
"${gobin}/informer-gen" \
         --input-dirs "${FQ_APIS}" \
         --versioned-clientset-package "${OUTPUT_PKG}/${CLIENTSET_PKG_NAME}/${CLIENTSET_NAME}" \
         --listers-package "${OUTPUT_PKG}/listers" \
         --output-package "${OUTPUT_PKG}/informers" \
         ${COMMON_FLAGS}

echo "Generating register at ${FQ_APIS}"
"${gobin}/register-gen" --output-package "${FQ_APIS}" --input-dirs ${FQ_APIS} ${COMMON_FLAGS}
