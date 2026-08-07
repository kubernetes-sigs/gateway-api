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

if [ -z "${VERSION-}" ]; then
    VERSION="latest"
fi

mkdir -p release

for CHANNEL in "${CHANNELS[@]}"; do
    echo "$CHANNEL"
    go run ./tools/openapi-generator \
      --name "Gateway API ${CHANNEL} channel" \
      --version "$VERSION" \
      --output "api/openapi-spec/${CHANNEL}-swagger.json" \
      --add-gateway-api-object-defs \
      --pretty-print \
      "./config/crd/${CHANNEL}/gateway"*
done

echo "Generated:" "api/openapi-spec/"*-swagger.json
