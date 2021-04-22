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

FROM golang:1.16 AS build-env
RUN mkdir -p /go/src/sig.k8s.io/gateway-api
WORKDIR /go/src/sig.k8s.io/gateway-api
COPY  . .
RUN useradd -u 10001 webhook
RUN cd cmd/admission/ && CGO_ENABLED=0 GOOS=linux go build -a -o gateway-api-webhook && chmod +x gateway-api-webhook

FROM scratch
COPY --from=build-env /go/src/sig.k8s.io/gateway-api/cmd/admission/gateway-api-webhook .
COPY --from=build-env /etc/passwd /etc/passwd
USER webhook
ENTRYPOINT ["/gateway-api-webhook"]
