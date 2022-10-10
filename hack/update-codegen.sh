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

readonly SCRIPT_ROOT="$(cd "$(dirname "${BASH_SOURCE}")"/.. && pwd)"

# Keep outer module cache so we don't need to redownload them each time.
# The build cache already is persisted.
readonly GOMODCACHE="$(go env GOMODCACHE)"
readonly GO111MODULE="on"
readonly GOFLAGS="-mod=readonly"
readonly GOPATH="$(mktemp -d)"
readonly MIN_REQUIRED_GO_VER="1.19"

function go_version_matches {
  go version | perl -ne "exit 1 unless m{go version go([0-9]+.[0-9]+)}; exit 1 if (\$1 < ${MIN_REQUIRED_GO_VER})"
  return $?
}

if ! go_version_matches; then
  echo "Go v${MIN_REQUIRED_GO_VER} or later is required to run code generation"
  exit 1
fi

export GOMODCACHE GO111MODULE GOFLAGS GOPATH

# Even when modules are enabled, the code-generator tools always write to
# a traditional GOPATH directory, so fake on up to point to the current
# workspace.
mkdir -p "$GOPATH/src/sigs.k8s.io"
ln -s "${SCRIPT_ROOT}" "$GOPATH/src/sigs.k8s.io/gateway-api"

readonly OUTPUT_PKG=sigs.k8s.io/gateway-api/pkg/client
readonly APIS_PKG=sigs.k8s.io/gateway-api
readonly CLIENTSET_NAME=versioned
readonly CLIENTSET_PKG_NAME=clientset

if [[ "${VERIFY_CODEGEN:-}" == "true" ]]; then
  echo "Running in verification mode"
  readonly VERIFY_FLAG="--verify-only"
fi

readonly COMMON_FLAGS="${VERIFY_FLAG:-} --go-header-file ${SCRIPT_ROOT}/hack/boilerplate/boilerplate.generatego.txt"

echo "Generating CRDs"
go run ./pkg/generator

echo "Generating clientset at ${OUTPUT_PKG}/${CLIENTSET_PKG_NAME}"
go run k8s.io/code-generator/cmd/client-gen \
  --clientset-name "${CLIENTSET_NAME}" \
  --input-base "" \
  --input "${APIS_PKG}/apis/v1alpha2,${APIS_PKG}/apis/v1beta1" \
  --output-package "${OUTPUT_PKG}/${CLIENTSET_PKG_NAME}" \
  ${COMMON_FLAGS}

echo "Generating listers at ${OUTPUT_PKG}/listers"
go run k8s.io/code-generator/cmd/lister-gen \
  --input-dirs "${APIS_PKG}/apis/v1alpha2,${APIS_PKG}/apis/v1beta1" \
  --output-package "${OUTPUT_PKG}/listers" \
  ${COMMON_FLAGS}

echo "Generating informers at ${OUTPUT_PKG}/informers"
go run k8s.io/code-generator/cmd/informer-gen \
  --input-dirs "${APIS_PKG}/apis/v1alpha2,${APIS_PKG}/apis/v1beta1" \
  --versioned-clientset-package "${OUTPUT_PKG}/${CLIENTSET_PKG_NAME}/${CLIENTSET_NAME}" \
  --listers-package "${OUTPUT_PKG}/listers" \
  --output-package "${OUTPUT_PKG}/informers" \
  ${COMMON_FLAGS}

for VERSION in v1alpha2 v1beta1
do
  echo "Generating ${VERSION} register at ${APIS_PKG}/apis/${VERSION}"
  go run k8s.io/code-generator/cmd/register-gen \
    --input-dirs "${APIS_PKG}/apis/${VERSION}" \
    --output-package "${APIS_PKG}/apis/${VERSION}" \
    ${COMMON_FLAGS}

  echo "Generating ${VERSION} deepcopy at ${APIS_PKG}/apis/${VERSION}"
  go run sigs.k8s.io/controller-tools/cmd/controller-gen \
    object:headerFile=${SCRIPT_ROOT}/hack/boilerplate/boilerplate.generatego.txt \
    paths="${APIS_PKG}/apis/${VERSION}" 

done
