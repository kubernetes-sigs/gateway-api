#!/bin/bash

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

KUBE_ROOT=$(dirname "${BASH_SOURCE}")/..

cd "${KUBE_ROOT}"

GENERATED_FILE_NAME="zz_generated"

find_non_generated_files() {
    find . -not \( \
        \( \
        -wholename './.git' \
        -o -wholename '*/vendor/*' \
        -o -wholename "**/${GENERATED_FILE_NAME}.*.go" \
        \) -prune \
        \) -name '*.go'
}

# get a list of all stale generated files
stale_files=
for directory in $(find_non_generated_files | xargs -L 1 dirname | uniq); do
    # filter out go directorys without generated files
    if ! compgen -G "${directory}/${GENERATED_FILE_NAME}.*.go" >/dev/null; then
        continue
    fi

    # latest modified date of all input files
    latest_modified_date=$(find $directory -type f -not -name '*zz_generated*' | xargs -L 1 stat -c '%Y' | sort -r | head -1)

    for generated_file in $(find $directory -type f -name '*zz_generated*'); do
        # modified date of the generated file
        generated_file_modified_date=$(stat -c '%Y' $generated_file)

        if [ "$generated_file_modified_date" -le "$latest_modified_date" ]; then
            # generated file was made before the latest change
            stale_files+=($generated_file)
        fi
    done

done

if [[ -n "${stale_files}" ]]; then
    echo "!!! following generated files are stale (run 'make -f kubebuilder.mk generate'): "
    echo "${stale_files}"
    exit 1
fi
