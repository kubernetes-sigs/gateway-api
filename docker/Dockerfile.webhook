# Copyright 2021 The Kubernetes Authors.
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

ARG BUILDPLATFORM=linux/amd64
FROM --platform=$BUILDPLATFORM golang:1.19 AS build-env
RUN mkdir -p /go/src/sig.k8s.io/gateway-api
WORKDIR /go/src/sig.k8s.io/gateway-api
COPY  . .
ARG TARGETARCH
ARG TAG
ARG COMMIT
RUN CGO_ENABLED=0 GOARCH=$TARGETARCH GOOS=linux go build -a -o gateway-api-webhook \
      -ldflags "-s -w -X main.VERSION=$TAG -X main.COMMIT=$COMMIT" ./cmd/admission

FROM gcr.io/distroless/static:nonroot
WORKDIR /
COPY --from=build-env /go/src/sig.k8s.io/gateway-api/gateway-api-webhook .
# Use uid of nonroot user (65532) because kubernetes expects numeric user when applying pod security policies
USER 65532
ENTRYPOINT ["/gateway-api-webhook"]
