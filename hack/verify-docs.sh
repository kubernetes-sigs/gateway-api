#!/bin/bash

# Copyright 2014 The Kubernetes Authors.
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

cd $SCRIPT_ROOT
# Run the docs make
make -f docs.mk

# If there's any uncommitted changes, fail.
if git status -s docs/ 2>&1 | grep -E -q '^\s+[MADRCU]'; then
		echo "Uncommitted changes in docs:" ;
		git status -s docs;
		exit 1;
fi