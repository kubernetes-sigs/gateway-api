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
readonly LOCAL_IMAGE="registry.k8s.io/gateway-api/admission-server:latest"

export KUBECONFIG="${GOPATH}/.kubeconfig"
export GOFLAGS GO111MODULE GOPATH
export PATH="${GOPATH}/bin:${PATH}"

# Cleanup logic for cleanup on exit
CLEANED_UP=false
cleanup() {
  if [ "$CLEANED_UP" = "true" ]; then
    return
  fi

  rm config/webhook/kustomization.yaml

  if [ "${KIND_CREATE_ATTEMPTED:-}" = true ]; then
    kind delete cluster --name "${CLUSTER_NAME}" || true
  fi
  CLEANED_UP=true
}

trap cleanup INT TERM

# For exit code
res=0

# Install kind
(cd $GOPATH && go install sigs.k8s.io/kind@v0.17.0) || res=$?

# Create cluster
KIND_CREATE_ATTEMPTED=true
kind create cluster --name "${CLUSTER_NAME}" || res=$?

cat <<EOF >config/webhook/kustomization.yaml
resources:
  - 0-namespace.yaml
  - certificate_config.yaml
  - admission_webhook.yaml
patches:
  - patch: |-
      - op: replace
        path: /spec/template/spec/containers/0/image
        value: ${LOCAL_IMAGE}
      - op: replace
        path: /spec/template/spec/containers/0/imagePullPolicy
        value: IfNotPresent
    target:
      group: apps
      version: v1
      kind: Deployment
      name: gateway-api-admission-server
EOF

# Install webhook
docker build -t ${LOCAL_IMAGE} -f docker/Dockerfile.webhook .
kind load docker-image ${LOCAL_IMAGE} --name "${CLUSTER_NAME}"
kubectl apply -k config/webhook/

# Wait for webhook to be ready
for check in {1..10}; do
  sleep 5
  NUM_COMPLETED=$(kubectl get po -n gateway-system | grep Completed | wc -l | xargs || echo Failed to get completed Pods)
  if [ "${NUM_COMPLETED}" = "2" ]; then
    echo "Webhook successfully configured"
    break
  elif [ "${check}" = "10" ]; then
    echo "Timed out waiting for webhook setup to complete"
    cleanup
    exit 1
  fi
  echo "Webhook not ready yet, will check again in 5 seconds"
done

for CHANNEL in experimental standard; do
  ##### Test v1alpha2 CRD apply and that invalid examples are invalid.
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

  # Install invalid gateway-api resources.
  # None of these examples should be successfully configured
  # This is very hacky, sorry.
  # Firstly, apply the examples, remembering that errors are on stdout
  kubectl apply --recursive -f hack/invalid-examples 2>&1 |
    # First, we grep out the expected responses.
    # After this, if everything is as expected, the output should be empty.
    grep -v 'is invalid' |
    grep -v 'missing required field' |
    grep -v 'denied the request' |
    # Then, we grep for anything else.
    # If anything else is found, this will return 0
    # which is *not* what we want.
    grep -e '.' &&
    res=2 ||
    echo Examples failed as expected
done

# Clean up and exit
cleanup || res=$?
exit $res
