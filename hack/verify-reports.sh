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

error() {
  echo "ERROR: $*" 1>&2
}

info() {
  echo "INFO: $*" 1>&2
}

# Check if the provided Gateway API version is greater than or equal to v1.1.0
check_ge_v1.1.0() {
    local version=$1
    local minimum_version="v1.1.0"

    # Normalize versions to remove leading 'v' and compare using sort -V
    if [[ $(echo -e "${version#v}\n${minimum_version#v}" | sort -V | head -n1) == "${minimum_version#v}" ]]; then
        return 0
    else
        return 1
    fi
}

# Check if the report fields are valid according to the rules defined in https://github.com/kubernetes-sigs/gateway-api/blob/release-1.1/conformance/reports/README.md
check_report_fields() {
    local report=$1
    local expected_gateway_api_version=$2

    # Check if the implementation version is a valid semver
    local gateway_api_version=$(yq eval '.gatewayAPIVersion' "$report")
    local gateway_api_channel=$(yq eval '.gatewayAPIChannel' "$report")
    local mode=$(yq eval '.mode' "$report")

    if [[ ${gateway_api_version} != ${expected_gateway_api_version} ]]; then
        error "$report gatewayAPIVersion does not match Gateway API version folder"
        EXIT_VALUE=1
    fi
    if [[ $gateway_api_channel != "standard" && $gateway_api_channel != "experimental" ]]; then
        error "$report gatewayAPIChannel is neither standard nor experimental"
        EXIT_VALUE=1
    fi
    if [[ $mode == "" ]]; then
        error "$report mode must be set"
        EXIT_VALUE=1
    fi
}

REPORTS_DIR=$(dirname "${BASH_SOURCE}")/../conformance/reports
# Regex to match the report file name pattern defined in https://github.com/kubernetes-sigs/gateway-api/blob/release-1.1/conformance/reports/README.md#how-this-folder-is-structured
FILE_NAMING_PATTERN="^(standard|experimental)-v?[0-9]+\.[0-9]+(\.[0-9]+)?-[^-]+-report\.yaml$"
EXIT_VALUE=0

for dir in ${REPORTS_DIR}/*
do
    element="${dir##*/}"
    if check_ge_v1.1.0 "${element}"; then
        if [[ -d "${dir}" ]]; then
            gateway_api_version="${element}"
            for implementation_dir in ${dir}/*
            do
                implementation=$(basename "${implementation_dir}")
                info "Checking ${implementation} project directory for Gateway API version ${gateway_api_version}"
            
                if [[ -f "${implementation_dir}/README.md" ]]; then
                # Check if the README.md has broken links
                    docker run -v $(readlink -f "$implementation_dir"):/${implementation}:ro --rm -i ghcr.io/tcort/markdown-link-check:stable /${implementation}/README.md
                else
                    error "missing README.md in ${implementation_dir}"
                    EXIT_VALUE=1
                fi
                for report in ${implementation_dir}/*.yaml
                do
                    report_name="${report##*/}"
                    if [[ ! $report_name =~ $FILE_NAMING_PATTERN ]]; then
                        error "$report_name does not match the naming pattern defined in the README.md"
                        EXIT_VALUE=1
                    fi
                    check_report_fields "${report}" "${gateway_api_version}"
                done
            done
        fi
    fi
done

exit ${EXIT_VALUE}
