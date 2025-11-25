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

# We need all the Make variables exported as env vars.
# Note that the ?= operator works regardless.

# Enable Go modules.
export GO111MODULE=on

# The registry to push container images to.
export REGISTRY ?= gcr.io/k8s-staging-gateway-api

# These are overridden by cloudbuild.yaml when run by Prow.

# Prow gives this a value of the form vYYYYMMDD-hash.
# (It's similar to `git describe` output, and for non-tag
# builds will give vYYYYMMDD-COMMITS-HASH where COMMITS is the
# number of commits since the last tag.)
export GIT_TAG ?= dev

# Prow gives this the reference it's called on.
# The test-infra config job only allows our cloudbuild to
# be called on `main` and semver tags, so this will be
# set to one of those things.
export BASE_REF ?= main

# The commit hash of the current checkout
# Used to pass a binary version for main,
# overridden to semver for tagged versions.
# Cloudbuild will set this in the environment to the
# commit SHA, since the Prow does not seem to check out
# a git repo.
export COMMIT ?= $(shell git rev-parse --short HEAD)

DOCKER ?= docker
# TOP is the current directory where this Makefile lives.
TOP := $(dir $(firstword $(MAKEFILE_LIST)))
# ROOT is the root of the mkdocs tree.
ROOT := $(abspath $(TOP))

# Command-line flags passed to "go test" for the conformance
# test. These are passed after the "-args" flag.
CONFORMANCE_FLAGS ?=
GO_TEST_FLAGS ?=

# Flags for CRD validation tests
CEL_TEST_K8S_VERSION ?= 
CEL_TEST_CRD_CHANNEL ?= standard

# Compilation flags for binaries
GOARCH ?= $(shell go env GOARCH)
GOOS ?= $(shell go env GOOS)

all: generate vet fmt verify test conformance-bin

.PHONY: clean-generated
clean-generated:
	rm -rf pkg/client/clientset
	rm -rf pkg/client/listers
	rm -rf pkg/client/informers

# Run generators for protos, Deepcopy funcs, CRDs, and docs.
.PHONY: generate
generate: clean-generated update-codegen tidy

.PHONY: update-codegen
update-codegen:
	hack/update-codegen.sh

.PHONY: build-install-yaml
build-install-yaml:
	hack/build-install-yaml.sh

.PHONY: build-monthly-yaml
build-monthly-yaml:
	hack/build-monthly-yaml.sh

# Run go fmt against code
fmt:
	go fmt ./...

# Run go vet against code
vet:
	go vet ./...

# Run go test against code
test:
	go test -race -cover ./apis/... ./conformance/utils/...
# Run tests for each submodule.
	cd "conformance/echo-basic" && go test -race -cover ./...

.PHONY: tidy
tidy:
	go work sync
	find . -name go.mod -execdir sh -c 'go mod tidy' \;

# Run tests for CRDs validation
.PHONY: test.crds-validation
test.crds-validation:
	K8S_VERSION=$(CEL_TEST_K8S_VERSION) CRD_CHANNEL=$(CEL_TEST_CRD_CHANNEL) go test ${GO_TEST_FLAGS} -count=1 -timeout=120s --tags=$(CEL_TEST_CRD_CHANNEL) -v ./tests/cel
	K8S_VERSION=$(CEL_TEST_K8S_VERSION) CRD_CHANNEL=$(CEL_TEST_CRD_CHANNEL) go test ${GO_TEST_FLAGS} -count=1 -timeout=120s -v ./tests/crd

# Run conformance tests against controller implementation
.PHONY: conformance
conformance:
	go test ${GO_TEST_FLAGS} -v ./conformance -run TestConformance -args ${CONFORMANCE_FLAGS}

# Build a conformance.test binary that can be used as a standalone binary to run conformance test
.PHONY: conformance-bin
conformance-bin:
	GOOS=$(GOOS) GOARCH=$(GOARCH) go test -c -v ./conformance 

# Install CRD's and example resources to a preexisting cluster.
.PHONY: install
install: crd example

# Install the CRD's to a preexisting cluster.
.PHONY: crd
crd:
	kubectl kustomize config/crd | kubectl apply -f -

# Install the example resources to a preexisting cluster.
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
	hack/verify-all.sh -v

# Build the documentation.
.PHONY: docs
docs:
	hack/make-docs.sh

.PHONY: update-conformance-image-refs
update-conformance-image-refs:
	hack/update-conformance-image-refs.sh

# Verify if support Docker Buildx.
.PHONY: image.buildx.verify
image.buildx.verify:
	docker version
	$(eval PASS := $(shell docker buildx --help | grep "docker buildx" ))
	@if [ -z "$(PASS)" ]; then \
		echo "Cannot find docker buildx, please install first."; \
		exit 1;\
	else \
		echo "===========> Support docker buildx"; \
		docker buildx version; \
	fi

export BUILDX_CONTEXT = gateway-api-builder
export BUILDX_PLATFORMS = linux/amd64,linux/arm64

# Setup multi-arch docker buildx environment.
.PHONY: image.multiarch.setup
image.multiarch.setup: image.buildx.verify
# Ensure qemu is in binfmt_misc.
# Docker desktop already has these in versions recent enough to have buildx,
# We only need to do this setup on linux hosts.
	@if [ "$(shell uname)" == "Linux" ]; then \
		docker run --rm --privileged multiarch/qemu-user-static --reset -p yes; \
	fi
# Ensure we use a builder that can leverage it, we need to recreate one.
	docker buildx rm $(BUILDX_CONTEXT) || :
	docker buildx create --use --name $(BUILDX_CONTEXT) --platform "${BUILDX_PLATFORMS}"

# Build and Push Multi Arch Images.
.PHONY: release-staging
release-staging: image.multiarch.setup
	hack/build-and-push.sh

# Docs

DOCS_BUILD_CONTAINER_NAME ?= gateway-api-mkdocs

.PHONY: build-docs
build-docs: update-geps
	docker build --pull -t gaie/mkdocs hack/mkdocs/image
	docker rm -f $(DOCS_BUILD_CONTAINER_NAME) || true
	docker run --name $(DOCS_BUILD_CONTAINER_NAME) --rm -v ${PWD}:/docs gaie/mkdocs build

.PHONY: build-docs-netlify
build-docs-netlify: update-geps api-ref-docs
	pip install -r hack/mkdocs/image/requirements.txt
	python -m mkdocs build

.PHONY: live-docs
live-docs: update-geps
	docker build -t gw/mkdocs hack/mkdocs/image
	docker run --rm -it -p 3000:3000 -v ${PWD}:/docs gw/mkdocs

.PHONY: update-geps
update-geps:
	hack/update-geps.sh

.PHONY: api-ref-docs
api-ref-docs:
	hack/mkdocs/generate.sh
