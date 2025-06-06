#!/bin/sh

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

# Set these as needed.
LINKERD_VERSION=${LINKERD_VERSION:-version-2.18}
LINKERD_EDGE_VERSION=${LINKERD_EDGE_VERSION:-edge-25.4.4}
GATEWAY_API_CHANNEL=${GATEWAY_API_CHANNEL:-standard}
GATEWAY_API_VERSION=${GATEWAY_API_VERSION:-v1.3.0}

UNSUPPORTED_FEATURES="MeshHTTPRouteRedirectPath,MeshHTTPRouteRewritePath"

CONFORMANCE_PRODUCT=linkerd-linkerd
CONFORMANCE_VERSION=$(echo $LINKERD_VERSION | cut -d- -f2-)
GATEWAY_API_BASE_URL=https://github.com/kubernetes-sigs/gateway-api/releases/download

echo "Using Linkerd version $LINKERD_VERSION (AKA $LINKERD_EDGE_VERSION)"
echo "Using Gateway API $GATEWAY_API_VERSION $GATEWAY_API_CHANNEL"

# Install the Linkerd CLI.
curl --proto '=https' --tlsv1.2 -sSfL https://run.linkerd.io/install-edge \
  | env LINKERD2_VERSION=${LINKERD_EDGE_VERSION} sh

export PATH=$HOME/.linkerd2/bin:$PATH

# Install the Gateway API CRDs.

kubectl apply -f ${GATEWAY_API_BASE_URL}/${GATEWAY_API_VERSION}/${GATEWAY_API_CHANNEL}-install.yaml

# Install the Linkerd control plane.
linkerd install --crds | kubectl apply -f -
linkerd install | kubectl apply -f -
linkerd check

# Run the conformance tests.

REPORT_NAME=${GATEWAY_API_CHANNEL}-${CONFORMANCE_VERSION}-default-report.yaml
REPORT_PATH=reports/${GATEWAY_API_VERSION}/${CONFORMANCE_PRODUCT}/${REPORT_NAME}

go test \
    -p 4 \
    ./conformance \
    -run TestConformance \
    -args \
        --organization Linkerd \
        --project Linkerd \
        --url https://github.com/linkerd/linkerd2 \
        --version ${LINKERD_VERSION} \
        --contact https://github.com/linkerd/linkerd2/blob/main/MAINTAINERS.md \
        --report-output ${REPORT_PATH} \
        --conformance-profiles=MESH-HTTP,MESH-GRPC \
        --all-features \
        --exempt-features=Gateway,ReferenceGrant,${UNSUPPORTED_FEATURES} \
        --namespace-annotations=linkerd.io/inject=enabled
