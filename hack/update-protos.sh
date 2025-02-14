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

set -o errexit
set -o nounset
set -o pipefail

readonly SCRIPT_ROOT="$(cd "$(dirname "${BASH_SOURCE}")"/.. && pwd)"

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
