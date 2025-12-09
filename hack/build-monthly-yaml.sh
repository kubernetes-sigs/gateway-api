#!/usr/bin/env bash

# Copyright 2025 The Kubernetes Authors.
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

TAG=${TAG:-monthly-$(date +"%Y.%m")}

go run ./tools/generator --experimental-only --version=$TAG
bash hack/build-install-yaml.sh --experimental-only

DO_RESTORE=

set +o nounset

if [ -z "$MONTHLY_NONINTERACTIVE" ]; then
    read -p "Do you want to git restore config/crd/experimental? [Y/n]: " -r

    if [[ $REPLY =~ ^[Yy]$ ]]; then
        DO_RESTORE=yes
    fi
fi

set -o nounset

if [ -n "$DO_RESTORE" ]; then
    # DON'T commit the generated YAML here.
    git restore config/crd/experimental
fi

mv release/experimental-install.yaml release/${TAG}-install.yaml
echo "Generated release/${TAG}-install.yaml"

