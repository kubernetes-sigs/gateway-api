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

SCRIPT_ROOT=$(dirname "${BASH_SOURCE}")/..
DIFFROOT="${SCRIPT_ROOT}/config/crd/bases"
TMP_DIFFROOT="${SCRIPT_ROOT}/_tmp/config/crd/bases"
_tmp="${SCRIPT_ROOT}/_tmp"
# The controller-gen command for generating CRDs from API definitions.
CONTROLLER_GEN="go run sigs.k8s.io/controller-tools/cmd/controller-gen"
# Need v1 to support defaults in CRDs, unfortunately limiting us to k8s 1.16+
CRD_OPTIONS="crd:crdVersions=v1"

cd "${SCRIPT_ROOT}"

cleanup() {
  rm -rf "${_tmp}"
}
trap "cleanup" EXIT SIGINT

cleanup

mkdir -p "${TMP_DIFFROOT}"
cp -a "${DIFFROOT}"/* "${TMP_DIFFROOT}"

${CONTROLLER_GEN} ${CRD_OPTIONS} rbac:roleName=manager-role webhook \
paths="./..." output:crd:artifacts:config=${TMP_DIFFROOT}

echo "diffing ${DIFFROOT} against freshly generated codegen in ${TMP_DIFFROOT}"
ret=0
diff -Naupr "${DIFFROOT}" "${TMP_DIFFROOT}" || ret=$?
if [[ $ret -eq 0 ]]
then
  echo "${DIFFROOT} up to date."
else
  echo "${DIFFROOT} is out of date. Please run 'make manifests'"
  exit 1
fi
