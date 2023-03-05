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

all: generate vet fmt verify test

# Run generators for protos, Deepcopy funcs, CRDs, and docs.
.PHONY: generate
generate: update-codegen update-webhook-yaml

.PHONY: update-codegen
update-codegen:
	hack/update-codegen.sh

.PHONY: update-webhook-yaml
update-webhook-yaml:
	hack/update-webhook-yaml.sh

.PHONY: build-install-yaml
build-install-yaml:
	hack/build-install-yaml.sh

# Run go fmt against code
fmt:
	go fmt ./...

# Run go vet against code
vet:
	go vet ./...

# Run go test against code
test:
	go test -race -cover ./pkg/... ./apis/... ./conformance/utils/...

# Run conformance tests against controller implementation
.PHONY: conformance
conformance:
	go test -v ./conformance/... -args ${CONFORMANCE_FLAGS}

# Install CRD's and example resources to a pre-existing cluster.
.PHONY: install
install: crd example

# Install the CRD's to a pre-existing cluster.
.PHONY: crd
crd:
	kubectl kustomize config/crd | kubectl apply -f -

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
	hack/verify-all.sh -v

# Build the documentation.
.PHONY: docs
docs:
	hack/make-docs.sh

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

BUILDX_CONTEXT = gateway-api-builder
BUILDX_PLATFORMS = linux/amd64,linux/arm64

# Setup multi-arch docker buildx enviroment.
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

# Generate a virtualenv install, which is useful for hacking on the
# docs since it installs mkdocs and all the right dependencies.
#
# On Ubuntu, this requires the python3-venv package.
virtualenv: .venv
.venv: requirements.txt
	@echo Creating a virtualenv in $@"... "
	@python3 -m venv $@ || (rm -rf $@ && exit 1)
	@echo Installing packages in $@"... "
	@$@/bin/python3 -m pip install -q -r requirements.txt || (rm -rf $@ && exit 1)
	@echo To enter the virtualenv type \"source $@/bin/activate\",  to exit type \"deactivate\"
