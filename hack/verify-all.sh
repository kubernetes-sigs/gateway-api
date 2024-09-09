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
source "${SCRIPT_ROOT}/hack/kube-env.sh"

SILENT=true
FAILED_TEST=()

function is-excluded {
  for e in $EXCLUDE; do
    if [[ $1 -ef ${BASH_SOURCE} ]]; then
      return
    fi
    if [[ $1 -ef "$SCRIPT_ROOT/hack/$e" ]]; then
      return
    fi
  done
  return 1
}

while getopts ":v" opt; do
  case $opt in
    v)
      SILENT=false
      ;;
    \?)
      echo "Invalid flag: -$OPTARG" >&2
      exit 1
      ;;
  esac
done

if $SILENT ; then
  echo "Running in the silent mode, run with -v if you want to see script logs."
fi

EXCLUDE="verify-all.sh"

# TODO(mlavacca): once the prow configuration will be updated with a new target for 
# test-crds-validation.sh, we can remove it from the find command, otherwise it will be run twice.
SCRIPTS=$(find "${SCRIPT_ROOT}"/hack -name "verify-*.sh" -o -name "test-crds-validation.sh")

ret=0
for t in $SCRIPTS;
do
  if is-excluded "${t}" ; then
    echo "Skipping $t"
    continue
  fi
  if $SILENT ; then
    echo -e "Verifying $t"
    if bash "$t" &> /dev/null; then
      echo -e "${color_green}SUCCESS${color_norm}"
    else
      echo -e "${color_red}FAILED: $t ${color_norm}"
      FAILED_TEST+=("$t")
      ret=1
    fi
  else
    if bash "$t"; then
      echo -e "${color_green}SUCCESS: $t ${color_norm}"
    else
      echo -e "${color_red}Test FAILED: $t ${color_norm}"
       FAILED_TEST+=("$t")
      ret=1
    fi
  fi
done

if [ ${#FAILED_TEST[@]} -ne 0 ]; then
  echo -e "\n${color_red}Summary of failed tests:${color_norm}"
  for test in "${FAILED_TEST[@]}"; do
    echo -e "${color_red}- $test${color_norm}"
  done
else
  echo -e "\n${color_green}All tests passed successfully.${color_norm}"
fi

exit $ret

# ex: ts=2 sw=2 et filetype=sh
