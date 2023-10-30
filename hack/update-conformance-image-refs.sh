#!/bin/bash

# Copyright 2023 The Kubernetes Authors.
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


set -o errexit
set -o nounset
set -o pipefail

# Wrap sed to deal with GNU and BSD sed flags.
run::sed() {
    local -r vers="$(sed --version < /dev/null 2>&1 | grep -q GNU && echo gnu || echo bsd)"
    case "$vers" in
        gnu) sed -i "$@" ;;
        *) sed -i '' "$@" ;;
esac
}

IMAGES=(
  "${REGISTRY}/echo-advanced"
  "${REGISTRY}/echo-basic"
)

# Export so we can use run::sed in subshells
# https://stackoverflow.com/questions/4321456/find-exec-a-shell-function-in-linux
export -f run::sed

for IMAGE in ${IMAGES[@]}; do
  echo "Fetching latest tags for $IMAGE"
  TAG_FILE=$(mktemp)
  go run github.com/google/go-containerregistry/cmd/gcrane@latest ls "$IMAGE" --json \
    | jq -r '.tags[]' \
    | grep -v latest \
    | sort -rV \
    | head -n1 > "$TAG_FILE"

  export REPO=${IMAGE#"${REGISTRY}"}
  export IMAGE_TAG=$(cat "$TAG_FILE")
  export IMAGE
  echo "Found tag $IMAGE_TAG - updating manifests..."
  find . -type f -name "*.yaml" -exec bash -c 'run::sed -e "s,image:.*${REPO}.*$,image: ${IMAGE}:${IMAGE_TAG},g" "$0"' {} \;
done


