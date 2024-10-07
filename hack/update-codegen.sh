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
readonly MIN_REQUIRED_GO_VER="$(go list -m -f '{{.GoVersion}}')"

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
readonly VERSIONS=(v1alpha2 v1alpha3 v1beta1 v1)

GATEWAY_INPUT_DIRS_SPACE=""
GATEWAY_INPUT_DIRS_COMMA=""
for VERSION in "${VERSIONS[@]}"
do
  GATEWAY_INPUT_DIRS_SPACE+="${APIS_PKG}/apis/${VERSION} "
  GATEWAY_INPUT_DIRS_COMMA+="${APIS_PKG}/apis/${VERSION},"
done
GATEWAY_INPUT_DIRS_SPACE="${GATEWAY_INPUT_DIRS_SPACE%,}" # drop trailing space
GATEWAY_INPUT_DIRS_COMMA="${GATEWAY_INPUT_DIRS_COMMA%,}" # drop trailing comma


if [[ "${VERIFY_CODEGEN:-}" == "true" ]]; then
  echo "Running in verification mode"
  readonly VERIFY_FLAG="--verify-only"
fi

readonly COMMON_FLAGS="${VERIFY_FLAG:-} --go-header-file ${SCRIPT_ROOT}/hack/boilerplate/boilerplate.generatego.txt"

echo "Generating CRDs"
go run ./pkg/generator

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
  --output-dir "apis/applyconfiguration" \
  --output-pkg "${APIS_PKG}/apis/applyconfiguration" \
  ${COMMON_FLAGS} \
  ${GATEWAY_INPUT_DIRS_SPACE}


# Temporary hack until https://github.com/kubernetes/kubernetes/pull/124371 is released
function fix_applyconfiguration() {
  local package="$1"
  local version="$(basename $1)"

  echo $package
  echo $version
  pushd $package > /dev/null

  # Replace import
  for filename in *.go; do
    import_line=$(grep "$package" "$filename")
    if [[ -z "$import_line" ]]; then
      continue
    fi
    import_prefix=$(echo "$import_line" | awk '{print $1}')
    sed -i'.bak' -e "s,${import_prefix} \"sigs.k8s.io/gateway-api/${package}\",,g" "$filename"
    sed -i'.bak' -e "s,\[\]${import_prefix}\.,\[\],g" "$filename"
    sed -i'.bak' -e "s,&${import_prefix}\.,&,g" "$filename"
    sed -i'.bak' -e "s,*${import_prefix}\.,*,g" "$filename"
    sed -i'.bak' -e "s,^\t${import_prefix}\.,,g" "$filename"
  done

  rm *.bak
  go fmt .
  find . -type f -name "*.go" -exec sed -i'.bak' -e "s,import (),,g" {} \;
  rm *.bak
  go fmt .

  popd > /dev/null
}

export -f fix_applyconfiguration
find apis/applyconfiguration/apis -name "v*" -type d -exec bash -c 'fix_applyconfiguration $0' {} \;

echo "Generating clientset at ${OUTPUT_PKG}/${CLIENTSET_PKG_NAME}"
go run k8s.io/code-generator/cmd/client-gen \
  --clientset-name "${CLIENTSET_NAME}" \
  --input-base "${APIS_PKG}" \
  --input "${GATEWAY_INPUT_DIRS_COMMA//${APIS_PKG}/}" \
  --output-dir "pkg/client/${CLIENTSET_PKG_NAME}" \
  --output-pkg "${OUTPUT_PKG}/${CLIENTSET_PKG_NAME}" \
  --apply-configuration-package "${APIS_PKG}/apis/applyconfiguration" \
  ${COMMON_FLAGS}

echo "Generating listers at ${OUTPUT_PKG}/listers"
go run k8s.io/code-generator/cmd/lister-gen \
  --output-dir "pkg/client/listers" \
  --output-pkg "${OUTPUT_PKG}/listers" \
  ${COMMON_FLAGS} \
  ${GATEWAY_INPUT_DIRS_SPACE}

echo "Generating informers at ${OUTPUT_PKG}/informers"
go run k8s.io/code-generator/cmd/informer-gen \
  --versioned-clientset-package "${OUTPUT_PKG}/${CLIENTSET_PKG_NAME}/${CLIENTSET_NAME}" \
  --listers-package "${OUTPUT_PKG}/listers" \
  --output-dir "pkg/client/informers" \
  --output-pkg "${OUTPUT_PKG}/informers" \
  ${COMMON_FLAGS} \
  ${GATEWAY_INPUT_DIRS_SPACE}

