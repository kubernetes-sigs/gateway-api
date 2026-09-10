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
LINKERD_VERSION=${LINKERD_VERSION:-2.20}
LINKERD_EDGE_VERSION=${LINKERD_EDGE_VERSION:-edge-26.6.3}

GATEWAY_API_CHANNEL=${GATEWAY_API_CHANNEL:-standard}
GATEWAY_API_VERSION=${GATEWAY_API_VERSION:-v1.6.2}

UNSUPPORTED_FEATURES="MeshHTTPRouteRedirectPath,MeshHTTPRouteRewritePath,HTTPRoute303RedirectStatusCode,HTTPRoute307RedirectStatusCode,HTTPRoute308RedirectStatusCode"
SKIP_TESTS="MeshGRPCRouteRequestHeaderModifier"

CONFORMANCE_PRODUCT=linkerd-linkerd
CONFORMANCE_VERSION=$(echo $LINKERD_VERSION | cut -d- -f2-)
GATEWAY_API_BASE_URL=https://github.com/kubernetes-sigs/gateway-api/releases/download

echo "Using Linkerd version $LINKERD_VERSION (AKA $LINKERD_EDGE_VERSION)"
echo "Using Gateway API $GATEWAY_API_VERSION $GATEWAY_API_CHANNEL"

# # Install the Linkerd CLI.
# curl --proto '=https' --tlsv1.2 -sSfL https://run.linkerd.io/install-edge \
#   | env LINKERD2_VERSION=${LINKERD_EDGE_VERSION} sh

export PATH=$HOME/.linkerd2/bin:$PATH

installed_version=$(linkerd version --client --proxy | awk '{ print $NF }')

if [ "$installed_version" != "$LINKERD_EDGE_VERSION" ]; then
  echo "Install Linkerd CLI version $LINKERD_EDGE_VERSION" >&2
  exit 1
else
  echo "Linkerd CLI version $LINKERD_EDGE_VERSION is already installed"
fi

# Install the Gateway API CRDs.

kubectl apply -f ${GATEWAY_API_BASE_URL}/${GATEWAY_API_VERSION}/${GATEWAY_API_CHANNEL}-install.yaml

# Install the Linkerd control plane.
linkerd install --crds | kubectl apply -f -
linkerd install | kubectl apply -f -
linkerd check

# Run the conformance tests.

echo "Starting conformance tests..."

REPORT_NAME=${GATEWAY_API_CHANNEL}-${CONFORMANCE_VERSION}-default-report.yaml
GATEWAY_API_VERSION_MAJOR_MINOR=$(echo ${GATEWAY_API_VERSION} | sed 's/v\([0-9]*\.[0-9]*\).*/\1/')
REPORT_PATH=reports/v${GATEWAY_API_VERSION_MAJOR_MINOR}/${CONFORMANCE_PRODUCT}/${REPORT_NAME}

skip_tests_arg=""

if [ -n "$SKIP_TESTS" ]; then
  skip_tests_arg="--skip-tests=${SKIP_TESTS}"
fi

go test \
    -p 1 \
    -v \
    ./conformance \
    -run TestConformance \
    -args \
        --organization Linkerd \
        --project Linkerd \
        --url https://github.com/linkerd/linkerd2 \
        --version ${LINKERD_VERSION} \
        --contact "gateway-api@buoyant.io" \
        --report-output ${REPORT_PATH} \
        --allow-crds-mismatch \
        --conformance-profiles=MESH-HTTP,MESH-GRPC \
        --all-features \
        --exempt-features=Gateway,ReferenceGrant,${UNSUPPORTED_FEATURES} \
        $skip_tests_arg --namespace-annotations=linkerd.io/inject=enabled
