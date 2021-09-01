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

readonly GO111MODULE="on"
readonly GOFLAGS="-mod=readonly"
readonly GOPATH="$(mktemp -d)"

export GO111MODULE GOFLAGS GOPATH

# Even when modules are enabled, the code-generator tools always write to
# a traditional GOPATH directory, so fake on up to point to the current
# workspace.
mkdir -p "$GOPATH/src/sigs.k8s.io"
ln -s "${SCRIPT_ROOT}" "$GOPATH/src/sigs.k8s.io/gateway-api"

readonly OUTPUT_PKG=sigs.k8s.io/gateway-api/pkg/client
readonly FQ_APIS=sigs.k8s.io/gateway-api/apis/v1alpha1,sigs.k8s.io/gateway-api/apis/v1alpha2
readonly APIS_PKG=sigs.k8s.io/gateway-api
readonly CLIENTSET_NAME=versioned
readonly CLIENTSET_PKG_NAME=clientset

if [[ "${VERIFY_CODEGEN:-}" == "true" ]]; then
  echo "Running in verification mode"
  readonly VERIFY_FLAG="--verify-only"
fi

readonly COMMON_FLAGS="${VERIFY_FLAG:-} --go-header-file ${SCRIPT_ROOT}/hack/boilerplate/boilerplate.generatego.txt"

echo "Generating v1alpha1 CRDs and deepcopy"
go run sigs.k8s.io/controller-tools/cmd/controller-gen \
        object:headerFile=./hack/boilerplate/boilerplate.generatego.txt \
        crd:crdVersions=v1 \
        output:crd:artifacts:config=config/crd/v1alpha1 \
        paths=./apis/v1alpha1

echo "Generating v1alpha2 CRDs and deepcopy"
go run sigs.k8s.io/controller-tools/cmd/controller-gen \
        object:headerFile=./hack/boilerplate/boilerplate.generatego.txt \
        crd:crdVersions=v1 \
        output:crd:artifacts:config=config/crd/v1alpha2 \
        paths=./apis/v1alpha2

# TODO(robscott): Change this once v1alpha2 has received formal API approval.
sed -i -e 's/controller\-gen\.kubebuilder\.io\/version\:\ v0\.6\.2/api\-approved\.kubernetes\.io\:\ unapproved/g' config/crd/v1alpha2/gateway.networking.k8s.io*

for VERSION in v1alpha1 v1alpha2
do
        GROUP="gateway"
        if [[ "${VERSION}" == "v1alpha1" ]]; then
                GROUP="networking"
        fi
        echo "Generating ${VERSION} clientset at ${OUTPUT_PKG}/${CLIENTSET_PKG_NAME}/${GROUP}"
        go run k8s.io/code-generator/cmd/client-gen \
                --clientset-name "${CLIENTSET_NAME}" \
                --input-base "" \
                --input "${APIS_PKG}/apis/${VERSION}" \
                --output-package "${OUTPUT_PKG}/${CLIENTSET_PKG_NAME}/${GROUP}" \
                ${COMMON_FLAGS}

        echo "Generating ${VERSION} listers at ${OUTPUT_PKG}/listers/${GROUP}"
        go run k8s.io/code-generator/cmd/lister-gen \
                --input-dirs "${APIS_PKG}/apis/${VERSION}" \
                --output-package "${OUTPUT_PKG}/listers/${GROUP}" \
                ${COMMON_FLAGS}

        echo "Generating ${VERSION} informers at ${OUTPUT_PKG}/informers/${GROUP}"
        go run k8s.io/code-generator/cmd/informer-gen \
                --input-dirs "${APIS_PKG}/apis/${VERSION}" \
                --versioned-clientset-package "${OUTPUT_PKG}/${CLIENTSET_PKG_NAME}/${GROUP}/${CLIENTSET_NAME}" \
                --listers-package "${OUTPUT_PKG}/listers/${GROUP}" \
                --output-package "${OUTPUT_PKG}/informers/${GROUP}" \
                ${COMMON_FLAGS}

        echo "Generating ${VERSION} register at ${APIS_PKG}/apis/${VERSION}"
        go run k8s.io/code-generator/cmd/register-gen \
                --input-dirs "${APIS_PKG}/apis/${VERSION}" \
                --output-package "${APIS_PKG}/apis/${VERSION}" \
                ${COMMON_FLAGS}

done
