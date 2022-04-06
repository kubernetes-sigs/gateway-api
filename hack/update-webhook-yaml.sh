#!/usr/bin/env bash

# Copyright 2020 The Kubernetes Authors.
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

if [[ -z "${BASE_REF-}" ]];
then
    echo "BASE_REF env var must be set and nonempty."
    exit 1
fi

semver='^v[0-9]+\.[0-9]+\.[0-9]+$'

if [[ "${BASE_REF}" =~ $semver ]]
then
    echo "Working on semver, need to replace."
    for yaml in `ls config/webhook/*.yaml`
    do
        sed -i "s/version: dev/version: ${BASE_REF}/g" $yaml
        sed -i "s/:latest/:${BASE_REF}/g" $yaml
    done
else
    echo "No version requested with BASE_REF, nothing to do."
fi

