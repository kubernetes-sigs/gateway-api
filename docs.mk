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

# Build rules for the documentation
#
# `make help` for a summary of top-level targets.

DOCKER ?= docker
MKDOCS_IMAGE ?= k8s.gcr.io/service-apis-mkdocs:latest
MKDOCS ?= mkdocs
SERVE_BIND_ADDRESS ?= 127.0.0.1

# TOP is the current directory where this Makefile lives.
TOP := $(dir $(firstword $(MAKEFILE_LIST)))
# DOCROOT is the root of the mkdocs tree.
DOCROOT := $(abspath $(TOP))

# Grab the uid/gid to fix permissions due to running in a docker container.
GID := $(shell id -g)
UID := $(shell id -u)

# Support docker images.
.PHONY: images
images: .mkdocs.dockerfile.timestamp

# build the image for mkdocs
.mkdocs.dockerfile.timestamp: mkdocs.dockerfile mkdocs.requirements.txt
	docker build -t $(MKDOCS_IMAGE) -f mkdocs.dockerfile .
	date > $@

# verify that the docs can be successfully built
.PHONY: verify
verify: images
	$(DOCKER) run \
		--mount type=bind,source=$(DOCROOT),target=/d \
		--sig-proxy=true \
		--rm \
		$(MKDOCS_IMAGE) \
		/bin/bash -c "cd /d && $(MKDOCS) build --site-dir=/tmp/d"

# serve runs mkdocs as a local webserver for interactive development.
# This will serve the live copy of the docs on 127.0.0.1:8000.
.PHONY: serve
serve: images
	$(DOCKER) run \
		-it \
		--sig-proxy=true \
		--mount type=bind,source=$(DOCROOT),target=/d \
		-p $(SERVE_BIND_ADDRESS):8000:8000 \
		--rm \
		$(MKDOCS_IMAGE) \
		/bin/bash -c "cd /d && $(MKDOCS) serve -a 0.0.0.0:8000"

# deploy will publish generated docs into gh-pages branch.
.PHONY: deploy
deploy: images
	$(DOCKER) run \
		-it \
		--sig-proxy=true \
		--mount type=bind,source=$(DOCROOT),target=/d \
		--rm \
		$(MKDOCS_IMAGE) \
		/bin/bash -c "cd /d && git config --global url."https://$$GITHUB_USER:$$GITHUB_TOKEN@github.com/".pushInsteadOf "https://github.com/" && $(MKDOCS) gh-deploy --site-dir=/tmp/d"

# help prints usage for this Makefile.
.PHONY: help
help:
	@echo "Usage:"
	@echo ""
	@echo "make        Build the documentation"
	@echo "make help   Print this help message"
	@echo "make verify Verify docs can be successfully generated"
	@echo "make serve  Run the webserver for live editing (ctrl-C to quit)"
	@echo "make deploy Deploy generated docs into gh-pages branch"

# init creates a new mkdocs template. This is included for completeness.
.PHONY: images init
init:
	$(DOCKER) run \
		--mount type=bind,source=$(DOCROOT),target=/d \
		--sig-proxy=true \
		--rm \
		$(MKDOCS_IMAGE) \
		/bin/bash -c "$(MKDOCS) new d; find /d -exec chown $(UID):$(GID) {} \;"
