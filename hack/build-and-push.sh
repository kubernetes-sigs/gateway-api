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

if [[ -z "${VERIFY-}" ]];
then
  export DOCKER_PUSH_FLAG="--push"
else
  export DOCKER_PUSH_FLAG=""
fi

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

echo "Building and pushing echo-advanced image (from Istio) ...${BUILDX_PLATFORMS}"

docker buildx build \
    -t ${REGISTRY}/echo-advanced:${GIT_TAG} \
    -t ${REGISTRY}/echo-advanced:${VERSION_TAG} \
    --platform ${BUILDX_PLATFORMS} \
    ${DOCKER_PUSH_FLAG} \
    -f docker/Dockerfile.echo-advanced \
    .

echo "Building and pushing echo-basic image (previously in Ingress Controller Conformance Repo) ...${BUILDX_PLATFORMS}"

docker buildx build \
    -t ${REGISTRY}/echo-basic:${GIT_TAG} \
    -t ${REGISTRY}/echo-basic:${VERSION_TAG} \
    --platform ${BUILDX_PLATFORMS} \
    ${DOCKER_PUSH_FLAG} \
    -f docker/Dockerfile.echo-basic \
    .
