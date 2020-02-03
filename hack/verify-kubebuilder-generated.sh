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

GENERATED_FILE_NAME="zz_generated.deepcopy.go"

find_non_generated_files() {
    # ignoring tools since it's an invalid go package
    find . -not \( \
        \( \
        -wholename './.git' \
        -o -wholename '*/vendor/*' \
        -o -wholename './tools' \
        -o -wholename "**/${GENERATED_FILE_NAME}" \
        \) -prune \
        \) -name '*.go'
}

generate() {
    # taken from the kubebuilder.mk generate rule
    GOFLAGS=-mod=vendor go run sigs.k8s.io/controller-tools/cmd/controller-gen object:headerFile=./hack/boilerplate.go.txt paths="$1" output:stdout
}

sha() {
    sha256sum "$1" | cut -d ' ' -f 1
}

# get a list of all stale generated files
stale_packages=()
for directory in $(find_non_generated_files | xargs -L 1 dirname | uniq); do
    # filter out go directories without generated files
    if (! compgen -G "${directory}/${GENERATED_FILE_NAME}" >/dev/null) && generate "$directory"; then
        continue
    fi

    generated_file_output=$(sha <(generate "$directory"))
    current_generated_file=$(sha "${directory}/${GENERATED_FILE_NAME}")

    if [ "$generated_file_output" != "$current_generated_file" ]; then
        stale_packages+=("$directory")
    fi

done

# check if there exists any stale packages
if [[ -n "${stale_packages:+1}" ]]; then
    echo "!!! following packages have stale generated files (run 'make -f kubebuilder.mk generate' to fix): "
    echo "${stale_packages}"
    exit 1
fi
