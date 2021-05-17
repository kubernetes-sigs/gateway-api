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

# Install CRDs
kubectl apply --kubeconfig "${KUBECONFIG}" -f config/crd/bases || res=$?

# Install all example gateway-api resources.
kubectl apply --kubeconfig "${KUBECONFIG}" --recursive -f examples || res=$?

# Clean up and exit
cleanup || res=$?
exit $res