echo "Generating ${VERSION} register at ${APIS_PKG}/apis/${VERSION}"
go run k8s.io/code-generator/cmd/register-gen \
  --output-file zz_generated.register.go \
  ${COMMON_FLAGS} \
  ${GATEWAY_INPUT_DIRS_SPACE}

for VERSION in "${VERSIONS[@]}"
do
  echo "Generating ${VERSION} deepcopy at ${APIS_PKG}/apis/${VERSION}"
  go run sigs.k8s.io/controller-tools/cmd/controller-gen \
    object:headerFile=${SCRIPT_ROOT}/hack/boilerplate/boilerplate.generatego.txt \
    paths="${APIS_PKG}/apis/${VERSION}"
done

echo "Generating gRPC/Protobuf code"

readonly PROTOC_CACHE_DIR="/tmp/protoc.cache"
readonly PROTOC_BINARY="${PROTOC_CACHE_DIR}/bin/protoc"
readonly PROTOC_VERSION="22.2"
readonly PROTOC_REPO="https://github.com/protocolbuffers/protobuf"

readonly PROTOC_LINUX_X86_URL="${PROTOC_REPO}/releases/download/v${PROTOC_VERSION}/protoc-${PROTOC_VERSION}-linux-x86_64.zip"
readonly PROTOC_LINUX_X86_CHECKSUM="4805ba56594556402a6c327a8d885a47640ee363  ${PROTOC_BINARY}"

readonly PROTOC_LINUX_ARM64_URL="${PROTOC_REPO}/releases/download/v${PROTOC_VERSION}/protoc-${PROTOC_VERSION}-linux-aarch_64.zip"
readonly PROTOC_LINUX_ARM64_CHECKSUM="47285b2386f990da319e9eef92cadec2dfa28733  ${PROTOC_BINARY}"

readonly PROTOC_MAC_UNIVERSAL_URL="${PROTOC_REPO}/releases/download/v${PROTOC_VERSION}/protoc-${PROTOC_VERSION}-osx-universal_binary.zip"
readonly PROTOC_MAC_UNIVERSAL_CHECKSUM="2a79d0eb235c808eca8de893762072b94dc6144c  ${PROTOC_BINARY}"

PROTOC_URL=""
PROTOC_CHECKSUM=""

ARCH=$(uname -m)
OS=$(uname)

if [[ "${OS}" != "Linux" ]] && [[ "${OS}" != "Darwin" ]]; then
  echo "Unsupported operating system ${OS}" >/dev/stderr
  exit 1
fi

if [[ "${OS}" == "Linux" ]]; then
  if [[ "$ARCH" == "x86_64" ]]; then
    PROTOC_URL="$PROTOC_LINUX_X86_URL"
    PROTOC_CHECKSUM="$PROTOC_LINUX_X86_CHECKSUM"
  elif [[ "$ARCH" == "arm64" ]]; then
    PROTOC_URL="$PROTOC_LINUX_ARM64_URL"
    PROTOC_CHECKSUM="$PROTOC_LINUX_ARM64_CHECKSUM"
  elif [[ "$ARCH" == "aarch64" ]]; then
    PROTOC_URL="$PROTOC_LINUX_ARM64_URL"
    PROTOC_CHECKSUM="$PROTOC_LINUX_ARM64_CHECKSUM"
  else
    echo "Architecture ${ARCH} is not supported on OS ${OS}." >/dev/stderr
    exit 1
  fi
elif [[ "${OS}" == "Darwin" ]]; then
    PROTOC_URL="$PROTOC_MAC_UNIVERSAL_URL"
    PROTOC_CHECKSUM="$PROTOC_MAC_UNIVERSAL_CHECKSUM"
fi

function verify_protoc {
  if ! echo "${PROTOC_CHECKSUM}" | shasum -c >/dev/null; then
    echo "Downloaded protoc binary failed checksum." >/dev/stderr
    exit 1
  fi
}

function ensure_protoc {
  mkdir -p "${PROTOC_CACHE_DIR}"
  if [ ! -f "${PROTOC_BINARY}" ]; then
    curl -sL -o "${PROTOC_CACHE_DIR}/protoc.zip" "${PROTOC_URL}"
    unzip -d "${PROTOC_CACHE_DIR}" "${PROTOC_CACHE_DIR}/protoc.zip"
  fi
  verify_protoc
}

ensure_protoc
go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.28
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.2

(cd conformance/echo-basic && \
  export PATH="$PATH:$GOPATH/bin" && \
  "${PROTOC_BINARY}" --go_out=grpcechoserver --go_opt=paths=source_relative \
    --go-grpc_out=grpcechoserver --go-grpc_opt=paths=source_relative \
    grpcecho.proto
)
