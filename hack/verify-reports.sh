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

# Normalize a version string: strip leading 'v' and append '.0' if no patch.
normalize_version() {
    local v="${1#v}"
    if [[ "$v" =~ ^[0-9]+\.[0-9]+$ ]]; then
        v="${v}.0"
    fi
    echo "$v"
}

# Check if the provided Gateway API version is greater than or equal to v1.1.0
check_ge_v1.1.0() {
    local version=$(normalize_version "$1")
    local minimum_version="1.1.0"

    # Compare using sort -V
    if [[ $(echo -e "${version}\n${minimum_version}" | sort -V | head -n1) == "${minimum_version}" ]]; then
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

    # Compare only major.minor, ignoring patch version
    local expected_major_minor=$(echo "${expected_gateway_api_version}" | sed 's/^\(v\?[0-9]\+\.[0-9]\+\).*/\1/')
    local actual_major_minor=$(echo "${gateway_api_version}" | sed 's/^\(v\?[0-9]\+\.[0-9]\+\).*/\1/')
    if [[ ${actual_major_minor} != ${expected_major_minor} ]]; then
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
EXIT_VALUE=0

# Function to validate a specific implementation directory
validate_implementation_dir() {
    local implementation_dir=$1
    local gateway_api_version=$2
    local implementation=$(basename "${implementation_dir}")

    info "Checking ${implementation} project directory for Gateway API version ${gateway_api_version}"

    if [[ -f "${implementation_dir}/README.md" ]]; then
    # Check if the README.md has broken links
        docker run -v "$(readlink -f "$implementation_dir"):/${implementation}:ro" --rm -i ghcr.io/tcort/markdown-link-check:stable /${implementation}/README.md
    else
        error "missing README.md in ${implementation_dir}"
        EXIT_VALUE=1
    fi
    for report in "${implementation_dir}"/*.yaml
    do
        # Skip if no yaml files exist (glob didn't match)
        [[ -e "${report}" ]] || continue
        check_report_fields "${report}" "${gateway_api_version}"
    done
}

# If specific directories are provided as arguments, validate only those
if [[ $# -gt 0 ]]; then
    info "Validating specific directories: $@"

    # First pass: resolve symlinks and deduplicate
    declare -A validated_dirs
    for target_dir in "$@"
    do
        # Normalize path (remove trailing slash, make relative to repo root)
        target_dir="${target_dir%/}"

        # Check if directory exists
        if [[ ! -d "${target_dir}" ]]; then
            error "Directory does not exist: ${target_dir}"
            EXIT_VALUE=1
            continue
        fi

        # Resolve symlinks to canonical path
        canonical_dir=$(readlink -f "${target_dir}")

        # Check if we've already validated this canonical path
        if [[ -v "validated_dirs[$canonical_dir]" ]]; then
            info "Skipping ${target_dir} (already validated as ${validated_dirs[$canonical_dir]})"
            continue
        fi

        # Extract version from the resolved canonical path
        version=$(echo "${canonical_dir}" | sed -n 's|.*/conformance/reports/\(v[0-9]\+\.[0-9]\+\(\.[0-9]\+\)\?\)/.*|\1|p')
        if [[ -z "${version}" ]]; then
            error "Could not extract version from path: ${canonical_dir}"
            EXIT_VALUE=1
            continue
        fi

        if ! check_ge_v1.1.0 "${version}"; then
            info "Skipping ${target_dir} (version ${version} is < v1.1.0)"
            continue
        fi

        # Mark this canonical path as validated and store the original path for display
        validated_dirs[$canonical_dir]="${target_dir}"

        validate_implementation_dir "${canonical_dir}" "${version}"
    done
else
    # No arguments provided, validate all directories (original behavior)
    for dir in ${REPORTS_DIR}/*
    do
        element="${dir##*/}"
        if check_ge_v1.1.0 "${element}"; then
            if [[ -d "${dir}" ]]; then
                gateway_api_version="${element}"
                for implementation_dir in ${dir}/*
                do
                    validate_implementation_dir "${implementation_dir}" "${gateway_api_version}"
                done
            fi
        fi
    done
fi

exit ${EXIT_VALUE}
