#!/usr/bin/env bash

# Copyright 2023 The Kubernetes Authors.
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

# TODO(https://github.com/kubernetes-sigs/gateway-api/issues/2239): Make this a
# "verify" script (so that it runs along with the verification presubmit) once
# we are confident about these tests.

set -eo pipefail

readonly GOFLAGS="-mod=readonly"
readonly GO111MODULE="on"
readonly GOPATH="$(mktemp -d)"
readonly CLUSTER_NAME="verify-gateway-api-validation"

export KUBECONFIG="${GOPATH}/.kubeconfig" # Do not modify the users kube-context by creating a separate kubeconfig
export GOFLAGS GO111MODULE GOPATH
export PATH="${GOPATH}/bin:${PATH}"

cleanup() {
  kind delete cluster --name "${CLUSTER_NAME}" || true
}

main() {
  # Install KIND and create a KIND cluster.
  (cd $GOPATH && go install sigs.k8s.io/kind@v0.20.0)
  kind create cluster --name "${CLUSTER_NAME}"

  # Install Gateway CRDs.
  #
  # It's expected that `make generate` has already been run to generate the
  # lastest CRD yamls.
  kubectl apply -k config/crd

  # Run tests.
  go test -timeout 120s sigs.k8s.io/gateway-api/hack/cel-validation

  cleanup
}

exit_handler() {
  if (($? != 0)); then
    cleanup
    echo "FAILED"
    exit 1
  fi
}

trap exit_handler EXIT

main "${@}"
