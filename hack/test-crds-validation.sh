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

set -o nounset
set -o pipefail

readonly GO111MODULE="on"
readonly GOFLAGS="-mod=readonly"
readonly GOPATH="$(mktemp -d)"
readonly CLUSTER_NAME="verify-gateway-api"

export KUBECONFIG="${GOPATH}/.kubeconfig"
export GOFLAGS GO111MODULE GOPATH
export PATH="${GOPATH}/bin:${PATH}"

# Cleanup logic for cleanup on exit
CLEANED_UP=false
cleanup() {
  if [ "$CLEANED_UP" = "true" ]; then
    return
  fi

  if [ "${KIND_CREATE_ATTEMPTED:-}" = true ]; then
    kind delete cluster --name "${CLUSTER_NAME}" || true
  fi
  CLEANED_UP=true
}

trap cleanup INT TERM EXIT

# TODO(mlavacca): find a good way to keep this dependency up to date.
KIND_VERSION="v0.26.0"

# list of kind images taken from https://github.com/kubernetes-sigs/kind/releases/tag/v0.26.0.
# they need to be updated when kind is updated.
KIND_IMAGES=(
  "kindest/node:v1.28.15@sha256:a7c05c7ae043a0b8c818f5a06188bc2c4098f6cb59ca7d1856df00375d839251"
  "kindest/node:v1.29.12@sha256:62c0672ba99a4afd7396512848d6fc382906b8f33349ae68fb1dbfe549f70dec"
  "kindest/node:v1.30.8@sha256:17cd608b3971338d9180b00776cb766c50d0a0b6b904ab4ff52fd3fc5c6369bf"
  "kindest/node:v1.31.4@sha256:2cb39f7295fe7eafee0842b1052a599a4fb0f8bcf3f83d96c7f4864c357c6c30"
  "kindest/node:v1.32.0@sha256:2458b423d635d7b01637cac2d6de7e1c1dca1148a2ba2e90975e214ca849e7cb"
)

if [ "$#" -gt 1 ]; then
    echo "Error: Too many arguments provided. Only 1 argument is allowed."
    exit 1
fi

DEFAULT_INDEX=$((1))

if [ "$#" -eq 1 ]; then
  # Check if the argument is a valid number between 1 and 5
  if ! [[ "$1" =~ ^[1-5] ]]; then
      echo "Error: Argument is not a valid integer between 1 and 5."
      exit 1
  fi
  INDEX=$(($1))
else
  INDEX=$((DEFAULT_INDEX))
fi

K8S_IMAGE=${KIND_IMAGES[$((INDEX-1))]}
echo "Using Kubernetes image: ${K8S_IMAGE}"

# For exit code
res=0

# Install kind
(cd "${GOPATH}" && go install sigs.k8s.io/kind@${KIND_VERSION}) || res=$?

# Create cluster
KIND_CREATE_ATTEMPTED=true
kind create cluster --name "${CLUSTER_NAME}" --image "${K8S_IMAGE}" || res=$?

# Verify CEL validation
for CHANNEL in experimental standard; do
  # Install CRDs.
  kubectl apply -f "config/crd/${CHANNEL}/gateway*.yaml"

  # Run tests.
  go test -v -timeout=120s -count=1 --tags ${CHANNEL} sigs.k8s.io/gateway-api/pkg/test/cel || res=$?

  # Delete CRDs to reset environment.
  kubectl delete -f "config/crd/${CHANNEL}/gateway*.yaml"
done

# Temporary workaround for https://github.com/kubernetes/kubernetes/issues/104090
sleep 8

## Validate example YAMLs for each channel

for CHANNEL in experimental standard; do
  ##### Test valid CRD apply and that invalid examples are invalid.
  # Install CRDs
  kubectl apply -f "config/crd/${CHANNEL}/gateway*.yaml" || res=$?

  # Temporary workaround for https://github.com/kubernetes/kubernetes/issues/104090
  sleep 8

  kubectl apply --recursive -f examples/standard || res=$?

  # Install all experimental example gateway-api resources when experimental mode is enabled
  if [[ "${CHANNEL}" == "experimental" ]]; then
    echo "Experimental mode enabled: deploying experimental examples"
    kubectl apply --recursive -f examples/experimental || res=$?
  fi

  # Find all our invalid examples and check them one by one.
  # This lets us check the output in a cleaner way than a grep pipeline.
  for file in $(find hack/invalid-examples -name "*.yaml"); do
    # Don't check alpha resources in Standard checks
    if [[ "$file" =~ "experimental" && "$CHANNEL" == "standard" ]]; then
      continue
    fi

    KUBECTL_OUTPUT=$(kubectl apply -f "$file" 2>&1)

    if [[ \
          ! ("$KUBECTL_OUTPUT" =~ "is invalid") && \
          ! ("$KUBECTL_OUTPUT" =~ "missing required field") &&  \
          ! ("$KUBECTL_OUTPUT" =~ "denied the request") && \
          ! ("$KUBECTL_OUTPUT" =~ "Invalid value") \
          ]]; then
      res=2
      cat<<EOF

Error: Example $file in channel $CHANNEL failed in an unexpected way with CEL validation.
$KUBECTL_OUTPUT
EOF
    else
    echo "Example $file in channel $CHANNEL failed as expected with CEL validation."
    fi

  done
  kubectl delete -f "config/crd/${CHANNEL}/gateway*.yaml" || res=$?
done

### We've trapped EXIT with cleanup(), so just exit with what we've got.
exit $res
