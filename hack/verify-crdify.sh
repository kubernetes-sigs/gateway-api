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

readonly GOTOOL="go tool"
readonly SCRIPT_ROOT=$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)
CRDIFY_ENFORCE=${CRDIFY_ENFORCE:-true}
CRDIFY_BASE_REF=${CRDIFY_BASE_REF:-${PULL_BASE_SHA:-main}}
REMOTE=${REMOTE:-origin}

cd "${SCRIPT_ROOT}"

if ! git rev-parse --verify "${CRDIFY_BASE_REF}" >/dev/null 2>&1; then
  git fetch "${REMOTE}" "${CRDIFY_BASE_REF}"
fi

error_count=0
error_files=""

for file in config/crd/standard/*.yaml; do
  filename="$(basename "${file}")"

  if [[ "${filename}" == "kustomization.yaml" || "${filename}" == *_vap_*.yaml ]]; then
    continue
  fi

  if ! git cat-file -e "${CRDIFY_BASE_REF}:${file}" 2>/dev/null; then
    echo "Skipping crdify check for ${file}: new CRD file (not present in ${CRDIFY_BASE_REF})"
    continue
  fi

  echo -e "\n${filename}:"
  if ! ${GOTOOL} sigs.k8s.io/crdify \
    "git://${CRDIFY_BASE_REF}?path=${file}" \
    "file://${SCRIPT_ROOT}/${file}"; then
    error_count=$((error_count + 1))
    error_files="${error_files}\n- ${filename}"
  fi
done

echo -e "\nCRD compatibility check summary:"
if [[ ${error_count} -gt 0 ]]; then
  echo -e "Breaking changes detected in ${error_count} CRD file(s):${error_files}"
  if [[ "${CRDIFY_ENFORCE}" == "true" ]]; then
    exit 1
  fi
else
  echo "No breaking changes detected."
fi
exit 0
