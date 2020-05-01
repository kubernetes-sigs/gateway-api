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

readonly SCRIPT_ROOT=$(dirname "${BASH_SOURCE}")/..
readonly APIDOC="${SCRIPT_ROOT}/docs-src/spec.md"

ret=0
diff -Naq "$APIDOC" <(./hack/api-docs/generate.sh /dev/fd/1 2>/dev/null) || ret=$?
if [[ $ret -eq 0 ]]
then
  echo "${APIDOC} up to date."
else
  echo "${APIDOC} is out of date. Please run 'make generate'"
  exit 1
fi
