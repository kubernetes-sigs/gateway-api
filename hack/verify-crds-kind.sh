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

  rm -f config/webhook/kustomization.yaml

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

# Verify CEL validations before installing webhook.
for CHANNEL in experimental standard; do
  # Install CRDs.
  kubectl apply -f "config/crd/${CHANNEL}/gateway*.yaml"

  # Run tests.
  go test -v -timeout=120s -count=1 --tags ${CHANNEL} sigs.k8s.io/gateway-api/pkg/test/cel

  # Delete CRDs to reset environment.
  kubectl delete -f "config/crd/${CHANNEL}/gateway*.yaml"
done

# Temporary workaround for https://github.com/kubernetes/kubernetes/issues/104090
sleep 8

## Check using example YAMLs as well
## with _only_ CEL validation


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

###
# This section and below can be REMOVED once the webhook is removed.
###
# Install webhook and check the _invalid_ examples again.
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
  ##### Test valid CRD apply and that invalid examples are invalid.
  # Install CRDs
  kubectl apply -f "config/crd/${CHANNEL}/gateway*.yaml" || res=$?

  # Temporary workaround for https://github.com/kubernetes/kubernetes/issues/104090
  sleep 8

  # Note that we skip the working examples since we did them already with
  # just CEL validation.

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

Error: Example $file in channel $CHANNEL failed in an unexpected way with webhook validation.
$KUBECTL_OUTPUT
EOF
    else
    echo "Example $file in channel $CHANNEL failed as expected with webhook validation."
    fi

  done
  kubectl delete -f "config/crd/${CHANNEL}/gateway*.yaml" || res=$?
done

# We've trapped EXIT with cleanup(), so just exit with what we've got.
exit $res
