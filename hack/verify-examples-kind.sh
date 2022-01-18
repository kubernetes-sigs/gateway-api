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

readonly GO111MODULE="on"
readonly GOFLAGS="-mod=readonly"
readonly GOPATH="$(mktemp -d)"
readonly CLUSTER_NAME="verify-gateway-api"
readonly KUBECONFIG="${GOPATH}/.kubeconfig"

export GOFLAGS GO111MODULE GOPATH
export PATH="${PATH}:${GOPATH}/bin"

# Cleanup logic for cleanup on exit
CLEANED_UP=false
cleanup() {
  if [ "$CLEANED_UP" = "true" ]; then
    return
  fi
  if [ "${KIND_CREATE_ATTEMPTED:-}" = true ]; then
    kind delete cluster --name "${CLUSTER_NAME}" --kubeconfig "${KUBECONFIG}" || true
  fi
  CLEANED_UP=true
}

trap cleanup INT TERM

# For exit code
res=0

# Install kind
(cd $GOPATH && go get -u sigs.k8s.io/kind) || res=$?

# Create cluster
KIND_CREATE_ATTEMPTED=true
kind create cluster --name "${CLUSTER_NAME}" --kubeconfig "${KUBECONFIG}" || res=$?

for CHANNEL in experimental stable; do
  ##### Test v1alpha2 CRD apply and that invalid examples are invalid.
  # Install CRDs
  kubectl apply --kubeconfig "${KUBECONFIG}" -f "config/crd/${CHANNEL}" || res=$?

  # Temporary workaround for https://github.com/kubernetes/kubernetes/issues/104090
  sleep 8

  # Install all example gateway-api resources.
  kubectl apply --kubeconfig "${KUBECONFIG}" --recursive -f examples/v1alpha2 || res=$?

  # Install all experimental example gateway-api resources when experimental mode is enabled
  if [[ "${CHANNEL}" == "experimental" ]] ; then
      echo "Experimental mode enabled: deploying experimental examples"
      kubectl apply --kubeconfig "${KUBECONFIG}" --recursive -f examples/experimental/v1alpha2 || res=$?
  fi

  # Install invalid gateway-api resources.
  # None of these examples should be successfully configured
  # This is very hacky, sorry.
  # Firstly, apply the examples, remembering that errors are on stdout
  kubectl apply --kubeconfig "${KUBECONFIG}" --recursive -f hack/invalid-examples 2>&1 | \
        # First, we grep out the expected responses.
        # After this, if everything is as expected, the output should be empty.
        grep -v 'is invalid' | \
        grep -v 'missing required field' | \
        # Then, we grep for anything else.
        # If anything else is found, this will return 0
        # which is *not* what we want.
        grep -e '.' && \
        res=2 || \
        echo Examples failed as expected
done

# Clean up and exit
cleanup || res=$?
exit $res
