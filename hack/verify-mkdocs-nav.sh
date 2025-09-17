#!/bin/bash

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

SCRIPT_ROOT=$(dirname "${BASH_SOURCE}")/..

TMP_DIFFROOT="${SCRIPT_ROOT}/_tmp"
TMP_FILE="${TMP_DIFFROOT}/nav.yml"

GEPS_TOC_SKIP="${GEPS_TOC_SKIP:-696}"
NAV_CONF="${NAV_CONF:-${SCRIPT_ROOT}/nav.yml}"
NAV_TEMPLATE="${NAV_TEMPLATE:-${NAV_CONF}.tmpl}"
GEPS_TOC_DIR=${GEPS_TOC_DIR:-${SCRIPT_ROOT}/geps}

cleanup() {
  rm -rf "${TMP_DIFFROOT}"
}
trap "cleanup" EXIT SIGINT

cleanup

mkdir -p "${TMP_DIFFROOT}"

go run cmd/gepstoc/main.go -g "${GEPS_TOC_DIR}/" -t "${NAV_TEMPLATE}" -s "${GEPS_TOC_SKIP}" > "${TMP_FILE}"

echo "diffing ${NAV_CONF} against freshly generated configuration"
ret=0
diff -Naupr --no-dereference "${NAV_CONF}" "${TMP_FILE}" || ret=1

if [[ $ret -eq 0 ]]; then
  echo "${NAV_CONF} up to date."
else
  echo "${NAV_CONF} is out of date. Please run hack/update-mkdocs-nav.sh"
  exit 1
fi
