#!/bin/bash

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

SCRIPT_ROOT=$(dirname "${BASH_SOURCE}")/..

IMPLS_DIR=${IMPLS_DIR:-${SCRIPT_ROOT}/conformance/list/implementations}
INTEGS_DIR=${INTEGS_DIR:-${SCRIPT_ROOT}/conformance/list/integrations}
REPORTS_DIR=${REPORTS_DIR:-${SCRIPT_ROOT}/conformance/reports}
OUTPUT_FILE=${OUTPUT_FILE:-${SCRIPT_ROOT}/site/content/en/docs/implementations/list.md}
go run tools/implist/main.go -d "${IMPLS_DIR}/" -i "${INTEGS_DIR}" -r "${REPORTS_DIR}" -o "${OUTPUT_FILE}"

