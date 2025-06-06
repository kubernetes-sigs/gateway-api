# Set these as needed.
LINKERD_VERSION=${LINKERD_VERSION:-enterprise-2.18}
GATEWAY_API_CHANNEL=${GATEWAY_API_CHANNEL:-standard}
GATEWAY_API_VERSION=${GATEWAY_API_VERSION:-v1.3.0}

CONFORMANCE_PRODUCT=buoyant-enterprise-for-linkerd
CONFORMANCE_VERSION=$(echo $LINKERD_VERSION | cut -d- -f2-)
GATEWAY_API_BASE_URL=https://github.com/kubernetes-sigs/gateway-api/releases/download

echo "Using Buoyant Enterprise for Linkerd version $LINKERD_VERSION"
echo "Using Gateway API $GATEWAY_API_VERSION $GATEWAY_API_CHANNEL"

# Install the Linkerd CLI.
curl --proto '=https' --tlsv1.2 -sSfL https://enterprise.buoyant.io/install \
  | env LINKERD2_VERSION=${LINKERD_EDGE_VERSION} sh

# Install the Gateway API CRDs.

kubectl apply -f ${GATEWAY_API_BASE_URL}/${GATEWAY_API_VERSION}/${GATEWAY_API_CHANNEL}-install.yaml

# Install the Linkerd control plane.
linkerd install --crds | kubectl apply -f -
linkerd install | kubectl apply -f -
linkerd check

# Run the conformance tests.

REPORT_NAME=${GATEWAY_API_CHANNEL}-${CONFORMANCE_VERSION}-default-report.yaml
REPORT_PATH=reports/${GATEWAY_API_VERSION}/${CONFORMANCE_PRODUCT}/${REPORT_NAME}

        # --supported-features=Mesh,HTTPRoute,GRPCRoute \

go test \
    -p 4 \
    ./conformance \
    -run TestConformance \
    -args \
        --organization Buoyant \
        --project "Buoyant Enterprise for Linkerd" \
        --url https://buoyant.io/ \
        --version enterprise-2.18 \
        --contact "gateway-api@buoyant.io" \
        --report-output ${REPORT_PATH} \
        --conformance-profiles=MESH-HTTP,MESH-GRPC \
        --all-features \
        --exempt-features=Gateway,ReferenceGrant \
        --namespace-annotations=linkerd.io/inject=enabled

