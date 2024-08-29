#!/bin/bash

# Copyright 2024 The Kubernetes Authors.
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

# This file contains code derived from Envoy Gateway,
# https://github.com/envoyproxy/gateway
# from the source file
# https://github.com/envoyproxy/gateway/blob/main/tools/hack/create-cluster.sh
# and is provided here subject to the following:
# Copyright Envoy Gateway Authors
# SPDX-License-Identifier: Apache-2.0

set -euo pipefail

# Setup default values
CLUSTER_NAME=${CLUSTER_NAME:-"envoy-gateway"}
METALLB_VERSION=${METALLB_VERSION:-"v0.13.10"}
KIND_NODE_TAG=${KIND_NODE_TAG:-"v1.28.0"}
NUM_WORKERS=${NUM_WORKERS:-""}


KIND_CFG=$(cat <<-EOM
kind: Cluster
apiVersion: kind.x-k8s.io/v1alpha4
nodes:
- role: control-plane
EOM
)

# https://kind.sigs.k8s.io/docs/user/quick-start/#multi-node-clusters
if [[ -n "${NUM_WORKERS}" ]]; then
for _ in $(seq 1 "${NUM_WORKERS}"); do
  KIND_CFG+=$(printf "\n%s" "- role: worker")
done
fi

## Check if kind cluster already exists.
if kind get clusters | grep -q "${CLUSTER_NAME}"; then
  echo "Cluster ${CLUSTER_NAME} already exists."
else
## Create kind cluster.
if [[ -z "${KIND_NODE_TAG}" ]]; then
  cat << EOF | kind create cluster --name "${CLUSTER_NAME}" --config -
${KIND_CFG}
EOF
else
  cat << EOF | kind create cluster --image "kindest/node:${KIND_NODE_TAG}" --name "${CLUSTER_NAME}" --config -
${KIND_CFG}
EOF
fi
fi


## Install MetalLB.
kubectl apply -f https://raw.githubusercontent.com/metallb/metallb/"${METALLB_VERSION}"/config/manifests/metallb-native.yaml
needCreate="$(kubectl get secret -n metallb-system memberlist --no-headers --ignore-not-found -o custom-columns=NAME:.metadata.name)"
if [ -z "$needCreate" ]; then
    kubectl create secret generic -n metallb-system memberlist --from-literal=secretkey="$(openssl rand -base64 128)"
fi

# Wait for MetalLB to become available.
kubectl rollout status -n metallb-system deployment/controller --timeout 5m
kubectl rollout status -n metallb-system daemonset/speaker --timeout 5m

# Apply config with addresses based on docker network IPAM.
subnet=$(docker network inspect kind | jq -r '.[].IPAM.Config[].Subnet | select(contains(":") | not)')
# Assume default kind network subnet prefix of 16, and choose addresses in that range.
address_first_octets=$(echo "${subnet}" | awk -F. '{printf "%s.%s",$1,$2}')
address_range="${address_first_octets}.255.200-${address_first_octets}.255.250"
kubectl apply -f - <<EOF
apiVersion: metallb.io/v1beta1
kind: IPAddressPool
metadata:
  namespace: metallb-system
  name: kube-services
spec:
  addresses:
  - ${address_range}
---
apiVersion: metallb.io/v1beta1
kind: L2Advertisement
metadata:
  name: kube-services
  namespace: metallb-system
spec:
  ipAddressPools:
  - kube-services
EOF
