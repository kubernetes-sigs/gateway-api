# Copyright 2019 The Kubernetes Authors.
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

DOCKER ?= docker
# Image to build protobufs
PROTO_IMG ?= k8s.gcr.io/kube-cross:v1.13.6-1
# TOP is the current directory where this Makefile lives.
TOP := $(dir $(firstword $(MAKEFILE_LIST)))
# ROOT is the root of the mkdocs tree.
ROOT := $(abspath $(TOP))

all: controller generate verify

# Build manager binary and run static analysis.
.PHONY: controller
controller:
	$(MAKE) -f kubebuilder.mk manager

# Run code generators for protos, Deepcopy funcs, CRDs, etc..
.PHONY: generate
generate:
	$(MAKE) proto
	$(MAKE) -f kubebuilder.mk generate
	$(MAKE) manifests

# Generate manifests e.g. CRD, RBAC etc.
.PHONY: manifests
manifests:
	$(MAKE) -f kubebuilder.mk manifests

# Generate protobufs
.PHONY: proto
proto:
	$(DOCKER) run -it \
		--mount type=bind,source=$(ROOT),target=/go/src/sigs.k8s.io/service-apis  \
		--mount type=bind,source=$(GOPATH)/pkg/mod,target=/go/pkg/mod  \
		--env GOPATH=/go \
		--env GOCACHE=/go/.cache \
		--rm \
		--user "$(shell id -u):$(shell id -g)" \
		-w /go/src/sigs.k8s.io/service-apis \
		$(PROTO_IMG) \
		hack/update-proto.sh

# Verify protobuf generation
.PHONY: verify-proto
verify-proto:
	$(DOCKER) run \
		--mount type=bind,source=$(ROOT),target=/realgo/src/sigs.k8s.io/service-apis \
		--env GOPATH=/go \
		--env GOCACHE=/go/.cache \
		--rm \
		--user "$(shell id -u):$(shell id -g)" \
		-w /go \
		$(PROTO_IMG) \
		/bin/bash -c "mkdir -p src/sigs.k8s.io/service-apis && \
			cp -r /realgo/src/sigs.k8s.io/service-apis/ src/sigs.k8s.io && \
			cd src/sigs.k8s.io/service-apis && \
			hack/update-proto.sh && \
			diff -r api /realgo/src/sigs.k8s.io/service-apis/api"

# Install CRD's and example resources to a pre-existing cluster.
.PHONY: install
install: manifests crd example

# Install the CRD's to a pre-existing cluster.
.PHONY: crd
crd:
	$(MAKE) -f kubebuilder.mk install

# Install the example resources to a pre-existing cluster.
.PHONY: example
example:
	hack/install-examples.sh

# Remove installed CRD's and CR's.
.PHONY: uninstall
uninstall:
	hack/delete-crds.sh

# Run static analysis.
.PHONY: verify
verify:
	hack/verify-all.sh

# Build the documentation.
.PHONY: docs
docs:
	# The docs image must be built locally until issue #141 is fixed.
	docker build --tag k8s.gcr.io/service-apis-mkdocs:latest -f mkdocs.dockerfile .
	$(MAKE) -f docs.mk

# Serve the docs site locally at http://localhost:8000.
.PHONY: serve
serve:
	$(MAKE) -f docs.mk serve

# Clean deletes generated documentation files.
.PHONY: clean
clean:
	$(MAKE) -f docs.mk clean
