#!/usr/bin/env bash

# Copyright 2025 The Kubernetes Authors.
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

# https://github.com/kubernetes/code-generator provides generator code to generate a custom typed
# and versioned client for custom API types similar to what https://github.com/kubernetes/client-go
# provides for core types. This script generates such a client for the Gateway API types, in service of any
# projects that need them.

set -o errexit
set -o nounset
set -o pipefail


readonly SCRIPT_ROOT="$(cd "$(dirname "${BASH_SOURCE}")"/.. && pwd)"

if [[ "${VERIFY_CODEGEN:-}" == "true" ]]; then
  echo "Running in verification mode"
  readonly VERIFY_FLAG="--verify-only"
fi

readonly COMMON_FLAGS="${VERIFY_FLAG:-} --go-header-file ${SCRIPT_ROOT}/hack/boilerplate/boilerplate.generatego.txt"

readonly APIS_PKG=sigs.k8s.io/gateway-api
readonly CLIENTSET_NAME=versioned
readonly CLIENTSET_PKG_NAME=clientset
readonly OUTPUT_DIR=pkg/client
readonly OUTPUT_PKG=sigs.k8s.io/gateway-api/pkg/client
readonly API_PATHS=(apis apisx)

GATEWAY_INPUT_DIRS_SPACE=""
GATEWAY_INPUT_DIRS_COMMA=""
GATEWAY_API_DIRS_COMMA=""

for API_PATH in "${API_PATHS[@]}"; do
  VERSIONS=($(find ./${API_PATH} -maxdepth 1 -name "v*" -exec bash -c 'basename {}' \; | LC_ALL=C sort -u))
  for VERSION in "${VERSIONS[@]}"; do
    GATEWAY_INPUT_DIRS_SPACE+="${APIS_PKG}/${API_PATH}/${VERSION} "
    GATEWAY_INPUT_DIRS_COMMA+="${APIS_PKG}/${API_PATH}/${VERSION},"
    GATEWAY_API_DIRS_COMMA+="${API_PATH}/${VERSION},"
  done
done

GATEWAY_INPUT_DIRS_SPACE="${GATEWAY_INPUT_DIRS_SPACE%,}" # drop trailing space
GATEWAY_INPUT_DIRS_COMMA="${GATEWAY_INPUT_DIRS_COMMA%,}" # drop trailing comma
GATEWAY_API_DIRS_COMMA="${GATEWAY_API_DIRS_COMMA%,}" # drop trailing comma

# throw away
new_report="$(mktemp -t "$(basename "$0").api_violations.XXXXXX")"

echo "Generating openapi schema"
go run k8s.io/kube-openapi/cmd/openapi-gen \
  --output-file zz_generated.openapi.go \
  --report-filename "${new_report}" \
  --output-dir "pkg/generated/openapi" \
  --output-pkg "sigs.k8s.io/gateway-api/pkg/generated/openapi" \
  ${COMMON_FLAGS} \
  $GATEWAY_INPUT_DIRS_SPACE \
  k8s.io/apimachinery/pkg/apis/meta/v1 \
  k8s.io/apimachinery/pkg/runtime \
  k8s.io/apimachinery/pkg/version


echo "Generating apply configuration"
go run k8s.io/code-generator/cmd/applyconfiguration-gen \
  --openapi-schema <(go run ${SCRIPT_ROOT}/cmd/modelschema) \
  --output-dir "applyconfiguration" \
  --output-pkg "${APIS_PKG}/applyconfiguration" \
  ${COMMON_FLAGS} \
  ${GATEWAY_INPUT_DIRS_SPACE}

echo "Generating clientset at ${OUTPUT_PKG}/${CLIENTSET_PKG_NAME}"
go run k8s.io/code-generator/cmd/client-gen \
  --clientset-name "${CLIENTSET_NAME}" \
  --input-base "${APIS_PKG}" \
  --input "${GATEWAY_INPUT_DIRS_COMMA//${APIS_PKG}/}" \
  --output-dir "${OUTPUT_DIR}/${CLIENTSET_PKG_NAME}" \
  --output-pkg "${OUTPUT_PKG}/${CLIENTSET_PKG_NAME}" \
  --apply-configuration-package "${APIS_PKG}/applyconfiguration" \
  ${COMMON_FLAGS}

echo "Generating listers at ${OUTPUT_PKG}/listers"
go run k8s.io/code-generator/cmd/lister-gen \
  --output-dir "${OUTPUT_DIR}/listers" \
  --output-pkg "${OUTPUT_PKG}/listers" \
  ${COMMON_FLAGS} \
  ${GATEWAY_INPUT_DIRS_SPACE}

echo "Generating informers"
go run k8s.io/code-generator/cmd/informer-gen \
  --versioned-clientset-package "${OUTPUT_PKG}/${CLIENTSET_PKG_NAME}/${CLIENTSET_NAME}" \
  --listers-package "${OUTPUT_PKG}/listers" \
  --output-dir "${OUTPUT_DIR}/informers" \
  --output-pkg "${OUTPUT_PKG}/informers" \
  ${COMMON_FLAGS} \
  ${GATEWAY_INPUT_DIRS_SPACE}

echo "Generating register helpers"
go run k8s.io/code-generator/cmd/register-gen \
  --output-file zz_generated.register.go \
  ${COMMON_FLAGS} \
  ${GATEWAY_INPUT_DIRS_SPACE}

echo "Generating deepcopy"
go run sigs.k8s.io/controller-tools/cmd/controller-gen \
  object:headerFile=${SCRIPT_ROOT}/hack/boilerplate/boilerplate.generatego.txt \
  paths="./apis/..." \
  paths="./apisx/..."
