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

# For exit code
res=0

# Install kind
(cd $GOPATH && go install sigs.k8s.io/kind@v0.20.0) || res=$?

# Create cluster
KIND_CREATE_ATTEMPTED=true
kind create cluster --name "${CLUSTER_NAME}"

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
