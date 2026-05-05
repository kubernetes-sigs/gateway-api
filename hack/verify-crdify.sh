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

readonly CRDIFY_VERSION="v0.5.0"
SCRIPT_ROOT=$(dirname "${BASH_SOURCE}")/..
CRDIFY_BASE_REF=${CRDIFY_BASE_REF:-${PULL_BASE_SHA:-main}}

cd "${SCRIPT_ROOT}"
go install "sigs.k8s.io/crdify@${CRDIFY_VERSION}"

if ! git rev-parse --verify "${CRDIFY_BASE_REF}" >/dev/null 2>&1; then
  git fetch origin "${CRDIFY_BASE_REF}:${CRDIFY_BASE_REF}"
fi

failed=false
error_count=0

for crd_dir in config/crd/standard config/crd/experimental; do
  for file in "${crd_dir}"/*.yaml; do
    if [[ "$(basename "${file}")" == "kustomization.yaml" ]]; then
      continue
    fi

    if ! git cat-file -e "${CRDIFY_BASE_REF}:${file}" 2>/dev/null; then
      echo "Skipping crdify check for ${file}: new CRD file (not present in ${CRDIFY_BASE_REF})"
      continue
    fi

    if ! crdify \
      "git://${CRDIFY_BASE_REF}?path=${file}" \
      "file://${SCRIPT_ROOT}/${file}"; then
      failed=true
      error_count=$((error_count + 1))
    fi
  done
done

echo
echo "CRD compatibility check summary:"
if ${failed}; then
  echo "Breaking changes detected in ${error_count} CRD file(s)."
  exit 1
else
  echo "No breaking changes detected."
  exit 0
fi
