#!/usr/bin/env bash

# Copyright The Kubernetes Authors.
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

ROOT=$(dirname "${BASH_SOURCE[0]}")/..
cd "$ROOT" || exit 1

CHANNELS=(standard experimental)

OUTPUT_DIR="${OUTPUT_DIR:-"release"}"

while [ -n "${1-}" ]; do
  case "$1" in
      "--experimental-only")
          CHANNELS=(experimental)
          ;;
      "--monthly")
          VERSION=${TAG:-monthly-$(date +"%Y.%m")}
          ;;
      "--pretty-print")
          PRETTY=true
          ;;
      "--version-as-filename")
          VERSION_AS_NAME=true
          ;;
      "--version"|"-v")
          VERSION="$2"
          shift
          ;;
      "--version="*)
          VERSION="${1#"--version="}"
          ;;
      *)
          printf 'Error: unknown argument "%s"\n' "$1" >> /dev/stderr
          exit 1
          ;;
  esac
  shift
done

if [ -z "${VERSION-}" ]; then
    # If a tag matches this commit, return that tag. Otherwise, generate a
    # valid Git reference string unique for this commit that contains the most
    # recent previous version string.
    VERSION="$(git describe --tags --match 'v*' --match 'monthly-*')"
fi

mkdir -p "$OUTPUT_DIR"

for CHANNEL in "${CHANNELS[@]}"; do
    echo "$CHANNEL"

    if [[ "${VERSION_AS_NAME-}" == "true" ]]; then
      name="$VERSION"
    else
      name="$CHANNEL"
    fi

    go run ./tools/openapi-generator \
      --name "Gateway API ${CHANNEL} channel" \
      --version "$VERSION" \
      --output "${OUTPUT_DIR}/${name}-swagger.json" \
      --add-gateway-api-object-defs \
      --pretty-print="${PRETTY:-false}" \
      "./config/crd/${CHANNEL}/gateway"*
done

echo "Generated:" "$OUTPUT_DIR"/*-swagger.json
