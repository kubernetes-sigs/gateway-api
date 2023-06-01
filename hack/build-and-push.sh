#!/bin/bash

# Copyright 2022 The Kubernetes Authors.
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

# This file is run by cloudbuild, from cloudbuild.yaml, using `make release-staging`.

set -o errexit
set -o nounset
set -o pipefail

if [[ -z "${GIT_TAG-}" ]];
then
    echo "GIT_TAG env var must be set and nonempty."
    exit 1
fi

if [[ -z "${BASE_REF-}" ]];
then
    echo "BASE_REF env var must be set and nonempty."
    exit 1
fi

if [[ -z "${COMMIT-}" ]];
then
    echo "COMMIT env var must be set and nonempty."
    exit 1
fi

if [[ -z "${REGISTRY-}" ]];
then
    echo "REGISTRY env var must be set and nonempty."
    exit 1
fi

# If our base ref == "main" then we will tag :latest.
VERSION_TAG=latest

# We tag the go binary with the git-based tag by default
BINARY_TAG=$GIT_TAG

# $BASE_REF has only two things that it can be set to by cloudbuild and Prow,
# `main`, or a semver tag.
# This is controlled by k8s.io/test-infra/config/jobs/image-pushing/k8s-staging-gateway-api.yaml.
if [[ "${BASE_REF}" != "main" ]]
then
    # Since we know this is built from a tag or release branch, we can set the VERSION_TAG
    VERSION_TAG="${BASE_REF}"

    # Include the semver tag in the binary instead of the git-based tag
    BINARY_TAG="${BASE_REF}"
fi

# Support multi-arch image build and push.
BUILDX_PLATFORMS="linux/amd64,linux/arm64"

echo "Building and pushing admission-server image...${BUILDX_PLATFORMS}"

# First, build the image, with the version info passed in.
# Note that an image will *always* be built tagged with the GIT_TAG, so we know when it was built.
# And, we add an extra version tag - either :latest or semver.
# The buildx integrate build and push in one line.
docker buildx build \
    -t ${REGISTRY}/admission-server:${GIT_TAG} \
    -t ${REGISTRY}/admission-server:${VERSION_TAG} \
    --build-arg "COMMIT=${COMMIT}" \
    --build-arg "TAG=${BINARY_TAG}" \
    --platform ${BUILDX_PLATFORMS} \
    --push \
    -f docker/Dockerfile.webhook \
    .

echo "Building and pushing echo-server image...${BUILDX_PLATFORMS}"

docker buildx build \
    -t ${REGISTRY}/echo-server:${GIT_TAG} \
    -t ${REGISTRY}/echo-server:${VERSION_TAG} \
    --platform ${BUILDX_PLATFORMS} \
    --push \
    -f docker/Dockerfile.echo \
    .
