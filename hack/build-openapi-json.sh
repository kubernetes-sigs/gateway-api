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
set -o pipefail

CHANNELS=(standard experimental)

set -o nounset

while [ -n "${1-}" ]; do
  case "$1" in
      "--experimental-only")
          CHANNELS=(experimental)
          ;;
      "--version"|"-v")
          version="$2"
          shift
          ;;
      "--version="*)
          version="${1#"--version="}"
          ;;
      *)
          printf 'Error: unknown argument "%s"\n' "$1" >> /dev/stderr
          exit 1
          ;;
  esac
  shift
done

if [ -z "${version-}" ]; then
    # If a tag matches this commit, return that tag. Otherwise, generate a
    # valid Git reference string unique for this commit that contains the most
    # recent previous version string.
    version="$(git describe --tags --match 'v*' --match 'monthly-*')"
fi

mkdir -p release/

for CHANNEL in "${CHANNELS[@]}"; do
    echo "$CHANNEL"
    go run ./tools/openapi-generator \
      --name "Gateway API ${CHANNEL} channel" \
      --version "$version" \
      --output "release/${CHANNEL}-swagger.json" \
      "./config/crd/${CHANNEL}/gateway"*
done

echo "Generated:" release/*-swagger.json
