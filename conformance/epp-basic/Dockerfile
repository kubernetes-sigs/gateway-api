# Dockerfile has specific requirement to put this ARG at the beginning:
# https://docs.docker.com/engine/reference/builder/#understand-how-arg-and-from-interact
ARG BUILDER_IMAGE=golang:1.25
ARG BASE_IMAGE=gcr.io/distroless/static:nonroot

## Multistage build
FROM ${BUILDER_IMAGE} AS builder
ENV CGO_ENABLED=0
ENV GOOS=linux
ENV GOARCH=amd64
ARG COMMIT_SHA=unknown
ARG BUILD_REF

# Dependencies
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download

# Sources
# For the lightweight EPP in Gateway API, all its files are located within conformance/epp-basic.
COPY conformance/epp-basic ./conformance/epp-basic

WORKDIR /src/conformance/epp-basic/cmd
RUN go build -ldflags="-X sigs.k8s.io/gateway-api/conformance/epp-basic/pkg/server.CommitSHA=${COMMIT_SHA} -X sigs.k8s.io/gateway-api/conformance/epp-basic/pkg/server.BuildRef=${BUILD_REF}" -o /lwepp

## Multistage deploy
FROM ${BASE_IMAGE}

WORKDIR /
COPY --from=builder /lwepp /lwepp

ENTRYPOINT ["/lwepp"]
