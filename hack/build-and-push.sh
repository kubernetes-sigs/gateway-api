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


LATEST=false

VERSION_TAG=$GIT_TAG

BINARY_VERSION=$COMMIT

# $BASE_REF has only two things that it can be set to by cloudbuild and Prow,
# `master`, or a semver tag.
# This is controlled by k8s.io/test-infra/config/jobs/image-pushing/k8s-staging-gateway-api.yaml.
if [[ "${BASE_REF}" != "master" ]]
then
    # Since we know this is built from a tag or release branch, we can set the VERSION_TAG
    VERSION_TAG="${BASE_REF}"
    # We want the binary version to show up correctly too.
    BINARY_VERSION="${BASE_REF}"
    # Use some bash magic to check if the semver does not end with -sometext, that
    # would indicate a prerelease version. If this is not a prerelease, then we want to set
    # the `latest` tag too.
    if [[ ! "${BASE_REF}" =~ -(.+)$ ]];
    then
    LATEST=true
    fi
fi

# First, build the image, with the version info passed in.
# Note that an image will *always* be built tagged with the GIT_TAG, so we know when it was built.
echo docker build --build-arg COMMIT=${BINARY_VERSION} --build-arg TAG=${VERSION_TAG} \
  			-t ${REGISTRY}/admission-server:${GIT_TAG} .

echo docker push ${REGISTRY}/admission-server:${GIT_TAG}

# Then, we add extra tags if required.
if [[ $VERSION_TAG != $GIT_TAG ]]
then
	echo docker tag ${REGISTRY}/admission-server:${GIT_TAG} ${REGISTRY}/admission-server:${VERSION_TAG}
	echo docker push ${REGISTRY}/admission-server:${VERSION_TAG}
fi

if [[ $LATEST == true ]]
then
	echo docker tag ${REGISTRY}/admission-server:${GIT_TAG} ${REGISTRY}/admission-server:latest
	echo docker push ${REGISTRY}/admission-server:latest
fi
